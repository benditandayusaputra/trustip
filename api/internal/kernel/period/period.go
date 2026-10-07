package period

import "time"

const (
	jakartaOffsetSeconds = 7 * 60 * 60
	monthKeyLayout       = "200601"
)

func Jakarta() *time.Location {
	return time.FixedZone("WIB", jakartaOffsetSeconds)
}

func MonthKey(at time.Time) string {
	return at.In(Jakarta()).Format(monthKeyLayout)
}
