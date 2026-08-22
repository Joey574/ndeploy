package sink

import (
	"crypto/rand"
	"sync"
	"sync/atomic"
	"testing"
)

func TestRingBuffer_Sequences(t *testing.T) {
	cases := []struct {
		name   string
		cap    int
		writes []string
		want   string
	}{
		{"simple append", 10, []string{"abc", "def"}, "abcdef"},
		{"exact fill", 6, []string{"abcd", "ef"}, "abcdef"},
		{"evict no wrap", 6, []string{"abcdef", "gh"}, "cdefgh"},
		{"evict with wrap", 6, []string{"abcdef", "gh", "ij", "klm"}, "hijklm"},
		{"oversized truncate", 4, []string{"abcdefgh"}, "efgh"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := NewRingBuffer(c.cap)
			for _, w := range c.writes {
				r.Write([]byte(w))
			}

			buf, _ := r.ReadAll()
			if string(buf) != c.want {
				t.Errorf("got %q, want %q", buf, c.want)
			}
		})
	}
}

func BenchmarkRingBuffer_SingleThreaded(b *testing.B) {
	const bufSize = 8 * 1024 * 1024 // 8MB
	const chunkSize = 4 * 1024      // 4KB writes

	r := NewRingBuffer(bufSize)

	writeBuf := make([]byte, chunkSize)
	if _, err := rand.Read(writeBuf); err != nil {
		b.Fatal(err)
	}
	readBuf := make([]byte, chunkSize)

	b.SetBytes(chunkSize)
	b.ResetTimer()

	for b.Loop() {
		if _, err := r.Write(writeBuf); err != nil {
			b.Fatalf("write: %v", err)
		}
		if _, err := r.Consume(readBuf); err != nil {
			b.Fatalf("consume: %v", err)
		}
	}
}

func BenchmarkRingBuffer_MultiThreaded(b *testing.B) {
	const bufSize = 32 * 1024 * 1024 // 32MB
	const chunkSize = 4 * 1024       // 4KB

	const numWriters = 4
	const numReaders = 4
	const numConsumers = 4

	r := NewRingBuffer(bufSize)

	payload := make([]byte, chunkSize)
	if _, err := rand.Read(payload); err != nil {
		b.Fatal(err)
	}

	var opsRemaining int64 = int64(b.N)

	b.SetBytes(chunkSize)
	b.ResetTimer()

	var wg sync.WaitGroup
	var stop atomic.Bool

	wg.Add(numWriters)
	wg.Add(numReaders)
	wg.Add(numConsumers)

	for range numWriters {
		go func() {
			defer wg.Done()
			buf := make([]byte, chunkSize)
			copy(buf, payload)
			for atomic.AddInt64(&opsRemaining, -1) >= 0 {
				if _, err := r.Write(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	for range numReaders {
		go func() {
			defer wg.Done()
			buf := make([]byte, chunkSize)
			for !stop.Load() {
				if _, err := r.Read(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	for range numConsumers {
		go func() {
			defer wg.Done()
			buf := make([]byte, chunkSize)
			for !stop.Load() {
				if _, err := r.Consume(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	var writerWG sync.WaitGroup
	writerWG.Add(1)
	go func() {
		defer writerWG.Done()
		for atomic.LoadInt64(&opsRemaining) >= 0 {
		}
	}()
	writerWG.Wait()
	stop.Store(true)
	wg.Wait()
}

func BenchmarkRingBuffer_MultiThreaded_LogSimLoad(b *testing.B) {
	const bufSize = 32 * 1024 * 1024 // 32MB
	const chunkSize = 4 * 1024       // 4KB

	const numWriters = 8
	const numReaders = 1

	r := NewRingBuffer(bufSize)

	payload := make([]byte, chunkSize)
	if _, err := rand.Read(payload); err != nil {
		b.Fatal(err)
	}

	var opsRemaining int64 = int64(b.N)

	b.SetBytes(chunkSize)
	b.ResetTimer()

	var wg sync.WaitGroup
	var stop atomic.Bool

	wg.Add(numWriters)
	wg.Add(numReaders)

	for range numWriters {
		go func() {
			defer wg.Done()
			buf := make([]byte, chunkSize)
			copy(buf, payload)
			for atomic.AddInt64(&opsRemaining, -1) >= 0 {
				if _, err := r.Write(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	for range numReaders {
		go func() {
			defer wg.Done()
			buf := make([]byte, chunkSize)
			for !stop.Load() {
				if _, err := r.Read(buf); err != nil {
					b.Error(err)
					return
				}
			}
		}()
	}

	var writerWG sync.WaitGroup
	writerWG.Add(1)
	go func() {
		defer writerWG.Done()
		for atomic.LoadInt64(&opsRemaining) >= 0 {
		}
	}()
	writerWG.Wait()
	stop.Store(true)
	wg.Wait()
}
