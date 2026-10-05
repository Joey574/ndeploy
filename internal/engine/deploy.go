package engine

import (
	"context"
	"errors"
	"fmt"
	"io"
	"ndeploy/v2/internal/db"
	"os/exec"
	"strings"
)

type DeploymentType int

const (
	Switch DeploymentType = iota
	Boot
	Test
	Build
	DryActivate
	DryBuild
	BuildVM
	BuildVMWithBootloader
)

var (
	ErrNoTarget = errors.New("deployment has no target node")
	ErrNoUser   = errors.New("build node has no user")
	ErrNoHost   = errors.New("build node has no host")
	ErrNoConfig = errors.New("deployment has no configuration file")
)

func DeploymentTypes() []DeploymentType {
	return []DeploymentType{Switch, Boot, Test, Build, DryActivate, DryBuild, BuildVM, BuildVMWithBootloader}
}

func ParseDeploymentType(s string) (DeploymentType, bool) {
	for _, t := range DeploymentTypes() {
		if t.String() == s {
			return t, true
		}
	}

	return 0, false
}

// Reports whether the type changes the running system of the target
func (d DeploymentType) Activate() bool {
	switch d {
	case Switch, Boot, Test, DryActivate:
		return true
	default:
		return false
	}
}

type Deployment struct {
	Type   DeploymentType
	Config string

	Upgrade  bool
	Rollback bool

	Sudo bool

	Target  *db.Node
	Builder *db.Node

	Output io.Writer
}

// Validates the deployment parameters
func (d *Deployment) Validate() error {
	if d.Target == nil {
		return ErrNoTarget
	}

	if d.Target.User == "" {
		return ErrNoUser
	}

	if d.Target.Host == "" {
		return ErrNoHost
	}

	if d.Rollback && !d.Type.Activate() {
		return fmt.Errorf("rollback is only valid with switch, boot, test, or dry-activate, not %s", d.Type)
	}

	if !d.Rollback && d.Config == "" {
		return ErrNoConfig
	}

	return nil
}

func (d *Deployment) Args() []string {
	args := []string{d.Type.String()}

	if d.Rollback {
		args = append(args, "--rollback")
	} else {
		args = append(args, "-I", "nixos-config="+d.Config)
	}

	if d.Upgrade {
		args = append(args, "--upgrade", "--upgrade-all")
	}

	if d.Sudo {
		args = append(args, "--sudo")
	}

	args = append(args, "--target-host", userAtHost(d.Target))
	if d.Builder != nil {
		args = append(args, "--build-host", userAtHost(d.Builder))
	}

	return args
}

func (d *Deployment) CommandLine() string {
	return "nixos-rebuild " + strings.Join(d.Args(), " ")
}

func (d *Deployment) CommandContext(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, "nixos-rebuild", d.Args()...)
}

func (d *Deployment) nodes() []*db.Node {
	nodes := []*db.Node{d.Target}
	if d.Builder != nil && d.Builder.ID != d.Target.ID {
		nodes = append(nodes, d.Builder)
	}

	return nodes
}

func (d DeploymentType) String() string {
	switch d {
	case Switch:
		return "switch"
	case Boot:
		return "boot"
	case Test:
		return "test"
	case Build:
		return "build"
	case DryActivate:
		return "dry-activate"
	case DryBuild:
		return "dry-build"
	case BuildVM:
		return "build-vm"
	case BuildVMWithBootloader:
		return "build-vm-with-bootloader"
	default:
		return "unknown"
	}
}
