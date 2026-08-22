package sink

import (
	"sync"
)

type RingBuffer struct {
	size  uint64
	start uint64
	end   uint64
	buf   []byte

	mx sync.RWMutex
}

func NewRingBuffer(size uint64) *RingBuffer {
	return &RingBuffer{
		size:  size,
		start: 0,
		end:   0,
		buf:   make([]byte, size),
	}
}

func (r *RingBuffer) Close() error {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.buf = nil
	return nil
}

func (r *RingBuffer) Write(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if len(p) >= len(r.buf) {
		return r.truncateCopy(p)
	} else if int(r.end)+len(p) <= len(r.buf) {
		return r.copy(p)
	} else {
		return r.overflowCopy(p)
	}
}

func (r *RingBuffer) Read(p []byte) (int, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	return r.readBuffer(p)
}

func (r *RingBuffer) ReadAll() ([]byte, error) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	n := max(int(r.start)-int(r.end), int(r.end)-int(r.start))
	buf := make([]byte, n)
	_, err := r.readBuffer(buf)
	return buf, err
}

func (r *RingBuffer) Consume(p []byte) (int, error) {
	r.mx.Lock()
	defer r.mx.Unlock()

	n, err := r.readBuffer(p)
	if err == nil {
		r.start = (r.start + uint64(n)) % r.size
	}
	return n, err
}

func (r *RingBuffer) Dump() {

}

// Lock must already be set before this is called
// Handles copy for data which is larger than the buffer, requiring truncation
func (r *RingBuffer) truncateCopy(p []byte) (int, error) {
	idx := len(p) - len(r.buf)
	copy(r.buf, p[idx:])
	r.start = 0
	r.end = r.size
	return len(r.buf), nil
}

// Lock must already be set before this is called
// Handles copies for when data fits into the buffer, without
func (r *RingBuffer) copy(p []byte) (int, error) {
	copy(r.buf[r.end:], p)
	r.end += uint64(len(p))
	return len(p), nil
}

// Lock must already be set before this is called
// Handles copy for data when it would require an overflow
func (r *RingBuffer) overflowCopy(p []byte) (int, error) {
	eidx := len(r.buf) - int(r.end)
	copy(r.buf[r.end:], p[:eidx])
	copy(r.buf[0:], p[eidx:])

	r.end = uint64(len(p) - eidx)
	return len(p), nil
}

// Lock must already be set before this is called
// TODO: special cases can definitely make this faster
func (r *RingBuffer) readBuffer(p []byte) (int, error) {
	idx := r.start
	count := 0

	for count < len(p) && idx != r.end {
		p[count] = r.buf[idx]

		count++
		idx = (idx + 1) % r.size
	}

	return count, nil

	// if p is greater than buffer size
	// if p will overflow
	// best case
}
