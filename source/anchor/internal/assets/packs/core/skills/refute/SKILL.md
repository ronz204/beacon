---
description: Have a reviewer with clean context attack a non-trivial decision, then record the outcome with the alternatives it dropped.
disable-model-invocation: true
---

# Refute

Test a decision against someone who has not seen your reasoning.

## When

- The person asks to challenge a design, a dependency or a plan before it is built.
- A decision is costly to undo and affects other capabilities.

## Steps

1. State the decision in two sentences: what and why. Run `beacon find <words>` for earlier decisions on the same thing.
2. Start a subagent with a clean context. Give it only the decision, the records that apply and the files it touches, not your reasoning, and ask it to find the strongest reasons it is wrong.
3. Read the attack. For each point say whether it holds, with the evidence, and change the decision where it does.
4. Repeat from step 2 with the revised decision, up to three rounds. Stop earlier when a round finds nothing new.
5. Record the result with `beacon record` as a `decision`: `what`, `why`, and in `rejected` each alternative the attack raised and why it lost.
6. Tell the person what survived, what changed and what stays uncertain.

## Gate

- The decision is a record in `.beacon/` with its `rejected` list.
- No round used your own justification as the reviewer's input.

## Alarms

- A fourth round: stop, the disagreement is the person's to settle.
- An attack you dismissed without evidence: answer it or keep it as an open `question`.
- A decision that replaces an active one: it needs the person's yes.

## Close

- Name the decision id, the rounds used and the objections that remain.
