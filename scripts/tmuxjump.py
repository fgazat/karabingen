#!/usr/bin/env python3
"""
tmuxjump - Jump to tmux sessions via Karabiner-Elements
Replaces the shell script with a pure Python implementation.
"""
import logging
import os
import re
import subprocess
import sys

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
    handlers=[logging.FileHandler(os.path.expanduser("~/tmuxjump.log")), logging.StreamHandler(sys.stderr)],
)


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
            # Strip whitespace
            line = line.strip()

            # Skip empty lines and comments
            if not line or line.startswith("#"):
                continue

            # Check if line starts with key (letter or number): "key:rest"
            match = re.match(r"^([0-9a-zA-Z]+):(.+)$", line)
            if match:
                keyname = match.group(1)
                rest = match.group(2)

                # Only use first occurrence of each key
                if keyname not in num_map:
                    # Check if rest contains directory: "name:dir"
                    if ":" in rest:
                        sname, srawdir = rest.split(":", 1)
                        # Expand tilde
                        sdir = os.path.expanduser(srawdir)
                        sessions.append(sname)
                        dirs[sname] = sdir
                        num_map[keyname] = sname
                        logging.debug(f"Parsed key mapping: key={keyname} name={sname} dir={sdir}")
                    else:
                        sessions.append(rest)
                        num_map[keyname] = rest
                        logging.debug(f"Parsed key mapping: key={keyname} name={rest}")

            # No key prefix, check if it's "name:dir"
            elif ":" in line:
                sname, srawdir = line.split(":", 1)
                sdir = os.path.expanduser(srawdir)
                sessions.append(sname)
                dirs[sname] = sdir
                logging.debug(f"Parsed session: name={sname} dir={sdir}")

            # Just a name
            else:
                sessions.append(line)
                logging.debug(f"Parsed session: name={line}")

    return sessions, dirs, num_map


def get_session_and_dir(index, sessions, dirs, num_map):
    """Resolve index to session name and directory"""
    # If NUM_MAP has this index, use it; otherwise fall back to array index
    if index in num_map:
        session = num_map[index]
        logging.debug(f"Resolved index '{index}' via key mapping to session '{session}'")
    else:
        # Try to convert index to integer for array access
        try:
            idx = int(index)
            if idx >= len(sessions):
                error_exit(f"Index {idx} out of range (have {len(sessions)} sessions)", 65)
            session = sessions[idx]
            logging.debug(f"Resolved index {idx} via array to session '{session}'")
        except (ValueError, IndexError):
            error_exit(f"Invalid index: {index}", 65)

    # Determine directory
    if session in dirs:
        directory = dirs[session]
        logging.debug(f"Using configured directory for session '{session}': {directory}")
    elif os.path.isdir(os.path.expanduser(f"~/{session}")):
        directory = os.path.expanduser(f"~/{session}")
        logging.debug(f"Using default directory for session '{session}': {directory}")
    else:
        directory = os.path.expanduser("~")
        logging.debug(f"Using home directory for session '{session}'")

    return session, directory


def ensure_tmux_session(tmux_bin, session, directory):
    """Ensure tmux session exists (create if needed)"""
    _, retcode = run_command(f'"{tmux_bin}" has-session -t "{session}" 2>/dev/null')
    if retcode != 0:
        logging.debug(f"Creating new tmux session '{session}' in directory {directory}")
        run_command(f'"{tmux_bin}" new-session -d -s "{session}" -c "{directory}"')
    else:
        logging.debug(f"Tmux session '{session}' already exists")


def get_most_recent_client(tmux_bin):
    """Get the most recently used tmux client"""
    # Tmux uses #{variable} syntax for format strings
    output, _ = run_command(
        f"\"{tmux_bin}\" list-clients -F '#{{client_tty}} #{{client_activity}}' 2>/dev/null | sort -k2nr | awk 'NR==1{{print $1}}'"
    )
    return output


def switch_existing_client(tmux_bin, client, session):
    """Switch existing tmux client to session and focus Alacritty"""
    run_command(f'"{tmux_bin}" switch-client -c "{client}" -t "{session}"')
    run_command("/usr/bin/open -a Alacritty", check=False)


def count_alacritty_windows():
    """Count Alacritty windows using AppleScript"""
    script = """
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
"""
    output, _ = run_osascript(script)
    try:
        return int(output)
    except ValueError:
        return 0


def type_into_alacritty(session):
    """Focus Alacritty and type tmux attach command"""
    script = f"""
tell application "Alacritty" to activate
delay 0.05
tell application "System Events"
  keystroke "tmux attach -t {session}"
  key code 36
end tell
"""
    run_osascript(script)


def create_new_window(alacritty_bin, tmux_bin, session):
    """Create new Alacritty window attached to session"""
    cmd = [alacritty_bin, "-e", tmux_bin, "attach", "-t", session]
    os.execvp(alacritty_bin, cmd)


def main():
    logging.debug(f"Starting tmuxjump with args: {sys.argv}")

    # Configuration
    TMUX_BIN = "/opt/homebrew/bin/tmux"  # Intel: /usr/local/bin/tmux
    ALACRITTY_BIN = "/Applications/Alacritty.app/Contents/MacOS/alacritty"

    # Parse arguments
    if len(sys.argv) < 2:
        error_exit(f"Usage: {sys.argv[0]} <key> [jumplist_path]", 64)

    key_input = sys.argv[1]

    # Allow jumplist path to be specified as second argument or use default
    if len(sys.argv) >= 3:
        MENU_FILE = os.path.expanduser(sys.argv[2])
    else:
        MENU_FILE = os.path.expanduser("~/.tmuxjumplist")

    logging.debug(f"Using jumplist file: {MENU_FILE}")

    # Parse menu file
    sessions, dirs, num_map = parse_tmuxjumplist_file(MENU_FILE)
    logging.debug(f"Loaded {len(sessions)} sessions, {len(num_map)} key mappings")

    # Resolve session and directory
    session, directory = get_session_and_dir(key_input, sessions, dirs, num_map)

    # Ensure session exists
    ensure_tmux_session(TMUX_BIN, session, directory)

    # Try to switch existing client
    most_recent_client = get_most_recent_client(TMUX_BIN)
    if most_recent_client:
        logging.debug(f"Found existing tmux client: {most_recent_client}")
        logging.debug(f"Switching to session '{session}' and focusing Alacritty")
        switch_existing_client(TMUX_BIN, most_recent_client, session)
        sys.exit(0)

    logging.debug("No tmux clients found")

    # No tmux clients. Check if Alacritty has windows
    window_count = count_alacritty_windows()
    logging.debug(f"Alacritty window count: {window_count}")
    if window_count > 0:
        logging.debug(f"Typing attach command into existing Alacritty window")
        type_into_alacritty(session)
        sys.exit(0)

    # Last resort: create new window
    logging.debug(f"Creating new Alacritty window attached to session '{session}'")
    create_new_window(ALACRITTY_BIN, TMUX_BIN, session)


if __name__ == "__main__":
    main()
