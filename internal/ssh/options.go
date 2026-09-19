package ssh

import (
	"github.com/Joey574/sink/v2/pkg/sink"
	"golang.org/x/crypto/ssh"
)

func SetSink(s sink.Sink) func(*Client) {
	return func(c *Client) {
		c.sink = s
	}
}

func SetHostKey(key ssh.PublicKey) func(*Client) {
	return func(c *Client) {
		c.hostKey = key
	}
}

func SetIdentityFile(path string) func(*Client) {
	return func(c *Client) {
		c.identityFile = path
	}
}

func SetPassphrasePrompt(prompt PassphrasePrompt) func(*Client) {
	return func(c *Client) {
		c.prompt = prompt
	}
}

func SetSignerCache(cache *SignerCache) func(*Client) {
	return func(c *Client) {
		c.cache = cache
	}
}
