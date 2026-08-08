//go:build !linux

package bptcommon

import "github.com/shirou/gopsutil/v3/mem"

func CalculateMemoryLimits() (MemoryConfig, error) {
	mem, err := mem.VirtualMemory()
	if err == nil {
		return MemoryConfig{
			TotalSystemRAM: int64(mem.Total),
			GoMemLimit:     int64(0.80 * float64(mem.Total)),
			SemaphoreLimit: int64(0.55 * float64(mem.Total)),
		}, nil
	}

	const gb = 1 << 30
	return MemoryConfig{
		TotalSystemRAM: gb,
		GoMemLimit:     int64(0.80 * gb),
		SemaphoreLimit: int64(0.55 * gb),
	}, nil
}
