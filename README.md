# Karabingen

CLI tool to generate karabiner.json file from OVERsimplified yaml. Example of expected config:

```yaml
version: 1
disable_command_tab: true # disables cmd + tab switches
disable_left_ctrl: true # disables left control key (useful with HHKB mode)
fix_c_c: true # fix option-c usage: for fzf usage.
use_hhkb: true # HHKB mode: maps Caps Lock to Left Control
hyperkey: caps_lock # key to use as hyperkey (caps_lock, right_command, right_option, right_shift, etc.)
fix_g502: # fixes back button of g502 mouse in safari
  enable: true # turn the rule on/off
  safari_only: true # only remap when Safari is frontmost (recommended)
  back_button: button4 # adjust if your EventViewer shows different codes
  forward_button: button5
tmux_jump:
  enable: true
  terminal: alacritty # or termianl, iterm2, ghosty
  jumplist_path: ~/tmuxjumplist.txt
  modifiers: ['right_command']
  all_letters: true
keybindings:
  option:
    '1':
      val: '/Applications/Safari.app'
      type: 'app'
    '2':
      val: '/Applications/Alacritty.app'
      type: 'app'
    '3':
      val: '/Applications/Bear.app'
      type: 'app'
  layers:
    - key: 'o'
      type: 'app'
      sub:
        't': '/Applications/Telegram.app'
        's': '/Applications/Safari.app'
        'c': '/Applications/Visual Studio Code.app'
        'b': '/Applications/Bear.app'
        'p': '/Applications/Postman.app'
        'f': '/System/Library/CoreServices/Finder.app'
        'z': '/Applications/Zen Browser.app'
    - key: 'w'
      type: 'web'
      sub:
        g: 'https://chatgpt.com/'
        r: 'https://reddit.com/'
        y: 'https://news.ycombinator.com'
        p: https://mxstbr.com/
        x: https://x.com/
```

Tool has no validation of config. If something goes wrong check **Karabiner-Elements** -> **Settings** -> **Log**.

## Configuration Options

### HHKB Mode

Set `use_hhkb: true` to map Caps Lock to Left Control, matching the Happy Hacking Keyboard layout. This completely
disables Caps Lock functionality, preventing it from being accidentally activated.

**Note:** If you enable HHKB mode and want to use a hyperkey, you must set `hyperkey` to something other than
`caps_lock` (see below).

### Disable Left Control

Set `disable_left_ctrl: true` to completely disable the physical left control key. This is particularly useful when
combined with HHKB mode, allowing you to rely solely on Caps Lock as your Control key.

**Example:** HHKB mode with disabled left control:

```yaml
use_hhkb: true
disable_left_ctrl: true
hyperkey: right_command
```

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

Using Brew:

```shell
brew install fgazat/tap/karabingen
```

Or using go:

```shell
go install github.com/fgazat/karabingen@latest
```

## Usage

```shell
karabingen generate [PATH_TO_YAML_CONFIG]
```

It will write to `~/.config/karabiner/karabiner.json` file.

## Credits

- [https://github.com/tekezo](https://github.com/tekezo)
- [https://github.com/pqrs-org/Karabiner-Elements](https://github.com/pqrs-org/Karabiner-Elements)
- Layers impl is taken from this dude: https://github.com/mxstbr/karabiner, vid with explanation
  https://www.youtube.com/watch?v=j4b_uQX3Vu0[
