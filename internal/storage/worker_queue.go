package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/transcode"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/lucsky/cuid"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

type JobStatus string

const (
	Created    JobStatus = "Created"
	InProgress JobStatus = "InProgress"
	Errored    JobStatus = "Errored"
	Done       JobStatus = "Done"
)

type Job struct {
	ID         string    `json:"id,omitempty"`
	Status     JobStatus `json:"status,omitempty"`
	StatusTime time.Time `json:"statusTime,omitempty"`
	Attempt    int       `json:"attempt,omitempty"`
	Version    int       `json:"version"`
}

type TranscodeJob struct {
	Job
	TranscodeType constants.TranscodeType `json:"transcodeType,omitempty"`
	VideoFileImpl transcode.VideoFileImpl `json:"videoFileImpl,omitempty"`
	VideoID       int64                   `json:"videoId,omitempty"`
}

func (t *TranscodeJob) MarshalBinary() (data []byte, err error) {
	return json.Marshal(t)
}

func (t *TranscodeJob) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, t)
}

type TranscodeWorker interface {
	EnqueueJob(ctx context.Context, job *TranscodeJob) error
	DequeueJob(ctx context.Context, work func(ctx context.Context, job TranscodeJob) error) error
	HandleErrored(ctx context.Context) error
}

type TranscodeWorkerImpl struct {
	client *redis.Client
	lock   *redislock.Client
}

type TranscodeFileWorkerImpl struct {
	driver *Driver
}

func (t TranscodeFileWorkerImpl) EnqueueJob(ctx context.Context, job *TranscodeJob) error {
	job.ID = cuid.New()
	job.Status = Created
	job.StatusTime = time.Now()
	return t.driver.Write("jobs", job.ID, job)
}

func (t TranscodeFileWorkerImpl) DequeueJob(ctx context.Context, work func(ctx context.Context, job TranscodeJob) error) error {
	job := TranscodeJob{}

	err := t.driver.GetRandom("jobs", &job)
	if err != nil {
		return err
	}
	if job.ID == "" {
		select {
		case <-time.After(5 * time.Second):
		case <-ctx.Done():
			return nil
		}
		return nil
	}
	job.Attempt++
	job.Status = InProgress
	job.StatusTime = time.Now()
	err = t.driver.Write("jobs", job.ID, job)
	if err != nil {
		return errors.Wrap(err, "Failed to save job")
	}

	err = work(ctx, job)
	if err != nil {
		return err
	}

	return t.driver.Delete("jobs", job.ID)
}

func (t TranscodeFileWorkerImpl) HandleErrored(ctx context.Context) error {
	return nil
}

func GetTranscodeWorker() TranscodeWorker {
	workPath := config.GetConfig().WorkDirectory

	client, err := New(workPath, nil)
	if err != nil {
		panic(err)
	}

	return TranscodeFileWorkerImpl{client}
}

func (t TranscodeWorkerImpl) HandleErrored(ctx context.Context) error {
	lock, err := t.lock.Obtain(ctx, "lock:transcode-job-inprogress", 5*time.Minute, nil)
	if err != nil {
		return err
	}
	defer lock.Release(ctx)

	items, err := t.client.LRange(ctx, "transcode-job-queue-inprogress", 0, 100).Result()
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		err := lock.Refresh(ctx, 5*time.Minute, nil)
		if err != nil {
			return errors.Wrap(err, "Failed to refresh lock")
		}
		jobLock, err := t.lock.Obtain(ctx, "lock:transcode-job:"+items[i], 1*time.Minute, nil)
		if err == redislock.ErrNotObtained {
			continue // expected if job is currently being worked on
		} else if err != nil {
			return err
		}

		job := TranscodeJob{}

		err = t.client.Get(ctx, "transcode-job:"+items[i]).Scan(&job)
		if err != nil {
			jobLock.Release(ctx)
			return errors.Wrap(err, "Failed to get job")
		}
		if job.Status == Done {
			log.Info().Str("job", items[i]).Msg("Removing Done job")
			err := t.client.LRem(ctx, "transcode-job-queue-inprogress", 0, items[i]).Err()
			if err != nil {
				jobLock.Release(ctx)
				return err
			}
		} else {
			tx := t.client.TxPipeline()
			if job.Attempt >= 3 {
				log.Info().Str("job", items[i]).Msg("Removing job after too many attempts")
			} else {
				log.Info().Str("job", items[i]).Msg("Enqueuing unfinished job")
				err = tx.LPush(ctx, "transcode-job-queue", items[i]).Err()
				if err != nil {
					jobLock.Release(ctx)
					tx.Discard()
					return err
				}
			}
			err := tx.LRem(ctx, "transcode-job-queue-inprogress", 0, items[i]).Err()
			if err != nil {
				jobLock.Release(ctx)
				tx.Discard()
				return err
			}
			_, err = tx.Exec(ctx)
			if err != nil {
				jobLock.Release(ctx)
				return err
			}
			jobLock.Release(ctx)
		}
	}

	return nil
}

func (t TranscodeWorkerImpl) EnqueueJob(ctx context.Context, job *TranscodeJob) error {
	id, err := t.client.Incr(ctx, "transcode-job-id-count").Result()
	if err != nil {
		return errors.Wrap(err, "Failed to generate new id")
	}
	job.ID = fmt.Sprintf("%d", id)
	job.Status = Created
	job.StatusTime = time.Now()
	tx := t.client.TxPipeline()

	if err != nil {
		return err
	}
	err = tx.Set(ctx, "transcode-job:"+job.ID, job, 0).Err()
	if err != nil {
		return err
	}
	err = tx.LPush(ctx, "transcode-job-queue", job.ID).Err()
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx)
	return err
}

func (t TranscodeWorkerImpl) DequeueJob(ctx context.Context, work func(ctx context.Context, job TranscodeJob) error) error {
	jobID, err := t.client.BRPopLPush(ctx, "transcode-job-queue", "transcode-job-queue-inprogress", 60*time.Second).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "Failed to dequeue job")
	}

	lock, err := t.lock.Obtain(ctx, "lock:transcode-job:"+jobID, 5*time.Minute, nil)
	if err != nil {
		return errors.Wrap(err, "Failed to get exclusive lock on job: "+jobID)
	}
	defer func() {
		lerr := lock.Release(ctx)
		if lerr != nil {
			log.Err(lerr).Msg("Failed to release lock")
		}
	}()

	job := TranscodeJob{}

	err = t.client.Get(ctx, "transcode-job:"+jobID).Scan(&job)
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "Failed to get job")
	}
	if job.Status == Done {
		return nil
	}

	job.Attempt++
	job.Status = InProgress
	job.StatusTime = time.Now()

	err = t.client.Set(ctx, "transcode-job:"+jobID, &job, 0).Err()
	if err != nil {
		return err
	}

	log.Debug().Str("id", job.ID).Msg("Working on transcode job")
	done := make(chan error, 1)

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		lerr := work(childCtx, job)
		done <- lerr
		close(done)
	}()

	errCount := 0
	err = nil
	loop := true
	for loop {
		select {
		case <-childCtx.Done():
			return childCtx.Err()
		case wErr := <-done:
			err = wErr
			loop = false
		case <-time.After(1 * time.Minute):
			err = lock.Refresh(childCtx, 5*time.Minute, nil)
			if err != nil {
				errCount++
			} else {
				errCount--
			}
			if errCount >= 3 {
				cancel()
				return errors.Wrap(err, "Failed to refresh exclusive lock")
			}
		}
	}
	if err != nil {
		return err
	}
	job.Status = Done
	err = t.client.Set(ctx, "transcode-job:"+jobID, &job, 72*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}
