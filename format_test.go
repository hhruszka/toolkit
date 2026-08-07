package bptcommon

import "testing"

var supportedExtensions = []string{"txt", "md", "xlsx"}

func TestValidateFormatFormat(t *testing.T) {
	var testCases = []struct {
		filePath         string
		expectedFormat   string
		expectedFilePath string
		error            bool
	}{
		{"test.json", "", "", true},
		{"test.", "", "", true},
		{"test.txt", "txt", "test.txt", false},
		{"test", "", "", true},
		{"txt", "txt", "", false},
		{".txt", "", "", true},
		{".tx", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.filePath, func(t *testing.T) {
			format, actualFilePath, err := ValidateFormat(tc.filePath, supportedExtensions)
			if actualFilePath != tc.expectedFilePath || format != tc.expectedFormat || (err != nil) != tc.error {
				t.Errorf("expected filepath %s, got %s, expected format %s got %s, expected error %t got %t", tc.expectedFilePath, actualFilePath, tc.expectedFormat, format, tc.error, err != nil)
			}
		})
	}
}
