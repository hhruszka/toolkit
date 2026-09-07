package pool

import "sync"

type BufferPool struct {
	pool        sync.Pool
	keepLimit   int
	wasteFactor float64
}

// NewBufferPool initializes and returns a new BufferPool with buffers capped at the specified limit.
func NewBufferPool(max int, maxWasteFactor float64) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() any {
				return make([]byte, 0)
			},
		},
		keepLimit:   max,
		wasteFactor: maxWasteFactor,
	}
}

// Capacity returns the maximum capacity of buffers allocated from the pool
func (bp *BufferPool) Capacity() int {
	return bp.keepLimit
}

// Reset resets the pool, discarding all buffers
func (bp *BufferPool) Reset() {
	bp.pool = sync.Pool{
		New: func() any {
			return make([]byte, 0)
		},
	}
}

// Get retrieves a byte slice from the pool with at least the specified size, creating a new one if necessary.
func (bp *BufferPool) Get(size int) []byte {
	buf := bp.pool.Get().([]byte)

	// Check if it's big enough
	if cap(buf) < size || float64(cap(buf)) > float64(size)*bp.wasteFactor {
		// Too small? Discard it and make a new one.
		buf = make([]byte, 0, size)
	}

	// Return a slice with length 0, ready to be appended to or Read into
	return buf[:0]
}

// Put returns a byte slice back to the pool if its capacity is within the pool's limit. Otherwise, it is discarded.
func (bp *BufferPool) Put(b []byte) {
	if cap(b) > bp.keepLimit {
		return
	}
	bp.pool.Put(b[:0])
}
