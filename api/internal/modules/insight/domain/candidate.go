package domain

const matchingPartnerMinScore = 75

type Candidate struct {
	OrganizationID  string
	Name            string
	Score           int
	IsPartner       bool
	HasOpenProposal bool
}

func MarkOpenProposals(candidates []Candidate, counterpartIDs []string) []Candidate {
	open := make(map[string]bool, len(counterpartIDs))
	for _, id := range counterpartIDs {
		open[id] = true
	}
	marked := make([]Candidate, 0, len(candidates))
	for _, candidate := range candidates {
		candidate.HasOpenProposal = candidate.HasOpenProposal || open[candidate.OrganizationID]
		marked = append(marked, candidate)
	}
	return marked
}

func findMatchingPartner(candidate Candidate) []Insight {
	if !candidate.isSuggestable() {
		return nil
	}
	return []Insight{{
		Kind: KindMatchingPartner, Level: LevelLow, Score: candidate.Score,
		Subject: Subject{OrganizationID: candidate.OrganizationID, OrganizationName: candidate.Name},
	}}
}

func (c Candidate) isSuggestable() bool {
	return c.Score >= matchingPartnerMinScore && !c.IsPartner && !c.HasOpenProposal
}
