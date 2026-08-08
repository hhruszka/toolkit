package bptcommon

// MemoryConfig holds the calculated limits in bytes
type MemoryConfig struct {
	TotalSystemRAM     int64
	GoMemLimit         int64 // 80% of total
	SemaphoreLimit     int64 // 55% of total
	AppMemoryFootprint int64 // RSS
}
