package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type ExecutionStatus struct {
	ExitStatus int
	Error      error
	Stdout     string
	Stderr     string
}

// App global variables
var (
	hostname string
)

func NewExecutionStatus(exitStatus int, error error, stdout string, stderr string) *ExecutionStatus {
	return &ExecutionStatus{ExitStatus: exitStatus, Error: error, Stdout: stdout, Stderr: stderr}
}

func cmdexecv1(cmdstr string) *ExecutionStatus {
	var (
		retcode int
		cmdOut  strings.Builder
		cmdErr  strings.Builder
		err     error
	)

	cmd := exec.Command("sh")
	cmd.Stdin = strings.NewReader(cmdstr)

	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr

	if err = cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			// The program has exited with an exit code != 0
			// Attempt to extract the exit code if possible
			retcode = exitErr.ExitCode()
		} else {
			retcode = -1
		}
	}

	return NewExecutionStatus(retcode, err, cmdOut.String(), cmdErr.String())
}

func cmdexec(cmdstr string) {
	// Command to execute

	cmd := exec.Command("sh", "-c", cmdstr)

	// Buffers to capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Running the command
	err := cmd.Run()

	// Printing stdout and stderr regardless of the command's success
	//fmt.Printf("stdout:\n%s\n", stdoutBuf.String())
	//fmt.Printf("stderr:\n%s\n", stderrBuf.String())

	// Handling errors and obtaining the exit code
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// The program has exited with an exit code != 0

			// Attempt to extract the exit code if possible
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				fmt.Printf("exit status: %d\n", status.ExitStatus())
			}
		} else {
			fmt.Printf("cmd.Run() failed with %s\n", err)
		}
	} else {
		// Command was successful
		fmt.Println("exit status: 0")
	}
}

func init() {
	var err error
	hostname, err = os.Hostname()

	if err != nil {
		log.Fatal(err.Error())
	}
}

func result(res bool) string {
	return map[bool]string{true: "PASSED", false: "FAILED"}[res]
}

func main() {
	var buf bytes.Buffer

	t := table.NewWriter()
	t.SetOutputMirror(&buf)
	t.SetTitle("Host: %s", hostname)
	t.Render()

	for _, tc := range TestCases {
		exitStatus := cmdexecv1(tc.Command)
		if tc.IsTest {
			t := table.NewWriter()
			t.SetOutputMirror(&buf)
			t.AppendRows([]table.Row{{tc.Id, tc.Abstract, result(tc.ResultFunc(exitStatus))}})
			t.AppendSeparator()
			t.Render()
			t = table.NewWriter()
			t.SetOutputMirror(&buf)
			t.AppendRow(table.Row{"Exit status:", exitStatus.ExitStatus})
			t.AppendSeparator()
			t.AppendRow(table.Row{"Stdout:", exitStatus.Stdout})
			t.AppendSeparator()
			t.AppendRow(table.Row{"Stderr:", exitStatus.Stderr})
			t.AppendSeparator()
			t.Render()
		}
	}
	fmt.Println(buf.String())
}
