# chronos

https://github.com/user-attachments/assets/1cfeaea3-5a87-4c07-b127-910b0da2808e

A terminal calendar application with vim-style keybindings.

Still highly experimental! Use at your own risk... I highly recommend using the
file config option rather than caldav until I get some more testing done!

## Features

- **Experimental sync with calendar applications**: based on iCal
- **Week view** with configurable zoom levels (30min, 15min, 5min, 1min)
- **Vim-style navigation and operations** - hjkl movement, visual mode, search
- **Event management** - create, edit, delete, copy/paste events
- **Recurring events** - create and manage repeating events
- **Multiple calendars** - organize events across different calendars
- **Search** - find events quickly with search mode
- **Agenda view** - see upcoming events at a glance
- **Undo/redo** - revert changes with u/U

## Installation & Usage

Build from source:

```bash
git clone https://github.com/samuelstranges/chronos
cd chronos
go build .

# Create calendar folder
mkdir -p ~/.config/chronos/calendars/

# Copy your calendar
cp /path/to/your/calendar.ics ~/.config/chronos/calendars/

# Run application
./chronos
```

### Key bindings

#### Keyhints

- `Space` - open keyhints

#### Movement

- `hjkl` or arrow keys - navigate calendar
- `H`/`L` - previous/next week
- `w`/`b`/`e` - next/previous/end event
- `t` - jump to current time
- `gg` - start of day
- `G` - end of day
- `^`/`$` - start of first/end of last event of day

#### Navigation

- `gh##` - go to hour (e.g., `gh14` for 2pm)
- `gm##` - go to minute
- `gt####` - go to time (e.g., `gt1430` for 2:30pm)
- `gd##` - go to day
- `gM##` - go to month
- `gY####` - go to year

#### Event operations

- `a` - add event
- `cc` - edit event
- `y` - copy event
- `x` - cut event
- `p` - paste event
- `enter` - view event details
- `o` - open event URL

#### Event editing

- `cc` - change all details
- `ct` - change title
- `cs` - change start time
- `ce` - change end time
- `cd` - change duration
- `cD` - change description
- `cl` - change location
- `cL` - change link
- `cC` - change color

#### Form input (when editing fields)

- `ctrl+u` - clear field
- `ctrl+a` / `home` - move to start
- `ctrl+e` / `end` - move to end
- `ctrl+w` - delete word backward
- `ctrl+k` - delete after cursor
- `esc` - cancel

#### View modes

- `v` - enter visual mode
- `/` - search mode
- `A` - toggle agenda view
- `tab` - toggle all-day events
- `+`/`-` - zoom in/out
- `T` - toggle event text color

#### Other

- `u`/`U` - undo/redo
- `C` - calendar manager
- `q` or `ctrl+c` - quit

## Configuration

### Storage Location

By default, Chronos stores calendars in `~/.config/chronos/calendars/` as
standard .ics files.

### Configuration File

Chronos can be configured via `~/.config/chronos/config.json`. If no config file
exists, it defaults to local file storage.

#### Storage Types

Chronos supports two storage backends:

1. **`file`** (default) - Local .ics files in `~/.config/chronos/calendars/`
2. **`caldav`** - Sync with CalDAV server (Nextcloud, Baikal, iCloud, etc.).
   This is still experimental and can break stuff... I've been running it with a
   though Baikal server which has been alot of fun though...

#### CalDAV Configuration

To sync with a CalDAV server, create `~/.config/chronos/config.json`:

```json
{
    "storage_type": "caldav",
    "caldav": {
        "server_url": "http://localhost:8086",
        "username": "sam",
        "password_command": "echo 'my_password'",
        "calendar_home_url": "/dav.php/calendars/sam/",
        "sync_interval": 60
    }
}
```

##### Password Options (choose one)

For security, use `password_command` (recommended) to fetch password from your
password manager, or use an environment variable, or plain text (not
recommended). Only one is needed to work:

```json
"password_command": "pass show chronos/caldav"
"password_env": "CHRONOS_CALDAV_PASSWORD"
"password": "your-password"
```

## Requirements

- Go 1.23.3+

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - styling
- [Huh](https://github.com/charmbracelet/huh) - forms
- [go-ical](https://github.com/emersion/go-ical) - iCalendar parser
- [rrule-go](https://github.com/teambition/rrule-go) - recurring rules
