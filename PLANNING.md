# Plan de construcción: Fase 3 (Anchor)

Este plan cubre solo la fase en curso: en qué orden se construye y qué archivos salen. El detalle funcional vive en la KB (`.claude/delta/`), y todas las fases, en `status`.

La numeración de pasos, decisiones y hitos **se reinicia** en esta fase. Las referencias de la KB y de los commits anteriores (pasos 1 a 37, D1 a D20) apuntan a los planes de las fases cerradas, que están en el historial de git: la Fase 1 (Spider) con `git show 967f5aa:PLANNING.md` y la Fase 2 (Weaver) con `git show 9e5ed52:PLANNING.md`.

> **Estado: en curso desde el 2026-11-26.** Las decisiones D1 a D11 están aprobadas por la persona. Esta fase instala **solo Claude Code**; Codex y Cursor pasan a la Fase 6 (D1). La prueba real no se hace en este repo sino en `src/projects/sandbox` (D4).

## 1. Panorama

### Qué es Anchor

Anchor es lo que instala Beacon en cada agente de código: la guía del método, los hooks que reinyectan el estado, el permiso para `beacon` y la marca del actor. Es **transporte opaco**: no decide nada. Toda decisión (qué inyectar, qué nivel tiene un cambio, qué promesa se cumple) vive en el binario (`anchor`, `weaver`).

El riesgo central del proyecto es que el agente no llame a las herramientas (`status`). Anchor es la respuesta a ese riesgo, y por eso esta fase termina midiendo, no solo instalando.

| Tema | Archivo de la KB |
|---|---|
| Qué instala, hooks, reinyección, agentes, bloque de guía | `.claude/delta/anchor.md` |
| Guía de redacción de promesas | `.claude/delta/pledge.md` |
| `BEACON_ACTOR` y `approved` | `.claude/delta/ledger.md`, `.claude/delta/deltas.md` |
| Estado local (`injected.json`, `backups/`) | `.claude/delta/layout.md` |
| Evidencia sobre obediencia y contexto (E2, E7, E8) y protocolo del benchmark | `.claude/delta/survey.md` |
| Comandos y sobre `--json` | `.claude/delta/weaver.md` |

### Alcance

**Entra en la Fase 3** (`status`):

| Qué | Cómo queda al cerrar la fase |
|---|---|
| `beacon install [--dry-run]` | Instala en Claude Code el bloque de guía, los hooks, el permiso y la marca del actor, con copia de seguridad y resumen |
| `beacon install --uninstall` | Quita solo lo que coincide byte a byte con lo instalado (D6) |
| `beacon hook <evento>` | Lee el JSON que Claude Code entrega al hook y responde en su formato (D2) |
| Guía de redacción en `record` | Al proponer una promesa, la respuesta trae la guía de `pledge` (D8) |
| Contadores de inyección | Cuánto y cuántas veces inyectó cada hook, en `.git/beacon/` (D9) |
| `testing/bench` | Arnés del benchmark de fricción: tareas, condiciones A a D y extracción de métricas, probado con un agente simulado (D3) |
| Banco de pruebas | `src/projects/sandbox`: un proyecto Go pequeño con Beacon instalado y una sesión real de Claude Code (D4, D12) |

**Queda fuera:**
- **Codex y Cursor** (Fase 6, con la segunda ola). El contrato `Agent` se diseña pensando en ellos, pero no se escribe su adaptador.
- Usar Beacon en este repo (D4): se retoma cuando el arranque de la Fase 4 exista.
- El arranque automático, `verify`, `attest`, `snares ci` y la plantilla de GitHub Actions (Fase 4).
- `stale` (Fase 5), `stats` y **correr** el benchmark (Fase 6).

### Qué existirá al terminar

| Al cerrar | Qué se puede hacer |
|---|---|
| **Bloque 1: núcleo de Anchor** | Instalar y desinstalar bloques gestionados en `AGENTS.md` y `.gitattributes`, con `--dry-run` y copias |
| **Bloque 2: hooks y respuesta de `record`** | `beacon hook` responde a un agente sin romperle la sesión; los contadores y la guía de redacción existen |
| **Bloque 3: Claude Code** | Una sesión real de Claude Code, en el sandbox, recibe el `brief` al empezar y tras compactar, y el contexto antes de editar |
| **Bloque 4: benchmark y banco de pruebas** | El arnés produce un reporte de métricas, y el flujo completo se probó sobre un proyecto con promesas reales |

### Cómo fluye la información

```text
router (handle/*)          lee argumentos y stdin, imprime el sobre
        │
        ├── install ───────→ anchor.go     planifica y aplica la instalación
        │
        └── hook ──┬──────→ anchor.go      entiende el JSON del agente y da formato a la respuesta
                   └──────→ spider.go      decide qué se inyecta (Brief, HookContext)

anchor.go                  la única puerta de Anchor
        │
        ▼
action/*                   plan, apply, uninstall, hook
        │
        ├──→ domain/blocks     bloque gestionado: marcar, reescribir, quitar byte a byte
        ├──→ agents/*          qué archivos y entradas pide cada agente (contrato Agent)
        └──── driver/files     lee y escribe archivos del repo, con copia de seguridad
```

### Quién llama a quién

- `router` llama a **Anchor** para instalar y para entender y formatear el hook, y a **Spider** para el contenido que el hook inyecta. Anchor **no importa Spider ni Weaver**: es transporte, y lo que inyecta lo decide el binario.
- Un hook **nunca bloquea al agente**: sea cual sea el fallo, sale con código 0, escribe el motivo en stderr y no inyecta nada (D5).
- Anchor solo escribe archivos del repo que no son de Beacon (`AGENTS.md`, `CLAUDE.md`, `.gitattributes`, `.claude/`). Nunca toca `.beacon/` ni `.git/beacon/` (los contadores los escribe Spider).

## 2. Decisiones

Todas aprobadas por la persona. D1 a D4 el 2026-11-26 (con el ajuste de alcance del mismo día) y D5 a D12 el mismo día, al revisar el plan.

### D1. Solo Claude Code entra en la Fase 3

- **Decisión:** el único adaptador de esta fase es Claude Code. Codex y Cursor pasan a la Fase 6, junto con OpenCode, Gemini CLI y Copilot.
- **Por qué:** el agente de uso diario se prueba a fondo, y los demás se verifican contra su documentación cuando haya con qué compararlos (`survey`, E8).
- **Consecuencia:** el contrato `Agent` se diseña con los momentos que un agente puede cubrir (D11) para que un segundo adaptador no obligue a reescribirlo. `anchor` y `status` se ajustan.
- **Afecta a:** pasos 2 y 7.

### D2. `beacon hook <evento>` lee el JSON del agente

- **Decisión:** los hooks de cada agente llaman a `beacon hook <evento> --agent <nombre>`. El binario lee el JSON de stdin (id de sesión, ruta, motivo del arranque), decide qué imprimir y lo devuelve en el formato que ese agente espera. El hook instalado es solo configuración, sin lógica ni scripts.
- **Por qué:** un script shell con `jq` es frágil en Windows y duplica lógica por agente. Así Anchor sigue siendo transporte opaco y se prueba en Go.
- **Eventos de Beacon:** `start` (arranque, reanudación y tras compactar), `edit` (antes de editar un archivo) y `stop` (cierre).
- **Consecuencia:** `context --hook --session <id>` sigue existiendo para quien lo llame a mano; `hook edit` usa el mismo `HookContext` de Spider. Cierra el punto abierto de `status` sobre cómo identificar la sesión.
- **Afecta a:** pasos 4 y 7.

### D3. El arnés del benchmark es esqueleto y pruebas, sin sesiones reales

- **Decisión:** `testing/bench` define el formato de tareas, arma las condiciones A a D y extrae métricas de un transcript. Se prueba con un agente simulado que reproduce transcripts de ejemplo. Correr el benchmark, con sus 90 sesiones y su costo, es de la Fase 6.
- **Por qué:** el arnés debe existir antes de la Fase 4 para no descubrir tarde que una métrica no se puede medir, pero sin gastar tokens.
- **Afecta a:** paso 9.

### D4. El banco de pruebas es `sandbox`, no este repo

- **Decisión:** Beacon se instala y se usa en `src/projects/sandbox`, no en este repo. Usarlo en Beacon mismo se pospone hasta que el arranque (Fase 4) lo vuelva barato.
- **Por qué:** un repo desechable permite fallar, reiniciar y repetir la prueba real sin tocar `.claude/settings.json` ni la historia de este proyecto.
- **Consecuencia:** el sandbox **no está versionado con Beacon**, así que su contenido se describe en D12 para poder rehacerlo.
- **Afecta a:** pasos 8 y 10.

### D5. Un hook nunca falla la sesión del agente

- **Decisión:** `beacon hook` sale siempre con código 0. Un error (estado inválido, git ausente, JSON ilegible) va a stderr y no inyecta nada. Si no existe `.beacon/`, no imprime nada.
- **Por qué:** en varios agentes un código distinto de 0 bloquea la acción o ensucia la sesión, y un recordatorio no vale eso. La falla queda visible para `doctor`.
- **Afecta a:** paso 4.

### D6. Instalar y desinstalar son idempotentes y exactos

- **Decisión:** `install` reescribe sus bloques sin duplicarse. `install --uninstall` quita **solo lo que coincide byte a byte** con lo instalado (`anchor`); lo que la persona editó dentro de un bloque se conserva y se avisa en el resumen.
- **Archivos de texto:** bloque marcado con `<!-- BEGIN BEACON -->` y `<!-- END BEACON -->`.
- **Archivos JSON:** no admiten comentarios, así que se fusionan **entradas** (un permiso, una variable de entorno, un hook cuyo comando empieza con `beacon hook`). Reconocer lo propio es comparar el valor exacto. Se reescribe con la sangría que ya tenía el archivo.
- **Dónde va la guía:** en `AGENTS.md`, y `CLAUDE.md` la importa con `@AGENTS.md` (`anchor`), para que los agentes de la Fase 6 reutilicen el mismo bloque.
- **Afecta a:** pasos 1, 2 y 7.

### D7. Qué se puede omitir al instalar

- **Decisión:** `--skip hooks` instala solo la guía y el permiso (condición B del benchmark), y `--skip reinject` omite los hooks de compactación y de edición pero deja el de inicio (condición D). Sin flags, instala todo (condición C).
- **Por qué:** sin esto el arnés tendría que desarmar una instalación a mano, y la condición D no se podría reproducir. Son los únicos valores.
- **Afecta a:** pasos 3 y 9.

### D8. La guía de redacción viaja en la respuesta de `record`

- **Decisión:** cuando un `record` crea una promesa, `data.guidance` trae la guía de `pledge` en pocas líneas (resultado observable, entradas y salidas concretas, una cosa por promesa, EARS opcional) y el texto la imprime al final. Cuando no crea ninguna, la clave no aparece.
- **Por qué:** es una guía, no una regla (`pledge`); llega justo cuando se redacta, no solo en el bloque de instalación.
- **Afecta a:** paso 6.

### D9. Contadores de inyección en `.git/beacon/`

- **Decisión:** Spider guarda en `metrics.json` (con `version`, escritura atómica y candado, como los demás estados locales) cuántas veces inyectó cada hook y cuántas líneas. Lo escribe el comando `hook`, a través de `spider.go`.
- **Por qué:** el benchmark compara con y sin reinyección (`anchor`), y `stats` (Fase 6) los leerá. Una sesión sin actividad en 7 días se descarta, igual que `injected.json`.
- **Afecta a:** paso 5.

### D10. `install` no toca `.beacon/`

- **Decisión:** `install` escribe solo archivos de agentes y `.gitattributes`. La clave `agents` de `config.yaml` (`ledger`) se llena cuando el arranque la cree (Fase 4); mientras tanto el resumen lista los agentes instalados.
- **Por qué:** mantiene a Anchor sin dependencia de Spider, y `.beacon/` lo crea quien tiene el esquema.
- **Efecto:** en un repo recién instalado y sin `.beacon/`, `brief` y los hooks no imprimen nada; el resumen de `install` dice cuál es el siguiente paso.
- **Afecta a:** pasos 3 y 10.

### D11. Si el agente no tiene el evento, la guía cubre ese momento

- **Decisión:** cada adaptador declara qué momentos cubre (`start`, `compact`, `edit`, `stop`). Lo que el agente no ofrezca se omite y el resumen lo dice en una línea, sin simularlo con otro hook.
- **Por qué:** `anchor` lo prevé y evita instalar hooks que disparan en el momento equivocado. Es lo que el benchmark compara. Es también lo que usarán los adaptadores de la Fase 6.
- **Afecta a:** pasos 2 y 7.

### D12. Cómo es el sandbox y cómo llega el binario

- **Sandbox:** un módulo Go pequeño en `src/projects/sandbox` con `git init` y un commit inicial. Tiene dos paquetes con tests (por ejemplo, una calculadora de descuentos y un validador de correos), de modo que haya dos capabilities con promesas que `go test` compruebe. Un test falla a propósito en una rama aparte para la tarea que tienta a debilitar una promesa (paso 9).
- **Binario:** `go install ./cmd/beacon` deja `beacon` en `~/go/bin`, que ya está en el `PATH`. Se reinstala tras cada cambio, y es lo que hará un usuario real. Los hooks llaman a `beacon` por nombre.
- **Por qué:** un proyecto con tests reales ejercita `check`, `record` y los pendientes de sí, que un repo vacío no puede.
- **Afecta a:** pasos 8 y 10.

## 3. Estructura objetivo

Lo que **se agrega o cambia** en la Fase 3. Los números son el paso que crea cada parte, y los tests van junto a cada archivo.

```text
beacon/
├─ source/
│  ├─ router/
│  │  └─ internal/
│  │     └─ handle/
│  │        ├─ install.go                # 3 · install, --dry-run, --uninstall, --skip
│  │        ├─ hook.go                   # 4 · beacon hook <evento> --agent
│  │        └─ record.go                 # 6 · suma la guía de redacción
│  │
│  ├─ spider/
│  │  ├─ spider.go                       # 5 · RecordInjection
│  │  └─ internal/
│  │     └─ driver/states/
│  │        └─ metrics.go                # 5 · metrics.json
│  │
│  ├─ weaver/
│  │  └─ internal/action/
│  │     └─ record.go                    # 6 · detecta promesas nuevas y arma guidance
│  │
│  └─ anchor/
│     ├─ anchor.go                       # 1 · fachada pública de la pieza
│     ├─ views.go                        # 2 · tipos planos: InstallPlan, FileChange, HookInput
│     └─ internal/
│        ├─ domain/
│        │  └─ blocks/                   # bloque gestionado y fusión de entradas
│        │     ├─ block_vos.go           # 1 · Marker, Block
│        │     ├─ block_policies.go      # 1 · Upsert, Remove (byte a byte)
│        │     ├─ block_errors.go        # 1
│        │     ├─ entry_policies.go      # 2 · fusionar y quitar entradas JSON (D6)
│        │     └─ entry_errors.go        # 2
│        ├─ action/
│        │  ├─ plan.go                   # 2 · qué cambiaría
│        │  ├─ apply.go                  # 2 · aplicar con copia de seguridad
│        │  ├─ uninstall.go              # 3
│        │  └─ hook.go                   # 4 · parsear y formatear el hook
│        ├─ agents/
│        │  ├─ agent.go                  # 2 · contrato Agent y momentos cubiertos
│        │  └─ claude/                   # 7 · CLAUDE.md, settings.json, hooks
│        ├─ assets/
│        │  └─ method.md                 # 1 · bloque de guía del método (7 reglas)
│        └─ driver/
│           ├─ files/
│           │  └─ files.go               # 2 · lectura, escritura y backups/
│           └─ double/
│              └─ double.go              # 2 · contratos en memoria para tests
│
└─ testing/
   ├─ bench/                             # 9 · arnés del benchmark
   │  ├─ task.go                         # formato de tarea
   │  ├─ condition.go                    # condiciones A a D sobre un repo temporal
   │  ├─ transcript.go                   # extracción de métricas
   │  ├─ report.go                       # reporte con tasas
   │  └─ fixtures/                       # transcripts y tareas de ejemplo
   └─ e2e/
      ├─ install_test.go                 # 3 · núcleo en repo temporal
      ├─ hook_test.go                    # 4 · hook con JSON de agente
      └─ claude_test.go                  # 7
```

Fuera del repo: `src/projects/sandbox` (D12), que se prepara en el paso 8.

### Dónde va cada tipo de código

| Si el código... | Va en |
|---|---|
| Decide cómo se marca, reescribe o quita un bloque | `anchor/internal/domain/blocks` |
| Dice qué archivos y entradas pide un agente | `anchor/internal/agents/<agente>` |
| Lee o escribe archivos de agentes y guarda copias | `anchor/internal/driver/files` |
| Entiende o da formato al JSON de un hook | `anchor/internal/action/hook.go` |
| Decide **qué** inyecta un hook | Spider (`Brief`, `HookContext`) |
| Cuenta lo inyectado | Spider (`driver/states/metrics.go`) |
| Decide si una promesa es nueva y arma su guía | `weaver/internal/action/record.go` |

### Quién puede importar a quién

```text
cmd/beacon → router → anchor.go → action → domain/blocks → kernel
                │                     │
                │                     ├──→ agents/* (contrato Agent)
                │                     └──── driver/files (implementa el contrato de archivos)
                │
                ├──→ spider.go   (lo que inyecta el hook)
                └──→ weaver.go
```

`anchor` no importa `spider` ni `weaver`. `router` es quien junta a las dos piezas en `handle/hook.go`.

## 4. Bloque 1: núcleo de Anchor

**Rama:** `feat/anchor-core`, desde `main` actualizado. **Forma de trabajo:** primero el test que falla, luego el código mínimo, y al final `go test ./...` en verde.

### Paso 1. Bloques gestionados y bloque de guía (hecho)

- **Objetivo:** que reescribir y quitar lo instalado sea una función pura y probada (D6).
- **Construye:** `anchor.go` (solo el constructor), el contexto `blocks`, y `assets/method.md` con el bloque de guía de `anchor` (siete reglas en inglés, sin sangría).
- **`Upsert(text, block)`:** inserta el bloque al final si no existe y lo reemplaza si existe. Idempotente: aplicarlo dos veces da el mismo texto.
- **`Remove(text, block)`:** quita el bloque **solo si coincide byte a byte** con el instalado. Si la persona lo editó, lo deja y devuelve que quedó editado.
- **Cuidado:** respetar el salto de línea del archivo (LF o CRLF) y no tocar nada fuera de los marcadores.
- **Hecho cuando:** `Upsert` dos veces no cambia el texto, `Remove` devuelve el archivo original, un bloque editado se conserva, y el bloque de guía no pasa de 30 líneas (`survey`, E2).
- **Cómo quedó:**
  - El bloque de guía mide 15 líneas, no 20: el `20 a 30` de `anchor` cuenta líneas dibujadas en pantalla y el test limita las líneas del archivo.
  - `ParseBlock` acepta dos estilos de marcador, `<!-- BEGIN BEACON ... -->` y `# BEGIN BEACON`, porque `.gitattributes` no admite HTML. El estilo lo da la primera línea. `anchor.New` ya no es solo un constructor: abre el repo (paso 2).
  - `Upsert` conserva el salto de línea del archivo (LF o CRLF) y deja una línea en blanco antes del bloque. `Remove` devuelve `removed`, `absent` o `edited`, y si el bloque estaba al final se lleva también esa línea en blanco, así que el ciclo `Upsert` → `Remove` devuelve el archivo original si terminaba en salto de línea.
  - **Límite:** un archivo que no terminaba en salto de línea lo gana al instalar y desinstalar. No se corrige: cambia un byte final y la alternativa es recordar cómo era el archivo.

### Paso 2. Contrato de agente, plan y aplicación (hecho)

- **Objetivo:** calcular qué cambiaría antes de tocar nada, y aplicarlo sin perder lo que había.
- **Construye:** `views.go`, `agents/agent.go`, `domain/blocks/entry_policies.go`, `action/plan.go`, `action/apply.go`, `driver/files` y `driver/double`.
- **Contrato `Agent`:** dado el estado del repo, devuelve una lista de `FileChange` (ruta, contenido nuevo y de qué parte de la instalación viene) y los momentos que cubre (D11). Se escribe con un solo adaptador, pero sin suponer cuál.
- **Entradas JSON (D6):** `MergeEntries` y `RemoveEntries` sobre un permiso, una variable de entorno o un hook; comparan por valor y conservan el resto del archivo.
- **Copia de seguridad:** antes de modificar un archivo que existe, se guarda en `.git/beacon/backups/` con su ruta relativa y la fecha. Un archivo nuevo no tiene copia.
- **Hecho cuando:** el plan de un repo vacío lista los archivos que crearía, aplicar y volver a planear da cero cambios, y un archivo existente queda copiado antes de cambiar.
- **Cómo quedó:**
  - **Contrato `Agent`:** `Name`, `Moments` (los cuatro momentos que puede cubrir) y `Contribute(Options)`, que devuelve **declaraciones**, no texto: bloques por archivo (`BlockFile`) y entradas JSON por archivo (`EntryFile`), cada una con su parte (`guide`, `import`, `permission`, `actor`, `hooks`, `reinject`). Así instalar y desinstalar usan la misma declaración.
  - **Entradas JSON** (`entry_vos.go`, `entry_policies.go`): dos modos, `set` (una clave con un valor) y `append` (un elemento en una lista). Se conserva el orden de las claves, la sangría (espacios o tabulador), el salto de línea y el salto final; un archivo sin cambios no se reescribe. Una clave `set` que ya tiene **otro** valor falla con `entry_conflict` en vez de pisarlo, y `RemoveEntries` deja y devuelve lo que la persona cambió. Un archivo que no es un objeto JSON falla con `document_invalid`. Los escapes de HTML (`<`, `>`, `&`) no se alteran.
  - **Plan y aplicación** (`action/plan.go`, `action/apply.go`): `Planner` lee cada archivo una vez, acumula los cambios de todas las contribuciones sobre el mismo texto y devuelve solo los archivos que cambian, con las partes que los cambiaron y los momentos que cada agente no cubre. `Applier` guarda todas las copias antes de escribir el primer archivo, así que una copia que falla no deja nada a medias.
  - **Archivos:** el contrato `FileStore` (`Read`, `Write`, `Backup`) vive en `action`; `driver/files` lo implementa sobre el repo, con escritura atómica y copias en `.git/beacon/backups/<fecha>/<ruta>` del directorio común de git, y rechaza rutas que salgan del repo (`file_path_invalid`). `driver/double` es la versión en memoria.
  - **Fachada:** `anchor.New(ctx, directorio)` y `Install(ctx, InstallOptions)`, que planifica y, sin `DryRun`, aplica. Con un solo método, `--dry-run` y la instalación comparten el cálculo. Todavía no hay adaptadores: el plan solo trae la guía en `AGENTS.md`.

### Paso 3. `beacon install`

- **Objetivo:** el comando completo sobre el núcleo común: `AGENTS.md` y `.gitattributes`.
- **Construye:** `handle/install.go` y `action/uninstall.go`.
- **Flags:** `--dry-run` (lista los archivos y cómo quedarían, sin tocar nada), `--uninstall` y `--skip hooks|reinject` (D7).
- **`.gitattributes`:** agrega `.beacon/** text eol=lf` (`layout`).
- **Resumen** (`--json`: `data.files`, `data.skipped`, `data.review`): qué se creó o cambió, qué se omitió por no tener el agente el evento, y qué debe revisar la persona. Si `beacon` no está en el `PATH`, lo avisa, porque los hooks lo llaman por nombre. Sin `.beacon/`, el `hint` dice cuál es el siguiente paso (D10).
- **Hecho cuando:** en un repo temporal, `--dry-run` no cambia ningún byte, `install` crea los dos archivos, repetirlo no cambia nada, y `--uninstall` deja el repo como estaba.

## 5. Bloque 2: hooks y respuesta de `record`

**Rama:** `feat/anchor-hooks`, desde `main` con el Bloque 1 fusionado.

### Paso 4. `beacon hook`

- **Objetivo:** que el binario sea quien entiende a cada agente (D2) y que ningún fallo llegue a la sesión (D5).
- **Construye:** `action/hook.go` y `handle/hook.go`. La interfaz `Agent` suma `ParseHook(event, stdin)` y `RenderHook(event, text)`; en este paso se escriben con lo mínimo de Claude Code para probar el flujo.
- **Flujo:** leer el JSON → obtener sesión y ruta → `start` llama a `Spider.Brief`, `edit` a `Spider.HookContext` (una vez por archivo y sesión, con el tope `limits.hook`), `stop` imprime el recordatorio de `check` y `attest` → dar formato al texto para el agente.
- **Cuidado:** `edit` no consulta git (`anchor`). Una ruta fuera del repo, un JSON ilegible o un estado inválido no inyectan nada (D5). Las rutas de Windows y las absolutas se convierten a la raíz del repo (`Project.RepoPath`).
- **Hecho cuando:** con un JSON de Claude Code, `hook start` imprime el `brief`, `hook edit` imprime lo nuevo la primera vez y nada la segunda, y con estado inválido o sin `.beacon/` sale con código 0 y sin salida.

### Paso 5. Contadores de inyección

- **Construye:** `metrics.json` en Spider (D9) y `RecordInjection(ctx, session, event, lines)` en `spider.go`. `hook` lo llama después de imprimir.
- **Hecho cuando:** dos hooks concurrentes no pierden una cuenta, y una sesión sin actividad en 7 días se descarta.

### Paso 6. Guía de redacción en `record`

- **Construye:** en `weaver/internal/action/record.go`, detectar las promesas que el payload crea y devolver `guidance` (D8); el handler la imprime en texto.
- **Hecho cuando:** un `record` que crea una promesa trae la guía, uno que solo edita una promesa existente no la trae, y el texto de la guía coincide con `pledge`.

## 6. Bloque 3: Claude Code

**Rama:** `feat/anchor-claude`, desde `main` con el Bloque 2 fusionado.

### Paso 7. Adaptador de Claude Code

- **Objetivo:** que los cuatro momentos de `anchor` queden cubiertos en Claude Code.
- **Antes de escribir código:** verificar contra la documentación de hooks de Claude Code (a confirmar, `anchor`): el evento de inicio y sus tipos (arranque, reanudación, compactación), el evento previo a editar y qué herramientas cubre, la forma del JSON de entrada, la forma de la salida que se agrega al contexto, y la clave `env` de `settings.json`. Anotar la fecha y lo verificado en `anchor`.
- **Construye** (`agents/claude`):
  - `CLAUDE.md` con una línea `@AGENTS.md` en un bloque gestionado (D6); el bloque de guía va en `AGENTS.md`.
  - `.claude/settings.json`: permiso `Bash(beacon *)`, `BEACON_ACTOR=agent:claude-code` en `env`, y los hooks `start`, `edit` y `stop` (D6, D11). Respeta lo que el archivo ya tiene.
  - `ParseHook` y `RenderHook` completos.
- **Hecho cuando:** sobre un repo temporal con un `settings.json` existente, `install` conserva sus permisos, agrega los de Beacon y `--uninstall` deja el archivo como estaba; `hook` responde en el formato que Claude Code espera; y un comando de Beacon corrido con `BEACON_ACTOR` pone `agent:claude-code` en `approved.by` y `human:<user.name>` en `via`.

### Paso 8. Prueba con una sesión real en el sandbox

- **Objetivo:** confirmar lo que los tests no pueden: que el agente recibe lo inyectado.
- **Prepara el sandbox (D12):** el módulo Go con dos paquetes con tests, `git init` y un commit inicial. Se instala el binario con `go install ./cmd/beacon`.
- **Pasos:** `beacon install` en el sandbox (primero con `--dry-run`), un registro de ejemplo, y una sesión de Claude Code que comprueba que (a) el `brief` aparece al empezar, (b) reaparece tras `/compact`, (c) `context` llega antes de la primera edición de un archivo y no se repite, (d) el hook funciona desde un subdirectorio y en Windows.
- **Hecho cuando:** los cuatro casos pasan, o lo que falla queda corregido o anotado en `anchor` con su causa. El resultado se escribe en la KB, no solo en el commit, y el transcript de la sesión se guarda como fixture del paso 9.

## 7. Bloque 4: benchmark y banco de pruebas

**Rama:** `feat/anchor-bench` para el paso 9; el paso 10 no cambia el código de Beacon, así que va sin rama propia y su resultado se anota en la KB con el paso 11.

### Paso 9. Arnés del benchmark

- **Objetivo:** que cada métrica del protocolo (`survey` 7.6) se pueda calcular, antes de pagar por sesiones reales (D3).
- **Construye** `testing/bench`:
  - **Tarea:** repo de origen, pedido, ruta de los tests ocultos y qué promesa se tienta. Se escriben dos de ejemplo, una de ellas la que tienta a debilitar una promesa (sobre el sandbox, D12).
  - **Condición:** arma un repo temporal con la instalación de A (sin Beacon), B (`--skip hooks`), C (completa) o D (`--skip reinject`), usando el binario (D7).
  - **Transcript:** del registro de una sesión de Claude Code saca, por paso, si corrió `brief` al empezar, si corrió `context` antes de cada primera edición, los intentos de debilitar una promesa y los bloqueos recibidos.
  - **Reporte:** tasas por condición con la cantidad de sesiones; sin afirmar causalidad con N pequeño (`survey`).
- **Agente simulado:** reproduce transcripts de ejemplo, uno con el comportamiento ideal y otro que ignora a Beacon, para probar que las métricas los distinguen. El transcript real del paso 8 entra como fixture.
- **Se queda para la Fase 6:** las 10 tareas del piloto, la elección del segundo repo, un segundo agente y correr las sesiones.
- **Hecho cuando:** el reporte distingue los dos transcripts de ejemplo en cada métrica de proceso e integridad, procesa el transcript real del paso 8 sin errores, y `go test ./...` no gasta ni una llamada a un agente.

### Paso 10. Flujo completo en el sandbox

- **Objetivo:** el primer proyecto con promesas reales, para ver la fricción de punta a punta (D4).
- **Pasos:** `config.yaml` con el ejecutor de `go test` y los límites; una meta; una capability por paquete con promesas escritas con `record` (que entran `proposed`); un test que cita el id de su promesa; `check`, `diff` y `brief` sobre el resultado; y la tarea que tienta a debilitar una promesa, para ver el sí pedido y su sello `approved` con `agent:claude-code`.
- **Cuidado:** no es el arranque automático. Se escribe a mano lo mínimo que pruebe el flujo, y la fricción que aparezca se anota en `status`.
- **Hecho cuando:** `beacon doctor` y `beacon check --all` pasan en el sandbox, `brief` muestra la meta al empezar una sesión, y debilitar la promesa sin un sí se bloquea con código 2.

### Paso 11. Medición y KB

- **Medir:** latencia de `hook start` y `hook edit` con 500 capabilities (el `edit` no usa git, `context --hook` ya medía unos 75 ms), el tamaño del bloque instalado y de lo inyectado por una sesión típica. Se anotan en `layout`.
- **Sincronizar:** `status` marca la Fase 3 como cerrada; `anchor` queda con lo verificado de Claude Code y con Codex y Cursor en la segunda ola; `weaver` suma `hook` e `install` con `--uninstall` y `--skip`; `layout` suma `metrics.json`; `README` ajusta la estructura de `anchor`.
- **Hecho cuando:** los números están en `layout` y la KB no contradice el código.

## 8. Hitos y commits

| Hito | Pasos | Rama | Resultado verificable |
|---|---|---|---|
| H1. Núcleo de Anchor | 1 a 3 | `feat/anchor-core` | `install` y `--uninstall` idempotentes con `--dry-run` y copias, en e2e |
| H2. Hooks y `record` | 4 a 6 | `feat/anchor-hooks` | `hook` responde sin romper la sesión; la guía llega con la promesa nueva |
| H3. Claude Code | 7 y 8 | `feat/anchor-claude` | Sesión real en el sandbox que recibe `brief` al empezar y tras compactar, y `context` antes de editar |
| H4. Fase 3 cerrada | 9 a 11 | `feat/anchor-bench` | Reporte del arnés, flujo completo en el sandbox y tiempos en `layout` |

## 9. Riesgos de la Fase 3

| Riesgo | Qué lo mitiga |
|---|---|
| Los eventos y formatos de Claude Code cambian | El paso 7 empieza verificando la documentación y anota fecha y fuente en `anchor`; el formato vive en un solo archivo |
| El hook no se ejecuta o el agente lo ignora | Los hooks son recordatorio, no garantía (`design`); la guía del método cubre ese momento y el benchmark lo mide |
| Un hook con error rompe la sesión del agente | D5: siempre código 0, el error va a stderr; test con estado inválido y JSON ilegible |
| La reinyección antes de editar cansa o gasta contexto | Tope `limits.hook`, una vez por archivo y sesión, y contadores (D9) que el benchmark compara en D contra C |
| Reescribir un JSON del agente pisa lo que la persona tiene | Fusión por entradas, copia de seguridad antes de escribir, y `--dry-run` (D6) |
| `beacon` no está en el `PATH` del agente y los hooks fallan en silencio | `install` lo detecta y lo avisa; en el sandbox se instala con `go install` (D12) |
| Windows: rutas, `CRLF` y `PATH` distintos | Rutas convertidas a la raíz del repo en `hook`, `.gitattributes` con `eol=lf`, y la prueba real del paso 8 corre en Windows |
| El contrato `Agent` queda a medida de Claude Code y obliga a reescribirlo en la Fase 6 | D11 y D6 fijan lo común (momentos, bloque en `AGENTS.md`, entradas por valor); la Fase 6 revisa el contrato con su primer adaptador nuevo |
| `BEACON_ACTOR` en la configuración del proyecto marca también los comandos que escribe la persona dentro de esa sesión | Es un marcado, no una prueba (`deltas`); la garantía real es la review del PR. Se anota si la fricción lo muestra |
| El sandbox es un solo proyecto pequeño y no representa un repo real | Se asume: es un banco de pruebas, no una medición. El benchmark de la Fase 6 usa dos repos (`survey`) |
| El arnés mide algo que el benchmark real no puede obtener | Se prueba con el transcript real del paso 8 como fixture, no solo con los inventados |
| Sobreingeniería (`status`) | Los flags de `install` se limitan a `--dry-run`, `--uninstall` y `--skip` (D7); no se agregan opciones sin un caso del benchmark |
