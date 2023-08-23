package gowp

import (
	"fmt"
	"sync/atomic"
)

func (pwp *WorkerPool) AddJob(job JobProcessor) {
	defer func() {
		if panicState := recover(); panicState != nil {
			fmt.Printf("ERROR: Recovered from panic state:", panicState)
		}
	}()

	id := atomic.AddUint64(&pwp.jobcnt, 1)
	j := Job {
		ID: id,
		Data: job,
	}

	pwp.jobq <- j
}
