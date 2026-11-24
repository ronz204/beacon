---
title: Arquitectura general
kind: notes
applies: [src/**]
always: true
review: 12mo
reviewed: 2026-11-22
tags: [backend]
audience: both
status: current
refers: [bc-7f3a]
---
Tres componentes: API, worker de colas y cliente web. La API nunca habla
directo con el worker; todo pasa por la cola.
