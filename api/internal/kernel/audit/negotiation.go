package audit

const (
	ProposalCountered Action = "PROPOSAL_COUNTERED"
	NeedPosted        Action = "NEED_POSTED"
	NeedExtended      Action = "NEED_EXTENDED"
	NeedFulfilled     Action = "NEED_FULFILLED"
	NeedClosed        Action = "NEED_CLOSED"
)

const NeedEntity EntityType = "NEED"
