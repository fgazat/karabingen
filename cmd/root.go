package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "karabingen",
	Short: "CLI tool to generate Karabiner configuration",
	Long:  `karabingen is a CLI tool to generate karabiner.json from simplified YAML configuration.`,
}

var tmuxCmd = &cobra.Command{
	Use:   "tmux",
	Short: "Tmux session management commands",
	Long:  `Commands for managing tmux sessions and bookmarks.`,
}

var safariCmd = &cobra.Command{
	Use:   "safari",
	Short: "Safari browser utilities",
	Long:  `Commands for managing Safari tabs and windows.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add generate command directly to root
	rootCmd.AddCommand(generateCmd)

	// Add tmux parent command
	rootCmd.AddCommand(tmuxCmd)

	// Add tmux subcommands
	tmuxCmd.AddCommand(switchTmuxCmd)
	tmuxCmd.AddCommand(bookmarkTmuxCmd)

	// Add safari parent command
	rootCmd.AddCommand(safariCmd)

	// Add safari subcommands
	safariCmd.AddCommand(switchSafariCmd)
}
