package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	outputPath string
	noBackup   bool
)

var generateCmd = &cobra.Command{
	Use:   "generate <config_path>",
	Short: "Generate Karabiner configuration from YAML",
	Long: `Generate karabiner.json from a simplified YAML configuration file.
By default, writes to ~/.config/karabiner/karabiner.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		configPath := args[0]
		return generateKarabinerConfig(configPath, outputPath, noBackup)
	},
}

func init() {
	generateCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Path to output karabiner.json file")
	generateCmd.Flags().BoolVar(&noBackup, "no-backup", false, "Skip creating backup of existing karabiner.json file")
}

func generateKarabinerConfig(configPath, outputPath string, noBackup bool) error {
	// Load and parse config
	config, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	// Determine output path
	var filePath string
	if outputPath != "" {
		filePath = outputPath
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		filePath = filepath.Join(home, ".config", "karabiner", "karabiner.json")
	}

	// Load existing karabiner config to preserve devices and other settings
	var existingKarabinerConfig KarabinerConfig
	if data, err := os.ReadFile(filePath); err == nil {
		json.Unmarshal(data, &existingKarabinerConfig)
	}

	// Create profile
	profile := Profile{
		Name:     "base",
		Selected: true,
		VirtualHIDKeyboard: &VirtualHIDKeyboard{
			KeyboardTypeV2: "iso",
		},
		SimpleModifications: []SimpleModification{},
		ComplexModifications: &ComplexModifications{
			Rules: []Rule{},
		},
	}

	// Preserve existing devices configuration if it exists
	for _, p := range existingKarabinerConfig.Profiles {
		if p.Name == "base" && p.Devices != nil {
			profile.Devices = p.Devices
			break
		}
	}

	// Handle fix_c_c (simple modification)
	if config.FixCC != nil && *config.FixCC {
		profile.SimpleModifications = append(profile.SimpleModifications, SimpleModification{
			From: KeyCode{KeyCode: "grave_accent_and_tilde"},
			To:   []KeyCode{{KeyCode: "non_us_backslash"}},
		})
	}

	// Generate complex modification rules
	rules := []Rule{}

	// Add HHKB mode if requested
	if config.UseHHKB {
		rules = append(rules, createHHKBModeRule())
		// If hyperkey is not caps_lock, add hyperkey rule
		if config.Hyperkey != "caps_lock" {
			rules = append(rules, createHyperKeyRule(config.Hyperkey))
		}
	} else {
		rules = append(rules, createHyperKeyRule(config.Hyperkey))
	}

	// Apply optional rules based on config
	optionalRules := []struct {
		enabled bool
		rule    func() Rule
	}{
		{config.DisableLeftCtrl, createDisableLeftCtrlRule},
		{config.DisableCommandTab, createDisableCommandTabRule},
		{config.SwitchSafariTabsHL, createSwitchTabsRule},
		{config.FixG502.Enable, func() Rule {
			return createFixG502Rule(
				config.FixG502.SafariOnly,
				config.FixG502.BackButton,
				config.FixG502.ForwardButton,
			)
		}},
	}

	for _, opt := range optionalRules {
		if opt.enabled {
			rules = append(rules, opt.rule())
		}
	}

	// Tmux jump
	if config.TmuxJump.Enable {
		tmuxRule, err := createTmuxJumpRule(config)
		if err != nil {
			return fmt.Errorf("failed to create tmux jump rule: %w", err)
		}
		rules = append(rules, tmuxRule)
	}

	// Option keybindings
	for key, binding := range config.Keybindings.Option {
		rules = append(rules, createOptionKeybindingRule(key, binding))
	}

	// HJKL arrow keys
	rules = append(rules, createHJKLRule())

	// Layer rules
	rules = append(rules, createLayerRules(config.Keybindings.Layers)...)

	// Set rules in profile
	profile.ComplexModifications.Rules = rules

	// Create final Karabiner config
	karabinerConfig := KarabinerConfig{
		Global: Global{
			ShowProfileNameInMenuBar: true,
		},
		Profiles: []Profile{profile},
	}

	// Preserve existing global settings if they exist
	if existingKarabinerConfig.Global.ShowProfileNameInMenuBar {
		karabinerConfig.Global = existingKarabinerConfig.Global
	}

	// Ensure output directory exists
	if err = os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create backup if file exists and backup is not disabled
	if !noBackup {
		if _, err := os.Stat(filePath); err == nil {
			timestamp := time.Now().Format("20060102_150405")
			backupName := fmt.Sprintf("backup_%s.json", timestamp)
			backupPath := filepath.Join(filepath.Dir(filePath), backupName)
			if err := copyFile(filePath, backupPath); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to create backup: %v\n", err)
			} else {
				fmt.Printf("Backup created: %s\n", backupPath)
			}
		}
	}

	// Write output file
	data, err := json.MarshalIndent(karabinerConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	fmt.Printf("Configuration written to: %s\n", filePath)
	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
