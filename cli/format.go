package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// ValidateFormat validates the given format against a list of supported formats and returns the format and file path.
func ValidateFormat(format string, supportedFormats []string) (string, string, error) {
	return validateFormat(format, supportedFormats)
}

// validateFormat validates the provided report format against a list of supported formats and checks the file path validity.
func validateFormat(reportFormat string, supportedFormats []string) (string, string, error) {
	var reportFile string

	reportFormat = strings.ToLower(reportFormat)
	reportFormat = strings.TrimSpace(reportFormat)

	// 1. Check if config.Format specifies format and supported
	// 2. Since config.Format does not specify format, check if file path specifies format through extension
	if slices.Contains(supportedFormats, reportFormat) {
		return reportFormat, "", nil
	}

	reportFile = reportFormat
	extension := filepath.Ext(reportFormat)

	if extension != "" {
		reportFormat = extension[1:]
	}
	if extension == "" || reportFormat == "" {
		return "", "", fmt.Errorf("missing extension in the provided file path %s", reportFile)
	}

	if !slices.Contains(supportedFormats, reportFormat) {
		return "", "", fmt.Errorf("%s is not a valid report format for the output option (-o or --output), aborting", reportFormat)
	}

	if strings.TrimSuffix(reportFile, filepath.Ext(reportFile)) == "" {
		return "", "", fmt.Errorf("missing file name in the provided file path %s", reportFile)
	}

	if reportFile != "" {
		reportFile = filepath.Clean(reportFile)
		stat, err := os.Stat(filepath.Dir(reportFile))
		if err == nil && stat != nil && !stat.IsDir() {
			return "", "", fmt.Errorf("the provided file path is not valid; aborting")
		}
		if err != nil {
			return "", "", fmt.Errorf("failed to access the provided file path due to: %w", err)
		}
	}
	return reportFormat, reportFile, nil
}
