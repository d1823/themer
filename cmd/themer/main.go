package main

import (
	"fmt"
	"github.com/adrg/xdg"
	"github.com/d1823/themer/internal/adapter"
	"github.com/d1823/themer/internal/config"
	"github.com/d1823/themer/internal/freedesktop"
	"github.com/godbus/dbus/v5"
	"log"
	"os"
	"strings"
)

func main() {
	if len(os.Args) == 2 && strings.Contains(os.Args[1], "-h") {
		fmt.Printf("Usage: %s\n", os.Args[0])
		fmt.Println("Themer is a versatile command-line utility that monitors the org.freedesktop.appearance.color-scheme over DBus and seamlessly triggers color scheme switching in various applications, based on the user's configuration.")
		fmt.Println()
		fmt.Printf(
			"To use Themer, create a JSON configuration file at \"%s/themer/config.json\" that specifies the adapters you want to use and along with any necessary parameters.\n",
			xdg.ConfigHome,
		)

		os.Exit(0)
	}

	p, err := xdg.SearchConfigFile("themer/config.json")
	if err != nil {
		log.Fatalf("reading the config file: %v", err)
	}
	d, err := os.ReadFile(p)
	if err != nil {
		log.Fatalf("reading the config file at %s: %v", p, err)
	}

	c, err := config.Parse(d)
	if err != nil {
		log.Fatalf("parsing the config file from %s: %v", p, err)
	}

	conn, err := dbus.SessionBus()
	if err != nil {
		log.Fatalf("Failed to connect to the D-Bus session bus: %v", err)
	}
	defer conn.Close()

	matchRule := "type='signal',path='/org/freedesktop/portal/desktop',interface='org.freedesktop.impl.portal.Settings',member='SettingChanged'"

	call := conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, matchRule)
	if call.Err != nil {
		log.Fatalf("Failed to add D-Bus signal filter: %v", call.Err)
	}

	signals := make(chan *dbus.Signal, 10)

	conn.Signal(signals)

	for {
		select {
		case signal := <-signals:
			if signal == nil {
				continue
			}

			var settingsChanged freedesktop.SettingChanged
			settingsChanged, err = freedesktop.ParseSettingChangedSignal(signal)
			if err != nil {
				log.Printf("Parsing the SettingsChanged signal: %v", err)
				continue
			}

			if settingsChanged.Key != "color-scheme" {
				continue
			}

			v, ok := settingsChanged.Value.Value().(uint32)
			if !ok {
				continue
			}

			var colorSchemePreference freedesktop.ColorSchemePreference
			colorSchemePreference, err = freedesktop.ParseColorSchemePreference(int(v))
			if err != nil {
				log.Printf("Parsing the ColorSchemePreference signal: %v", err)
				continue
			}

			if colorSchemePreference == freedesktop.NoPreference {
				switch c.NoPreferenceFallback {
				case config.NoPreferenceFallbackDark:
					colorSchemePreference = freedesktop.PreferDarkAppearance
				case config.NoPreferenceFallbackLight:
					colorSchemePreference = freedesktop.PreferLightAppearance
				}
			}

			for _, a := range c.Adapters {
				switch a.(type) {
				case config.SymlinkAdapter:
					err := adapter.ExecuteSymlinkAdapter(colorSchemePreference, a.(config.SymlinkAdapter))
					if err != nil {
						log.Printf("Executing the SymlinkAdapter: %v", err)
					}
				case config.TmuxAdapter:
					err := adapter.ExecuteTmuxAdapter(colorSchemePreference, a.(config.TmuxAdapter))
					if err != nil {
						log.Printf("Executing the TmuxAdapter: %v", err)
					}
				case config.AlacrittyAdapter:
					err := adapter.ExecuteAlacrittyAdapter(colorSchemePreference, a.(config.AlacrittyAdapter))
					if err != nil {
						log.Printf("Executing the AlacrittyAdapter: %v", err)
					}
				case config.KonsoleAdapter:
					err := adapter.ExecuteKonsoleAdapter(colorSchemePreference, a.(config.KonsoleAdapter))
					if err != nil {
						log.Printf("Executing the KonsoleAdapter: %v", err)
					}
				default:
					log.Fatalf("Unknown adapter: %T", a)
				}
			}
		}
	}
}
