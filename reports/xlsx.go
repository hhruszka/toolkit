package reports

import (
	"bptvnftester/testengine"
	"fmt"
	"github.com/xuri/excelize/v2"
	"maps"
	"os"
	"slices"
	"strings"
	"time"
)

func prepXLSXFile(xlsxFile *excelize.File) error {
	if idx, err := xlsxFile.GetSheetIndex("Sheet1"); idx == 0 && err == nil {
		if err := xlsxFile.DeleteSheet("Sheet1"); err != nil {
			return fmt.Errorf("failed to delete sheet due to: %w", err)
		}
	}

	return nil
}

func setSheetName(xlsxFile *excelize.File, sheetName string) error {
	if len(sheetName) > 31 {
		sheetName = sheetName[:31]
	}

	if idx, err := xlsxFile.GetSheetIndex("Sheet1"); idx == 0 && err == nil {
		if err := xlsxFile.SetSheetName("Sheet1", sheetName); err != nil {
			return fmt.Errorf("failed to delete sheet due to: %w", err)
		}
		return nil
	}
	if _, err := xlsxFile.NewSheet(sheetName); err != nil {
		return fmt.Errorf("failed to generate xlsx file due to: %w", err)
	}
	return nil
}

// saveAllAccountsTestReportsToXLSX writes test reports from multiple namespaces to an Excel file using the provided structure.
// Takes a slice of NamespaceTestReport and an Excel file to modify. Returns an error on failure.
func saveAllAccountsTestReportsToXLSX(report *HostTestReport, xlsxFile *excelize.File) error {
	for _, accountReport := range report.Tests {
		if err := saveAccountTestReportToXLSX(report.HostName, report.Version, accountReport, accountReport.UserName, xlsxFile); err != nil {
			return err
		}
	}
	if err := saveTestResultsVNFBPT66ToXLS(report, "VNFBPT66 Details", xlsxFile); err != nil {
		return err
	}
	return nil
}

// saveNamespaceTestReportToXLSX writes the test report data into an XLSX file, creating or updating the appropriate sheet.
// It formats the sheet header, writes data rows, and applies styling and column adjustments.
// Parameters: `report` contains the namespace test data, `xlsxFile` represents the Excel file to modify.
// Returns an error if any operation fails during the process.
func saveAccountTestReportToXLSX(hostName string, version string, accountResults *testengine.AccountTestResults, sheetName string, xlsxFile *excelize.File) error {
	if err := setSheetName(xlsxFile, sheetName); err != nil {
		return err
	}

	// Set a header row
	headerRow := 1
	headers := []string{"Host", "Time", "App Version", "User", "Test ID", "Test Title", "Result", "Return Code", "Errors", "Stdout", "Stderr", "JIRA Ticker", "Remarks"}
	for col, header := range headers {
		if err := xlsxFile.SetCellValue(sheetName, _cell(col+1, headerRow), header); err != nil {
			return err
		}
	}

	// Write data rows
	row := headerRow + 1
	col, _ := excelize.ColumnNameToNumber("A")

	tests := accountResults.ExecTestStatuses
	testIDs := slices.Sorted(maps.Keys(tests))
	for _, testId := range testIDs {
		test := tests[testId]
		_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), hostName)
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+1, row), test.ExecStatus.ExecTime.Format(time.RFC822))
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+2, row), version)
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+3, row), accountResults.UserName)
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+4, row), testId)
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+5, row), testengine.GetAbstract(testId))
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+6, row), testResult(test, false))
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+7, row), test.ExecStatus.RetCode)
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+8, row), strings.Join(test.ExecStatus.Error, "\n"))
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+9, row), strings.Join(test.ExecStatus.Stdout, "\n"))
		_ = xlsxFile.SetCellValue(sheetName, _cell(col+10, row), strings.Join(test.ExecStatus.Stderr, "\n"))
		row++
	}

	return formatCols(xlsxFile, sheetName, col, col+12, headerRow, row)
}

// saveScanResultsVNFBPT66ToXLS saves VNFBPT66 test results from multiple namespaces into an Excel sheet.
// Takes a slice of NamespaceTestReport, the sheet name, and an Excel file to modify.
// Returns an error if creating or writing to the sheet fails.
func saveTestResultsVNFBPT66ToXLS(report *HostTestReport, sheetName string, xlsxFile *excelize.File) error {
	if err := setSheetName(xlsxFile, sheetName); err != nil {
		return err
	}

	// Set a header row
	headerRow := 1
	headers := []string{"Host", "Time", "App Version", "User", "Name", "Value", "Type", "Confidence", "File Path", "Readable"}
	for col, header := range headers {
		if err := xlsxFile.SetCellValue(sheetName, _cell(col+1, headerRow), header); err != nil {
			return err
		}
	}

	// Write data rows
	row := headerRow + 1
	col, _ := excelize.ColumnNameToNumber("A")

	hostName := report.HostName
	for _, accountReport := range report.Tests {
		userName := accountReport.UserName
		for _, test := range accountReport.ExecTestStatuses {
			if test.ExecStatus.Data != nil {
				vnfbpt66results, ok := test.ExecStatus.Data.([]*testengine.SecretDetectorResult)
				if ok && vnfbpt66results != nil {
					for _, result := range vnfbpt66results {
						_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), hostName)
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+1, row), test.ExecStatus.ExecTime.Format(time.RFC822))
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+2, row), report.Version)
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+3, row), userName)
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+4, row), result.Name())
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+5, row), result.Value())
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+6, row), result.Type())
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+7, row), result.Confidence())
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+8, row), result.File())
						_ = xlsxFile.SetCellValue(sheetName, _cell(col+9, row), result.Readable())
						row++
					}
				}
			}
		}
	}

	if row-headerRow == 1 {
		_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), "Nothing to report")
	}
	return formatCols(xlsxFile, sheetName, col, col+9, headerRow, row)
}

func noResults(xlsxFile *excelize.File, sheetName string, row int, col int) {
	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), "No secrets found")
}

//func saveNamespaceCNFBPT57ToXLS(report []*NamespaceTestReport, sheetName string, xlsxFile *excelize.File) error {
//	var err error
//
//	if len(sheetName) > 31 {
//		sheetName = sheetName[:31]
//	}
//
//	if xlsxFile.SheetCount == 1 && xlsxFile.GetSheetName(0) == "Sheet1" {
//		// When the xlsx file gets created, then the 'Sheet1' is automatically created as well.
//		// We need to rename it to our namespace.
//		err = xlsxFile.SetSheetName(xlsxFile.GetSheetName(0), sheetName)
//		if err != nil {
//			// this is the first and the only tab. We cannot change its name to 31 char name
//			return err
//		}
//	} else {
//		// check if this sheet exists, if no then create it
//		if idx, _ := xlsxFile.GetSheetIndex(sheetName); idx == -1 {
//			if _, err = xlsxFile.NewSheet(sheetName); err != nil {
//				return fmt.Errorf("failed to generate xlsx file due to: %w", err)
//			}
//		}
//	}
//
//	// Set a header row
//	headerRow := 1
//	headers := []string{"Namespace", "Time", "App Version", "Pod", "Container", "Image", "Line No", "Value", "Type", "Confidence", "Context"}
//	for col, header := range headers {
//		if err := xlsxFile.SetCellValue(sheetName, _cell(col+1, headerRow), header); err != nil {
//			return err
//		}
//	}
//
//	// Write data rows
//	row := headerRow + 1
//	col, _ := excelize.ColumnNameToNumber("A")
//	for _, namespaceReport := range report {
//		for _, pod := range namespaceReport.PodsTestResults {
//			for _, container := range pod.ContainersTestResults {
//				for _, test := range container.TestsResults {
//					cnfbpt57, ok := test.TestData.([]*core.Secret)
//					if strings.ToUpper(test.TestId) == "CNFBPT57" && ok && cnfbpt57 != nil {
//						for _, resultToString := range cnfbpt57 {
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), namespaceReport.Metadata.Namespace)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+1, row), test.ExecTime.Format(time.RFC822))
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+2, row), namespaceReport.Metadata.AppVersion)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+3, row), pod.PodName)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+4, row), container.ContainerName)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+5, row), container.ContainerImage)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+6, row), resultToString.LineNumber)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+7, row), resultToString.SecretValue)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+8, row), resultToString.SecretType)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+9, row), resultToString.Confidence)
//							_ = xlsxFile.SetCellValue(sheetName, _cell(col+10, row), resultToString.Context)
//							row++
//						}
//					}
//				}
//			}
//		}
//	}
//	err = formatCols(xlsxFile, sheetName, col, col+10, headerRow, row)
//	return err
//}

// setHeadersRow sets headers for an Excel sheet based on provided metadata, pod, and container information.
// It writes metadata-related rows, followed by column headers for test results.
// Returns the updated row number and an error if any error occurs during the cell value setting.
//func setHeadersRow(meta *NamespaceTestReportMeta, podName string, containerName string, xlsxFile *excelize.File, sheetName string, row int, col int) (int, error) {
//	// Set a header row
//	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), fmt.Sprintf("Namespace: %s", meta.Namespace))
//	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row+1), fmt.Sprintf("Date: %s", meta.Timestamp))
//	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row+2), fmt.Sprintf("Version: %s", meta.AppVersion))
//	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row+3), fmt.Sprintf("Pod: %s", podName))
//	_ = xlsxFile.SetCellValue(sheetName, _cell(col, row+4), fmt.Sprintf("Container: %s", containerName))
//	row += 6
//
//	headers := []string{"Container", "Image", "Test ID", "Test Title", "UID", "Result", "Return Code", "Errors", "Stdout", "Stderr"}
//	for offset, header := range headers {
//		if err := xlsxFile.SetCellValue(sheetName, _cell(col+offset, row), header); err != nil {
//			return -1, err
//		}
//	}
//
//	return row + 1, nil
//}

// saveContainerTestResultsToXLSX writes container test results to an Excel sheet starting from a specified row and column.
// containerTestResults is the container's test resultToString data to be written.
// xlsxFile is the Excel file object where data will be stored.
// sheetName specifies the worksheet to write the data in.
// Row and col define the starting row and column respectively for the data insertion.
// Returns the next available row after data insertion and an error if any operation fails.
//func saveContainerTestResultsToXLSX(containerTestResults *ContainerTestResults, xlsxFile *excelize.File, sheetName string, row int, col int) (int, error) {
//	for _, test := range containerTestResults.TestsResults {
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col, row), containerTestResults.ContainerName)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+1, row), containerTestResults.ContainerImage)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+2, row), test.TestId)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+3, row), test.TestName)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+4, row), containerTestResults.Uid)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+5, row), test.Result)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+6, row), test.RetCode)
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+7, row), strings.Join(test.Error, "\n"))
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+8, row), strings.Join(test.Stdout, "\n"))
//		_ = xlsxFile.SetCellValue(sheetName, _cell(col+9, row), strings.Join(test.Stderr, "\n"))
//		row++
//	}
//
//	return row, nil
//}

// SaveToXLSXFile generates and saves an Excel report for the given namespace test report to a timestamped XLSX file.
// It includes per-container and aggregated namespace data, creating sheets for each container in the namespace.
// Returns an error if the report generation or file saving fails.
func SaveToXLSXFile(report *HostTestReport) error {
	// Create a new Excel file
	xlsxFile := excelize.NewFile()

	// Add data to the file
	if err := saveAllAccountsTestReportsToXLSX(report, xlsxFile); err != nil {
		return err
	}

	// Save the file
	filePath := REPORT_NAME + "-" + report.HostName + "-" + time.Now().Format("2006-01-02_15-04-05_MST") + ".xlsx"
	if err := xlsxFile.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to generate xlsx file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}
