package report

import (
	"path/filepath"
	"strings"
)

// SanitizeFilePath ensures a file path has a specified default extension and returns a sanitized version of the path.
func SanitizeFilePath(filePath, defaultFilePath, defaultExtension string) string {
	return sanitizeFilePath(filePath, defaultFilePath, defaultExtension)
}

// sanitizeFilePath ensures a file path has a default extension and returns a cleaned version of the path.
func sanitizeFilePath(filePath, defaultFilePath, defaultExtension string) string {
	if len(defaultExtension) == 0 {
		panic("default extension cannot be empty")
	}

	if len(defaultFilePath) == 0 {
		panic("default file path cannot be empty")
	}

	if filePath == "" {
		filePath = defaultFilePath
	}

	if len(defaultExtension) > 0 && defaultExtension[0] != '.' {
		defaultExtension = "." + defaultExtension
	}

	fileExt := filepath.Ext(filePath)

	if fileExt == "" {
		filePath = filePath + defaultExtension
		fileExt = filepath.Ext(filePath)
	}

	if fileExt != defaultExtension {
		filePath = strings.TrimSuffix(filePath, fileExt) + defaultExtension
	}

	return filepath.Clean(filePath)
}
