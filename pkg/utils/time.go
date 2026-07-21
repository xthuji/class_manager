package utils

import "time"

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

func GetToday() string {
	return time.Now().Format("2006-01-02")
}

func GetWeekStart() time.Time {
	now := time.Now()
	weekday := now.Weekday()
	if weekday == 0 {
		weekday = 7
	}
	return now.AddDate(0, 0, -int(weekday)+1)
}

func GetWeekEnd() time.Time {
	return GetWeekStart().AddDate(0, 0, 6)
}

func DaysOfWeek() []string {
	return []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
}

func Periods() []string {
	return []string{"上午", "下午", "晚上"}
}
