package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

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
