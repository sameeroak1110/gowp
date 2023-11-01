package gowp

import (
	"fmt"
	"math/rand"
	"time"
	"runtime"
	"context"
	"sync"
	"sync/atomic"

	"github.com/sameeroak1110/gowp/helper"
)


/* *****************************************************************************
Description : gowp package init().

Arguments   : NA

Return value: NA

Additional note:
- Seeding to nanoseconds granularity. Introduced a small usec delay before a new worker
is spun off. This delay ensures that the workers don't start off all at once.
***************************************************************************** */
func init() {
	rand.Seed(time.Now().UnixNano())     // seeding to nanoseconds granularity.
	runtime.GOMAXPROCS(runtime.NumCPU()) // allocates one logical processor for the scheduler to use
	newPoolID = 0
}



/* *****************************************************************************
Description : Returns worker-pool context. This context is created from the upstream.

Receiver    :
*WorkerPool: Reference of the newly created worker-pool.

Implements  : NA

Arguments   : NA

Return value:
1> context.Context: Worker-pool context.

Additional note: NA
***************************************************************************** */
func (pwp *WorkerPool) GetContext() context.Context {
	return pwp.ctx
}


/* *****************************************************************************
Description : Returns worker-pool context cancel function. This context function is created from
the upstream is part of the context of worker-pool.

Receiver    :
*WorkerPool: Reference of the newly created worker-pool.

Implements  : NA

Arguments   : NA

Return value:
1> context.CancelFunc: Worker-pool context cancel function.

Additional note: NA
***************************************************************************** */
func (pwp *WorkerPool) GetCancelFunc() context.CancelFunc {
	return pwp.cancelFunc
}


/* *****************************************************************************
Description : Returns max job cnt a worker-pool will execute if WorkerPoolOptions.ShouldTerminate
is set to true.

Receiver    :
*WorkerPool: Reference of the newly created worker-pool.

Implements  : NA

Arguments   : NA

Return value:
1> int: WorkerPool.maxJobCnt

Additional note: NA
***************************************************************************** */
func (pwp *WorkerPool) GetMaxJobCnt() int {
	return pwp.maxJobCnt
}


/* *****************************************************************************
Description : Creates a new worker-pool instance. New worker-pool is created through this function.
Size of job-queue is 100 times the number of workers (denoted by wpsize in the function call).

Receiver    : NA

Implements  : NA

Arguments   :
1> tmpctx context.Context: Worker-pool context. This context is created in the upstream.
2> cfunc context.CancelFunc: Cancel function of context.
3> wpsize int32: Number of workers, denotes worker-pool size. Minimum size is 10 and maximum
allowed size is 100.
TODO: 4> isResponse bool: true if business logic needs job execution response.
TODO: 5> jobctrl bool: context-timeout in the exec functions.
6> opts WorkerPoolOptions: WorkerPool options. start-message, cancel-message, maxjobcnt, and shouldterminate flag.

Return value:
1> *WorkerPool: Reference to the newly created worker-pool.
2> int32: ID of newly created worker-pool.
3> error: Only possible error is while creating a UUID.

Additional note:
wcnt, avlwcnt, and ID are book-keeping members. They're updated using atomic.AddInt32() function.
atomic.Add... functions ensure thread safe update of their argument. These members need not be of
int32 type, however, there's no atomic.Add... function for uint8 or int8 type value.
wcnt keeps track of the number of workers that're in the run at any given instance in time.
avlwcnt keeps track of the number of available workes at any given instance in time.
***************************************************************************** */
/* func NewWorkerPool(tmpctx context.Context, cfunc context.CancelFunc, wpsize int32, _name, smsg, cmsg string, isResponse bool,
	jobctrl bool) (*WorkerPool, int32, error) { */
func NewWorkerPool(tmpctx context.Context, cfunc context.CancelFunc, wpsize int32, _name, smsg, cmsg string, opts WorkerPoolOptions) (*WorkerPool, int32, error) {
	if wpsize < minWPSize {
		wpsize = minWPSize
	}

	if wpsize > maxWPSize {
		wpsize = maxWPSize
	}

	jpsize := wpsize * jpwpfactor

	uuid, err := helper.NewUUID()
	if err != nil {
		return nil, 0, err
	}

	wpID := atomic.AddInt32(&newPoolID, 1)
	pwp := &WorkerPool {
		id: wpID,
		uuid: uuid,
		jobq: make(chan Job, jpsize),
		workers: make(chan int32, wpsize),
		ctx: tmpctx,
		cancelFunc: cfunc,
		singletonCtrl: &sync.Mutex{},
		wg: sync.WaitGroup{},
		name: _name,
		startMsg: smsg,
		cancelMsg: cmsg,
		maxJobCnt: opts.MaxJobCnt,
		shouldTerminate: opts.ShouldTerminate,
	}

	for i := int32(1); i <= wpsize; i++ {
		pwp.workers <- i
	}
	pwp.avlwcnt = wpsize

	return pwp, pwp.id, nil
}


func (pwp *WorkerPool) exec(job Job, wid, wcnt, avlwcnt int32) {
	defer func() {
		pwp.workers <- wid  // one more worker is made available.
		atomic.AddInt32(&pwp.wcnt, -1)
		atomic.AddInt32(&pwp.avlwcnt, 1)
		pwp.wg.Done()
	}()

	c := make(chan JobStatus, 1)
	wg := sync.WaitGroup{}
	wg.Add(1)  // one go-routine for the job process method.
	go func(pwg *sync.WaitGroup) {
		defer pwg.Done()

		js := JobStatus{}
		js.data, js.err = job.data.Process(pwp.GetContext(), cancelFunc)
		c <- js
	}(&wg)

	select {
		case <-pwp.GetContext().Done():
			wg.Wait()
			return

		// TODO: result processing.
		case <-c:
			wg.Wait()
	}

	return
}


/* *****************************************************************************
Description : Starts the worker-pool.

Receiver    : *WorkerPool 

Implements  : NA

Arguments   :
1> ctx context.Context: Context passed from the upstream.
2> pwg *sync.WaitGroup: Passed 

Return value: NA

Additional note:
ctx and pwg combiningly used for concurrency control.
***************************************************************************** */
func (pwp *WorkerPool) Start(ctx context.Context, pwg *sync.WaitGroup) {
	defer func() {
		if panicState := recover(); panicState != nil {
			fmt.Printf("ERROR: Recovered from panic: state: %#v", panicState)
		}

		pwg.Done()
	}()

	pwp.singletonCtrl.Lock()
	if pwp.startFlag {
		pwp.singletonCtrl.Unlock()
		return
	}
	pwp.startFlag = true
	pwp.stopFlag = false
	pwp.singletonCtrl.Unlock()

	for {
		select {
			//case <-pwp.GetContext().Done():
			case <-ctx.Done():
				pwp.wg.Wait()  // waits for each exec() method finish its respective job.
				return

			case job := <-pwp.jobq:  // there's a job for one of the workers.
				select {
					case wid := <-pwp.workers:
						if wid > 0 {
							wcnt := atomic.AddInt32(&pwp.wcnt, 1)
							avlwcnt := atomic.AddInt32(&pwp.avlwcnt, -1)
							//jobcnt := atomic.AddUint64(&pwp.jobcnt, 1)
							atomic.AddUint64(&pwp.jobcnt, 1)
							pwp.wg.Add(1)
							//time.Sleep(time.Duration(helper.RandomInt(1000, 2000)) * time.Millisecond)
							go pwp.exec(job, wid, wcnt, avlwcnt)
						}
						break
				}
				break
		}
	}

	return
}


/* *****************************************************************************
Description : Stops a worker-pool.

Receiver    : *WorkerPool 

Implements  : NA

Arguments   : NA

Return value: NA

Additional note: NA
***************************************************************************** */
func (pwp *WorkerPool) Stop() {
	pwp.singletonCtrl.Lock()
	defer pwp.singletonCtrl.Unlock()

	if pwp.stopFlag {
		return
	}

	pwp.stopFlag = true
	pwp.startFlag = false

	close(pwp.jobq)
	close(pwp.workers)

	return
}
