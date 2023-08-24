## Why do we need a worker-pool in the first place?
Those who work in go must have come across this statement that go is cabable enough to run hundreds of thousands of go-routines at any given instance in time.
Apparently, this indeed is true. But then, in my humble opinion, we should raise a quiestion, why would we really want our service to run hundreds of thousands of concurrent go-routines.

Go's memory model is amazing and the effectiveness of go's memory management across multiple go-routines is equally good. As a matter of fact, go's GC is one of the fastest ones amongst the GC oriented programming languages. As well, the small size of go-routines is very effective in the sense that it's far better when it comes to context switching and swap-in and swap-out. But, no matter how effective the go-routine management is, running those many go-routines concurrently (I'm using concurrently and parallelly loosely here, and thus, to mean the same) should still make us believe that we may need to revist the approach.

Keeping the concurrently running go-routines at a generously limited size is a better approach. This is important because creating too many worker go-routines can lead to performance issues and resource contention. Though, size of each go-routine is small and typically is in the range of 2K to 4K, sum total of the memory consumed by a huge number of concurrently running go-routines is still going to be a costlier affair as, no matter how rich in number and configurations, the resources will still be
limited.
The other important aspect is as the number of go-routines increases chances of thrashing will increase
in equal proportions.
The sum effect of all this is the scalability issue. I'm dis-accounting the performance issue since it'll mostly be related to the way each job is going to be executed.
As anyway since we spoke about performance issue and scalability issue, they're clearly different from each other:
Performance issue is, if our algorithim is taking more time for a single task execution, we've a performance problem.
Scalability issu is, if our system performs well for one task execution, however, slows down if the size of set of tasks increases.

Thus, spinning off more and more go-routines to execute concurrently at any given instance in time
may result into sacalability issue.
Thus, the first approach should be limiting the number of go-routines. But at the same time each job
needs to be executed. And this should happen judicially, meaning no job - once its execution is started - can be dropped in order to control the number of concurrently running go-routines.
Therefore, the only possibility remains is to control the number of concurrently running go-routines.
The best way is through creating a team of fixed number of go-routines.
Thus, each go-routine in the team may either be free or be executing a job at any given instance
in time. This team of go-routines is a go-routine pool, also termed as worker-pool in generic terms.

## Implementation
A simple implementation strategy is by using buffered channel, one each for jobs and workers.
Workers are meant to work on the jobs. The jobs are published by any upstream application.
It'd rather be more appropriate to implement worker-pool as a distributed application rather than
making it a part of a monolith. This's a typical implementation scenario in publisher-subscriber model.
Job queue has jobs that're pushed from the upstream. Job queue should be large enough to accommodate the incoming traffic. Workers on the other hand are awaiting for jobs to arrive. The moment
there's a job in the job queue, one of the workers is picked up and assigned the job to work upon.
From the implementation stand point, both job queue and worker-pool are implemented as buffered
channels.

Let's look at the WorkerPool type to start with.
The most prominent members of WorkerPool are **jobPool**, **workers**, **ctx**, and **wg**.
**jobPool** is a buffered channel of **Job** type and **workers** is a buffered channel of int32.
**ctx** and **wg** help in concurrency control. Each downstream go-routine is context oriented
so that the context cancellation from upstream is handled gracefully. The objective is
to let each downstream go-routine finish its job gracefully and in entirety.

**startLock** and **stopLock** are used for singleton control. A worker-pool once started
shouldn't be accidentally restarted. Similarly a worker-pool once stopped shouldn't be accidentally
stopped again. **stopLock** is not being used at present.
Note: sync.Once along with func (o *Once) Do(f func()) from sync package can also be used to
achieve singleton control.

There're some book-keeping members in the WorkerPool, they're wcnt, avlwcnt, and jobcnt. wcnt denotes
the number of concurrent workers in the run and avlwcnt denotes the number of workers that're waiting
for jobs.
jobcnt denotes the total number of jobs served by this worker-pool instance since it was started.

### Types:
```
type Job struct {
	ID uint64         // generated internally using atomic.AddUint64().
	Name string       // job name, optional.
	Data JobProcessor // data part, any type that implements JobProcessor.
}

type WorkerPool struct {
	ID int32                      // generated internally using atomic.AddInt32().
	UUID string                   // generated internally.
	Name string                   // user defined name of worker-pool.
	jobPool chan Job              // jobs that workers are going to work on.
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
```

### Interfaces:
```
// - A type that implements JobProcessor is assumed to have all the necessary data.
// - Process() method definition on a specific type is assumed to make use of these data.
// - The upstream is supposed to pass context to the downstream - ie, Process() method - so that
// each running job (Process() method in execution) can terminate gracefully to avoid go-routine
// leak.
// - This's essential as a specific business requirment may need each specific running job to finish
// in entirety and gracefully and thus conclude logically. For instance, a job is updating some rows
// in a db table and needs to handle status of the updation.
// - However, there may be a circumstance where the business logic need is to terminate the running
// job instantly when it receives the context cancellation.
// - Therefore, it's upto the implementation how to handle the upstream context cancellation.
type JobProcessor interface {
	GetName() string
	Process(context.Context) (interface{}, error)
}

```

### Exported functions/methods:
A new worker-pool is created using function **NewWorkerPool()**.
```
func NewWorkerPool(tmpctx context.Context, cfunc context.CancelFunc,
	wpsize int32, name, smsg, cmsg string) (*WorkerPool, int32, error)
```

Please read the function header for more details. main() of sampleapp shows how to invoke this function.

Newly created worker-pool is started using method **Start()** over pointer receiver of type WorkerPool
returned by **NewWorkerPool()**.
```
func (pwp *WorkerPool) Start(ctx context.Context, pwg *sync.WaitGroup)
```
main() of sampleapp shows how to invoke this function.

A job is added to the worker-pool through method Add() on pointer receiver of type WorkerPool.
```
func (pwp *WorkerPool) AddJob(job JobProcessor)
```
main() of sampleapp has addjobs() function that demonstrates how jobs are added to a worker-pool.


A job is pulled out from jobq and a worker is assigned to handle the pulled job.
Method (*WorkerPool).exec() executes the job. A variable of type Job has a member Data of type
JobProcessor interface. Upstream publishes job as an object of some type **t** where **t** should implement this interface so that exec() method can invoke the actual implementation of Process() methods
over object of **t**.
Please go through the sampleapp for the example.


## Sample application
Sample application has a function function addjobs(). It's invoked as a go-routine. addjobs() publlishes
jobs until parent context created in the main() is cancelled.
The sampleapp is using a logger (github.com/sameeroak1110/logger) package. The logger package is supposed to exit in the last so as to allow each job processor dump the logs. This example application
is executing each job processor method to end gracefully rather than
Before application ends, main() waits for all the remaining logs to get flushed out.
It's a good practice to do final clean up before the application exits.

TestJobData implements JobProcessor interface. addjobs() function is publishing jobs each of
type TestJobData.

## TODO
gowp package doesn't have result processing enabled at present. The way the result can be processed is
only known to the application that uses gowp. It's, therefore, a good idea to define an interface with
method for job result process.

There can be following interfaces.
```
type JobProcessor interface {
    GetName() string
    Process(context.Context) (JobResultProcessor, error)
}       

type JobResultProcessor interface {
    ProcessResult(context.Context) (interface{}, error)
}
```

Method Process() of JobProcessor returns an object of type JobresultProcessor. ProcessResult() method
of JobresultProcessor interface is supposed to deal with the result.

On cancellation of the parent context, it's possible that job queue still has some jobs that're
yet to be served. exec() method waits for the job processing go-routine to finish.
In turn, the Start() method waits for all the job processing go-routines to finish. This may result
some jobs being dropped from the job queue.
