---
description: Before saying work is done, run the checks that prove it and report each promise's real state, then walk the manual ones.
---

# Finish

Claim nothing without fresh evidence from a command that ran in this session.

## When

- You are about to say the work is done, fixed or ready.
- The person asks whether it is finished.
- Manual promises are pending after code changed.

## Steps

1. Run `beacon verify` for the related checks and the risk of the change.
2. Run `beacon check` on each capability you touched.
3. Report each promise as it is: holds, attested, unverified, violated or drifting. Say holds only when a check passed in this session.
4. Run `beacon attest` to list the manual promises whose code changed.
5. Ask the person about one manual promise at a time, with what they should look at. Attest it with `beacon attest <id> --evidence <path or link>` only after they say yes.
6. Never commit, push or open a pull request: the person delivers.

## Gate

- Every claim of "done" is tied to a command and its output from this session.
- No manual promise is attested without the person's answer.

## Alarms

- "Should work", "probably" or "looks right": run the command instead.
- A red check you did not cause: report it, do not hide it or change the promise.
- A check you skipped: say so in the closing lines.

## Close

- End with the promises touched, the changes to project state, and what is pending.
