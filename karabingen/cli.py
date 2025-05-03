import json
import sys
from pathlib import Path

import yaml


def load_config(path="./config.yaml"):
    with open(path) as f:
        return yaml.safe_load(f)


def create_hyper_key_rule():
    return {
        "description": "Hyper Key",
        "manipulators": [
            {
                "description": "Caps Lock -> Hyper Key",
                "from": {"key_code": "caps_lock", "modifiers": {"optional": ["any"]}},
                "to": [{"set_variable": {"name": "hyper", "value": 1}}],
                "to_after_key_up": [{"set_variable": {"name": "hyper", "value": 0}}],
                "to_if_alone": [{"key_code": "escape"}],
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


def create_option_keybinding_rule(key, binding):
    to = {}
    if binding["type"] == "app":
        to = {"software_function": {"open_application": {"file_path": binding["val"]}}}
    elif binding["type"] == "web":
        to = {"shell_command": f"open {binding['val']}"}

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
    keybindings = config.get("keybingings", {})
    option_keybindings = keybindings.get("option", {})
    layers = keybindings.get("layers", [])

    profile = {
        "name": "base",
        "selected": True,
        "virtual_hid_keyboard": {"keyboard_type_v2": "iso"},
        "simple_modifications": [],
        "complex_modifications": {"rules": []},
    }

    if fix_c_c:
        profile["simple_modifications"].append(
            {
                "from": {"key_code": "grave_accent_and_tilde"},
                "to": [{"key_code": "non_us_backslash"}],
            }
        )

    rules = [create_hyper_key_rule()]

    if disable_command_tab:
        rules.append(create_disable_command_tab_rule())

    for key, binding in option_keybindings.items():
        rules.append(create_option_keybinding_rule(key, binding))

    rules.extend(create_layer_rules(layers))

    profile["complex_modifications"]["rules"] = rules

    karabiner_config = {
        "global": {"show_profile_name_in_menu_bar": True},
        "profiles": [profile],
    }

    home = Path.home()
    file_path = home / ".config" / "karabiner" / "karabiner.json"

    file_path.parent.mkdir(parents=True, exist_ok=True)
    with open(file_path, "w+") as f:
        json.dump(karabiner_config, f, indent=2)


if __name__ == "__main__":
    main()
