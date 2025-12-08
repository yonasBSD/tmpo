package storage

import "time"

// TimeEntry represents a recorded period of work on a project.
// It includes a unique identifier, the project name, the start time,
// an optional end time (nil indicates the entry is still in progress),
// and a free-form description of the work performed.
type TimeEntry struct {
	ID int64
	ProjectName string
	StartTime time.Time
	EndTime *time.Time
	Description string
}

// Duration returns the elapsed time for the TimeEntry.
// If EndTime is non-nil, it returns the difference EndTime.Sub(StartTime).
// If EndTime is nil (the entry is ongoing), it returns time.Since(StartTime).
func (t *TimeEntry) Duration() time.Duration {
	if( t.EndTime == nil) {
		return time.Since(t.StartTime)
	}

	return t.EndTime.Sub(t.StartTime)
}

// IsRunning reports whether the TimeEntry is currently running.
// It returns true when EndTime is nil, indicating no end timestamp has been set.
func (t *TimeEntry) IsRunning() bool {
	return t.EndTime == nil
}