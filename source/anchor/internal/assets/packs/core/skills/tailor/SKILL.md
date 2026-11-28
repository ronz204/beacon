---
description: Adjust the Beacon packs to this project by proposing a change to the packs section of config.yaml, which the person approves.
disable-model-invocation: true
---

# Tailor

Fit the installed rules to the project without writing rules by hand.

## When

- The person asks to adapt the packs, change the language or the commit length.
- A rule loads on the wrong files, or a stack has no rules scoped to it.

## Steps

1. Run `beacon doctor` and read `.beacon/config.yaml`, the `packs:` section above all.
2. Read the project to find what differs from the defaults: the language of identifiers and errors, the language of the interface, the stacks, the commit subject length.
3. Say each change you propose in one line, with its reason, and ask the person which to keep.
4. Write the whole config, with the chosen changes, to a file and send it with `beacon record --file <path>`. The record is `{"type":"config","record":{...}}` and replaces the current one, so copy every key you do not change.
5. The command exits with code 2 and names what needs a yes. Show it to the person and wait.
6. After they say yes, send the same file again with `--approve`, then run `beacon install` so the rules are rendered again.
7. Run `beacon doctor`.

## Gate

- The change went through `beacon record`, and the person gave the yes.
- `beacon doctor` reports no problems.

## Alarms

- Editing a rule file under `.claude/rules/`: never, the next install replaces it. Change `packs:` instead.
- Dropping a key from the config because it was not copied: compare with the file before sending.
- A binding the project does not need: leave the default.

## Close

- List what changed in `packs:` and which rules were rendered again.
