package ssh

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type KnownHostsStatus int

const handshakeTimeout = 15 * time.Second

const (
	KnownHostsUnknown KnownHostsStatus = iota
	KnownHostsMatch
	KnownHostsMismatch
)

type HostKeyInfo struct {
	Key        ssh.PublicKey
	KnownHosts KnownHostsStatus
}

type HostKeyMismatchError struct {
	Host string
	Want ssh.PublicKey
	Got  ssh.PublicKey
}

func (e *HostKeyMismatchError) Error() string {
	return fmt.Sprintf("host key mismatch for %s: trusted %s, got %s",
		e.Host, Fingerprint(e.Want), Fingerprint(e.Got))
}

var ErrNoHostKey = errors.New("no trusted host key configured for host")
var errProbeDone = errors.New("host keyt probe complete")

func FetchHostKey(ctx context.Context, host string) (*HostKeyInfo, error) {
	info := &HostKeyInfo{}

	config := &ssh.ClientConfig{
		User: "ndeploy-probe",
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			info.Key = key
			info.KnownHosts = checkKnownHosts(hostname, remote, key)
			return errProbeDone
		},
	}

	client, err := handshake(ctx, host, config, handshakeTimeout)
	if client != nil {
		client.Close()
	}

	if info.Key == nil {
		return nil, err
	}

	return info, nil
}

func checkKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) KnownHostsStatus {
	home, err := os.UserHomeDir()
	if err != nil {
		return KnownHostsUnknown
	}

	callback, err := knownhosts.New(filepath.Join(".ssh", "known_hosts"))
	if err != nil {
		return KnownHostsUnknown
	}

	err = callback(hostname, remote, key)
	if err == nil {
		return KnownHostsMatch
	}

	var keyErr *knownhosts.KeyError
	if !errors.As(err, &keyErr) {
		return KnownHostsUnknown
	}

	for _, want := range keyErr.Want {
		if want.Key.Type() == key.Type() {
			return KnownHostsMismatch
		}
	}

	return KnownHostsUnknown
}

func pinnedHostKey(want ssh.PublicKey) ssh.HostKeyCallback {
	wantBytes := want.Marshal()

	return func(hostname string, _ net.Addr, got ssh.PublicKey) error {
		if string(got.Marshal()) == string(wantBytes) {
			return nil
		}

		return &HostKeyMismatchError{Host: hostname, Want: want, Got: got}
	}
}

func hostKeyAlgorithms(key ssh.PublicKey) []string {
	if key.Type() == ssh.KeyAlgoRSA {
		return []string{ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256, ssh.KeyAlgoRSA}
	}

	return []string{key.Type()}
}

func MarshalHostKey(key ssh.PublicKey) string {
	line := ssh.MarshalAuthorizedKey(key)
	return string(line[:len(line)-1])
}

func ParseHostKey(stored string) (ssh.PublicKey, error) {
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(stored))
	return key, err
}

func Fingerprint(key ssh.PublicKey) string {
	if key == nil {
		return "<none>"
	}

	return ssh.FingerprintSHA256(key)
}

func hostAddr(host string) string {
	if _, _, err := net.SplitHostPort(host); err == nil {
		return host
	}

	return net.JoinHostPort(host, "22")
}

func handshake(ctx context.Context, host string, config *ssh.ClientConfig, timeout time.Duration) (*ssh.Client, error) {
	addr := hostAddr(host)

	dialer := net.Dialer{Timeout: handshakeTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	if timeout > 0 {
		conn.SetDeadline(time.Now().Add(timeout))
	}

	stop := context.AfterFunc(ctx, func() { conn.Close() })
	defer stop()

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return nil, err
	}

	conn.SetDeadline(time.Time{})
	return ssh.NewClient(c, chans, reqs), nil
}
