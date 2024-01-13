package config

import (
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
)

// SymlinkAdapter represents the Symlink adapter configuration.
type SymlinkAdapter struct {
	NoPreferenceFile    string `mapstructure:"no_preference_file"`
	DarkPreferenceFile  string `mapstructure:"dark_preference_file"`
	LightPreferenceFile string `mapstructure:"light_preference_file"`
	TargetFile          string `mapstructure:"target_file"`
}

// TmuxAdapter represents the Tmux adapter configuration.
type TmuxAdapter struct {
	SymlinkAdapter `mapstructure:",squash"`
	TmuxConfigFile string `mapstructure:"tmux_config_file"`
}

// AlacrittyAdapter represents the Alacritty adapter configuration.
type AlacrittyAdapter struct {
	SymlinkAdapter      `mapstructure:",squash"`
	AlacrittyConfigFile string `mapstructure:"alacritty_config_file"`
}

// KonsoleAdapter represents the Konsole adapter configuration.
type KonsoleAdapter struct {
	SymlinkAdapter          `mapstructure:",squash"`
	NoPreferenceProfileName string `mapstructure:"no_preference_profile_name"`
	DarkProfileName         string `mapstructure:"dark_profile_name"`
	LightProfileName        string `mapstructure:"light_profile_name"`
}

// Configuration represents the top-level configuration structure.
type Configuration struct {
	Adapters []interface{} `json:"adapters"`
}

// UnmarshalJSON customizes the JSON unmarshaling process for Configuration.
func (c *Configuration) UnmarshalJSON(data []byte) error {
	var configData map[string]interface{}
	if err := json.Unmarshal(data, &configData); err != nil {
		return fmt.Errorf("failed to unmarshal configuration JSON data: %v", err)
	}

	adaptersData, ok := configData["adapters"].([]interface{})
	if !ok {
		return fmt.Errorf("configuration JSON does not contain an 'adapters' array")
	}

	for _, adapterData := range adaptersData {
		adapter, ok := adapterData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid adapter data in 'adapters' array")
		}

		switch adapter["adapter"] {
		case "symlink":
			symlink := SymlinkAdapter{}
			if err := mapstructure.Decode(adapter, &symlink); err != nil {
				return fmt.Errorf("failed to decode 'symlink' adapter: %v", err)
			}
			c.Adapters = append(c.Adapters, symlink)
		case "tmux":
			tmux := TmuxAdapter{}
			if err := mapstructure.Decode(adapter, &tmux); err != nil {
				return fmt.Errorf("failed to decode 'tmux' adapter: %v", err)
			}
			c.Adapters = append(c.Adapters, tmux)
		case "alacritty":
			alacritty := AlacrittyAdapter{}
			if err := mapstructure.Decode(adapter, &alacritty); err != nil {
				return fmt.Errorf("failed to decode 'alacritty' adapter: %v", err)
			}
			c.Adapters = append(c.Adapters, alacritty)
		case "konsole":
			konsole := KonsoleAdapter{}
			if err := mapstructure.Decode(adapter, &konsole); err != nil {
				return fmt.Errorf("failed to decode 'konsole' adapter: %v", err)
			}
			c.Adapters = append(c.Adapters, konsole)
		default:
			return fmt.Errorf("unknown adapter type: %v", adapter["adapter"])
		}
	}

	return nil
}

// Parse parses a JSON byte slice into a Configuration.
func Parse(jsonData []byte) (*Configuration, error) {
	var config Configuration

	if err := json.Unmarshal(jsonData, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON data into Configuration: %v", err)
	}

	return &config, nil
}
