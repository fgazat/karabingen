package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func createHyperKeyRule(hyperKey string) Rule {
	return Rule{
		Description: fmt.Sprintf("Hyper Key (%s)", hyperKey),
		Manipulators: []Manipulator{
			{
				Type:        "basic",
				Description: fmt.Sprintf("%s -> Hyper Key", hyperKey),
				From: From{
					KeyCode:   hyperKey,
				},
				To: []To{
					{SetVariable: &SetVariable{Name: "hyper", Value: 1}},
				},
				ToAfterKeyUp: []To{
					{SetVariable: &SetVariable{Name: "hyper", Value: 0}},
				},
				ToIfAlone: []To{
					{KeyCode: "escape"},
				},
			},
		},
	}
}

func createHHKBModeRule() Rule {
	return Rule{
		Description: "HHKB Mode (Caps Lock -> Left Control)",
		Manipulators: []Manipulator{
			{
				Type:        "basic",
				Description: "Caps Lock -> Left Control",
				From: From{
					KeyCode: "caps_lock",
				},
				To: []To{
					{KeyCode: "left_control"},
				},
			},
		},
	}
}

func createDisableLeftCtrlRule() Rule {
	return Rule{
		Description: "Disable Left Control",
		Manipulators: []Manipulator{
			{
				Type:        "basic",
				Description: "Left Control -> None",
				From: From{
					KeyCode: "left_control",
				},
				To: []To{
					{KeyCode: "vk_none"},
				},
			},
		},
	}
}

func createDisableCommandTabRule() Rule {
	return Rule{
		Description: "Disable Command + Tab",
		Manipulators: []Manipulator{
			{
				Type: "basic",
				From: From{
					KeyCode: "tab",
					Modifiers: &Modifiers{
						Mandatory: []string{"command"},
					},
				},
				To: []To{
					{KeyCode: "vk_none"},
				},
			},
		},
	}
}

func createFixG502Rule(safariOnly bool, backButton, forwardButton string) Rule {
	var conditions []Condition
	if safariOnly {
		conditions = []Condition{
			{
				Type:              "frontmost_application_if",
				BundleIdentifiers: []string{"^com\\.apple\\.Safari$"},
			},
		}
	}

	return Rule{
		Description: "G502: map side buttons to Safari Back/Forward",
		Manipulators: []Manipulator{
			{
				Type:        "basic",
				Description: "G502 Back → ⌘[",
				From: From{
					PointingButton: backButton,
				},
				To: []To{
					{KeyCode: "open_bracket", Modifiers: []string{"command"}},
				},
				Conditions: conditions,
			},
			{
				Type:        "basic",
				Description: "G502 Forward → ⌘]",
				From: From{
					PointingButton: forwardButton,
				},
				To: []To{
					{KeyCode: "close_bracket", Modifiers: []string{"command"}},
				},
				Conditions: conditions,
			},
		},
	}
}

func createOptionKeybindingRule(key string, binding KeyBinding) Rule {
	var to To

	switch binding.Type {
	case "app":
		to = To{
			SoftwareFunction: &SoftwareFunction{
				OpenApplication: &OpenApplication{
					FilePath: binding.Val,
				},
			},
		}
	case "web":
		to = To{
			ShellCommand: fmt.Sprintf("open %s", binding.Val),
		}
	case "shell":
		to = To{
			ShellCommand: binding.Val,
		}
	}

	return Rule{
		Description: "Open TBD",
		Manipulators: []Manipulator{
			{
				Type: "basic",
				From: From{
					KeyCode: key,
					Modifiers: &Modifiers{
						Mandatory: []string{"left_option"},
						Optional:  []string{"caps_lock"},
					},
				},
				To: []To{to},
			},
		},
	}
}

func createHJKLRule() Rule {
	return Rule{
		Description: "Map Option + H/J/K/L to Arrow Keys",
		Manipulators: []Manipulator{
			{
				Type: "basic",
				From: From{
					KeyCode:   "h",
					Modifiers: &Modifiers{Mandatory: []string{"option"}},
				},
				To: []To{{KeyCode: "left_arrow"}},
			},
			{
				Type: "basic",
				From: From{
					KeyCode:   "j",
					Modifiers: &Modifiers{Mandatory: []string{"option"}},
				},
				To: []To{{KeyCode: "down_arrow"}},
			},
			{
				Type: "basic",
				From: From{
					KeyCode:   "k",
					Modifiers: &Modifiers{Mandatory: []string{"option"}},
				},
				To: []To{{KeyCode: "up_arrow"}},
			},
			{
				Type: "basic",
				From: From{
					KeyCode:   "l",
					Modifiers: &Modifiers{Mandatory: []string{"option"}},
				},
				To: []To{{KeyCode: "right_arrow"}},
			},
			{
				Type: "basic",
				From: From{
					KeyCode:   "m",
					Modifiers: &Modifiers{Mandatory: []string{"option"}},
				},
				To: []To{{KeyCode: "return_or_enter"}},
			},
		},
	}
}

func createTmuxJumpRule(config *Config) (Rule, error) {
	tmuxConfig := config.TmuxJump

	// Get the path to karabingen executable
	executable, err := os.Executable()
	if err != nil {
		return Rule{}, fmt.Errorf("failed to get executable path: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return Rule{}, fmt.Errorf("failed to get absolute path: %w", err)
	}

	manipulators := []Manipulator{}

	// Create the base command
	baseCmd := fmt.Sprintf("%s tmux switch --tmux %s --jumplist %s --terminal %s",
		executable,
		tmuxConfig.TmuxPath,
		tmuxConfig.JumplistPath,
		tmuxConfig.Terminal,
	)

	modifierNames := make([]string, len(tmuxConfig.Modifiers))
	for i, mod := range tmuxConfig.Modifiers {
		modifierNames[i] = strings.Title(mod)
	}
	modStr := strings.Join(modifierNames, "+")

	// 0 opens the tmuxjumplist file in a new terminal window
	// Find the editor executable with full path
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	// Try to find the full path to the editor
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		// Fallback to common locations
		commonPaths := []string{
			"/opt/homebrew/bin/" + editor,
			"/usr/local/bin/" + editor,
			"/usr/bin/" + editor,
		}
		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				editorPath = path
				break
			}
		}
		if editorPath == "" {
			editorPath = editor // Use as-is if we can't find it
		}
	}

	// Expand tilde in jumplist path
	jumplistPath := tmuxConfig.JumplistPath
	if strings.HasPrefix(jumplistPath, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			jumplistPath = filepath.Join(home, jumplistPath[2:])
		}
	}

	var editCmd string
	switch tmuxConfig.Terminal {
	case "alacritty":
		editCmd = fmt.Sprintf("open -a Alacritty -n --args -e %s %s", editorPath, jumplistPath)
	case "ghostty":
		editCmd = fmt.Sprintf("open -a Ghostty -n --args -e %s %s", editorPath, jumplistPath)
	case "iterm2":
		editCmd = fmt.Sprintf(`osascript -e 'tell application "iTerm" to create window with default profile command "%s %s"'`, editorPath, jumplistPath)
	case "terminal":
		editCmd = fmt.Sprintf(`osascript -e 'tell application "Terminal" to do script "%s %s"'`, editorPath, jumplistPath)
	default:
		editCmd = fmt.Sprintf("open -a Alacritty -n --args -e %s %s", editorPath, jumplistPath)
	}

	manipulators = append(manipulators, Manipulator{
		Type: "basic",
		From: From{
			KeyCode:   "0",
			Modifiers: &Modifiers{Mandatory: tmuxConfig.Modifiers},
		},
		To: []To{
			{ShellCommand: editCmd},
		},
		Description: fmt.Sprintf("%s+0 → edit tmuxjumplist", modStr),
	})

	// 1-9 jump to tmux sessions
	for i := 1; i <= 9; i++ {
		digit := fmt.Sprintf("%d", i)
		manipulators = append(manipulators, Manipulator{
			Type: "basic",
			From: From{
				KeyCode:   digit,
				Modifiers: &Modifiers{Mandatory: tmuxConfig.Modifiers},
			},
			To: []To{
				{ShellCommand: fmt.Sprintf("%s %s", baseCmd, digit)},
			},
			Description: fmt.Sprintf("%s+%s → tmux session %s", modStr, digit, digit),
		})
	}

	// Letters jump to tmux sessions
	for _, letter := range tmuxConfig.Letters {
		manipulators = append(manipulators, Manipulator{
			Type: "basic",
			From: From{
				KeyCode:   letter,
				Modifiers: &Modifiers{Mandatory: tmuxConfig.Modifiers},
			},
			To: []To{
				{ShellCommand: fmt.Sprintf("%s %s", baseCmd, letter)},
			},
			Description: fmt.Sprintf("%s+%s → tmux session %s", modStr, letter, letter),
		})
	}

	return Rule{
		Description:  fmt.Sprintf("%s+Key → tmux session jump (%s)", modStr, tmuxConfig.Terminal),
		Manipulators: manipulators,
	}, nil
}

func createLayerRules(layers []LayerConfig) []Rule {
	rules := []Rule{}
	allLayerKeys := make([]string, len(layers))
	for i, layer := range layers {
		allLayerKeys[i] = layer.Key
	}

	for _, layer := range layers {
		key := layer.Key
		subBindings := layer.Sub
		layerType := layer.Type

		// Build conditions for other layers being off
		otherLayerConditions := []Condition{}
		for _, k := range allLayerKeys {
			if k != key {
				otherLayerConditions = append(otherLayerConditions, Condition{
					Type:  "variable_if",
					Name:  fmt.Sprintf("hyper_sublayer_%s", k),
					Value: 0,
				})
			}
		}

		// Build all conditions (hyper + other layers off)
		toggleConditions := append(
			[]Condition{{Type: "variable_if", Name: "hyper", Value: 1}},
			otherLayerConditions...,
		)

		// Toggle manipulator
		toggleManipulator := Manipulator{
			Type:        "basic",
			Description: fmt.Sprintf("Toggle Hyper sublayer %s", key),
			From: From{
				KeyCode:   key,
			},
			To: []To{
				{SetVariable: &SetVariable{
					Name:  fmt.Sprintf("hyper_sublayer_%s", key),
					Value: 1,
				}},
			},
			ToAfterKeyUp: []To{
				{SetVariable: &SetVariable{
					Name:  fmt.Sprintf("hyper_sublayer_%s", key),
					Value: 0,
				}},
			},
			Conditions: toggleConditions,
		}

		manipulators := []Manipulator{toggleManipulator}

		// Sub-key manipulators
		for subkey, val := range subBindings {
			var to To
			if layerType == "app" {
				to = To{
					SoftwareFunction: &SoftwareFunction{
						OpenApplication: &OpenApplication{
							FilePath: val,
						},
					},
				}
			} else if layerType == "web" {
				to = To{
					ShellCommand: fmt.Sprintf("open %s", val),
				}
			}

			manipulators = append(manipulators, Manipulator{
				Type:        "basic",
				Description: "Open ",
				From: From{
					KeyCode:   subkey,
				},
				To: []To{to},
				Conditions: []Condition{
					{
						Type:  "variable_if",
						Name:  fmt.Sprintf("hyper_sublayer_%s", key),
						Value: 1,
					},
				},
			})
		}

		rules = append(rules, Rule{
			Description:  fmt.Sprintf("Hyper Key sublayer \"%s\"", key),
			Manipulators: manipulators,
		})
	}

	return rules
}

func createSwitchTabsRule() Rule {
	return Rule{
		Description: "Remap ⌘+⌥+H/L to switch tabs",
		Manipulators: []Manipulator{
			{
				Type:        "basic",
				Description: "⌘+⌥+l → Next Tab (⌃+Tab)",
				From: From{
					KeyCode: "l",
					Modifiers: &Modifiers{
						Mandatory: []string{"command", "option"},
					},
				},
				To: []To{
					{KeyCode: "tab", Modifiers: []string{"control"}},
				},
			},
			{
				Type:        "basic",
				Description: "⌘+⌥+h → Previous Tab (⌃+⇧+Tab)",
				From: From{
					KeyCode: "h",
					Modifiers: &Modifiers{
						Mandatory: []string{"command", "option"},
					},
				},
				To: []To{
					{KeyCode: "tab", Modifiers: []string{"control", "shift"}},
				},
			},
		},
	}
}
