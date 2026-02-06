package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

type bookmark struct {
	Key  string
	Name string
	Path string
}

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
		bookmarkFile = expandTilde(bookmarkFile)

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

	// Read existing bookmarks
	bookmarks, err := getBookmarks(bookmarkFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read bookmark file: %w", err)
	}

	// Display existing bookmarks if any
	if len(bookmarks) > 0 {
		// Sort bookmarks: numbers first, then letters
		slices.SortFunc(bookmarks, func(a, b bookmark) int {
			ka, kb := a.Key, b.Key
			isNumA := ka >= "0" && ka <= "9"
			isNumB := kb >= "0" && kb <= "9"

			if isNumA && !isNumB {
				return -1
			}
			if !isNumA && isNumB {
				return 1
			}
			if ka < kb {
				return -1
			}
			if ka > kb {
				return 1
			}
			return 0
		})

		// Find max name length for alignment
		maxNameLen := 0
		for _, b := range bookmarks {
			if len(b.Name) > maxNameLen {
				maxNameLen = len(b.Name)
			}
		}

		fmt.Println("Existing bookmarks:")
		for _, b := range bookmarks {
			fmt.Printf("  %s:\t%-*s\t%s\n", b.Key, maxNameLen, b.Name, b.Path)
		}
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
	existingKeys := make([]string, len(bookmarks))
	for i, b := range bookmarks {
		existingKeys[i] = b.Key
	}
	if slices.Contains(existingKeys, key) {
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
	f, err := os.OpenFile(bookmarkFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
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

func getBookmarks(bookmarkFile string) ([]bookmark, error) {
	file, err := os.Open(bookmarkFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bookmarkRegex := regexp.MustCompile(`^([0-9a-zA-Z]+):([^:]+):(.+)$`)
	var bookmarks []bookmark

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if matches := bookmarkRegex.FindStringSubmatch(line); len(matches) == 4 {
			bookmarks = append(bookmarks, bookmark{
				Key:  matches[1],
				Name: matches[2],
				Path: matches[3],
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return bookmarks, nil
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
	return os.WriteFile(bookmarkFile, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
