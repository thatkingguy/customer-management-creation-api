package utils

import (
	"fmt"
	"time"
)

// ParseDate parses a date string in ISO 8601 or RFC 2822 format
func ParseDate(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date string is empty")
	}

	// Try ISO 8601 formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05",
		"2006-01-02",
		time.RFC1123,
		time.RFC1123Z,
		time.RFC822,
		time.RFC822Z,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("date must be a valid date string in ISO 8601 or RFC 2822 format")
}

// FormatDate formats a time.Time to YYYY-MM-DD string
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// CalculateAge calculates age from date of birth
func CalculateAge(dateOfBirth time.Time) int {
	now := time.Now()
	age := now.Year() - dateOfBirth.Year()
	
	// Adjust if birthday hasn't occurred this year
	if now.YearDay() < dateOfBirth.YearDay() {
		age--
	}
	
	return age
}



