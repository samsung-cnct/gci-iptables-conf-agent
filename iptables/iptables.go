//
// Copyright © 2016 Samsung CNCT
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package iptables provides some a very minimal interface to Linux iptables.
package iptables

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

const (
	cmd        = "iptables"
	cmdSave    = "iptables-save"
	cmdRestore = "iptables-restore"

	argC       = "-c"
	argCounter = "--counter"
)

// Save executes the system command "iptables-save -c" and returns either
// success: the resultant byte array containing stdout, error = nil
// failure: the resultant byte array containing stderr, error is set
func Save() ([]byte, error) {

	// iptables-save with couter values (for now, until this becomes an option: arg)..
	cmd := exec.Command(cmdSave, argCounter)
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf

	if err := cmd.Run(); err != nil {
		log.Print(fmt.Sprintf("Save: cmd.Run error return: %v", err))
		return stderrBuf.Bytes(), err
	}

	return stdoutBuf.Bytes(), nil
}

// Restore executes the system command "iptables-restore < 'stdin'"
func Restore(stdin []byte) error {

	// iptables-restore with couter values (for now, until this becomes an option: arg).
	cmd := exec.Command(cmdRestore, argCounter)
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}
	cmd.Stdout = stdoutBuf
	cmd.Stderr = stderrBuf
	cmd.Stdin = bytes.NewBuffer(stdin)

	if err := cmd.Run(); err != nil {
		log.Print(fmt.Sprintf("Restore: cmd.Run error return: %v, stderr: %s, stdout: %s", err, stderrBuf, stdoutBuf))
		return err
	}
	return nil
}

// ContainsRulePart iterates over the contents of buf looking for any occurance
// of the stubstring 'match' as any portion of every individual byte array.  It
// returns the index of the first match or -1 if none found.
func ContainsRulePart(match string, buf [][]byte) int {

	matchBytes := []byte(match)
	for i := range buf {
		if bytes.Contains(buf[i], matchBytes) {
			return i
		}
	}

	return -1
}
