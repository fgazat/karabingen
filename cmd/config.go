package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// KeyBinding represents a single key binding configuration
type KeyBinding struct {
	Type string                `yaml:"type"` // "app", "web", "shell", or "layer" (v2 only)
	Val  string                `yaml:"val"`
	Sub  map[string]KeyBinding `yaml:"sub"` // only used when Type == "layer"
}

// LayerConfig represents a hyperkey layer configuration
type LayerConfig struct {
	Key  string            `yaml:"key"`
	Type string            `yaml:"type"` // "app" or "web"
	Sub  map[string]string `yaml:"sub"`
}

// TmuxJumpConfig represents tmux session jumping configuration
type TmuxJumpConfig struct {
	Enable           bool     `yaml:"enable"`
	Modifiers        []string `yaml:"modifiers"`
	JumplistPath     string   `yaml:"jumplist_path"`
	Letters          []string `yaml:"letters"`
	AllLetters       bool     `yaml:"all_letters"`
	AllLettersExcept []string `yaml:"all_letters_except"`
	Terminal         string   `yaml:"terminal"`
	TmuxPath         string   `yaml:"tmux_path"`
}

// FixG502Config represents G502 mouse button remapping configuration
type FixG502Config struct {
	Enable        bool   `yaml:"enable"`
	SafariOnly    bool   `yaml:"safari_only"`
	BackButton    string `yaml:"back_button"`
	ForwardButton string `yaml:"forward_button"`
}

// QuickAppSwitchConfig represents quick app switching configuration
type QuickAppSwitchConfig struct {
	Enable     bool     `yaml:"enable"`
	Modifiers  []string `yaml:"modifiers"`
	BackKey    string   `yaml:"back_key"`    // Key for previous app (default: "o")
	ForwardKey string   `yaml:"forward_key"` // Key for next app (default: "i")
}

// KeybindingsConfig represents all keybindings configuration
type KeybindingsConfig struct {
	Option map[string]KeyBinding `yaml:"option"`
	Layers []LayerConfig         `yaml:"layers"`
}

// Config represents the complete configuration
type Config struct {
	Version            int                  `yaml:"version"`
	DisableCommandTab  bool                 `yaml:"disable_command_tab"`
	DisableLeftCtrl    bool                 `yaml:"disable_left_ctrl"`
	FixCC              bool                 `yaml:"fix_c_c"`
	UseHHKB            bool                 `yaml:"use_hhkb"`
	Hyperkey           string               `yaml:"hyperkey"`
	Keybindings        KeybindingsConfig    `yaml:"keybindings"`
	TmuxJump           TmuxJumpConfig       `yaml:"tmux_jump"`
	FixG502            FixG502Config        `yaml:"fix_g502"`
	SwitchSafariTabsHL bool                 `yaml:"switch_safari_tabs_hl"`
	QuickAppSwitch     QuickAppSwitchConfig `yaml:"quick_app_switch"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	// Set defaults
	config.Version = 1
	config.Hyperkey = "caps_lock"
	config.TmuxJump.Terminal = "alacritty"
	config.TmuxJump.Modifiers = []string{"option", "control"}
	config.TmuxJump.JumplistPath = "~/.tmuxjumplist"
	config.TmuxJump.TmuxPath = "/opt/homebrew/bin/tmux"
	config.FixG502.SafariOnly = true
	config.FixG502.BackButton = "button4"
	config.FixG502.ForwardButton = "button5"
	config.QuickAppSwitch.Modifiers = []string{"option"}
	config.QuickAppSwitch.BackKey = "o"
	config.QuickAppSwitch.ForwardKey = "i"

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate version
	if config.Version != 1 && config.Version != 2 {
		return nil, fmt.Errorf("unsupported config version: %d (supported: 1, 2)", config.Version)
	}

	// Validate option keybindings
	for key, binding := range config.Keybindings.Option {
		if binding.Type == "layer" {
			if config.Version < 2 {
				return nil, fmt.Errorf("option key '%s' has type 'layer' which requires version: 2", key)
			}
			if len(binding.Sub) == 0 {
				return nil, fmt.Errorf("option key '%s' has type 'layer' but no sub-bindings", key)
			}
			for subKey, subBinding := range binding.Sub {
				if subBinding.Type == "layer" {
					return nil, fmt.Errorf("option key '%s' sub-key '%s' cannot be a layer (nested layers not supported)", key, subKey)
				}
				if subBinding.Val == "" {
					return nil, fmt.Errorf("option key '%s' sub-key '%s' has empty val", key, subKey)
				}
			}
		}
	}

	// Process all_letters_except or all_letters
	if config.TmuxJump.AllLettersExcept != nil {
		allLetters := "abcdefghijklmnopqrstuvwxyz"
		excludeMap := make(map[rune]bool)
		for _, letter := range config.TmuxJump.AllLettersExcept {
			if len(letter) > 0 {
				excludeMap[rune(letter[0])] = true
			}
		}

		config.TmuxJump.Letters = []string{}
		for _, char := range allLetters {
			if !excludeMap[char] {
				config.TmuxJump.Letters = append(config.TmuxJump.Letters, string(char))
			}
		}
	} else if config.TmuxJump.AllLetters {
		config.TmuxJump.Letters = []string{}
		for char := 'a'; char <= 'z'; char++ {
			config.TmuxJump.Letters = append(config.TmuxJump.Letters, string(char))
		}
	}

	return &config, nil
}
