package engine

import (
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"

	"github.com/Joey574/sink/v2/pkg/sink"
)

type Engine struct {
	sink sink.Sink
}

func NewEngine(options ...func(*Engine)) *Engine {
	e := &Engine{}

	for _, o := range options {
		o(e)
	}

	return e
}

func (e *Engine) TestConnection(n *db.Node) error {
	client, err := ssh.NewClient(
		n.User,
		n.Host,
		ssh.SetSink(
			sink.New(
				sink.Wrap(e.sink),
			),
		),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	c, err := client.Dial()
	if err != nil {
		return err
	}

	c.Close()
	return nil
}
