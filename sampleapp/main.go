package main

import (
	"fmt"
	"runtime"
	"math/rand"
	"context"
	"os"
	"bufio"
	"sync"
	"time"

	"github.com/sameeroak1110/logger"
	"github.com/sameeroak1110/gowp"
	"github.com/sameeroak1110/gowp/helper"

	"sampleapp/job"
)

const pkgname string = "main"


func addjobs(ctx context.Context, pwg *sync.WaitGroup, pwp *gowp.WorkerPool) {
	defer func() {
		if panicState := recover(); panicState != nil {
			logger.Log(pkgname, logger.ERROR, "Recovered from panic: state: %#v", panicState)
		}

		logger.Log(pkgname, logger.WARNING, "exiting.")
		pwg.Done()
	}()

	i := int64(1)
	for {
		select {
			case <-ctx.Done():
			//case <-pwp.GetContext().Done():
				logger.Log(pkgname, logger.WARNING, "got ctx cancel.")
				return

			default:
				//waitForMS := helper.RandomInt(1, 5)
				//waitForMS := helper.RandomInt(10, 100)
				waitForMS := helper.RandomInt(100, 300)
				job := job.TestJobData {
					ID: i,
					Name: fmt.Sprintf("TestJob-%d", i),
				}
				time.Sleep(time.Duration(waitForMS) * time.Millisecond)
				pwp.AddJob(job)
				logger.Log(pkgname, logger.DEBUG, "[%s:%d]  waited for %d ms before new job(%s:%d) was added\n", pwp.Name, pwp.ID, waitForMS, job.Name, job.ID)
				//fmt.Printf("[addJobs]  DEBUG: [%s:%d]  waited for %d ms before new job(%s:%d) was added\n", pwp.Name, pwp.ID, waitForMS, job.Name, job.ID)
				i++
		}
	}

	return
}


func main() {
	rand.Seed(time.Now().UnixNano())  // seeding to nanoseconds granularity.
	runtime.GOMAXPROCS(runtime.NumCPU()) // allocates one logical processor for the scheduler to use

	ctx := context.Background()
	ctxParent, cancelParent := context.WithCancel(ctx)
	ctxAppEnd, cancelFuncAppEnd := context.WithCancel(ctx)

	wg := sync.WaitGroup{}

	//endapp:= make(chan bool)
	appdone := make(chan bool)
	//if isSuccess := logger.Init(ctxAppEnd, endapp, appdone, false, "./", "DEBUG"); !isSuccess {
	if isSuccess := logger.Init(ctxAppEnd, appdone, false, "./", "DEBUG"); !isSuccess {
		fmt.Printf("Error-102: Unable to initilize logger, exiting application.\n")
		os.Exit(102)
	}
	logger.Log(pkgname, logger.DEBUG, "logger initialized.")
	logger.Log(pkgname, logger.DEBUG, "log dispatcher started.")

	pwp, _, err := gowp.NewWorkerPool(ctxParent, cancelParent, 100, "wp1", fmt.Sprintf("started wp-1"), fmt.Sprintf("cancelled wp-1"))
	if err != nil {
		logger.Log(pkgname, logger.ERROR, "new worker-pool error: %s\n", err.Error())
		return
	}

	wg.Add(2)
	logger.Log(pkgname, logger.DEBUG, "application started.")
	time.Sleep(time.Duration(helper.RandomInt(1000, 2000)) * time.Millisecond)
	go pwp.Start(ctxParent, &wg)
	time.Sleep(time.Duration(helper.RandomInt(100, 1000)) * time.Millisecond)
	go addjobs(ctxParent, &wg, pwp)

	go func() {
		pS := bufio.NewScanner(os.Stdin)
		pS.Scan()
		// pReader := bufio.NewReader(os.Stdin)
		// pReader.ReadString('\n')
		fmt.Println("WARNING: received termination.")
		cancelParent()
	}()

	wg.Wait()
	logger.Log(pkgname, logger.DBGRM, "ch-2")
	cancelFuncAppEnd()
	//fmt.Println("INFO: application stopped.")
	fmt.Printf("INFO: application stopped (%t).\n", <-appdone)

	return
}
