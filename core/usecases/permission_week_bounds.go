package usecases

import (
	"time"
)

func WeekBoundsForDate(weekendDays []int, date time.Time) (time.Time, time.Time) {
	day := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	weekendSet := make(map[int]struct{}, len(weekendDays))
	for _, d := range weekendDays {
		if d >= 0 && d <= 6 {
			weekendSet[d] = struct{}{}
		}
	}

	if len(weekendSet) == 7 {
		return day.AddDate(0, 0, -6), day
	}

	start := day
	for i := 0; i < 7; i++ {
		candidate := day.AddDate(0, 0, -i)
		prev := candidate.AddDate(0, 0, -1)
		if _, currentIsWeekend := weekendSet[int(candidate.Weekday())]; currentIsWeekend {
			continue
		}
		if _, prevIsWeekend := weekendSet[int(prev.Weekday())]; prevIsWeekend || len(weekendSet) == 0 {
			start = candidate
			break
		}
		start = candidate
	}

	end := start.AddDate(0, 0, 6)
	return start, end
}
