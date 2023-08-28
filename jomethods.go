package gowp

import (
	"sync/atomic"
)


func (pwp *WorkerPool) NewJob(_name string, _data JobProcessor) Job {
	tmpID := atomic.AddUint64(&pwp.jobcnt, 1)
	return Job {
		id: tmpID,
		name: _name,
		data: _data,
	}
}


func (job Job) GetID() uint64 {
	return job.id
}


func (job Job) GetName() string {
	return job.name
}


func (job Job) GetData() JobProcessor {
	return job.data
}
