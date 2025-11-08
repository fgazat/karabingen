#!/bin/zsh
set -euo pipefail

TMUX_BIN="/opt/homebrew/bin/tmux"          # Intel: /usr/local/bin/tmux
ALACRITTY_BIN="/Applications/Alacritty.app/Contents/MacOS/alacritty"
MENU_FILE="$HOME/.tmf"

KEY_INPUT="${1:-}"
if [[ -z "$KEY_INPUT" ]]; then
  print -u2 "Usage: $0 <key>"
  exit 64
fi
INDEX="$KEY_INPUT"

# --- load ~/.tmf (supports "name", "name:/path", or "num:name:/path")
typeset -a SESSIONS; SESSIONS=()
typeset -A DIRS
typeset -A NUM_MAP
if [[ ! -f "$MENU_FILE" ]]; then
  print -u2 "Missing $MENU_FILE"
  exit 66
fi
while IFS= read -r line; do
  line="${${line##[[:space:]]#}%[[:space:]]#}"
  [[ -z "$line" || "$line" == '#'* ]] && continue

  # Check if line starts with key (letter or number):
  if [[ "$line" =~ ^([0-9a-zA-Z]+):(.+)$ ]]; then
    local keyname="${match[1]}"
    local rest="${match[2]}"

    # Only use first occurrence of each key
    if [[ -z "${NUM_MAP[$keyname]-}" ]]; then
      if [[ "$rest" == *:* ]]; then
        local sname="${rest%%:*}"
        local srawdir="${rest#*:}"
        local sdir="${~srawdir}"
        SESSIONS+=("$sname")
        DIRS[$sname]="$sdir"
        NUM_MAP[$keyname]="$sname"
        echo "DEBUG PARSE: key=$keyname name=$sname dir=$sdir" >&2
      else
        SESSIONS+=("$rest")
        NUM_MAP[$keyname]="$rest"
        echo "DEBUG PARSE: key=$keyname name=$rest" >&2
      fi
    fi
  elif [[ "$line" == *:* ]]; then
    local sname="${line%%:*}"
    local srawdir="${line#*:}"
    local sdir="${~srawdir}"
    SESSIONS+=("$sname")
    DIRS[$sname]="$sdir"
    echo "DEBUG PARSE: name=$sname dir=$sdir" >&2
  else
    SESSIONS+=("$line")
    echo "DEBUG PARSE: name=$line" >&2
  fi
done < "$MENU_FILE"

if (( INDEX >= ${#SESSIONS[@]} )); then
  print -u2 "Index $INDEX out of range (have ${#SESSIONS[@]} sessions)"
  exit 65
fi

# If NUM_MAP has this index, use it; otherwise fall back to array index
if [[ -n "${NUM_MAP[$INDEX]-}" ]]; then
  SESSION="${NUM_MAP[$INDEX]}"
else
  SESSION="${SESSIONS[$((INDEX+1))]}"   # zsh arrays are 1-based
fi

echo "DEBUG: INDEX=$INDEX SESSION=$SESSION" >&2
echo "DEBUG: DIRS[$SESSION]=${DIRS[$SESSION]-EMPTY}" >&2
if [[ -n "${DIRS[$SESSION]-}" ]]; then
  DIR="${DIRS[$SESSION]}"
elif [[ -d "$HOME/$SESSION" ]]; then
  DIR="$HOME/$SESSION"
else
  DIR="$HOME"
fi

# Debug output
echo "DEBUG: SESSION=$SESSION" >&2
echo "DEBUG: DIR=$DIR" >&2

# 1) Ensure the session exists (detached)
if ! "$TMUX_BIN" has-session -t "$SESSION" 2>/dev/null; then
  "$TMUX_BIN" new-session -d -s "$SESSION" -c "$DIR"
fi

# 2) If there is any tmux client, switch it and focus Alacritty (no new window)
MOST_RECENT_CLIENT="$("$TMUX_BIN" list-clients -F '#{client_tty} #{client_activity}' 2>/dev/null \
  | sort -k2nr | awk 'NR==1{print $1}')"

if [[ -n "${MOST_RECENT_CLIENT:-}" ]]; then
  "$TMUX_BIN" switch-client -c "$MOST_RECENT_CLIENT" -t "$SESSION"
  /usr/bin/open -a Alacritty
  exit 0
fi

# 3) No tmux clients. If Alacritty has a window, type the tmux command into it (no new window).
HAS_ALA_WINDOW="$(/usr/bin/osascript <<'APPLESCRIPT'
tell application "System Events"
  set isRunning to (exists process "Alacritty")
  if isRunning then
    try
      set winCount to count windows of process "Alacritty"
    on error
      set winCount to 0
    end try
  else
    set winCount to 0
  end if
end tell
return winCount
APPLESCRIPT
)"
if [[ "$HAS_ALA_WINDOW" != "0" ]]; then
  # Focus Alacritty and type the tmux attach/create command into the current shell
  # Requires Accessibility permission for Terminal/karabiner to control "System Events".
  /usr/bin/osascript <<APPLESCRIPT
tell application "Alacritty" to activate
delay 0.05
tell application "System Events"
  keystroke "tmux attach -t ${SESSION}"
  key code 36  -- Return
end tell
APPLESCRIPT
  exit 0
fi

# 4) Last resort: create one window attached to the session
exec "$ALACRITTY_BIN" -e /bin/zsh -lc "\"$TMUX_BIN\" attach -t '$SESSION'"
