package croner

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ParseCrontab parses a crontab expression and returns the number of seconds between executions
func parseCrontab(cronExpr string) (int64, error) {
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 {
		return 0, fmt.Errorf("invalid crontab expression")
	}

	minutes, err := parseField(fields[0], 0, 59)
	if err != nil {
		return 0, err
	}

	hours, err := parseField(fields[1], 0, 23)
	if err != nil {
		return 0, err
	}

	daysOfMonth, err := parseField(fields[2], 1, 31)
	if err != nil {
		return 0, err
	}

	months, err := parseField(fields[3], 1, 12)
	if err != nil {
		return 0, err
	}

	daysOfWeek, err := parseField(fields[4], 0, 6)
	if err != nil {
		return 0, err
	}

	// Calculate the number of seconds until the next execution
	now := time.Now()
	nextExec := now.Add(1 * time.Second)
	for {
		if matches(nextExec, minutes, hours, daysOfMonth, months, daysOfWeek) {
			break
		}
		nextExec = nextExec.Add(1 * time.Second)
	}
	return int64(nextExec.Sub(now).Seconds()), nil

}

func parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		values := make([]int, max-min+1)
		for i := min; i <= max; i++ {
			values[i-min] = i
		}
		return values, nil
	}

	parts := strings.Split(field, ",")
	var values []int
	for _, part := range parts {
		if strings.Contains(part, "-") {
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range in field: %s", field)
			}
			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				values = append(values, i)
			}
		} else {
			value, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
	}

	return values, nil
}

func matches(t time.Time, minutes, hours, daysOfMonth, months, daysOfWeek []int) bool {
	if !contains(minutes, t.Minute()) ||
		!contains(hours, t.Hour()) ||
		!contains(daysOfMonth, t.Day()) ||
		!contains(months, int(t.Month())) ||
		!contains(daysOfWeek, int(t.Weekday())) {
		return false
	}
	return true
}

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
