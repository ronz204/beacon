package record

import (
	"strings"

	"github.com/ronz204/beacon/internal/kernel"
)

const (
	maximumSlugLength = 50
	fallbackSlug      = "record"
)

var slugFoldGroups = map[string]string{
	"a":  "áàäâãå",
	"ae": "æ",
	"c":  "ç",
	"e":  "éèëê",
	"i":  "íìïî",
	"n":  "ñ",
	"o":  "óòöôõø",
	"oe": "œ",
	"ss": "ß",
	"u":  "úùüû",
	"y":  "ýÿ",
}

var slugFolds = strings.NewReplacer(slugFoldPairs()...)

func FileName(
	id kernel.RecordID,
	title kernel.RecordText,
	extension string,
) string {
	return id.String() + "-" + slugOf(title.String()) + extension
}

func slugFoldPairs() []string {
	var pairs []string

	for ascii, letters := range slugFoldGroups {
		for _, letter := range letters {
			pairs = append(pairs, string(letter), ascii)
		}
	}

	return pairs
}

func slugOf(title string) string {
	folded := slugFolds.Replace(strings.ToLower(title))
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
