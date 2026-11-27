---
paths:
  - ".beacon/**"
---

# Beacon Record Conventions

This governs every file under `.beacon/`: the goal, capabilities, promises, decisions, questions, guides, work and config of the project. Beacon checks them with code, so how they change matters more than what they say.

---

## Edit through the CLI

- Create, change and close records with `beacon record` and `beacon close`, never by editing the files, because the binary validates the shape, classifies the change and asks the person for a yes when a promise would weaken.
- Read them with `beacon get`, `beacon find` and `beacon context`, not by scanning the folder, because the commands return what applies and what is blocked.
- Pass a long payload with `beacon record --file <path>`, because quoting it through the shell breaks it.
- After an unavoidable hand edit, run `beacon fmt` and then `beacon doctor`.

## When Beacon asks for a yes

- A change that weakens a promise, a runner or the config exits with code 2 and names what is pending. Stop and ask the person, and send the same payload again with `--approve` only after they say yes.
- Never rewrite a promise, a check or a baseline to make something pass.

---

## Non-goals

- Does not decide which records a project needs.
- Does not replace the method guide in `AGENTS.md`.
