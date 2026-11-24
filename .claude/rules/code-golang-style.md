---
paths:
  - "**/*.go"
---

# Go Style Conventions

This governs the visual shape of Go source and tests: how far a line may run, how conditions and calls are laid out, and what test data may contain. `gofmt` is assumed and never wraps a line, so these rules cover what it leaves open. Domain modeling belongs to the domain-layer rule.

---

## Vertical over horizontal

- A line is at most 80 columns, counting a tab as 4, because code is read top to bottom and side-by-side diffs cut long lines.
- A function body never sits on its signature line, even when it is one statement.
- A condition joins at most two operands. More than that becomes separate guards, a `switch` or a named predicate, because a predicate's name says what the condition means.
- A signature, call or composite literal that does not fit puts one item per line with a trailing comma.
- A composite literal with three or more fields puts one field per line.
- A nested call that hurts reading is split with a named intermediate value.
- A blank line separates guards from the path that follows them.

## Comments

- Source and tests carry no comments. Names and small functions explain the code.
- The only exception is a directive the toolchain reads, such as `//go:embed` or `//go:build`.

## Tests

- Test data is clearly fictional (`jane`, `sample-agent`). It never holds the real name, handle, email or machine details of anyone working on the project.
- A table of inputs puts one case per line.

---

## Non-goals

- Doesn't decide package layout or domain shapes.
- Doesn't replace `gofmt` or `go vet`.
