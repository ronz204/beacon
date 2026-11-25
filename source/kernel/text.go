package kernel

import "strings"

var textFoldGroups = map[string]string{
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

var textFolds = strings.NewReplacer(textFoldPairs()...)

type RecordText struct {
	value string
}

func ParseRecordText(raw string) (RecordText, error) {
	value := strings.TrimSpace(raw)

	if value == "" {
		return RecordText{}, invalidValue(KernelTextInvalid, raw)
	}

	return RecordText{value: value}, nil
}

func (t RecordText) String() string {
	return t.value
}

func FoldText(text string) string {
	return textFolds.Replace(strings.ToLower(text))
}

func textFoldPairs() []string {
	var pairs []string

	for ascii, letters := range textFoldGroups {
		for _, letter := range letters {
			pairs = append(pairs, string(letter), ascii)
		}
	}

	return pairs
}
