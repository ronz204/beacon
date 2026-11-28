---
description: Build a promise test-first, with its test failing before it passes, then check it with Beacon.
---

# Ensure

Make a promise true with a test that fails first and passes after.

## When

- A promise has a runner but no test behind it.
- A new behavior is about to be built, or a bug is about to be fixed.

## Steps

1. Run `beacon context <paths>` for the files you will touch, and read the promises that apply.
2. If the behavior has no promise, propose one with `beacon record`: a concrete result with concrete inputs and outputs.
3. Write the test first and make it cite the promise id in its name or first line.
4. Run it and see it fail for the right reason. A test that passes before the code exists proves nothing.
5. Write the least code that makes it pass, then run the whole suite.
6. Run `beacon check <capability>` and read the real state of each promise.

## Gate

- `beacon check` shows the promise as holding, from a check that ran now.
- The test cites the promise id.

## Alarms

- The test passed on the first run: it checks nothing yet, fix the test.
- A promise, runner or baseline would have to change for the test to pass: stop and ask the person.
- A test that mirrors the code line by line: test the result, not the steps.

## Close

- Name the promises touched, the test that backs each, and any that remain unverified.
