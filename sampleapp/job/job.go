package job

import (
	"time"
	"context"

	"github.com/sameeroak1110/logger"
	"github.com/sameeroak1110/gowp/helper"
)

const pkgname string = "job"

type TestJobData struct {
	ID int64
	Name string
	Desc string
}

func (job TestJobData) GetName() string {
	return job.Name
}

func (job TestJobData) Process(ctx context.Context, cancelFunc context.CancelFunc, cnt int, shouldTerminate bool) (interface{}, error) {
	defer func() {
		if panicState := recover(); panicState != nil {
			logger.Log(pkgname, logger.ERROR, "Recovered from panic. state: %#v", panicState)
		} else {
			logger.Log(pkgname, logger.DEBUG, "job done: %d:%s", job.ID, job.Name)
		}
	}()

	select {
		case val := <-ctx.Done():
			logger.Log(pkgname, logger.WARNING, "[%d] ProcessJob ctx.Cancel: %s,  value: %#v\n", job.ID, job.Name, val)
			return nil, nil

		default:
			logger.Log(pkgname, logger.DEBUG, "TestJobData process(%d:%s) starts.", job.ID, job.Name)
			//execForMS := helper.RandomInt(5000, 10000)
			//execForMS := helper.RandomInt(15000, 20000)
			//execForMS := helper.RandomInt(5000, 8000)
			execForMS := helper.RandomInt(3000, 5000)
			time.Sleep(time.Duration(execForMS) * time.Millisecond)
			logger.Log(pkgname, logger.DEBUG, "TestJobData process(%d:%s) executed for %d ms", job.ID, job.Name, execForMS)
			if shouldTerminate == true {
				if job.ID == int64(cnt) {
					logger.Log(pkgname, logger.DEBUG, "All done. Invoking termination.")
					cancelFunc()
				}
			}
	}

	return nil, nil
}
