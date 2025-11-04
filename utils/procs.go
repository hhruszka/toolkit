//go:build linux

package utils

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ProcessInfo combines information from /proc/[PID]/stat,
// /proc/[PID]/status, and /proc/[PID]/uid_map
type ProcessInfo struct {
	PID          int // Process ID
	UID          int
	Cmd          string
	CmdLine      string
	CmdLineCmd   string
	CmdLineArgs  string
	ProcFileInfo os.FileInfo
}

func NewProcessInfo(pid int) (*ProcessInfo, error) {
	info := &ProcessInfo{PID: pid}

	uid, err := ParseUID(pid)
	if err != nil {
		return nil, err
	}
	info.UID = uid

	// Read command line
	info.Cmd, err = ParseComm(pid)
	if err != nil {
		return nil, err
	}

	// Read command line
	info.CmdLine, info.CmdLineCmd, info.CmdLineArgs, err = ParseCmdLine(pid)
	if err != nil {
		return nil, err
	}

	info.ProcFileInfo, err = FileInfo(info.CmdLineCmd)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func ParseComm(pid int) (string, error) {
	comm, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "comm"))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(comm)), nil
}

func ParseCmdLine(pid int) (string, string, string, error) {
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return "", "", "", err
	}

	cmdLine := strings.ReplaceAll(string(data), "\x00", " ")
	cmdLineTokens := strings.Fields(cmdLine)

	// Sometimes processes modify comm and cmdline files, so we need to find first executable if present cmdline.
	var (
		shift int
		token string
		cmd   string
		args  string
	)

	for shift, token = range cmdLineTokens {
		if DoesFileExists(token) && IsExecutable(token) {
			cmd = cmdLineTokens[shift]
			args = strings.Join(cmdLineTokens[shift+1:], " ")
			break
		}
	}

	return strings.TrimSpace(cmdLine), cmd, args, nil
}

func FileInfo(filePath string) (os.FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		//fmt.Println("Error:", err)
		return nil, err
	}

	return fileInfo, nil
}

func ParseUID(pid int) (int, error) {
	fd, err := os.Open(filepath.Join("/proc", strconv.Itoa(pid), "status"))
	if err != nil {
		return 0, err
	}
	defer func() { _ = fd.Close() }()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Uid:") {
			parse := strings.Fields(line)
			if len(parse) >= 2 {
				uid, err := strconv.Atoi(parse[1])
				if err != nil {
					return 0, err
				}
				return uid, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("uid not found for %d", pid)
}

func ProcessSnapshot(procs map[int]*ProcessInfo) error {
	files, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			pid, err := strconv.Atoi(file.Name()) // The name is a number (i.e., a PID)
			if err != nil {
				// This is not a process directory so we continue interating over the file/directory list
				continue
			}

			if pid == os.Getpid() {
				// skip our application
				continue
			}

			if _, exists := procs[pid]; exists {
				// already captured process so we continue searching for new ones
				continue
			}

			procInfo, err := NewProcessInfo(pid)
			if err != nil {
				// Log the error but continue iterating, it could be caused by permissions
				// TODO: log error
				continue
			}

			// if it is nil then we will try again later
			if procInfo != nil {
				procs[pid] = procInfo
			}
		}
	}

	return nil
}

func ProcessMonitor(interval int, probing int) (map[int]*ProcessInfo, error) {
	var procs = make(map[int]*ProcessInfo, 50)

	var err error

	if interval > 0 {
		wg := sync.WaitGroup{}
		timer := time.NewTimer(time.Second * time.Duration(interval))
		ticker := time.NewTicker(time.Millisecond * time.Duration(probing))

		wg.Add(1)

		go func() {
			defer wg.Done()

			stop := false
			for !stop {
				select {
				case <-timer.C:
					ticker.Stop()
					stop = true
					break
				case <-ticker.C:
					if err = ProcessSnapshot(procs); err != nil {
						ticker.Stop()
						timer.Stop()
						stop = true
						break
					}
				}
			}
		}()

		wg.Wait()
	} else {
		err = ProcessSnapshot(procs)
	}

	return procs, err
}
