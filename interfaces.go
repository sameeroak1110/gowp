/* *****************************************************************************
Copyright (c) 2023, sameeroak1110 (sameeroak1110@gmail.com)
BSD 3-Clause License.

Package     : github.com/sameeroak1110/gowp
Filename    : github.com/sameeroak1110/gowp/interfaces.go
File-type   : GoLang source code file

Compiler/Runtime: go version go1.20.5 linux/amd64

Version History
Version     : 1.0
Author      : Sameer Oak (sameeroak1110@gmail.com)

Description :
- Interfaces that other types should implement.
***************************************************************************** */
package gowp

import (
	"context"
)

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
// - Therefore, it's upto the implementation how to handle the upstream context.
type JobProcessor interface {
	GetName() string
	Process(context.Context) (interface{}, error)
}
