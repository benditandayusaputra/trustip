package domain

const (
	reputationMaxPoints      = 10
	newcomerReputationPoints = 5
	ratedFromSettlements     = 3
	establishedFromPartners  = 3
	basisPointsPerPoint      = 1000
	roundingHalfBasisPoints  = basisPointsPerPoint / 2
)

type Reputation interface {
	component() component
}

type StoreReputation struct {
	DueSettlements int
	OnTimeBPS      int
	ActivePartners int
}

func (r StoreReputation) IsNew() bool {
	return r.DueSettlements < ratedFromSettlements
}

func (r StoreReputation) component() component {
	if r.IsNew() {
		return component{points: newcomerReputationPoints}
	}
	return component{
		points: (r.OnTimeBPS + roundingHalfBasisPoints) / basisPointsPerPoint,
		reason: Reason{Code: ReasonOnTimePayment, OnTimeBPS: r.OnTimeBPS}, isWon: true,
	}
}

type ConsignorReputation struct {
	ActivePartners int
	ProvenReports  int
}

func (r ConsignorReputation) IsEstablished() bool {
	return r.ActivePartners >= establishedFromPartners && r.ProvenReports == 0
}

func (r ConsignorReputation) component() component {
	if !r.IsEstablished() {
		return component{points: newcomerReputationPoints}
	}
	return component{points: reputationMaxPoints, reason: Reason{Code: ReasonEstablished}, isWon: true}
}
