package gowp

import (
	"context"
	"sync"
)


type Job struct {
	id uint64         // generated internally using atomic.AddUint64().
	name string       // job name, optional.
	data JobProcessor // data part, any type that implements JobProcessor.
}

// - a workerpool has ID, UUID, and a name.
// - context orientation - context and waitgroup - ensures that a workerpool instance can be terminated
// gracefully.
// - startMsg and cancelMsg are some informative messages that can be logged.
// - startMsg is supposedly to indicate the wp has been started. cancelMsg is supposedly to indicate that
// the wp has been cancelled/stopped.
type WorkerPool struct {
	id int32                      // generated internally using atomic.AddInt32().
	uuid string                   // generated internally.
	name string                   // user defined name of worker-pool.
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
	singletonCtrl *sync.Mutex     // ensures worker-pool is started/stopped only once while it's in action.
	startFlag bool
	stopFlag bool
	isResponse bool               // true if upstream needs job execution status.
	jobctrl bool                  // context-timeout in exec function.

	// worker-pool cancellation:
	maxJobCnt       int    // maximum of jobs worker-pool has executed before cancellation. Process() method of JobProcessor{} interface uses this count.
	shouldTerminate bool   // if true, Process() method of JobProcessor{} interface invokes cancel function to terminate the worker-pool.
}

type WorkerPoolOptions struct {
	MaxJobCnt       int    // maximum of jobs worker-pool has executed before cancellation. Process() method of JobProcessor{} interface uses this count.
	                       // if current jobcnt reaches MaxJobCnt, Process() may invoke cancellation if ShouldTerminate flag is set to true.
						   // default value is 0 to indicate cancellation is ignored.
	ShouldTerminate bool   // if true, Process() method of JobProcessor{} interface invokes cancel function to terminate the worker-pool.
}

// Status of execution of each job.
type JobStatus struct {
	data interface{}
	err error
}
