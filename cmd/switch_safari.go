package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var fzfPath string

var switchSafariCmd = &cobra.Command{
	Use:   "switch",
	Short: "Switch between Safari tabs using fzf",
	Long: `Opens an interactive fzf menu to search and switch between all open Safari tabs.
Requires fzf to be installed (brew install fzf).`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return switchSafariTab(fzfPath)
	},
}

func init() {
	switchSafariCmd.Flags().StringVar(&fzfPath, "fzf", "/opt/homebrew/bin/fzf", "Path to fzf binary")
}

func switchSafariTab(fzfPath string) error {
	// Use ASCII Unit Separator (0x1F) as delimiter - rarely appears in text
	const delimiter = "\x1F"

	// AppleScript to list all Safari tabs
	listTabsScript := `
tell application "Safari"
	set output to ""
	repeat with w from 1 to count windows
		repeat with t from 1 to count tabs of window w
			set tabURL to URL of tab t of window w
			set tabName to name of tab t of window w

			-- Replace pipe characters to avoid breaking delimiter
			set AppleScript's text item delimiters to "|"
			set tabNameParts to text items of tabName
			set AppleScript's text item delimiters to "¦"
			set tabName to tabNameParts as string

			set urlParts to text items of tabURL
			set tabURL to urlParts as string
			set AppleScript's text item delimiters to ""

			-- Truncate or pad to exactly 70 characters
			if length of tabName > 70 then
				set tabName to text 1 thru 67 of tabName & "..."
			else
				set padding to ""
				repeat (70 - (length of tabName)) times
					set padding to padding & " "
				end repeat
				set tabName to tabName & padding
			end if

			set output to output & w & (ASCII character 31) & t & (ASCII character 31) & tabName & (ASCII character 31) & tabURL & linefeed
		end repeat
	end repeat
	return output
end tell
`

	// Get list of tabs from Safari
	osascriptCmd := exec.Command("osascript", "-e", listTabsScript)
	tabsOutput, err := osascriptCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get Safari tabs (is Safari running?): %w", err)
	}

	// Pipe to fzf for selection
	fzfCmd := exec.Command(fzfPath, "--delimiter="+delimiter, "--with-nth=3,4")
	fzfCmd.Stdin = strings.NewReader(string(tabsOutput))
	fzfOutput, err := fzfCmd.Output()
	if err != nil {
		// User probably cancelled (Ctrl+C or ESC)
		return nil
	}

	selection := strings.TrimSpace(string(fzfOutput))
	if selection == "" {
		return nil
	}

	// Parse selection: window<delim>tab<delim>name<delim>url
	parts := strings.Split(selection, delimiter)
	if len(parts) < 2 {
		return fmt.Errorf("invalid selection format")
	}

	window := parts[0]
	tab := parts[1]

	// Switch to selected tab
	switchScript := fmt.Sprintf(`tell application "Safari" to tell window %s to set current tab to tab %s`, window, tab)
	if err := exec.Command("osascript", "-e", switchScript).Run(); err != nil {
		return fmt.Errorf("failed to switch tab: %w", err)
	}

	// Activate Safari
	activateScript := `tell application "Safari" to activate`
	if err := exec.Command("osascript", "-e", activateScript).Run(); err != nil {
		return fmt.Errorf("failed to activate Safari: %w", err)
	}

	return nil
}
