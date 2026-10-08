package claimed

import "time"

type Time struct {
	at      time.Time
	isGiven bool
}

func At(at time.Time) Time {
	return Time{at: at, isGiven: true}
}

func Absent() Time {
	return Time{}
}

func (t Time) Or(fallback time.Time) time.Time {
	if !t.isGiven {
		return fallback
	}
	return t.at
}

func (t Time) Text() string {
	if !t.isGiven {
		return ""
	}
	return t.at.UTC().Format(time.RFC3339Nano)
}
