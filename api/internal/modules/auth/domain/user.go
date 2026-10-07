package domain

import (
	"errors"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/contact"
)

var ErrEmailTaken = errors.New("email is already registered")

type Account struct {
	ID         string
	Email      string
	FullName   string
	Membership actor.Membership
}

type StoredPassword struct {
	Hash          string
	PepperVersion string
}

type User struct {
	ID                string
	Email             string
	FullName          string
	WhatsApp          contact.WhatsAppNumber
	Password          StoredPassword
	PrivacyAcceptedAt time.Time
}
