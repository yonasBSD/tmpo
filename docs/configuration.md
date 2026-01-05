# Configuration Guide

Learn how to configure tmpo for your projects and workflow.

## Storage Location

All time tracking data and configuration is stored locally on your machine:

```text
~/.tmpo/
  ├── tmpo.db          # SQLite database with time entries
  └── config.yaml      # Global configuration (optional)
```

Your data never leaves your machine. Both files can be backed up, copied, or version controlled if desired.

> [!NOTE]
> **Contributors**, when developing tmpo with `TMPO_DEV=1` or `TMPO_DEV=true`, both files are stored in `~/.tmpo-dev/` instead to keep development work separate from your production data.

## Global Configuration

### The `tmpo config` Command

Use `tmpo config` to set user-wide preferences that apply across all projects:

```bash
tmpo config
```

This launches an interactive configuration wizard where you can set:

- **Currency** - Your preferred currency for displaying billing rates and earnings
- **Date Format** - Choose between MM/DD/YYYY, DD/MM/YYYY, or YYYY-MM-DD
- **Time Format** - Choose between 24-hour (15:30) or 12-hour (3:30 PM)
- **Timezone** - IANA timezone for your location (e.g., America/New_York, Europe/London)
- **Export Path** - Default directory for exported files (type "clear" to remove)

### Global Settings

Global preferences are stored in `~/.tmpo/config.yaml`:

```yaml
currency: USD
date_format: MM/DD/YYYY
time_format: 12-hour (AM/PM)
timezone: America/New_York
export_path: ~/Documents/timesheets
```

These settings affect how tmpo displays times and currencies throughout the application:

#### Currency

Your currency choice determines the symbol displayed for all billing information across all projects:

**Supported Currencies:**

tmpo supports 30+ currencies including:

- **Americas:** USD ($), CAD (CA$), BRL (R$), MXN (MX$)
- **Europe:** EUR (€), GBP (£), CHF (Fr), SEK (kr), NOK (kr)
- **Asia:** JPY (¥), CNY (¥), INR (₹), KRW (₩), SGD (S$)
- **Oceania:** AUD (A$), NZD (NZ$)

See the [full currency code list](https://en.wikipedia.org/wiki/ISO_4217#Active_codes).

#### Date & Time Formats

Choose how dates and times are displayed and entered throughout tmpo:

**Date Formats:**

- `MM/DD/YYYY` - US format (01/15/2024)
- `DD/MM/YYYY` - European format (15/01/2024)
- `YYYY-MM-DD` - ISO format (2024-01-15)

> [!NOTE]
> Your date format setting affects both display output (in logs, stats, etc.) and input prompts (when using `tmpo manual` or `tmpo edit`). The prompts will show and accept dates in your configured format.

**Time Formats:**

- `24-hour` - Military time (14:30, 23:45)
- `12-hour (AM/PM)` - Standard time (2:30 PM, 11:45 PM)

#### Timezone

Set your IANA timezone for accurate time tracking when working across time zones. Common examples:

- North America: `America/New_York`, `America/Chicago`, `America/Los_Angeles`
- Europe: `Europe/London`, `Europe/Paris`, `Europe/Berlin`
- Asia: `Asia/Tokyo`, `Asia/Singapore`, `Asia/Hong_Kong`
- Oceania: `Australia/Sydney`, `Pacific/Auckland`

Full list: [IANA Time Zone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones)

#### Export Path

Set a default directory where exported files (CSV, JSON) will be saved. This can be overridden per-project in `.tmporc` files.

**Setting the export path:**

```bash
tmpo config
# Export path (press Enter to keep current): ~/Documents/timesheets
```

**Clearing the export path:**

To remove the export path setting and revert to saving in the current directory:

```bash
tmpo config
# Export path (press Enter to keep current): clear
```

**How it works:**

- **Global setting** (`~/.tmpo/config.yaml`): Applies to all projects unless overridden
- **Project setting** (`.tmporc`): Overrides global setting for that project
- **If not set**: Files are exported to your current working directory
- **Supports `~`**: Use `~/Documents` instead of `/Users/yourname/Documents`

**Examples:**

```yaml
# Export to home directory
export_path: ~/exports

# Export to specific folder
export_path: /Users/dylan/Dropbox/timesheets

# No default (export to current directory)
export_path: ""
```

## Project Configuration

### The `.tmporc` File

Place a `.tmporc` file in your project root to customize tracking settings for that project. When you run tmpo commands from within the project directory (or any subdirectory), it will automatically use these settings.

### Creating a Configuration File

Use `tmpo init` to create a `.tmporc` file using an interactive form:

```bash
cd ~/projects/my-project
tmpo init
# You'll be prompted for:
# - Project name (defaults to auto-detected name)
# - Hourly rate (optional, press Enter to skip)
# - Description (optional, press Enter to skip)
# - Export path (optional, press Enter to skip)
```

For quick setup without prompts, use the `--accept-defaults` flag:

```bash
tmpo init --accept-defaults
# Creates .tmporc with auto-detected project name and default values
```

This creates a `.tmporc` file in the current directory.

### File Format

The `.tmporc` file uses YAML format:

```yaml
# tmpo project configuration
# This file configures time tracking settings for this project

# Project name (used to identify time entries)
project_name: My Awesome Project

# [OPTIONAL] Hourly rate for billing calculations (set to 0 to disable)
hourly_rate: 125.50

# [OPTIONAL] Description for this project
description: Client project for Acme Corp

# [OPTIONAL] Default export path for this project (overrides global export path)
export_path: ~/Documents/acme-timesheets
```

### Configuration Fields

#### `project_name` (required)

The name used to identify time entries for this project. This overrides automatic detection from git or directory names.

**Example:**

```yaml
project_name: Client Website Redesign
```

#### `hourly_rate` (optional)

Your billing rate per hour. When set, tmpo will calculate estimated earnings based on tracked time. The currency symbol displayed is determined by your global currency setting (see `tmpo config`).

**Example:**

```yaml
hourly_rate: 150.00
```

Set to `0` or omit to disable rate tracking:

```yaml
hourly_rate: 0
```

#### `description` (optional)

A longer description or notes about the project. This is for your reference and doesn't affect time tracking.

**Example:**

```yaml
description: Q1 2024 website redesign for Acme Corp. Main contact: john@acme.com
```

#### `export_path` (optional)

Default directory for exported files (CSV, JSON) for this project. This overrides the global export path setting from `tmpo config`.

**Example:**

```yaml
export_path: ~/Documents/client-timesheets
```

**How priority works:**

1. **Project `.tmporc` export path** - Highest priority (used if set)
2. **Global config export path** - Used if no project-specific path
3. **Current directory** - Default if neither is set

**Supports home directory expansion:**

```yaml
export_path: ~/Dropbox/timesheets     # Expands to /Users/yourname/Dropbox/timesheets
export_path: /absolute/path/exports   # Absolute paths work too
```

Set to empty string to export to current directory for this project:

```yaml
export_path: ""
```

## Project Detection Priority

When you run `tmpo start`, the project name is determined in this order:

1. **`.tmporc` file** - If present in current directory or any parent directory
2. **Git repository name** - The name of the git repository root folder
3. **Current directory name** - The name of your current working directory

This means you can override automatic detection by adding a `.tmporc` file.

### Example Scenarios

#### **Scenario 1:** Git repo with custom name

```bash
# Directory: ~/code/website-2024/
# Git repo name: website-2024
# No .tmporc file
tmpo start
# → Tracks to project "website-2024"
```

#### **Scenario 2:** With .tmporc override

```bash
# Directory: ~/code/website-2024/
# .tmporc contains: project_name: "Acme Website"
tmpo start
# → Tracks to project "Acme Website"
```

#### **Scenario 3:** Subdirectory detection

```bash
# Directory: ~/code/my-project/src/components/
# .tmporc exists at: ~/code/my-project/.tmporc
tmpo start
# → Uses .tmporc from project root
```

## Multi-Project Setup

### Separate Projects with Different Rates

Create a `.tmporc` in each project directory using `tmpo init`:

```bash
# Client A - $150/hour
cd ~/projects/client-a
tmpo init
# Project name: Client A - Web Development
# Hourly rate: 150
# Description: [press Enter to skip]

# Client B - different rate
cd ~/projects/client-b
tmpo init
# Project name: Client B - Game Development
# Hourly rate: 175
# Description: [press Enter to skip]

# Personal project - no billing
cd ~/projects/my-app
tmpo init --accept-defaults  # Quick setup with defaults
```

To change currency display (affects all projects):

```bash
tmpo config
# Select your preferred currency (USD, EUR, GBP, etc.)
```

Alternatively, you can manually create `.tmporc` files:

```bash
# Client configuration
cat > ~/projects/client-project/.tmporc << EOF
project_name: Client Project - Web Development
hourly_rate: 150.00
EOF
```

### Monorepo with Sub-Projects

In a monorepo, you can track different sub-projects separately:

```bash
# Main repo tracks to "My Company Platform"
cd ~/monorepo
echo "project_name: My Company Platform" > .tmporc

# But frontend team tracks separately
cd ~/monorepo/frontend
echo "project_name: Platform - Frontend" > .tmporc

# And backend team tracks separately
cd ~/monorepo/backend
echo "project_name: Platform - Backend" > .tmporc
```

## Version Control

### Should I commit `.tmporc`?

**Yes, commit it** *if*:

- Your team wants shared project naming
- It's an open source project and contributors might want to track time
- The configuration has no sensitive information

**Don't commit it** *if*:

- The hourly rate is personal/confidential
- Each team member prefers their own project naming

### Using `.gitignore`

To keep `.tmporc` files local:

```bash
echo ".tmporc" >> .gitignore
```

Or create a global gitignore:

```bash
echo ".tmporc" >> ~/.gitignore_global
git config --global core.excludesfile ~/.gitignore_global
```

## Migrating Data

### Backing Up Your Data

```bash
# Create a backup of your time tracking database
cp ~/.tmpo/tmpo.db ~/backups/tmpo-backup-$(date +%Y%m%d).db

# Optionally backup your global config too
cp ~/.tmpo/config.yaml ~/backups/tmpo-config-backup-$(date +%Y%m%d).yaml
```

### Moving to a New Machine

```bash
# On old machine - backup both database and config
cp ~/.tmpo/tmpo.db ~/tmpo-export.db
cp ~/.tmpo/config.yaml ~/tmpo-config.yaml

# Transfer files to new machine, then:
mkdir -p ~/.tmpo
cp ~/tmpo-export.db ~/.tmpo/tmpo.db
cp ~/tmpo-config.yaml ~/.tmpo/config.yaml
```

### Exporting for External Tools

Use `tmpo export` to get your data in portable formats:

```bash
# Export everything to CSV
tmpo export --output all-time-data.csv

# Export to JSON for programmatic access
tmpo export --format json --output all-time-data.json
```

See the [Usage Guide](usage.md#tmpo-export) for more export options.
