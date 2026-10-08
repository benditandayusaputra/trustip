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
	"github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
)

type RecordCountCommand struct {
	IdempotencyKey string
	Input          domain.RecordInput
}

type CountReceipt struct {
	View     domain.SealedView
	IsReplay bool
}

type RecordCount struct {
	deps   Deps
	reader *GetCount
}

func NewRecordCount(deps Deps) *RecordCount {
	return &RecordCount{deps: deps, reader: NewGetCount(deps)}
}

func (uc *RecordCount) Execute(
	ctx context.Context, signedIn actor.Actor, cmd RecordCountCommand,
) (CountReceipt, error) {
	if err := policy.Require(signedIn, policy.RecordCounts); err != nil {
		return CountReceipt{}, fmt.Errorf("check count recorder: %w", err)
	}
	fingerprint, err := cmd.Input.Fingerprint()
	if err != nil {
		return CountReceipt{}, fmt.Errorf("fingerprint count request: %w", err)
	}
	key := domain.IdempotencyKey{StoreID: signedIn.Membership.OrganizationID, Key: cmd.IdempotencyKey}
	var countID string
	err = uc.deps.Tx.WithinTx(ctx, func(ctx context.Context) error {
		countID, err = uc.record(ctx, signedIn, cmd, fingerprint)
		return err
	})
	if errors.Is(err, domain.ErrDuplicateKey) {
		return uc.replay(ctx, signedIn, key, fingerprint)
	}
	if err != nil {
		return CountReceipt{}, fmt.Errorf("record count on %s: %w", cmd.Input.PartnershipID, err)
	}
	return uc.receipt(ctx, signedIn, countID, false)
}

func (uc *RecordCount) record(
	ctx context.Context, signedIn actor.Actor, cmd RecordCountCommand, fingerprint []byte,
) (string, error) {
	storeID := signedIn.Membership.OrganizationID
	_, isStored, err := uc.deps.Counts.FindByKey(ctx, domain.IdempotencyKey{StoreID: storeID, Key: cmd.IdempotencyKey})
	if err != nil {
		return "", fmt.Errorf("find count by key: %w", err)
	}
	if isStored {
		return "", domain.ErrDuplicateKey
	}
	sheetKey := domain.SheetKey{StoreID: storeID, PartnershipID: cmd.Input.PartnershipID}
	sheet, err := uc.deps.Counts.LockSheet(ctx, sheetKey)
	if err != nil {
		return "", fmt.Errorf("lock count sheet: %w", err)
	}
	now := uc.deps.Clock.Now()
	draft, err := sheet.Check(cmd.Input, now)
	if err != nil {
		return "", fmt.Errorf("check count: %w", err)
	}
	recording := domain.Recording{
		RecordedBy: signedIn.UserID, IdempotencyKey: cmd.IdempotencyKey, Fingerprint: fingerprint, At: now,
	}
	if recording.IDs, err = uc.newIDs(draft.IDCount()); err != nil {
		return "", err
	}
	if recording.Sequence, err = uc.deps.Counts.NextNumber(ctx, domain.SequenceOf(storeID, now)); err != nil {
		return "", fmt.Errorf("number count: %w", err)
	}
	count := draft.Record(recording)
	return count.ID, uc.save(ctx, count, sheet.StoreName)
}

func (uc *RecordCount) newIDs(total int) ([]string, error) {
	ids := make([]string, 0, total)
	for range total {
		id, err := uc.deps.IDs.NewID()
		if err != nil {
			return nil, fmt.Errorf("generate count ids: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (uc *RecordCount) save(ctx context.Context, count domain.Count, storeName string) error {
	if err := uc.deps.Counts.CreateCount(ctx, count); err != nil {
		return fmt.Errorf("save count: %w", err)
	}
	if err := uc.deps.Sealer.Seal(ctx, count.SealRequest()); err != nil {
		return fmt.Errorf("seal count %s: %w", count.Number, err)
	}
	if err := uc.deps.Announcer.Announce(ctx, count.Notice(storeName)); err != nil {
		return fmt.Errorf("announce count %s: %w", count.Number, err)
	}
	err := uc.deps.Audit.Record(ctx, audit.Entry{
		OrganizationID: count.StoreID, ActorID: count.RecordedBy, Action: audit.CountRecorded,
		EntityType: audit.CountEntity, EntityID: count.ID, At: count.ReceivedAt,
	})
	if err != nil {
		return fmt.Errorf("audit count %s: %w", count.Number, err)
	}
	return nil
}

func (uc *RecordCount) replay(
	ctx context.Context, signedIn actor.Actor, key domain.IdempotencyKey, fingerprint []byte,
) (CountReceipt, error) {
	stored, isStored, err := uc.deps.Counts.FindByKey(ctx, key)
	if err != nil {
		return CountReceipt{}, fmt.Errorf("find replayed count: %w", err)
	}
	if !isStored {
		return CountReceipt{}, fmt.Errorf("find replayed count: %w", domain.ErrDuplicateKey)
	}
	if !bytes.Equal(stored.Fingerprint, fingerprint) {
		return CountReceipt{}, apperror.New(apperror.CodeIdempotencyKeyReused)
	}
	return uc.receipt(ctx, signedIn, stored.CountID, true)
}

func (uc *RecordCount) receipt(
	ctx context.Context, signedIn actor.Actor, countID string, isReplay bool,
) (CountReceipt, error) {
	view, err := uc.reader.read(ctx, signedIn, countID)
	if err != nil {
		return CountReceipt{}, err
	}
	return CountReceipt{View: view, IsReplay: isReplay}, nil
}
