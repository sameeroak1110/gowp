package gowp

import (
)

func (pwp *WorkerPool) GetID() int32 {
	if pwp == nil {
		return -1
	}

	return pwp.id
}

func (pwp *WorkerPool) GetName() string {
	if pwp == nil {
		return EMPTY_STRING
	}

	return pwp.name
}

func (pwp *WorkerPool) GetUUID() string {
	if pwp == nil {
		return EMPTY_STRING
	}

	return pwp.uuid
}
