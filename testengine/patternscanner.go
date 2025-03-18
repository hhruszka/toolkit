package testengine

import (
	"bufio"
	"os"
)

func ScanWithRegex(text string) []RegexMatches {
	// TODO: it has to iterate through all patterns and return a slice with found patterns and matches
	//       since text line potentially can contain multiple patterns
	var foundMatches []RegexMatches

	for _, pattern := range DefaultPatterns.Get() {
		if matches := pattern.CompiledRegex.FindAllString(text, -1); len(matches) > 0 {
			foundMatches = append(foundMatches, RegexMatches{pattern: &pattern, matches: matches})
		}
	}
	return foundMatches
}

func ScanFileWithRegex(file string) *SecretScanResults {
	f, err := os.Open(file)

	if err != nil && !os.IsNotExist(err) {
		//log.Println(err.Error())
		return nil
	}
	defer func() { _ = f.Close() }()

	// Splits on newlines by default.
	scanner := bufio.NewScanner(f)

	line := 1
	foundSecrets := map[int]Secret{}

	for scanner.Scan() {
		foundMatches := ScanWithRegex(scanner.Text())

		for _, found := range foundMatches {
			for _, match := range found.matches {
				foundSecrets[line] = Secret{SecretType: found.pattern.Name, SecretValue: match, Likelihood: VeryLikely, LineNumber: line}
			}
		}
		line++
	}

	if len(foundSecrets) > 0 {
		return &SecretScanResults{file: file, secrets: foundSecrets}
	} else {
		return nil
	}
}
