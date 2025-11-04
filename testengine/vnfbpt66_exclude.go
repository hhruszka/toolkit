package testengine

import (
	"fmt"
	"go.uber.org/zap"
	"regexp"
)

var excludeEnvVars []*regexp.Regexp = []*regexp.Regexp{
	regexp.MustCompile(`(?i)^NSS_WRAPPER_PASSWD$`),
	regexp.MustCompile(`(?i)^HOME$`),
	regexp.MustCompile(`(?i)^PATH$`),
	regexp.MustCompile(`(?i)^(?:PWD|OLDPWD)$`),
	regexp.MustCompile(`(?i)^USER$`),
	regexp.MustCompile(`(?i)^LANG$`),
	regexp.MustCompile(`(?i)^HISTFILESIZE$`),
	regexp.MustCompile(`(?i)^SHELL$`),
	regexp.MustCompile(`(?i)^MAIL$`),
	regexp.MustCompile(`(?i).*_?ADDR`),
	regexp.MustCompile(`(?i).*_?REALM$`),
	regexp.MustCompile(`(?i).*_?RESOURCE$`),
	regexp.MustCompile(`(?i).*_?REQUIRED$`),
	regexp.MustCompile(`(?i).*_?(?:FLAG|SWITCH)$`),
	regexp.MustCompile(`(?i).*_?SIZE$`),
	regexp.MustCompile(`(?i).*_?LEVEL$`),
	regexp.MustCompile(`(?i).*_?CAPACITY$`),
	regexp.MustCompile(`(?i).*_?VALIDITY$`),
	regexp.MustCompile(`(?i).*_?HOST$`),
	regexp.MustCompile(`(?i).*_?PREFIX$`),
	regexp.MustCompile(`(?i).*_?FQDN$`),
	regexp.MustCompile(`(?i).*LOGIN.*BANNER.*`),
	regexp.MustCompile(`(?i).*_?UUID$`),
	regexp.MustCompile(`(?i).*SECRET.*NAME$`),
	regexp.MustCompile(`(?i)^__doozer_key$`),
	regexp.MustCompile(`(?i).*WELCOME.*(?:MSG|MESSAGE).*`),
	regexp.MustCompile(`(?i).*(?:USR|USER).*(?:DIR|PATH|FOLDER)$`),
	regexp.MustCompile(`(?i)^SERVICE_ACCOUNT$`),
	regexp.MustCompile(`(?i).*SECRET_NAME$`),
}

func isExcludedEnvVar(envVar string) bool {
	for _, re := range excludeEnvVars {
		if re.MatchString(envVar) {
			zap.L().Debug("Skipping excluded env var", zap.String("envVar", envVar))
			return true
		}
	}
	return false
}

func AddExcludedEnvVars(envVars []string) {
	for _, envVar := range envVars {
		if isExcludedEnvVar(envVar) {
			continue
		}
		excludeEnvVars = append(excludeEnvVars, regexp.MustCompile(fmt.Sprintf("%s", envVar)))
	}
}
