package engine

import (
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"

	"github.com/Joey574/sink/v2/pkg/sink"
)

type Engine struct {
	sink sink.Sink
}

func NewEngine() *Engine {
	return &Engine{}
}

func (e *Engine) TestConnection(n *db.Node) error {
	client, err := ssh.NewClient(n.User, n.Host, ssh.SetSink(e.sink))
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
