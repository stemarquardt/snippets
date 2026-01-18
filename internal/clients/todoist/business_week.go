package todoist

import "time"

// BusinessWeek represents a week from Monday to Sunday
type BusinessWeek struct {
	Start time.Time
	End   time.Time
}

// GetCurrentBusinessWeek returns the current business week (Monday to today or Sunday, whichever is earlier)
func GetCurrentBusinessWeek() BusinessWeek {
	now := time.Now()
	return GetBusinessWeekForDate(now)
}

// GetBusinessWeekForDate returns the business week for a given date
// The week runs from Monday (start of day) to Sunday (end of day)
func GetBusinessWeekForDate(date time.Time) BusinessWeek {
	// Normalize to start of day
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// Calculate days since Monday (0 = Monday, 6 = Sunday)
	weekday := int(date.Weekday())
	if weekday == 0 { // Sunday is 0 in time.Weekday, we want it to be 6
		weekday = 6
	} else {
		weekday = weekday - 1
	}

	// Get Monday of this week
	monday := date.AddDate(0, 0, -weekday)

	// Get Sunday of this week (end of day)
	sunday := monday.AddDate(0, 0, 6)
	sunday = time.Date(sunday.Year(), sunday.Month(), sunday.Day(), 23, 59, 59, 0, sunday.Location())

	return BusinessWeek{
		Start: monday,
		End:   sunday,
	}
}

// GetCurrentBusinessWeekToDate returns the current business week from Monday to today (not including future days)
func GetCurrentBusinessWeekToDate() BusinessWeek {
	now := time.Now()
	week := GetBusinessWeekForDate(now)

	// If current time is before the week's end, use current time as end
	if now.Before(week.End) {
		week.End = now
	}

	return week
}

// GetPreviousBusinessWeek returns the business week prior to the given date
func GetPreviousBusinessWeek(date time.Time) BusinessWeek {
	// Go back 7 days to get into the previous week
	previousWeekDate := date.AddDate(0, 0, -7)
	return GetBusinessWeekForDate(previousWeekDate)
}

// GetBusinessWeeksBack returns N business weeks prior to the current week
// Returns weeks in chronological order (oldest first)
func GetBusinessWeeksBack(n int) []BusinessWeek {
	weeks := make([]BusinessWeek, n)
	currentDate := time.Now()

	for i := n - 1; i >= 0; i-- {
		weeksBack := i + 1
		dateInWeek := currentDate.AddDate(0, 0, -7*weeksBack)
		weeks[n-1-i] = GetBusinessWeekForDate(dateInWeek)
	}

	return weeks
}

// IsInBusinessWeek checks if a given time falls within the business week
func (bw BusinessWeek) IsInBusinessWeek(t time.Time) bool {
	return (t.Equal(bw.Start) || t.After(bw.Start)) && (t.Equal(bw.End) || t.Before(bw.End))
}

// String returns a human-readable representation of the business week
func (bw BusinessWeek) String() string {
	return bw.Start.Format("Jan 2") + " - " + bw.End.Format("Jan 2, 2006")
}

// WeekOf returns the Monday date of this business week
func (bw BusinessWeek) WeekOf() time.Time {
	return bw.Start
}