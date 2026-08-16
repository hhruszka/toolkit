package reports

import (
	"bptvnftester/testengine"

	"github.com/fatih/color"
)

func resultToString(res bool) string {
	return map[bool]string{true: color.GreenString("PASSED"), false: color.RedString("FAILED")}[res]
}

// testStatus determines the test resultToString based on execution status or boolean value, returning "TIMEOUT", "PASSED", or "FAILED".
func testResult(test *testengine.TestResult, colorFlg bool) string {
	var (
		passedFailed map[bool]string = map[bool]string{true: "PASSED", false: "FAILED"}
		timeout      string          = "TIMEOUT"
		canceled     string          = "CANCELED"
	)

	if colorFlg {
		passedFailed[false] = color.RedString(passedFailed[false])
		passedFailed[true] = color.GreenString(passedFailed[true])
		timeout = color.YellowString("TIMEOUT")
		canceled = color.RedString("CANCELED")
	}

	result := passedFailed[test.Result]
	if test.TimedOut() {
		result = timeout
	} else if test.Canceled() {
		result = canceled
	}

	return result
}
