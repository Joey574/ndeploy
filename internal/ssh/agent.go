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

func AgentIdentities() ([]Identity, error) {
	ag, conn, err := connectToAgent()
	if errors.Is(err, ErrNoAgent) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer conn.Close()

	keys, err := ag.List()
	if err != nil {
		return nil, err
	}

	identities := make([]Identity, 0, len(keys))
	for _, k := range keys {
		identities = append(identities, Identity{
			Comment:     k.Comment,
			Fingerprint: Fingerprint(k),
			FromAgent:   true,
		})
	}

	return identities, nil
}
