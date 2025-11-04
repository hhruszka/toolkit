package testengine

import (
	"bptvnftester/utils"
	"fmt"
	"strings"
	"time"
)

// vnfbpt06 tests for: known exploitable binaries with setuid cannot be present on the system.
func vnfbpt06(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	retCode := Success
	execTime := time.Now().UTC()

	depID := AllTestCases[testId].Dependencies[0].Id
	varName := AllTestCases[testId].Dependencies[0].VarName
	execStatus := depExecResults[depID][varName]

	//fmt.Println(execStatus.Stdout)

	var (
		execError string
		stdout    string
		stderr    string
	)

	if execStatus.RetCode == 0 || execStatus.RetCode == 1 {
		var exploitableBins []string
		for _, file := range execStatus.Stdout {
			if utils.IsFileListed(file, ExploitableSuidBinaries) {
				exploitableBins = append(exploitableBins, fmt.Sprintf("%s\t%s", utils.FilePermissions(file), file))
			}
		}
		stdout = strings.Join(exploitableBins, "\n")
	} else {
		retCode = execStatus.RetCode
		execError = strings.Join(execStatus.Error, "\n")
		stderr = strings.Join(execStatus.Stderr, "\n")
	}

	return NewExecutionStatus(retCode, execError, stdout, stderr, execTime)
}
