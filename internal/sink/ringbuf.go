package sink

import (
	"fmt"
	"sync"
)

type RingBuffer struct {
	capacity int
	size     int
	start    int
	end      int
	version  uint64
	buf      []byte

	mx sync.Mutex
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity == 0 {
		panic("size must be non 0")
	}

	return &RingBuffer{
		capacity: capacity,
		size:     0,
		start:    0,
		end:      0,
		version:  0,
		buf:      make([]byte, capacity),
	}
}

func (r *RingBuffer) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.buf = nil
	r.capacity = 0
	r.size = 0
	r.start = 0
	r.end = 0
	r.version = 0
	return nil
}

func (r *RingBuffer) Version() uint64 {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.version
}

func (r *RingBuffer) Capacity() int {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.capacity
}

func (r *RingBuffer) Write(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.buf == nil {
		return 0, fmt.Errorf("buffer is closed")
	}

	n, err := r.writeLocked(p)
	if err == nil {
		r.version++
	}

	return n, err
}

func (r *RingBuffer) writeLocked(p []byte) (int, error) {
	if len(p) >= r.capacity {
		return r.truncateCopyLocked(p)
	}

	if overflow := r.size + len(p) - r.capacity; overflow > 0 {
		r.start = (r.start + overflow) % r.capacity
	}

	r.size = min(r.size+len(p), r.capacity)
	n1 := copy(r.buf[r.end:], p)
	copy(r.buf, p[n1:])

	r.end = (r.end + len(p)) % r.capacity
	return len(p), nil
}

// Reads data from buffer without advancing the start of it
// if the buffer is bigger than the data available, only the available
// data will be returned
func (r *RingBuffer) Read(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.buf == nil {
		return 0, fmt.Errorf("buffer is closed")
	}

	return r.readBufferLocked(p)
}

func (r *RingBuffer) ReadAll() ([]byte, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.buf == nil {
		return nil, fmt.Errorf("buffer is closed")
	}

	buf := make([]byte, r.size)
	_, err := r.readBufferLocked(buf)
	return buf, err
}

func (r *RingBuffer) IsFull() bool {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.isFullLocked()
}

func (r *RingBuffer) isFullLocked() bool {
	return r.start == r.end && r.size != 0
}

func (r *RingBuffer) IsEmpty() bool {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.isEmptyLocked()
}

func (r *RingBuffer) isEmptyLocked() bool {
	return r.start == r.end && r.size == 0
}

// Reads data from buffer and advances the start
func (r *RingBuffer) Consume(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.buf == nil {
		return 0, fmt.Errorf("buffer is closed")
	}

	n, err := r.readBufferLocked(p)
	if err == nil {
		r.start = (r.start + n) % r.capacity
		r.size -= n
	}
	return n, err
}

// Handles copy for data which is larger than the buffer, requiring truncation
func (r *RingBuffer) truncateCopyLocked(p []byte) (int, error) {
	idx := len(p) - len(r.buf)
	copy(r.buf, p[idx:])
	r.start = 0
	r.end = 0
	r.size = r.capacity
	return len(r.buf), nil
}

func (r *RingBuffer) readBufferLocked(p []byte) (int, error) {
	n := min(len(p), r.size)
	if n == 0 {
		return 0, nil
	}

	idx := min(n, r.capacity-r.start)
	copy(p[:idx], r.buf[r.start:r.start+idx])

	if idx < n {
		copy(p[idx:n], r.buf[:n-idx])
	}

	return n, nil
}
