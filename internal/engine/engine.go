package engine

import (
	"context"
	"fmt"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"

	"github.com/Joey574/sink/v2/pkg/sink"
)

type Engine struct {
	sink    sink.Sink
	signers *ssh.SignerCache
}

func NewEngine(options ...func(*Engine)) *Engine {
	e := &Engine{
		sink:    sink.New(),
		signers: ssh.NewSignerCache(),
	}

	for _, o := range options {
		o(e)
	}

	return e
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
