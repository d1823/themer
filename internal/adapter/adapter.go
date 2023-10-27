package adapter

import (
	"bytes"
	"fmt"
	"github.com/d1823/themer/internal/config"
	"github.com/d1823/themer/internal/freedesktop"
	"os"
	"os/exec"
	"time"
)

func ExecuteSymlinkAdapter(preference freedesktop.ColorSchemePreference, config config.SymlinkAdapter) error {
	// NOTE: The target file might not yet exist.
	_ = os.Remove(config.TargetFile)

	preferenceFile := ""
	switch preference {
	case freedesktop.NoPreference:
		preferenceFile = config.NoPreferenceFile
	case freedesktop.PreferDarkAppearance:
		preferenceFile = config.DarkPreferenceFile
	case freedesktop.PreferLightAppearance:
		preferenceFile = config.LightPreferenceFile
	default:
		return fmt.Errorf("invalid preference: %v", preference)
	}

	if err := os.Symlink(preferenceFile, config.TargetFile); err != nil {
		return fmt.Errorf("symlinking %s to the target file %s: %w", preferenceFile, config.TargetFile, err)
	}

	return nil
}

func ExecuteAlacrittyAdapter(preference freedesktop.ColorSchemePreference, config config.AlacrittyAdapter) error {
	err := ExecuteSymlinkAdapter(preference, config.SymlinkAdapter)
	if err != nil {
		return fmt.Errorf("symlinking within alacritty adapter: %w", err)
	}

	t := time.Now().Local()
	err = os.Chtimes(config.AlacrittyConfigFile, t, t)
	if err != nil {
		return fmt.Errorf("touching within alacritty adapter: %w", err)
	}

	return nil
}

func ExecuteTmuxAdapter(preference freedesktop.ColorSchemePreference, config config.TmuxAdapter) error {
	err := ExecuteSymlinkAdapter(preference, config.SymlinkAdapter)
	if err != nil {
		return fmt.Errorf("symlinking within tmux adapter: %w", err)
	}

	cmd := exec.Command("tmux", "source-file", config.TmuxConfigFile)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("sourcing tmux config: %w", err)
	}

	return nil
}

func runCmd(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", err
	}
	cmd := exec.Command(tmux, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	outStr, errStr := string(stdout.Bytes()), string(stderr.Bytes())

	return outStr, errStr, err
}
