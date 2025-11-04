package testengine

import (
	"os"
	"regexp"
	"strings"
)

// convertSyntax converts custom variable syntax $(VAR) to ${VAR}
// so that os.Expand can recognize it.
func convertSyntax(input string) string {
	// Use a regular expression to match the $(VAR) pattern.
	re := regexp.MustCompile(`\$\((\w+)\)`)
	return re.ReplaceAllString(input, `$${$1}`)
}

// ExpandEnvMap recursively expands variables in a map.
// Given a map of variable names to values, where some values may reference
// other variables (using either ${VAR} or $(VAR) syntax),
// it returns a new map with fully expanded values.
func ExpandEnvMap(vars map[string]string) map[string]string {
	expanded := make(map[string]string)

	// Recursive helper to expand a single value.
	var expand func(string) string
	expand = func(s string) string {
		// First convert our custom syntax.
		s = convertSyntax(s)
		// Use os.Expand with a lookup function that looks up variables
		// in our map (recursively).
		result := os.Expand(s, func(key string) string {
			// If we've already expanded the variable, return it.
			if val, ok := expanded[key]; ok {
				return val
			}
			// Otherwise, if it exists in our original map, expand it.
			if v, ok := vars[key]; ok {
				res := expand(v)
				// Cache the expanded result.
				expanded[key] = res
				return res
			}
			// If not found, return an empty string.
			return ""
		})
		return result
	}

	// Expand every variable.
	for key, value := range vars {
		expanded[key] = expand(value)
	}
	return expanded
}

// getEnvVars extracts environment variables from the execution status of a dependent test case and returns them as a map.
func GetEnvVars(env []string) map[string]string {
	envVars := make(map[string]string)
	for _, str := range env {
		varName, varValue, found := strings.Cut(str, "=")
		if found {
			envVars[varName] = varValue
		}
	}

	//user, ok := envVars["USER"]
	//if !ok {
	//	user = os.Getenv("USER")
	//	_ = user
	//}

	return ExpandEnvMap(envVars)
}
