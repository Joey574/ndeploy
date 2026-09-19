package ssh

import (
	"bytes"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"golang.org/x/crypto/ssh"
)

type PassphrasePrompt func(path, fingerprint string, attempt int) ([]byte, error)

var (
	ErrPromptCanceled  = errors.New("passphrase prompt cancelled")
	ErrNoPrompt        = errors.New("private key is encrypted and no passphrase prompt is configured")
	ErrTooManyAttempts = errors.New("too many wrong passphrases")
)

const maxPassphraseAttempts = 3
const maxKeyFileSize = 64 * 1024

var defaultIdentityNames = []string{"id_ed25519", "id_ecdsa", "id_rsa"}

type Identity struct {
	Path        string
	Comment     string
	Fingerprint string
	Encrypted   bool
	FromAgent   bool
}

func (i Identity) Label() string {
	var b strings.Builder

	switch {
	case i.FromAgent:
		b.WriteString("agent: ")
		b.WriteString(i.Comment)
	default:
		b.WriteString(abbreviateHome(i.Path))
	}

	if i.Encrypted {
		b.WriteString(" (encrypted")
	}

	if i.Fingerprint != "" {
		b.WriteString(" ")
		b.WriteString(i.Fingerprint)
	}

	return b.String()
}

func abbreviateHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}

	if rel, ok := strings.CutPrefix(path, home+string(filepath.Separator)); ok {
		return filepath.Join("~", rel)
	}

	return path
}

func Dir() string {
	dir, _ := sshDir()
	return dir
}

func sshDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".ssh"), nil
}

func DefaultIdentityFiles() []string {
	dir, err := sshDir()
	if err != nil {
		return nil
	}

	var files []string
	for _, name := range defaultIdentityNames {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}

	return files
}

func DiscoverIdentities() ([]Identity, error) {
	dir, err := sshDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var identities []Identity
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".pub") {
			continue
		}

		identity, err := InspectIdentity(filepath.Join(dir, entry.Name()))
		if err != nil {
			continue
		}

		identities = append(identities, identity)
	}

	sort.Slice(identities, func(i, j int) bool {
		return identities[i].Path < identities[j].Path
	})

	return identities, nil
}

func readKeyFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.Size() > maxKeyFileSize {
		return nil, fmt.Errorf("%s is too large to be a private key", path)
	}

	return os.ReadFile(path)
}

func InspectIdentity(path string) (Identity, error) {
	identity, _, err := inspect(path)
	return identity, err
}

func inspect(path string) (Identity, ssh.PublicKey, error) {
	identity := Identity{Path: path}

	pem, err := readKeyFile(path)
	if err != nil {
		return identity, nil, err
	}

	if !bytes.Contains(pem, []byte("PRIVATE KEY-----")) {
		return identity, nil, fmt.Errorf("%s is not a private key", path)
	}

	pub, encrypted, err := publicKeyOf(path, pem)
	if err != nil {
		return identity, nil, err
	}

	identity.Encrypted = encrypted
	identity.Comment = commentOf(path)
	if pub != nil {
		identity.Fingerprint = Fingerprint(pub)
	}

	return identity, pub, nil
}

func publicKeyOf(path string, pem []byte) (ssh.PublicKey, bool, error) {
	signer, err := ssh.ParsePrivateKey(pem)
	if err == nil {
		return signer.PublicKey(), false, nil
	}

	var missing *ssh.PassphraseMissingError
	if !errors.As(err, &missing) {
		return nil, false, err
	}

	if missing.PublicKey != nil {
		return missing.PublicKey, true, nil
	}

	if line, err := os.ReadFile(path + ".pub"); err == nil {
		if pub, _, _, _, err := ssh.ParseAuthorizedKey(line); err == nil {
			return pub, true, nil
		}
	}

	return nil, true, nil
}

func commentOf(path string) string {
	line, err := os.ReadFile(path + ".pub")
	if err != nil {
		return ""
	}

	_, comment, _, _, err := ssh.ParseAuthorizedKey(line)
	if err != nil {
		return ""
	}

	return comment
}

func LoadSigner(path string, prompt PassphrasePrompt) (ssh.Signer, error) {
	pem, err := readKeyFile(path)
	if err != nil {
		return nil, err
	}

	signer, err := ssh.ParsePrivateKey(pem)

	var missing *ssh.PassphraseMissingError
	if !errors.As(err, &missing) {
		return signer, err
	}

	if prompt == nil {
		return nil, ErrNoPrompt
	}

	fingerprint := ""
	if pub, _, _ := publicKeyOf(path, pem); pub != nil {
		fingerprint = Fingerprint(pub)
	}

	for attempt := 1; attempt <= maxPassphraseAttempts; attempt++ {
		passphrase, err := prompt(path, fingerprint, attempt)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKeyWithPassphrase(pem, passphrase)
		clear(passphrase)

		if errors.Is(err, x509.IncorrectPasswordError) {
			continue
		}

		return signer, err
	}

	return nil, ErrTooManyAttempts
}

type SignerCache struct {
	mx      sync.Mutex
	signers map[string]ssh.Signer
}

func NewSignerCache() *SignerCache {
	return &SignerCache{signers: make(map[string]ssh.Signer)}
}

func (c *SignerCache) Get(path string) (ssh.Signer, bool) {
	c.mx.Lock()
	defer c.mx.Unlock()

	signer, ok := c.signers[path]
	return signer, ok
}

func (c *SignerCache) Put(path string, signer ssh.Signer) {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.signers[path] = signer
}

func (c *SignerCache) Forget(path string) {
	c.mx.Lock()
	defer c.mx.Unlock()

	delete(c.signers, path)
}

type lazySigner struct {
	pub  ssh.PublicKey
	load func() (ssh.Signer, error)

	once   sync.Once
	signer ssh.Signer
	err    error
}

func (l *lazySigner) resolve() (ssh.Signer, error) {
	l.once.Do(func() {
		l.signer, l.err = l.load()
	})

	return l.signer, l.err
}

func (l *lazySigner) PublicKey() ssh.PublicKey {
	return l.pub
}

func (l *lazySigner) Sign(rand io.Reader, data []byte) (*ssh.Signature, error) {
	signer, err := l.resolve()
	if err != nil {
		return nil, err
	}

	return signer.Sign(rand, data)
}

func (l *lazySigner) SignWithAlgorithm(rand io.Reader, data []byte, algorithm string) (*ssh.Signature, error) {
	signer, err := l.resolve()
	if err != nil {
		return nil, err
	}

	if as, ok := signer.(ssh.AlgorithmSigner); ok {
		return as.SignWithAlgorithm(rand, data, algorithm)
	}

	return signer.Sign(rand, data)
}
