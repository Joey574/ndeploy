package ssh

import (
	"fmt"
	"net"

	"github.com/Joey574/sink/v2/pkg/sink"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	user      string
	host      string
	agentConn net.Conn

	sink   sink.Sink
	config *ssh.ClientConfig
}

func NewClient(user, host string, options ...func(*Client)) (*Client, error) {
	c := &Client{
		user: user,
		host: host,
		sink: sink.New(sink.EnableStdOut()),
		config: &ssh.ClientConfig{
			User:            user,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO
		},
	}

	for _, o := range options {
		o(c)
	}

	// connect to agent after options are applied so we
	// write to the correct sink
	agent, conn, err := connectToAgent()
	if err != nil {
		c.sink.Printf(sink.WARN, "%v\n", err)
		return nil, err
	}

	c.agentConn = conn
	c.config.Auth = append(c.config.Auth, ssh.PublicKeysCallback(agent.Signers))
	return c, nil
}

func (c *Client) Dial() (*ssh.Client, error) {
	return ssh.Dial("tcp", c.host+":22", c.config)
}

func Dial(user, host string) (*ssh.Session, error) {
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(nil),
		},

		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", host), config)
	if err != nil {
		return nil, err
	}

	return client.NewSession()
}

func (c *Client) Close() error {
	if c.agentConn != nil {
		return c.agentConn.Close()
	}

	return nil
}
