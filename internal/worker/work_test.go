package worker

import (
	context2 "context"
	"media-web/internal/constants"
	"testing"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
)

type MockWorkerPoolFactory struct {
	makePool func() WorkerPool
}

type MockWorkerPool struct {
	middleware     func(fn interface{})
	jobWithOptions func(name string, jobOpts work.JobOptions, fn interface{})
	start          func()
	stop           func()
}

func (m MockWorkerPool) Middleware(fn interface{}) {
	m.middleware(fn)
}

func (m MockWorkerPool) JobWithOptions(name string, jobOpts work.JobOptions, fn interface{}) {
	m.jobWithOptions(name, jobOpts, fn)
}

func (m MockWorkerPool) Start() {
	m.start()
}

func (m MockWorkerPool) Stop() {
	m.stop()
}

func (w MockWorkerPoolFactory) NewWorkerPool(ctx interface{}, concurrency uint, namespace string, pool *redis.Pool) WorkerPool {
	return w.makePool()
}

func TestWorkerPoolSetup(t *testing.T) {
	ctx, cancel := context2.WithCancel(context2.Background())
	cancel()
	var start = false
	var stop = false
	var middleware = false
	var jobs = make([]string, 0)
	makePool := func() WorkerPool {
		return MockWorkerPool{
			middleware: func(fn interface{}) { middleware = true },
			jobWithOptions: func(name string, jobOpts work.JobOptions, fn interface{}) {
				jobs = append(jobs, name)
			},
			start: func() { start = true },
			stop:  func() { stop = true },
		}
	}

	context := WorkerContext{}
	err := context.Log(&work.Job{}, func() error {
		return nil
	})

	StartWorkerPool(ctx, &context, MockWorkerPoolFactory{makePool: makePool})

	assert.NoError(t, err)
	assert.True(t, start)
	assert.True(t, stop)
	assert.True(t, middleware)
	assert.ElementsMatch(t, jobs, []string{constants.UpdateSonarrJobName, constants.UpdateRadarrJobName})
}
