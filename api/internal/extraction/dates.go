package extraction

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DateNormalizer parses and normalizes date strings.
type DateNormalizer struct {
	// Reference time for relative dates
	ReferenceTime time.Time
}

// NewDateNormalizer creates a new date normalizer.
func NewDateNormalizer() *DateNormalizer {
	return &DateNormalizer{
		ReferenceTime: time.Now(),
	}
}

// Patterns for date matching
var (
	// Exact date patterns
	exactDatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(\d{1,2})\s+(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{4})\b`),
		regexp.MustCompile(`(?i)(January|February|March|April|May|June|July|August|September|October|November|December)\s+(\d{1,2}),?\s+(\d{4})`),
		regexp.MustCompile(`(?i)\b(\d{1,2})/(\d{1,2})/(\d{4})\b`),
		regexp.MustCompile(`(?i)\b(\d{4})-(\d{1,2})-(\d{1,2})\b`),
	}

	// Day of week patterns
	dayOfWeekPattern = regexp.MustCompile(`(?i)\b(Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)\b`)

	// Month patterns
	monthPattern = regexp.MustCompile(`(?i)\b(that\s+)?(this\s+)?(January|February|March|April|May|June|July|August|September|October|November|December)\b`)

	// Season patterns
	seasonPattern = regexp.MustCompile(`(?i)\b(that\s+)?(this\s+)?(the\s+)?(following\s+)?(next\s+)?(spring|summer|autumn|fall|winter)\b`)

	// Year pattern
	yearPattern = regexp.MustCompile(`(?i)\b(in\s+)?(the\s+)?(year\s+)?(\d{4})\b`)

	// Relative patterns
	relativePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(\d+)\s+days?\s+(later|after|ago)\b`),
		regexp.MustCompile(`(?i)\b(\d+)\s+weeks?\s+(later|after|ago)\b`),
		regexp.MustCompile(`(?i)\b(\d+)\s+months?\s+(later|after|ago)\b`),
		regexp.MustCompile(`(?i)\b(\d+)\s+years?\s+(later|after|ago)\b`),
		regexp.MustCompile(`(?i)\b(the\s+)?(following|next)\s+(day|week|month|year)\b`),
		regexp.MustCompile(`(?i)\b(a\s+)?few\s+days?\s+later\b`),
		regexp.MustCompile(`(?i)\bsome\s+time\s+later\b`),
		regexp.MustCompile(`(?i)\b(the\s+)?next\s+(morning|afternoon|evening|night)\b`),
	}

	// Approximate patterns
	approximatePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\bone\s+(morning|afternoon|evening|night)\b`),
		regexp.MustCompile(`(?i)\bsometime\s+(later|afterwards|soon)\b`),
		regexp.MustCompile(`(?i)\ba\s+while\s+later\b`),
	}
)

// monthToNumber maps month names to numbers.
var monthToNumber = map[string]int{
	"january":   1,
	"february":  2,
	"march":     3,
	"april":     4,
	"may":       5,
	"june":      6,
	"july":      7,
	"august":    8,
	"september": 9,
	"october":   10,
	"november":  11,
	"december":  12,
}

// seasonToMonths maps seasons to month ranges.
var seasonToMonths = map[string][2]int{
	"spring": {3, 5},
	"summer": {6, 8},
	"autumn": {9, 11},
	"fall":   {9, 11},
	"winter": {12, 2},
}

// NormalizeDate parses a date string and returns normalized date range with precision.
func (n *DateNormalizer) NormalizeDate(dateText string, referenceDate *time.Time) (*ParsedDate, error) {
	if referenceDate == nil {
		ref := n.ReferenceTime
		referenceDate = &ref
	}

	dateText = strings.TrimSpace(dateText)

	// Try exact date patterns first
	for _, pattern := range exactDatePatterns {
		if matches := pattern.FindStringSubmatch(dateText); matches != nil {
			return n.parseExactDate(matches)
		}
	}

	// Check for day of week
	if dayOfWeekPattern.MatchString(dateText) {
		return &ParsedDate{
			DatePrecision: "exact",
		}, nil
	}

	// Check for month patterns
	if matches := monthPattern.FindStringSubmatch(dateText); matches != nil {
		monthName := matches[len(matches)-1]
		if monthNum, ok := monthToNumber[strings.ToLower(monthName)]; ok {
			year := referenceDate.Year()
			return &ParsedDate{
				DateStart:     date(year, monthNum, 1),
				DateEnd:       date(year, monthNum, daysInMonth(monthNum, year)),
				DatePrecision: "month",
			}, nil
		}
	}

	// Check for season patterns
	if matches := seasonPattern.FindStringSubmatch(dateText); matches != nil {
		seasonName := matches[len(matches)-1]
		if months, ok := seasonToMonths[strings.ToLower(seasonName)]; ok {
			year := referenceDate.Year()
			startMonth := months[0]
			endMonth := months[1]

			// Handle winter spanning years
			if startMonth > endMonth {
				return &ParsedDate{
					DateStart:     date(year, startMonth, 1),
					DateEnd:       date(year+1, endMonth, daysInMonth(endMonth, year+1)),
					DatePrecision: "season",
				}, nil
			}

			return &ParsedDate{
				DateStart:     date(year, startMonth, 1),
				DateEnd:       date(year, endMonth, daysInMonth(endMonth, year)),
				DatePrecision: "season",
			}, nil
		}
	}

	// Check for year patterns
	if matches := yearPattern.FindStringSubmatch(dateText); matches != nil {
		yearStr := matches[len(matches)-1]
		var year int
		if _, err := fmt.Sscanf(yearStr, "%d", &year); err == nil {
			return &ParsedDate{
				DateStart:     date(year, 1, 1),
				DateEnd:       date(year, 12, 31),
				DatePrecision: "year",
			}, nil
		}
	}

	// Check for relative patterns
	for _, pattern := range relativePatterns {
		if pattern.MatchString(dateText) {
			return &ParsedDate{
				DatePrecision: "relative",
			}, nil
		}
	}

	// Check for approximate patterns
	for _, pattern := range approximatePatterns {
		if pattern.MatchString(dateText) {
			return &ParsedDate{
				DatePrecision: "approximate",
			}, nil
		}
	}

	// Unknown precision
	return &ParsedDate{
		DatePrecision: "unknown",
	}, nil
}

// parseExactDate parses exact date patterns.
func (n *DateNormalizer) parseExactDate(matches []string) (*ParsedDate, error) {
	// Try different pattern formats
	if len(matches) >= 4 {
		// Format: "15 March 1805" or "March 15, 1805"
		var day, month, year int
		var yearStr string

		// Check if first group is day or month
		if _, err := fmt.Sscanf(matches[1], "%d", &day); err == nil && day <= 31 {
			// Day first: "15 March 1805"
			monthName := matches[2]
			yearStr = matches[3]
			if monthNum, ok := monthToNumber[strings.ToLower(monthName)]; ok {
				month = monthNum
			}
		} else {
			// Month first: "March 15, 1805"
			monthName := matches[1]
			yearStr = matches[3]
			if monthNum, ok := monthToNumber[strings.ToLower(monthName)]; ok {
				month = monthNum
			}
			fmt.Sscanf(matches[2], "%d", &day)
		}

		if _, err := fmt.Sscanf(yearStr, "%d", &year); err == nil {
			return &ParsedDate{
				DateStart:     date(year, month, day),
				DateEnd:       date(year, month, day),
				DatePrecision: "exact",
			}, nil
		}
	}

	// Format: "2025-03-15" or "03/15/2025"
	if len(matches) >= 4 {
		var year, month, day int
		if _, err := fmt.Sscanf(matches[1], "%d", &year); err == nil && year > 1000 {
			// ISO format: "2025-03-15"
			fmt.Sscanf(matches[2], "%d", &month)
			fmt.Sscanf(matches[3], "%d", &day)
		} else {
			// US format: "03/15/2025"
			fmt.Sscanf(matches[1], "%d", &month)
			fmt.Sscanf(matches[2], "%d", &day)
			fmt.Sscanf(matches[3], "%d", &year)
		}

		return &ParsedDate{
			DateStart:     date(year, month, day),
			DateEnd:       date(year, month, day),
			DatePrecision: "exact",
		}, nil
	}

	return &ParsedDate{
		DatePrecision: "unknown",
	}, nil
}

// ResolveRelativeDate resolves a relative date against a reference date.
func (n *DateNormalizer) ResolveRelativeDate(dateText string, referenceDate time.Time) (time.Time, error) {
	dateText = strings.ToLower(strings.TrimSpace(dateText))

	// Pattern: "3 days later"
	re := regexp.MustCompile(`(\d+)\s+days?\s+(later|after|ago)`)
	if matches := re.FindStringSubmatch(dateText); matches != nil {
		var days int
		fmt.Sscanf(matches[1], "%d", &days)
		if matches[2] == "ago" {
			return referenceDate.AddDate(0, 0, -days), nil
		}
		return referenceDate.AddDate(0, 0, days), nil
	}

	// Pattern: "3 weeks later"
	re = regexp.MustCompile(`(\d+)\s+weeks?\s+(later|after|ago)`)
	if matches := re.FindStringSubmatch(dateText); matches != nil {
		var weeks int
		fmt.Sscanf(matches[1], "%d", &weeks)
		if matches[2] == "ago" {
			return referenceDate.AddDate(0, 0, -weeks*7), nil
		}
		return referenceDate.AddDate(0, 0, weeks*7), nil
	}

	// Pattern: "the following day/week"
	if strings.Contains(dateText, "following day") || strings.Contains(dateText, "next day") {
		return referenceDate.AddDate(0, 0, 1), nil
	}
	if strings.Contains(dateText, "following week") || strings.Contains(dateText, "next week") {
		return referenceDate.AddDate(0, 0, 7), nil
	}

	return time.Time{}, fmt.Errorf("unrecognized relative date pattern")
}

// date creates a time.Time pointer for a given date.
func date(year, month, day int) *time.Time {
	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &d
}

// daysInMonth returns the number of days in a month.
func daysInMonth(month int, year int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if isLeapYear(year) {
			return 29
		}
		return 28
	default:
		return 30
	}
}

// isLeapYear checks if a year is a leap year.
func isLeapYear(year int) bool {
	if year%4 != 0 {
		return false
	}
	if year%100 != 0 {
		return true
	}
	return year%400 == 0
}
