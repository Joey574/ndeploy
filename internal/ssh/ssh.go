package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	user   string
	host   string
	config *ssh.ClientConfig
}

func NewClient(user, host string) (*Client, error) {
	agent, err := connectToAgent()
	if err != nil {
		//sink.Printf(sink.WARN, "%v\n", err)
		return nil, err
	}

	c := &Client{
		user: user,
		host: host,
		config: &ssh.ClientConfig{
			User: user,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeysCallback(agent.Signers),
			},

			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}

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
