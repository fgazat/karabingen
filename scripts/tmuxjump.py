#!/usr/bin/env python3
"""
tmuxjump - Jump to tmux sessions via Karabiner-Elements
Replaces the shell script with a pure Python implementation.
"""
import os
import re
import subprocess
import sys

# Optional logging - only enable if TMUXJUMP_DEBUG=1 env var is set
_DEBUG = os.environ.get("TMUXJUMP_DEBUG") == "1"

if _DEBUG:
    import logging

    logging.basicConfig(
        level=logging.DEBUG,
        format="%(asctime)s - %(levelname)s - %(message)s",
        handlers=[logging.FileHandler(os.path.expanduser("~/tmuxjump.log")), logging.StreamHandler(sys.stderr)],
    )


def log_debug(msg):
    """Lightweight debug logging - only active if TMUXJUMP_DEBUG=1"""
    if _DEBUG:
        logging.debug(msg)


def error_exit(msg, code=1):
    """Print error message and exit with code"""
    print(msg, file=sys.stderr)
    sys.exit(code)


def run_command(cmd, capture=True, check=False):
    """Run shell command and optionally capture output"""
    try:
        if capture:
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True, check=check)
            return result.stdout.strip(), result.returncode
        else:
            result = subprocess.run(cmd, shell=True, check=check)
            return "", result.returncode
    except subprocess.CalledProcessError as e:
        return "", e.returncode


def run_osascript(script):
    """Run AppleScript and return output"""
    result = subprocess.run(["osascript", "-e", script], capture_output=True, text=True)
    return result.stdout.strip(), result.returncode


def _parse_key_mapping(keyname, rest, sessions, dirs, num_map):
    """Parse a line with key mapping format: key:name or key:name:dir"""
    # Skip if key already mapped
    if keyname in num_map:
        return

    # Check if rest contains directory: "name:dir"
    if ":" not in rest:
        sessions.append(rest)
        num_map[keyname] = rest
        log_debug(f"Parsed key mapping: key={keyname} name={rest}")
        return

    sname, srawdir = rest.split(":", 1)
    sdir = os.path.expanduser(srawdir)
    sessions.append(sname)
    dirs[sname] = sdir
    num_map[keyname] = sname
    log_debug(f"Parsed key mapping: key={keyname} name={sname} dir={sdir}")


def _parse_session_line(line, sessions, dirs):
    """Parse a line without key mapping: name or name:dir"""
    if ":" not in line:
        sessions.append(line)
        log_debug(f"Parsed session: name={line}")
        return

    sname, srawdir = line.split(":", 1)
    sdir = os.path.expanduser(srawdir)
    sessions.append(sname)
    dirs[sname] = sdir
    log_debug(f"Parsed session: name={sname} dir={sdir}")


def parse_tmuxjumplist_file(menu_file):
    """
    Parse ~/.tmuxjumplist file and return:
    - sessions: list of session names
    - dirs: dict mapping session name to directory
    - num_map: dict mapping key to session name
    """
    if not os.path.exists(menu_file):
        error_exit(f"Missing {menu_file}", 66)

    sessions = []
    dirs = {}
    num_map = {}

    with open(menu_file, "r") as f:
        for line in f:
            line = line.strip()

            # Skip empty lines and comments
            if not line or line.startswith("#"):
                continue

            # Try to match key mapping format: "key:rest"
            match = re.match(r"^([0-9a-zA-Z]+):(.+)$", line)
            if match:
                keyname = match.group(1)
                rest = match.group(2)
                _parse_key_mapping(keyname, rest, sessions, dirs, num_map)
            else:
                _parse_session_line(line, sessions, dirs)

    return sessions, dirs, num_map


def get_session_and_dir(index, sessions, dirs, num_map):
    """Resolve index to session name and directory"""
    # If NUM_MAP has this index, use it; otherwise fall back to array index
    if index in num_map:
        session = num_map[index]
        log_debug(f"Resolved index '{index}' via key mapping to session '{session}'")
    else:
        # Try to convert index to integer for array access
        try:
            idx = int(index)
            if idx >= len(sessions):
                error_exit(f"Index {idx} out of range (have {len(sessions)} sessions)", 65)
            session = sessions[idx]
            log_debug(f"Resolved index {idx} via array to session '{session}'")
        except (ValueError, IndexError):
            error_exit(f"Invalid index: {index}", 65)

    # Determine directory
    if session in dirs:
        directory = dirs[session]
        log_debug(f"Using configured directory for session '{session}': {directory}")
    elif os.path.isdir(os.path.expanduser(f"~/{session}")):
        directory = os.path.expanduser(f"~/{session}")
        log_debug(f"Using default directory for session '{session}': {directory}")
    else:
        directory = os.path.expanduser("~")
        log_debug(f"Using home directory for session '{session}'")

    return session, directory


def ensure_tmux_session(tmux_bin, session, directory):
    """Ensure tmux session exists (create if needed)"""
    _, retcode = run_command(f'"{tmux_bin}" has-session -t "{session}" 2>/dev/null')
    if retcode != 0:
        log_debug(f"Creating new tmux session '{session}' in directory {directory}")
        run_command(f'"{tmux_bin}" new-session -d -s "{session}" -c "{directory}"')
    else:
        log_debug(f"Tmux session '{session}' already exists")


def get_most_recent_client(tmux_bin):
    """Get the most recently used tmux client"""
    # Tmux uses #{variable} syntax for format strings
    output, _ = run_command(
        f"\"{tmux_bin}\" list-clients -F '#{{client_tty}} #{{client_activity}}' 2>/dev/null | sort -k2nr | awk 'NR==1{{print $1}}'"
    )
    return output


def switch_existing_client(tmux_bin, client, session, terminal_app):
    """Switch existing tmux client to session and focus terminal"""
    run_command(f'"{tmux_bin}" switch-client -c "{client}" -t "{session}"')
    run_command(f"/usr/bin/open -a {terminal_app}", check=False)


def count_terminal_windows(terminal_app):
    """Count terminal windows using AppleScript"""
    script = f"""
tell application "System Events"
  set isRunning to (exists process "{terminal_app}")
  if isRunning then
    try
      set winCount to count windows of process "{terminal_app}"
    on error
      set winCount to 0
    end try
  else
    set winCount to 0
  end if
end tell
return winCount
"""
    output, _ = run_osascript(script)
    try:
        return int(output)
    except ValueError:
        return 0


def type_into_terminal(terminal_app, session, tmux_bin):
    """Focus terminal and type tmux attach command"""
    script = f"""
tell application "{terminal_app}" to activate
delay 0.05
tell application "System Events"
  keystroke "{tmux_bin} attach -t {session}"
  key code 36
end tell
"""
    run_osascript(script)


def create_new_window(terminal_type, terminal_bin, tmux_bin, session):
    """Create new terminal window attached to session"""
    if terminal_type in ["alacritty", "ghostty"]:
        cmd = [terminal_bin, "-e", tmux_bin, "attach", "-t", session]
        os.execvp(terminal_bin, cmd)
    elif terminal_type == "iterm2":
        script = f"""
tell application "iTerm"
    create window with default profile command "{tmux_bin} attach -t {session}"
end tell
"""
        run_osascript(script)
    elif terminal_type == "terminal":
        script = f"""
tell application "Terminal"
    do script "{tmux_bin} attach -t {session}"
    activate
end tell
"""
        run_osascript(script)
    else:
        error_exit(f"Unknown terminal type: {terminal_type}", 1)


def main():
    log_debug(f"Starting tmuxjump with args: {sys.argv}")

    # Configuration
    TMUX_BIN = "/opt/homebrew/bin/tmux"  # Intel: /usr/local/bin/tmux

    # Terminal configuration defaults
    TERMINAL_CONFIGS = {
        "alacritty": {"app_name": "Alacritty", "bin": "/Applications/Alacritty.app/Contents/MacOS/alacritty"},
        "iterm2": {"app_name": "iTerm", "bin": None},  # Uses AppleScript only
        "terminal": {"app_name": "Terminal", "bin": None},  # Uses AppleScript only
        "ghostty": {"app_name": "Ghostty", "bin": "/Applications/Ghostty.app/Contents/MacOS/ghostty"},
    }

    # Parse arguments
    if len(sys.argv) < 2:
        error_exit(f"Usage: {sys.argv[0]} <key> [jumplist_path] [terminal_type]", 64)

    key_input = sys.argv[1]

    # Allow jumplist path to be specified as second argument or use default
    if len(sys.argv) >= 3:
        MENU_FILE = os.path.expanduser(sys.argv[2])
    else:
        MENU_FILE = os.path.expanduser("~/.tmuxjumplist")

    # Allow terminal type to be specified as third argument or use default
    if len(sys.argv) >= 4:
        terminal_type = sys.argv[3]
    else:
        terminal_type = "alacritty"

    if terminal_type not in TERMINAL_CONFIGS:
        error_exit(f"Unknown terminal type: {terminal_type}. Supported: {list(TERMINAL_CONFIGS.keys())}", 64)

    terminal_config = TERMINAL_CONFIGS[terminal_type]
    log_debug(f"Using terminal: {terminal_type} ({terminal_config['app_name']})")
    log_debug(f"Using jumplist file: {MENU_FILE}")

    # Parse menu file
    sessions, dirs, num_map = parse_tmuxjumplist_file(MENU_FILE)
    log_debug(f"Loaded {len(sessions)} sessions, {len(num_map)} key mappings")

    # Resolve session and directory
    session, directory = get_session_and_dir(key_input, sessions, dirs, num_map)

    # Ensure session exists
    ensure_tmux_session(TMUX_BIN, session, directory)

    # Try to switch existing client
    most_recent_client = get_most_recent_client(TMUX_BIN)
    if most_recent_client:
        log_debug(f"Found existing tmux client: {most_recent_client}")
        log_debug(f"Switching to session '{session}' and focusing {terminal_config['app_name']}")
        switch_existing_client(TMUX_BIN, most_recent_client, session, terminal_config["app_name"])
        sys.exit(0)

    log_debug("No tmux clients found")

    # No tmux clients. Check if terminal has windows
    window_count = count_terminal_windows(terminal_config["app_name"])
    log_debug(f"{terminal_config['app_name']} window count: {window_count}")
    if window_count > 0:
        log_debug(f"Typing attach command into existing {terminal_config['app_name']} window")
        type_into_terminal(terminal_config["app_name"], session, TMUX_BIN)
        sys.exit(0)

    # Last resort: create new window
    log_debug(f"Creating new {terminal_config['app_name']} window attached to session '{session}'")
    create_new_window(terminal_type, terminal_config["bin"], TMUX_BIN, session)


if __name__ == "__main__":
    main()
