/* *****************************************************************************
Copyright (c) 2023, sameeroak1110 (sameeroak1110@gmail.com)
BSD 3-Clause License.

Package     : github.com/sameeroak1110/gowp/helper
Filename    : github.com/sameeroak1110/gowp/helper/helper.go
File-type   : GoLang source code file

Compiler/Runtime: go version go1.20.5 linux/amd64 when tested lately.

Version History
Version     : 1.0
Author      : Sameer Oak (sameeroak1110@gmail.com)

Description :
- Helper functions.
***************************************************************************** */
package helper

import (
	"fmt"
	"math/rand"
)


/* *****************************************************************************
Description : Returns random integer within min and max bounaries, both included.

Arguments   :
1> min int: Lower boundary.
2> max int: Upper boundary.

Return value:
1> int: Random number between lower and upper boundary.

Additional note: NA
***************************************************************************** */
func RandomInt(min int, max int) int {
	if min >= max {
		return -1
	}

	return rand.Intn(max - min + 1) + min
}


/* *****************************************************************************
Description : Generated a new 40 character (UTF-8) wide UUID.

Arguments   : NA

Return value:
1> string: Newly generated UUID.
2> error: Error in case of error.

Additional note: NA
***************************************************************************** */
func NewUUID() (string, error) {
	uuidBuffer := make([]byte, uuidLen)
	if _, err := rand.Read(uuidBuffer); err != nil {
		return "", fmt.Errorf("ERROR: Reading random string: %s", err.Error())
	}

	if (uuidBuffer == nil) || (len(uuidBuffer) < 1) {
		return "", fmt.Errorf("ERROR: Nil or empty uuidBuffer.")
	}

	uuid := fmt.Sprintf("%x%x%x%x%x", uuidBuffer[0:6], uuidBuffer[6:8], uuidBuffer[8:10], uuidBuffer[10:12], uuidBuffer[12:])
	fmt.Printf("len(uuidBuffer): %d,  newuuid: %s\n", len(uuidBuffer), uuid)

	return uuid, nil
}
