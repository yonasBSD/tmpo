# Usage Guide

Complete reference for all tmpo commands and features.

## Basic Commands

### `tmpo start [description]`

Start tracking time for the current project. Automatically detects the project name from:

1. `.tmporc` configuration file (if present)
2. Git repository name
3. Current directory name

**Examples:**

```bash
tmpo start                             # Start tracking
tmpo start "Fix authentication bug"    # Start with description
```

### `tmpo stop`

Stop the currently running time entry.

```bash
tmpo stop
```

### `tmpo status`

View the current tracking session with elapsed time.

```bash
tmpo status
# Output:
# [tmpo] Currently tracking: my-project
#     Started: 2:30 PM
#     Duration: 1h 23m
#     Description: Implementing feature
```

### `tmpo log`

View your time tracking history.

**Options:**

- `--limit N` - Show N most recent entries (default: 20)

**Examples:**

```bash
tmpo log             # Show recent entries
tmpo log --limit 50  # Show more entries
```

### `tmpo stats`

Display statistics about your tracked time.

**Options:**

- `--today` - Show only today's statistics
- `--week` - Show this week's statistics
- `--month` - Show this month's statistics

**Examples:**

```bash
tmpo stats          # All-time stats
tmpo stats --today  # Today's stats
tmpo stats --week   # This week's stats
```

## Project Configuration

### `tmpo init`

Create a `.tmporc` configuration file for the current project.

**Options:**

- `--name "Project Name"` - Specify custom project name
- `--rate 150` - Set hourly rate for billing calculations

**Examples:**

```bash
tmpo init                                   # Auto-detect project name
tmpo init --name "My Project"               # Specify name
tmpo init --name "Client Work" --rate 150   # Set hourly rate
```

See [Configuration Guide](configuration.md) for details on the `.tmporc` file format.

## Advanced Features

### `tmpo manual`

Create manual time entries for past work using an interactive prompt.

```bash
tmpo manual
# Prompts for:
# - Project name
# - Start date and time
# - End date and time
# - Description
```

This is useful for:

- Recording time before you started using tmpo
- Adding entries when you forgot to start the timer
- Correcting tracking mistakes

### `tmpo export`

Export your time tracking data to CSV or JSON.

**Options:**

- `--format [csv|json]` - Output format (default: csv)
- `--project "Name"` - Filter by specific project
- `--today` - Export only today's entries
- `--week` - Export this week's entries
- `--month` - Export this month's entries
- `--output filename` - Specify output file path

**Examples:**

```bash
tmpo export                              # Export all as CSV
tmpo export --format json                # Export as JSON
tmpo export --project "My Project"       # Filter by project
tmpo export --today                      # Export today's entries
tmpo export --week                       # Export this week
tmpo export --output timesheet.csv       # Specify output file
```

**CSV Format:**

```csv
Project,Description,Start,End,Duration (hours)
my-project,Implementing feature,2024-01-15 14:30:00,2024-01-15 16:45:00,2.25
```

**JSON Format:**

```json
[
  {
    "project": "my-project",
    "description": "Implementing feature",
    "start": "2024-01-15T14:30:00Z",
    "end": "2024-01-15T16:45:00Z",
    "duration_hours": 2.25
  }
]
```

## Tips and Workflows

### Quick Daily Review

```bash
# See what you worked on today
tmpo stats --today
tmpo log --limit 10
```

### Weekly Timesheet Export

```bash
# Export this week's entries for invoicing
tmpo export --week --output timesheet-$(date +%Y-%m-%d).csv
```

### Multi-Project Workflow

Create a `.tmporc` file in each project directory with different hourly rates:

```bash
cd ~/projects/client-a
tmpo init --name "Client A" --rate 150

cd ~/projects/client-b
tmpo init --name "Client B" --rate 175
```

Now `tmpo start` will automatically track to the correct project when you're in each directory.

### Tracking Without Descriptions

You can always start tracking immediately and add context later by checking your git commits:

```bash
tmpo start
# ... do work ...
tmpo stop

# Later, correlate with git log to recall what you did
git log --since="2 hours ago" --oneline
```
