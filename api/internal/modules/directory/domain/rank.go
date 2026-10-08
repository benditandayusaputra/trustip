package domain

import (
	"cmp"
	"slices"
)

type rankKey struct {
	score  int
	meters int
	id     string
}

type Listed interface {
	rankKey() rankKey
	SortPosition() Position
}

func RankByMatch[T Listed](cards []T) []T {
	ranked := slices.Clone(cards)
	slices.SortStableFunc(ranked, func(left, right T) int {
		leftKey, rightKey := left.rankKey(), right.rankKey()
		return cmp.Or(
			cmp.Compare(rightKey.score, leftKey.score),
			cmp.Compare(leftKey.meters, rightKey.meters),
			cmp.Compare(leftKey.id, rightKey.id),
		)
	})
	return ranked
}

func PageAfter[T Listed](ranked []T, afterID string, limit int) ([]T, bool) {
	start := 0
	if afterID != "" {
		start = slices.IndexFunc(ranked, func(card T) bool { return card.rankKey().id == afterID }) + 1
	}
	rest := ranked[start:]
	if len(rest) <= limit {
		return rest, false
	}
	return rest[:limit], true
}

func (c StoreCard) ScoredMatch() (Match, bool) {
	return c.Match, c.HasMatch
}

func (c ConsignorCard) ScoredMatch() (Match, bool) {
	return c.Match, c.HasMatch
}

func (c ProductCard) ScoredMatch() (Match, bool) {
	return c.Match, c.HasMatch
}
