package reports

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func SaveReportToJSONFile(report *HostTestReport) error {
	jsonBuff, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		return fmt.Errorf("internal application error: %s\n", err.Error())
	}

	filePath := REPORT_NAME + "-" + report.HostName + "-" + time.Now().Format("2006-01-02_15-04-05_MST") + ".json"
	if err := os.WriteFile(filePath, jsonBuff, 0444); err != nil {
		return fmt.Errorf("failed to generate json file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}
