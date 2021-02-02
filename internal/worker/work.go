package worker

import (
	"context"
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/utils"
	"media-web/internal/web"
	"time"

	"github.com/floostack/transcoder"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type WorkerContext struct {
	GetTranscoder func() transcoder.Transcoder
	SonarrClient  web.SonarrClient
	RadarrClient  web.RadarrClient
	Enqueuer      WorkScheduler
	Sleep         func(d time.Duration)
}

type WorkScheduler interface {
	EnqueueUnique(jobName string, args map[string]interface{}) (*work.Job, error)
}

var worker = work.NewEnqueuer(config.GetConfig().JobQueueNamespace, &storage.RedisPool)

var Enqueuer = worker

func (c *WorkerContext) Log(job *work.Job, next work.NextMiddlewareFunc) error {
	log.Info().Str("jobId", job.ID).Msg("Starting job: " + job.ID)
	return next()
}

func (c *WorkerContext) Metrics(job *work.Job, next work.NextMiddlewareFunc) error {
	utils.InflightJob.WithLabelValues(job.Name).Inc()
	defer utils.InflightJob.WithLabelValues(job.Name).Dec()
	start := time.Now()
	err := next()
	dur := time.Since(start)
	status := "success"
	if err != nil {
		status = "error"
	}
	utils.JobCount.WithLabelValues(job.Name, status).Inc()
	utils.JobTime.WithLabelValues(job.Name, status).Observe(dur.Seconds())
	return err
}

var workerContext = WorkerContext{
	GetTranscoder: GetTranscoder,
	SonarrClient:  web.GetSonarrClient(),
	RadarrClient:  web.GetRadarrClient(),
	Enqueuer:      Enqueuer,
	Sleep:         time.Sleep,
}

func GetWorkerContext() WorkerContext {
	return workerContext
}

type WorkerPoolFactory interface {
	NewWorkerPool(ctx interface{}, concurrency uint, namespace string, pool *redis.Pool) WorkerPool
}

type WorkerPoolFactoryImpl struct {
}

func (f WorkerPoolFactoryImpl) NewWorkerPool(ctx interface{}, concurrency uint, namespace string, pool *redis.Pool) WorkerPool {
	return &WorkerPoolImpl{
		pool: work.NewWorkerPool(ctx, concurrency, namespace, pool),
	}
}

type WorkerPool interface {
	Middleware(fn interface{})
	JobWithOptions(name string, jobOpts work.JobOptions, fn interface{})
	Start()
	Stop()
}

type WorkerPoolImpl struct {
	pool *work.WorkerPool
}

func (w WorkerPoolImpl) Middleware(fn interface{}) {
	w.pool.Middleware(fn)
}
func (w WorkerPoolImpl) JobWithOptions(name string, jobOpts work.JobOptions, fn interface{}) {
	w.pool.JobWithOptions(name, jobOpts, fn)
}
func (w WorkerPoolImpl) Start() {
	w.pool.Start()
}
func (w WorkerPoolImpl) Stop() {
	w.pool.Stop()
}

func StartWorkerPool(ctx context.Context, wctx WorkerContext, factory WorkerPoolFactory) {
	log.Info().Msg("Starting worker pool")
	// Note: normally the worker context isn't shared and would be unique per job
	// However, here we use it as a mechanism to inject dependencies into the job handler
	pool := factory.NewWorkerPool(wctx, 20, config.GetConfig().JobQueueNamespace, &storage.RedisPool)
	pool.Middleware(wctx.Log)
	pool.Middleware(wctx.Metrics)

	pool.JobWithOptions(constants.TranscodeJobType, work.JobOptions{
		Priority:       1,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 1,
	}, wctx.TranscodeJobHandler)

	pool.JobWithOptions(constants.UpdateSonarrJobName, work.JobOptions{
		Priority:       2,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, wctx.UpdateTVShow)

	pool.JobWithOptions(constants.UpdateRadarrJobName, work.JobOptions{
		Priority:       2,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, wctx.UpdateMovie)

	// Start processing jobs
	pool.Start()

	<-ctx.Done()

	// Stop the pool
	pool.Stop()
	log.Info().Msg("Worker pool stopped")
}
