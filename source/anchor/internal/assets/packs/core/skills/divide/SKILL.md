---
description: Split a large piece of work into tasks, each with a result that can be checked and the tasks that block it.
---

# Divide

Turn a big request into steps that can be built and checked one by one.

## When

- The work is too big to resume from the request plus `git diff`.
- The person asks for a plan or a breakdown.

## Steps

1. Run `beacon ready` and `beacon brief` to see the work in progress and what blocks it.
2. If no `work` exists, create one with `beacon record`, with the person's request word for word.
3. List the promises the work must make true, and propose the missing ones, concrete and `proposed`.
4. Cut the work into tasks. Each task has a title, a `parent` that is the work, the promise it `implements`, and a result you can check with a command.
5. Order them with `blocks`: a task that needs another one names it. Keep each task small enough to finish in one sitting.
6. Run `beacon doctor` to validate the hierarchy and `beacon ready` to see which task is first.
7. Show the person the list and wait for their changes before building.

## Gate

- Every task names the promise it implements and how to check it.
- `beacon doctor` reports no problems and `beacon ready` shows a first task.

## Alarms

- A task titled "refactor" or "clean up" with no result to check: rewrite it or drop it.
- A task that needs a decision nobody made: record a `question` that blocks it.
- More than about eight tasks: the work is two works, split it.

## Close

- List the tasks in order, the first one ready, and the questions still open.
