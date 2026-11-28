---
description: Write or explain a change so another person understands it, with every claim tied to a record id or a file and line.
---

# Digest

Write for someone who was not in the session.

## When

- You write a summary, a handoff, a pull request description or documentation.
- The person asks you to explain what changed or why.

## Steps

1. Run `beacon brief` and `beacon get <id>` for the records the change touches, so the why comes from them and not from memory.
2. Say what changed in one sentence, then why, then how, in that order.
3. Give an example with a concrete input and its result.
4. Tie each claim to a record id or to `file:line`. Drop the claims you cannot tie.
5. Name what is still open or unverified, and which promises are not holding yet.
6. Cut what the reader would not miss: no history of attempts, no restating the diff.

## Gate

- Every id you cite resolves with `beacon get <id>`.
- Every `file:line` exists as written.

## Alarms

- A sentence that would still be true for any change: it says nothing, remove it.
- A word the reader may not know: define it once or use a plainer one.
- A claim of "done" or "works" without a check you ran: run it or say it is unverified.

## Close

- End with the records cited and what the reader should do next.
