package sudo

import (
	"regexp"
	"strings"
)

// SudoCommand represents a single executable permission
type SudoCommand struct {
	RunAsUser  string
	RunAsGroup string // Optional, usually irrelevant for security audits but good to have
	NeedsPass  bool
	Command    string
}

// ParseSudoOutput parses the output of a sudoers file or command for executable permissions and returns a list of SudoCommand.
func ParseSudoOutput(output string) []SudoCommand {
	var results []SudoCommand
	lines := strings.Split(output, "\n")

	// Strict Regex: Must start with indentation followed by opening parenthesis
	// Example matches: "    (root)", "    (ALL : ALL)"
	reRunAs := regexp.MustCompile(`^\s+\(\s*([^: )]+)(?:\s*:\s*([^)]+))?\s*\)`)

	for _, line := range lines {
		// 1. Strict Filter: If the line does not match the (User) pattern, SKIP IT.
		// This effectively filters out "Matching Defaults..." headers and config lines.
		runAsMatch := reRunAs.FindStringSubmatch(line)
		if runAsMatch == nil {
			continue
		}

		// 2. Extract RunAs User (Group 1)
		runAsUser := runAsMatch[1]

		// 3. Extract the "Body"
		// Remove the (User) prefix to get the clean command list
		body := reRunAs.ReplaceAllString(line, "")
		body = strings.TrimSpace(body)

		// 4. Stateful Parsing (PASSWD vs NOPASSWD)
		currentNeedsPass := true // Defaults to PASSWD
		tokens := strings.Split(body, ",")

		for _, token := range tokens {
			token = strings.TrimSpace(token)

			// Handle State Switches
			if strings.HasPrefix(token, "NOPASSWD:") {
				currentNeedsPass = false
				token = strings.TrimPrefix(token, "NOPASSWD:")
			} else if strings.HasPrefix(token, "PASSWD:") {
				currentNeedsPass = true
				token = strings.TrimPrefix(token, "PASSWD:")
			}

			cmd := strings.TrimSpace(token)
			if cmd == "" {
				continue
			}

			results = append(results, SudoCommand{
				RunAsUser: runAsUser,
				NeedsPass: currentNeedsPass,
				Command:   cmd,
			})
		}
	}

	return results
}
