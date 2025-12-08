package export

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/storage"
)

// ToCSV writes the provided slice of time entries to a CSV file at the given
// filename. The file is created (or truncated if it already exists) and a CSV
// writer is used to emit a header row followed by one record per entry.
//
// The CSV contains the following columns in order:
//   - "Project"            : entry.ProjectName
//   - "Start Time"         : entry.StartTime formatted as "2006-01-02 15:04:05"
//   - "End Time"           : entry.EndTime formatted as "2006-01-02 15:04:05" or
//                           an empty string if EndTime is nil
//   - "Duration (hours)"   : entry.Duration() expressed in hours, formatted
//                           with two decimal places
//   - "Description"        : entry.Description
//
// The function returns an error if the file cannot be created or if writing
// any header/record fails. The CSV writer is flushed before returning, and the
// file is closed via deferred cleanup. The caller should ensure the provided
// filename is writable.
func ToCSV(entries []*storage.TimeEntry, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)

	defer writer.Flush()

	header := []string{"Project", "Start Time", "End Time", "Duration (hours)", "Description"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	for _, entry := range entries {
		endTime := ""
		if entry.EndTime != nil {
			endTime = entry.EndTime.Format("2006-01-02 15:04:05")
		}

		duration := entry.Duration().Hours()

		record := []string{
			entry.ProjectName,
			entry.StartTime.Format("2006-01-02 15:04:05"),
			endTime,
			fmt.Sprintf("%.2f", duration),
			entry.Description,
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	return nil
}