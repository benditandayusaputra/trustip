package domain

import (
	"time"

	"github.com/benditandayusaputra/trustip/api/internal/kernel/pricing"
)

type Allocation struct {
	ID         string
	ProductID  string
	OccurredAt time.Time
	Qty        int
	Terms      pricing.Terms
	Split      pricing.Split
}

type Line struct {
	ProductID string
	Terms     pricing.Terms
	Qty       int
	Split     pricing.Split
}

type Summary struct {
	Lines []Line
	Qty   int
	Split pricing.Split
}

type lineKey struct {
	productID string
	terms     pricing.Terms
}

func Summarize(allocations []Allocation) Summary {
	summary := Summary{Lines: []Line{}}
	positions := map[lineKey]int{}
	splits := make([]pricing.Split, 0, len(allocations))
	for _, allocation := range allocations {
		key := lineKey{productID: allocation.ProductID, terms: allocation.Terms}
		position, isKnown := positions[key]
		if !isKnown {
			position = len(summary.Lines)
			positions[key] = position
			summary.Lines = append(summary.Lines, Line{ProductID: allocation.ProductID, Terms: allocation.Terms})
		}
		summary.Lines[position] = summary.Lines[position].add(allocation)
		summary.Qty += allocation.Qty
		splits = append(splits, allocation.Split)
	}
	summary.Split = pricing.Summarize(splits)
	return summary
}

func (l Line) add(allocation Allocation) Line {
	l.Qty += allocation.Qty
	l.Split = pricing.Summarize([]pricing.Split{l.Split, allocation.Split})
	return l
}
