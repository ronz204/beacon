package record

import (
	"strings"

	"github.com/ronz204/beacon/source/kernel"
)

const (
	maximumSlugLength = 50
	fallbackSlug      = "record"
)

func FileName(
	id kernel.RecordID,
	title kernel.RecordText,
	extension string,
) string {
	return id.String() + "-" + slugOf(title.String()) + extension
}

func slugOf(title string) string {
	folded := kernel.FoldText(title)
	words := strings.FieldsFunc(folded, isSlugBreak)

	if len(words) == 0 {
		return fallbackSlug
	}

	return limitSlug(words)
}

func isSlugBreak(char rune) bool {
	isLetter := char >= 'a' && char <= 'z'
	isDigit := char >= '0' && char <= '9'

	return !isLetter && !isDigit
}

func limitSlug(words []string) string {
	first := words[0]
	if len(first) > maximumSlugLength {
		return first[:maximumSlugLength]
	}

	slug := first
	for _, word := range words[1:] {
		candidate := slug + "-" + word
		if len(candidate) > maximumSlugLength {
			break
		}

		slug = candidate
	}

	return slug
}
