# Karabingen

CLI tool to generate karabiner.json file from OVERsimplified yaml. Example of expected config:

```yaml
disable_command_tab: true # disables cmd + tab switches
fix_c_c: true # fix option-c usage: for fzf usage.
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

Tool has almost no validation of config. If something goes wrong check **Karabiner-Elements** -> **Settings** ->
**Log**.

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

Layers impl is taken from this dude: https://github.com/mxstbr/karabiner, vid with explanation:
https://www.youtube.com/watch?v=j4b_uQX3Vu0[
