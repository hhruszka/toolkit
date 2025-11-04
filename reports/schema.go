package reports

import (
	"bptvnftester/testengine"
)

const REPORT_NAME = "vnf-test-report"

// TestResult represents the outcome of an individual test execution, including metadata and execution details.
type HostTestReport struct {
	HostName string
	Date     string
	Version  string
	Tests    []*testengine.AccountTestResults
}

// NewTestResult creates a new instance of TestResult by mapping test details and execution status from TestResult.
func NewHostTestReport(hostName string, date string, version string, results []*testengine.AccountTestResults) *HostTestReport {
	return &HostTestReport{
		HostName: hostName,
		Date:     date,
		Version:  version,
		Tests:    results,
	}
}
