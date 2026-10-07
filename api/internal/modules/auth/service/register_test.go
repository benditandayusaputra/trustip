package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

const acceptableInput = "Kp7#mQz2vX"

type fakeCaptcha struct {
	err      error
	verified []service.CaptchaChallenge
}

func (f *fakeCaptcha) Verify(_ context.Context, challenge service.CaptchaChallenge) error {
	f.verified = append(f.verified, challenge)
	return f.err
}

type fakeHasher struct {
	err    error
	hashed []string
}

func (f *fakeHasher) Hash(password string) (domain.StoredPassword, error) {
	f.hashed = append(f.hashed, password)
	return domain.StoredPassword{Hash: "$argon2id$hash-of-" + password, PepperVersion: "v1"}, f.err
}

type fakeUsers struct {
	err     error
	created []domain.User
}

func (f *fakeUsers) CreateUser(_ context.Context, user domain.User) error {
	if f.err != nil {
		return f.err
	}
	f.created = append(f.created, user)
	return nil
}

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

type fixedIDs struct{}

func (fixedIDs) NewID() (string, error) { return "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e", nil }

type registerFixture struct {
	captcha *fakeCaptcha
	hasher  *fakeHasher
	users   *fakeUsers
	useCase *service.RegisterUser
}

func testNow() time.Time {
	return time.Date(2026, 10, 19, 8, 0, 0, 0, time.UTC)
}

func newRegisterFixture() registerFixture {
	fixture := registerFixture{captcha: &fakeCaptcha{}, hasher: &fakeHasher{}, users: &fakeUsers{}}
	fixture.useCase = service.NewRegisterUser(service.RegisterDeps{
		Captcha: fixture.captcha,
		Hasher:  fixture.hasher,
		Users:   fixture.users,
		Policy:  domain.NewPasswordPolicy(),
		Clock:   fixedClock{now: testNow()},
		IDs:     fixedIDs{},
	})
	return fixture
}

func registerCommand() service.RegisterCommand {
	return service.RegisterCommand{
		Form: domain.RegistrationForm{
			FullName:        "Sari Wulandari",
			Email:           "sari@contoh.id",
			Password:        acceptableInput,
			WhatsApp:        "081234567890",
			PrivacyAccepted: true,
		},
		CaptchaToken: "widget-answer",
		ClientIP:     "203.0.113.7",
	}
}

func TestRegister_BRAUTH05_FailedCaptchaStopsBeforeAnythingElse(t *testing.T) {
	fixture := newRegisterFixture()
	fixture.captcha.err = apperror.New(apperror.CodeCaptchaFailed)

	_, err := fixture.useCase.Execute(t.Context(), registerCommand())

	if apperror.From(err).Code() != apperror.CodeCaptchaFailed {
		t.Fatalf("err = %v", err)
	}
	if len(fixture.hasher.hashed) != 0 || len(fixture.users.created) != 0 {
		t.Fatalf("hashed %d, created %d", len(fixture.hasher.hashed), len(fixture.users.created))
	}
}

func TestRegister_InvalidForm_IsRejectedWithoutHashing(t *testing.T) {
	fixture := newRegisterFixture()
	command := registerCommand()
	command.Form.Email = "bukan-email"

	_, err := fixture.useCase.Execute(t.Context(), command)

	if apperror.From(err).Code() != apperror.CodeValidationFailed || len(fixture.hasher.hashed) != 0 {
		t.Fatalf("err = %v, hashed = %d", err, len(fixture.hasher.hashed))
	}
}

func TestRegister_EmailAlreadyRegistered_TellsTheUserToSignIn(t *testing.T) {
	fixture := newRegisterFixture()
	fixture.users.err = domain.ErrEmailTaken

	_, err := fixture.useCase.Execute(t.Context(), registerCommand())

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "email" || fields[0].Message != "Email sudah terdaftar. Silakan masuk." {
		t.Fatalf("err = %v, fields = %+v", err, fields)
	}
}

func TestRegister_RSEC07_BusyHasherIsUnavailable(t *testing.T) {
	fixture := newRegisterFixture()
	fixture.hasher.err = apperror.Unavailable(1)

	_, err := fixture.useCase.Execute(t.Context(), registerCommand())

	if apperror.From(err).Code() != apperror.CodeServiceUnavailable || len(fixture.users.created) != 0 {
		t.Fatalf("err = %v, created = %d", err, len(fixture.users.created))
	}
}

func TestRegister_StorageFailure_IsReturnedWithContext(t *testing.T) {
	fixture := newRegisterFixture()
	broken := errors.New("connection reset")
	fixture.users.err = broken

	_, err := fixture.useCase.Execute(t.Context(), registerCommand())

	if !errors.Is(err, broken) {
		t.Fatalf("err = %v", err)
	}
}

func TestRegister_ChecksTheSignupCaptchaForTheClientIP(t *testing.T) {
	fixture := newRegisterFixture()

	if _, err := fixture.useCase.Execute(t.Context(), registerCommand()); err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := service.CaptchaChallenge{Token: "widget-answer", Action: "signup", RemoteIP: "203.0.113.7"}
	if len(fixture.captcha.verified) != 1 || fixture.captcha.verified[0] != want {
		t.Fatalf("verified = %+v", fixture.captcha.verified)
	}
}

func TestRegister_FRAUTH01_StoresUserWithHashedPasswordAndConsentTime(t *testing.T) {
	fixture := newRegisterFixture()

	user, err := fixture.useCase.Execute(t.Context(), registerCommand())
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(fixture.users.created) != 1 || fixture.users.created[0] != user {
		t.Fatalf("created = %+v", fixture.users.created)
	}
	if user.Password.Hash != "$argon2id$hash-of-"+acceptableInput || !user.PrivacyAcceptedAt.Equal(testNow()) {
		t.Fatalf("user = %+v", user)
	}
	if user.WhatsApp.String() != "6281234567890" {
		t.Fatalf("whatsapp = %s", user.WhatsApp)
	}
}
