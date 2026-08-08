package bptcommon

import (
	"math"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// MemoryConfig holds the calculated limits in bytes
type MemoryConfig struct {
	TotalSystemRAM int64
	GoMemLimit     int64 // 80% of total
	SemaphoreLimit int64 // 55% of total
}

// CalculateMemoryLimits determines the optimal memory bounds for the application.
func CalculateMemoryLimits() (MemoryConfig, error) {
	totalRAM := getSystemMemory()

	config := MemoryConfig{
		TotalSystemRAM: totalRAM,
		GoMemLimit:     int64(float64(totalRAM) * 0.80),
		SemaphoreLimit: int64(float64(totalRAM) * 0.55),
	}

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
