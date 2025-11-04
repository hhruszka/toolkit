package reports

import (
	"bptvnftester/testengine"
	"time"
)

func ConvertTestData(hostName string, version string, accountTestsResults []*testengine.AccountTestResults) *HostTestReport {
	return NewHostTestReport(hostName, time.Now().Format(time.RFC822), version, accountTestsResults)
}
