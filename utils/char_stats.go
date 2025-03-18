package utils

import (
	"strings"
	"unicode"
)

type CharStatistics struct {
	Lowers  int `json:"Lowers"`
	Uppers  int `json:"Uppers"`
	Digits  int `json:"Digits"`
	Symbols int `json:"Symbols"`
}

// https://owasp.org/www-community/password-special-characters
var symbols = " !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"

func (s CharStatistics) len() int {
	return s.Digits + s.Uppers + s.Lowers + s.Symbols
}

// CharStats
func CharStats(word string) CharStatistics {
	var stats CharStatistics
	for _, char := range word {
		switch {
		case unicode.IsLower(char):
			stats.Lowers++
		case unicode.IsUpper(char):
			stats.Uppers++
		case unicode.IsDigit(char):
			stats.Digits++
		case strings.ContainsRune(symbols, char):
			stats.Symbols++
		default:
			continue
		}
	}

	return stats
}
