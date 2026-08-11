package report

import "bytes"

// WrapText combines a slice of strings into a single string and inserts line breaks every n characters in each string.
func WrapText(text []string, n int) string {
	return wrapText(text, n)
}

// wrapText formats a slice of strings into a single string with line breaks every n characters per string.
func wrapText(text []string, n int) string {
	var buffer bytes.Buffer

	for _, str := range text {
		wrapLen := n - 1
		strLen := len(str) - 1
		for idx, c := range str {
			buffer.WriteRune(c)
			if idx%n == wrapLen && idx != strLen {
				buffer.WriteRune('\n')
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}
