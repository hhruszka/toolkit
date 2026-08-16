package reports

import (
	"bptvnftester/testengine"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func SaveToCSVFile(hostTestReport *HostTestReport, filePath string) error {
	report := new(bytes.Buffer)
	w := csv.NewWriter(report)

	err := w.Write([]string{"Host", "User", "Test Id", "Abstract", "Result", "Execution Code", "Stdout", "Stderr"})
	if err != nil {
		return fmt.Errorf("failed to write header to CSV file: %v", err)
	}

	for _, testsResults := range hostTestReport.Tests {
		var sortedTestIds []string

		for testId, _ := range testsResults.ExecTestStatuses {
			sortedTestIds = append(sortedTestIds, testId)
		}
		sort.Strings(sortedTestIds)

		for _, testId := range sortedTestIds {
			if testengine.AllTestCases[testId].IsTest {
				err = w.Write([]string{hostTestReport.HostName, testsResults.UserName, testId, testengine.GetAbstract(testId), resultToString(testsResults.ExecTestStatuses[testId].Result), strconv.Itoa(int(testsResults.ExecTestStatuses[testId].ExecStatus.RetCode)), strings.Join(testsResults.ExecTestStatuses[testId].ExecStatus.Stdout, "\n"), strings.Join(testsResults.ExecTestStatuses[testId].ExecStatus.Stderr, "\n")})
				if err != nil {
					return fmt.Errorf("failed to write header to CSV file: %v", err)
				}
			}
		}
	}

	w.Flush()
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
