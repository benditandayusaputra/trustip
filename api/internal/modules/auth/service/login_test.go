package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

const knownUserID = "0192f0c4-7b1e-7c3a-9d2e-5f6a7b8c9d0e"

type fakeCredentials struct {
	stored      map[string]domain.Credentials
	savedPass   []domain.StoredPassword
	lockedCalls int
}

func (f *fakeCredentials) FindCredentials(_ context.Context, email string) (domain.Credentials, bool, error) {
	credentials, ok := f.stored[email]
	return credentials, ok, nil
}

func (f *fakeCredentials) LockLoginAttempts(_ context.Context, userID string) (domain.LoginAttempts, error) {
	f.lockedCalls++
	for _, credentials := range f.stored {
		if credentials.UserID == userID {
			return credentials.Attempts, nil
		}
	}
	return domain.LoginAttempts{}, errors.New("missing user")
}

func (f *fakeCredentials) SaveLoginAttempts(_ context.Context, userID string, attempts domain.LoginAttempts) error {
	for email, credentials := range f.stored {
		if credentials.UserID == userID {
			credentials.Attempts = attempts
			f.stored[email] = credentials
		}
	}
	return nil
}

func (f *fakeCredentials) SavePassword(_ context.Context, _ string, password domain.StoredPassword) error {
	f.savedPass = append(f.savedPass, password)
	return nil
}

type fakePasswords struct {
	needsRehash bool
	verified    int
	absent      int
}

func (f *fakePasswords) Hash(password string) (domain.StoredPassword, error) {
	return domain.StoredPassword{Hash: "$argon2id$new-" + password, PepperVersion: "v2"}, nil
}

func (f *fakePasswords) Verify(stored domain.StoredPassword, password string) (domain.PasswordCheck, error) {
	f.verified++
	matches := stored.Hash == "$argon2id$"+password
	return domain.PasswordCheck{Matches: matches, NeedsRehash: matches && f.needsRehash}, nil
}

func (f *fakePasswords) VerifyAbsent(string) error {
	f.absent++
	return nil
}

type fakeSessions struct{ created []domain.Session }

func (f *fakeSessions) CreateSession(_ context.Context, session domain.Session) error {
	f.created = append(f.created, session)
	return nil
}

type fakeTokens struct{ signed []domain.AccessClaims }

func (f *fakeTokens) NewRefreshToken(familyID string) (domain.RefreshToken, error) {
	if familyID == "" {
		return domain.RefreshToken{}, errors.New("refresh token needs its family")
	}
	return domain.RefreshToken{Plain: "refresh-plain", Hash: []byte("refresh-hash")}, nil
}

func (f *fakeTokens) SuccessorOf(previous string) domain.RefreshToken {
	return domain.RefreshToken{Plain: previous + "+next", Hash: f.HashRefreshToken(previous + "+next")}
}

func (f *fakeTokens) HashRefreshToken(plain string) []byte {
	return []byte("hash-of-" + plain)
}

func (f *fakeTokens) SignAccessToken(claims domain.AccessClaims) (string, error) {
	f.signed = append(f.signed, claims)
	return "signed-access", nil
}

type fakeMemberships struct{ byUser map[string]actor.Membership }

func (f *fakeMemberships) FindMembership(_ context.Context, userID string) (actor.Membership, bool, error) {
	membership, ok := f.byUser[userID]
	return membership, ok, nil
}

func storeOwnerMembership() actor.Membership {
	return actor.Membership{
		OrganizationID:   "0192f0c4-9c22-7e66-a1bb-4d2f3e5d6a7c",
		OrganizationType: actor.Store,
		Role:             actor.Owner,
	}
}

type inlineTx struct{ count int }

func (tx *inlineTx) WithinTx(ctx context.Context, work func(context.Context) error) error {
	tx.count++
	return work(ctx)
}

type sequenceIDs struct{ next int }

func (s *sequenceIDs) NewID() (string, error) {
	s.next++
	return []string{"", "session-1", "family-1", "session-2", "family-2"}[s.next], nil
}

type loginFixture struct {
	captcha     *fakeCaptcha
	credentials *fakeCredentials
	passwords   *fakePasswords
	sessions    *fakeSessions
	tokens      *fakeTokens
	memberships *fakeMemberships
	useCase     *service.LoginUser
}

func newLoginFixture(attempts domain.LoginAttempts) loginFixture {
	fixture := loginFixture{
		captcha: &fakeCaptcha{},
		credentials: &fakeCredentials{stored: map[string]domain.Credentials{"sari@contoh.id": {
			UserID:   knownUserID,
			Email:    "sari@contoh.id",
			FullName: "Sari Wulandari",
			Password: domain.StoredPassword{Hash: "$argon2id$" + acceptableInput, PepperVersion: "v1"},
			Attempts: attempts,
		}}},
		passwords:   &fakePasswords{},
		sessions:    &fakeSessions{},
		tokens:      &fakeTokens{},
		memberships: &fakeMemberships{byUser: map[string]actor.Membership{}},
	}
	fixture.useCase = service.NewLoginUser(service.LoginDeps{
		Captcha:     fixture.captcha,
		Passwords:   fixture.passwords,
		Credentials: fixture.credentials,
		Sessions:    fixture.sessions,
		Tokens:      fixture.tokens,
		Memberships: fixture.memberships,
		Tx:          &inlineTx{},
		Clock:       fixedClock{now: testNow()},
		IDs:         &sequenceIDs{},
	})
	return fixture
}

func loginCommand(email, password string) service.LoginCommand {
	return service.LoginCommand{
		Email: email, Password: password, CaptchaToken: "widget-answer", ClientIP: "203.0.113.7",
	}
}

func errorCode(err error) apperror.Code {
	return apperror.From(err).Code()
}

func TestLogin_BRAUTH05_InvalidCaptchaIsRejectedWithoutCheckingThePassword(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})
	fixture.captcha.err = apperror.New(apperror.CodeCaptchaFailed)

	_, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput))

	if errorCode(err) != apperror.CodeCaptchaFailed || fixture.passwords.verified != 0 {
		t.Fatalf("err = %v, verified = %d", err, fixture.passwords.verified)
	}
}

func TestLogin_RSEC06_UnknownEmailStillSpendsOneHash(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})

	_, err := fixture.useCase.Execute(t.Context(), loginCommand("asing@contoh.id", acceptableInput))

	if errorCode(err) != apperror.CodeLoginFailed || fixture.passwords.absent != 1 {
		t.Fatalf("err = %v, absent = %d", err, fixture.passwords.absent)
	}
}

func TestLogin_RSEC08_WrongPasswordGetsTheSameMessageAsUnknownEmail(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})

	_, wrongPassword := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", "Salah#2026x"))
	_, unknownEmail := fixture.useCase.Execute(t.Context(), loginCommand("asing@contoh.id", acceptableInput))

	if apperror.From(wrongPassword).Message() != apperror.From(unknownEmail).Message() {
		t.Fatalf("messages differ: %q, %q", wrongPassword, unknownEmail)
	}
}

func TestLogin_BRAUTH01_WrongPasswordIsCountedUnderARowLock(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{FailedCount: 2})

	_, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", "Salah#2026x"))

	attempts := fixture.credentials.stored["sari@contoh.id"].Attempts
	if errorCode(err) != apperror.CodeLoginFailed || attempts.FailedCount != 3 ||
		fixture.credentials.lockedCalls != 1 {
		t.Fatalf("err = %v, attempts = %+v", err, attempts)
	}
}

func TestLogin_BRAUTH01_FifthWrongPasswordLocksTheAccount(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{FailedCount: 4})

	_, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", "Salah#2026x"))

	appErr := apperror.From(err)
	if appErr.Code() != apperror.CodeAccountLocked || appErr.RetryAfter() != 900 {
		t.Fatalf("err = %v, retry after = %d", err, appErr.RetryAfter())
	}
}

func TestLogin_RSEC45_LockedAccountRefusesEvenTheRightPassword(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{LockedUntil: testNow().Add(10 * time.Minute)})

	_, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput))

	sessions := len(fixture.sessions.created)
	if errorCode(err) != apperror.CodeAccountLocked || fixture.passwords.verified != 0 || sessions != 0 {
		t.Fatalf("err = %v, verified = %d, sessions = %d", err, fixture.passwords.verified, sessions)
	}
}

func TestLogin_RSEC19_SuccessStartsANewSessionFamily(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{FailedCount: 3})

	result, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	session := fixture.sessions.created[0]
	if session.FamilyID != "family-1" || session.UserID != knownUserID || string(session.TokenHash) != "refresh-hash" {
		t.Fatalf("session = %+v", session)
	}
	if result.Tokens.Access != "signed-access" || result.Tokens.Refresh != "refresh-plain" ||
		!result.Tokens.RefreshExpires.Equal(session.ExpiresAt) || result.Account.FullName != "Sari Wulandari" {
		t.Fatalf("result = %+v", result)
	}
	if fixture.credentials.stored["sari@contoh.id"].Attempts.FailedCount != 0 {
		t.Fatal("a successful login must clear the failure streak")
	}
}

func TestLogin_RSEC16_AccessTokenNamesUserAndSessionFamily(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})

	result, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput))
	if err != nil {
		t.Fatalf("execute: %v", err)
	}

	claims := fixture.tokens.signed[0]
	if claims.Subject != knownUserID || claims.SessionFamily != "family-1" ||
		!result.Tokens.AccessExpires.Equal(claims.ExpiresAt) || claims.Membership != (actor.Membership{}) {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestLogin_RSEC16_AccessTokenCarriesTheMembership(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})
	fixture.memberships.byUser[knownUserID] = storeOwnerMembership()

	if _, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput)); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if claims := fixture.tokens.signed[0]; claims.Membership != storeOwnerMembership() {
		t.Fatalf("claims = %+v", claims)
	}
}

func TestLogin_FRAUTH05_SignedInAccountTellsWhetherOnboardingIsDone(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})
	fixture.memberships.byUser[knownUserID] = storeOwnerMembership()

	result, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput))

	if err != nil || result.Account.Membership != storeOwnerMembership() {
		t.Fatalf("account = %+v, err = %v", result.Account, err)
	}
}

func TestLogin_RSEC05_OutdatedHashIsReplacedAfterSuccess(t *testing.T) {
	fixture := newLoginFixture(domain.LoginAttempts{})
	fixture.passwords.needsRehash = true

	if _, err := fixture.useCase.Execute(t.Context(), loginCommand("sari@contoh.id", acceptableInput)); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if len(fixture.credentials.savedPass) != 1 || fixture.credentials.savedPass[0].PepperVersion != "v2" {
		t.Fatalf("saved = %+v", fixture.credentials.savedPass)
	}
}
