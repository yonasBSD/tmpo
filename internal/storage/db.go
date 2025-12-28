package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Database is a thin wrapper around an *sql.DB connection pool.
// It centralizes database access for the package and provides a
// single place to implement queries, transactions, migrations and
// other persistence-related helpers.
//
// The underlying *sql.DB is safe for concurrent use, but a Database
// with a nil internal pointer is not usable — it must be initialized
// (for example via sql.Open) and closed when no longer needed.
type Database struct {
	db *sql.DB
}

// Initialize ensures the on-disk storage for the application exists, opens the
// SQLite database, and returns a Database wrapper.
//
// Specifically, Initialize:
//   - determines the current user's home directory,
//   - checks the TMPO_DEV environment variable:
//     - if TMPO_DEV is "1" or "true", uses "$HOME/.tmpo-dev" (development mode),
//     - otherwise uses "$HOME/.tmpo" (production mode, the default),
//   - creates the appropriate directory if it does not already exist,
//   - opens (or creates) the SQLite database file "tmpo.db" in that directory,
//   - ensures the time_entries table exists with the schema:
//       id INTEGER PRIMARY KEY AUTOINCREMENT,
//       project_name TEXT NOT NULL,
//       start_time DATETIME NOT NULL,
//       end_time DATETIME,
//       description TEXT
//
// On success it returns a *Database containing the opened *sql.DB. It returns
// a non-nil error if any step fails: obtaining the home directory, creating
// the directory, opening the database, or creating the table. The caller is
// responsible for closing the database when finished.
func Initialize() (*Database, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	// Switch out directory depending on build environment
	tmpoDir := filepath.Join(homeDir, ".tmpo")
	if devMode := os.Getenv("TMPO_DEV"); devMode == "1" || devMode == "true" {
		tmpoDir = filepath.Join(homeDir, ".tmpo-dev")
	}

	if err := os.MkdirAll(tmpoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .tmpo directory: %w", err)
	}

	dbPath := filepath.Join(tmpoDir, "tmpo.db")
	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS time_entries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			description TEXT,
			hourly_rate REAL
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	// Create milestones table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS milestones (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_name TEXT NOT NULL,
			name TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			UNIQUE(project_name, name)
		)
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to create milestones table: %w", err)
	}

	// Add milestone_name column to time_entries (safe if already exists)
	_, err = db.Exec(`ALTER TABLE time_entries ADD COLUMN milestone_name TEXT`)
	// Ignore error if column already exists (SQLite doesn't have IF NOT EXISTS for ALTER)
	if err != nil && !isColumnExistsError(err) {
		return nil, fmt.Errorf("failed to add milestone_name column: %w", err)
	}

	// Create indexes for performance
	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_time_entries_milestone ON time_entries(milestone_name)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_milestones_project_active ON milestones(project_name, end_time)`)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	return &Database{db: db}, nil
}

// isColumnExistsError checks if the error is due to a duplicate column.
// SQLite returns "duplicate column name" when trying to add a column that already exists.
func isColumnExistsError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate column name") ||
		strings.Contains(errMsg, "duplicate column")
}

// CreateEntry inserts a new time entry for the specified projectName with the given
// description, hourlyRate, and milestoneName. The entry's start_time is set to the current time.
// If hourlyRate is nil, the hourly_rate column will be set to NULL. If milestoneName is nil,
// the milestone_name column will be set to NULL. On success it returns the created *TimeEntry
// (retrieved by querying the database for the last insert id). If the insert or the subsequent
// retrieval fails, an error wrapping the underlying database error is returned.
func (d *Database) CreateEntry(projectName, description string, hourlyRate *float64, milestoneName *string) (*TimeEntry, error) {
	var rate sql.NullFloat64
	if hourlyRate != nil {
		rate = sql.NullFloat64{Float64: *hourlyRate, Valid: true}
	}

	var milestone sql.NullString
	if milestoneName != nil {
		milestone = sql.NullString{String: *milestoneName, Valid: true}
	}

	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, description, hourly_rate, milestone_name) VALUES (?, ?, ?, ?, ?)",
		projectName,
		time.Now(),
		description,
		rate,
		milestone,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create entry: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetEntry(id)
}

// CreateManualEntry inserts a completed time entry with specific start and end times.
// Unlike CreateEntry which uses the current time and leaves end_time NULL, this method
// creates a fully specified historical entry for manual record-keeping.
// If hourlyRate is nil, the hourly_rate column will be set to NULL. If milestoneName is nil,
// the milestone_name column will be set to NULL. On success it returns the created *TimeEntry
// (retrieved by querying the database for the last insert id). If the insert or the subsequent
// retrieval fails, an error wrapping the underlying database error is returned.
func (d *Database) CreateManualEntry(projectName, description string, startTime, endTime time.Time, hourlyRate *float64, milestoneName *string) (*TimeEntry, error) {
	var rate sql.NullFloat64
	if hourlyRate != nil {
		rate = sql.NullFloat64{Float64: *hourlyRate, Valid: true}
	}

	var milestone sql.NullString
	if milestoneName != nil {
		milestone = sql.NullString{String: *milestoneName, Valid: true}
	}

	result, err := d.db.Exec(
		"INSERT INTO time_entries (project_name, start_time, end_time, description, hourly_rate, milestone_name) VALUES (?, ?, ?, ?, ?, ?)",
		projectName,
		startTime,
		endTime,
		description,
		rate,
		milestone,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create manual entry: %w", err)
	}

	id, err := result.LastInsertId()

	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetEntry(id)
}

// GetRunningEntry retrieves the most recently started time entry that is still running
// (i.e. has a NULL end_time) from the time_entries table. The query orders by
// start_time descending and returns at most one row.
//
// If there is no running entry, GetRunningEntry returns (nil, nil). If the database
// query or scan fails, it returns a non-nil error describing the failure.
//
// The function scans id, project_name, start_time, end_time, description, hourly_rate,
// and milestone_name into a TimeEntry. The EndTime field on the returned TimeEntry is set
// only if the scanned end_time is non-NULL. The HourlyRate field is set only if the scanned
// hourly_rate is non-NULL. The MilestoneName field is set only if the scanned milestone_name
// is non-NULL.
func (d *Database) GetRunningEntry() (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE end_time IS NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get running entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

// GetLastStoppedEntry retrieves the most recently stopped time entry (i.e. has a non-NULL
// end_time) from the time_entries table. The query orders by start_time descending and
// returns at most one row.
//
// If there is no stopped entry, GetLastStoppedEntry returns (nil, nil). If the database
// query or scan fails, it returns a non-nil error describing the failure.
//
// The function scans id, project_name, start_time, end_time, description, hourly_rate,
// and milestone_name into a TimeEntry. Since this query only returns stopped entries, the
// EndTime field will always be non-nil. The HourlyRate and MilestoneName fields are set
// only if their scanned values are non-NULL.
func (d *Database) GetLastStoppedEntry() (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE end_time IS NOT NULL
		ORDER BY start_time DESC
		LIMIT 1
	`).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get last stopped entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

// StopEntry sets the end_time of the time entry identified by id to the current time.
// It updates the corresponding row in the time_entries table using time.Now().
// If the update fails (for example if the row does not exist or the database returns an error),
// an error is returned wrapped with context. This method overwrites any existing end_time value
// and does not return the updated entry or perform additional validation on id.
func (d *Database) StopEntry(id int64) error {
	_, err := d.db.Exec(
		"UPDATE time_entries SET end_time = ? WHERE id = ?",
		time.Now(),
		id,
	)

	if(err != nil) {
		return fmt.Errorf("failed to stop entry: %w", err)
	}

	return nil
}

// GetEntry retrieves a TimeEntry by its ID from the database.
// It queries the time_entries table for id, project_name, start_time, end_time, description, hourly_rate,
// and milestone_name, scans the result into a TimeEntry value, and returns a pointer to it.
// If the end_time column is NULL in the database, the returned TimeEntry.EndTime will be nil;
// otherwise EndTime will point to the retrieved time value. If the hourly_rate column is NULL,
// the returned TimeEntry.HourlyRate will be nil; otherwise HourlyRate will point to the retrieved value.
// If the milestone_name column is NULL, the returned TimeEntry.MilestoneName will be nil.
// If no row is found or an error occurs during query/scan, an error is returned (wrapped with
// the context "failed to get entry").
func (d *Database) GetEntry(id int64) (*TimeEntry, error) {
	var entry TimeEntry
	var endTime sql.NullTime
	var hourlyRate sql.NullFloat64
	var milestoneName sql.NullString

	err := d.db.QueryRow(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE id = ?
	`, id).Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)

	if err != nil {
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	if endTime.Valid {
		entry.EndTime = &endTime.Time
	}

	if hourlyRate.Valid {
		entry.HourlyRate = &hourlyRate.Float64
	}

	if milestoneName.Valid {
		entry.MilestoneName = &milestoneName.String
	}

	return &entry, nil
}

// GetEntries retrieves time entries from the Database.
//
// It returns time entries ordered by start_time in descending order. If limit > 0,
// at most `limit` entries are returned; if limit <= 0 all matching entries are returned.
// Each returned element is a pointer to a TimeEntry. The EndTime field of a TimeEntry
// will be nil when the corresponding end_time column in the database is NULL. The HourlyRate
// field will be nil when the corresponding hourly_rate column in the database is NULL. The
// MilestoneName field will be nil when the corresponding milestone_name column is NULL.
//
// The function performs a SQL query selecting id, project_name, start_time, end_time,
// description, hourly_rate, and milestone_name. It returns a slice of entries and an error if
// the query or row scanning fails; any underlying error is wrapped.
func (d *Database) GetEntries(limit int) ([]*TimeEntry, error) {
	query := `
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		ORDER BY start_time DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetEntriesByProject retrieves time entries for the specified projectName from the
// time_entries table. Results are ordered by start_time in descending order (newest first).
//
// For each row a TimeEntry is populated. If the end_time column is NULL the returned
// TimeEntry.EndTime will be nil; otherwise EndTime will point to the scanned time.Time.
// If the hourly_rate column is NULL the returned TimeEntry.HourlyRate will be nil;
// otherwise HourlyRate will point to the scanned float64. If the milestone_name column
// is NULL the returned TimeEntry.MilestoneName will be nil.
//
// On success the function returns a slice of pointers to TimeEntry. If there are no
// matching rows the returned slice will have length 0 (it may be nil). On failure the
// function returns a non-nil error and a nil slice. Errors may originate from the
// query execution, row scanning, or row iteration.
func (d *Database) GetEntriesByProject(projectName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE project_name = ?
		ORDER BY start_time DESC
	`, projectName)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetEntriesByDateRange retrieves time entries whose start_time falls between start and end (inclusive).
// Results are returned in descending order by start_time.
// The provided start and end times are passed to the database driver as-is; callers should ensure they use the intended timezone/representation.
// For rows with a NULL end_time the returned TimeEntry.EndTime will be nil; otherwise EndTime points to the parsed time value.
// For rows with a NULL hourly_rate the returned TimeEntry.HourlyRate will be nil; otherwise HourlyRate points to the parsed float64.
// For rows with a NULL milestone_name the returned TimeEntry.MilestoneName will be nil.
// Returns a slice of pointers to TimeEntry (which may be empty) or an error if the database query or row scanning fails.
func (d *Database) GetEntriesByDateRange(start, end time.Time) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE start_time BETWEEN ? AND ?
		ORDER BY start_time DESC
	`, start, end)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// GetAllProjects retrieves all distinct project names from the time_entries table.
// The results are returned in ascending order by project_name.
// On success it returns a slice of project names (which will be empty if no projects exist)
// and a nil error. If the underlying database query or a row scan fails, it returns a
// non-nil error describing the failure.
func (d *Database) GetAllProjects() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT project_name
		FROM time_entries
		ORDER BY project_name
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	defer rows.Close()

	var projects []string

	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetProjectsWithCompletedEntries retrieves all distinct project names that have at least
// one completed time entry (end_time IS NOT NULL) from the time_entries table.
// The results are returned in ascending order by project_name.
// On success it returns a slice of project names (which will be empty if no completed entries exist)
// and a nil error. If the underlying database query or a row scan fails, it returns a
// non-nil error describing the failure.
func (d *Database) GetProjectsWithCompletedEntries() ([]string, error) {
	rows, err := d.db.Query(`
		SELECT DISTINCT project_name
		FROM time_entries
		WHERE end_time IS NOT NULL
		ORDER BY project_name
	`)

	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	defer rows.Close()

	var projects []string

	for rows.Next() {
		var project string
		if err := rows.Scan(&project); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// GetCompletedEntriesByProject retrieves completed time entries (where end_time IS NOT NULL)
// for the specified projectName from the time_entries table. Results are ordered by start_time
// in descending order (newest first).
//
// For each row a TimeEntry is populated. Since this query only returns completed entries,
// the EndTime field will always be non-nil. If the hourly_rate column is NULL the returned
// TimeEntry.HourlyRate will be nil; otherwise HourlyRate will point to the scanned float64.
// If the milestone_name column is NULL the returned TimeEntry.MilestoneName will be nil.
//
// On success the function returns a slice of pointers to TimeEntry. If there are no
// matching rows the returned slice will have length 0 (it may be nil). On failure the
// function returns a non-nil error and a nil slice. Errors may originate from the
// query execution, row scanning, or row iteration.
func (d *Database) GetCompletedEntriesByProject(projectName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(`
		SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name
		FROM time_entries
		WHERE project_name = ? AND end_time IS NOT NULL
		ORDER BY start_time DESC
	`, projectName)

	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}

	defer rows.Close()

	var entries []*TimeEntry

	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// UpdateTimeEntry updates an existing time entry in the database with the values from the provided
// TimeEntry. It updates the project_name, start_time, end_time, description, hourly_rate, and
// milestone_name fields for the entry with the matching ID.
//
// If the provided entry's EndTime is nil, the end_time column will be set to NULL.
// If the provided entry's HourlyRate is nil, the hourly_rate column will be set to NULL.
// If the provided entry's MilestoneName is nil, the milestone_name column will be set to NULL.
//
// Returns an error if the update fails. Does not verify that a row with the given ID exists;
// if no rows are affected the function will still return nil (no error).
func (d *Database) UpdateTimeEntry(id int64, entry *TimeEntry) error {
	var endTime sql.NullTime
	if entry.EndTime != nil {
		endTime = sql.NullTime{Time: *entry.EndTime, Valid: true}
	}

	var hourlyRate sql.NullFloat64
	if entry.HourlyRate != nil {
		hourlyRate = sql.NullFloat64{Float64: *entry.HourlyRate, Valid: true}
	}

	var milestoneName sql.NullString
	if entry.MilestoneName != nil {
		milestoneName = sql.NullString{String: *entry.MilestoneName, Valid: true}
	}

	_, err := d.db.Exec(`
		UPDATE time_entries
		SET project_name = ?, start_time = ?, end_time = ?, description = ?, hourly_rate = ?, milestone_name = ?
		WHERE id = ?
	`, entry.ProjectName, entry.StartTime, endTime, entry.Description, hourlyRate, milestoneName, id)

	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	return nil
}

// DeleteTimeEntry deletes a time entry from the database by its ID.
// Returns an error if the deletion fails. Does not verify that a row with the given ID exists;
// if no rows are affected the function will still return nil (no error).
func (d *Database) DeleteTimeEntry(id int64) error {
	_, err := d.db.Exec("DELETE FROM time_entries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}

// CreateMilestone inserts a new milestone for the specified projectName with the given
// name. The milestone's start_time is set to the current time and end_time is NULL
// (indicating active status). On success it returns the created *Milestone (retrieved
// by querying the database for the last insert id). If a milestone with the same
// project_name and name already exists, an error is returned.
func (d *Database) CreateMilestone(projectName, name string) (*Milestone, error) {
	result, err := d.db.Exec(
		"INSERT INTO milestones (project_name, name, start_time) VALUES (?, ?, ?)",
		projectName,
		name,
		time.Now(),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create milestone: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return d.GetMilestone(id)
}

// GetMilestone retrieves a Milestone by its ID from the database.
// Returns (nil, nil) if no milestone with the given ID exists.
func (d *Database) GetMilestone(id int64) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE id = ?",
		id,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get milestone: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

// GetActiveMilestoneForProject retrieves the currently active milestone for the
// specified projectName (i.e., has a NULL end_time). Returns (nil, nil) if no
// active milestone exists for the project.
func (d *Database) GetActiveMilestoneForProject(projectName string) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? AND end_time IS NULL ORDER BY start_time DESC LIMIT 1",
		projectName,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get active milestone: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

// GetMilestoneByName retrieves a milestone by its project name and milestone name.
// Returns (nil, nil) if not found.
func (d *Database) GetMilestoneByName(projectName, milestoneName string) (*Milestone, error) {
	var milestone Milestone
	var endTime sql.NullTime

	err := d.db.QueryRow(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? AND name = ?",
		projectName,
		milestoneName,
	).Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get milestone by name: %w", err)
	}

	if endTime.Valid {
		milestone.EndTime = &endTime.Time
	}

	return &milestone, nil
}

// GetMilestonesByProject retrieves all milestones for the specified projectName,
// ordered by start_time DESC (newest first).
func (d *Database) GetMilestonesByProject(projectName string) ([]*Milestone, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, name, start_time, end_time FROM milestones WHERE project_name = ? ORDER BY start_time DESC",
		projectName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		var milestone Milestone
		var endTime sql.NullTime

		err := rows.Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}

		if endTime.Valid {
			milestone.EndTime = &endTime.Time
		}

		milestones = append(milestones, &milestone)
	}

	return milestones, nil
}

// GetAllMilestones retrieves all milestones from all projects,
// ordered by start_time DESC (newest first).
func (d *Database) GetAllMilestones() ([]*Milestone, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, name, start_time, end_time FROM milestones ORDER BY start_time DESC",
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get all milestones: %w", err)
	}
	defer rows.Close()

	var milestones []*Milestone
	for rows.Next() {
		var milestone Milestone
		var endTime sql.NullTime

		err := rows.Scan(&milestone.ID, &milestone.ProjectName, &milestone.Name, &milestone.StartTime, &endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to scan milestone: %w", err)
		}

		if endTime.Valid {
			milestone.EndTime = &endTime.Time
		}

		milestones = append(milestones, &milestone)
	}

	return milestones, nil
}

// FinishMilestone sets the end_time of the milestone identified by id to the current time.
// This marks the milestone as completed.
func (d *Database) FinishMilestone(id int64) error {
	_, err := d.db.Exec(
		"UPDATE milestones SET end_time = ? WHERE id = ?",
		time.Now(),
		id,
	)

	if err != nil {
		return fmt.Errorf("failed to finish milestone: %w", err)
	}

	return nil
}

// GetEntriesByMilestone retrieves time entries for the specified projectName and
// milestoneName from the time_entries table. Results are ordered by start_time DESC.
func (d *Database) GetEntriesByMilestone(projectName, milestoneName string) ([]*TimeEntry, error) {
	rows, err := d.db.Query(
		"SELECT id, project_name, start_time, end_time, description, hourly_rate, milestone_name FROM time_entries WHERE project_name = ? AND milestone_name = ? ORDER BY start_time DESC",
		projectName,
		milestoneName,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get entries by milestone: %w", err)
	}
	defer rows.Close()

	var entries []*TimeEntry
	for rows.Next() {
		var entry TimeEntry
		var endTime sql.NullTime
		var hourlyRate sql.NullFloat64
		var milestoneName sql.NullString

		err := rows.Scan(&entry.ID, &entry.ProjectName, &entry.StartTime, &endTime, &entry.Description, &hourlyRate, &milestoneName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		if endTime.Valid {
			entry.EndTime = &endTime.Time
		}

		if hourlyRate.Valid {
			entry.HourlyRate = &hourlyRate.Float64
		}

		if milestoneName.Valid {
			entry.MilestoneName = &milestoneName.String
		}

		entries = append(entries, &entry)
	}

	return entries, nil
}

// Close closes the Database, releasing any underlying resources.
// It delegates to the wrapped database's Close method and returns any error encountered.
// After Close is called, the Database must not be used for further operations.
func (d *Database) Close() error {
	return d.db.Close()
}