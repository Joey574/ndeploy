package ssh

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

func readKeyFile(path string) ([]byte, error) {}

func InspectIdentity(path string) (Identity, error) {}

func inspect(path string) (Identity, ssh.PublicKey, error) {}

func publicKeyOf(path string, pem []byte) (pub ssh.PublicKey, encrypted bool, error) {}

func commentOf(path string) string {}

func LoadSigner(path string, prompt PassphrasePrompt) (ssh.Signer, error) {}
