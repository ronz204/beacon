package kernel

import "strings"

type ApprovalInput struct {
	By    string
	At    string
	Via   string
	Count int
}

type Approval struct {
	by    Actor
	at    CalendarDate
	via   Actor
	count int
}

func ParseApproval(input ApprovalInput) (Approval, error) {
	if input.Count < 0 {
		return Approval{}, invalidValue(KernelApprovalInvalid, "count")
	}

	by, err := ParseActor(input.By)
	if err != nil {
		return Approval{}, err
	}

	at, err := ParseCalendarDate(input.At)
	if err != nil {
		return Approval{}, err
	}

	if strings.TrimSpace(input.Via) == "" {
		return Approval{by: by, at: at, count: input.Count}, nil
	}

	via, err := ParseActor(input.Via)
	if err != nil {
		return Approval{}, err
	}
	if via.IsHuman() {
		return Approval{}, invalidValue(KernelApprovalInvalid, input.Via)
	}

	return Approval{
		by:    by,
		at:    at,
		via:   via,
		count: input.Count,
	}, nil
}

func (a Approval) By() Actor {
	return a.by
}

func (a Approval) At() CalendarDate {
	return a.at
}

func (a Approval) Count() int {
	return max(a.count, 1)
}

func (a Approval) SameStamp(other Approval) bool {
	return a.by == other.by && a.at == other.at && a.via == other.via
}

func (a Approval) Recounted(count int) Approval {
	a.count = count

	return a
}

func (a Approval) Via() (Actor, bool) {
	return a.via, a.via != Actor{}
}
