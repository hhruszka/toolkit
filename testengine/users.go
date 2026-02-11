package testengine

import (
	"bptvnftester/log"
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"
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

const DEFAULT_MIN_USER_UID = 1000

// getMinUID parses the /etc/login.defs file to find the minimum user ID.
func getMinUID() int {
	file, err := os.Open("/etc/login.defs")
	if err != nil {
		return DEFAULT_MIN_USER_UID
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 2 && fields[0] == "UID_MIN" {
			minUID, err := strconv.Atoi(fields[1])
			if err != nil {
				return DEFAULT_MIN_USER_UID
			}
			return minUID
		}
	}

	if err := scanner.Err(); err != nil {
		return DEFAULT_MIN_USER_UID
	}

	return DEFAULT_MIN_USER_UID
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

func getUsers() map[string]string {
	var users map[string]string = make(map[string]string)

	passwd, err := getent()
	if err != nil {
		return users
	}

	// Create a scanner to read the file
	scanner := bufio.NewScanner(passwd)

	// Iterate through each line
	for scanner.Scan() {
		// Split the line into fields separated by colons
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			log.Logln("Wrong format of /etc/passwd")
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
		log.Fatal("Error reading /etc/passwd:", err)
	}

	return users
}
