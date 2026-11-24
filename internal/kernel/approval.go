package kernel

import "strings"

type ApprovalInput struct {
	By  string
	At  string
	Via string
}

type Approval struct {
	by  Actor
	at  CalendarDate
	via Actor
}

func ParseApproval(input ApprovalInput) (Approval, error) {
	by, err := ParseActor(input.By)
	if err != nil {
		return Approval{}, err
	}

	at, err := ParseCalendarDate(input.At)
	if err != nil {
		return Approval{}, err
	}

	if strings.TrimSpace(input.Via) == "" {
		return Approval{by: by, at: at}, nil
	}

	via, err := ParseActor(input.Via)
	if err != nil {
		return Approval{}, err
	}
	if via.IsHuman() {
		return Approval{}, invalidValue(KernelApprovalInvalid, input.Via)
	}

	return Approval{
		by:  by,
		at:  at,
		via: via,
	}, nil
}

func (a Approval) By() Actor {
	return a.by
}

func (a Approval) At() CalendarDate {
	return a.at
}

func (a Approval) Via() (Actor, bool) {
	return a.via, a.via != Actor{}
}
