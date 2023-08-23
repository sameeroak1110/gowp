package gowp

import (
	"context"
	"sync"
)


type Job struct {
	ID uint64         // generated internally using atomic.AddUint64().
	Name string       // job name, optional.
	Data JobProcessor // data part, any type that implements JobProcessor.
}

// - a workerpool has ID, UUID, and a name.
// - context orientation - context and waitgroup - ensures that a workerpool instance can be terminated
// gracefully.
// - startMsg and cancelMsg are some informative messages that can be logged.
// - startMsg is supposedly to indicate the wp has been started. cancelMsg is supposedly to indicate that
// the wp has been cancelled/stopped.
type WorkerPool struct {
	ID int32                      // generated internally using atomic.AddInt32().
	UUID string                   // generated internally.
	Name string                   // user defined name of worker-pool.
	jobq chan Job                 // jobs that workers are going to work on.
	jobcnt uint64                 // total no. of jobs served by this wp. updated using atomic.AddUint64().
	workers chan int32            // limited number of workers that are going to work on jobs.
	wcnt int32                    // no. of workers in action at any given instance in time. updated using atomic.AddInt32().
	avlwcnt int32                 // available workers at any given instance in time. updated using atomic.AddInt32().
	startMsg string               // optional worker-pool start message.
	cancelMsg string              // indicates the wp has been cancelled/stopped, optional.
	ctx context.Context           // passed on through upstream.
	cancelFunc context.CancelFunc // cancel function of context. passed on through upstream.
	wg sync.WaitGroup             // concurrency control, used in conjunction with ctx.
	startLock *sync.Mutex         // ensures workerpool is started only once while it's in the run at any given instance in time.
	startFlag bool
	stopLock *sync.Mutex          // ensures workerpool is stopped only once while it's in the run at any given instance in time.
	stopFlag bool
}

// Status of execution of each job.
type JobStatus struct {
	Data interface{}
	Err error
}
