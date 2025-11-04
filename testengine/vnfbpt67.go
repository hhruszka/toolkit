//go:build ignore

package testengine

import (
	"time"
)

func getvnfbpt67Dependencies(testId string, depExecResults map[string]map[string]*ExecutionStatus) []string {
	depID := AllTestCases[testId].Dependencies[0].Id
	varName := AllTestCases[testId].Dependencies[0].VarName
	execStatus := depExecResults[depID][varName]

	return execStatus.Stdout
}

func vnfbpt67(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	execTime := time.Now().UTC()

	retCode := Success
	errStr := ""

	return NewExecutionStatusWithData(
		retCode,
		errStr,
		"",
		"",
		execTime,
		results.Secrets,
	)
}
