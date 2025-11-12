import argparse
import json
import os
import shutil
import sys
from datetime import datetime
from pathlib import Path

import yaml


def load_config(path="./config.yaml"):
    with open(path) as f:
        return yaml.safe_load(f)


def get_config_version(config):
    return config.get("version", 1)


def deploy_tmuxjump_script():
    """
    Deploy the bundled tmuxjump Python script to ~/.config/karabiner/scripts/tmuxjump.py
    Returns the path to the deployed script.
    """
    # Get the bundled script path
    package_dir = Path(__file__).parent.parent
    bundled_script = package_dir / "scripts" / "tmuxjump.py"

    if not bundled_script.exists():
        raise FileNotFoundError(f"Bundled tmuxjump script not found at {bundled_script}")

    # Target location
    target_dir = Path.home() / ".config" / "karabiner" / "scripts"
    target_dir.mkdir(parents=True, exist_ok=True)
    target_script = target_dir / "tmuxjump.py"

    shutil.copy2(bundled_script, target_script)
    target_script.chmod(0o755)
    print(f"Deployed tmuxjump script to: {target_script}")
    return str(target_script)


def create_hyper_key_rule(hyper_key="caps_lock"):
    """
    Create a hyperkey rule using the specified key.

    Args:
        hyper_key: The key to use as hyperkey. Options include:
                   - "caps_lock" (default)
                   - "right_command"
                   - "right_option"
                   - "right_shift"
                   - "tab" (hold)
                   - "return_or_enter" (hold)
                   - "grave_accent_and_tilde" (backtick)
    """
    return {
        "description": f"Hyper Key ({hyper_key})",
        "manipulators": [
            {
                "description": f"{hyper_key} -> Hyper Key",
                "from": {"key_code": hyper_key, "modifiers": {"optional": ["any"]}},
                "to": [{"set_variable": {"name": "hyper", "value": 1}}],
                "to_after_key_up": [{"set_variable": {"name": "hyper", "value": 0}}],
                "to_if_alone": [{"key_code": "escape"}],
                "type": "basic",
            }
        ],
    }


def create_hhkb_mode_rule():
    """
    HHKB mode: Map Caps Lock to Left Control.
    This is the standard HHKB (Happy Hacking Keyboard) layout.

    Note: This only remaps Caps Lock -> Control, not bidirectional.
    This prevents Caps Lock from ever being activated accidentally.
    """
    return {
        "description": "HHKB Mode (Caps Lock -> Left Control)",
        "manipulators": [
            {
                "description": "Caps Lock -> Left Control",
                "from": {"key_code": "caps_lock", "modifiers": {"optional": ["any"]}},
                "to": [{"key_code": "left_control"}],
                "type": "basic",
            },
        ],
    }


def create_disable_left_ctrl_rule():
    """
    Disable the left control key completely.
    Useful in combination with HHKB mode where Caps Lock becomes Control.
    """
    return {
        "description": "Disable Left Control",
        "manipulators": [
            {
                "description": "Left Control -> None",
                "from": {"key_code": "left_control", "modifiers": {"optional": ["any"]}},
                "to": [{"key_code": "vk_none"}],
                "type": "basic",
            }
        ],
    }


def create_disable_command_tab_rule():
    return {
        "description": "Disable Command + Tab",
        "manipulators": [
            {
                "from": {"key_code": "tab", "modifiers": {"mandatory": ["command"]}},
                "to": [{"key_code": "vk_none"}],
                "type": "basic",
            }
        ],
    }


def create_fix_g502_rule(safari_only=True, back_button="button4", forward_button="button5"):
    """
    Remap G502 side buttons to Safari Back/Forward (⌘[ / ⌘]) by default.
    If safari_only is True, the rule triggers only when Safari is frontmost.
    """
    conditions = []
    if safari_only:
        conditions = [
            {
                "type": "frontmost_application_if",
                "bundle_identifiers": ["^com\\.apple\\.Safari$"],
            }
        ]

    return {
        "description": "G502: map side buttons to Safari Back/Forward",
        "manipulators": [
            {
                "type": "basic",
                "from": {"pointing_button": back_button},
                "to": [{"key_code": "open_bracket", "modifiers": ["command"]}],
                "conditions": conditions,
                "description": "G502 Back → ⌘[",
            },
            {
                "type": "basic",
                "from": {"pointing_button": forward_button},
                "to": [{"key_code": "close_bracket", "modifiers": ["command"]}],
                "conditions": conditions,
                "description": "G502 Forward → ⌘]",
            },
        ],
    }


def create_option_keybinding_rule(key, binding):
    to = {}
    if binding["type"] == "app":
        to = {"software_function": {"open_application": {"file_path": binding["val"]}}}
    elif binding["type"] == "web":
        to = {"shell_command": f"open {binding['val']}"}
    elif binding["type"] == "shell":
        to = {"shell_command": binding["val"]}

    return {
        "description": "Open TBD",
        "manipulators": [
            {
                "from": {
                    "key_code": key,
                    "modifiers": {"mandatory": ["left_option"], "optional": ["caps_lock"]},
                },
                "to": [to],
                "type": "basic",
            }
        ],
    }


def hjkl():
    return {
        "description": "Map Option + H/J/K/L to Arrow Keys",
        "manipulators": [
            {
                "type": "basic",
                "from": {"key_code": "h", "modifiers": {"mandatory": ["option"]}},
                "to": [{"key_code": "left_arrow"}],
            },
            {
                "type": "basic",
                "from": {"key_code": "j", "modifiers": {"mandatory": ["option"]}},
                "to": [{"key_code": "down_arrow"}],
            },
            {
                "type": "basic",
                "from": {"key_code": "k", "modifiers": {"mandatory": ["option"]}},
                "to": [{"key_code": "up_arrow"}],
            },
            {
                "type": "basic",
                "from": {"key_code": "l", "modifiers": {"mandatory": ["option"]}},
                "to": [{"key_code": "right_arrow"}],
            },
            {
                "type": "basic",
                "from": {"key_code": "m", "modifiers": {"mandatory": ["option"]}},
                "to": [{"key_code": "return_or_enter"}],
            },
        ],
    }


def create_tmux_jump_rule(
    script_path="~/bin/tmuxjump.py",
    modifiers=None,
    tmuxjumplist_path="~/tmuxjumplist",
    letters=None,
    all_letters=False,
    all_letters_except=None,
    terminal="alacritty",
):
    """
    Create rules for tmux session jumping with digits 1-9, 0 for editing tmuxjumplist, and optional letters.
    Uses option+control by default for easier pressing.
    If all_letters=True, creates rules for all a-z letters automatically.
    If all_letters_except is provided, creates rules for all letters except the specified ones.
    Calls Python script directly for better cross-platform compatibility.
    Supports multiple terminals: alacritty, iterm2, terminal, ghostty.
    """
    if modifiers is None:
        modifiers = ["option", "control"]
    if letters is None:
        letters = []

    # If all_letters_except is provided, use all letters except the specified ones
    if all_letters_except is not None:
        all_alphabet = set("abcdefghijklmnopqrstuvwxyz")
        excluded = set(all_letters_except)
        letters = sorted(list(all_alphabet - excluded))
    # If all_letters is True, add all a-z to the letters list
    elif all_letters:
        letters = list("abcdefghijklmnopqrstuvwxyz")

    manipulators = []

    # 0 opens the tmuxjumplist file in the terminal editor
    # Use the same terminal as configured for jumping
    if terminal == "iterm2":
        edit_command = f'osascript -e \'tell application "iTerm" to create window with default profile command "nvim {tmuxjumplist_path}"\''
    elif terminal == "terminal":
        edit_command = f'osascript -e \'tell application "Terminal" to do script "nvim {tmuxjumplist_path}"\''
    else:  # alacritty, ghostty, or fallback
        # Try to use tmux if available, otherwise open in new terminal window
        edit_command = f"/opt/homebrew/bin/tmux new-window 'nvim {tmuxjumplist_path}' 2>/dev/null || /usr/bin/env python3 {script_path} 0 {tmuxjumplist_path} {terminal}"

    zero_manipulator = {
        "type": "basic",
        "from": {"key_code": "0", "modifiers": {"mandatory": modifiers}},
        "to": [{"shell_command": edit_command}],
        "description": f"{'+'.join([m.capitalize() for m in modifiers])}+0 → edit tmuxjumplist",
    }
    manipulators.append(zero_manipulator)

    # 1-9 jump to tmux sessions
    for digit in ["1", "2", "3", "4", "5", "6", "7", "8", "9"]:
        manipulator = {
            "type": "basic",
            "from": {"key_code": digit, "modifiers": {"mandatory": modifiers}},
            "to": [{"shell_command": f"/usr/bin/env python3 {script_path} {digit} {tmuxjumplist_path} {terminal}"}],
            "description": f"{'+'.join([m.capitalize() for m in modifiers])}+{digit} → tmux session {digit}",
        }
        manipulators.append(manipulator)

    # Letters jump to tmux sessions
    for letter in letters:
        manipulator = {
            "type": "basic",
            "from": {"key_code": letter, "modifiers": {"mandatory": modifiers}},
            "to": [{"shell_command": f"/usr/bin/env python3 {script_path} {letter} {tmuxjumplist_path} {terminal}"}],
            "description": f"{'+'.join([m.capitalize() for m in modifiers])}+{letter} → tmux session {letter}",
        }
        manipulators.append(manipulator)

    return {
        "description": f"{'+'.join([m.capitalize() for m in modifiers])}+Key → tmux session jump ({terminal})",
        "manipulators": manipulators,
    }


def create_layer_rules(layers):
    rules = []
    all_layer_keys = [layer["key"] for layer in layers]

    for layer in layers:
        key = layer["key"]
        sub_bindings = layer["sub"]
        layer_type = layer["type"]

        toggle_rule = {
            "description": f'Hyper Key sublayer "{key}"',
            "manipulators": [
                {
                    "description": f"Toggle Hyper sublayer {key}",
                    "from": {"key_code": key, "modifiers": {"optional": ["any"]}},
                    "to": [{"set_variable": {"name": f"hyper_sublayer_{key}", "value": 1}}],
                    "to_after_key_up": [{"set_variable": {"name": f"hyper_sublayer_{key}", "value": 0}}],
                    "type": "basic",
                    "conditions": [{"name": "hyper", "type": "variable_if", "value": 1}]
                    + [
                        {"name": f"hyper_sublayer_{k}", "type": "variable_if", "value": 0}
                        for k in all_layer_keys
                        if k != key
                    ],
                },
            ],
        }

        for subkey, val in sub_bindings.items():
            to = {}
            if layer_type == "app":
                to = {"software_function": {"open_application": {"file_path": val}}}
            elif layer_type == "web":
                to = {"shell_command": f"open {val}"}

            toggle_rule["manipulators"].append(
                {
                    "description": "Open ",
                    "from": {"key_code": subkey, "modifiers": {"optional": ["any"]}},
                    "to": [to],
                    "type": "basic",
                    "conditions": [{"name": f"hyper_sublayer_{key}", "type": "variable_if", "value": 1}],
                }
            )

        rules.append(toggle_rule)

    return rules


def create_switch_tabs_rule():
    """
    Remap ⌘+⌥+H/L to switch tabs.
    h -> Previous Tab  (⌃+⇧+Tab)
    l -> Next Tab  (⌃+Tab)
    """
    return {
        "description": "Remap ⌘+⌥+H/L to switch tabs",
        "manipulators": [
            {
                "type": "basic",
                "from": {
                    "key_code": "l",
                    "modifiers": {"mandatory": ["command", "option"], "optional": ["any"]},
                },
                "to": [{"key_code": "tab", "modifiers": ["control"]}],
                "description": "⌘+⌥+l → Next Tab (⌃+Tab)",
            },
            {
                "type": "basic",
                "from": {
                    "key_code": "h",
                    "modifiers": {"mandatory": ["command", "option"], "optional": ["any"]},
                },
                "to": [{"key_code": "tab", "modifiers": ["control", "shift"]}],
                "description": "⌘+⌥+h → Previous Tab (⌃+⇧+Tab)",
            },
        ],
    }


def parse_config_v1(config):
    """
    Parse configuration file version 1.
    This maintains backward compatibility with existing configs.
    """
    disable_command_tab = config.get("disable_command_tab", False)
    disable_left_ctrl = config.get("disable_left_ctrl", False)
    fix_c_c = config.get("fix_c_c")
    use_hhkb = config.get("use_hhkb", False)
    hyperkey = config.get("hyperkey", "caps_lock")
    keybindings = config.get("keybingings", {})
    option_keybindings = keybindings.get("option", {})
    layers = keybindings.get("layers", [])

    # Tmux jump configuration
    tmux_cfg = config.get("tmux_jump", {})
    enable_tmux = tmux_cfg.get("enable", False) if isinstance(tmux_cfg, dict) else config.get("enable_tmux", False)

    # Auto-deploy bundled script if no script_path is specified and tmux is enabled
    if isinstance(tmux_cfg, dict) and "script_path" in tmux_cfg:
        tmux_script_path = tmux_cfg["script_path"]
    elif enable_tmux:
        # Deploy the bundled script
        tmux_script_path = deploy_tmuxjump_script()
    else:
        tmux_script_path = "~/bin/tmuxjump.py"

    tmux_modifiers = (
        tmux_cfg.get("modifiers", ["option", "control"]) if isinstance(tmux_cfg, dict) else ["option", "control"]
    )
    tmux_jumplist_path = (
        tmux_cfg.get("jumplist_path", "~/.tmuxjumplist") if isinstance(tmux_cfg, dict) else "~/tmuxjumplist"
    )
    tmux_letters = tmux_cfg.get("letters", []) if isinstance(tmux_cfg, dict) else []
    tmux_all_letters = tmux_cfg.get("all_letters", False) if isinstance(tmux_cfg, dict) else False
    tmux_all_letters_except = tmux_cfg.get("all_letters_except", None) if isinstance(tmux_cfg, dict) else None
    tmux_terminal = tmux_cfg.get("terminal", "alacritty") if isinstance(tmux_cfg, dict) else "alacritty"

    # Fix G502 configuration
    fix_g502_cfg = config.get("fix_g502", {})
    switch_tabs_hl = config.get("switch_safari_tabs_hl", False)

    return {
        "disable_command_tab": disable_command_tab,
        "disable_left_ctrl": disable_left_ctrl,
        "fix_c_c": fix_c_c,
        "use_hhkb": use_hhkb,
        "hyperkey": hyperkey,
        "option_keybindings": option_keybindings,
        "layers": layers,
        "enable_tmux": enable_tmux,
        "tmux_script_path": tmux_script_path,
        "tmux_modifiers": tmux_modifiers,
        "tmux_tmuxjumplist_path": tmux_jumplist_path,
        "tmux_letters": tmux_letters,
        "tmux_all_letters": tmux_all_letters,
        "tmux_all_letters_except": tmux_all_letters_except,
        "tmux_terminal": tmux_terminal,
        "fix_g502_cfg": fix_g502_cfg,
        "switch_tabs_hl": switch_tabs_hl,
    }


def parse_config(config):
    """
    Parse configuration file based on version.
    Routes to appropriate version-specific parser.
    """
    version = get_config_version(config)

    if version == 1:
        return parse_config_v1(config)
    else:
        raise ValueError(f"Unsupported config version: {version}. Supported versions: 1")


def main():
    parser = argparse.ArgumentParser(
        description="Generate Karabiner-Elements configuration from YAML config file",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument("config_path", help="Path to the YAML configuration file")
    parser.add_argument(
        "-o",
        "--output",
        dest="output_path",
        help="Path to output karabiner.json file (default: ~/.config/karabiner/karabiner.json)",
    )
    parser.add_argument("--no-backup", action="store_true", help="Skip creating backup of existing karabiner.json file")

    args = parser.parse_args()

    config = load_config(args.config_path)
    parsed = parse_config(config)

    # Extract parsed values
    disable_command_tab = parsed["disable_command_tab"]
    disable_left_ctrl = parsed["disable_left_ctrl"]
    fix_c_c = parsed["fix_c_c"]
    use_hhkb = parsed["use_hhkb"]
    hyperkey = parsed["hyperkey"]
    option_keybindings = parsed["option_keybindings"]
    layers = parsed["layers"]
    enable_tmux = parsed["enable_tmux"]
    tmux_script_path = parsed["tmux_script_path"]
    tmux_modifiers = parsed["tmux_modifiers"]
    tmux_tmuxjumplist_path = parsed["tmux_tmuxjumplist_path"]
    tmux_letters = parsed["tmux_letters"]
    tmux_all_letters = parsed["tmux_all_letters"]
    tmux_all_letters_except = parsed["tmux_all_letters_except"]
    tmux_terminal = parsed["tmux_terminal"]
    fix_g502_cfg = parsed["fix_g502_cfg"]
    switch_tabs_hl_cfg = parsed["switch_tabs_hl"]

    # Determine output path
    if args.output_path:
        file_path = Path(args.output_path)
    else:
        home = Path.home()
        file_path = home / ".config" / "karabiner" / "karabiner.json"

    # Load existing karabiner config to preserve devices and other settings
    existing_config = {}
    if file_path.exists():
        try:
            with open(file_path, "r") as f:
                existing_config = json.load(f)
        except (json.JSONDecodeError, FileNotFoundError):
            pass

    profile = {
        "name": "base",
        "selected": True,
        "virtual_hid_keyboard": {"keyboard_type_v2": "iso"},
        "simple_modifications": [],
        "complex_modifications": {"rules": []},
    }

    # Preserve existing devices configuration if it exists
    existing_profile = None
    if "profiles" in existing_config:
        for p in existing_config["profiles"]:
            if p.get("name") == "base":
                existing_profile = p
                break

    if existing_profile and "devices" in existing_profile:
        profile["devices"] = existing_profile["devices"]

    if fix_c_c:
        profile["simple_modifications"].append(
            {
                "from": {"key_code": "grave_accent_and_tilde"},
                "to": [{"key_code": "non_us_backslash"}],
            }
        )

    rules = []

    # Add HHKB mode if requested (maps caps lock to left control)
    # Note: HHKB mode and hyperkey are mutually exclusive if both use caps_lock
    if use_hhkb:
        rules.append(create_hhkb_mode_rule())
        # If hyperkey is set to caps_lock but HHKB is enabled, skip hyperkey rule
        if hyperkey != "caps_lock":
            rules.append(create_hyper_key_rule(hyperkey))
    else:
        rules.append(create_hyper_key_rule(hyperkey))

    # Disable left control if requested (useful with HHKB mode)
    if disable_left_ctrl:
        rules.append(create_disable_left_ctrl_rule())

    if fix_g502_cfg.get("enable", False):
        rules.append(
            create_fix_g502_rule(
                safari_only=fix_g502_cfg.get("safari_only", True),
                back_button=fix_g502_cfg.get("back_button", "button4"),
                forward_button=fix_g502_cfg.get("forward_button", "button5"),
            )
        )
    if switch_tabs_hl_cfg:
        rules.append(create_switch_tabs_rule())

    if disable_command_tab:
        rules.append(create_disable_command_tab_rule())

    if enable_tmux:
        rules.append(
            create_tmux_jump_rule(
                script_path=tmux_script_path,
                modifiers=tmux_modifiers,
                tmuxjumplist_path=tmux_tmuxjumplist_path,
                letters=tmux_letters,
                all_letters=tmux_all_letters,
                all_letters_except=tmux_all_letters_except,
                terminal=tmux_terminal,
            )
        )

    for key, binding in option_keybindings.items():
        rules.append(create_option_keybinding_rule(key, binding))

    rules.append(hjkl())
    rules.extend(create_layer_rules(layers))

    profile["complex_modifications"]["rules"] = rules

    karabiner_config = {
        "global": existing_config.get("global", {"show_profile_name_in_menu_bar": True}),
        "profiles": [profile],
    }

    file_path.parent.mkdir(parents=True, exist_ok=True)

    # Create backup of existing file if it exists and backup is not disabled
    if not args.no_backup and file_path.exists():
        timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
        backup_name = f"backup_{timestamp}.json"
        backup_path = file_path.parent / backup_name
        shutil.copy2(file_path, backup_path)
        print(f"Backup created: {backup_path}")

    with open(file_path, "w+") as f:
        json.dump(karabiner_config, f, indent=2)


if __name__ == "__main__":
    main()
