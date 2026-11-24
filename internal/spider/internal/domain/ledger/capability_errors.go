package ledger

import "github.com/ronz204/beacon/internal/kernel"

type CapabilityCode string

const (
	CapabilityIDInvalid    CapabilityCode = "capability_id_invalid"
	CapabilityFieldMissing CapabilityCode = "capability_field_missing"
	PromiseIDInvalid       CapabilityCode = "promise_id_invalid"
	PromiseStatusInvalid   CapabilityCode = "promise_status_invalid"
	PromiseCheckInvalid    CapabilityCode = "promise_check_invalid"
	PromiseOriginInvalid   CapabilityCode = "promise_origin_invalid"
	PromiseAttestInvalid   CapabilityCode = "promise_attest_invalid"
	PromiseAttestManual    CapabilityCode = "promise_attest_manual"
	PromiseFieldMissing    CapabilityCode = "promise_field_missing"
	PromiseRepeated        CapabilityCode = "promise_repeated"
)

type CapabilityError = kernel.Error[CapabilityCode]

var capabilityErrors = kernel.ErrorCatalog[CapabilityCode]{
	CapabilityIDInvalid:    "capability id must be a root record id",
	CapabilityFieldMissing: "capability is missing a required field",
	PromiseIDInvalid:       "promise id must be a root record id",
	PromiseStatusInvalid:   "promise status must be proposed or active",
	PromiseCheckInvalid:    "check needs a runner with safe params or manual",
	PromiseOriginInvalid:   "promise origin must be a root work id",
	PromiseAttestInvalid:   "attestation needs a human, date and commit",
	PromiseAttestManual:    "only manual promises can be attested",
	PromiseFieldMissing:    "promise is missing a required field",
	PromiseRepeated:        "promise id is repeated in the capability",
}

func invalidCapabilityValue(
	code CapabilityCode,
	raw string,
) CapabilityError {
	details := map[string]any{"value": raw}
	return capabilityErrors.Fail(code, details)
}
