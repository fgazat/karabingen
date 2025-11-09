#!/usr/bin/env bash
# tmb - TmuX Bookmark: Add current directory to tmux jump list
# Usage: tmb [jumplist_file]

# Parse arguments - allow jumplist path to be specified or use default
if [ $# -ge 1 ]; then
    BOOKMARK_FILE="$1"
else
    BOOKMARK_FILE="$HOME/.tmuxjumplist"
fi

# Get current directory name
name="$(basename "$PWD")"

# Extract already used keys from bookmark file
used_keys=$(grep -oE '^[0-9a-zA-Z]+:' "$BOOKMARK_FILE" 2>/dev/null | sed 's/:$//' | sort -u | tr '\n' ' ')

if [ -n "$used_keys" ]; then
    echo "Already used keys: $used_keys"
fi

echo -n "Enter key for '$name': "
read key

if [ -z "$key" ]; then
    echo "No key provided, aborting."
    exit 1
fi

# Check if key already exists
if grep -qE "^${key}:" "$BOOKMARK_FILE" 2>/dev/null; then
    echo "Warning: key '$key' already exists in $BOOKMARK_FILE"
    echo -n "Overwrite? (y/n/a - y:replace, n:cancel, a:append): "
    read confirm
    case "$confirm" in
        y|Y)
            # Remove existing entry with this key
            sed -i.bak "/^${key}:/d" "$BOOKMARK_FILE"
            ;;
        a|A)
            # Just append, don't remove existing
            ;;
        *)
            echo "Cancelled."
            exit 1
            ;;
    esac
fi

echo "${key}:${name}:${PWD}" >> "$BOOKMARK_FILE"
echo "Added: ${key}:${name}:${PWD}"
