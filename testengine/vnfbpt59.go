//go:build ignore

package testengine

import (
	"bptvnftester/utils"
	"fmt"
	"github.com/hhruszka/secretscanner/checkers"
	"github.com/hhruszka/secretscanner/core"
	"golang.org/x/exp/utf8string"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

var passwordRegex = regexp.MustCompile(`(?i)([^\s]*(?:p[a@]?ssw[0ord]*?|\w+pwd\w+|p[@a]ssw?|pss?wd|secr[et]*?|cert|su?[per]*?use?r|use?r|passphrase)[^\s]*)`)
var sensitiveFilesRegex = regexp.MustCompile(`(?i)\.(pem|crt|cer|der|pfx|p12|cred|cert)$`)
var sensitiveFilePathsRegex = regexp.MustCompile(`(?i).*(?:[\\/](?:config|configs|credential(?:s)?|secret(?:s)?|pass(?:word|wd)?|key(?:s)?|private|cert(?:ificate)?(?:s)?|token|ssl|tls|pki|trust)(?:[\\/]|$)).*`)
var findFiles = regexp.MustCompile(`(?P<filepath>[^\s'"]*\/[^\s'"]+)`)

// PasswordCheckerIsFilePath checks if a word could be a file path
func PasswordCheckerIsFilePath(word string) bool {
	if filepath.IsAbs(word) {
		return true
	}
	return false
}

// PasswordCheckerIsExistingFilePath checks if a word is an existing directory or a file
func PasswordCheckerIsExistingFilePath(word string) bool {
	if filepath.IsAbs(word) {
		if _, err := os.Stat(word); err == nil {
			return true
		}
	}
	return false
}

// PasswordCheckerIsTimeOrDate checks if a word is a time or a date typical for Linux (ML candidate)
func PasswordCheckerIsTimeOrDate(word string) bool {
	for _, reg := range timeRegexes {
		if reg.MatchString(word) {
			return true
		}
	}
	return false
}

// PasswordCheckerIsFlag checks if a word is one of flags
func PasswordCheckerIsFlag(word string) bool {
	flags := []string{"true", "false", "yes", "y", "no", "n", "on", "off"}
	return slices.Contains(flags, strings.ToLower(word))
}

// PasswordCheckerThirteen checks if a word consists only of ASCII characters
func PasswordCheckerIsASCII(word string) bool {
	return utf8string.NewString(word).IsASCII()
}

// PasswordCheckerSymbolsAndDigitsOnly checks if a word consists of digits and symbols only
func PasswordCheckerSymbolsAndDigitsOnly(word string) bool {
	if stats := utils.CharStats(word); stats.Digits > 0 && stats.Symbols >= 0 && stats.Lowers == 0 && stats.Uppers == 0 {
		return true
	}
	return false
}

// PasswordCheckerDigitsOnly checks if a word consists of digits only
func PasswordCheckerDigitsOnly(word string) bool {
	_, err := strconv.Atoi(word)
	return err == nil
}

const MIN_LENGHT = 4

// PasswordCheckerDigitsOnly checks if a word consists of digits only
func PasswordCheckerTooShort(word string) bool {
	return len(word) < MIN_LENGHT
}

// IsSecret checks if a word is a potential secret
func IsSecret(candidate string) bool {
	switch {
	case PasswordCheckerIsTimeOrDate(candidate):
		return false
	case PasswordCheckerIsFlag(candidate):
		return false
	case !PasswordCheckerIsASCII(candidate):
		return false
	case PasswordCheckerSymbolsAndDigitsOnly(candidate):
		return false
	case PasswordCheckerDigitsOnly(candidate):
		return false
	case PasswordCheckerTooShort(candidate):
		return false
	}
	return true
}

// Sensitive patterns
// var secretNamePattern = regexp.MustCompile(`(?i)(secret|passwd|cred|token)`)
var secretPattern = regexp.MustCompile(`(?i)(pswd|p[a@]?ssw?[0ord]*?|secret|cred|credential|token|key)`)

var YesNo = map[bool]string{true: "Yes", false: "No"}

func stringify(lines []string) string {
	return strings.Join(lines, "\n")
}

// cnfbpt49 tests whether environment variables defined as part of a container's environment contains secrets.
func vnfbpt59(testId string, depExecResults map[string]map[string]*ExecutionStatus) *ExecutionStatus {
	var (
		secrets []string
	)

	execTime := time.Now().UTC()

	depID := AllTestCases[testId].Dependencies[0].Id
	varName := AllTestCases[testId].Dependencies[0].VarName
	execStatus := depExecResults[depID][varName]

	runtimeEnv := execStatus.Stdout

	runtimeEnvVars := make(map[string]string)
	for _, str := range runtimeEnv {
		tokens := strings.Split(str, "=")
		if len(tokens) >= 2 {
			key := tokens[0]
			value := strings.Join(tokens[1:], "=")
			runtimeEnvVars[key] = value
		}
	}

	user, ok := runtimeEnvVars["USER"]
	if !ok {
		user = os.Getenv("USER")
		_ = user
	}

	is := core.NewSecretScanner()
	retCode := Success
	envVars := make(map[string]string)

	envVars = ExpandEnvMap(runtimeEnvVars)

	for name, value := range envVars {
		isLine := len(strings.Fields(value)) > 1
		isMultiLine := len(strings.Split(value, "\n")) > 1
		isToken := !isLine && !isMultiLine

		if isExcluded(name) {
			continue
		}
		match := passwordRegex.FindAllString(name, 1)
		if match != nil {
			//fmt.Printf("match %s=%s\n", name, value)
			if isToken && IsSecret(value) {
				secret := fmt.Sprintf("%s=%s", name, value)
				if slices.Contains(secrets, secret) {
					continue
				}
				secrets = append(secrets, secret)

				files := findFiles.FindAllString(value, -1)
				//fmt.Println(len(files), value)
				for _, file := range files {
					if checkers.PasswordCheckerIsFilePath(file) {
						readable := utils.IsReadable(file)
						if readable {
							secrets = append(secrets, fmt.Sprintf("\tReadable: %s", file))
						}
					}
				}
				retCode = GeneralError
				continue
			}
		}

		match = sensitiveFilePathsRegex.FindAllString(value, 1)
		if match != nil && checkers.PasswordCheckerIsFilePath(value) {
			secrets = append(secrets, fmt.Sprintf("%s=%s", name, value))

			files := findFiles.FindAllString(value, -1)
			for _, file := range files {
				if checkers.PasswordCheckerIsFilePath(file) {
					readable := utils.IsReadable(file)
					if readable {
						secrets = append(secrets, fmt.Sprintf("\tReadable: %s", file))
					}
				}
			}
			retCode = GeneralError
			continue
		}

		match = sensitiveFilesRegex.FindAllString(value, 1)
		if match != nil && checkers.PasswordCheckerIsFilePath(value) {
			secrets = append(secrets, fmt.Sprintf("%s=%s", name, value))
			files := findFiles.FindAllString(value, -1)
			for _, file := range files {
				if checkers.PasswordCheckerIsFilePath(file) {
					readable := utils.IsReadable(file)
					if readable {
						secrets = append(secrets, fmt.Sprintf("\tReadable: %s", file))
					}
				}
			}
			retCode = GeneralError
			continue
		}

		if isToken || isLine {
			valueMatches := is.ScanLine([]byte(value))
			if valueMatches != nil && len(valueMatches) > 0 {
				for _, valueMatch := range valueMatches {
					if valueMatch != nil && valueMatch.Pattern.Confidence >= 3 {
						secret := valueMatch.ToSecrets(0, nil)
						secrets = append(secrets, fmt.Sprintf("%s: %s=%s", secret.SecretType, name, secret.SecretValue))
						retCode = GeneralError
					}
				}
			}
			continue
		}

		if isMultiLine {
			_, valueMatches := is.ScanFile("", []byte(value))
			if valueMatches != nil && len(valueMatches) > 0 {
				for _, valueMatch := range valueMatches {
					if valueMatch != nil && valueMatch.Confidence >= 3 {
						secrets = append(secrets, fmt.Sprintf("%s: %s=%s", valueMatch.SecretType, name, valueMatch.SecretValue))
					}
				}
				retCode = GeneralError
			}
		}
	}

	if len(secrets) > 0 {
		retCode = GeneralError
	}
	return NewExecutionStatus(retCode, "", strings.Join(runtimeEnv, "\n"), strings.Join(secrets, "\n"), execTime)
}
