/*
Copyright 2015 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

*/

package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

const (
	ShellToUse            = "bash"
	CommandTimeoutSeconds = 10
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func ExecCommandLine(commandline string) (string, error) {
	return ExecCommandLineWithTimeOut(commandline, CommandTimeoutSeconds)
}

func ExecCommandLineWithTimeOut(commandline string, timeout int) (string, error) {
	var cmd *exec.Cmd

	background := false
	if strings.HasSuffix(commandline, " &") {
		background = true
		timeout = 0
	}

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
		defer cancel()

		cmd = exec.CommandContext(ctx, ShellToUse, "-c", commandline)
	} else {
		cmd = exec.Command(ShellToUse, "-c", commandline)
	}

	exitCode := 0
	var output []byte
	var err error

	if background {
		if err := cmd.Start(); err != nil {
			output = []byte("Failed to start the command")
			exitCode = 1
		}
	} else {
		if output, err = cmd.CombinedOutput(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				exitCode = exitError.ExitCode()
			}
		}
	}

	var finalErr error
	if exitCode != 0 || err != nil {
		finalErr = fmt.Errorf("Command (%v) failed: exitcode: %v, output (%v), error: %v", commandline, exitCode, string(output), err)
	} else {
		klog.V(3).Infof("Running Command (%v) succeeded", commandline)
	}

	return string(output), finalErr
}
