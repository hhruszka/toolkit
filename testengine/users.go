package testengine

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func contains(str string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.HasSuffix(str, p) {
			return true
		}
	}
	return false
}

func getent() (io.Reader, error) {
	var passwd bytes.Buffer

	cmd := exec.Command("getent", "passwd")
	cmd.Stdout = &passwd
	if err := cmd.Run(); err == nil {
		return &passwd, nil
	}

	if data, err := os.ReadFile("/etc/passwd"); err == nil {
		return bytes.NewBuffer(data), nil
	}
	return nil, errors.New("failed to retrieve user list")
}

var GetUsers = getUsers

func getUsers() (map[string]string, error) {
	var users map[string]string = make(map[string]string)

	passwd, err := getent()
	if err != nil {
		return nil, err
	}

	// Create a scanner to read the file
	scanner := bufio.NewScanner(passwd)

	// Iterate through each line
	for scanner.Scan() {
		// Split the line into fields separated by colons
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			//log.Logln("Wrong format of /etc/passwd")
			continue
		}

		user := fields[0]
		shell := fields[6]

		if contains(shell, "/bash", "/ksh", "/zsh", "/csh", "/fish", "/tcsh", "/dash", "/elvish", "/ion", "/xonsh", "/nushell", "/sh", "/wsh") {
			// The first field is the username
			if fields[0] == "root" || fields[2] == "0" {
				continue
			}
			users[user] = shell
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading /etc/passwd: %w", err)
	}

	return users, nil
}
