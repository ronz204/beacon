package random

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"github.com/ronz204/beacon/source/kernel"
)

const (
	initialHexLength  = 4
	maximumHexLength  = 16
	attemptsPerLength = 16
)

type Issuer struct {
	source io.Reader
}

func NewIssuer() Issuer {
	return NewIssuerFrom(rand.Reader)
}

func NewIssuerFrom(source io.Reader) Issuer {
	return Issuer{source: source}
}

func (i Issuer) IssueRecordID(
	taken func(kernel.RecordID) bool,
) (kernel.RecordID, error) {
	for length := initialHexLength; length <= maximumHexLength; length++ {
		for range attemptsPerLength {
			id, err := i.candidate(length)
			if err != nil {
				return kernel.RecordID{}, kernel.FailRecordIDIssue(err)
			}

			if !taken(id) {
				return id, nil
			}
		}
	}

	return kernel.RecordID{}, kernel.FailRecordIDIssue(nil)
}

func (i Issuer) candidate(length int) (kernel.RecordID, error) {
	buffer := make([]byte, (length+1)/2)

	if _, err := io.ReadFull(i.source, buffer); err != nil {
		return kernel.RecordID{}, err
	}

	digits := hex.EncodeToString(buffer)[:length]

	return kernel.ParseRecordID("bc-" + digits)
}
