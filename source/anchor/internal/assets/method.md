<!-- BEGIN BEACON (managed block: do not edit inside) -->
## Beacon: project memory with checked promises

Beacon keeps the project's goal, promises, decisions and open questions in `.beacon/`. Use the `beacon` CLI (add `--json`); never edit those files by hand.

1. **Start.** At the start of a session run `beacon brief`: active goal, blockers, red or drifting promises, work in progress.
2. **Orient.** Before editing code run `beacon context <paths>` for the files you will touch, and respect what it returns. If the request pulls away from the goal, say so in one line.
3. **Don't guess.** If you cannot decide something without inventing it, ask the user (one question at a time) or record it as a question with `beacon record`.
4. **Record what lasts.** Record decisions, questions and new promises (they enter as `proposed`; state a concrete result with concrete inputs and outputs) as they happen; pass long text with `beacon record --file <path>`. Small changes need no record. If work is too big to resume from the request plus `git diff`, create a `work` with the user's request verbatim.
5. **Promises are the contract.** Tests cite the id of the promise they prove. Never edit, weaken or delete an active promise, a runner or a visual baseline to make something pass: Beacon will ask the user for a yes, and you must wait for it.
6. **Check before you finish.** Run `beacon check` on the capabilities you touched. Report each promise's real state (Holds, Attested, Unverified, Violated, drift). Never call something Holds unless a check passed.
7. **The person delivers.** Never commit, push or open a PR yourself. Then ask the person about each manual promise whose code changed, one at a time, with `beacon attest`. End with: promises touched, changes to project state, what is pending.

When Beacon blocks something, its message says exactly what to do. Follow it.
<!-- END BEACON -->
