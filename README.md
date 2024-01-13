<a href="https://1823.pl/#gh-light-mode-only">
  <img src="./.github/images/d1823.webp" align="right" alt="1823 logo" title="1823" height="60">
</a>

<a href="https://1823.pl/#gh-dark-mode-only">
  <img src="./.github/images/d1823-light.webp" align="right" alt="1823 logo" title="1823" height="60">
</a>

# README

Themer is a versatile command-line utility that monitors the org.freedesktop.appearance.color-scheme over DBus and seamlessly triggers color scheme switching in various applications, based on the user's configuration.

# Build
Use the attached Makefile. Run `make` to see the available options.

# Usage
## Installation
To install the application system-wide, run `sudo make install`. Alternatively, to install it just for the current user, run `PREFIX=~/.local SERVICE_PREFIX=~/.config/systemd/user XDG_CONFIG_HOME=~/.config make install`.

## Configuration
To use Themer, create a JSON configuration file at `$XDG_CONFIG_HOME/themer/config.json`, specifying the theme-switching adapters to execute. Here's an example configuration file:

```json
{
    "no_preference_fallback": "dark",
    "adapters": [
        {
            "adapter": "alacritty",
            "dark_preference_file": "/home/user/.config/alacritty/themes/dark-theme.yml",
            "light_preference_file": "/home/user/.config/alacritty/themes/light-theme.yml",
            "target_file": "/home/user/.config/alacritty/themes/current-theme.yml",
            "alacritty_config_file": "/home/user/.config/alacritty/alacritty.yml"
        },
        {
            "adapter": "tmux",
            "dark_preference_file": "/home/user/.config/tmux/themes/dark-theme.conf",
            "light_preference_file": "/home/user/.config/tmux/themes/light-theme.conf",
            "target_file": "/home/user/.config/tmux/themes/current-theme.conf",
            "tmux_config_file": "/home/user/.config/tmux/tmux.conf"
        },
        {
            "adapter": "konsole",
            "dark_profile_name": "Dark",
            "light_profile_name": "Light"
        }
    ]
}
```

This example configuration supports theme-switching adapters that execute when the org.freedesktop.appearance.color-scheme is modified, configuring theme switching for Alacritty, Tmux and Konsole.

### Available adapters

#### Symlink
It receives the current color-scheme preference, selects the correct the theme file and symlinks it to the target path.

#### Tmux
It uses the *symlink* adapter internally, but also executes `tmux source-file <tmux_config_file>` to make tmux reload its configuration.

Tmux requires a manual trigger to detect config changes, but the rest follows the same setup as in case of Alacritty.
Add the following line to your tmux configuration:

```conf
source-file ~/.config/tmux/themes/current-theme.conf
```

#### Alacritty
It uses the *symlink* adapter internally, but touches the `<alacritty_config_file>` to make alacritty reload its configuration.

Alacritty automatically detects config changes, but only on the main config file. Changing the included `/home/user/.config/tmux/themes/current-theme.conf` won't trigger the config reload. That's why this adapter touches the config file.
Add the following line to your Alacritty configuration:

```yaml
import:
 - ~/.config/alacritty/themes/current-theme.yml
```

#### Konsole
It utilizes the DBus integration to make Konsole switch to the configured profile.

To make this work, make sure you have two Konsole profiles configured for dark & light themes. Afterward, put their
names into the `dark_profile_name` and `light_profile_name` properties.

> :warning: **If you are running KDE**: Make sure to set the `no_preference_fallback` to "light".
> Otherwise, switching to the light color scheme won't be possible.
>
> [See why](https://github.com/d1823/themer/commit/2e341291ff4b169bfca0b240dec69c886366fb49).

## Setup

To autostart themer, use your desktop environment's configuration or the provided *.service file to register a new systemd service.
Make sure to use the correct unit-level dependency - `default.target` for user units, `multi-user.target` for system units.

# License
This project is licensed under the 3-Clause BSD license.
