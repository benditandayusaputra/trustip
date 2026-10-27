package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/domain"
	"github.com/benditandayusaputra/trustip/api/internal/modules/auth/service"
)

const ownAvatarID = "0192f0c4-aa00-7e66-a1bb-4d2f3e5d6a92"

type fakeProfiles struct {
	account domain.Account
	changes []domain.AccountChange
}

func (f *fakeProfiles) GetAccount(_ context.Context, _ string) (domain.Account, error) {
	return f.account, nil
}

func (f *fakeProfiles) UpdateProfile(_ context.Context, change domain.AccountChange) error {
	if change.Profile.AvatarFileID != "" && change.Profile.AvatarFileID != ownAvatarID {
		return domain.ErrUnknownAvatar
	}
	f.changes = append(f.changes, change)
	f.account.FullName = change.Profile.FullName
	f.account.WhatsApp = change.Profile.WhatsApp.String()
	f.account.AvatarFileID = change.Profile.AvatarFileID
	return nil
}

func signedInSari() actor.Actor {
	return actor.Actor{UserID: knownUserID, SessionFamily: "family-1", Membership: storeOwnerMembership()}
}

func newUpdateAccount(profiles *fakeProfiles, trail *fakeAudit) *service.UpdateAccount {
	return service.NewUpdateAccount(service.UpdateAccountDeps{
		Accounts: profiles, Audit: trail, Tx: &inlineTx{}, Clock: fixedClock{now: testNow()},
	})
}

func TestUpdateAccount_SavesNameWhatsAppAndAvatarThenAudits(t *testing.T) {
	profiles := &fakeProfiles{account: domain.Account{ID: knownUserID, Email: "sari@contoh.id"}}
	trail := &fakeAudit{}
	form := domain.ProfileForm{FullName: "Sari W", WhatsApp: "0812 3456 7890", AvatarFileID: ownAvatarID}

	account, err := newUpdateAccount(profiles, trail).Execute(t.Context(), signedInSari(), form)
	if err != nil {
		t.Fatalf("update account: %v", err)
	}

	if account.FullName != "Sari W" || account.WhatsApp != "6281234567890" || account.AvatarFileID != ownAvatarID ||
		account.Membership != storeOwnerMembership() {
		t.Fatalf("account = %+v", account)
	}
	change := profiles.changes[0]
	if change.UserID != knownUserID || change.OrganizationID != storeOwnerMembership().OrganizationID {
		t.Fatalf("change = %+v", change)
	}
	if len(trail.entries) != 1 || trail.entries[0].Action != audit.AccountUpdated {
		t.Fatalf("audit = %+v", trail.entries)
	}
}

func TestUpdateAccount_RejectsAnAvatarThatIsNotTheUsersOwn(t *testing.T) {
	profiles := &fakeProfiles{}
	trail := &fakeAudit{}
	form := domain.ProfileForm{FullName: "Sari W", WhatsApp: "081234567890", AvatarFileID: "someone-else"}

	_, err := newUpdateAccount(profiles, trail).Execute(t.Context(), signedInSari(), form)

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "avatar_file_id" || len(trail.entries) != 0 {
		t.Fatalf("err = %v, audit = %+v", err, trail.entries)
	}
}

func TestUpdateAccount_InvalidFormNeverReachesTheStore(t *testing.T) {
	profiles := &fakeProfiles{}

	_, err := newUpdateAccount(profiles, &fakeAudit{}).Execute(t.Context(), signedInSari(), domain.ProfileForm{})

	if errorCode(err) != apperror.CodeValidationFailed || len(profiles.changes) != 0 {
		t.Fatalf("err = %v, changes = %d", err, len(profiles.changes))
	}
}

type fakeOwners struct {
	owner domain.PasswordOwner
	saved []domain.StoredPassword
}

func (f *fakeOwners) LockPasswordOwner(_ context.Context, _ string) (domain.PasswordOwner, error) {
	return f.owner, nil
}

func (f *fakeOwners) SavePassword(_ context.Context, _ string, password domain.StoredPassword) error {
	f.saved = append(f.saved, password)
	return nil
}

type fakeOtherSessions struct {
	kept    string
	revoked []string
}

func (f *fakeOtherSessions) RevokeOtherSessions(
	_ context.Context, _ string, keepFamily string, _ time.Time,
) ([]string, error) {
	f.kept = keepFamily
	return f.revoked, nil
}

type passwordFixture struct {
	owners   *fakeOwners
	sessions *fakeOtherSessions
	access   *fakeAccessRevoker
	audit    *fakeAudit
	useCase  *service.ChangePassword
}

func newPasswordFixture() passwordFixture {
	fixture := passwordFixture{
		owners: &fakeOwners{owner: domain.PasswordOwner{
			FullName: "Sari Wulandari", Email: "sari@contoh.id",
			Password: domain.StoredPassword{Hash: "$argon2id$Lama-sandi-2026", PepperVersion: "v1"},
		}},
		sessions: &fakeOtherSessions{revoked: []string{"family-2", "family-3"}},
		access:   &fakeAccessRevoker{until: map[string]time.Time{}},
		audit:    &fakeAudit{},
	}
	fixture.useCase = service.NewChangePassword(service.ChangePasswordDeps{
		Owners: fixture.owners, Sessions: fixture.sessions, Passwords: &fakePasswords{},
		Policy: domain.NewPasswordPolicy(), Access: fixture.access, Audit: fixture.audit,
		Tx: &inlineTx{}, Clock: fixedClock{now: testNow()},
	})
	return fixture
}

func TestChangePassword_SavesTheNewHashAndEndsEveryOtherSession(t *testing.T) {
	fixture := newPasswordFixture()
	form := domain.PasswordChangeForm{Current: "Lama-sandi-2026", New: acceptableInput}

	if err := fixture.useCase.Execute(t.Context(), signedInSari(), form); err != nil {
		t.Fatalf("change password: %v", err)
	}

	if len(fixture.owners.saved) != 1 || fixture.owners.saved[0].Hash != "$argon2id$new-"+acceptableInput {
		t.Fatalf("saved = %+v", fixture.owners.saved)
	}
	if fixture.sessions.kept != "family-1" || len(fixture.access.until) != 2 {
		t.Fatalf("kept = %s, revoked access = %v", fixture.sessions.kept, fixture.access.until)
	}
	if len(fixture.audit.entries) != 1 || fixture.audit.entries[0].Action != audit.PasswordChanged {
		t.Fatalf("audit = %+v", fixture.audit.entries)
	}
}

func TestChangePassword_WrongCurrentPasswordChangesNothing(t *testing.T) {
	fixture := newPasswordFixture()
	form := domain.PasswordChangeForm{Current: "Bukan-sandinya-1", New: acceptableInput}

	err := fixture.useCase.Execute(t.Context(), signedInSari(), form)

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Code != string(apperror.FieldWrongCurrent) {
		t.Fatalf("err = %v", err)
	}
	if len(fixture.owners.saved) != 0 || len(fixture.access.until) != 0 || len(fixture.audit.entries) != 0 {
		t.Fatalf("saved = %v, access = %v, audit = %v",
			fixture.owners.saved, fixture.access.until, fixture.audit.entries)
	}
}

func TestChangePassword_WeakNewPasswordIsRejectedBeforeHashing(t *testing.T) {
	fixture := newPasswordFixture()
	form := domain.PasswordChangeForm{Current: "Lama-sandi-2026", New: "password123"}

	err := fixture.useCase.Execute(t.Context(), signedInSari(), form)

	fields := apperror.From(err).Fields()
	if len(fields) != 1 || fields[0].Field != "new_password" || len(fixture.owners.saved) != 0 {
		t.Fatalf("err = %v, saved = %v", err, fixture.owners.saved)
	}
}
