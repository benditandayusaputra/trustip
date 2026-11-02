package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/claimed"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	countdomain "github.com/benditandayusaputra/trustip/api/internal/modules/count/domain"
	countservice "github.com/benditandayusaputra/trustip/api/internal/modules/count/service"
	dropdomain "github.com/benditandayusaputra/trustip/api/internal/modules/drop/domain"
	dropservice "github.com/benditandayusaputra/trustip/api/internal/modules/drop/service"
	filedomain "github.com/benditandayusaputra/trustip/api/internal/modules/file/domain"
	settlementdomain "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
	settlementservice "github.com/benditandayusaputra/trustip/api/internal/modules/settlement/service"
)

const (
	keepPendingDays   = 1
	keepUnconfirmed   = 2
	jitterSpread      = 121
	jitterCentre      = 60
	jitterDayStep     = 37
	jitterProductStep = 11
	cashPayment       = "CASH"
	cashReceipt       = "bukti/nota-tunai"
)

func (s *demoSeeder) recordDrops(ctx context.Context, daysAgo int) error {
	for _, deal := range s.world.Partnerships {
		if !deal.Drops.fallsOn(daysAgo) || !s.isActive(deal, daysAgo) {
			continue
		}
		if err := s.recordDrop(ctx, deal); err != nil {
			return fmt.Errorf("drop for %s: %w", deal.Key, err)
		}
	}
	return nil
}

func (s *demoSeeder) dropItems(deal demoPartnership) []dropdomain.ItemInput {
	factor := permille
	if weekday := s.clock.Now().In(period.Jakarta()).Weekday(); weekday == time.Saturday || weekday == time.Sunday {
		factor = deal.Drops.WeekendPermille
	}
	items := make([]dropdomain.ItemInput, 0, len(deal.Items))
	for _, item := range deal.Items {
		if qty := item.Qty * factor / permille; qty > 0 {
			items = append(items, dropdomain.ItemInput{ProductID: s.cast.products[item.Product], Qty: qty})
		}
	}
	return items
}

func (s *demoSeeder) recordDrop(ctx context.Context, deal demoPartnership) error {
	items := s.dropItems(deal)
	if len(items) == 0 {
		return nil
	}
	consignor, store := s.partiesOf(deal)
	sheet, err := s.use.count.Sheet.Execute(ctx, store, s.cast.partnerships[deal.Key])
	if err != nil {
		return fmt.Errorf("read sheet: %w", err)
	}
	key, err := newDemoKey()
	if err != nil {
		return err
	}
	now := claimed.At(s.clock.Now())
	receipt, err := s.use.drop.Record.Execute(ctx, store, dropservice.RecordDropCommand{
		IdempotencyKey: key, Input: dropdomain.RecordInput{
			PartnershipID: sheet.PartnershipID, AgreementVersion: sheet.AgreementVersion, OccurredAt: now,
			ClientRecordedAt: now, Items: items,
		},
	})
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}
	s.drops = append(s.drops, demoReview{reviewer: consignor, id: receipt.View.ID})
	return nil
}

func (s *demoSeeder) confirmDrops(ctx context.Context, daysAgo int) error {
	pending := s.drops
	s.drops = nil
	if daysAgo <= keepPendingDays {
		return nil
	}
	for _, drop := range pending {
		if _, err := s.use.drop.Confirm.Execute(ctx, drop.reviewer, drop.id); err != nil {
			return fmt.Errorf("confirm drop %s: %w", drop.id, err)
		}
	}
	return nil
}

func (s *demoSeeder) recordCounts(ctx context.Context, daysAgo int) error {
	for _, deal := range s.world.Partnerships {
		if !deal.Counts.fallsOn(daysAgo) || !s.isActive(deal, daysAgo) {
			continue
		}
		if err := s.recordCount(ctx, deal, daysAgo); err != nil {
			return fmt.Errorf("count for %s: %w", deal.Key, err)
		}
	}
	return nil
}

func sellPermille(routine demoRoutine, daysAgo, productIndex int) int {
	jitter := (daysAgo*jitterDayStep+productIndex*jitterProductStep)%jitterSpread - jitterCentre
	return min(permille, max(0, routine.SellPermille+jitter))
}

func countLines(sheet countdomain.Sheet, routine demoRoutine, daysAgo int) []countdomain.RequestLine {
	lines := make([]countdomain.RequestLine, 0, len(sheet.Products))
	for index, product := range sheet.Products {
		recorded := 0
		for _, batch := range product.Batches {
			recorded += batch.Stock.Balance
		}
		if recorded == 0 {
			continue
		}
		left := recorded - recorded*sellPermille(routine, daysAgo, index)/permille
		line := countdomain.RequestLine{ProductID: product.ProductID, RecordedQty: recorded, CountedQty: left}
		if routine.ReturnRest && left > 0 {
			line.ReturnedQty, line.ReturnReason = left, routine.Reason
		}
		lines = append(lines, line)
	}
	return lines
}

func (s *demoSeeder) recordCount(ctx context.Context, deal demoPartnership, daysAgo int) error {
	consignor, store := s.partiesOf(deal)
	sheet, err := s.use.count.Sheet.Execute(ctx, store, s.cast.partnerships[deal.Key])
	if err != nil {
		return fmt.Errorf("read sheet: %w", err)
	}
	lines := countLines(sheet, deal.Counts, daysAgo)
	if len(lines) == 0 {
		return nil
	}
	key, err := newDemoKey()
	if err != nil {
		return err
	}
	now := claimed.At(s.clock.Now())
	receipt, err := s.use.count.Record.Execute(ctx, store, countservice.RecordCountCommand{
		IdempotencyKey: key, Input: countdomain.RecordInput{
			PartnershipID: sheet.PartnershipID, AgreementVersion: sheet.AgreementVersion, CountedAt: now,
			ClientRecordedAt: now, Occasion: string(countdomain.StoreClosing), Lines: lines,
		},
	})
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}
	s.counts = append(s.counts, demoReview{reviewer: consignor, id: receipt.View.ID})
	return nil
}

func (s *demoSeeder) confirmCounts(ctx context.Context, daysAgo int) error {
	pending := s.counts
	s.counts = nil
	if daysAgo <= keepUnconfirmed {
		return nil
	}
	for _, count := range pending {
		if _, err := s.use.count.Confirm.Execute(ctx, count.reviewer, count.id); err != nil {
			return fmt.Errorf("confirm count %s: %w", count.id, err)
		}
	}
	return nil
}

func (s *demoSeeder) daysAgoOf(at time.Time) int {
	local := at.In(period.Jakarta())
	day := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, period.Jakarta())
	return int(s.today.Sub(day) / (hoursPerDay * time.Hour))
}

func (s *demoSeeder) payPartners(ctx context.Context, daysAgo int) error {
	for _, deal := range s.world.Partnerships {
		if err := s.payPartner(ctx, deal, daysAgo); err != nil {
			return fmt.Errorf("pay %s: %w", deal.Key, err)
		}
	}
	return nil
}

func (s *demoSeeder) payPartner(ctx context.Context, deal demoPartnership, daysAgo int) error {
	partnershipID, isFormed := s.cast.partnerships[deal.Key]
	if !isFormed {
		return nil
	}
	_, store := s.partiesOf(deal)
	unpaid, err := s.queries(ctx).ListUnpaidSettlements(ctx, partnershipID)
	if err != nil {
		return fmt.Errorf("list unpaid settlements: %w", err)
	}
	for _, settlement := range unpaid {
		if s.daysAgoOf(settlement.IssuedAt)-daysAgo < deal.PayAfterDays {
			continue
		}
		key, err := newDemoKey()
		if err != nil {
			return err
		}
		proofID, err := s.photos.put(ctx, store, filedomain.PaymentProof, cashReceipt)
		if err != nil {
			return err
		}
		_, err = s.use.settlement.Payments.Record.Execute(ctx, store, settlementservice.RecordPaymentCommand{
			SettlementID: settlement.ID, IdempotencyKey: key, Input: settlementdomain.PaymentInput{
				Amount: settlement.ConsignorShareAmount, Method: cashPayment,
				PaidOn: s.clock.Now().In(period.Jakarta()).Format(time.DateOnly), ProofFileID: proofID,
			},
		})
		if err != nil {
			return fmt.Errorf("record payment for %s: %w", settlement.ID, err)
		}
	}
	return nil
}

func (s *demoSeeder) confirmPayments(ctx context.Context, daysAgo int) error {
	for _, deal := range s.world.Partnerships {
		if err := s.confirmPartnerPayments(ctx, deal, daysAgo); err != nil {
			return fmt.Errorf("confirm payments of %s: %w", deal.Key, err)
		}
	}
	return nil
}

func (s *demoSeeder) confirmPartnerPayments(ctx context.Context, deal demoPartnership, daysAgo int) error {
	partnershipID, isFormed := s.cast.partnerships[deal.Key]
	if !isFormed {
		return nil
	}
	consignor, _ := s.partiesOf(deal)
	pending, err := s.queries(ctx).ListPendingPayments(ctx, partnershipID)
	if err != nil {
		return fmt.Errorf("list pending payments: %w", err)
	}
	for _, payment := range pending {
		if s.daysAgoOf(payment.IssuedAt)-daysAgo < deal.PayAfterDays+deal.ConfirmAfterDays {
			continue
		}
		if _, err := s.use.settlement.Payments.Confirm.Execute(ctx, consignor, payment.ID); err != nil {
			return fmt.Errorf("confirm payment %s: %w", payment.ID, err)
		}
	}
	return nil
}

func (s *demoSeeder) finish(ctx context.Context) error {
	s.moveTo(s.now)
	return s.use.jobs.summarize(ctx, s.clock.Now())
}
