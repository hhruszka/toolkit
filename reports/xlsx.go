package reports

import (
	"bptvnftester/testengine"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	rp "gitlabe1.ext.net.nokia.com/cn-pentesting-repo/bpt-common/report"
)

func prepXLSXFile(xlsxFile *excelize.File) error {
	if idx, err := xlsxFile.GetSheetIndex("Sheet1"); idx == 0 && err == nil {
		if err := xlsxFile.DeleteSheet("Sheet1"); err != nil {
			return fmt.Errorf("failed to delete sheet due to: %w", err)
		}
	}

	return nil
}

// saveTestReportsToXLSX writes test reports from multiple namespaces to an Excel file using the provided structure.
// Takes a slice of NamespaceTestReport and an Excel file to modify. Returns an error on failure.
func saveTestReportsToXLSX(report *HostTestReport, xlsxFile *excelize.File) error {
	for _, accountReport := range report.Tests {
		if err := saveTestReportToXLSX(report.HostName, report.Version, accountReport, accountReport.UserName, xlsxFile); err != nil {
			return err
		}
	}
	if err := saveTestResultsVNFBPT66ToXLS(report, "VNFBPT66 Details", xlsxFile); err != nil {
		return err
	}
	return nil
}

// saveTestReportToXLSX writes the test report data into an XLSX file, creating or updating the appropriate sheet.
// It formats the sheet header, writes data rows, and applies styling and column adjustments.
// Parameters: `report` contains the namespace test data, `xlsxFile` represents the Excel file to modify.
// Returns an error if any operation fails during the process.
func saveTestReportToXLSX(hostName string, version string, accountResults *testengine.AccountTestResults, sheetName string, xlsxFile *excelize.File) error {
	var err error

	sheetName, err = rp.SetSheetName(xlsxFile, sheetName)
	if err != nil {
		return fmt.Errorf("failed to set sheet name; %w", err)
	}

	// Set a header row
	headerRow := 1
	headers := []string{"Host", "Time", "App Version", "User", "Test ID", "Test Title", "Result", "Return Code", "Errors", "Stdout", "Stderr", "JIRA Ticker", "Remarks"}
	for col, header := range headers {
		if err := xlsxFile.SetCellValue(sheetName, rp.Cell(col+1, headerRow), header); err != nil {
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
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col, row), hostName)
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+1, row), test.ExecStatus.ExecTime.Format(time.RFC822))
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+2, row), version)
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+3, row), accountResults.UserName)
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+4, row), testId)
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+5, row), testengine.GetAbstract(testId))
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+6, row), testResult(test, false))
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+7, row), test.ExecStatus.RetCode)
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+8, row), strings.Join(test.ExecStatus.Error, "\n"))
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+9, row), strings.Join(test.ExecStatus.Stdout, "\n"))
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+10, row), strings.Join(test.ExecStatus.Stderr, "\n"))
		row++
	}

	return rp.FormatCols(xlsxFile, sheetName, col, col+12, headerRow, row)
}

// saveScanResultsVNFBPT66ToXLS saves VNFBPT66 test results from multiple namespaces into an Excel sheet.
// Takes a slice of NamespaceTestReport, the sheet name, and an Excel file to modify.
// Returns an error if creating or writing to the sheet fails.
func saveTestResultsVNFBPT66ToXLS(report *HostTestReport, sheetName string, xlsxFile *excelize.File) error {
	var err error

	sheetName, err = rp.SetSheetName(xlsxFile, sheetName)
	if err != nil {
		return fmt.Errorf("failed to set sheet name; %w", err)
	}

	// Set a header row
	headerRow := 1
	headers := []string{"Host", "Time", "App Version", "User", "Name", "Value", "Type", "Confidence", "File Path", "Readable"}
	for col, header := range headers {
		if err := xlsxFile.SetCellValue(sheetName, rp.Cell(col+1, headerRow), header); err != nil {
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
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col, row), hostName)
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+1, row), test.ExecStatus.ExecTime.Format(time.RFC822))
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+2, row), report.Version)
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+3, row), userName)
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+4, row), result.Name())
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+5, row), result.Value())
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+6, row), result.Type())
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+7, row), result.Confidence())
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+8, row), result.File())
						_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col+9, row), result.Readable())
						row++
					}
				}
			}
		}
	}

	if row-headerRow == 1 {
		_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col, row), "Nothing to report")
	}
	return rp.FormatCols(xlsxFile, sheetName, col, col+9, headerRow, row)
}

func noResults(xlsxFile *excelize.File, sheetName string, row int, col int) {
	_ = xlsxFile.SetCellValue(sheetName, rp.Cell(col, row), "No secrets found")
}

// SaveToXLSXFile generates and saves an Excel report for the given namespace test report to a timestamped XLSX file.
// It includes per-container and aggregated namespace data, creating sheets for each container in the namespace.
// Returns an error if the report generation or file saving fails.
func SaveToXLSXFile(report *HostTestReport, filePath string, reportToXLSX func(*HostTestReport, *excelize.File) error) error {
	// Create a new Excel file
	xlsxFile := excelize.NewFile()

	// Add data to the file
	if err := reportToXLSX(report, xlsxFile); err != nil {
		return err
	}

	filePath = rp.SanitizeFilePath(filePath, REPORT_NAME+"-"+report.HostName+"-"+time.Now().Format("2006-01-02_15-04-05_MST")+".xlsx", ".xlsx")

	if err := xlsxFile.SaveAs(filePath); err != nil {
		return fmt.Errorf("failed to generate xlsx file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}
