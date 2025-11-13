package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var bookmarkTmuxCmd = &cobra.Command{
	Use:   "bookmark [jumplist_file]",
	Short: "Add current directory to tmux jump list",
	Long: `TmuX Bookmark: Add the current working directory to the tmux jump list.
If no jumplist file is specified, defaults to ~/.tmuxjumplist.

The bookmark format is: key:name:directory
Where:
  - key: A single character (0-9, a-z, A-Z) to trigger the session
  - name: The session name (defaults to directory basename)
  - directory: The full path to the directory`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine jumplist file path
		var bookmarkFile string
		if len(args) >= 1 {
			bookmarkFile = args[0]
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			bookmarkFile = filepath.Join(home, ".tmuxjumplist")
		}

		// Expand home directory in path
		if strings.HasPrefix(bookmarkFile, "~/") {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			bookmarkFile = filepath.Join(home, bookmarkFile[2:])
		}

		return addBookmark(bookmarkFile)
	},
}

func addBookmark(bookmarkFile string) error {
	// Get current directory
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Get directory name
	name := filepath.Base(pwd)

	// Read existing bookmarks to find used keys
	usedKeys, err := getUsedKeys(bookmarkFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read bookmark file: %w", err)
	}

	// Display used keys if any
	if len(usedKeys) > 0 {
		fmt.Printf("Already used keys: %s\n", strings.Join(usedKeys, " "))
	}

	// Prompt for key
	fmt.Printf("Enter key for '%s': ", name)
	reader := bufio.NewReader(os.Stdin)
	key, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	key = strings.TrimSpace(key)

	if key == "" {
		fmt.Println("No key provided, aborting.")
		return nil
	}

	// Check if key already exists
	if keyExists(usedKeys, key) {
		fmt.Printf("Warning: key '%s' already exists in %s\n", key, bookmarkFile)
		fmt.Print("Overwrite? (y/n/a - y:replace, n:cancel, a:append): ")
		confirm, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		confirm = strings.TrimSpace(strings.ToLower(confirm))

		switch confirm {
		case "y":
			// Remove existing entry with this key
			if err := removeKeyFromFile(bookmarkFile, key); err != nil {
				return fmt.Errorf("failed to remove existing key: %w", err)
			}
		case "a":
			// Just append, don't remove existing
		default:
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Append new bookmark
	entry := fmt.Sprintf("%s:%s:%s\n", key, name, pwd)
	f, err := os.OpenFile(bookmarkFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open bookmark file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		return fmt.Errorf("failed to write bookmark: %w", err)
	}

	fmt.Printf("Added: %s:%s:%s\n", key, name, pwd)
	return nil
}

func getUsedKeys(bookmarkFile string) ([]string, error) {
	file, err := os.Open(bookmarkFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	keyRegex := regexp.MustCompile(`^([0-9a-zA-Z]+):`)
	keysMap := make(map[string]bool)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := keyRegex.FindStringSubmatch(line); len(matches) > 1 {
			keysMap[matches[1]] = true
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	keys := make([]string, 0, len(keysMap))
	for key := range keysMap {
		keys = append(keys, key)
	}

	return keys, nil
}

func keyExists(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

func removeKeyFromFile(bookmarkFile, keyToRemove string) error {
	// Read all lines
	file, err := os.Open(bookmarkFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	keyRegex := regexp.MustCompile(`^` + regexp.QuoteMeta(keyToRemove) + `:`)

	for scanner.Scan() {
		line := scanner.Text()
		// Skip lines that start with the key to remove
		if !keyRegex.MatchString(line) {
			lines = append(lines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Write back to file
	return os.WriteFile(bookmarkFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
