# Karabingen

CLI tool to generate karabiner.json file from OVERsimplified yaml. Example of expected config:

```yaml
disable_command_tab: true # disables cmd + tab switches
fix_c_c: true # fix option-c usage: for fzf usage.
use_hhkb: true # HHKB mode: swaps Caps Lock with Left Control
hyperkey: caps_lock # key to use as hyperkey (caps_lock, right_command, right_option, right_shift, etc.)
fix_g502: # fixes back button of g502 mouse in safari
  enable: true           # turn the rule on/off
  safari_only: true      # only remap when Safari is frontmost (recommended)
  back_button: button4   # adjust if your EventViewer shows different codes
  forward_button: button5
keybingings:
  option: # option + key keybindings
    '1':
      val: /Applications/Zen Browser.app
      type: app
    '2':
      val: /Applications/Alacritty.app
      type: app
  layers:
    - key: o # caps_lock + o + subkey
      type: app
      sub:
        t: /Applications/Telegram.app
        s: /Applications/Safari.app
        b: /Applications/Bear.app
        p: /Applications/Postman.app
    - key: w # caps_lock + w + subkey
      type: web
      sub:
        g: https://chatgpt.com/
        r: https://reddit.com/
        y: https://news.ycombinator.com
        p: https://mxstbr.com/
        x: https://x.com/
```

Tool has no validation of config. If something goes wrong check **Karabiner-Elements** -> **Settings** ->
**Log**.

## Configuration Options

### HHKB Mode
Set `use_hhkb: true` to map Caps Lock to Left Control, matching the Happy Hacking Keyboard layout. This completely disables Caps Lock functionality, preventing it from being accidentally activated.

**Note:** If you enable HHKB mode and want to use a hyperkey, you must set `hyperkey` to something other than `caps_lock` (see below).

### Hyperkey Options
The `hyperkey` setting lets you choose which key becomes your hyperkey. Available options:
- `caps_lock` (default) - Most common choice
- `right_command` - **Recommended alternative** - rarely used, easy to reach with right thumb
- `right_option` - Good if you don't need international character input
- `right_shift` - Comfortable for touch typists
- `tab` - Hold for hyper, tap for tab
- `return_or_enter` - Hold for hyper, tap for enter
- `grave_accent_and_tilde` - If you rarely use the backtick key

**Example:** Using HHKB mode with right_command as hyperkey:
```yaml
use_hhkb: true
hyperkey: right_command
```

## Instalation

```shell
pipx install karabingen
```

## Usage

```shell
karabingen [PATH_TO_YAML_CONFIG]
```

It will write to `~/.config/karabiner/karabiner.json` file.

## Credits
* [https://github.com/tekezo](https://github.com/tekezo)
* [https://github.com/pqrs-org/Karabiner-Elements](https://github.com/pqrs-org/Karabiner-Elements)
* Layers impl is taken from this dude: https://github.com/mxstbr/karabiner, vid with explanation
https://www.youtube.com/watch?v=j4b_uQX3Vu0[
