---
description: Interview the person one question at a time to settle open questions and turn each answer into a Beacon record.
---

# Elicit

Settle what is unclear with the person before any code depends on it.

## When

- `beacon ready` shows work blocked by an open question.
- The request cannot be built without inventing an answer.
- The person asks to be interviewed, or the scope is vague.

## Steps

1. Run `beacon brief` and `beacon ready` to see the open questions and the goal.
2. Read what the project already says with `beacon find <words>` so no question repeats a decision.
3. Ask exactly one question per message. Offer your recommended answer and why, and wait.
4. After each answer, record it with `beacon record` (pass long text with `--file`):
   - a choice that lasts becomes a `decision`, with the alternatives dropped;
   - something still unknown becomes a `question`;
   - a concrete result with concrete inputs and outputs becomes a `proposed` promise.
5. Close a question that is answered with `beacon close <id> --reason <text>`.
6. Stop when no open question blocks the work, and say which records you wrote.

## Gate

- Do not start building while `beacon ready` still shows a blocking question.
- Never answer a question yourself to move on: ask, or leave it recorded as open.

## Alarms

- Two questions in one message: split them.
- An answer that changes a promise: it needs the person's yes, so run the same `beacon record` with `--approve` only after they say it.
- A long interview with nothing recorded: stop and record what is settled.

## Close

- List the records created or closed and the questions still open.
