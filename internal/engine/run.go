package engine

import (
	"context"
	"os/exec"
	"sync"
	"time"
)

type Run struct {
	ID         int64
	Deployment *Deployment
	Command    string
	StartedAt  time.Time

	ctx    context.Context
	cmd    *exec.Cmd
	cancel context.CancelFunc
	done   chan struct{}

	mx        sync.Mutex
	err       error
	exitCode  int
	cancelled bool
}

func (r *Run) Done() <-chan struct{} {
	return r.done
}

func (r *Run) Wait() error {
	<-r.done
	return r.Err()
}

func (r *Run) Err() error {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.err
}

func (r *Run) ExitCode() int {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.exitCode
}

func (r *Run) Cancelled() bool {
	r.mx.Lock()
	defer r.mx.Unlock()
	return r.cancelled
}

func (r *Run) markCancelled() {
	r.mx.Lock()
	defer r.mx.Unlock()
	r.cancelled = true
}

func (r *Run) Running() bool {
	select {
	case <-r.done:
		return false
	default:
		return true
	}
}

func (r *Run) Cancel() {
	r.markCancelled()
	r.cancel()
}

func (r *Run) finish(err error, exitCode int) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.err = err
	r.exitCode = exitCode
}
