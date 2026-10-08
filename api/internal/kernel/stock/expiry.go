package stock

import "time"

const (
	hoursPerDay            = 24
	nearExpiryDays         = 7
	shortShelfLifeDays     = 7
	shortShelfLifeNearDays = 1
)

type Flag string

const (
	Safe       Flag = "SAFE"
	NearExpiry Flag = "NEAR_EXPIRY"
	Expired    Flag = "EXPIRED"
)

type Expiry struct {
	date  time.Time
	given bool
}

func NoExpiry() Expiry {
	return Expiry{}
}

func ExpiresOn(date time.Time) Expiry {
	return Expiry{date: calendarDay(date), given: true}
}

type ShelfLife struct {
	days  int
	given bool
}

func UnknownShelfLife() ShelfLife {
	return ShelfLife{}
}

func ShelfLifeOf(days int) ShelfLife {
	return ShelfLife{days: days, given: true}
}

func (s ShelfLife) nearExpiryThreshold() int {
	if s.given && s.days <= shortShelfLifeDays {
		return shortShelfLifeNearDays
	}
	return nearExpiryDays
}

func (e Expiry) FlagOn(today time.Time, shelfLife ShelfLife) Flag {
	if !e.given {
		return Safe
	}
	daysLeft := int(e.date.Sub(calendarDay(today)).Hours() / hoursPerDay)
	switch {
	case daysLeft < 0:
		return Expired
	case daysLeft <= shelfLife.nearExpiryThreshold():
		return NearExpiry
	}
	return Safe
}

func (e Expiry) compare(other Expiry) int {
	switch {
	case e.given && other.given:
		return e.date.Compare(other.date)
	case e.given:
		return -1
	case other.given:
		return 1
	}
	return 0
}

func calendarDay(at time.Time) time.Time {
	return time.Date(at.Year(), at.Month(), at.Day(), 0, 0, 0, 0, time.UTC)
}
