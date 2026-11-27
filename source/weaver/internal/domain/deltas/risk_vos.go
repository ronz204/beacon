package deltas

type Risk string

const (
	RiskLow  Risk = "low"
	RiskHigh Risk = "high"
)

type RiskReason string

const (
	RiskWeakening RiskReason = "weakening_change"
	RiskRed       RiskReason = "red_promise"
	RiskDrift     RiskReason = "drifted_promise"
	RiskRaised    RiskReason = "raised_on_request"
)

type Zone struct {
	Capability string
	Red        bool
	Drifted    bool
}

type Assessment struct {
	Level   Risk
	Reasons []RiskReason
}
