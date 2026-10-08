package domain

const matchingPartnerMinScore = 75

type Candidate struct {
	OrganizationID  string
	Name            string
	Score           int
	IsPartner       bool
	HasOpenProposal bool
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
