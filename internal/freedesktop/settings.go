package freedesktop

import (
	"fmt"
	"github.com/godbus/dbus/v5"
)

type ColorSchemePreference int

const (
	NoPreference ColorSchemePreference = iota
	PreferDarkAppearance
	PreferLightAppearance
)

type SettingChanged struct {
	Namespace string
	Key       string
	Value     dbus.Variant
}

func ParseColorSchemePreference(value int) (ColorSchemePreference, error) {
	castedValue := ColorSchemePreference(value)

	if castedValue < NoPreference || castedValue > PreferLightAppearance {
		return NoPreference, fmt.Errorf("the %d is not a valid ColorSchemePreference value", value)
	}

	return castedValue, nil
}

func ParseSettingChangedSignal(v *dbus.Signal) (sc SettingChanged, err error) {
	if len(v.Body) == 0 {
		err = fmt.Errorf("parsing SettingsChanged: invalid body length")
		return
	}

	var ok bool
	sc.Namespace, ok = v.Body[0].(string)
	if !ok {
		err = fmt.Errorf("parsing SettingsChanged: unable to parse namespace from body")
		return
	}

	sc.Key, ok = v.Body[1].(string)
	if !ok {
		err = fmt.Errorf("parsing SettingsChanged: unable to parse key from body")
		return
	}

	sc.Value, ok = v.Body[2].(dbus.Variant)
	if !ok {
		err = fmt.Errorf("parsing SettingsChanged: unable to parse value from body")
		return
	}

	return
}
