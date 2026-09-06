package ssh

import (
	"fmt"
	"net"
	"os"

	"golang.org/x/crypto/ssh/agent"
)

func connectToAgent() (agent.ExtendedAgent, error) {
	socketPath := os.Getenv("SSH_AUTH_SOCK")
	if socketPath == "" {
		//sink.Println(sink.WARN, "SSH_AUTH_SOCK env variable not set")
		return nil, fmt.Errorf("SSH_AUTH_SOCK env variable not set")
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		//sink.Printf(sink.WARN, "failed to connect to ssh agent: %v\n", err)
		return nil, err
	}
	defer conn.Close()

	agentClient := agent.NewClient(conn)
	return agentClient, nil
}
