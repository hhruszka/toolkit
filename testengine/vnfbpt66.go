package testengine

import (
	"go.uber.org/zap"
	"time"
)

func getvnfbpt66Dependencies(testId string, depExecResults map[string]map[string]*ExecutionStatus) []string {
	depID := AllTestCases[testId].Dependencies[0].Id
	varName := AllTestCases[testId].Dependencies[0].VarName
	execStatus := depExecResults[depID][varName]

	return execStatus.Stdout
}

func vnfbpt66(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	execTime := time.Now().UTC()

	depValues := getvnfbpt66Dependencies(testId, depExecResults)
	cf := func() map[string]string { return GetEnvVars(depValues) }
	ec := NewEnvVarCollector(WithCollectorFunc(cf))

	sd := NewSecretDetector(WithExcludeEnvVarFunc(isExcludedEnvVar), WithLogger(zap.L(), "VNFBPT66"))
	envVars := ec.CollectAllEnVars()
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
