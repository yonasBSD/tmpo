package export

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
)

// ExportEntry represents a single time-tracking record prepared for JSON export.
// It contains the project name, the start timestamp, an optional end timestamp,
// the duration expressed in hours, and an optional human-readable description.
//
// Project is the associated project identifier or name.
// StartTime is the entry start timestamp as a string (for example, RFC3339).
// EndTime is the optional end timestamp; it will be omitted from JSON when empty.
// Duration is the total duration of the entry in hours as a floating-point value.
// Description is an optional text note; it will be omitted from JSON when empty.
type ExportEntry struct {
	Project string `json:"project"`
	StartTime string `json:"start_time"`
	EndTime string `json:"end_time,omitempty"`
	Duration float64 `json:"duration_hours"`
	Description string `json:"description,omitempty"`
}

// ToJson writes the given time entries to filename in pretty-printed JSON.
// Each storage.TimeEntry is converted to an ExportEntry with these mappings:
//   - Project: entry.ProjectName
//   - StartTime: formatted using layout "2006-01-02T15:04:05Z07:00" (RFC3339-like)
//   - EndTime: formatted using the same layout if entry.EndTime is non-nil; omitted otherwise
//   - Duration: entry.Duration().Hours() (floating-point hours)
//   - Description: entry.Description
//
// The function creates or truncates the target file, encodes the slice of
// ExportEntry values with json.Encoder and indentation, and closes the file
// before returning. It returns an error if the file cannot be created or if
// JSON encoding fails. Callers must ensure the destination path is writable.
func ToJson(entries []*storage.TimeEntry, filename string) error {
	var exportEntries []ExportEntry

	for _, entry := range entries {
		export := ExportEntry{
			Project: entry.ProjectName,
			StartTime: entry.StartTime.Format("2006-01-02T15:04:05Z07:00"),
			Duration: entry.Duration().Hours(),
			Description: entry.Description,
		}

		if entry.EndTime != nil {
			export.EndTime = entry.EndTime.Format("2006-01-02T15:04:05Z07:00")
		}

		exportEntries = append(exportEntries, export)
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create JSON file: %w", err)
	}

	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(exportEntries); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}