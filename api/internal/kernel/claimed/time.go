package claimed

import "time"

const ClockTolerance = 5 * time.Minute

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

func (t Time) SettledBy(receivedAt time.Time) time.Time {
	at := t.Or(receivedAt)
	if at.After(receivedAt.Add(ClockTolerance)) {
		return receivedAt
	}
	return at
}

func Parse(raw *string) (Time, bool) {
	if raw == nil {
		return Absent(), true
	}
	at, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		return Absent(), false
	}
	return At(at), true
}
