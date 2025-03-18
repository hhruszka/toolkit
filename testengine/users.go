package testengine

import (
	"bufio"
	"linuxtester/log"
	"os"
	"strconv"
	"strings"
)

func contains(str string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(str, p) {
			return true
		}
	}
	return false
}

func GetUsers() []string {
	var users []string
	var uid int

	// Open the /etc/passwd file
	file, err := os.Open("/etc/passwd")
	if err != nil {
		log.Fatal("Error opening /etc/passwd:", err)
	}
	defer file.Close()

	// Create a scanner to read the file
	scanner := bufio.NewScanner(file)

	// Iterate through each line
	for scanner.Scan() {
		// Split the line into fields separated by colons
		fields := strings.Split(scanner.Text(), ":")
		if len(fields) < 7 {
			log.Logln("Wrong format of /etc/passwd")
			continue
		}

		uid, err = strconv.Atoi(fields[2])
		if err != nil {
			log.Logln("Error converting uid to int:", err)
			continue
		}
		if uid >= 1000 && contains(fields[6], "/bash", "/ksh", "/zsh", "/csh", "/fish", "/tcsh", "/dash", "/elvish", "/ion", "/xonsh", "/nushell", "/sh", "/wsh") {
			// The first field is the username
			users = append(users, fields[0])
		}
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading /etc/passwd:", err)
	}

	return users
}
