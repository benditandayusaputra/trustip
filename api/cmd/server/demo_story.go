package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/actor"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/apperror"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/period"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/kernel/terms"
	needdomain "github.com/benditandayusaputra/trustip/api/internal/modules/need/domain"
	partnershipservice "github.com/benditandayusaputra/trustip/api/internal/modules/partnership/service"
	proposaldomain "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/domain"
	proposalservice "github.com/benditandayusaputra/trustip/api/internal/modules/proposal/service"
)

type demoMoment struct {
	hour   int
	minute int
	run    func(ctx context.Context, daysAgo int) error
}

const (
	unreadDaysAgo   = 2
	dropHour        = 6
	morningHour     = 9
	halfHour        = 30
	proposalHour    = 10
	offerHour       = 11
	declineHour     = 12
	endingHour      = 13
	countHour       = 20
	countReviewHour = 21
	paymentHour     = 22
	unreadHour      = 23
)

func (s *demoSeeder) moments() []demoMoment {
	return []demoMoment{
		{hour: 0, minute: dailyMinute, run: s.runNightJobs},
		{hour: 0, minute: settlementMinute, run: s.issueSettlements},
		{hour: dropHour, run: s.recordDrops},
		{hour: morningHour, run: s.postNeeds},
		{hour: morningHour, minute: halfHour, run: s.confirmDrops},
		{hour: proposalHour, run: s.formPartnerships},
		{hour: offerHour, run: s.sendProposals},
		{hour: declineHour, run: s.declineProposals},
		{hour: endingHour, run: s.endPartnerships},
		{hour: countHour, run: s.recordCounts},
		{hour: countReviewHour, run: s.confirmCounts},
		{hour: countReviewHour, minute: halfHour, run: s.payPartners},
		{hour: paymentHour, run: s.confirmPayments},
		{hour: unreadHour, run: s.readOldNotifications},
	}
}

func (s *demoSeeder) playHistory(ctx context.Context) error {
	for daysAgo := s.world.SetupDaysAgo; daysAgo >= 1; daysAgo-- {
		for _, moment := range s.moments() {
			s.moveTo(s.dayAt(daysAgo, moment.hour).Add(time.Duration(moment.minute) * time.Minute))
			if err := moment.run(ctx, daysAgo); err != nil {
				return fmt.Errorf("demo day -%d at %02d.%02d: %w", daysAgo, moment.hour, moment.minute, err)
			}
		}
	}
	return nil
}

func (s *demoSeeder) runNightJobs(ctx context.Context, _ int) error {
	return s.use.jobs.daily(ctx, s.clock.Now())
}

func (s *demoSeeder) issueSettlements(ctx context.Context, _ int) error {
	day := s.clock.Now().In(period.Jakarta())
	if day.Weekday() == time.Monday {
		if err := s.use.jobs.issue(terms.ScheduleWeekly)(ctx, s.clock.Now()); err != nil {
			return err
		}
	}
	if day.Day() == 1 {
		return s.use.jobs.issue(terms.ScheduleMonthly)(ctx, s.clock.Now())
	}
	return nil
}

func (s *demoSeeder) postNeeds(ctx context.Context, daysAgo int) error {
	for _, need := range s.world.Needs {
		var err error
		if need.DaysAgo == daysAgo {
			err = s.postNeed(ctx, need)
		}
		if need.FulfilledDaysAgo == daysAgo && err == nil {
			_, err = s.use.need.Fulfill.Execute(ctx, s.cast.owners[need.Store], s.cast.needs[need.Key])
		}
		if err != nil {
			return fmt.Errorf("demo need %s: %w", need.Key, err)
		}
	}
	return nil
}

func (s *demoSeeder) postNeed(ctx context.Context, need demoNeed) error {
	categoryID, err := s.category(need.Category)
	if err != nil {
		return err
	}
	posted, err := s.use.need.Post.Execute(ctx, s.cast.owners[need.Store], needdomain.Form{
		Title: need.Title, CategoryID: categoryID, Description: need.Description,
		WeeklyQuantity: amountOf(need.WeeklyQty), MaxRetailPrice: amountOf(need.MaxRetailPrice),
		PreferredModel: need.PreferredModel, OfferedCommissionBPS: amountOf(need.CommissionBPS),
	})
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	s.cast.needs[need.Key] = posted.ID
	return nil
}

func amountOf(value int64) needdomain.Amount {
	return needdomain.Amount{Value: value, IsGiven: value > 0}
}

func (s *demoSeeder) revision(deal demoTerms, items []demoItem, message string) proposaldomain.RevisionInput {
	inputs := make([]proposaldomain.ItemInput, 0, len(items))
	for _, item := range items {
		price := pricing.TermsInput{
			Model: item.Model, RetailPrice: item.Retail, CommissionBPS: item.CommissionBPS, DepositPrice: item.Deposit,
		}
		inputs = append(inputs, proposaldomain.ItemInput{ProductID: s.cast.products[item.Product], Terms: price})
	}
	return proposaldomain.RevisionInput{Items: inputs, Message: message, Terms: terms.Input{
		SettlementSchedule: deal.Schedule, PaymentTermsDays: deal.PaymentTermsDays, CountRhythm: deal.Rhythm,
		Notes: deal.Notes,
	}}
}

func (s *demoSeeder) formPartnerships(ctx context.Context, daysAgo int) error {
	if daysAgo != s.world.SetupDaysAgo-1 {
		return nil
	}
	for _, deal := range s.world.Partnerships {
		consignor, store := s.cast.owners[deal.Consignor], s.cast.owners[deal.Store]
		sent, err := s.use.proposal.Propose.Execute(ctx, consignor, proposalservice.DirectCommand{
			TargetID: store.Membership.OrganizationID, NeedID: s.cast.needs[deal.Need],
			Input: s.revision(deal.demoTerms, deal.Items, deal.Notes),
		})
		if err != nil {
			return fmt.Errorf("propose demo partnership %s: %w", deal.Key, err)
		}
		accepted, err := s.use.proposal.Accept.Execute(ctx, store, sent.ID)
		if err != nil {
			return fmt.Errorf("accept demo partnership %s: %w", deal.Key, err)
		}
		s.cast.partnerships[deal.Key] = accepted.PartnershipID
	}
	return nil
}

func (s *demoSeeder) sendProposals(ctx context.Context, daysAgo int) error {
	for index, offer := range s.world.Proposals {
		if offer.DaysAgo != daysAgo {
			continue
		}
		sent, err := s.use.proposal.Propose.Execute(ctx, s.cast.owners[offer.From], proposalservice.DirectCommand{
			TargetID: s.cast.owners[offer.To].Membership.OrganizationID,
			Input:    s.revision(offer.demoTerms, offer.Items, offer.Message),
		})
		if err != nil {
			return fmt.Errorf("send demo proposal %d: %w", index, err)
		}
		s.cast.proposals[index] = sent.ID
	}
	return nil
}

func (s *demoSeeder) declineProposals(ctx context.Context, daysAgo int) error {
	for index, offer := range s.world.Proposals {
		if offer.Declined == nil || offer.Declined.DaysAgo != daysAgo {
			continue
		}
		_, err := s.use.proposal.Decline.Execute(ctx, s.cast.owners[offer.To], proposalservice.DeclineCommand{
			ProposalID: s.cast.proposals[index], Reason: offer.Declined.Reason,
		})
		if err != nil {
			return fmt.Errorf("decline demo proposal %d: %w", index, err)
		}
	}
	return nil
}

func (s *demoSeeder) endPartnerships(ctx context.Context, daysAgo int) error {
	for _, deal := range s.world.Partnerships {
		if deal.Ended == nil || deal.Ended.DaysAgo != daysAgo {
			continue
		}
		if err := s.endPartnership(ctx, deal, daysAgo); err != nil {
			return fmt.Errorf("end demo partnership %s: %w", deal.Key, err)
		}
	}
	return nil
}

func (s *demoSeeder) endPartnership(ctx context.Context, deal demoPartnership, daysAgo int) error {
	ender := s.cast.owners[deal.Consignor]
	if deal.Ended.By == "store" {
		ender = s.cast.owners[deal.Store]
	}
	command := partnershipservice.EndCommand{PartnershipID: s.cast.partnerships[deal.Key], Reason: deal.Ended.Reason}
	_, err := s.use.partnership.Ending.End.Execute(ctx, ender, command)
	if err == nil {
		return nil
	}
	if apperror.From(err).Code() != apperror.CodePartnershipBlocked {
		return fmt.Errorf("end: %w", err)
	}
	if err := s.payPartner(ctx, deal, daysAgo); err != nil {
		return err
	}
	if err := s.confirmPartnerPayments(ctx, deal, daysAgo); err != nil {
		return err
	}
	if _, err := s.use.partnership.Ending.End.Execute(ctx, ender, command); err != nil {
		return fmt.Errorf("end after settling: %w", err)
	}
	return nil
}

func (s *demoSeeder) readOldNotifications(ctx context.Context, daysAgo int) error {
	if daysAgo != unreadDaysAgo {
		return nil
	}
	for key, owner := range s.cast.owners {
		if err := s.use.inbox.MarkAllRead.Execute(ctx, owner); err != nil {
			return fmt.Errorf("read notifications of %s: %w", key, err)
		}
	}
	return nil
}

func (s *demoSeeder) isActive(deal demoPartnership, daysAgo int) bool {
	return deal.Ended == nil || daysAgo > deal.Ended.DaysAgo
}

func (s *demoSeeder) partiesOf(deal demoPartnership) (actor.Actor, actor.Actor) {
	return s.cast.owners[deal.Consignor], s.cast.owners[deal.Store]
}
