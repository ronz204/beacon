# Guides

El **contexto orientativo** del proyecto: producto, dominio, arquitectura, convenciones y runbooks. Es prosa, así que es la única parte del estado que no es YAML puro.

## Formato

Markdown con frontmatter YAML. Los metadatos son livianos: dos campos obligatorios y el resto opcional con defaults por tipo. Todos son de una sola palabra (`ledger`).

```markdown
---
title: Arquitectura general
kind: architecture
---
Tres componentes: API, worker de colas y cliente web. La API nunca habla
directo con el worker; todo pasa por la cola.
```

| Campo | Obligatorio | Qué hace |
|---|---|---|
| `title` | Sí | Título |
| `kind` | Sí | `product`, `domain`, `architecture`, `convention` o `runbook` |
| `applies` | No | Globs de cuándo es relevante; si falta, solo se carga cuando se pide o aparece en el índice del `brief` |
| `always` | No | `true` para que entre siempre al `brief` |
| `review` | No | Cada cuánto revisarla, como duración (`6mo`, `12mo`, `30d`); default por tipo |
| `reviewed` | No | Fecha de la última revisión (`YYYY-MM-DD`) |
| `tags` | No | Etiquetas de una palabra |
| `audience` | No | `agent`, `human` o `both` |
| `status` | No | `draft` o `current` |

- El `id` sale del nombre del archivo (`layout`).
- Las relaciones solo cuando hacen falta.
- La **intención narrativa** de una funcionalidad (el "qué queremos y por qué" para una persona) vive en una guide enlazada a la capability con `refers`. El YAML de la capability tiene solo una línea de intención.

## Carga: qué entra y cuándo

El contexto debe ser poco y preciso (`survey`, E1 y E2).

**El `brief` (lo que entra siempre al empezar una sesión):**

- La meta activa.
- Los bloqueos y rojos: preguntas abiertas que bloquean y promesas en rojo o con deriva.
- El trabajo en curso.
- Un índice de una línea por guide.

**Bajo demanda:** todo lo demás aparece cuando `context <paths>` coincide con el `applies` de una guide. Solo las guides con `always: true` entran al `brief` completas.

## Frescura

Una guide se marca como posiblemente desactualizada cuando:

- pasó su plazo (`reviewed` más `review`), o
- el código bajo su `applies` cambió mucho desde la última revisión (se calcula con git).

La segunda señal es determinista y sin modelo. Aparece en `beacon stale` y en el `brief` (`spider`).

**Fecha base:** `reviewed`. Si falta, en una guide se usa la fecha del último commit que tocó el archivo y en una decisión su `created`. Una revisión es un cambio de nivel 1: la persona confirma que sigue vigente y el agente actualiza `reviewed`.

**Plazos por defecto** (`review` en `config.yaml`, `ledger`): convención 6 meses, arquitectura 12, producto 12, dominio 12, runbook 6. Las decisiones también se revisan por antigüedad, con 6 meses. Son valores iniciales sin evidencia detrás: la señal fuerte es el cambio de código, y los plazos se ajustan con el uso.

## Mantenimiento

- `beacon doctor` avisa cuando una guide supera el máximo de `limits.guide` líneas (`ledger`); las largas se parten.
- Una guide obsoleta se **borra** (git guarda la historia); no pasa a `archive/`.