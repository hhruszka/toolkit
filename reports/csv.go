package reports

import (
	"bptvnftester/testengine"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func SaveToCSVFile(hostTestReport *HostTestReport, filePath string) error {
	var report bytes.Buffer

	_, _ = fmt.Fprintf(&report, "Host,User,Test Id,Abstract, Result,Execution Code, Stdout, Stderr\n")

	for _, testsResults := range hostTestReport.Tests {
		var sortedTestIds []string

		for testId, _ := range testsResults.ExecTestStatuses {
			sortedTestIds = append(sortedTestIds, testId)
		}
		sort.Strings(sortedTestIds)

		for _, testId := range sortedTestIds {
			if testengine.AllTestCases[testId].IsTest {
				_, _ = fmt.Fprintf(&report, "%q,%q,%q,%q,%q,%d,%q,%q\n", hostTestReport.HostName, testsResults.UserName, testId, testengine.GetAbstract(testId), resultToString(testsResults.ExecTestStatuses[testId].Result), testsResults.ExecTestStatuses[testId].ExecStatus.RetCode, testsResults.ExecTestStatuses[testId].ExecStatus.Stdout, testsResults.ExecTestStatuses[testId].ExecStatus.Stderr)
			}
		}
	}

	return saveToCSVFile(hostTestReport.HostName, filePath, report.Bytes())
}

func saveToCSVFile(hostName string, filePath string, report []byte) error {
	if filePath == "" {
		filePath = filepath.Join(REPORT_NAME + "-" + hostName + "-" + time.Now().Format("2006-01-02_15-04-05_MST") + ".csv")
	}
	filePath = filepath.Clean(filePath)
	if err := os.WriteFile(filePath, report, 0644); err != nil {
		return fmt.Errorf("failed to save to file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}
