package reports

import (
	"bptvnftester/testengine"
	"bytes"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func generateTextReport(hostTestReport *HostTestReport) error {
	var report bytes.Buffer

	for _, testsResults := range hostTestReport.Tests {
		var sortedTestIds []string

		for testId, _ := range testsResults.ExecTestStatuses {
			sortedTestIds = append(sortedTestIds, testId)
		}
		sort.Strings(sortedTestIds)

		t := table.NewWriter()
		t.SetOutputMirror(&report)
		t.SetTitle("Host: %s\nUser: %s\nDate: %s\nVersion: %s\n", hostTestReport.HostName, testsResults.UserName, hostTestReport.Date, hostTestReport.Version)
		t.SetAllowedRowLength(140)

		for _, testId := range sortedTestIds {
			if testengine.AllTestCases[testId].IsTest {
				t.AppendRows([]table.Row{{testId, testengine.AllTestCases[testId].Abstract, resultToString(testengine.AllTestCases[testId].ResultFunc(testsResults.ExecTestStatuses[testId].ExecStatus))}})
				t.AppendSeparator()
			}
		}
		t.Render()
	}
	//fmt.Println(report.String())
	return saveTextReportToFile(hostTestReport.HostName, report.Bytes())
}

func generateDetailedTextReport(hostTestReport *HostTestReport) error {
	var report bytes.Buffer
	for _, testsResults := range hostTestReport.Tests {
		var buf bytes.Buffer
		var sortedTestIds []string

		for testId, _ := range testsResults.ExecTestStatuses {
			sortedTestIds = append(sortedTestIds, testId)
		}
		sort.Strings(sortedTestIds)

		t := table.NewWriter()
		t.SetOutputMirror(&buf)
		t.SetTitle("Host: %s\nUser: %s\nDate: %s\nVersion: %s", hostTestReport.HostName, testsResults.UserName, hostTestReport.Date, hostTestReport.Version)
		t.SetAllowedRowLength(140)

		for _, testId := range sortedTestIds {
			if testengine.AllTestCases[testId].IsTest {
				t.AppendRow(nil)
				t.AppendSeparator()
				t.AppendRows([]table.Row{{testId, testengine.AllTestCases[testId].Abstract, resultToString(testengine.AllTestCases[testId].ResultFunc(testsResults.ExecTestStatuses[testId].ExecStatus))}})
				t.AppendSeparator()
				t.SetOutputMirror(&buf)
				t.AppendRow(table.Row{"Exit status:", testsResults.ExecTestStatuses[testId].ExecStatus.RetCode})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Error", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Error, 100)})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Stdout:", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Stdout, 100)})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Stderr:", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Stderr, 100)})
				t.AppendSeparator()
			}
		}
		t.Render()
		fmt.Fprintf(&report, buf.String())
	}

	return saveTextReportToFile(hostTestReport.HostName, report.Bytes())
}

func saveTextReportToFile(hostName string, report []byte) error {
	filePath := filepath.Join(REPORT_NAME + "-" + hostName + "-" + time.Now().Format("2006-01-02_15-04-05_MST") + ".txt")
	if err := os.WriteFile(filePath, report, 0444); err != nil {
		return fmt.Errorf("failed to save to file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}
