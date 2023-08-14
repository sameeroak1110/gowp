/* *****************************************************************************
Copyright (c) 2023, sameeroak1110 (sameeroak1110@gmail.com)
BSD 3-Clause License.

Package     : github.com/sameeroak1110/gowp
Filename    : github.com/sameeroak1110/gowp/data.go
File-type   : GoLang source code file

Compiler/Runtime: go version go1.20.5 linux/amd64 when tested lately.

Version History
Version     : 1.0
Author      : Sameer Oak (sameeroak1110@gmail.com)

Description :
- All internal and exported data.
***************************************************************************** */
package gowp

const pkgname string = "gowp"

// specific jobID. updated using atomic.AddInt32().
var newPoolID int32

// keeps track of how many workers are in action at any given instance in time.
// updated using atomic.AddInt32().
var workercnt int32

const minWPSize int32 = 10
const maxWPSize int32 = 100
const jpwpfactor int32 = 100
