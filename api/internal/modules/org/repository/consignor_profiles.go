package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/benditandayusaputra/trustip/api/internal/db/sqlc"
	"github.com/benditandayusaputra/trustip/api/internal/modules/org/domain"
	"github.com/benditandayusaputra/trustip/api/internal/platform/database"
)

const paymentQRPurpose = "PAYMENT_QR"

func (r *Profiles) UpdateConsignorProfile(ctx context.Context, change domain.ConsignorProfileChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	profile := change.Profile
	updated, err := sqlc.New(r.pool.Querier(ctx)).UpdateConsignorProfile(writeCtx, sqlc.UpdateConsignorProfileParams{
		OrganizationID:       change.OrganizationID,
		HasDeliveryRadius:    profile.DeliveryRadius.Given,
		DeliveryRadiusKm:     int64(profile.DeliveryRadius.Kilometres),
		DeliveryDays:         orEmpty(profile.DeliveryDays),
		PreferredStoreTypes:  orEmpty(profile.PreferredStoreTypes),
		BusinessIDNo:         profile.Legal.BusinessIDNo,
		DistributionPermitNo: profile.Legal.DistributionPermitNo,
		HalalCertNo:          profile.Legal.HalalCertNo,
		CanProvideSample:     profile.CanProvideSample,
		UpdatedAt:            change.At,
	})
	if err != nil {
		return fmt.Errorf("update consignor profile of %s: %w", change.OrganizationID, err)
	}
	if updated == 0 {
		return fmt.Errorf("consignor profile of %s is missing: %w", change.OrganizationID, pgx.ErrNoRows)
	}
	return nil
}

func (r *Profiles) UpdatePaymentInfo(ctx context.Context, change domain.PaymentInfoChange) error {
	writeCtx, cancel := database.WriteContext(ctx)
	defer cancel()
	queries := sqlc.New(r.pool.Querier(ctx))
	if err := checkOwnFile(writeCtx, queries, ownFile{
		organizationID: change.OrganizationID, fileID: change.Payment.QRFileID, purpose: paymentQRPurpose,
	}); err != nil {
		return explainMissingFile(err, domain.ErrUnknownQR)
	}
	err := queries.UpdatePaymentInfo(writeCtx, sqlc.UpdatePaymentInfoParams{
		ID: change.OrganizationID, PaymentInfo: change.Payment.Info, PaymentQrFileID: change.Payment.QRFileID,
		UpdatedAt: change.At,
	})
	if err != nil {
		return fmt.Errorf("update payment info of %s: %w", change.OrganizationID, err)
	}
	return nil
}

func orEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func consignorSettings(
	ctx context.Context, queries *sqlc.Queries, organizationID string,
) (domain.ConsignorProfile, domain.PaymentSettings, error) {
	row, err := queries.GetConsignorProfile(ctx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConsignorProfile{}, domain.PaymentSettings{}, nil
	}
	if err != nil {
		return domain.ConsignorProfile{}, domain.PaymentSettings{}, fmt.Errorf("get consignor profile: %w", err)
	}
	payment, err := queries.GetPaymentInfo(ctx, organizationID)
	if err != nil {
		return domain.ConsignorProfile{}, domain.PaymentSettings{}, fmt.Errorf("get payment info: %w", err)
	}
	profile := domain.ConsignorProfile{
		DeliveryRadius:      domain.DeliveryRadius{Given: row.HasDeliveryRadius, Kilometres: int(row.DeliveryRadiusKm)},
		DeliveryDays:        row.DeliveryDays,
		PreferredStoreTypes: row.PreferredStoreTypes,
		Legal: domain.LegalNumbers{
			BusinessIDNo: row.BusinessIDNo, DistributionPermitNo: row.DistributionPermitNo,
			HalalCertNo: row.HalalCertNo,
		},
		CanProvideSample: row.CanProvideSample,
	}
	settings := domain.PaymentSettings{
		Info: payment.PaymentInfo, QR: domain.FileRef{ID: payment.QrFileID, StorageKey: payment.QrStorageKey},
	}
	return profile, settings, nil
}
