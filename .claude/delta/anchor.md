# Anchor

Adaptadores que instalan Beacon en cada agente. Son **transporte opaco**: no toman decisiones; toda decisión vive en el binario (`weaver`).

## Qué instala

1. **Bloque de guía del método:** unas 20 líneas en inglés (`survey`, E2), siete reglas (ver abajo).
2. **Hooks** donde el agente los soporte (ver "Hooks y reinyección"). Son recordatorio, no garantía (`design`).
3. **Permiso para `beacon`:** la regla que evita que el agente pida aprobación en cada comando. En Claude Code son dos, una por herramienta de shell (`Bash(beacon *)` y `PowerShell(beacon *)`). Un comando compuesto o una tubería hacia `beacon` sigue pidiendo permiso (ver "Hooks de Claude Code").
4. **Nombre del agente para el sello:** no se instala nada. Cada `beacon hook` anota qué agente corrió en el worktree y el binario lo usa como `via` cuando el sí llega sin terminal (`ledger`, sección Actor). Es un marcado, no una prueba (`deltas`).
5. **Comandos** nativos si el agente los admite.
6. **Packs de rules y skills:** rules con `paths:` y skills propias, con el prefijo `beacon-` (`packs`). Aprobado; en construcción, con la instalación y `packs:` hechas (`status`, Fase 4b).

## Hooks y reinyección

La obediencia a las instrucciones decae a lo largo de la sesión, y reorganizar el archivo de instrucciones no lo arregla (`survey`, E7 y E8). Por eso Anchor no se limita a un recordatorio al empezar: donde el agente tenga el evento, reinyecta el estado en los momentos en que se pierde o importa.

| Momento | Qué corre | Por qué |
|---|---|---|
| Inicio y reanudación de sesión | `beacon hook start` (el `brief`) | El agente empieza con la meta, los bloqueos y los rojos sin pedirlos |
| **Tras compactar el contexto** (y tras `/clear`) | `beacon hook start` (el `brief`) | La compactación borra el `brief`; sin esto el agente sigue sin saber lo que sabía |
| **Antes de editar un archivo** | `beacon hook edit` (el `context --hook`) | El contexto que aplica llega justo cuando se necesita, no solo al principio |
| Cierre de sesión | **No se instala.** Recordatorio de `beacon check` y `beacon attest` | `Stop` dispara al terminar cada respuesta, no la sesión, y `attest` llega en la Fase 4. Se retoma entonces |

Reglas de la inyección antes de editar, para que no se vuelva ruido:

- **Una vez por archivo y sesión:** `--hook` recuerda en `.git/beacon/` qué ya inyectó (`layout`); la sesión la identifica el hook con lo que el agente le pase.
- **Con tope** de `limits.hook` líneas (20 por defecto, `ledger`); lo que no cabe se resume con un aviso, igual que el `brief` (`spider`).
- **Sin salida** cuando no hay nada nuevo que decir.
- **Sin consultas a git:** corre antes de cada edición, así que solo lee los registros. La deriva y la frescura llegan por el `brief` al empezar y tras compactar, y por `doctor`.
- Se mide: el binario cuenta cuánto se inyectó y cuántas veces, y el benchmark compara con y sin esta reinyección (`status`).

**Qué eventos existen varía por agente y cambia rápido.** Codex lista `SessionStart`, `Stop`, `UserPromptSubmit`, `PreCompact`, `PostCompact` y `SessionEnd`; para Claude Code y Cursor hay que verificar al implementar cada adaptador qué hook sirve antes de editar y tras compactar (el hook de inicio de Claude Code distingue el arranque, la reanudación y la compactación según su documentación, a confirmar). Si un agente no tiene el evento, ese momento queda cubierto solo por la guía del método.

## Instalación

- Bloques **gestionados e idempotentes**: marcados, se reescriben sin duplicarse.
- Al actualizar o desinstalar, borra **solo lo que coincide byte a byte** con lo instalado; lo que la persona editó se conserva. Los archivos enteros de los packs no son bloques: su huella en el encabezado decide si alguien los editó (`packs`).
- `beacon install` detecta qué agentes hay en el repo.
- **`--preview`:** muestra qué archivos crearía o cambiaría y cómo quedarían, sin tocar nada.
- **Copia de seguridad:** antes de modificar un archivo existente guarda una copia en `.git/beacon/backups/` (`layout`). Así una instalación que sale mal se deshace y no depende de que el archivo estuviera en git.
- **Saltos de línea:** agrega `.beacon/** text eol=lf` a `.gitattributes`, para que en Windows los registros no aparezcan como modificados (`layout`). Con los packs agrega también `.claude/rules/beacon-*` y `.claude/skills/beacon-*/**`, porque desinstalar compara byte a byte (`packs`).
- **Solo toca archivos del repo** (`CLAUDE.md`, `AGENTS.md`, `.gitattributes`, `.claude/rules/`, `.claude/skills/`, `.codex/`, `.cursor/`, la configuración de permisos del proyecto). Lo que haga falta en la configuración de usuario del agente lo **imprime** para que la persona decida.
- Termina con un **resumen** de lo cambiado y de lo que el agente pide revisar, por ejemplo los hooks nuevos de Codex, que hay que confiar desde `/hooks`.

## Bloque de guía del método

Es el texto que se instala en el archivo de instrucciones de cada agente. Usa los comandos del CLI.

```markdown
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
```

## Agentes

**Primera ola (Fase 3):** Claude Code. **Segunda ola (Fase 6):** Codex, Cursor, OpenCode, Gemini CLI, Copilot. Las filas de la tabla de abajo son notas de investigación, a verificar al implementar cada adaptador.

| Agente | Instrucciones | Notas |
|---|---|---|
| Claude Code | `CLAUDE.md`; lee `AGENTS.md` solo si no hay `CLAUDE.md` (desde v2.1.277, no en Bedrock, Vertex ni Foundry) | Poner el bloque en `AGENTS.md` y que `CLAUDE.md` lo importe con `@AGENTS.md` |
| Codex | `AGENTS.md` (tope de 32 KiB en el primer turno) | Hooks soportados, ver abajo |
| Cursor | `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/*.mdc` | `hooks.json` |
| OpenCode | `AGENTS.md` primario, `CLAUDE.md` de respaldo; campo `instructions` | — |
| Gemini CLI | `GEMINI.md` (configurable) | Hooks desde v0.26 |
| Copilot | `.github/copilot-instructions.md` | — |

Las capacidades de cada agente cambian rápido: verificar contra su documentación al implementar cada adaptador, incluido su mecanismo de permisos (paso 3 de "Qué instala") y qué eventos de hook existen (sección "Hooks y reinyección"). Además, el bloque de guía se prueba **agente por agente**, porque el cumplimiento de las mismas instrucciones varía entre agentes (`survey`, E8).

### Hooks de Codex

Verificado el 2026-11-22 en la documentación de Codex (developers.openai.com/codex/hooks):

- Se declaran en `<repo>/.codex/hooks.json` o en tablas `[hooks]` de `<repo>/.codex/config.toml`; también existen las versiones de usuario en `~/.codex/`.
- Eventos útiles para Beacon: `SessionStart` (con matcher `startup|resume`), `Stop`, y para la reinyección `PreCompact` y `PostCompact`. También existen `UserPromptSubmit` y `SessionEnd`. No verifiqué si hay un evento antes de editar archivos.
- Los comandos corren con el directorio de la sesión como cwd. Conviene resolver la ruta desde la raíz de git (`git rev-parse --show-toplevel`), porque Codex puede lanzarse desde un subdirectorio.
- Codex pide revisar y confiar los hooks nuevos o modificados desde `/hooks`. Instalar o actualizar el hook de Beacon exige ese paso, y `beacon install` debe avisarlo.
- El hook de `SessionStart` llama al **CLI** (`beacon brief`).
- Un usuario reportó (issue #35863 del repo de Codex) que en Codex Desktop un hook de `SessionStart` aparece en la interfaz pero no se ejecuta. Es coherente con la regla de que los hooks son recordatorio y no garantía (`design`).

### Hooks de Claude Code

Verificado el 2026-11-26 leyendo la documentación (code.claude.com/docs/en/hooks, `memory`, `permissions` y `env-vars`) a través de resúmenes, y **confirmado el 2026-11-27 con sesiones reales** de Claude Code 2.1.321 (`claude -p --include-hook-events`) sobre un proyecto de prueba en Windows. Lo marcado como confirmado sale de esas sesiones.

- **Registro:** `.claude/settings.json`, `hooks.<Evento>` con una lista de `{matcher, hooks: [{type: "command", command, timeout}]}`. Un matcher de solo letras, dígitos, `_`, `-`, `,` y `|` es una lista de nombres exactos; cualquier otro es una expresión regular sin anclar.
- **Inicio y reanudación:** `SessionStart` entrega `source` con `startup`, `resume`, `clear`, `compact` (compactación manual o automática) y `fork`. Un matcher `startup|resume|compact` cubre el inicio, la reanudación y la reinyección tras compactar con un solo hook.
- **Antes de editar:** `PreToolUse` con matcher `Edit|Write`. Entrega por stdin `session_id`, `cwd`, `tool_name` y `tool_input.file_path`, que es **absoluta y con barras invertidas en Windows**. `MultiEdit` no aparece en lo leído; no se usa.
- **Cómo se agrega contexto:** en `SessionStart`, texto plano por stdout o `{"hookSpecificOutput": {"hookEventName": "SessionStart", "additionalContext": "..."}}`, con tope de 10.000 caracteres. En `PreToolUse`, `hookSpecificOutput` con `hookEventName: "PreToolUse"` y `additionalContext`; para no tocar el permiso de la edición, **no** se envía `permissionDecision`, porque `allow` se salta la confirmación.
- **Códigos de salida:** 0 con stdout vacío no decide nada y la herramienta sigue su flujo normal. El 2 bloquea. Cualquier otro código es un error que no bloquea. Es lo que apoya la regla de que un hook nunca falla la sesión (decisión D5 del plan de la Fase 3, en el historial de git).
- **Cierre:** `Stop` entrega `stop_hook_active` y `last_assistant_message`; con `stop_hook_active` verdadero hay que salir sin hacer nada, para no encadenar continuaciones. Se puede dar guía sin forzar continuar con `additionalContext`. **Sin confirmar:** si el stdout plano de `Stop` llega al modelo o a la persona.
- **Actor:** no se usa la clave `env` de `settings.json`; el nombre del agente sale de su propio hook (`--agent claude-code`).
- **Permiso:** `Bash(beacon *)` (con espacio antes del `*`) permite `beacon` y sus subcomandos, y no `beaconx`. No permite un comando compuesto (`beacon x && otro`), que pide permiso por cada parte. **Confirmado:** en Windows el agente ejecuta con la herramienta `PowerShell`, que `Bash(...)` no cubre; hace falta además `PowerShell(beacon *)`, y `install` escribe las dos.
- **Windows (confirmado):** el hook `beacon hook ... --agent claude-code` se ejecuta y resuelve `beacon` por el `PATH` (Git Bash instalado). Al arrancar Claude Code en un subdirectorio del proyecto **no se cargan los hooks**: toma `.claude/settings.json` del directorio donde arranca. Si la sesión cambia de directorio después, el hook recibe ese `cwd` y `beacon` busca el proyecto hacia arriba. En Git Bash, un `/compact` escrito en un comando se convierte en una ruta: exportar `MSYS_NO_PATHCONV=1`.
- **Instrucciones:** Claude Code lee `AGENTS.md` por sí solo desde la v2.1.277, pero solo cuando no existe ningún `CLAUDE.md`. Con uno, solo lee ese, y `@AGENTS.md` dentro de él lo importa (hasta cuatro saltos; se ignora dentro de bloques de código). Por eso la instalación crea `CLAUDE.md` con el import, y el hook de inicio no repite el bloque de guía.
- **Transcript:** `transcript_path` apunta a un JSONL en `~/.claude/projects/<proyecto>/<sesión>.jsonl`, escrito de forma asíncrona. El esquema de cada línea no está documentado: el arnés del benchmark lo lee de transcripts reales (paso 9 del plan de la Fase 3, en el historial de git).
- **Compactación y reanudación (confirmado):** `/compact` dispara `SessionStart:compact` y `--resume` dispara `SessionStart:resume`; en ambos el hook entregó el `brief` actualizado. El modelo cita el contexto que llega por `SessionStart` y por `PreToolUse:Edit`. `Stop` dispara al terminar **cada respuesta**, y su stdout sigue sin confirmar; por eso Beacon no instala un hook de cierre.
