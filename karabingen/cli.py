import json
import sys
from pathlib import Path

import yaml


def load_config(path="./config.yaml"):
    with open(path) as f:
        return yaml.safe_load(f)


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


def create_tmux_jump_rule(script_path="~/bin/tmuxjump.sh", modifiers=None, tmf_path="~/tmf", letters=None, all_letters=False):
    """
    Create rules for tmux session jumping with digits 1-9, 0 for editing tmf, and optional letters.
    Uses option+control by default for easier pressing.
    If all_letters=True, creates rules for all a-z letters automatically.
    """
    if modifiers is None:
        modifiers = ["option", "control"]
    if letters is None:
        letters = []

    # If all_letters is True, add all a-z to the letters list
    if all_letters:
        letters = list("abcdefghijklmnopqrstuvwxyz")

    manipulators = []

    # 0 opens a temporary tmux window to edit ~/tmf
    zero_manipulator = {
        "type": "basic",
        "from": {"key_code": "0", "modifiers": {"mandatory": modifiers}},
        "to": [{"shell_command": f"/usr/bin/env zsh -lc 'tmux new-window \"nvim {tmf_path}\"'"}],
        "description": f"{'+'.join([m.capitalize() for m in modifiers])}+0 → edit tmf in nvim",
    }
    manipulators.append(zero_manipulator)

    # 1-9 jump to tmux sessions
    for digit in ["1", "2", "3", "4", "5", "6", "7", "8", "9"]:
        manipulator = {
            "type": "basic",
            "from": {"key_code": digit, "modifiers": {"mandatory": modifiers}},
            "to": [{"shell_command": f"/usr/bin/env zsh -lc '{script_path} {digit}'"}],
            "description": f"{'+'.join([m.capitalize() for m in modifiers])}+{digit} → tmux session {digit}",
        }
        manipulators.append(manipulator)

    # Letters jump to tmux sessions
    for letter in letters:
        manipulator = {
            "type": "basic",
            "from": {"key_code": letter, "modifiers": {"mandatory": modifiers}},
            "to": [{"shell_command": f"/usr/bin/env zsh -lc '{script_path} {letter}'"}],
            "description": f"{'+'.join([m.capitalize() for m in modifiers])}+{letter} → tmux session {letter}",
        }
        manipulators.append(manipulator)

    return {
        "description": f"{'+'.join([m.capitalize() for m in modifiers])}+Key → tmux session jump",
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


def main():
    if len(sys.argv) == 1:
        print("pass config path")
        exit(1)

    config = load_config(sys.argv[1])
    disable_command_tab = config.get("disable_command_tab", False)
    fix_c_c = config.get("fix_c_c")
    use_hhkb = config.get("use_hhkb", False)
    hyperkey = config.get("hyperkey", "caps_lock")
    keybindings = config.get("keybingings", {})
    option_keybindings = keybindings.get("option", {})
    layers = keybindings.get("layers", [])

    # Tmux jump configuration
    tmux_cfg = config.get("tmux_jump", {})
    enable_tmux = tmux_cfg.get("enable", False) if isinstance(tmux_cfg, dict) else config.get("enable_tmux", False)
    tmux_script_path = (
        tmux_cfg.get("script_path", "~/bin/tmuxjump.sh") if isinstance(tmux_cfg, dict) else "~/bin/tmuxjump.sh"
    )
    tmux_modifiers = (
        tmux_cfg.get("modifiers", ["option", "control"]) if isinstance(tmux_cfg, dict) else ["option", "control"]
    )
    tmux_tmf_path = tmux_cfg.get("tmf_path", "~/tmf") if isinstance(tmux_cfg, dict) else "~/tmf"
    tmux_letters = tmux_cfg.get("letters", []) if isinstance(tmux_cfg, dict) else []
    tmux_all_letters = tmux_cfg.get("all_letters", False) if isinstance(tmux_cfg, dict) else False

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

    # Add HHKB mode if requested (swaps caps lock with left control)
    # Note: HHKB mode and hyperkey are mutually exclusive if both use caps_lock
    if use_hhkb:
        rules.append(create_hhkb_mode_rule())
        # If hyperkey is set to caps_lock but HHKB is enabled, skip hyperkey rule
        if hyperkey != "caps_lock":
            rules.append(create_hyper_key_rule(hyperkey))
    else:
        rules.append(create_hyper_key_rule(hyperkey))

    fix_g502_cfg = config.get("fix_g502", {})
    if fix_g502_cfg.get("enable", False):
        rules.append(
            create_fix_g502_rule(
                safari_only=fix_g502_cfg.get("safari_only", True),
                back_button=fix_g502_cfg.get("back_button", "button4"),
                forward_button=fix_g502_cfg.get("forward_button", "button5"),
            )
        )

    if disable_command_tab:
        rules.append(create_disable_command_tab_rule())

    if enable_tmux:
        rules.append(
            create_tmux_jump_rule(
                script_path=tmux_script_path,
                modifiers=tmux_modifiers,
                tmf_path=tmux_tmf_path,
                letters=tmux_letters,
                all_letters=tmux_all_letters,
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
    with open(file_path, "w+") as f:
        json.dump(karabiner_config, f, indent=2)


if __name__ == "__main__":
    main()
