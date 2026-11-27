package blocks

import (
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
)

const (
	frontmatterFence = "---"
	stampKey         = "beacon: "
	stampLength      = 16
)

func StampManaged(content string) (string, error) {
	lines := strings.Split(unixLines(content), "\n")

	if !hasFrontmatter(lines) {
		return "", failManagedFrontmatter()
	}

	stamp := stampKey + digest(strings.Join(lines, "\n"))
	stamped := slices.Insert(lines, 1, stamp)

	return strings.Join(stamped, "\n"), nil
}

func InspectManaged(text string) Ownership {
	lines := strings.Split(unixLines(text), "\n")

	if len(lines) < 3 || lines[0] != frontmatterFence {
		return OwnershipForeign
	}

	recorded, stamped := strings.CutPrefix(lines[1], stampKey)
	if !stamped {
		return OwnershipForeign
	}

	original := slices.Delete(slices.Clone(lines), 1, 2)
	if recorded == digest(strings.Join(original, "\n")) {
		return OwnershipManaged
	}

	return OwnershipEdited
}

func hasFrontmatter(lines []string) bool {
	if len(lines) == 0 || lines[0] != frontmatterFence {
		return false
	}

	return slices.Contains(lines[1:], frontmatterFence)
}

func unixLines(text string) string {
	return strings.ReplaceAll(text, "\r\n", "\n")
}

func digest(text string) string {
	sum := sha256.Sum256([]byte(text))

	return hex.EncodeToString(sum[:])[:stampLength]
}
