package domain

import (
	"math"
	"time"
)

const (
	MaxFailedLogins = 5
	LockDuration    = 15 * time.Minute
)

type LoginAttempts struct {
	FailedCount int
	LockedUntil time.Time
}

func (a LoginAttempts) IsLockedAt(now time.Time) bool {
	return now.Before(a.LockedUntil)
}

func (a LoginAttempts) AfterFailure(now time.Time) LoginAttempts {
	failedCount := a.FailedCount + 1
	if failedCount >= MaxFailedLogins {
		return LoginAttempts{LockedUntil: now.Add(LockDuration)}
	}
	return LoginAttempts{FailedCount: failedCount, LockedUntil: a.LockedUntil}
}

func (a LoginAttempts) AfterSuccess() LoginAttempts {
	return LoginAttempts{}
}

func (a LoginAttempts) SecondsUntilUnlock(now time.Time) int {
	return int(math.Ceil(a.LockedUntil.Sub(now).Seconds()))
}

type PasswordCheck struct {
	Matches     bool
	NeedsRehash bool
}

type Credentials struct {
	UserID   string
	Email    string
	FullName string
	Password StoredPassword
	Attempts LoginAttempts
}
