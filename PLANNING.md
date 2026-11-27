# Plan de construcción: Fase 4 (Snares)

Este plan cubre solo la fase en curso: en qué orden se construye y qué archivos salen. El detalle funcional vive en la KB (`.claude/delta/`), y todas las fases, en `status`.

La numeración de pasos, decisiones y hitos **se reinicia** en esta fase. Las referencias de la KB y de los commits anteriores apuntan a los planes de las fases cerradas, que están en el historial de git: la Fase 1 (Spider) con `git show 967f5aa:PLANNING.md`, la Fase 2 (Weaver) con `git show 9e5ed52:PLANNING.md` y la Fase 3 (Anchor) con `git show 8457fc7:PLANNING.md`.

> **Estado: por empezar (2026-11-27).** D1 a D19 las aprobó la persona el mismo día. Las verificaciones previas (GitHub, Bun, Python y la versión que estampa Go) están hechas y quedan en los pasos 6, 7 y 8.

## 1. Panorama

### Qué es Snares

Snares es la capa que **no depende de que el agente coopere**: un git hook que corre antes de cada commit y un comando de CI que corre en cada PR y sobre la fusión. Es lo que atrapa un debilitamiento que el agente se saltó, y la garantía en equipo sigue siendo la review del PR por una persona (`snares`, `deltas`).

La fase trae también lo que Weaver todavía no tiene (`attest` y `verify`, con el riesgo) y lo que hace que Beacon sea fácil de empezar en un repo existente (`beacon init`). Por eso son cuatro bloques y no uno.

| Tema | Archivo de la KB |
|---|---|
| Pre-commit, CI, fusión, equipos, seguridad | `.claude/delta/snares.md` |
| Atestaciones, estados, promesas manuales y de UI, ejecutores | `.claude/delta/pledge.md` |
| Niveles, aprobación, riesgo y revisión independiente | `.claude/delta/deltas.md` |
| Arranque en un repo existente, `work` y `task` | `.claude/delta/method.md` |
| Comandos y sobre `--json` | `.claude/delta/weaver.md` |
| Campos (`attested`), `config.yaml`, esquema | `.claude/delta/ledger.md` |
| Estado local, formato y mediciones | `.claude/delta/layout.md` |
| Qué instala Anchor y la guía del método | `.claude/delta/anchor.md` |

### Alcance

**Entra en la Fase 4** (`status`):

| Qué | Cómo queda al cerrar la fase |
|---|---|
| `beacon record --file <ruta>` | El agente pasa el payload por un archivo, sin tubería (D9) |
| `beacon attest [promesas] [--evidence <ref>]` | Sin argumentos lista las promesas manuales cuyo código cambió; con ids las atestigua (D11) |
| `beacon verify` | Corre los checks relacionados, calcula el riesgo y, si es alto, arma el paquete de revisión independiente (D2, D12) |
| `beacon snares pre-commit` | Bloquea el commit con un nivel 3 sin sello, un registro inválido o un check en rojo (D3, D15) |
| `beacon snares ci` | Recalcula los checks, usa los ejecutores de la base, verifica los síes y escribe el comentario del PR (D8, D13, D14) |
| `beacon install` | Instala además el pre-commit (D3) y, con `--ci github`, la plantilla de GitHub Actions (D5) |
| `beacon init` | Escaneo mecánico de Go, Bun, Node y Python; escribe `config.yaml`, la pregunta de la meta y las capabilities candidatas (D6, D7) |
| Guía y ejemplos de UI | El bloque de guía suma `attest` y `--file`; los ejecutores de ejemplo para UI quedan documentados (D16, D17) |
| Beacon en Beacon | Este repo con `init` e `install` y su meta y capabilities reales (D10) |

**Queda fuera:**
- La segunda ola de agentes (Fase 6) y el hook de cierre `Stop`, que sigue siendo ruido (`anchor`).
- `stale` y la frescura de guides (Fase 5), y el lint de tests que citan promesas inexistentes (`pledge`), que se evalúa con ellos.
- `stats` y correr el benchmark de fricción (Fase 6).
- Etiquetas, Releases, Homebrew y Scoop (distribución, sin fase todavía; la persona agregará las etiquetas más adelante) y plantillas de CI distintas de GitHub (`snares`).

### Qué existirá al terminar

| Al cerrar | Qué se puede hacer |
|---|---|
| **Bloque 1: atestar y verificar** | `record --file`, `attest` y `verify` con riesgo y paquete de revisión |
| **Bloque 2: Snares** | Un commit con un nivel 3 sin sello se bloquea en local, y un PR igual se bloquea en el CI, con el comentario clasificado por nivel |
| **Bloque 3: arranque** | `beacon init` deja un repo existente listo para empezar a preguntar, sin inventar nada |
| **Bloque 4: uso propio y cierre** | Beacon se usa en su propio repo, con tiempos medidos y la KB sincronizada |

### Cómo fluye la información

```text
router (handle/*)          lee argumentos, stdin y archivos; imprime el sobre
        │
        ├── attest, verify, record ───────→ weaver.go     clasifica, comprueba y sella
        ├── init ─────────────────────────→ spider.go     escanea y escribe registros
        ├── snares pre-commit | ci ───────→ snares.go ──→ weaver.go ──→ spider.go
        └── install (pre-commit, --ci) ───→ anchor.go     escribe archivos del repo y de .git

snares.go                  la única puerta de Snares: orquesta, no calcula
        │
        ▼
action/*                   precommit, ci, comment
        │
        └──→ weaver.go     pendientes de sí, checks, riesgo; spider.go para el estado
```

### Quién llama a quién

- La dependencia sigue siendo `router → snares → weaver → spider` (`README`). **Snares no calcula nada nuevo**: junta lo que Weaver y Spider ya saben (pendientes de sí, checks, validación) y decide si el commit o el PR pasa.
- **Anchor sigue siendo el único que escribe archivos de agentes, de git y de CI.** El pre-commit y la plantilla de GitHub los instala `install`, no Snares (D3, D5). Snares conoce el comando; Anchor, el archivo.
- `init` va por **Spider** (escribe registros) y no por Anchor: Anchor nunca toca `.beacon/` (Fase 3, D10).
- Un git hook se salta con `git commit --no-verify`: el pre-commit es un recordatorio fuerte, no la garantía. **La garantía es el CI y la review del PR** (`deltas`).

## 2. Decisiones

D1 a D10 se cerraron con la persona el 2026-11-27, y D11 a D19, que eran propuestas del plan, las aprobó el mismo día.

### D1. Todo en la Fase 4, en cuatro bloques

- **Decisión:** pre-commit, `snares ci`, el arranque, `attest` y `verify` van en una sola fase, ordenados por dependencias: atestar y verificar, Snares, arranque, uso propio.
- **Por qué:** `snares ci` verifica síes y riesgo, que Weaver calcula; y el arranque se prueba mejor cuando el CI ya existe.
- **Afecta a:** todos los pasos.

### D2. `verify` corre los checks, calcula el riesgo y arma un paquete de revisión

- **Decisión:** `beacon verify` corre los checks relacionados, calcula el riesgo (D12) y, si es alto, imprime un **paquete de revisión**: las promesas tocadas con su estado y el diff clasificado por nivel, para que lo revise un agente con contexto limpio o una persona. **No lanza ningún agente.**
- **Por qué:** lanzar un revisor ataría el binario a un agente y gastaría tokens; el paquete funciona con cualquiera.
- **Afecta a:** paso 3.

### D3. El pre-commit lo instala `install` y bloquea lo que debilita

- **Decisión:** `beacon install` instala el pre-commit como un bloque gestionado en el hook de git (con copia de seguridad y `--uninstall`). Bloquea el commit si hay un cambio de nivel 3 sin sello, un registro inválido o un check relacionado en rojo. La deriva solo avisa. `--skip precommit` lo omite.
- **Por qué:** una sola instalación para el agente y para git, y `--no-verify` queda como la salida de emergencia de git, no de Beacon.
- **Afecta a:** pasos 4 y 5.

### D4. El CI instala `beacon` con `go install`

- **Decisión:** la plantilla instala el binario con `go install github.com/ronz204/beacon/cmd/beacon@<versión>`. No se publican Releases en esta fase.
- **Detalle (verificado el 2026-11-27):** `install --ci github` escribe como versión la del binario que lo ejecuta, con `runtime/debug.ReadBuildInfo`. Con Go 1.24 o más, una compilación local dentro de un repo git **ya no da `(devel)`**: estampa una pseudoversión (`v0.0.0-20261127145601-c96620ea8bf8`, medida sobre el binario de este repo), que `go install` acepta si ese commit está en el remoto, y con una etiqueta (`go install ...@v1.2.3`) da la etiqueta. Si termina en `+dirty` no se puede instalar: se usa `latest` y el resumen avisa. La persona agregará etiquetas y Releases más adelante; con ellas la versión queda fija sin cambiar nada de esto.
- **Afecta a:** paso 7.

### D5. La plantilla de GitHub se entrega con `beacon install --ci github`

- **Decisión:** un flag escribe `.github/workflows/beacon.yml` como archivo gestionado (copia de seguridad y `--uninstall`). Sin el flag, `install` no toca nada de CI.
- **Por qué:** cambiar el CI es intrusivo y se pide a propósito.
- **Afecta a:** paso 7.

### D6. `beacon init` reconoce Go, Bun, Node y Python

- **Decisión:** el arranque (`method`) es un comando, `beacon init`, con un escaneo mecánico del stack, los paquetes y el comando de test de **Go, Bun (el principal), Node y Python**. Sirve para **proponer los ejecutores** (`runners` de `config.yaml`, que son los comandos que corren los checks) y las carpetas candidatas a capability.
- **Los ejecutores entran con el sí de siempre:** son de nivel 3 porque corren código en el CI. `init` crea `config.yaml` solo con `schema: 1` y muestra los propuestos; la persona los aprueba con `record --approve` de tipo `config` (Fase 2, D13).
- **Afecta a:** pasos 8 y 9.

### D7. Qué escribe `init`

- **Decisión:** `config.yaml` con `schema: 1`; una `question` abierta para la meta; y una capability **sin promesas** por cada carpeta candidata (tope de 10; el resto se resume en una sola pregunta), cada una con su `question`. **Nada queda `active`.**
- **Sin inventar:** el campo `intent` es obligatorio (`ledger`), así que `init` escribe un hecho verificable (`Code in <ruta>; its purpose is an open question.`) y deja la pregunta para la persona.
- **Afecta a:** paso 9.

### D8. `snares ci` recalcula todo; `--scope touched` es opcional

- **Decisión:** `beacon snares ci` corre por defecto **todos** los checks (la fuente compartida de verdad, `pledge`). `--scope touched` limita a las capabilities tocadas por el PR, para suites lentas.
- **Afecta a:** paso 6.

### D9. `record --file`

- **Decisión:** `beacon record --file <ruta>` lee el payload de un archivo. `record -` sigue igual. Pasar los dos o ninguno es un error de uso.
- **Por qué:** en Windows el agente usa PowerShell, y una tubería hacia `beacon` no la cubre la regla de permiso `PowerShell(beacon *)`: lo comprobó el paso 10 de la Fase 3.
- **Afecta a:** paso 1.

### D10. Beacon se usa en su propio repo, al final

- **Decisión:** el último paso de la fase corre `init` e `install` sobre este repo, con la meta y las capabilities reales (Spider, Weaver, Anchor, Snares y el router). Los síes los da la persona.
- **Por qué:** es la primera prueba en un repo real, y D4 de la Fase 3 la pospuso hasta que el arranque existiera.
- **Afecta a:** paso 11.

### D11. `attest`: un comando que lista y atestigua

- **Decisión:** `beacon attest` sin argumentos lista las promesas **manuales cuyo código cambió** desde la base (`pledge`), con una pregunta de sí o no por cada una, que el agente hace de una en una. `beacon attest <promesa>... [--evidence <ref>]` las atestigua. El sello usa las mismas reglas que `approved` (D13 de la Fase 3): `by` es la persona y `via` es el agente si el sí llega sin terminal. `attested` suma `via` (opcional) a `{by, at, commit, evidence}`.
- **`commit`:** el commit `HEAD` del momento. Si hay cambios sin commitear en los `paths` de la promesa, la atestación queda con deriva, como cualquier resultado (`pledge`).
- **Afecta a:** paso 2.

### D12. Cómo se calcula y se muestra el riesgo

- **Decisión:** el riesgo es **alto** si hay un cambio de nivel 3 (incluye ejecutores, esquema y capturas base) o se toca una capability con promesas en rojo o con deriva; si no, **bajo** (`deltas`). `verify` lo calcula y `diff` lo muestra en una línea. El agente puede subirlo (`verify --risk high`), nunca bajarlo.
- **Código de salida de `verify`:** 2 si algún check queda en rojo; con riesgo alto sin rojos sale 0 y el `hint` pide la revisión. La revisión nunca hace commit ni push.
- **Afecta a:** paso 3.

### D13. En el CI los ejecutores son los de la base

- **Decisión:** `snares ci` lee los `runners` de `config.yaml` de la **rama base** (`git show <base>:.beacon/config.yaml`). Solo usa los de la rama del PR si su `config.yaml` trae un `approved` fresco (la regla del sí fresco, Fase 2, D19). Si el PR cambia los ejecutores sin sello, las promesas que dependen de ellos **no se ejecutan** y el comando falla con código 2 y `runners_unapproved`.
- **Por qué:** un PR no puede ejecutar ejecutores que él mismo introduce (`snares`).
- **Afecta a:** paso 6.

### D14. Cómo verifica los síes el CI

- **Decisión:** reutiliza los pendientes de sí de `diff` (Fase 2, D19) calculados contra la base del evento: en `pull_request`, el merge-base; en `merge_group`, la base que trae el propio evento, pasada con `--base`. Un pendiente sale con código 2 y el comentario lo lista. Es un marcado: la garantía real es la review (`deltas`).
- **Afecta a:** paso 6.

### D15. El pre-commit compara con `HEAD` y corre solo lo relacionado

- **Decisión:** compara el árbol de trabajo con `HEAD` (lo que el commit está por sumar, sin importar qué se añadió al índice) y corre los checks de las capabilities tocadas respecto de la base (`check` sin argumentos), no todos. Un check en rojo, un nivel 3 sin sello o un registro inválido salen con código 2. Es un cálculo parecido al del CI, con menos alcance, para que el commit no tarde.
- **Límite anotado:** comparar con el árbol de trabajo y no con el índice puede bloquear por un archivo que no se va a commitear; se acepta por simplicidad y se afina con el uso.
- **Afecta a:** paso 4.

### D16. La guía suma `attest` y `--file`

- **Decisión:** el bloque de guía del método (`anchor`) cambia dos frases: la regla 4 recomienda `beacon record --file <ruta>` para textos largos, y la regla 7 pide, al terminar, preguntarle a la persona por las promesas manuales con `beacon attest`. Sigue por debajo de 30 líneas. El hook `Stop` sigue sin existir.
- **Afecta a:** paso 10.

### D17. Los ejecutores de UI son documentación

- **Decisión:** los ejemplos de `test-e2e` (Playwright) y `visual` (con su `baseline`) se documentan en la KB y en el `README` con Bun como ejemplo principal; `init` no los propone, porque no puede saber si el proyecto tiene UI ni dónde guarda sus capturas.
- **Afecta a:** paso 10.

### D18. `init` no escribe la clave `agents`

- **Decisión:** con un solo agente soportado la clave no informa nada, y Anchor no toca `.beacon/`. Se retoma en la Fase 6, cuando haya más de un adaptador.
- **Afecta a:** paso 9.

### D19. `snares ci` entrega un comentario y un resumen

- **Decisión:** `--comment <ruta>` escribe el comentario en Markdown (el diff por nivel, los pendientes, los rojos y el riesgo); la plantilla lo publica con `gh pr comment` solo en `pull_request` y lo copia al resumen del job. El binario no habla con la API de GitHub.
- **Afecta a:** pasos 6 y 7.

## 3. Estructura objetivo

Lo que **se agrega o cambia** en la Fase 4. Los números son el paso que crea cada parte, y los tests van junto a cada archivo.

```text
beacon/
├─ source/
│  ├─ router/internal/handle/
│  │  ├─ record.go                    # 1 · --file
│  │  ├─ attest.go                    # 2
│  │  ├─ verify.go                    # 3
│  │  ├─ snares.go                    # 4 · pre-commit · 6 · ci
│  │  ├─ install.go                   # 5 · --skip precommit · 7 · --ci github
│  │  └─ init.go                      # 9
│  │
│  ├─ spider/
│  │  ├─ attest.go                    # 2 · promesas manuales y escritura de `attested`
│  │  ├─ base.go                      # 6 · los runners de una revisión
│  │  ├─ scan.go                      # 8 · Scan
│  │  ├─ init.go                      # 9 · Init
│  │  └─ internal/
│  │     ├─ action/
│  │     │  ├─ attest.go              # 2
│  │     │  └─ init.go                # 9
│  │     └─ driver/
│  │        ├─ record/                # 2 · Draft.Attest
│  │        └─ scan/                  # 8 · Go, Bun, Node, Python
│  │
│  ├─ weaver/
│  │  ├─ attest.go                    # 2
│  │  ├─ verify.go                    # 3
│  │  └─ internal/
│  │     ├─ domain/deltas/
│  │     │  └─ risk_policies.go       # 3 · RiskOf
│  │     └─ action/
│  │        ├─ attest.go              # 2
│  │        └─ verify.go              # 3
│  │
│  ├─ snares/
│  │  ├─ snares.go                    # 4 · fachada pública de la pieza
│  │  ├─ views.go                     # 4
│  │  └─ internal/action/
│  │     ├─ precommit.go              # 4
│  │     ├─ ci.go                     # 6
│  │     └─ comment.go                # 6 · Markdown del PR
│  │
│  └─ anchor/internal/
│     ├─ agents/git/                  # 5 · el hook de pre-commit en el directorio de git
│     ├─ agents/github/               # 7 · .github/workflows/beacon.yml
│     └─ assets/
│        └─ beacon.yml                # 7 · plantilla
│
└─ testing/integration/
   ├─ record_test.go                  # 1
   ├─ attest_test.go                  # 2
   ├─ verify_test.go                  # 3
   ├─ precommit_test.go               # 4 y 5
   ├─ ci_test.go                      # 6 y 7
   └─ init_test.go                    # 8 y 9
```

### Dónde va cada tipo de código

| Si el código... | Va en |
|---|---|
| Calcula el riesgo o el nivel de un cambio | `weaver/internal/domain/deltas` |
| Escribe `attested` en un registro | Spider (`driver/record`, `action/attest.go`) |
| Decide si un commit o un PR pasa | `snares/internal/action` |
| Lee la estructura del repo para el arranque | `spider/internal/driver/scan` |
| Escribe un archivo de git o de CI | Anchor (`agents/git`, `agents/github`) |
| Arma el Markdown del comentario | `snares/internal/action/comment.go` |

### Quién puede importar a quién

```text
cmd/beacon → router → snares.go → action → weaver.go → spider.go
                │                  
                ├──→ weaver.go     (attest, verify, record)
                ├──→ spider.go     (init)
                └──→ anchor.go     (install)
```

`snares` solo importa `weaver` y `spider` por sus paquetes raíz. `anchor` no importa ninguno de los dos.

## 4. Bloque 1: atestar y verificar

**Rama:** `feat/weaver-attest`, desde `main` actualizado (en la práctica, desde la rama del plan, que aún no estaba fusionada). **Forma de trabajo:** primero el test que falla, luego el código mínimo, y al final `go test -count=1 ./...` en verde (los tests de integración se corren sin caché).

### Paso 1. `record --file` (hecho)

- **Objetivo:** que el agente pase el payload sin tubería (D9).
- **Construye:** el flag `--file` en `handle/record.go`. Lee el archivo con la ruta relativa al directorio actual; `-` y `--file` a la vez, o ninguno, dan `record_source_invalid`, con un `hint` que dice cuál usar.
- **Hecho cuando:** `record --file` hace lo mismo que `record -` con el mismo payload (incluido `--approve`), un archivo que no existe falla con un error claro, y el e2e cubre los tres casos.
- **Cómo quedó:** `record [-]` acepta un argumento opcional. Pasar `-` y `--file` a la vez, o ninguno, da `record_source_invalid` (código 1), y un archivo ilegible da `record_file_unreadable` con la ruta en `details`. El flujo de después es el mismo: `--approve` y los niveles funcionan igual con un archivo. La guía del método se actualiza en el paso 10.

### Paso 2. `attest` (hecho)

- **Objetivo:** que una promesa manual pueda llegar a `Attested` y volver a la deriva cuando cambia su código (D11).
- **Construye:**
  - En Spider, `ManualPromises(ctx, base)` (las manuales de las capabilities tocadas desde la base y sin atestación vigente) y `Attest(ctx, promesa, sello, commit, evidencia)`, que escribe `attested` con el mismo flujo de `Draft` que `record` (fusión sobre el archivo, escritura atómica, candado).
  - En Weaver, `Attester`: sin ids devuelve la lista; con ids atestigua, con el sello de `Whoami` (persona o agente) y el `HEAD` del momento.
  - El handler `beacon attest` y el sufijo `via` en `attested` (`ledger`).
- **Cuidado:** solo se atestiguan promesas **manuales**; una promesa automática falla con `attest_not_manual` y un `hint` que manda a `beacon check`. Una atestación **no es un cambio de nivel 3** (no debilita nada).
- **Hecho cuando:** una promesa manual tocada aparece en la lista; al atestiguarla pasa a `Attested` con su evidencia; editar un archivo de sus `paths` la vuelve a mostrar con deriva; y la persona y el agente quedan en `by` y `via`.
- **Cómo quedó:**
  - **Dominio:** `PromiseAttestation` suma `via` (opcional, siempre un agente) y el lector y el formateador de `attested` lo conocen: el orden canónico es `by`, `at`, `via`, `commit`, `evidence`.
  - **Spider:** `ManualPromises(revisión)` lista las promesas manuales y `active` de las capabilities tocadas que no están atestiguadas o tienen deriva, y `Attest` escribe `attested` con el mismo `Draft` y el mismo candado que `record`. `Project.Head` da el commit corto.
  - **Una atestación exige código commiteado:** si los `paths` de la capability tienen cambios sin commitear, `Attest` falla con `attest_uncommitted`. Esto **corrige a D11**, que decía que la atestación quedaba con deriva: un resultado de check guarda la huella de lo no commiteado, pero una atestación solo guarda un commit, y atestiguar antes de commitear habría dejado la promesa con deriva para siempre. El orden es: la persona commitea, y entonces se atestigua.
  - **Weaver:** `Attester` usa el sello de siempre (`StampApproval`): `by` es la persona y `via` el agente cuando el sí llega sin terminal. `beacon attest` sin ids lista y deja un `hint` que le dice al agente que pregunte de una en una.
  - **Errores:** `attest_not_manual` (promesa automática), `attest_uncommitted` y `record_missing`, cada uno con su `hint`.

### Paso 3. Riesgo y `verify` (hecho)

- **Objetivo:** que el riesgo lo calcule el código y que la revisión independiente tenga qué leer (D2, D12).
- **Construye:**
  - `deltas.RiskOf` (dominio): alto si hay un hallazgo de nivel 3 o una capability tocada con promesas en rojo o con deriva; devuelve también las razones.
  - `action/verify.go`: reutiliza `Checker` para los checks relacionados, calcula el riesgo, y con riesgo alto arma el paquete: por promesa tocada su texto y su estado, el diff por nivel y las instrucciones para el revisor (solo promesas y diff, nunca commit ni push).
  - `beacon verify [--risk high] [--base <ref>]`, y una línea de riesgo en `beacon diff`.
- **Hecho cuando:** un cambio de nivel 3 da riesgo alto con su razón; uno de mantenimiento, bajo; `--risk high` sube y nunca baja; un check en rojo sale con código 2; y el paquete lista cada promesa tocada.
- **Cómo quedó:**
  - **Dominio:** `deltas.AssessRisk(hallazgos, zonas, subido)` devuelve `low` o `high` y sus razones, sin repetirlas: `weakening_change`, `red_promise`, `drifted_promise` y `raised_on_request`. Una zona es una capability tocada con alguna promesa `violated` o con deriva (las manuales atestiguadas incluidas).
  - **`diff`** calcula el riesgo con las zonas del estado actual y lo muestra en la primera línea (`risk: high (weakening_change)`) y en `data.risk`.
  - **`verify`** corre primero los checks relacionados y después calcula el riesgo, para que una promesa que acaba de arreglarse no cuente como roja. Con riesgo alto, `data.review` trae las promesas tocadas (las comprobadas y las manuales pendientes) con su texto y su estado, el diff clasificado y cuatro reglas para el revisor. El `hint` pide la revisión; con un rojo, el `hint` es el de `check` y el código es 2.
  - **Cambio necesario:** `TouchedCapabilities` solo aceptaba un commit en hexadecimal, y `diff --since HEAD~1` se rompió al necesitarlo. Ahora resuelve cualquier nombre con `rev-parse` y rechaza lo que parezca una opción.
  - **Los textos de las promesas** llegan a `check` por `Analysis` (`PromiseState.Text`), sin duplicar la lectura.

## 5. Bloque 2: Snares

**Rama:** `feat/snares-ci`, desde `main` con el Bloque 1 fusionado.

### Paso 4. La pieza Snares y `snares pre-commit` (hecho)

- **Objetivo:** que un commit que debilita lo prometido se bloquee sin depender del agente (D3, D15).
- **Construye:** `source/snares` con su fachada (`snares.go`, `views.go`) y `action/precommit.go`, que llama a `Spider.Doctor`, a los pendientes de sí de Weaver contra `HEAD` y a `Check` de las capabilities tocadas. Sale con 0, o con 2 y un `hint` por cada bloqueo. Añade el grupo `beacon snares` al router.
- **Hecho cuando:** en un repo temporal, un nivel 3 sin sello, un check en rojo y un registro inválido salen cada uno con código 2 y su `hint`; la deriva solo avisa; y un commit sin nada de eso sale con 0.
- **Cómo quedó:**
  - `snares.New(project)` junta `Spider.Doctor` y `Weaver` (`Diff` desde `HEAD`, `Check` de lo tocado) detrás de una interfaz `Ports`, y la acción devuelve bloqueos (`invalid_state`, `unapproved_change`, `failed_check`) con su `hint`; el router los junta y sale con 2.
  - **Sin commits todavía:** si no hay `HEAD`, solo valida el estado y avisa. **Sin rama base:** salta los checks y avisa. Un ejecutor que no arranca avisa y no bloquea.

### Paso 5. Instalar el pre-commit (hecho)

- **Objetivo:** que `install` deje el hook funcionando (D3).
- **Construye:** el agente `git` de Anchor, que contribuye un bloque gestionado al hook de pre-commit (`#!/bin/sh` si el archivo es nuevo; `beacon snares pre-commit || exit $?`). La ruta sale de `git rev-parse --git-path hooks/pre-commit`, que respeta `core.hooksPath`; el archivo queda ejecutable; el `FileStore` sabe escribir dentro de la carpeta de git. Suma `--skip precommit`, y `.beacon/** text eol=lf linguist-generated` en el bloque de `.gitattributes` (`snares`).
- **Cuidado:** si el hook ya existe y termina con `exit 0`, el bloque se agrega antes de esa línea; si `core.hooksPath` apunta a una carpeta versionada del repo (husky), el resumen avisa de que se edita un archivo del proyecto.
- **Hecho cuando:** `install` crea el hook ejecutable, un `git commit` real se bloquea con un nivel 3 sin sello y pasa con `--no-verify`, repetir `install` no cambia nada y `--uninstall` deja el hook como estaba.
- **Cómo quedó:**
  - `agents/git` entrega un bloque con `#!/bin/sh` como preámbulo del archivo nuevo; `Planner` suma `Git` (con `--skip precommit` se omite) y `CI`. En la desinstalación, un hook que queda solo con el preámbulo se borra.
  - `blocks.Upsert` inserta el bloque antes de un `exit 0` final, y quitarlo devuelve el archivo igual que estaba. `files.Repo` conoce la ruta del hook (`git rev-parse --git-path`, que respeta `core.hooksPath`), la acepta aunque quede fuera del repo y le pone modo 0755.
  - Si el hook es un archivo del proyecto (husky), el resumen avisa. `.gitattributes` suma `linguist-generated`.

### Paso 6. Ejecutores de la base y `snares ci` (hecho)

- **Objetivo:** que el CI sea la garantía (D8, D13, D14, D19).
- **Construye:**
  - En Spider, `RunnersAt(ctx, rev)`, que lee los `runners` de `config.yaml` en una revisión; y en Weaver, un `Check` que acepta los ejecutores de otra fuente.
  - `action/ci.go`: elige los ejecutores (D13), recalcula los checks (`--all` por defecto, `--scope touched`), verifica los pendientes de sí contra la base, calcula el riesgo y falla con 2 si hay un rojo, un nivel 3 sin sello o ejecutores sin aprobar.
  - `action/comment.go` y `--comment <ruta>`: el Markdown con el diff por nivel.
  - `beacon snares ci [--base <ref>] [--scope all|touched] [--comment <ruta>]`.
- **Hecho cuando:** en un repo temporal con una rama que cambia los ejecutores sin sello, las promesas que dependen de ellos no se ejecutan y sale con 2; con el sello, se usan los de la rama; un nivel 3 sin sello sale con 2; un PR limpio sale con 0; y el comentario lista lo que debe leer el revisor.
- **Cómo quedó:**
  - `Project.RunnersAt` lee `config.yaml` en una revisión (vacío si no existía) y `CheckInput.RunnersAt` hace que el `Checker` use esos ejecutores y **omita** las promesas cuyo ejecutor no existe ahí (`skipped`), en vez de fallar con `runner_missing`.
  - `snares ci` calcula primero el diff contra la base. Si hay un cambio `runners_changed` sin sello, corre con los ejecutores de la base y bloquea con `runners_unapproved`; si lo tiene, usa los de la rama. Los demás pendientes salen como `unapproved_change`.
  - `--base` vacío cae a `origin/HEAD`, `main`, `master`, `origin/main` y `origin/master`. `--comment` escribe el Markdown (riesgo, bloqueos, omitidas, avisos y cambios por nivel).

### Paso 7. La plantilla de GitHub Actions (hecho)

- **Objetivo:** que un proyecto tenga el CI con una línea (D4, D5).
- **Verificado el 2026-11-27** en la documentación de GitHub (se anota en `snares`):
  - **Confirmado:** el evento se llama `merge_group`, su único tipo es `checks_requested`, corre "cuando un pull request se agrega a una cola de fusión", `GITHUB_SHA` es el SHA del grupo y el workflow **debe** incluir `merge_group` como disparador para que los checks requeridos reporten. El evento es distinto de `pull_request` y `push`. `actions/checkout` trae por omisión un solo commit del SHA que disparó el workflow, y `fetch-depth: 0` trae todo el historial.
  - **Para `gh pr comment`:** `--body-file <archivo>` lee el cuerpo de un archivo, el token va en `GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}`, y al declarar `permissions` lo no declarado queda en `none`, así que hacen falta `contents: read` y `pull-requests: write`. El resumen del job se escribe en `$GITHUB_STEP_SUMMARY` (tope de 1 MiB por paso). En un PR desde un fork el token es de solo lectura, así que el comentario falla ahí: la plantilla lo publica solo si el PR es del mismo repositorio.
  - **Sin confirmar:** los campos del payload de `merge_group` (`base_sha`, `base_ref`, `head_sha`) **no aparecen en las páginas leídas**, ni el formato de `GITHUB_REF`. La plantilla usa `github.event.merge_group.base_sha` para `--base` y, mientras no se confirme, `snares ci` cae a `git merge-base HEAD origin/<rama base>` cuando `--base` llega vacío. Queda como riesgo hasta que corra en una cola real.
  - **Versiones:** el README de `actions/checkout` y de `actions/setup-go` muestra `@v7` como la actual (no se consultó la página de releases); se vuelve a mirar al escribir el archivo.
- **Construye:** `agents/github` de Anchor y `assets/beacon.yml`: `pull_request` y `merge_group`, `actions/checkout` con `fetch-depth: 0`, la instalación de `beacon` (D4), `beacon snares ci --base ... --comment ...`, y la publicación del comentario con `gh pr comment` solo en `pull_request`. Suma `install --ci github`.
- **Hecho cuando:** `install --ci github` crea el archivo, repetirlo no cambia nada, `--uninstall` lo quita, el YAML tiene la forma que GitHub espera (se valida con un lector YAML en el test) y la verificación queda escrita en `snares`.
- **No se puede probar aquí:** que el workflow corra de verdad en GitHub. Se prueba el comando que llama y la forma del archivo.
- **Cómo quedó:**
  - `actions/checkout` y `actions/setup-go` siguen en `@v7` (v7.0.1 y v7.1.0, consultado el 2026-11-27 en la API pública de releases).
  - El archivo entero es un bloque gestionado (`# BEGIN BEACON` ... `# END BEACON`) con `{{version}}`, que `install` llena con `runtime/debug`; sin versión usable (`(devel)` o `+dirty`) pone `latest` y el resumen avisa.
  - `--base` es `origin/<base_ref>` en `pull_request` y `merge_group.base_sha` en la cola; si llega vacío, no se pasa. El comentario se publica solo si el PR es del mismo repositorio. `--ci` solo acepta `github`.

## 6. Bloque 3: arranque

**Rama:** `feat/spider-init`, desde `main` con el Bloque 2 fusionado.

### Paso 8. El escaneo mecánico

- **Objetivo:** saber qué hay en un repo sin inventar nada (D6).
- **Verificado el 2026-11-27** con Bun 1.3.14 y Python 3.14 en esta máquina (se anota en `method`):
  - **Bun:** `bun test <ruta> -t "<nombre>"` filtra por archivo y por nombre desde la raíz, `bun test --cwd <paquete>` corre un paquete de un workspace, y un test que falla sale con código 1. Ejecutor propuesto: `test-bun: [bun, test, --cwd, "{pkg}", "{target}"]`, con `{target}` como ruta de archivo relativa al paquete.
  - **Python:** `python -m pytest <archivo>::<test>` elige un test y `-k` filtra por nombre; un fallo sale con 1. Ejecutor propuesto: `test-python: [python, -m, pytest, "{target}"]`. `pytest` no está instalado en esta máquina: la prueba se hizo con `uv run --with pytest`, y el ejecutor da `runner_unavailable` si falta.
  - **Go:** `[go, test, "{pkg}", -run, "{target}"]` (ya en uso). **Node:** `[pnpm, --filter, "{pkg}", test, --, "{target}"]` (el ejemplo de la KB); npm y yarn llevan su propia variante.
- **Construye:** `driver/scan` y `Scan` en Spider: detecta Go (`go.mod`), Bun (`bun.lock`, `bunfig.toml`, `package.json`), Node (`pnpm-lock.yaml`, `yarn.lock`, `package-lock.json`) y Python (`pyproject.toml`, `requirements.txt`); los paquetes de un monorepo (workspaces o módulos); el comando de test de cada uno como **ejecutor propuesto**; y las carpetas candidatas a capability, por tamaño.
- **Cuidado:** solo lee; no escribe nada. Ignora `node_modules`, `.git`, `vendor`, `dist` y las carpetas ocultas.
- **Hecho cuando:** en repos temporales de cada ecosistema, el escaneo devuelve el stack, los ejecutores y los candidatos esperados; un repo sin ninguno devuelve la lista vacía con un `hint`; y un monorepo mixto detecta los dos stacks.

### Paso 9. `beacon init`

- **Objetivo:** dejar el repo listo para preguntar, con valor desde la primera sesión (D6, D7, D18).
- **Construye:** `Init` en Spider y `beacon init [--dry-run]`: escribe `config.yaml` con `schema: 1`, la `question` de la meta, y una capability sin promesas con su `question` por candidato (tope de 10 y una pregunta resumen del resto); muestra los ejecutores propuestos y el comando para aprobarlos (`record --approve` de tipo `config`); falla con `project_exists` si ya hay `.beacon/`.
- **Cuidado:** `init` no instala nada de agentes (`install`) y no escribe `agents`. Los registros usan el actor de siempre (Fase 3, D13).
- **Hecho cuando:** `init` en un repo temporal deja `doctor` en verde, el `brief` ya muestra las preguntas abiertas, los ejecutores propuestos entran con `record --approve`, y `--dry-run` no cambia ningún byte.

### Paso 10. Guía y ejemplos de UI

- **Objetivo:** que el agente conozca lo nuevo y que la UI tenga ejemplos (D16, D17).
- **Construye:** el bloque de guía con las dos frases de D16 (y su test de tamaño); los ejecutores de ejemplo `test-e2e` y `visual` con su `baseline`, con Bun como principal, en `ledger` y en el `README`; y la regla de qué cuenta como `observed` y qué pide `attest`, en `pledge`.
- **Hecho cuando:** el bloque sigue por debajo de 30 líneas, un `install` sobre un repo ya instalado solo reemplaza el bloque de guía, y la KB no contradice los comandos nuevos.

## 7. Bloque 4: uso propio y cierre

**Rama:** `feat/snares-adopt` para el paso 11 (los archivos de este repo y sus registros) y `docs/close-phase-4` para el paso 12.

### Paso 11. Beacon en Beacon

- **Objetivo:** el primer uso en serio, en un repo real (D10).
- **Pasos:** `beacon init` en este repo; revisar los candidatos (Spider, Weaver, Anchor, Snares, router y el kernel) y las preguntas; **la persona responde la meta y aprueba los ejecutores** (`go test`); `beacon install --ci github`; escribir las primeras promesas reales con `record` (entran `proposed`).
- **Cuidado:** los síes los da la persona desde su terminal. El agente no los fabrica, igual que en el paso 10 de la Fase 3. Este repo ya tiene `.claude/settings.json` y reglas propias: `install` fusiona, no pisa.
- **Hecho cuando:** `beacon doctor` y `beacon check --all` pasan en este repo, el `brief` muestra la meta real, el pre-commit bloquea un nivel 3 de prueba, y la fricción que aparezca queda anotada en `status`.

### Paso 12. Medición y KB

- **Medir:** cuánto tarda `snares pre-commit` y `snares ci` con 500 capabilities y 5.000 promesas (el mismo repo sintético de las fases 2 y 3), y `init` sobre un repo grande. Se anota en `layout`, junto a las mediciones anteriores.
- **Sincronizar:** `status` marca la Fase 4 como cerrada y suma las decisiones; `snares` queda con lo verificado de `merge_group`; `weaver` suma `attest`, `verify`, `init`, `snares` y `--file`; `method` con lo verificado de Bun y Python; `pledge` y `ledger` con `attested.via`; `anchor` con el hook de git y la plantilla; el `README` con la estructura de `snares`.
- **Hecho cuando:** los números están en `layout` y la KB no contradice el código.

## 8. Hitos y commits

| Hito | Pasos | Rama | Resultado verificable |
|---|---|---|---|
| H1. Atestar y verificar | 1 a 3 | `feat/weaver-attest` | `record --file`, `attest` y `verify` con riesgo y paquete, en integración |
| H2. Snares | 4 a 7 | `feat/snares-ci` | Un commit y un PR con un nivel 3 sin sello se bloquean; la plantilla se instala y se quita |
| H3. Arranque | 8 a 10 | `feat/spider-init` | `init` deja un repo de cada stack listo para preguntar; la guía conoce `attest` |
| H4. Fase 4 cerrada | 11 y 12 | `feat/snares-adopt`, `docs/close-phase-4` | Beacon en su propio repo y tiempos en `layout` |

## 9. Riesgos de la Fase 4

| Riesgo | Qué lo mitiga |
|---|---|
| El pre-commit tarda y la gente lo salta con `--no-verify` | Corre solo los checks de lo tocado (D15); se mide en el paso 12; la garantía es el CI, no el hook |
| El pre-commit bloquea por un cambio que no se va a commitear (compara con el árbol de trabajo, D15) | Límite anotado; se afina con el uso del paso 11 |
| Un hook existente (husky, `core.hooksPath` versionado) se edita sin querer | El bloque es gestionado, `install` guarda copia, el resumen avisa de que edita un archivo del proyecto y `--skip precommit` lo evita |
| El CI ejecuta código del PR | Ejecutores de la base hasta que se aprueben (D13), sin shell y con parámetros que no empiezan con `-` (`pledge`) |
| El payload de `merge_group` (`base_sha`) no se pudo confirmar en la documentación | `snares ci` cae a `git merge-base` si `--base` llega vacío; se confirma la primera vez que corra en una cola de fusión real |
| No hay etiquetas ni Releases, así que la versión que fija la plantilla es una pseudoversión de commit (D4) | Sirve mientras el commit esté en el remoto; con `+dirty` se cae a `latest` y se avisa. La persona agregará las etiquetas más adelante |
| El escaneo propone mal los ejecutores de Bun o Python | Los comandos se probaron en esta máquina (paso 8) y los ejecutores solo entran con un sí (D6); lo no instalado da `runner_unavailable`, no un rojo |
| `init` genera ruido en un monorepo grande | Tope de 10 capabilities y una pregunta resumen del resto (D7) |
| El agente fabrica un sí o una atestación | Es un marcado, no una prueba (`deltas`); la garantía es la review del PR. Los síes del paso 11 los da la persona |
| La atestación `via` agente se toma por una verificación humana | `attest` solo afirma que una persona dijo sí; la evidencia del agente es `observed` y no cuenta (`pledge`) |
| Sobreingeniería (`status`) | No hay otras plataformas de CI, ni lanzamiento de agentes, ni Releases en esta fase |
