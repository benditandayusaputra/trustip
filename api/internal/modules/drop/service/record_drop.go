package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/audit"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/policy"
	"github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
)

type RecordDropCommand struct {
	IdempotencyKey string
	Input          domain.RecordInput
}

type DropReceipt struct {
	View     domain.SealedView
	IsReplay bool
}

type RecordDrop struct {
	deps   Deps
	reader *GetDrop
}

func NewRecordDrop(deps Deps) *RecordDrop {
	return &RecordDrop{deps: deps, reader: NewGetDrop(deps)}
}

func (uc *RecordDrop) Execute(ctx context.Context, signedIn actor.Actor, cmd RecordDropCommand) (DropReceipt, error) {
	if err := policy.Require(signedIn, policy.RecordDrops); err != nil {
		return DropReceipt{}, fmt.Errorf("check drop recorder: %w", err)
	}
	fingerprint, err := cmd.Input.Fingerprint()
	if err != nil {
		return DropReceipt{}, fmt.Errorf("fingerprint drop request: %w", err)
	}
	key := domain.IdempotencyKey{StoreID: signedIn.Membership.OrganizationID, Key: cmd.IdempotencyKey}
	var dropID string
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		dropID, err = uc.record(ctx, signedIn, cmd, fingerprint)
		return err
	})
	if errors.Is(err, domain.ErrDuplicateKey) {
		return uc.replay(ctx, signedIn, key, fingerprint)
	}
	if err != nil {
		return DropReceipt{}, fmt.Errorf("record drop on %s: %w", cmd.Input.PartnershipID, err)
	}
	return uc.receipt(ctx, signedIn, dropID, false)
}

func (uc *RecordDrop) record(
	ctx context.Context, signedIn actor.Actor, cmd RecordDropCommand, fingerprint []byte,
) (string, error) {
	storeID := signedIn.Membership.OrganizationID
	_, isStored, err := uc.deps.Drops.FindByKey(ctx, domain.IdempotencyKey{StoreID: storeID, Key: cmd.IdempotencyKey})
	if err != nil {
		return "", fmt.Errorf("find drop by key: %w", err)
	}
	if isStored {
		return "", domain.ErrDuplicateKey
	}
	setting, err := uc.deps.Drops.GetSetting(ctx, domain.SettingKey{
		StoreID: storeID, PartnershipID: cmd.Input.PartnershipID,
	})
	if err != nil {
		return "", fmt.Errorf("load partnership: %w", err)
	}
	now := uc.deps.Clock.Now()
	draft, err := setting.Check(cmd.Input, now)
	if err != nil {
		return "", fmt.Errorf("check drop: %w", err)
	}
	recording := domain.Recording{
		RecordedBy: signedIn.UserID, IdempotencyKey: cmd.IdempotencyKey, Fingerprint: fingerprint, At: now,
	}
	if recording.IDs, err = uc.newIDs(len(draft.Items)); err != nil {
		return "", err
	}
	if recording.Sequence, err = uc.deps.Drops.NextNumber(ctx, domain.SequenceOf(storeID, now)); err != nil {
		return "", fmt.Errorf("number drop: %w", err)
	}
	drop := draft.Record(recording)
	return drop.ID, uc.save(ctx, drop, setting.StoreName)
}

func (uc *RecordDrop) newIDs(lines int) (domain.RecordIDs, error) {
	ids := domain.RecordIDs{ItemIDs: make([]string, 0, lines), MovementIDs: make([]string, 0, lines)}
	var errs []error
	var err error
	ids.DropID, err = uc.deps.IDs.NewID()
	errs = append(errs, err)
	for range lines {
		itemID, itemErr := uc.deps.IDs.NewID()
		movementID, movementErr := uc.deps.IDs.NewID()
		ids.ItemIDs = append(ids.ItemIDs, itemID)
		ids.MovementIDs = append(ids.MovementIDs, movementID)
		errs = append(errs, itemErr, movementErr)
	}
	if err := errors.Join(errs...); err != nil {
		return domain.RecordIDs{}, fmt.Errorf("generate drop ids: %w", err)
	}
	return ids, nil
}

func (uc *RecordDrop) save(ctx context.Context, drop domain.Drop, storeName string) error {
	if err := uc.deps.Drops.CreateDrop(ctx, drop); err != nil {
		return fmt.Errorf("save drop: %w", err)
	}
	if err := uc.deps.Sealer.Seal(ctx, drop.SealRequest()); err != nil {
		return fmt.Errorf("seal drop %s: %w", drop.Number, err)
	}
	if err := uc.deps.Announcer.Announce(ctx, drop.Notice(storeName)); err != nil {
		return fmt.Errorf("announce drop %s: %w", drop.Number, err)
	}
	err := uc.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: drop.StoreID, ActorID: drop.RecordedBy, Action: audit.DropRecorded,
		EntityType: audit.DropEntity, EntityID: drop.ID, At: drop.ReceivedAt,
	})
	if err != nil {
		return fmt.Errorf("audit drop %s: %w", drop.Number, err)
	}
	return nil
}

func (uc *RecordDrop) replay(
	ctx context.Context, signedIn actor.Actor, key domain.IdempotencyKey, fingerprint []byte,
) (DropReceipt, error) {
	stored, isStored, err := uc.deps.Drops.FindByKey(ctx, key)
	if err != nil {
		return DropReceipt{}, fmt.Errorf("find replayed drop: %w", err)
	}
	if !isStored {
		return DropReceipt{}, fmt.Errorf("find replayed drop: %w", domain.ErrDuplicateKey)
	}
	if !bytes.Equal(stored.Fingerprint, fingerprint) {
		return DropReceipt{}, apperror.New(apperror.CodeIdempotencyKeyReused)
	}
	return uc.receipt(ctx, signedIn, stored.DropID, true)
}

func (uc *RecordDrop) receipt(
	ctx context.Context, signedIn actor.Actor, dropID string, isReplay bool,
) (DropReceipt, error) {
	view, err := uc.reader.read(ctx, signedIn, dropID)
	if err != nil {
		return DropReceipt{}, err
	}
	return DropReceipt{View: view, IsReplay: isReplay}, nil
}
