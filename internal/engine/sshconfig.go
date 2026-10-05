package engine

import (
	"fmt"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"
	"os"
	"path/filepath"
	"strings"
)

const (
	sshConfigName     = "config"
	sshKnownHostsName = "known_hosts"
)

func writeSSHConfig(dir string, nodes []*db.Node) (string, error) {
	knownhosts := filepath.Join(dir, sshKnownHostsName)
	configPath := filepath.Join(dir, sshConfigName)

	var hosts, config strings.Builder

	for _, n := range nodes {
		key, err := ssh.ParseHostKey(n.HostKey)
		if err != nil {
			return "", fmt.Errorf("stored host key of %s is invalid: %w", n.Host, err)
		}

		host, port := splitHostPort(n.Host)

		fmt.Fprintf(&hosts, "%s %s\n", knownHostsName(host, port), ssh.MarshalHostKey(key))
		fmt.Fprintf(&config, "	Host %s\n", host)
		fmt.Fprintf(&config, "	User %s\n", n.User)
		if port != "" {
			fmt.Fprintf(&config, "	Port %s\n", port)
		}

		if n.IdentityFile != "" {
			fmt.Fprintf(&config, "	IdentityFile %s\n", n.IdentityFile)
			config.WriteString("	IdentitiesOnly yes\n")
		}
		config.WriteString("\n")
	}

	config.WriteString("Host *\n")
	fmt.Fprintf(&config, "	UserKnownHostsFile %s\n", knownhosts)
	config.WriteString("	StrictHostKeyChecking yes\n")
	config.WriteString("	PasswordAuthentication no\n")
	config.WriteString("	KdbInteractiveAuthentication no\n")
	config.WriteString("	ServerAliveInterval 30\n")

	if err := os.WriteFile(knownhosts, []byte(hosts.String()), 0o600); err != nil {
		return "", err
	}

	if err := os.WriteFile(configPath, []byte(config.String()), 0o600); err != nil {
		return "", err
	}

	return configPath, nil
}

func knownHostsName(host, port string) string {
	if port == "" || port == "22" {
		return host
	}

	return fmt.Sprintf("[%s]:%s", host, port)
}

func nixSSHOpts(config string) string {
	opts := "-F" + config
	if existing := os.Getenv("NIX_SSHOPTS"); existing != "" {
		opts += " " + existing
	}

	return opts
}
