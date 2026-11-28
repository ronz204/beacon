---
description: Review a diff against the promises and non-goals it touches, and weigh feedback on it by checking it before applying it.
---

# Review

Judge a change by what the project promised, not by taste.

## When

- You are asked to review a diff, a branch or a pull request.
- You receive review comments on your own change.
- Before handing work to the person.

## Steps

1. Run `beacon diff` to see the state changes by level, and `beacon verify` for the related checks and the risk.
2. Run `beacon context <paths>` on the changed files and read the promises, decisions and non-goals that apply.
3. Read the diff against them. Look first for a weakened promise, a changed runner or baseline, a feature a non-goal rules out, and a change with no test.
4. Report each finding with the file and line, why it matters, and the record it breaks. Order them by risk.
5. When you receive feedback, check each point against the code and the records before applying it. Apply what is right, answer with the evidence for what is not, and ask the person when unsure.

## Gate

- Every finding names a file and line or a record id.
- `beacon verify` ran in this session, and its result is quoted.

## Alarms

- Agreeing to a comment you have not checked: verify it first.
- Praise or filler where a finding should be: cut it.
- A finding of taste with no promise or bug behind it: mark it as a nit or drop it.

## Close

- End with the findings by risk, what you could not check, and the call: ready, ready with changes, or not ready.
