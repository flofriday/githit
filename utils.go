package main

import "time"

// Rounds the time down to the next hour and converts it to UNIX time
// Example 18:47 -> 18:00
func roundToUnixHour(t time.Time) int64 {
	// Round down to the hour
	hour := (t.UnixNano() / time.Hour.Nanoseconds()) * time.Hour.Nanoseconds()

	// Convert from Nanoseconds to Seconds
	hour /= time.Second.Nanoseconds()
	return hour
}
