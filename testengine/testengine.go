package testengine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"linuxtester/log"
	"maps"
	"os"
	"os/exec"
	"os/user"
	"slices"
	sort2 "sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var AllTestCases map[string]*TestCase
var selectedTestCases map[string]*TestCase
var selectedTestIds []string

var userName string
var hostName string

func init() {
	AllTestCases = make(map[string]*TestCase)

	for _, testCase := range TestCases {
		AllTestCases[testCase.Id] = testCase
	}

	userName = strconv.Itoa(os.Getuid())
	userInfo, err := user.Current()
	if err == nil {
		userName = userInfo.Name
	}

	hostName, _ = os.Hostname()
}

// initTestCases is a mutex for defining selectedTestCases global variable which is indirectly
// shared by all testing goroutines.
var initTestCases sync.Once

// getTestCases returns the slice with test cases based on a parameter tests.
// If tests is nil then all test cases are returned.
func getTestCases(tests []string) map[string]*TestCase {
	if tests == nil {
		return AllTestCases
	}

	initTestCases.Do(func() {
		selectedTestCases = make(map[string]*TestCase)
		for _, testId := range tests {
			selectedTestCases[testId] = AllTestCases[testId]
		}
	})

	return selectedTestCases
}

// getTestCases returns the slice with test cases based on a parameter tests.
// If tests is nil then all test cases are returned.
func getTestIds(tests []string) []string {
	if tests == nil {
		testIds := slices.Collect(maps.Keys(AllTestCases))
		sort2.Strings(testIds)
		return testIds
	}

	initTestCases.Do(func() {
		selectedTestCases = make(map[string]*TestCase)
		for _, testId := range tests {
			selectedTestCases[testId] = AllTestCases[testId]
		}
		selectedTestIds = slices.Collect(maps.Keys(selectedTestCases))
		sort2.Strings(selectedTestIds)
	})

	return selectedTestIds
}

func IsTestCase(testId string) bool {
	_, exists := AllTestCases[testId]
	return exists
}

// CmdExec executes command provided in cmdStdin with environment cmdEnv and timeout provided in seconds.
func CmdExec(cmdStdin io.Reader, cmdExec string, cmdArgs []string, cmdEnv []string, timeout time.Duration) *ExecutionStatus {
	var (
		retcode RetCode
		cmd     *exec.Cmd
		cmdOut  strings.Builder
		cmdErr  strings.Builder
		err     error
	)

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, cmdExec, cmdArgs...)
	} else {
		cmd = exec.Command(cmdExec, cmdArgs...)
	}

	cmd.Env = append(os.Environ(), cmdEnv...)
	cmd.Stdin = cmdStdin
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr

	if err = cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if ok := errors.As(err, &exitErr); ok {
			// The program has exited with an exit code != 0
			// Attempt to extract the exit code if possible
			retcode = RetCode(exitErr.ExitCode())
		} else if errors.Is(err, context.DeadlineExceeded) {
			retcode = ExecutionTimeOut
		} else {
			retcode = InternalAppError
		}
		return NewExecutionStatus(retcode, err.Error(), cmdOut.String(), cmdErr.String())
	}

	return NewExecutionStatus(retcode, "", cmdOut.String(), cmdErr.String())
}

// ExecTest executes a test case on a container.
func ExecTest(execTest func(io.Reader, string, []string, []string, time.Duration) *ExecutionStatus, testId string, cache map[string]*ExecutionStatus, cmdExec string, cmdArgs []string, timeout time.Duration) *ExecutionStatus {
	// check if cached - was already executed for a given pod and container
	if _, cached := cache[testId]; cached {
		return cache[testId]
	}

	log.Logf("Executing: %s\n", testId)

	// execute dependencies
	var cmdEnv []string
	var depExecStatuses map[string]map[string]*ExecutionStatus

	for _, dep := range AllTestCases[testId].Dependencies {
		// This is recurrent call. Exec() always returns an instance of k8sexec.ExecutionStatus
		depExecResult := ExecTest(execTest, dep.Id, cache, cmdExec, cmdArgs, timeout)
		if dep.Type == Stdout {
			// note "%q"
			//fmt.Printf("%s=%q\n", dep.VarName, strings.Join(depExecResult.Stdout, " "))
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%q\n", dep.VarName, strings.Join(depExecResult.Stdout, " ")))
		}
		if dep.Type == ExitCode {
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%t\n", dep.VarName, depExecResult.RetCode == Success))
		}
		if dep.Type == TestFunc {
			if depExecStatuses == nil {
				depExecStatuses = make(map[string]map[string]*ExecutionStatus)
			}
			depExecStatuses[dep.Id] = make(map[string]*ExecutionStatus)
			depExecStatuses[dep.Id][dep.VarName] = depExecResult
		}
	}

	if AllTestCases[testId].TestFunc != nil {
		cache[testId] = AllTestCases[testId].TestFunc(testId, depExecStatuses)
		log.Logf("%s/%s: %s CMD:\n%s\n", hostName, userName, testId, "")
	} else {
		if len(AllTestCases[testId].Command) == 0 {
			// this is an error condition since neither TestFunc nor Command have been defined for a test case
			panic(fmt.Sprintf("Internal error: for test id %s neither TestFunc nor Command have been defined.\nAborting. \n", testId))
		}

		// add command to stdin buffer

		cache[testId] = execTest(bytes.NewBufferString(AllTestCases[testId].Command), cmdExec, cmdArgs, cmdEnv, timeout)
	}

	log.Logf("%s/%s: %s RETCODE: %d\n", hostName, userName, testId, cache[testId].RetCode)
	log.Logf("%s/%s: %s STDOUT:\n%s\n", hostName, userName, testId, strings.Join(cache[testId].Stdout, "\n"))
	log.Logf("%s/%s: %s STDERR:\n%s\n", hostName, userName, testId, strings.Join(cache[testId].Stderr, "\n"))

	return cache[testId]
}

// ExecTests function create test environment for a container and executes test cases.
func ExecTests(tests []string, cmdExec string, cmdArgs []string, timeout time.Duration) map[string]*ExecutionStatus {
	var cache map[string]*ExecutionStatus = make(map[string]*ExecutionStatus)
	var execStatuses map[string]*ExecutionStatus = make(map[string]*ExecutionStatus)

	// since test cases are kept in a slice, we need to create it with test cases from tests parameter if tests != nil
	testIds := getTestIds(tests)

	for _, testId := range testIds {
		if AllTestCases[testId].IsTest {
			fmt.Fprintf(os.Stderr, "Executing: %s  %q\n", AllTestCases[testId].Id, AllTestCases[testId].Abstract)
			execStatuses[testId] = ExecTest(CmdExec, testId, cache, cmdExec, cmdArgs, timeout)
		}
	}

	return execStatuses
}

type AccountTestStatus struct {
	UserName         string                `json:"UserName"`
	ExecTestStatuses map[string]TestStatus `json:"ExecTestStatuses"`
}

func NewAccountTestStatus(userName string, execTestStatuses map[string]TestStatus) *AccountTestStatus {
	return &AccountTestStatus{UserName: userName, ExecTestStatuses: execTestStatuses}
}

type TestStatus struct {
	Status     bool             `json:"Status"`
	ExecStatus *ExecutionStatus `json:"Details"`
}

func RunTests(tests []string, cmdExec string, cmdArgs []string, timeout time.Duration) map[string]TestStatus {
	testsResults := make(map[string]TestStatus)
	execStatuses := ExecTests(tests, cmdExec, cmdArgs, timeout)

	for testId, status := range execStatuses {
		if AllTestCases[testId].IsTest {
			result := AllTestCases[testId].ResultFunc(status)
			testsResults[testId] = TestStatus{result, status}
		}
	}

	return testsResults
}

func RunAccountTests(tests []string, users []string, timeout time.Duration) []*AccountTestStatus {
	var userTestsResults []*AccountTestStatus
	for _, user := range users {
		fmt.Fprintf(os.Stderr, "Testing account: %s\n", user)
		userTestsResults = append(userTestsResults, NewAccountTestStatus(user, RunTests(tests, "su", []string{"-c", "sh", "-", user}, timeout)))
	}
	return userTestsResults
}
