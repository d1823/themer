package adapter

import (
	"encoding/xml"
	"fmt"
	"github.com/d1823/themer/internal/config"
	"github.com/d1823/themer/internal/freedesktop"
	"github.com/godbus/dbus/v5"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ExecuteGSettingsGTKThemeAdapter(preference freedesktop.ColorSchemePreference, config config.GSettingsGTKThemeAdapter) error {
	theme := ""
	switch preference {
	case freedesktop.PreferDarkAppearance:
		theme = config.DarkThemeName
	case freedesktop.PreferLightAppearance:
		theme = config.LightThemeName
	default:
		return fmt.Errorf("invalid preference: %v", preference)
	}

	cmd := exec.Command("gsettings", "set", "org.gnome.desktop.interface", "gtk-theme", theme)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set GTK theme: %w", err)
	}
	return nil
}

func ExecuteSymlinkAdapter(preference freedesktop.ColorSchemePreference, config config.SymlinkAdapter) error {
	// NOTE: The target file might not yet exist.
	_ = os.Remove(config.TargetFile)

	preferenceFile := ""
	switch preference {
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

type Node struct {
	XMLName xml.Name `xml:"node"`
	Name    string   `xml:"name,attr"`
	Nodes   []Node   `xml:"node"`
}

func ExecuteKonsoleAdapter(preference freedesktop.ColorSchemePreference, config config.KonsoleAdapter) error {
	// NOTE: The connection returned by the dbus.SessionBus is shared.
	//       Closing it here would mean the main package would no longer receive the signals it's waiting for.
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		log.Fatalf("Failed to connect to the D-Bus session bus: %v", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Printf("Failed to close the connection: %v", closeErr)
		}
	}()

	if err = conn.Auth(nil); err != nil {
		log.Fatalf("Failed to authentication with the D-Bus session bus: %v", err)
	}

	if err = conn.Hello(); err != nil {
		log.Fatalf("Failed to greet with the D-Bus session bus: %v", err)
	}

	var listedServices []string
	err = conn.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&listedServices)
	if err != nil {
		log.Fatalf("Failed to list service names: %v", err)
	}

	var profileName string
	switch preference {
	case freedesktop.PreferDarkAppearance:
		profileName = config.DarkProfileName
	case freedesktop.PreferLightAppearance:
		profileName = config.LightProfileName
	default:
		return fmt.Errorf("invalid preference: %v", preference)
	}

	var introspection string
	for _, service := range listedServices {
		if !strings.HasPrefix(service, "org.kde.konsole-") {
			continue
		}

		err = conn.Object(service, "/Sessions").
			Call("org.freedesktop.DBus.Introspectable.Introspect", 0).
			Store(&introspection)
		if err != nil {
			log.Fatalf("Failed to introspect the service: %v", err)
		}

		var root Node
		err = xml.Unmarshal([]byte(introspection), &root)
		if err != nil {
			log.Fatalf("Failed to parse the introspection: %v", err)
		}

		sessions := make(map[string]struct{})
		for _, node := range root.Nodes {
			sessions[fmt.Sprintf("/Sessions/%s", node.Name)] = struct{}{}
		}

		for session, _ := range sessions {
			call := conn.Object(service, dbus.ObjectPath(session)).
				Call("org.kde.konsole.Session.setProfile", 0, profileName)
			if call.Err != nil {
				log.Fatalf("Failed to set profile: %v", call.Err)
			}
		}
	}

	return nil
}
