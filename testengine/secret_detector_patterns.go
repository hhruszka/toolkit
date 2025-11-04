package testengine

import (
	secretpatterns "github.com/hhruszka/secretscanner/patterns"
	"regexp"
)

// SecretDetectorPatterns holds all regex patterns used for secret detection
type SecretDetectorPatterns struct {
	// Confidence 4 patterns (90% certainty)
	HighConfidencePasswordRegex *regexp.Regexp
	HighConfidenceSecretRegex   *regexp.Regexp
	CommonEnvVarPatterns        []*regexp.Regexp

	// Confidence 3 patterns (75% certainty)
	MediumConfidencePasswordRegex *regexp.Regexp
	UserRegex                     *regexp.Regexp
	SensitiveFilesRegex           *regexp.Regexp

	// Confidence 2 patterns (50% certainty)
	LowConfidencePasswordRegex *regexp.Regexp
	SensitiveFilePathsRegex    *regexp.Regexp
	SecretPattern              *regexp.Regexp

	// Utility patterns
	FindFilesRegex *regexp.Regexp
}

// NewSecretDetectorPatterns initializes and returns SecretDetectorPatterns with compiled regexes
func NewSecretDetectorPatterns() *SecretDetectorPatterns {
	return &SecretDetectorPatterns{
		HighConfidencePasswordRegex: secretpatterns.GetHighConfidencePasswordRegex(),
		HighConfidenceSecretRegex:   secretpatterns.GetHighConfidenceSecretRegex(),
		CommonEnvVarPatterns:        secretpatterns.GetCommonEnvVarPatterns(),

		MediumConfidencePasswordRegex: secretpatterns.GetMediumConfidencePasswordRegex(),
		UserRegex:                     secretpatterns.GetUserRegex(),
		SensitiveFilesRegex:           secretpatterns.GetSensitiveFilesRegex(),

		LowConfidencePasswordRegex: secretpatterns.GetLowConfidencePasswordRegex(),
		SensitiveFilePathsRegex:    secretpatterns.GetSensitiveFilePathsRegex(),
		SecretPattern:              secretpatterns.GetSecretGenericRegex(),
	}
}
