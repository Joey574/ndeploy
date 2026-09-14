package dbu

import (
	"context"
	"ndeploy/v2/internal/db"
	"sync"

	"github.com/Joey574/sink/v2/pkg/sink"
)

type DbStore struct {
	name string
	dbu  *Dbu

	mx sync.Mutex
}

func NewDBStore(name string, dbu *Dbu) *DbStore {
	return &DbStore{
		name: name,
		dbu:  dbu,
	}
}

func (s *DbStore) Write(level sink.LogLevel, p []byte) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.writeWithCtxLocked(context.Background(), level, p)
}

func (s *DbStore) WriteWithCtx(ctx context.Context, level sink.LogLevel, p []byte) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	return s.writeWithCtxLocked(ctx, level, p)
}

func (s *DbStore) writeWithCtxLocked(ctx context.Context, level sink.LogLevel, p []byte) error {
	q := s.dbu.Queries()
	_, err := q.CreateLog(
		ctx,
		db.CreateLogParams{
			Source:   s.name,
			Level:    level.String(),
			Message:  string(p),
			Metadata: nil,
		})

	return err
}
