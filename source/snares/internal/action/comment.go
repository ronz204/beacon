package action

import (
	"fmt"
	"strings"

	"github.com/ronz204/beacon/source/weaver"
)

func Comment(result CIResult) string {
	var out strings.Builder

	out.WriteString("## Beacon\n\n")
	out.WriteString(statusLine(result) + "\n")
	if result.Revision != "" {
		fmt.Fprintf(&out, "\n**Risk:** %s\n", riskText(result.Diff.Risk))
	}

	for _, block := range result.Blocks {
		fmt.Fprintf(&out, "\n### %s\n\n", block.Message)
		for _, item := range block.Items {
			fmt.Fprintf(&out, "- `%s`\n", item)
		}
	}
	if len(result.Check.Skipped) > 0 {
		out.WriteString("\n### Promises that did not run\n\n")
		for _, id := range result.Check.Skipped {
			fmt.Fprintf(&out, "- `%s`\n", id)
		}
	}
	if len(result.Warnings) > 0 {
		out.WriteString("\n### Warnings\n\n")
		for _, warning := range result.Warnings {
			fmt.Fprintf(&out, "- %s\n", warning)
		}
	}

	writeLevels(&out, result.Diff)

	return out.String()
}

func statusLine(result CIResult) string {
	if len(result.Blocks) > 0 {
		return "**Blocked.** A person has to read this before it merges."
	}

	return "**Passed.** Every check ran and no level 3 change waits for a yes."
}

func riskText(risk weaver.RiskView) string {
	if len(risk.Reasons) == 0 {
		return risk.Level
	}

	return fmt.Sprintf("%s (%s)", risk.Level, strings.Join(risk.Reasons, ", "))
}

func writeLevels(out *strings.Builder, diff weaver.DiffView) {
	if len(diff.Levels) == 0 {
		return
	}

	out.WriteString("\n### Changes by level\n")
	for _, level := range diff.Levels {
		fmt.Fprintf(out, "\n**Level %d**\n\n", level.Level)
		for _, entry := range level.Changes {
			fmt.Fprintf(out, "- `%s` %s %s", entry.Path, entry.Target, entry.ID)
			if entry.Reason != "" {
				fmt.Fprintf(out, " (%s)", entry.Reason)
			}
			if entry.NeedsYes && entry.Approved {
				out.WriteString(" approved")
			}
			out.WriteString("\n")
		}
	}
}
