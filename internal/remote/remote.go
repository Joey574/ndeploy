package remote

import (
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

type Remote struct {
	ConfigDir  string
	ConfigPath string
	User       string
	Host       string
}

func NewRemote(dir string) *Remote {
	host := filepath.Base(dir)

	user := "admin"
	if host == "desktop" {
		user = "joey"
	}

	return &Remote{
		ConfigDir:  dir,
		ConfigPath: dir + "/configuration.nix",
		User:       user,
		Host:       host,
	}
}

func (r *Remote) CreateClosure() error {
	config := &ssh.ClientConfig{
		User: r.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(nil),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}
