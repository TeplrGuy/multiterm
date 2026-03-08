package setup

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func IsTmuxInstalled() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func IsBrewInstalled() bool {
	_, err := exec.LookPath("brew")
	return err == nil
}

func GetOS() string {
	return runtime.GOOS
}

func InstallTmux() error {
	cmd := exec.Command("brew", "install", "tmux")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func EnsureTmux() error {
	if IsTmuxInstalled() {
		return nil
	}

	switch GetOS() {
	case "darwin":
		if !IsBrewInstalled() {
			return fmt.Errorf("tmux is required. Install it with: brew install tmux")
		}

		fmt.Println("tmux not found. Installing via Homebrew...")
		if err := InstallTmux(); err != nil {
			return fmt.Errorf("failed to install tmux: %w", err)
		}

		if !IsTmuxInstalled() {
			return fmt.Errorf("tmux installation succeeded but binary not found in PATH")
		}
	case "linux":
		return fmt.Errorf("tmux is required. Install it with: sudo apt install tmux (Debian/Ubuntu) or sudo yum install tmux (RHEL/CentOS)")
	default:
		return fmt.Errorf("tmux is required but automatic installation is not supported on %s", GetOS())
	}

	return nil
}
