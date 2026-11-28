---
description: Find the root cause of a failing check or a bug before fixing it, and leave the regression as a promise.
---

# Locate

Understand why it fails before changing anything.

## When

- A check is red, a test fails or behavior is wrong.
- A fix you just made did not work.

## Steps

1. Run `beacon check` on the capability and read which promise is red and why.
2. Run `beacon context <paths>` on the failing files to see the promises and decisions that apply.
3. Reproduce the failure with the smallest input that shows it.
4. Trace backwards from the symptom to the first place the data or state goes wrong. Fix there, not where it shows.
5. Change one thing at a time and run the check again after each change.
6. Add the regression as a promise with its test, citing the id, with `beacon record`.

## Gate

- `beacon check` moves from red to green from a run you made after the fix.
- You can say the cause in one sentence.

## Alarms

- A fix before the cause is known: stop and go back to step 3.
- Three fixes failed: record a `question` with what you learned and ask the person to rethink the approach.
- The check is red because the promise or the test looks wrong: do not change either, ask the person.

## Close

- State the cause, the fix, the promise that now guards it, and what you did not verify.
