package bptcommon

import (
	"context"
	"errors"

	"golang.org/x/sync/semaphore"
)

type MemoryBank struct {
	memoryBankLimit int64
	*semaphore.Weighted
}

var memoryBank *MemoryBank

// FileExceedsMemoryLimitError is returned when a file exceeds the memory limit.
var RequestExceedsMemoryLimitError = errors.New("file exceeds memory limit")
var NegativeMemoryRequestError = errors.New("request for negative memory")

// AcquireMemory is a function that acquires a semaphore for the given amount of memory and returns a function to release it.
func (mb *MemoryBank) AcquireMemory(ctx context.Context, size int64) (func(), error) {
	if size < 0 {
		return func() {}, NegativeMemoryRequestError
	}

	if size == 0 {
		return func() {}, nil
	}

	if size > mb.memoryBankLimit {
		return func() {}, RequestExceedsMemoryLimitError
	}
	err := mb.Acquire(ctx, size)
	if err != nil {
		return func() {}, err
	}
	return func() { mb.Release(size) }, nil
}

// NewMemoryBank initializes a singleton MemoryBank instance with the specified memory limit and returns it.
// If memoryBankLimit is non-positive, it returns nil. If an instance already exists, it returns the existing instance.
func NewMemoryBank(memoryBankLimit int64) *MemoryBank {
	if memoryBankLimit <= 0 {
		return nil
	}

	if memoryBank != nil {
		return memoryBank
	}

	memoryBank = &MemoryBank{
		memoryBankLimit: memoryBankLimit,
		Weighted:        semaphore.NewWeighted(memoryBankLimit),
	}

	return memoryBank
}
