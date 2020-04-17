package worker

import (
	"media-web/internal/config"
	"media-web/internal/constants"
	"media-web/internal/storage"
	"media-web/internal/web"
	"os"
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
)

type WorkerContext struct {
	GetTranscoder func() Transcoder
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

func StartWorkerPool(context WorkerContext, factory WorkerPoolFactory, stopChan <-chan os.Signal) {
	log.Info().Msg("Starting worker pool")
	// Note: normally the worker context isn't shared and would be unique per job
	// However, here we use it as a mechanism to inject dependencies into the job handler
	pool := factory.NewWorkerPool(context, 20, config.GetConfig().JobQueueNamespace, &storage.RedisPool)
	pool.Middleware(context.Log)

	pool.JobWithOptions(constants.TranscodeJobType, work.JobOptions{
		Priority:       1,
		MaxFails:       3,
		SkipDead:       false,
		MaxConcurrency: 1,
	}, context.TranscodeJobHandler)

	pool.JobWithOptions(constants.UpdateSonarrJobName, work.JobOptions{
		Priority:       2,
		MaxFails:       5,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, context.UpdateTVShow)

	pool.JobWithOptions(constants.UpdateRadarrJobName, work.JobOptions{
		Priority:       2,
		MaxFails:       5,
		SkipDead:       false,
		MaxConcurrency: 5,
	}, context.UpdateMovie)

	// Start processing jobs
	pool.Start()

	<-stopChan

	// Stop the pool
	pool.Stop()
	log.Info().Msg("Worker pool stopped")
}
