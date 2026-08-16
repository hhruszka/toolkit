package testengine

import (
	"bptvnftester/utils"
	"os"
	"strings"

	core "github.com/hhruszka/secretscanner/regex"
	"go.uber.org/zap"
)

type FilePathValidator interface {
	CheckReadability(path string) bool
	FindFilePaths(value string) []string
}

type FileSystemFilePathValidator struct {
}

func NewFileSystemFilePathValidator() *FileSystemFilePathValidator {
	return &FileSystemFilePathValidator{}
}

func (f *FileSystemFilePathValidator) CheckReadability(path string) bool {
	if path == "" {
		return false
	}
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return false
	}
	_ = file.Close()

	return true
}

func (f *FileSystemFilePathValidator) FindFilePaths(value string) []string {
	return utils.FindFilePaths(value)
}

// DetectionResult holds the results of secret detection
type DetectionResult struct {
	Secrets  []*SecretDetectorResult
	HasError bool
}

// SecretDetector handles detection of secrets in environment variables
type SecretDetector struct {
	scanner              *core.Scanner
	patterns             *SecretDetectorPatterns
	fileValidator        FilePathValidator
	isExcludedEnvVarFunc func(name string) bool
	logger               *zap.Logger
	loggerPrefix         string
}

// SecretDetectorOption defines a functional option for configuring a SecretDetector instance.
type SecretDetectorOption func(*SecretDetector)

// NewSecretDetector creates a new SecretDetector instance
func NewSecretDetector(options ...SecretDetectorOption) *SecretDetector {
	patterns := NewSecretDetectorPatterns()
	fileValidator := NewFileSystemFilePathValidator()
	sd := &SecretDetector{
		scanner:       core.NewScanner(),
		patterns:      patterns,
		fileValidator: fileValidator,
		logger:        zap.NewNop(),
		loggerPrefix:  "SecretDetector",
		isExcludedEnvVarFunc: func(name string) bool {
			return false
		},
	}
	for _, option := range options {
		option(sd)
	}
	return sd
}

// WithSecretScanner sets the secret scanner to use for advanced detection
func WithSecretScanner(scanner *core.Scanner) SecretDetectorOption {
	return func(sd *SecretDetector) {
		sd.scanner = scanner
	}
}

// WithSecretDetectorPatterns sets the secret detector patterns to use for detection
func WithSecretDetectorPatterns(patterns *SecretDetectorPatterns) SecretDetectorOption {
	return func(sd *SecretDetector) {
		sd.patterns = patterns
	}
}

// WithFilePathValidator sets the file path validator to use for detection
func WithFilePathValidator(fileValidator FilePathValidator) SecretDetectorOption {
	return func(sd *SecretDetector) {
		sd.fileValidator = fileValidator
	}
}

// WithLogger sets the logger to use for detection
func WithLogger(logger *zap.Logger, prefix string) SecretDetectorOption {
	return func(sd *SecretDetector) {
		sd.logger = logger
		if prefix != "" {
			sd.loggerPrefix = prefix
		}
	}
}

func WithExcludeEnvVarFunc(excludeFunc func(name string) bool) SecretDetectorOption {
	return func(sd *SecretDetector) {
		sd.isExcludedEnvVarFunc = excludeFunc
	}
}

// DetectSecrets analyzes environment variables for potential secrets
func (s *SecretDetector) DetectSecrets(envVars map[string]string) *DetectionResult {
	result := &DetectionResult{
		Secrets: make([]*SecretDetectorResult, 0),
	}

	for name, value := range envVars {
		if s.isExcludedEnvVarFunc(name) {
			s.logger.Info("Skipping env. variable", zap.String("name", name), zap.String("value", value))
			continue
		}

		if secrets := s.detectInVariable(name, value); len(secrets) > 0 {
			s.logger.Info("Detected secrets in env. variable", zap.String("name", name), zap.Int("found", len(secrets)))
			result.Secrets = append(result.Secrets, secrets...)
			result.HasError = true
		}
	}
	//for _, secret := range result.Secrets {
	//	fmt.Printf("%+v", secret)
	//}
	s.logger.Info("Finished detecting secrets", zap.Int("found", len(result.Secrets)))
	return result
}

// detectInVariable performs secret detection on a single environment variable
func (s *SecretDetector) detectInVariable(name, value string) []*SecretDetectorResult {
	//var secrets []string

	s.logger.Debug(s.loggerPrefix, zap.String("name", name), zap.String("value", value))
	// Check common secret env var names
	if detected := s.checkCommonSecretNames(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkCommonSecretNames"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check password regex patterns
	if detected := s.checkHighConfidencePasswordPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkCommonSecretNames"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check secret regex patterns
	if detected := s.checkHighConfidenceSecretPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkCommonSecretNames"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check password regex patterns
	if detected := s.checkMediumConfidencePasswordPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkMediumConfidencePasswordPatterns"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check user patterns
	if detected := s.checkUserPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkUserPatterns"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check sensitive file extensions
	if detected := s.checkSensitiveFilesPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkSensitiveFilesPatterns"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check sensitive file paths
	if detected := s.checkSensitiveFilePaths(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkSensitiveFilePaths"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check secret patterns
	if detected := s.checkSecretPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkSecretPatterns"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Check secret patterns
	if detected := s.checkLowConfidencePasswordPatterns(name, value); detected != nil {
		s.logger.Debug(s.loggerPrefix, zap.String("check", "checkLowConfidencePasswordPatterns"), zap.String("name", name), zap.String("value", value))
		return []*SecretDetectorResult{detected}
	}

	// Use secret scanner for advanced detection
	return s.scanWithSecretScanner(name, value)
}

// checkCommonSecretNames checks if the variable name matches common secret patterns
func (s *SecretDetector) checkCommonSecretNames(name, value string) *SecretDetectorResult {
	for _, re := range s.patterns.CommonEnvVarPatterns {
		if re.MatchString(name) {
			//s.logger.Info("detected secret", zap.String("name", name), zap.String("value", value))
			return NewSecretDetectorResult(name, value, false, false, 4, "High confidence common secret variable name")
		}
	}

	return nil
}

// checkHighConfidencePasswordPatterns identifies potential high-confidence password patterns in environment variable names.
// Validates whether the variable value appears to be a single token and a potential secret.
// Marks results where values contain readable file paths for added context in detection.
func (s *SecretDetector) checkHighConfidencePasswordPatterns(name, value string) *SecretDetectorResult {
	if !s.patterns.HighConfidencePasswordRegex.MatchString(name) {
		return nil
	}

	isLine := len(strings.Fields(value)) > 1
	isMultiLine := len(strings.Split(value, "\n")) > 1
	isToken := !isLine && !isMultiLine
	confidence := 4

	if !isToken {
		return nil
	}

	var secret *SecretDetectorResult
	if s.patterns.UserRegex.MatchString(name) {
		secret = NewSecretDetectorResult(name, value, false, false, 3, "Username env. variable name")
	} else {
		if !IsSecret(value) {
			confidence = 3
		}
		secret = NewSecretDetectorResult(name, value, false, false, confidence, "High confidence password env. variable name")
	}

	files := s.fileValidator.FindFilePaths(value)
	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false

	for _, file := range files {
		if s.fileValidator.CheckReadability(file) {
			readableFlag = true
			valueString.WriteString("\nReadable: ")
			valueString.WriteString(file)
		} else {
			valueString.WriteString("\nNot readable: ")
			valueString.WriteString(file)
		}
	}

	if len(files) > 0 {
		secret.value = valueString.String()
		secret.readable = readableFlag
		secret.file = true
		confidence = 2
		if readableFlag {
			confidence = 3
		}
		secret.sectype = "Sensitive file path in env. variable value"
	}

	return secret
}

// checkHighConfidenceSecretPatterns checks if the environment variable name matches high-confidence secret patterns.
// Evaluates the variable value to confirm it is a token and not multi-line or a list.
// Returns a SecretDetectorResult if the value is identified as a potential secret.
// Validates if the value contains readable file paths, marking them accordingly.
func (s *SecretDetector) checkHighConfidenceSecretPatterns(name, value string) *SecretDetectorResult {
	if !s.patterns.HighConfidenceSecretRegex.MatchString(name) {
		return nil
	}

	isLine := len(strings.Fields(value)) > 1
	isMultiLine := len(strings.Split(value, "\n")) > 1
	isToken := !isLine && !isMultiLine
	confidence := 4

	if !isToken {
		return nil
	}

	if !IsSecret(value) {
		confidence = 3
	}
	secret := NewSecretDetectorResult(name, value, false, false, confidence, " High confidence secret env. variable name")

	files := s.fileValidator.FindFilePaths(value)
	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false

	for _, file := range files {
		if s.fileValidator.CheckReadability(file) {
			readableFlag = true
			valueString.WriteString("\nReadable: ")
			valueString.WriteString(file)
		} else {
			valueString.WriteString("\nNot readable: ")
			valueString.WriteString(file)
		}
	}

	if len(files) > 0 {
		secret.value = valueString.String()
		secret.readable = readableFlag
		secret.file = true
		secret.confidence = 2
		if readableFlag {
			secret.confidence = 3
		}
		secret.sectype = "Sensitive file path in env. variable value"
	}

	return secret
}

// checkMediuConfidencePasswordPatterns checks if the variable name matches medium-confidence password patterns.
// Validates whether the variable value appears to be a single token and a potential secret.
// Marks results where values contain readable file paths for added context in detection.
// Returns a SecretDetectorResult if a medium-confidence password pattern is identified.
func (s *SecretDetector) checkMediumConfidencePasswordPatterns(name, value string) *SecretDetectorResult {
	if !s.patterns.MediumConfidencePasswordRegex.MatchString(name) {
		s.logger.Debug("No match", zap.String("check", "checkMediumConfidencePasswordPatterns"), zap.String("name", name), zap.String("value", value))
		return nil
	}

	isLine := len(strings.Fields(value)) > 1
	isMultiLine := len(strings.Split(value, "\n")) > 1
	isToken := !isLine && !isMultiLine
	confidence := 3

	if !isToken {
		return nil
	}
	if !IsSecret(value) {
		confidence = 2
	}

	var secret *SecretDetectorResult
	if s.patterns.UserRegex.MatchString(name) {
		secret = NewSecretDetectorResult(name, value, false, false, 3, "Username env. variable name")
	} else {
		secret = NewSecretDetectorResult(name, value, false, false, confidence, "Medium confidence password env. variable name")
	}

	files := s.fileValidator.FindFilePaths(value)
	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false

	for _, file := range files {
		if s.fileValidator.CheckReadability(file) {
			readableFlag = true
			valueString.WriteString("\nReadable: ")
			valueString.WriteString(file)
		} else {
			valueString.WriteString("\nNot readable: ")
			valueString.WriteString(file)
		}
	}

	if len(files) > 0 {
		secret.value = valueString.String()
		secret.readable = readableFlag
		secret.file = true
		secret.confidence = 2
		if readableFlag {
			secret.confidence = 3
		}
		secret.sectype = "Sensitive file path in env. variable value"
	}
	return secret
}

// checkMediuConfidencePasswordPatterns checks if the variable name matches medium-confidence password patterns.
// Validates whether the variable value appears to be a single token and a potential secret.
// Marks results where values contain readable file paths for added context in detection.
// Returns a SecretDetectorResult if a medium-confidence password pattern is identified.
func (s *SecretDetector) checkLowConfidencePasswordPatterns(name, value string) *SecretDetectorResult {
	if !s.patterns.LowConfidencePasswordRegex.MatchString(name) {
		return nil
	}

	isLine := len(strings.Fields(value)) > 1
	isMultiLine := len(strings.Split(value, "\n")) > 1
	isToken := !isLine && !isMultiLine

	if !isToken {
		return nil
	}

	if !IsSecret(value) {
		return nil
	}

	var secret *SecretDetectorResult
	if s.patterns.UserRegex.MatchString(name) {
		secret = NewSecretDetectorResult(name, value, false, false, 3, "Username env. variable name")
	} else {
		secret = NewSecretDetectorResult(name, value, false, false, 2, "Low confidence password env. variable name")
	}

	files := s.fileValidator.FindFilePaths(value)
	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false

	for _, file := range files {
		if s.fileValidator.CheckReadability(file) {
			readableFlag = true
			valueString.WriteString("\nReadable: ")
			valueString.WriteString(file)
		} else {
			valueString.WriteString("\nNot readable: ")
			valueString.WriteString(file)
		}
	}

	if len(files) > 0 {
		secret.value = valueString.String()
		secret.readable = readableFlag
		secret.file = true
		secret.confidence = 2
		if readableFlag {
			secret.confidence = 3
		}
		secret.sectype = "Sensitive file path in env. variable value"
	}
	return secret
}

// checkUserPatterns inspects the provided variable name against user-defined patterns to identify usernames as environment variables.
func (s *SecretDetector) checkUserPatterns(name, value string) *SecretDetectorResult {
	if s.patterns.UserRegex.MatchString(name) && IsSecret(value) {
		return NewSecretDetectorResult(name, value, false, false, 3, "Username env. variable name")
	}
	return nil
}

// checkSensitiveFiles checks if the value contains sensitive file extensions
func (s *SecretDetector) checkSensitiveFilesPatterns(name, value string) *SecretDetectorResult {
	// Check if it's a file path first (cheap operation)
	filePaths := s.fileValidator.FindFilePaths(value)
	if filePaths == nil || len(filePaths) == 0 {
		return nil
	}

	var sensitiveFiles []string
	for _, filePath := range filePaths {
		if s.patterns.SensitiveFilesRegex.MatchString(filePath) {
			sensitiveFiles = append(sensitiveFiles, filePath)
		}
	}

	// Only then check the regex pattern (expensive operation)
	if sensitiveFiles == nil || len(sensitiveFiles) == 0 {
		return nil
	}

	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false
	confidence := 3

	for _, file := range sensitiveFiles {
		if s.patterns.SensitiveFilesRegex.MatchString(file) {
			if s.fileValidator.CheckReadability(file) {
				readableFlag = true
				valueString.WriteString("\nReadable: ")
				valueString.WriteString(file)
			} else {
				confidence = 2
				valueString.WriteString("\nNot readable: ")
				valueString.WriteString(file)
			}
		}
	}

	return NewSecretDetectorResult(
		name, valueString.String(),
		true,
		readableFlag,
		confidence,
		"Sensitive file in env. variable value",
	)
}

// checkSensitiveFilePaths checks if the value contains sensitive file paths
func (s *SecretDetector) checkSensitiveFilePaths(name, value string) *SecretDetectorResult {
	// Check if it's a file path first (cheap operation)
	filePaths := s.fileValidator.FindFilePaths(value)
	if filePaths == nil || len(filePaths) == 0 {
		return nil
	}

	var sensitiveFilePaths []string
	for _, filePath := range filePaths {
		if s.patterns.SensitiveFilePathsRegex.MatchString(filePath) {
			sensitiveFilePaths = append(sensitiveFilePaths, filePath)
		}
	}

	// Only then check the regex pattern (expensive operation)
	if sensitiveFilePaths == nil || len(sensitiveFilePaths) == 0 {
		return nil
	}

	valueString := strings.Builder{}
	valueString.WriteString(value)
	readableFlag := false
	confidence := 2

	for _, file := range sensitiveFilePaths {
		if s.patterns.SensitiveFilesRegex.MatchString(file) {
			if s.fileValidator.CheckReadability(file) {
				confidence = 3
				readableFlag = true
				valueString.WriteString("\nReadable: ")
				valueString.WriteString(file)
			} else {
				valueString.WriteString("\nNot readable: ")
				valueString.WriteString(file)
			}
		}
	}

	return NewSecretDetectorResult(
		name, valueString.String(),
		true,
		readableFlag,
		confidence,
		"Sensitive file path in env. variable value",
	)
}

// checkSecretPatterns checks if the environment variable name matches secret patterns defined in the SecretDetector.
func (s *SecretDetector) checkSecretPatterns(name, value string) *SecretDetectorResult {
	if s.patterns.SecretPattern.MatchString(name) {
		//if s.patterns.UserRegex.MatchString(name) {
		//	return NewSecretDetectorResult(name, value, false, false, 3, "Username env. variable name")
		//}
		if !IsSecret(value) {
			return nil
		}
		return NewSecretDetectorResult(name, value, false, false, 2, "Low confidence secret env. variable name")
	}
	return nil
}

// scanWithSecretScanner uses the advanced secret scanner for detection
func (s *SecretDetector) scanWithSecretScanner(name, value string) []*SecretDetectorResult {
	var secrets []*SecretDetectorResult

	isLine := len(strings.Fields(value)) > 1
	isMultiLine := len(strings.Split(value, "\n")) > 1
	isToken := !isLine && !isMultiLine

	if !isMultiLine && !IsSecret(value) {
		return nil
	}

	if isToken || isLine {
		valueMatches := s.scanner.ScanLine([]byte(value))
		if valueMatches != nil && len(valueMatches) > 0 {
			for _, valueMatch := range valueMatches {
				if valueMatch != nil && valueMatch.Pattern.Confidence >= 3 {
					foundSecret := valueMatch.ToSecret(0, []byte(value))
					if foundSecret != nil {
						s := NewSecretDetectorResult(
							name,
							value,
							false,
							false,
							foundSecret.Confidence,
							foundSecret.SecretType,
						)
						secrets = append(secrets, s)
					}
				}
			}
		}
	}

	if isMultiLine {
		valueMatches, _ := s.scanner.ScanLines([]byte(value), "")
		if valueMatches != nil && len(valueMatches) > 0 {
			for _, valueMatch := range valueMatches {
				if valueMatch != nil && valueMatch.Confidence >= 3 {
					s := NewSecretDetectorResult(
						name,
						value,
						false,
						false,
						valueMatch.Confidence,
						valueMatch.SecretType,
					)
					secrets = append(secrets, s)
				}
			}
		}
	}

	var secType string
	if len(secrets) > 0 {
		secType = secrets[0].Type()
	}
	s.logger.Debug("scanWithSecretScanner", zap.String("name", name), zap.String("value", value), zap.String("type", secType))
	return secrets
}
