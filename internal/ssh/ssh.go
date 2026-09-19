package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/Joey574/sink/v2/pkg/sink"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

var ErrNoIdentities = errors.New("no ssh identities available, load a key into ssh-agent or pick a key file")

type Client struct {
	user string
	host string
	sink sink.Sink

	hostKey      ssh.PublicKey
	identityFile string

	prompt PassphrasePrompt
	cache  *SignerCache

	agent     agent.ExtendedAgent
	agentConn net.Conn
}

func NewClient(user, host string, options ...func(*Client)) (*Client, error) {
	if user == "" || host == "" {
		return nil, errors.New("ssh client needs a user and a host")
	}

	c := &Client{
		user: user,
		host: host,
		sink: sink.New(sink.EnableStdOut()),
	}

	for _, o := range options {
		o(c)
	}

	agent, conn, err := connectToAgent()
	if err != nil {
		c.sink.Printf(sink.DEBUG, "ssh-agent unavailable, using key files only: %v\n", err)
	} else {
		c.agent = agent
		c.agentConn = conn
	}

	return c, nil
}

func (c *Client) Dial(ctx context.Context) (*ssh.Client, error) {
	if c.hostKey == nil {
		return nil, ErrNoHostKey
	}

	config := &ssh.ClientConfig{
		User:              c.user,
		HostKeyCallback:   pinnedHostKey(c.hostKey),
		HostKeyAlgorithms: hostKeyAlgorithms(c.hostKey),

		Auth: []ssh.AuthMethod{ssh.PublicKeysCallback(c.signers)},
	}

	c.sink.Printf(sink.DEBUG, "connecting to %s@%s\n", c.user, hostAddr(c.host))

	client, err := handshake(ctx, c.host, config, 0)
	if err != nil {
		c.sink.Printf(sink.WARN, "connection to %s failed: %v\n", c.host, err)
		return nil, err
	}

	return client, nil
}

func (c *Client) signers() ([]ssh.Signer, error) {
	agentSigners := c.agentSigners()

	var signers []ssh.Signer
	if c.identityFile == "" {
		signers = append(signers, agentSigners...)

		for _, path := range DefaultIdentityFiles() {
			signer, err := c.fileSigner(path, agentSigners)
			if err != nil {
				c.sink.Printf(sink.DEBUG, "skipping %s: %v\n", path, err)
				continue
			}

			signers = append(signers, signer)
		}
	} else {
		signer, err := c.fileSigner(c.identityFile, agentSigners)
		if err != nil {
			return nil, fmt.Errorf("identity file %s: %w", c.identityFile, err)
		}

		signers = append(signers, signer)
	}

	signers = dedupeSigners(signers)
	if len(signers) == 0 {
		return nil, ErrNoIdentities
	}

	c.sink.Printf(sink.TRACE, "offering %d ssh identities\n", len(signers))
	return signers, nil
}

func (c *Client) agentSigners() []ssh.Signer {
	if c.agent == nil {
		return nil
	}

	signers, err := c.agent.Signers()
	if err != nil {
		c.sink.Printf(sink.WARN, "listing ssh-agent keys failed: %v\n", err)
		return nil
	}

	return signers
}

func (c *Client) fileSigner(path string, agentSigners []ssh.Signer) (ssh.Signer, error) {
	if c.cache != nil {
		if signer, ok := c.cache.Get(path); ok {
			return signer, nil
		}
	}

	identity, pub, err := inspect(path)
	if err != nil {
		return nil, err
	}

	if pub != nil {
		want := string(pub.Marshal())
		for _, signer := range agentSigners {
			if string(signer.PublicKey().Marshal()) == want {
				return signer, nil
			}
		}
	}

	load := func() (ssh.Signer, error) {
		signer, err := LoadSigner(path, c.prompt)
		if err != nil {
			return nil, err
		}

		if c.cache != nil {
			c.cache.Put(path, signer)
		}

		return signer, nil
	}

	if !identity.Encrypted {
		return load()
	}

	if pub == nil {
		return load()
	}

	return &lazySigner{pub: pub, load: load}, nil
}

func (c *Client) Close() error {
	if c.agentConn != nil {
		return c.agentConn.Close()
	}

	return nil
}

func dedupeSigners(signers []ssh.Signer) []ssh.Signer {
	seen := make(map[string]bool, len(signers))

	out := signers[:0]
	for _, signer := range signers {
		key := string(signer.PublicKey().Marshal())
		if seen[key] {
			continue
		}

		seen[key] = true
		out = append(out, signer)
	}

	return out
}
