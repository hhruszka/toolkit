package bptcommon

import (
	"testing"
)

func TestSanitizeFilePath(t *testing.T) {
	var testCases = []struct {
		filePath, defaultFilePath, defaultExtension string
		expectedFilePath                            string
	}{
		{"", "default.txt", ".txt", "default.txt"},
		{"existing.txt", "default.txt", ".txt", "existing.txt"},
		{"", "default.txt", ".md", "default.md"},
		{"report.xls", "default.xlsx", ".xlsx", "report.xlsx"},
		{"report.namespace.xls", "default.xlsx", "xlsx", "report.namespace.xlsx"},
		{"report.namespace.xlsx", "default.xlsx", "xlsx", "report.namespace.xlsx"},
	}

	for _, tc := range testCases {
		actualFilePath := sanitizeFilePath(tc.filePath, tc.defaultFilePath, tc.defaultExtension)
		if actualFilePath != tc.expectedFilePath {
			t.Errorf("sanitizeFilePath(%s, %s, %s) = %s, expected %s", tc.filePath, tc.defaultFilePath, tc.defaultExtension, actualFilePath, tc.expectedFilePath)
		}
	}
}
