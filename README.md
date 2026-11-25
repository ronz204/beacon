# Beacon

Beacon es un método de trabajo para agentes de código que comprueba con código lo que el proyecto promete. Es un solo binario en Go. La base de conocimiento vive en `.claude/delta/` y empieza en `portal.md`.

## Estructura

```text
beacon/
├─ .claude/
│  ├─ delta/                    # base de conocimiento
│  └─ rules/                    # convenciones del proyecto
├─ .github/workflows/           # CI
├─ cmd/beacon/main.go           # punto de entrada del binario
├─ testing/
│  ├─ e2e/                      # pruebas del binario real
│  ├─ utils/                    # ayudas compartidas de prueba
│  └─ bench/                    # arnés del benchmark de fricción
├─ source/
│  ├─ kernel/                   # bases puras: errores, ids, actor
│  ├─ system/                   # escritura atómica y ejecución de git
│  ├─ router/                   # CLI
│  │  ├─ router.go
│  │  └─ internal/
│  │     ├─ handle/             # comandos
│  │     └─ output/             # sobre --json y códigos de salida
│  ├─ spider/                   # registros, grafo y señales
│  │  ├─ spider.go
│  │  └─ internal/
│  │     ├─ domain/
│  │     │  ├─ ledger/          # goal, capability
│  │     │  ├─ memory/          # decision, question, guide
│  │     │  ├─ agenda/          # work, task, claim
│  │     │  └─ thread/          # grafo, relaciones y señales
│  │     ├─ action/             # load, validate, fmt, doctor, brief, context
│  │     └─ driver/
│  │        ├─ record/          # YAML en .beacon/
│  │        ├─ commit/          # consultas a git
│  │        ├─ random/          # generación de ids
│  │        ├─ match/           # globs
│  │        ├─ double/          # contratos en memoria para tests
│  │        └─ states/          # estado local en .git/beacon/
│  ├─ engine/                   # checks, niveles de cambio, riesgo y vistas
│  │  ├─ engine.go
│  │  └─ internal/
│  │     ├─ domain/
│  │     │  ├─ deltas/
│  │     │  └─ checks/
│  │     ├─ action/
│  │     ├─ driver/
│  │     └─ render/
│  ├─ anchor/                   # instalación en cada agente
│  │  ├─ anchor.go
│  │  └─ internal/
│  │     ├─ action/
│  │     ├─ agents/             # claude, codex, cursor
│  │     └─ assets/
│  └─ sentry/                   # git hooks y CI
│     ├─ sentry.go
│     └─ internal/
│        ├─ action/
│        └─ assets/
├─ go.mod
├─ AGENTS.md
└─ README.md
```

Las carpetas se crean cuando su fase las necesita (`.claude/delta/status.md`).

## Reglas de la estructura

- **Dependencias en una sola dirección:** `router → sentry → engine → spider`, y `router → anchor`. Todas pueden usar `kernel` y `system`, salvo `domain`, que solo usa `kernel`.
- **Cada pieza se expone por su archivo de entrada** (`spider.go`, `engine.go`). Lo demás va en su `internal/`, y el compilador impide importarlo desde fuera.
- **`source/` no es privado para Go:** al no llamarse `internal`, otro módulo podría importar `kernel`, `system`, `router` y las fachadas de cada pieza. No se promete que esas APIs sean estables; lo que de verdad es privado vive en el `internal/` de cada pieza.
- **Cada pieza tiene solo las capas que necesita:** `domain` (reglas), `action` (casos de uso) y `driver` (YAML, git y otros sistemas).
- **Una pieza nueva** se agrega como `source/<pieza>/` con su archivo de entrada y se conecta en `router`.
- **Nombres de carpeta:** una sola palabra en minúsculas, de seis letras cuando exista una natural.
- **Dominio:** cada contexto es un paquete Go y sus archivos son `<subslice>_<kind>.go` (`.claude/rules/code-domain-layer.md`).
- **Las pruebas unitarias** van junto al código como `*_test.go`. `testing/` guarda solo las que cruzan todo el binario y sus ayudas.
