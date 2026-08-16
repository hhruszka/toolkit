//go:build linux

package testengine

import (
	"bptvnftester/utils"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var (
	procs    map[int]*utils.ProcessInfo
	interval int = 5   // seconds
	probing  int = 100 // miliseconds
)

// var reIsFile = regexp.MustCompile(`(\$|\.\/|\.\.\/|\~|\/)+([^\/\n\r\t,:;= ]+\/)*[^\/\n\r\t,:;= ]+`)
var reIsFile = regexp.MustCompile(`(\$|\./|\.\./|~|/)+([^/\n\r\t,:;= ]+/)*[^/\n\r\t,:;= ]+`)

// enum05 tests for: known exploitable binaries with setuid cannot be present on the system.
func enum05(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	var (
		err     error
		retCode = Success
	)

	execTime := time.Now().UTC()
	procs, err = utils.ProcessMonitor(interval, probing)
	if err != nil {
		return NewExecutionStatus(InternalAppError, err.Error(), "", "", execTime)
	}

	return NewExecutionStatus(retCode, "", "", "", execTime)
}

// vnfbpt44 checks if processes run by root are not writable by other users
func vnfbpt44(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	if procs == nil {
		return nil
	}

	var (
		testError []string
	)

	for pid, proc := range procs {
		if proc.UID == 0 && proc.CmdLineCmd != "" && utils.IsWritable(proc.CmdLineCmd) {
			// TODO: do it better
			testError = append(testError, fmt.Sprintf("root's process %d executable %s is writable\n%s\t%s", pid, proc.Cmd, proc.ProcFileInfo.Mode().Perm().String(), utils.FileAbs(proc.CmdLineCmd)))
		}
	}

	if len(testError) > 0 {
		retCode = GeneralError
	}

	return NewExecutionStatus(retCode, strings.Join(testError, "\n"), "", "", execTime)
}

// vnfbpt45 checks if processes run by root are not invoked with files (as arguments) readable by other users
func vnfbpt45(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	if procs == nil {
		return nil
	}

	var (
		testError []string
	)

	for pid, proc := range procs {
		if proc.UID == 0 && proc.CmdLineCmd != "" {
			for _, filePath := range reIsFile.FindAllString(proc.CmdLineArgs, -1) {
				if utils.DoesFileExists(filePath) && utils.IsRegularFile(filePath) && utils.IsReadable(filePath) {
					testError = append(testError, fmt.Sprintf("file %s in a command line of a process %s (pid %d) run by root is readable", filePath, proc.CmdLineCmd, pid))
					owner, group := utils.FileOwnership(filePath)
					perms := utils.FilePermissions(filePath)
					testError = append(testError, fmt.Sprintf("\t%s:%s\t%s\t%s", owner, group, perms, filePath))
				}
			}
		}
	}

	if len(testError) > 0 {
		retCode = GeneralError
	}

	return NewExecutionStatus(retCode, strings.Join(testError, "\n"), "", "", execTime)
}

// vnfbpt46 checks if processes run by root are not invoked with files writable by other users
func vnfbpt46(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	if procs == nil {
		return nil
	}

	var (
		testError []string
	)

	//fmt.Printf("vnfbpt46() Process Monitor found %d processes\n", len(procs))
	for pid, proc := range procs {
		if proc.UID == 0 && proc.CmdLineCmd != "" {
			for _, filePath := range reIsFile.FindAllString(proc.CmdLineArgs, -1) {
				//fmt.Printf("\t%s\n", match)
				if utils.DoesFileExists(filePath) && utils.IsRegularFile(filePath) && utils.IsWritable(filePath) {
					testError = append(testError, fmt.Sprintf("file %s in a command line of a process %s (pid %d) run by root is writable", filePath, proc.CmdLineCmd, pid))
					owner, group := utils.FileOwnership(filePath)
					perms := utils.FilePermissions(filePath)
					testError = append(testError, fmt.Sprintf("\t%s:%s\t%s\t%s", owner, group, perms, filePath))
				}
			}
		}
	}

	if len(testError) > 0 {
		retCode = GeneralError
	}

	return NewExecutionStatus(retCode, strings.Join(testError, "\n"), "", "", execTime)
}

// vnfbpt47 checks if processes are not invoked with files containing secrets e.g. credentials
func vnfbpt47(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	if procs == nil {
		return nil
	}

	var (
		testError []string
	)

	//fmt.Printf("vnfbpt47() Process Monitor found %d processes\n", len(procs))
	for pid, proc := range procs {
		if proc.CmdLineCmd != "" {
			for _, filePath := range reIsFile.FindAllString(proc.CmdLineArgs, -1) {
				//fmt.Printf("\t%s\n", filePath)
				if utils.DoesFileExists(filePath) && utils.IsRegularFile(filePath) && utils.IsReadable(filePath) && utils.IsPlainTextFile(filePath) {
					results := ScanFileWithRegex(filePath)
					if results != nil {
						// TODO: do it better
						testError = append(testError, fmt.Sprintf("File %s in a command line of a process %s (pid %d) contains secrets:", filePath, proc.CmdLineCmd, pid))
						for _, secret := range results.secrets {
							testError = append(testError, fmt.Sprintf("\t%s: %s", secret.SecretType, secret.SecretValue))
						}
					}
				}
			}
		}
	}

	if len(testError) > 0 {
		retCode = GeneralError
	}

	return NewExecutionStatus(retCode, strings.Join(testError, "\n"), "", "", execTime)
}

// vnfbpt48 checks if processes are not invoked with secrets/credentials
func vnfbpt48(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	if procs == nil {
		return nil
	}

	var (
		testError []string
	)

	//fmt.Printf("vnfbpt48() Process Monitor found %d processes\n", len(procs))
	for pid, proc := range procs {
		if proc.CmdLine != "" {
			results := ScanWithRegex(proc.CmdLine)
			if results != nil {
				// TODO: do it better
				testError = append(testError, fmt.Sprintf("Command line of process %s (pid %d) contains secrets", proc.CmdLineCmd, pid))
				testError = append(testError, fmt.Sprintf("\tCommand line:\n%s", proc.CmdLine))
				testError = append(testError, fmt.Sprintf("\tSecrets:"))
				for _, secret := range results {
					testError = append(testError, fmt.Sprintf("\t\t%s: %s", secret.pattern.Name, strings.Join(secret.matches, ", ")))
				}
			}
		}
	}

	if len(testError) > 0 {
		retCode = GeneralError
	}
	return NewExecutionStatus(retCode, strings.Join(testError, "\n"), "", "", execTime)
}
