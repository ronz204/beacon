package ledger

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"unicode"

	"github.com/ronz204/beacon/source/kernel"
)

var attestedCommitPattern = regexp.MustCompile(`^[0-9a-f]{7,64}$`)

type CapabilityID struct {
	value kernel.RecordID
}

func ParseCapabilityID(raw string) (CapabilityID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return CapabilityID{}, invalidCapabilityValue(CapabilityIDInvalid, raw)
	}

	return CapabilityID{value: id}, nil
}

func (id CapabilityID) String() string {
	return id.value.String()
}

type PromiseID struct {
	value kernel.RecordID
}

func ParsePromiseID(raw string) (PromiseID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return PromiseID{}, invalidCapabilityValue(PromiseIDInvalid, raw)
	}

	return PromiseID{value: id}, nil
}

func (id PromiseID) String() string {
	return id.value.String()
}

type OriginWorkID struct {
	value kernel.RecordID
}

func ParseOriginWorkID(raw string) (OriginWorkID, error) {
	id, err := kernel.ParseRootRecordID(raw)

	if err != nil {
		return OriginWorkID{}, invalidCapabilityValue(PromiseOriginInvalid, raw)
	}

	return OriginWorkID{value: id}, nil
}

func (id OriginWorkID) String() string {
	return id.value.String()
}

type PromiseStatus string

const (
	PromiseProposed PromiseStatus = "proposed"
	PromiseActive   PromiseStatus = "active"
)

func PromiseStatuses() []PromiseStatus {
	return []PromiseStatus{
		PromiseProposed,
		PromiseActive,
	}
}

func ParsePromiseStatus(raw string) (PromiseStatus, error) {
	status := PromiseStatus(strings.TrimSpace(raw))

	if !slices.Contains(PromiseStatuses(), status) {
		return "", invalidCapabilityValue(PromiseStatusInvalid, raw)
	}

	return status, nil
}

type PromiseCheckInput struct {
	Runner string
	Params map[string]string
	Manual string
}

type PromiseCheck struct {
	runner string
	params map[string]string
	manual string
}

func ParsePromiseCheck(input PromiseCheckInput) (PromiseCheck, error) {
	runner := strings.TrimSpace(input.Runner)
	manual := strings.TrimSpace(input.Manual)

	if (runner == "") == (manual == "") {
		return invalidPromiseCheck()
	}
	if manual == "" {
		return parseRunnerCheck(runner, input.Params)
	}
	if len(input.Params) > 0 {
		return invalidPromiseCheck()
	}

	return PromiseCheck{manual: manual}, nil
}

func parseRunnerCheck(
	runner string,
	rawParams map[string]string,
) (PromiseCheck, error) {
	if !isCheckArgument(runner) {
		return invalidPromiseCheck()
	}

	params := make(map[string]string, len(rawParams))
	for key, raw := range rawParams {
		value := strings.TrimSpace(raw)

		if key == "" || key != strings.TrimSpace(key) {
			return invalidPromiseCheck()
		}
		if !isCheckArgument(value) {
			return invalidPromiseCheck()
		}

		params[key] = value
	}

	return PromiseCheck{runner: runner, params: params}, nil
}

func isCheckArgument(value string) bool {
	if value == "" || strings.HasPrefix(value, "-") {
		return false
	}

	return strings.IndexFunc(value, unicode.IsControl) < 0
}

func invalidPromiseCheck() (PromiseCheck, error) {
	return PromiseCheck{}, capabilityErrors.Fail(PromiseCheckInvalid, nil)
}

func (c PromiseCheck) IsManual() bool {
	return c.manual != ""
}

func (c PromiseCheck) IsZero() bool {
	return c.runner == "" && c.manual == ""
}

func (c PromiseCheck) Runner() string {
	return c.runner
}

func (c PromiseCheck) Params() map[string]string {
	return maps.Clone(c.params)
}

func (c PromiseCheck) Manual() string {
	return c.manual
}

type PromiseAttestationInput struct {
	By       string
	At       string
	Via      string
	Commit   string
	Evidence string
}

type PromiseAttestation struct {
	by       kernel.Actor
	at       kernel.CalendarDate
	via      kernel.Actor
	commit   string
	evidence string
}

func ParsePromiseAttestation(
	input PromiseAttestationInput,
) (PromiseAttestation, error) {
	by, byErr := kernel.ParseActor(input.By)
	at, atErr := kernel.ParseCalendarDate(input.At)
	commit := strings.TrimSpace(input.Commit)

	if byErr != nil || atErr != nil {
		return invalidPromiseAttestation()
	}
	if !by.IsHuman() || !attestedCommitPattern.MatchString(commit) {
		return invalidPromiseAttestation()
	}

	via, err := parseAttestationVia(input.Via)
	if err != nil {
		return invalidPromiseAttestation()
	}

	return PromiseAttestation{
		by:       by,
		at:       at,
		via:      via,
		commit:   commit,
		evidence: strings.TrimSpace(input.Evidence),
	}, nil
}

func parseAttestationVia(raw string) (kernel.Actor, error) {
	if strings.TrimSpace(raw) == "" {
		return kernel.Actor{}, nil
	}

	via, err := kernel.ParseActor(raw)
	if err != nil || via.IsHuman() {
		return kernel.Actor{}, capabilityErrors.Fail(PromiseAttestInvalid, nil)
	}

	return via, nil
}

func invalidPromiseAttestation() (PromiseAttestation, error) {
	failure := capabilityErrors.Fail(PromiseAttestInvalid, nil)
	return PromiseAttestation{}, failure
}

func (a PromiseAttestation) By() kernel.Actor {
	return a.by
}

func (a PromiseAttestation) At() kernel.CalendarDate {
	return a.at
}

func (a PromiseAttestation) Via() (kernel.Actor, bool) {
	return a.via, !kernel.IsMissing(a.via)
}

func (a PromiseAttestation) Commit() string {
	return a.commit
}

func (a PromiseAttestation) Evidence() string {
	return a.evidence
}

type VerificationStatus string

const (
	VerificationHolds      VerificationStatus = "holds"
	VerificationAttested   VerificationStatus = "attested"
	VerificationUnverified VerificationStatus = "unverified"
	VerificationViolated   VerificationStatus = "violated"
)

type Verification struct {
	status     VerificationStatus
	notRunHere bool
	commit     string
	footprint  string
	drifted    bool
}

func (v Verification) Status() VerificationStatus {
	return v.status
}

func (v Verification) NotRunHere() bool {
	return v.notRunHere
}

func (v Verification) Commit() string {
	return v.commit
}

func (v Verification) Footprint() string {
	return v.footprint
}

func (v Verification) Drifted() bool {
	return v.drifted
}

func (v Verification) WithDrift() Verification {
	v.drifted = true
	return v
}
