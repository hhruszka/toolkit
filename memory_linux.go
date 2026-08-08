//go:build linux

package bptcommon

import (
	"bytes"
	"errors"
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// CalculateMemoryLimits determines the optimal memory bounds for the application.
func CalculateMemoryLimits() (MemoryConfig, error) {
	totalRAM := getSystemMemory()

	config := MemoryConfig{
		TotalSystemRAM: totalRAM,
		GoMemLimit:     int64(float64(totalRAM) * 0.80),
		SemaphoreLimit: int64(float64(totalRAM) * 0.55),
	}

	rss, err := getRSS()
	if err != nil {
		return config, err
	}
	config.AppMemoryFootprint = int64(rss)

	return config, nil
}

// getContainerOrHostMemory attempts to read cgroup limits first,
// then falls back to host memory.
func getSystemMemory() int64 {
	// 1. Try Cgroups v2 (Modern Docker / K8s)
	if data, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		value := strings.TrimSpace(string(data))
		if value != "max" {
			if limit, err := strconv.ParseInt(value, 10, 64); err == nil {
				return limit
			}
		}
	}

	// 2. Try Cgroups v1 (Older Docker / K8s)
	if data, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		value := strings.TrimSpace(string(data))
		if limit, err := strconv.ParseInt(value, 10, 64); err == nil {
			// Cgroup v1 sets unconstrained memory to a massive number (usually 2^63-1)
			if limit < int64(math.MaxInt64)-1 {
				return limit
			}
		}
	}

	// 3. Fallback: Host System Memory (Bare metal or unconstrained container)
	var info syscall.Sysinfo_t
	if err := syscall.Sysinfo(&info); err == nil {
		// Totalram is in units of info.Unit (usually 4096 bytes)
		return int64(info.Totalram) * int64(info.Unit)
	}

	// 4. Absolute fallback if everything fails (default to 1GB to prevent panics)
	return 1024 * 1024 * 1024
}

var UnexpectedStatMemoryFormatError = errors.New("unexpected format in /proc/self/statm")

// getLinuxRSS retrieves the resident set size (RSS) memory used by the current process in bytes on Linux systems.
// It reads from the "/proc/self/statm" file and multiplies the RSS value (in pages) by the OS page size.
func getRSS() (uint64, error) {
	data, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0, err
	}

	fields := bytes.Fields(data)
	if len(fields) < 2 {
		return 0, UnexpectedStatMemoryFormatError
	}

	rssPages, err := strconv.ParseUint(string(fields[1]), 10, 64)
	if err != nil {
		return 0, err
	}

	pageSize := uint64(os.Getpagesize())
	rssBytes := rssPages * pageSize

	return rssBytes, nil
}
