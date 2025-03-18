package utils

import (
	"net/http"
	"os"
	"strings"
)

// const MAX_FILE_SIZE = 512 * 1024
const MAX_FILE_SIZE = 512

func readSampleOfFile(fileName string) ([]byte, error) {
	var fileData []byte

	fileInfo, err := os.Stat(fileName)
	if err != nil {
		return nil, err
	}

	if fileInfo.Size() >= MAX_FILE_SIZE {
		fileData = make([]byte, MAX_FILE_SIZE)
	}
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = file.Read(fileData)
	if err != nil {
		return nil, err
	}
	return fileData, nil
}

func IsPlainTextFile(fileName string) bool {
	data, err := readSampleOfFile(fileName)
	if err != nil {
		return false
	}
	mimeType := http.DetectContentType(data)
	return strings.Contains(mimeType, "text/plain")
}
