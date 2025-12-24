package performance

import (
	"sync"
	"sync/atomic"
)

// BufferPool manages a pool of reusable byte buffers to reduce garbage collection pressure.
// Using sync.Pool for buffers is essential for high-throughput applications like Modbus proxies.
type BufferPool struct {
	pool      sync.Pool
	size      int
	gets      uint64
	puts      uint64
	allocated uint64
}

// NewBufferPool creates a new buffer pool with the specified buffer size.
// Common sizes for Modbus:
//   - 260 bytes: Maximum Modbus TCP ADU size
//   - 512 bytes: Safe size for any Modbus frame + overhead
//   - 1024 bytes: For buffering multiple frames
func NewBufferPool(size int) *BufferPool {
	bp := &BufferPool{
		size: size,
	}

	bp.pool.New = func() interface{} {
		atomic.AddUint64(&bp.allocated, 1)
		return make([]byte, size)
	}

	return bp
}

// Get retrieves a buffer from the pool.
func (bp *BufferPool) Get() []byte {
	atomic.AddUint64(&bp.gets, 1)
	buf := bp.pool.Get().([]byte)
	return buf[:bp.size] // Reset length
}

// Put returns a buffer to the pool.
// The buffer should not be used after calling Put.
func (bp *BufferPool) Put(buf []byte) {
	if cap(buf) < bp.size {
		// Buffer is too small, don't return it to pool
		return
	}

	atomic.AddUint64(&bp.puts, 1)
	bp.pool.Put(buf[:bp.size]) // Reset length before returning
}

// Stats returns pool statistics.
func (bp *BufferPool) Stats() (gets, puts, allocated uint64, reuseRate float64) {
	gets = atomic.LoadUint64(&bp.gets)
	puts = atomic.LoadUint64(&bp.puts)
	allocated = atomic.LoadUint64(&bp.allocated)

	if gets > 0 {
		reuseRate = float64(gets-allocated) / float64(gets) * 100
	}

	return
}

// GlobalBufferPools provides pre-configured buffer pools for common sizes.
var GlobalBufferPools = struct {
	Small  *BufferPool // 260 bytes - Modbus TCP ADU
	Medium *BufferPool // 512 bytes - Safe buffer
	Large  *BufferPool // 1024 bytes - Multi-frame buffer
	Huge   *BufferPool // 4096 bytes - Large payload buffer
}{
	Small:  NewBufferPool(260),
	Medium: NewBufferPool(512),
	Large:  NewBufferPool(1024),
	Huge:   NewBufferPool(4096),
}

// GetBuffer retrieves a buffer of appropriate size from the global pools.
// The returned buffer will have at least minSize bytes.
func GetBuffer(minSize int) []byte {
	switch {
	case minSize <= 260:
		return GlobalBufferPools.Small.Get()
	case minSize <= 512:
		return GlobalBufferPools.Medium.Get()
	case minSize <= 1024:
		return GlobalBufferPools.Large.Get()
	default:
		return GlobalBufferPools.Huge.Get()
	}
}

// PutBuffer returns a buffer to the appropriate global pool.
func PutBuffer(buf []byte) {
	size := cap(buf)
	switch {
	case size <= 260:
		GlobalBufferPools.Small.Put(buf)
	case size <= 512:
		GlobalBufferPools.Medium.Put(buf)
	case size <= 1024:
		GlobalBufferPools.Large.Put(buf)
	case size <= 4096:
		GlobalBufferPools.Huge.Put(buf)
	}
	// Larger buffers are not pooled, let GC handle them
}

// ObjectPool is a generic pool for any type.
type ObjectPool[T any] struct {
	pool sync.Pool
	new  func() T
	gets uint64
	puts uint64
}

// NewObjectPool creates a new object pool with the given constructor.
func NewObjectPool[T any](new func() T) *ObjectPool[T] {
	op := &ObjectPool[T]{
		new: new,
	}

	op.pool.New = func() interface{} {
		return new()
	}

	return op
}

// Get retrieves an object from the pool.
func (op *ObjectPool[T]) Get() T {
	atomic.AddUint64(&op.gets, 1)
	return op.pool.Get().(T)
}

// Put returns an object to the pool.
func (op *ObjectPool[T]) Put(obj T) {
	atomic.AddUint64(&op.puts, 1)
	op.pool.Put(obj)
}

// Stats returns pool statistics.
func (op *ObjectPool[T]) Stats() (gets, puts uint64) {
	return atomic.LoadUint64(&op.gets), atomic.LoadUint64(&op.puts)
}
