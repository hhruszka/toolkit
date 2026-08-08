package bptcommon

import "sync"

type BufferPool struct {
	pool  sync.Pool
	limit int
}

// NewBufferPool initializes and returns a new BufferPool with buffers capped at the specified limit.
func NewBufferPool(limit int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() any {
				return make([]byte, 0, limit)
			},
		},
		limit: limit,
	}
}

// Capacity returns the maximum capacity of buffers allocated from the pool
func (bp *BufferPool) Capacity() int {
	return bp.limit
}

// Reset resets the pool, discarding all buffers
func (bp *BufferPool) Reset() {
	bp.pool = sync.Pool{
		New: func() any {
			return make([]byte, 0, bp.limit)
		},
	}
}

// Get retrieves a byte slice from the pool with at least the specified size, creating a new one if necessary.
func (bp *BufferPool) Get(size int) []byte {
	// Size of a slice requested size from the pool is bigger than the pool limit
	if size > bp.limit {
		return make([]byte, 0, size)
	}

	v := bp.pool.Get()

	// Pool was empty, create new empty buffer.
	// This is just in case the sync.Pool.New was not provided.
	// If sync.Pool.New is provided, this will never happen.
	if v == nil {
		return make([]byte, 0, size)
	}

	buf := v.([]byte)

	// Check if it's big enough
	if cap(buf) < size {
		// Too small? Discard it and make a new one.
		buf = make([]byte, 0, size)
	}

	// Return a slice with length 0, ready to be appended to or Read into
	return buf[:0]
}

// Put returns a byte slice back to the pool if its capacity is within the pool's limit. Otherwise, it is discarded.
func (bp *BufferPool) Put(b []byte) {
	if cap(b) > bp.limit {
		return
	}
	bp.pool.Put(b)
}
