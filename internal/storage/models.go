package storage

import (
	"math"
	"time"
)

// TimeEntry represents a recorded period of work on a project.
// It includes a unique identifier, the project name, the start time,
// an optional end time (nil indicates the entry is still in progress),
// a free-form description of the work performed, and an optional hourly rate
// (nil indicates no rate was configured when the entry was created).
type TimeEntry struct {
	ID int64
	ProjectName string
	StartTime time.Time
	EndTime *time.Time
	Description string
	HourlyRate *float64
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

// RoundedHours returns the duration in hours rounded to 2 decimal places.
// This rounding is used for earnings calculations to ensure transparency:
// the displayed hours value (e.g., "1.83 hours") matches exactly what is
// used in billing calculations.
//
// Future enhancement: This could be made configurable via user settings
// to support different rounding increments (e.g., 0.1 hours for 6-minute billing,
// or 0.25 hours for 15-minute billing).
func (t *TimeEntry) RoundedHours() float64 {
	return math.Round(t.Duration().Hours()*100) / 100
}