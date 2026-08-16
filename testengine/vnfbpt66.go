package testengine

import (
	"strings"
	"time"

	"go.uber.org/zap"
)

func getvnfbpt66Dependencies(testId string, depExecResults map[string]map[string]*ExecutionStatus) []string {
	depID := AllTestCases[testId].Dependencies[0].Id
	varName := AllTestCases[testId].Dependencies[0].VarName
	execStatus := depExecResults[depID][varName]

	return execStatus.Stdout
}

// getEnvVars extracts environment variables from the execution status of a dependent test case and returns them as a map.
func getEnvVars(env []string) map[string]string {
	envVars := make(map[string]string)
	for _, str := range env {
		varName, varValue, found := strings.Cut(str, "=")
		if found {
			envVars[varName] = varValue
		}
	}

	return envVars
}

func vnfbpt66(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	execTime := time.Now().UTC()

	depValues := getvnfbpt66Dependencies(testId, depExecResults)

	sd := NewSecretDetector(WithExcludeEnvVarFunc(isExcludedEnvVar), WithLogger(zap.L(), "VNFBPT66"))
	envVars := getEnvVars(depValues)
	results := sd.DetectSecrets(envVars)

	retCode := Success
	errStr := ""
	if results.HasError {
		retCode = GeneralError
		errStr = "found secrets in environment variables\ncheck details in \"VNFBPT66 Details\" tab"
	}

	return NewExecutionStatusWithData(
		retCode,
		errStr,
		"",
		"",
		execTime,
		results.Secrets,
	)
}
