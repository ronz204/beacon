---
title: Architecture overview
kind: architecture
applies: [src/**]
always: true
review: 12mo
reviewed: 2026-11-22
tags: [backend]
audience: both
status: current
refers: [bc-7f3a]
---
Three components: API, queue worker and web client. The API never talks
to the worker directly; everything goes through the queue.
