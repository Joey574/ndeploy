package ssh

import (
	"errors"
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh/agent"
)

var ErrNoAgent = errors.New("SSH_AUTH_SOCK env variable not set")

func connectToAgent() (agent.ExtendedAgent, net.Conn, error) {
	socketPath := os.Getenv("SSH_AUTH_SOCK")
	if socketPath == "" {
		return nil, nil, fmt.Errorf("SSH_AUTH_SOCK env variable not set")
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		return nil, nil, err
	}

	return agent.NewClient(conn), conn, nil
}
