package handler

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
	"github.com/benditandayusaputra/trustip/api/internal/modules/settlement/domain"
)

type partyResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type splitResponse struct {
	Gross          int64 `json:"gross"`
	StoreShare     int64 `json:"store_share"`
	ConsignorShare int64 `json:"consignor_share"`
}

type periodResponse struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type standingResponse struct {
	IsOverdue bool   `json:"is_overdue"`
	DaysLate  int    `json:"days_late"`
	Age       string `json:"age"`
}

type lineResponse struct {
	ProductID     string        `json:"product_id"`
	ProductName   string        `json:"product_name"`
	SKU           string        `json:"sku"`
	Unit          string        `json:"unit"`
	PricingModel  string        `json:"pricing_model"`
	RetailPrice   int64         `json:"retail_price"`
	CommissionBPS int           `json:"commission_bps"`
	DepositPrice  int64         `json:"deposit_price"`
	Qty           int           `json:"qty"`
	Split         splitResponse `json:"split"`
}

type countRefResponse struct {
	ID        string `json:"id"`
	Number    string `json:"number"`
	CountedAt string `json:"counted_at"`
}

type sealEventResponse struct {
	EventType     string `json:"event_type"`
	ActorOrgName  string `json:"actor_org_name"`
	ActorUserName string `json:"actor_user_name"`
	OccurredAt    string `json:"occurred_at"`
	Code          string `json:"code"`
}

type sealResponse struct {
	IsValid bool                `json:"is_valid"`
	Events  []sealEventResponse `json:"events"`
}

type settlementResponse struct {
	ID            string             `json:"id"`
	Number        string             `json:"number"`
	PartnershipID string             `json:"partnership_id"`
	Status        string             `json:"status"`
	Schedule      string             `json:"schedule"`
	Period        periodResponse     `json:"period"`
	DueDate       string             `json:"due_date"`
	IssuedAt      string             `json:"issued_at"`
	PaidAt        *string            `json:"paid_at"`
	Standing      standingResponse   `json:"standing"`
	Store         partyResponse      `json:"store"`
	Consignor     partyResponse      `json:"consignor"`
	Lines         []lineResponse     `json:"lines"`
	SoldQty       int                `json:"sold_qty"`
	Split         splitResponse      `json:"split"`
	Counts        []countRefResponse `json:"counts"`
	Seal          sealResponse       `json:"seal"`
}

func toSettlementResponse(sealed domain.SealedView) settlementResponse {
	view := sealed.View
	response := settlementResponse{
		ID: view.ID, Number: view.Number, PartnershipID: view.PartnershipID, Status: string(view.Status),
		Schedule: string(view.Schedule), Period: toPeriod(view.Period), DueDate: view.DueDate.Format(time.DateOnly),
		IssuedAt: timestamp(view.IssuedAt), Standing: toStanding(sealed.Standing),
		Store:     partyResponse{ID: view.StoreID, Name: view.StoreName},
		Consignor: partyResponse{ID: view.ConsignorID, Name: view.ConsignorName},
		Lines:     toLines(view), SoldQty: view.Summary.Qty, Split: toSplit(view.Summary.Split),
		Counts: toCountRefs(view.Counts), Seal: toSealResponse(sealed.Seal),
	}
	if !view.PaidAt.IsZero() {
		paidAt := timestamp(view.PaidAt)
		response.PaidAt = &paidAt
	}
	return response
}

func toPeriod(period domain.Period) periodResponse {
	return periodResponse{Start: timestamp(period.Start), End: timestamp(period.End)}
}

func toStanding(standing domain.Standing) standingResponse {
	return standingResponse{IsOverdue: standing.IsOverdue, DaysLate: standing.DaysLate, Age: string(standing.Age)}
}

func toLines(view domain.View) []lineResponse {
	lines := make([]lineResponse, 0, len(view.Summary.Lines))
	for _, line := range view.Summary.Lines {
		product := view.Products[line.ProductID]
		lines = append(lines, lineResponse{
			ProductID: line.ProductID, ProductName: product.Name, SKU: product.SKU, Unit: product.Unit,
			PricingModel: string(line.Terms.Model), RetailPrice: line.Terms.RetailPrice.Int64(),
			CommissionBPS: line.Terms.CommissionBPS, DepositPrice: line.Terms.DepositPrice.Int64(), Qty: line.Qty,
			Split: toSplit(line.Split),
		})
	}
	return lines
}

func toCountRefs(counts []domain.SourceCount) []countRefResponse {
	refs := make([]countRefResponse, 0, len(counts))
	for _, count := range counts {
		refs = append(refs, countRefResponse{ID: count.ID, Number: count.Number, CountedAt: timestamp(count.CountedAt)})
	}
	return refs
}

func toSplit(split pricing.Split) splitResponse {
	return splitResponse{
		Gross: split.Gross.Int64(), StoreShare: split.StoreShare.Int64(), ConsignorShare: split.ConsignorShare.Int64(),
	}
}

func toSealResponse(verdict domain.SealVerdict) sealResponse {
	events := make([]sealEventResponse, 0, len(verdict.Events))
	for _, event := range verdict.Events {
		events = append(events, sealEventResponse{
			EventType: event.Event, ActorOrgName: event.ActorOrgName, ActorUserName: event.ActorUserName,
			OccurredAt: timestamp(event.At), Code: event.Code,
		})
	}
	return sealResponse{IsValid: verdict.IsValid, Events: events}
}

type counterpartResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type summaryResponse struct {
	ID             string              `json:"id"`
	Number         string              `json:"number"`
	PartnershipID  string              `json:"partnership_id"`
	Status         string              `json:"status"`
	Schedule       string              `json:"schedule"`
	Period         periodResponse      `json:"period"`
	DueDate        string              `json:"due_date"`
	IssuedAt       string              `json:"issued_at"`
	SoldQty        int                 `json:"sold_qty"`
	ConsignorShare int64               `json:"consignor_share"`
	Counterpart    counterpartResponse `json:"counterpart"`
	Standing       standingResponse    `json:"standing"`
}

func toSummaryResponse(item domain.ListItem) summaryResponse {
	return summaryResponse{
		ID: item.ID, Number: item.Number, PartnershipID: item.PartnershipID, Status: string(item.Status),
		Schedule: item.Schedule, Period: toPeriod(item.Period), DueDate: item.DueDate.Format(time.DateOnly),
		IssuedAt: timestamp(item.IssuedAt), SoldQty: item.SoldQty, ConsignorShare: item.ConsignorShare,
		Counterpart: counterpartResponse{
			ID: item.CounterpartID, Name: item.CounterpartName, Type: item.CounterpartType,
		},
		Standing: toStanding(item.Standing),
	}
}

type runningResponse struct {
	PartnershipID string        `json:"partnership_id"`
	HasSales      bool          `json:"has_sales"`
	Since         *string       `json:"since"`
	SoldQty       int           `json:"sold_qty"`
	Split         splitResponse `json:"split"`
}

func toRunningResponse(running domain.Running) runningResponse {
	response := runningResponse{
		PartnershipID: running.PartnershipID, HasSales: running.HasSales, SoldQty: running.SoldQty,
		Split: toSplit(running.Split),
	}
	if running.HasSales {
		since := timestamp(running.Since)
		response.Since = &since
	}
	return response
}

func timestamp(at time.Time) string {
	return at.UTC().Format(time.RFC3339)
}
