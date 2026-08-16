package testengine

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"os/user"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
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
		userName = userInfo.Username
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

// getTestIds returns a sorted list of test IDs based on the input slice or all available test case IDs if input is nil.
func getTestIds(tests []string) []string {
	if tests == nil || len(tests) == 0 {
		testIds := slices.Collect(maps.Keys(AllTestCases))
		slices.Sort(testIds)
		return testIds
	}

	initTestCases.Do(func() {
		selectedTestCases = make(map[string]*TestCase)
		for _, testId := range tests {
			selectedTestCases[testId] = AllTestCases[testId]
		}
		selectedTestIds = slices.Collect(maps.Keys(selectedTestCases))
		slices.Sort(selectedTestIds)
	})

	return selectedTestIds
}

// IsTestCase checks if the provided testId exists in the AllTestCases map and returns true if it exists, otherwise false.
func IsTestCase(testId string) bool {
	_, exists := AllTestCases[testId]
	return exists
}

// GetAbstract retrieves the abstract description of a test case using its test ID as a key from the AllTestCases map.
func GetAbstract(testId string) string {
	return AllTestCases[testId].Abstract
}

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err) // crypto/rand failing is catastrophic; handle as you see fit
	}
	return hex.EncodeToString(b)
}

func trimMarker(out, marker string) string {
	lines := strings.Split(out, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == marker {
			return strings.Join(lines[i+1:], "\n")
		}
	}
	return out // marker not found: return raw so we don't silently drop data
}

// CmdExec executes command provided in cmdStdin with environment cmdEnv and timeout provided in seconds.
func CmdExec(ctx context.Context, cmdStdin string, cmdExec string, cmdArgs []string, cmdEnv []string, timeout time.Duration) *ExecutionStatus {
	var (
		retcode RetCode
		cmd     *exec.Cmd
		cmdOut  strings.Builder
		cmdErr  strings.Builder
		err     error
	)

	cmdExecCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		cmdExecCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	//zap.L().Debug("Executing command", zap.String("stdCmdin", cmdStdin), zap.String("cmdExec", cmdExec), zap.Strings("cmdArgs", cmdArgs))

	marker := "---sentinel-" + randHex(16) + "---"
	wrappedCmd := fmt.Sprintf("echo %s;%s", marker, cmdStdin)
	cmdArgs = append(cmdArgs, wrappedCmd)

	zap.L().Debug("Executing command", zap.String("cmdExec", cmdExec), zap.Strings("cmdArgs", cmdArgs))

	cmd = exec.CommandContext(cmdExecCtx, cmdExec, cmdArgs...)
	cmd.Env = cmdEnv
	cmd.Stdin = bytes.NewReader(nil)
	cmd.Stdout = &cmdOut
	cmd.Stderr = &cmdErr
	execTime := time.Now().UTC()

	err = cmd.Run()
	cmdOutStr := trimMarker(cmdOut.String(), marker)
	retcode = Success
	errMsg := ""

	var exitErr *exec.ExitError
	switch {
	case errors.Is(cmdExecCtx.Err(), context.DeadlineExceeded):
		retcode = ExecutionTimeOut
		errMsg = cmdExecCtx.Err().Error()
	case errors.Is(cmdExecCtx.Err(), context.Canceled):
		retcode = ExecutionCanceled
		errMsg = cmdExecCtx.Err().Error()
	case err != nil && errors.As(err, &exitErr):
		retcode = RetCode(exitErr.ExitCode())
		errMsg = err.Error()
	case err != nil:
		retcode = InternalAppError
		errMsg = err.Error()
	}

	return NewExecutionStatus(retcode, errMsg, cmdOutStr, cmdErr.String(), execTime)
}

// ExecTest executes a test case on a container.
func ExecTest(ctx context.Context, execTest func(context.Context, string, string, []string, []string, time.Duration) *ExecutionStatus, testId string, cache map[string]*ExecutionStatus, cmdExec string, cmdArgs []string, timeout time.Duration) *ExecutionStatus {
	// check if cached - was already executed for a given pod and container
	if _, cached := cache[testId]; cached {
		return cache[testId]
	}

	zap.L().Debug("Execution", zap.String("test id", testId))

	// execute dependencies
	var cmdEnv []string
	var depExecStatuses map[string]map[string]*ExecutionStatus

	for _, dep := range AllTestCases[testId].Dependencies {
		// This is recurrent call. Exec() always returns an instance of k8sexec.ExecutionStatus
		depExecResult := ExecTest(ctx, execTest, dep.Id, cache, cmdExec, cmdArgs, timeout)
		if dep.Type == Stdout {
			// note "%q"
			//fmt.Printf("%s=%q\n", dep.VarName, strings.Join(depExecResult.Stdout, " "))
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%q", dep.VarName, strings.Join(depExecResult.Stdout, " ")))
		}
		if dep.Type == ExitCode {
			cmdEnv = append(cmdEnv, fmt.Sprintf("%s=%t", dep.VarName, depExecResult.RetCode == Success))
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
		zap.L().Debug(testId, zap.String("host", hostName), zap.String("user", userName))
	} else {
		if len(AllTestCases[testId].Command) == 0 {
			// this is an error condition since neither TestFunc nor Command have been defined for a test case
			panic(fmt.Sprintf("Internal error: for test id %s neither TestFunc nor Command have been defined.\nAborting. \n", testId))
		}

		// add command to stdin buffer

		cache[testId] = execTest(ctx, AllTestCases[testId].Command, cmdExec, cmdArgs, cmdEnv, timeout)
		zap.L().Debug(
			testId,
			zap.String("host", hostName),
			zap.String("user", userName),
			zap.String("command", AllTestCases[testId].Command),
			zap.String("cmdExec", cmdExec),
			zap.String("cmdArgs", strings.Join(cmdArgs, " ")),
			zap.String("cmdEnv", strings.Join(cmdEnv, " ")),
		)
	}

	zap.L().Debug("cache", zap.String("test id", testId), zap.Any("cache", cache[testId]))
	return cache[testId]
}

// ExecTests function create test environment for a container and executes test cases.
func ExecTests(ctx context.Context, tests []string, cmdExec string, cmdArgs []string, timeout time.Duration) map[string]*ExecutionStatus {
	var cache map[string]*ExecutionStatus = make(map[string]*ExecutionStatus)
	var execStatuses map[string]*ExecutionStatus = make(map[string]*ExecutionStatus)

	// since test cases are kept in a slice, we need to create it with test cases from tests parameter if tests != nil
	testIds := getTestIds(tests)

	for _, testId := range testIds {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}

		if AllTestCases[testId].IsTest {
			fmt.Fprintf(os.Stderr, "Executing: %s  %q\n", AllTestCases[testId].Id, AllTestCases[testId].Abstract)
			execStatuses[testId] = ExecTest(ctx, CmdExec, testId, cache, cmdExec, cmdArgs, timeout)
		}
	}

	return execStatuses
}

type AccountTestResults struct {
	UserName         string                 `json:"UserName"`
	ExecTestStatuses map[string]*TestResult `json:"ExecTestStatuses"`
}

// NewAccountTestResults creates a new AccountTestResults instance with the provided username and execution test statuses.
func NewAccountTestResults(userName string, execTestStatuses map[string]*TestResult) *AccountTestResults {
	return &AccountTestResults{UserName: userName, ExecTestStatuses: execTestStatuses}
}

type TestResult struct {
	Result     bool             `json:"Status"`
	ExecStatus *ExecutionStatus `json:"Details"`
}

// NewTestResult creates a new TestResult instance with the specified test status and execution details.
func NewTestResult(status bool, execStatus *ExecutionStatus) *TestResult {
	return &TestResult{Result: status, ExecStatus: execStatus}
}

// TimedOut checks if the execution status indicates a timeout by comparing the return code to ExecutionTimeOut.
func (tr *TestResult) TimedOut() bool {
	return tr.ExecStatus.RetCode == ExecutionTimeOut
}

// Canceled checks if the execution status indicates a cancel by comparing the return code to ExecutionCanceled.
func (tr *TestResult) Canceled() bool {
	return tr.ExecStatus.RetCode == ExecutionCanceled
}

// RunTests executes a series of test cases, processes their results, and returns a map of test IDs to TestResult.
func RunTests(ctx context.Context, tests []string, cmdExec string, cmdArgs []string, timeout time.Duration) map[string]*TestResult {
	testsResults := make(map[string]*TestResult)
	execStatuses := ExecTests(ctx, tests, cmdExec, cmdArgs, timeout)

	for testId, status := range execStatuses {
		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}

		if AllTestCases[testId].IsTest {
			result := AllTestCases[testId].ResultFunc(status)
			testsResults[testId] = NewTestResult(result, status)
		}
	}

	return testsResults
}

// isParentSudo checks if the parent process is 'sudo' by examining process information in the /proc filesystem on Linux.
func isParentSudo() bool {
	ppid := os.Getppid()

	// Method 1: Check /proc/PPID/comm (Linux)
	if commBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid)); err == nil {
		comm := strings.TrimSpace(string(commBytes))
		if comm == "sudo" {
			return true
		}
	}

	// Method 2: Check /proc/PPID/cmdline (Linux)
	if cmdlineBytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", ppid)); err == nil {
		cmdline := string(cmdlineBytes)
		// cmdline has null separators, split on first null
		if parts := strings.Split(cmdline, "\x00"); len(parts) > 0 {
			if strings.HasSuffix(parts[0], "sudo") {
				return true
			}
		}
	}

	return false
}

var RunAccountTests = RunAccountTestsWithSU

// findSU locates a system `su` binary at a well-known path, bypassing PATH-based
// lookup which may resolve to broken vendor wrappers (e.g. Nokia LSS's /opt/LSS/bin/su
// which is a bash script without a shebang line).
func findSU() (string, error) {
	candidates := []string{"/bin/su", "/usr/bin/su"}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && !info.IsDir() && info.Mode()&0111 != 0 {
			return p, nil
		}
	}
	// Last-resort fallback to PATH lookup — may hit vendor wrappers.
	return exec.LookPath("su")
}

// RunAccountTestsWithSU executes tests for multiple user accounts through `su`, ensuring environment parity for each test.
// It iterates over provided user credentials, prints the currently tested account, and captures the results of each test.
// Returns a slice of AccountTestResults containing the outcomes of the executed tests per user.
func RunAccountTestsWithSU(ctx context.Context, tests []string, users map[string]string, timeout time.Duration) []*AccountTestResults {
	var userTestsResults []*AccountTestResults

	suPath, err := findSU()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: cannot locate 'su' binary: %v\n", err)
		return nil
	}
	if suPath != "/bin/su" && suPath != "/usr/bin/su" {
		fmt.Fprintf(os.Stderr, "WARNING: using non-standard 'su' at %q\n", suPath)
	}

	for login, shell := range users {
		fmt.Fprintf(os.Stderr, "Testing account: %s\n", login)
		_ = shell

		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}
		// OPTION 2: su is independent of linux distribution, however, bptvnftester should be run from root shell.
		// `shell -i -c bash` forces shell to run in the interactive mode, which ensures that the shell environment looks in the similar way when
		// a threat actor got initial footprint on the system (shell). This behaviour of a shell was confirmed with the shell man pages. Usage of `-i`
		// is unconditional - does not conflict with `-c`. The shell in the interactive mode sources rc files (.bashrc, .zshrc etc.) that contains
		// user account customizations e.g. additional env. variables or aliases.
		// Execution of `bash` as a command/argument of `-c` makes `bash` to inherit the environment created by `shell -i` since `bash` is launched
		// as a child process.
		//userTestsResults = append(userTestsResults, NewAccountTestResults(login, RunTests(ctx, tests, "su", []string{"-l", login, "--session-command", fmt.Sprintf("%s -i -c sh", shell)}, timeout)))
		userTestsResults = append(userTestsResults, NewAccountTestResults(login, RunTests(ctx, tests, suPath, []string{"-l", login, "--session-command"}, timeout)))
	}
	return userTestsResults
}

// RunAccountTestsWithSUDO executes a set of test cases for user accounts using `sudo` and returns the results.
func RunAccountTestsWithSUDO(ctx context.Context, tests []string, users map[string]string, timeout time.Duration) []*AccountTestResults {
	var userTestsResults []*AccountTestResults

	for login, shell := range users {
		fmt.Fprintf(os.Stderr, "Testing account: %s\n", login)
		_ = shell

		if ctx != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
			}
		}

		// OPTION 1: using sudo to execute commands works as expected, but it depends on availability of sudo
		// `shell -i -c bash` forces shell to run in the interactive mode, which ensures that the shell environment looks in the similar way when
		// a threat actor got initial footprint on the system (shell). This behaviour of a shell was confirmed with the shell man pages. Usage of `-i`
		// is unconditional - does not conflict with `-c`. The shell in the interactive mode sources rc files (.bashrc, .zshrc etc.) that contains
		// user account customizations e.g. additional env. variables or aliases.
		// Execution of `bash` as a command/argument of `-c` makes `bash` to inherit the environment created by `shell -i` since `bash` is launched
		// as a child process.
		userTestsResults = append(userTestsResults, NewAccountTestResults(login, RunTests(ctx, tests, "sudo", []string{"-iu", login, shell, "-i", "-c", "bash"}, timeout)))
	}
	return userTestsResults
}
