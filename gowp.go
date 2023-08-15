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


func (pwp *WorkerPool) GetContext() context.Context {
	return pwp.ctx
}


func NewWorkerPool(tmpctx context.Context, cfunc context.CancelFunc, wpsize int32, name, smsg, cmsg string) (*WorkerPool, int32, error) {
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
		ID: wpID,
		UUID: uuid,
		Name: name,
		jobPool: make(chan Job, jpsize),
		workers: make(chan int32, wpsize),
		startMsg: smsg,
		cancelMsg: cmsg,
		ctx: tmpctx,
		cancelFunc: cfunc,
		startLock: &sync.Mutex{},
		wg: sync.WaitGroup{},
	}

	for i := int32(1); i <= wpsize; i++ {
		pwp.workers <- i
	}
	pwp.avlwcnt = wpsize

	return pwp, pwp.ID, nil
}


func (pwp *WorkerPool) exec(job Job, wid, wcnt, avlwcnt int32) {
	//var err error

	defer func() {
		// TODO: log jobstatus
		/* jobStatus := "JobDone"
		if err != nil {
			jobStatus = "JobError: " + err.Error()
		} */

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
		js.Data, js.Err = job.Data.Process(pwp.GetContext())
		c <- js
	}(&wg)

	select {
		case <-pwp.GetContext().Done():
			wg.Wait()
			return

		// TODO: need to push the result to some channel.
		/* case js := <-c:
			err = js.Err
			wg.Wait() */

		case <-c:
			wg.Wait()
	}

	return
}


func (pwp *WorkerPool) Start(ctx context.Context, pwg *sync.WaitGroup) {
	defer func() {
		if panicState := recover(); panicState != nil {
			fmt.Printf("ERROR: Recovered from panic: state: %#v", panicState)
		}

		pwg.Done()
	}()

	pwp.startLock.Lock()
	if pwp.startFlag {
		pwp.startLock.Unlock()
		return
	}
	pwp.startFlag = true
	pwp.startLock.Unlock()

	for {
		select {
			//case <-pwp.GetContext().Done():
			case <-ctx.Done():
				pwp.wg.Wait()  // waits for each exec() method finish its respective job.
				return

			case job := <-pwp.jobPool:  // there's a job for one of the workers.
				select {
					case wid := <-pwp.workers:
						if wid > 0 {
							wcnt := atomic.AddInt32(&pwp.wcnt, 1)
							avlwcnt := atomic.AddInt32(&pwp.avlwcnt, -1)
							jobcnt := atomic.AddUint64(&pwp.jobcnt, 1)
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
