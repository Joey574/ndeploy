package engine

import (
	"context"
	"fmt"
	"ndeploy/v2/internal/db"
	"os/exec"
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

type Deployment struct {
	Type   DeploymentType
	Config string

	Upgrade  bool
	Rollback bool

	Target  *db.Node
	Builder *db.Node
}

func (d *Deployment) CommandContext(ctx context.Context) *exec.Cmd {
	args := []string{
		d.Type.String(),
		"--target-host",
		fmt.Sprintf("%s@%s", d.Target.User, d.Target.Host),
		"--build-host",
		fmt.Sprintf("%s@%s", d.Builder.User, d.Builder.Host),
		"-I", fmt.Sprintf("nixos-config=%s", d.Config),
	}

	if d.Upgrade {
		args = append(args, []string{"--upgrade", "--upgrade-all"}...)
	}

	if d.Rollback {
		args = append(args, "rollback")
	}

	return exec.CommandContext(ctx, "nixos-rebuild", args...)
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
