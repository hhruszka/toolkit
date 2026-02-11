package reports

import (
	"bptvnftester/testengine"
	"fmt"
)

var GenReport = genReport

// GenReport generates a report based on the input test results and saves it in the specified format (json, xlsx, text, or csv).
// The function uses global variables `format`, `namespace`, `AppVersion`, and `failedOnly` to control its behavior.
// Returns an error if the report generation or saving process fails.
func genReport(results []*testengine.AccountTestResults, hostName string, fileName string, format string, detailed bool, appVersion string) error {
	report := ConvertTestData(hostName, appVersion, results)
	switch format {
	case "json":
		if err := SaveReportToJSONFile(report, fileName); err != nil {
			return fmt.Errorf("internal application error: %s\n", err.Error())
		}
	case "xlsx", "xls":
		if err := SaveToXLSXFile(report, fileName, saveTestReportsXLSXStream); err != nil {
			return fmt.Errorf("internal application error: %s\n", err.Error())
		}
	case "text", "txt":
		if detailed {
			return SaveDetailedReportToTextFile(report, fileName)
		}
		return SaveReportToTextFile(report, fileName)
	case "csv":
		return SaveToCSVFile(report, fileName)
	default:
		return fmt.Errorf("unsupported report format: %s\n", format)
	}

	return nil
}
