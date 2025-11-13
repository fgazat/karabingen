package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	tmuxPath     string
	jumplistPath string
	terminal     string
)

var switchTmuxCmd = &cobra.Command{
	Use:   "switch <key>",
	Short: "Switch to a tmux session based on jumplist",
	Long: `Switch to a tmux session by reading the jumplist file and jumping to the session
corresponding to the provided key (0-9, a-z).`,
	Args:         cobra.ExactArgs(1),
	SilenceUsage: true, // Don't show usage on errors
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if err := switchTmuxSession(key, tmuxPath, jumplistPath, terminal); err != nil {
			// Log error to a file for debugging instead of stdout
			logError(err)
			return nil // Return nil to avoid showing usage and exit code 1
		}
		return nil
	},
}

func init() {
	switchTmuxCmd.Flags().StringVar(&tmuxPath, "tmux", "/opt/homebrew/bin/tmux", "Path to tmux binary")
	switchTmuxCmd.Flags().StringVar(&jumplistPath, "jumplist", "~/.tmuxjumplist", "Path to jumplist file")
	switchTmuxCmd.Flags().StringVar(&terminal, "terminal", "alacritty", "Terminal to use (alacritty, iterm2, terminal, ghostty)")
	switchTmuxCmd.MarkFlagRequired("jumplist")
}

func switchTmuxSession(key, tmuxPath, jumplistPath, terminal string) error {
	// Special case: 0 opens the jumplist file for editing
	if key == "0" {
		return editJumplist(jumplistPath, terminal)
	}

	// Expand home directory in jumplist path
	if strings.HasPrefix(jumplistPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		jumplistPath = filepath.Join(home, jumplistPath[2:])
	}

	// Read jumplist file
	sessions, err := readJumplist(jumplistPath)
	if err != nil {
		return fmt.Errorf("failed to read jumplist %s: %w", jumplistPath, err)
	}

	// Find session for the given key
	// Format: key:session_name or key:session_name:directory
	var sessionName, directory string
	for _, line := range sessions {
		parts := strings.Split(line, ":")
		if len(parts) >= 2 && strings.TrimSpace(parts[0]) == key {
			// Session name is always the second part
			sessionName = strings.TrimSpace(parts[1])
			// Directory is optional third part
			if len(parts) >= 3 {
				directory = strings.TrimSpace(parts[2])
				if strings.HasPrefix(directory, "~/") {
					home, _ := os.UserHomeDir()
					directory = filepath.Join(home, directory[2:])
				}
			}
			break
		}
	}

	if sessionName == "" {
		return fmt.Errorf("no session found for key '%s' in jumplist %s", key, jumplistPath)
	}

	// Use home directory if no directory specified
	if directory == "" {
		directory, _ = os.UserHomeDir()
	}

	// Ensure tmux session exists (create if needed)
	ensureTmuxSession(tmuxPath, sessionName, directory)

	// Get terminal app name from terminal type
	terminalApp := getTerminalAppName(terminal)

	// Try to switch existing tmux client first
	mostRecentClient := getMostRecentTmuxClient(tmuxPath)
	if mostRecentClient != "" {
		// Switch existing client and focus terminal
		exec.Command(tmuxPath, "switch-client", "-c", mostRecentClient, "-t", sessionName).Run()
		exec.Command("open", "-a", terminalApp).Run()
		return nil
	}

	// No tmux clients found. Check if terminal has windows
	windowCount := countTerminalWindows(terminalApp)
	if windowCount > 0 {
		// Type attach command into existing terminal window
		typeIntoTerminal(terminalApp, sessionName, tmuxPath)
		return nil
	}

	// Last resort: create new window
	return createNewWindow(terminal, tmuxPath, sessionName)
}

func readJumplist(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}

	return lines, scanner.Err()
}

func editJumplist(jumplistPath, terminal string) error {
	// Expand home directory
	if strings.HasPrefix(jumplistPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		jumplistPath = filepath.Join(home, jumplistPath[2:])
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nvim"
	}

	// Check if we're inside tmux
	if os.Getenv("TMUX") != "" {
		// Inside tmux: open in new window
		cmd := exec.Command("/opt/homebrew/bin/tmux", "new-window", fmt.Sprintf("%s %s", editor, jumplistPath))
		return cmd.Run()
	}

	// Outside tmux: open in terminal
	switch terminal {
	case "iterm2":
		script := fmt.Sprintf(`tell application "iTerm" to create window with default profile command "%s %s"`, editor, jumplistPath)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	case "terminal":
		script := fmt.Sprintf(`tell application "Terminal" to do script "%s %s"`, editor, jumplistPath)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	case "alacritty":
		cmd := exec.Command("/Applications/Alacritty.app/Contents/MacOS/alacritty", "-e", editor, jumplistPath)
		return cmd.Run()
	case "ghostty":
		cmd := exec.Command("/Applications/Ghostty.app/Contents/MacOS/ghostty", "-e", editor, jumplistPath)
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported terminal: %s", terminal)
	}
}

func getTerminalAppName(terminal string) string {
	switch terminal {
	case "alacritty":
		return "Alacritty"
	case "iterm2":
		return "iTerm"
	case "terminal":
		return "Terminal"
	case "ghostty":
		return "Ghostty"
	default:
		return terminal
	}
}

func ensureTmuxSession(tmuxPath, sessionName, directory string) {
	// Check if session exists
	cmd := exec.Command(tmuxPath, "has-session", "-t", sessionName)
	if err := cmd.Run(); err != nil {
		// Session doesn't exist, create it
		exec.Command(tmuxPath, "new-session", "-d", "-s", sessionName, "-c", directory).Run()
	}
}

func getMostRecentTmuxClient(tmuxPath string) string {
	// Get most recently used tmux client
	cmd := exec.Command("sh", "-c",
		fmt.Sprintf(`"%s" list-clients -F '#{client_tty} #{client_activity}' 2>/dev/null | sort -k2nr | awk 'NR==1{print $1}'`, tmuxPath))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func countTerminalWindows(terminalApp string) int {
	script := fmt.Sprintf(`
tell application "System Events"
  set isRunning to (exists process "%s")
  if isRunning then
    try
      set winCount to count windows of process "%s"
    on error
      set winCount to 0
    end try
  else
    set winCount to 0
  end if
end tell
return winCount`, terminalApp, terminalApp)

	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	var count int
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count
}

func typeIntoTerminal(terminalApp, sessionName, tmuxPath string) error {
	script := fmt.Sprintf(`
tell application "%s" to activate
delay 0.005
tell application "System Events"
  keystroke "%s attach -t %s"
  key code 36
end tell`, terminalApp, tmuxPath, sessionName)

	cmd := exec.Command("osascript", "-e", script)
	return cmd.Run()
}

func createNewWindow(terminal, tmuxPath, sessionName string) error {
	switch terminal {
	case "iterm2":
		script := fmt.Sprintf(`tell application "iTerm" to create window with default profile command "%s attach -t %s"`, tmuxPath, sessionName)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	case "terminal":
		script := fmt.Sprintf(`
tell application "Terminal"
    do script "%s attach -t %s"
    activate
end tell`, tmuxPath, sessionName)
		cmd := exec.Command("osascript", "-e", script)
		return cmd.Run()
	case "alacritty":
		cmd := exec.Command("/Applications/Alacritty.app/Contents/MacOS/alacritty", "-e", tmuxPath, "attach", "-t", sessionName)
		return cmd.Start() // Use Start instead of Run to not block
	case "ghostty":
		cmd := exec.Command("/Applications/Ghostty.app/Contents/MacOS/ghostty", "-e", tmuxPath, "attach", "-t", sessionName)
		return cmd.Start()
	default:
		return fmt.Errorf("unsupported terminal: %s", terminal)
	}
}

func logError(err error) {
	// Log errors to a debug file for troubleshooting
	home, homeErr := os.UserHomeDir()
	if homeErr != nil {
		return
	}

	logPath := filepath.Join(home, ".config", "karabiner", "karabingen-errors.log")
	f, fileErr := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if fileErr != nil {
		return
	}
	defer f.Close()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(f, "[%s] %v\n", timestamp, err)
}
