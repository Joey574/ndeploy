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
	buf      []byte

	mx sync.RWMutex
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
	return nil
}

func (r *RingBuffer) Write(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.buf == nil {
		return 0, fmt.Errorf("buffer is closed")
	}

	if len(p) >= r.capacity {
		return r.truncateCopy(p)
	}

	if overflow := r.size + len(p) - r.capacity; overflow > 0 {
		r.start = (r.start + overflow) % r.capacity
	}

	r.size = min(r.size+len(p), r.capacity)
	if r.end+len(p) <= r.capacity {
		return r.copy(p)
	}

	return r.overflowCopy(p)
}

// Reads data from buffer without advancing the start of it
func (r *RingBuffer) Read(p []byte) (int, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()
	if r.buf == nil {
		return 0, fmt.Errorf("buffer is closed")
	}

	return r.readBuffer(p)
}

func (r *RingBuffer) ReadAll() ([]byte, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	if r.buf == nil {
		return nil, fmt.Errorf("buffer is closed")
	}

	buf := make([]byte, r.size)
	_, err := r.readBuffer(buf)
	return buf, err
}

func (r *RingBuffer) IsFull() bool {
	r.mx.RLock()
	defer r.mx.RUnlock()
	return r.isFullLocked()
}

func (r *RingBuffer) isFullLocked() bool {
	return r.start == r.end && r.size != 0
}

func (r *RingBuffer) IsEmpty() bool {
	r.mx.RLock()
	defer r.mx.RUnlock()
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

	n, err := r.readBuffer(p)
	if err == nil {
		r.start = (r.start + n) % r.capacity
		r.size -= n
	}
	return n, err
}

// Lock must already be set before this is called
// Handles copy for data which is larger than the buffer, requiring truncation
func (r *RingBuffer) truncateCopy(p []byte) (int, error) {
	idx := len(p) - len(r.buf)
	copy(r.buf, p[idx:])
	r.start = 0
	r.end = 0
	r.size = r.capacity
	return len(r.buf), nil
}

// Lock must already be set before this is called
// Handles copies for when data fits into the buffer, without an overflow
func (r *RingBuffer) copy(p []byte) (int, error) {
	copy(r.buf[r.end:], p)
	r.end = (r.end + len(p)) % r.capacity
	return len(p), nil
}

// Lock must already be set before this is called
// Handles copy for data when it would require an overflow
func (r *RingBuffer) overflowCopy(p []byte) (int, error) {
	eidx := len(r.buf) - int(r.end)
	copy(r.buf[r.end:], p[:eidx])
	copy(r.buf[0:], p[eidx:])
	r.end = len(p) - eidx
	return len(p), nil
}

// Lock must already be set before this is called
func (r *RingBuffer) readBuffer(p []byte) (int, error) {
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
