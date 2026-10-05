package engine

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"
	"os"
	"os/exec"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/Joey574/sink/v2/pkg/sink"
)

const (
	cancelGracePeriod = 30 * time.Second
	maxOutputLine     = 1024 * 1024
)

type Engine struct {
	q *db.Queries

	sink    sink.Sink
	signers *ssh.SignerCache

	mx      sync.Mutex
	runs    map[int64]*Run
	nextRun int64
}

func NewEngine(options ...func(*Engine)) *Engine {
	e := &Engine{
		sink:    sink.New(),
		signers: ssh.NewSignerCache(),
		runs:    make(map[int64]*Run),
	}

	for _, o := range options {
		o(e)
	}

	return e
}

func (e *Engine) SetRecorder(q *db.Queries) {
	e.q = q
}

func (e *Engine) ProbeHost(ctx context.Context, host string) (*ssh.HostKeyInfo, error) {
	e.sink.Printf(sink.DEBUG, "probing host key of %s\n", host)

	info, err := ssh.FetchHostKey(ctx, host)
	if err != nil {
		e.sink.Printf(sink.WARN, "probing %s failed: %v\n", host, err)
		return nil, err
	}

	e.sink.Printf(sink.DEBUG, "%s presented %s %s\n", host, info.Key.Type(), ssh.Fingerprint(info.Key))
	return info, nil
}

func (e *Engine) TestConnection(ctx context.Context, n *db.Node, prompt ssh.PassphrasePrompt) error {
	client, err := e.newClient(n, prompt)
	if err != nil {
		return err
	}
	defer client.Close()

	conn, err := client.Dial(ctx)
	if err != nil {
		return err
	}

	return conn.Close()
}

func (e *Engine) StartDeployment(ctx context.Context, d *Deployment) (*Run, error) {
	if err := d.Validate(); err != nil {
		return nil, err
	}

	tmp, err := os.MkdirTemp("", "ndeploy-deploy-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}

	sshConfig, err := writeSSHConfig(tmp, d.nodes())
	if err != nil {
		os.RemoveAll(tmp)
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)
	cmd := d.CommandContext(ctx)
	cmd.Env = append(os.Environ(), "NIX_SSHOPTS="+nixSSHOpts(sshConfig))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
	cmd.WaitDelay = cancelGracePeriod

	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	run := &Run{
		Deployment: d,
		Command:    d.CommandLine(),
		StartedAt:  time.Now(),
		ctx:        ctx,
		cmd:        cmd,
		cancel:     cancel,
		done:       make(chan struct{}),
		exitCode:   -1,
	}

	run.ID = e.record(ctx, run)
	if err != nil {
		cancel()
		pw.Close()
		os.RemoveAll(tmp)
		return nil, fmt.Errorf("recording deployment: %w", err)
	}

	s := sink.New(
		sink.Wrap(e.sink),
		sink.SetName(fmt.Sprintf("deploy-%d", run.ID)),
	)

	s.Printf(sink.INFO, "deploying to %s@%s: %s\n", d.Target.User, d.Target.Host, run.Command)
	s.Printf(sink.DEBUG, "ssh config at %s\n", &sshConfig)

	if err := cmd.Start(); err != nil {
		cancel()
		pw.Close()
		os.RemoveAll(tmp)

		err = fmt.Errorf("starting nixos-rebuild: %w", err)
		s.Printf(sink.ERROR, "%v\n", err)
		e.finishRecord(run, err)
		return nil, err
	}

	e.track(run)

	pumped := make(chan struct{})
	go e.pump(run, s, pr, pumped)
	go e.wait(run, s, pw, pumped, tmp)

	return run, nil
}

func (e *Engine) pump(run *Run, s sink.Sink, r io.Reader, done chan<- struct{}) {
	defer close(done)

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), maxOutputLine)

	for scanner.Scan() {
		line := scanner.Text() + "\n"
		s.WriteString(sink.INFO, line)

		if run.Deployment.Output != nil {
			if _, err := io.WriteString(&run.Deployment.Output, line); err != nil {
				s.Printf(sink.WARN, "writing deployment output: %v\n", err)
				run.Deployment.Output = nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		s.Printf(sink.WARN, "reading nixos-rebuild output: %v\n", err)
	}
}

func (e *Engine) Runs() []*Run {
	e.mx.Lock()
	defer e.mx.Unlock()

	runs := make([]*Run, 0, len(e.runs))
	for _, r := range e.runs {
		runs = append(runs, r)
	}

	sort.Slice(runs, func(i, j int) bool { return runs[i].ID < runs[j].ID })
	return runs
}

func (e *Engine) Run(id int64) (*Run, bool) {
	e.mx.Lock()
	defer e.mx.Unlock()

	r, ok := e.runs[id]
	return r, ok
}

func (e *Engine) wait(run *Run, s sink.Sink, pw *io.PipeWriter, pumped <-chan struct{}, tmp string) {
	waitErr := run.cmd.Wait()

	pw.Close()
	<-pumped
	os.RemoveAll(tmp)

	if run.ctx.Err() != nil {
		run.markCancelled()
	}
	run.cancel()

	var err error
	exitCode := -1

	var exitErr *exec.ExitError
	isExit := errors.As(waitErr, &exitErr)

	switch {
	case run.Cancelled():
		err = context.Canceled
		if isExit {
			exitCode = exitErr.ExitCode()
		}
		s.Println(sink.WARN, "deployment cancelled")

	case isExit:
		exitCode = exitErr.ExitCode()
		err = fmt.Errorf("nixos-rebuild exited with status %d", exitCode)
		s.Printf(sink.ERROR, "%v\n", err)

	case waitErr != nil:
		err = waitErr
		s.Printf(sink.ERROR, "nixos-rebuild failed %v\n", err)

	default:
		exitCode = 0
		s.Printf(sink.INFO, "deployment finished succesfully in %s", time.Since(run.StartedAt).Round(time.Second))
	}

	run.finish(err, exitCode)
	e.finishRecord(run, err)
	e.untrack(run)
	close(run.done)
}

func (e *Engine) track(run *Run) {
	e.mx.Lock()
	defer e.mx.Unlock()
	e.runs[run.ID] = run
}

func (e *Engine) untrack(run *Run) {
	e.mx.Lock()
	defer e.mx.Unlock()
	delete(e.runs, run.ID)
}

func (e *Engine) record(ctx context.Context, run *Run) (int64, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	d := run.Deployment
	params := db.CreateDeploymentParams{
		TargetNode: d.Target.ID,
		Action:     d.Type.String(),
		ConfigFile: d.Config,
		Command: run.Command,
		IsUpgrade: d.Upgrade,
		IsRollback: d.Rollback
	}

	if d.Builder != nil {
		params.BuildNode = sql.NullInt64{Int64: d.Builder.ID, Valid: true}
	}

	rec, err := e.q.CreateDeployment(context.WithoutCancel(ctx), params)
	if err != nil {
		return 0, err
	}

	return rec.ID, nil
}

func (e *Engine) finishRecord(run *Run, runErr error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	ctx := context.Background()
	if run.Cancelled() {
		if err := e.q.SetCancelled(ctx, db.SetCancelledParams{IsCancelled: true, ID: run.ID}); err != nil {
			e.sink.Printf(sink.ERROR, "marking dpeloyment %d cancelled: %v\n", run.ID, err)
		}
	}

	msg := ""
	if runErr != nil {
		msg = runErr.Error()
	}

	err := e.q.FinishDeployment(ctx, db.FinishDeploymentParams{
		ExitCode: int64(run.ExitCode()),
		Error: msg,
		ID: run.ID
	})
	if err != nil {
		e.sink.Printf(sink.ERROR, "finishing deployment %d: %v\n", run.ID, err)
	}
}

func (e *Engine) newClient(n *db.Node, prompt ssh.PassphrasePrompt) (*ssh.Client, error) {
	hostKey, err := ssh.ParseHostKey(n.HostKey)
	if err != nil {
		return nil, fmt.Errorf("stored host key of %s is invalid: %w", n.Host, err)
	}

	return ssh.NewClient(
		n.User,
		n.Host,
		ssh.SetHostKey(hostKey),
		ssh.SetIdentityFile(n.IdentityFile),
		ssh.SetPassphrasePrompt(prompt),
		ssh.SetSignerCache(e.signers),
		ssh.SetSink(sink.New(
			sink.Wrap(e.sink),
			sink.SetName(fmt.Sprintf("ssh-%s", n.Host)),
		)),
	)
}
