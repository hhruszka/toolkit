package reports

import (
	"bptvnftester/testengine"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// saveTestReportsToXLSX writes test reports from multiple namespaces to an Excel file using the provided structure.
// Takes a slice of NamespaceTestReport and an Excel file to modify. Returns an error on failure.
func saveTestReportsXLSXStream(report *HostTestReport, xlsxFile *excelize.File) error {
	SetDefaultStyles(xlsxFile)

	for _, accountReport := range report.Tests {
		if err := saveTestReportXLSXStream(report.HostName, report.Version, accountReport, accountReport.UserName, xlsxFile); err != nil {
			return err
		}
	}
	if err := saveTestResultsVNFBPT66XLSStream(report, "VNFBPT66 Details", xlsxFile); err != nil {
		return err
	}
	return nil
}

// calcColumnWidthTestResults calculates column widths for displaying Go binary information from image scan results.
// imageScanResults is a slice of ImageScan objects containing scan data for images and their layers.
// Returns a slice of integers representing the maximum width for each column.
func calcColumnWidthTestResults(hostName string, version string, accountResults *testengine.AccountTestResults, headers []interface{}) []int {
	var colWidths = make([]int, len(headers))

	updateWidth := func(colIndex int, val interface{}) {
		var length int
		var strVal string

		switch v := val.(type) {
		case excelize.Cell:
			strVal = fmt.Sprintf("%v", v.Value)
		case string:
			strVal = v
		default:
			strVal = fmt.Sprintf("%v", v)
		}

		strVal = strings.Trim(strVal, " ")
		if strings.Contains(strVal, "\n") {
			strValues := strings.Split(strVal, "\n")
			for _, str := range strValues {
				length = max(len(str), length)
			}
		} else {
			length = len(strVal)
		}

		colWidths[colIndex] = max(length, colWidths[colIndex])
	}

	for i, header := range headers {
		updateWidth(i, header)
	}

	tests := accountResults.ExecTestStatuses
	testIDs := slices.Sorted(maps.Keys(tests))
	for _, testId := range testIDs {
		test := tests[testId]
		values := []interface{}{
			hostName,
			test.ExecStatus.ExecTime.Format(time.RFC822),
			version,
			accountResults.UserName,
			testId,
			testengine.GetAbstract(testId),
			testResult(test, false),
			test.ExecStatus.RetCode,
			strings.Join(test.ExecStatus.Error, "\n"),
			strings.Join(test.ExecStatus.Stdout, "\n"),
			strings.Join(test.ExecStatus.Stderr, "\n"),
		}
		for i, value := range values {
			updateWidth(i, value)
		}
	}
	return colWidths
}

// saveTestReportXLSXStream writes the test report data into an XLSX file, creating or updating the appropriate sheet.
// It formats the sheet header, writes data rows, and applies styling and column adjustments.
// Parameters: `report` contains the namespace test data, `xlsxFile` represents the Excel file to modify.
// Returns an error if any operation fails during the process.
func saveTestReportXLSXStream(hostName string, version string, accountResults *testengine.AccountTestResults, sheetName string, xlsxFile *excelize.File) error {
	var (
		err error
		sw  *excelize.StreamWriter
	)

	sheetName, err = setSheetName(xlsxFile, sheetName)
	if err != nil {
		return fmt.Errorf("failed to set sheet name; %w", err)
	}
	sw, err = xlsxFile.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create stream writer; %w", err)
	}

	col, _ := excelize.ColumnNameToNumber("A")

	// Set a header row
	headers := []interface{}{"Host", "Time", "App Version", "User", "Test ID", "Test Title", "Result", "Return Code", "Errors", "Stdout", "Stderr", "JIRA Ticker", "Remarks"}

	colWidth := calcColumnWidthTestResults(hostName, version, accountResults, headers)
	if err = setColWidthWithStreamWriter(sw, col, colWidth); err != nil {
		return fmt.Errorf("failed to set column width; %w", err)
	}

	// Write header row
	headerRow := 1
	err = sw.SetRow(_cell(col, headerRow), headers)
	if err != nil {
		return fmt.Errorf("failed to write header row; %w", err)
	}

	// Write data rows
	row := headerRow + 1

	tests := accountResults.ExecTestStatuses
	testIDs := slices.Sorted(maps.Keys(tests))
	for _, testId := range testIDs {
		test := tests[testId]
		values := []interface{}{
			excelize.Cell{StyleID: styleNotWrappedId, Value: hostName},
			excelize.Cell{StyleID: styleNotWrappedId, Value: test.ExecStatus.ExecTime.Format(time.RFC822)},
			excelize.Cell{StyleID: styleNotWrappedId, Value: version},
			excelize.Cell{StyleID: styleNotWrappedId, Value: accountResults.UserName},
			excelize.Cell{StyleID: styleNotWrappedId, Value: testId},
			excelize.Cell{StyleID: styleNotWrappedId, Value: testengine.GetAbstract(testId)},
			excelize.Cell{StyleID: styleNotWrappedId, Value: testResult(test, false)},
			excelize.Cell{StyleID: styleNotWrappedId, Value: test.ExecStatus.RetCode},
			excelize.Cell{StyleID: styleWrappedId, Value: strings.Join(test.ExecStatus.Error, "\n")},
			excelize.Cell{StyleID: styleWrappedId, Value: strings.Join(test.ExecStatus.Stdout, "\n")},
			excelize.Cell{StyleID: styleWrappedId, Value: strings.Join(test.ExecStatus.Stderr, "\n")},
		}
		err = sw.SetRow(_cell(col, row), values)
		if err != nil {
			return fmt.Errorf("failed to write data row; %w", err)
		}
		row++
	}

	if err = sw.Flush(); err != nil {
		return fmt.Errorf("failed to flush stream writer; %w", err)
	}
	return nil
}

// calcColumnWidthVNFBPT66TestResults calculates column widths for displaying Go binary information from image scan results.
// imageScanResults is a slice of ImageScan objects containing scan data for images and their layers.
// Returns a slice of integers representing the maximum width for each column.
func calcColumnWidthVNFBPT66TestResults(report *HostTestReport, headers []interface{}) []int {
	var colWidths = make([]int, len(headers))

	updateWidth := func(colIndex int, val interface{}) {
		var length int
		var strVal string

		switch v := val.(type) {
		case excelize.Cell:
			strVal = fmt.Sprintf("%v", v.Value)
		case string:
			strVal = v
		default:
			strVal = fmt.Sprintf("%v", v)
		}

		if strings.Contains(strVal, "\n") {
			strValues := strings.Split(strVal, "\n")
			for _, str := range strValues {
				length = max(len(str), length)
			}
		} else {
			length = len(strVal)
		}

		colWidths[colIndex] = max(length, colWidths[colIndex])
	}

	for i, header := range headers {
		updateWidth(i, header)
	}

	hostName := report.HostName
	for _, accountReport := range report.Tests {
		userName := accountReport.UserName
		for _, test := range accountReport.ExecTestStatuses {
			if test.ExecStatus.Data != nil {
				vnfbpt66results, ok := test.ExecStatus.Data.([]*testengine.SecretDetectorResult)
				if ok && vnfbpt66results != nil {
					for _, result := range vnfbpt66results {
						values := []interface{}{
							hostName,
							test.ExecStatus.ExecTime.Format(time.RFC822),
							report.Version,
							userName,
							result.Name(),
							result.Value(),
							result.Type(),
							result.Confidence(),
							result.File(),
							result.Readable(),
						}
						for i, value := range values {
							updateWidth(i, value)
						}
					}
				}
			}
		}
	}
	return colWidths
}

// saveTestResultsVNFBPT66XLSStream saves VNFBPT66 test results from multiple namespaces into an Excel sheet.
// Takes a slice of NamespaceTestReport, the sheet name, and an Excel file to modify.
// Returns an error if creating or writing to the sheet fails.
func saveTestResultsVNFBPT66XLSStream(report *HostTestReport, sheetName string, xlsxFile *excelize.File) error {
	var (
		err error
		sw  *excelize.StreamWriter
	)

	sheetName, err = setSheetName(xlsxFile, sheetName)
	if err != nil {
		return fmt.Errorf("failed to set sheet name; %w", err)
	}
	sw, err = xlsxFile.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create stream writer; %w", err)
	}

	headers := []interface{}{"Host", "Time", "App Version", "User", "Name", "Value", "Type", "Confidence", "File Path", "Readable"}

	col, _ := excelize.ColumnNameToNumber("A")
	colWidth := calcColumnWidthVNFBPT66TestResults(report, headers)
	if err = setColWidthWithStreamWriter(sw, col, colWidth); err != nil {
		return fmt.Errorf("failed to set column width; %w", err)
	}

	// Set a header row
	headerRow := 1
	if err = sw.SetRow(_cell(1, headerRow), headers); err != nil {
		return fmt.Errorf("failed to write header row; %w", err)
	}

	// Write data rows
	row := headerRow + 1

	hostName := report.HostName
	for _, accountReport := range report.Tests {
		userName := accountReport.UserName
		for _, test := range accountReport.ExecTestStatuses {
			if test.ExecStatus.Data != nil {
				vnfbpt66results, ok := test.ExecStatus.Data.([]*testengine.SecretDetectorResult)
				if ok && vnfbpt66results != nil {
					for _, result := range vnfbpt66results {
						if row >= excelize.TotalRows {
							return fmt.Errorf("exceeded maximum number of rows %d (allowed %d)", row, excelize.TotalRows)
						}
						values := []interface{}{
							hostName,
							test.ExecStatus.ExecTime.Format(time.RFC822),
							report.Version,
							userName,
							result.Name(),
							result.Value(),
							result.Type(),
							result.Confidence(),
							result.File(),
							result.Readable(),
						}
						if err = sw.SetRow(_cell(col, row), values); err != nil {
							return fmt.Errorf("failed to write data row; %w", err)
						}
						row++
					}
				}
			}
		}
	}

	if row-headerRow == 1 {
		_ = sw.SetRow(_cell(col, row), []interface{}{"Nothing to report"})
	}
	if err = sw.Flush(); err != nil {
		return fmt.Errorf("failed to flush stream writer; %w", err)
	}
	return nil
}
