# Anchor

Adaptadores que instalan Beacon en cada agente. Son **transporte opaco**: no toman decisiones; toda decisión vive en el binario (`engine`).

## Qué instala

1. **Bloque de guía del método:** unas 20 líneas en inglés (`research`, E2), siete reglas (ver abajo).
2. **Hooks** donde el agente los soporte (ver "Hooks y reinyección"). Son recordatorio, no garantía (`architecture`).
3. **Permiso para `beacon`:** la regla que evita que el agente pida aprobación en cada comando.
4. **Marca del actor:** hace que los comandos del agente corran con `BEACON_ACTOR=agent:<nombre>` (por hook o por la configuración del agente, donde lo permita). Es un marcado, no una prueba (`records`, `changes`).
5. **Comandos** nativos si el agente los admite.
6. **Configuración MCP** del agente, cuando exista el servidor (Fase 5, `engine`).

## Hooks y reinyección

La obediencia a las instrucciones decae a lo largo de la sesión, y reorganizar el archivo de instrucciones no lo arregla (`research`, E7 y E8). Por eso Anchor no se limita a un recordatorio al empezar: donde el agente tenga el evento, reinyecta el estado en los momentos en que se pierde o importa.

| Momento | Qué corre | Por qué |
|---|---|---|
| Inicio y reanudación de sesión | `beacon brief` | El agente empieza con la meta, los bloqueos y los rojos sin pedirlos |
| **Tras compactar el contexto** | `beacon brief` | La compactación borra el `brief`; sin esto el agente sigue sin saber lo que sabía |
| **Antes de editar un archivo** | `beacon context --hook <ruta>` | El contexto que aplica llega justo cuando se necesita, no solo al principio |
| Cierre de sesión | Recordatorio de `beacon check` y `beacon attest` | Que no se cierre sin comprobar |

Reglas de la inyección antes de editar, para que no se vuelva ruido:

- **Una vez por archivo y sesión:** `--hook` recuerda en `.git/beacon/` qué ya inyectó (`storage`); la sesión la identifica el hook con lo que el agente le pase.
- **Con tope** de `limits.hook` líneas (20 por defecto, `records`); lo que no cabe se resume con un aviso, igual que el `brief` (`spider`).
- **Sin salida** cuando no hay nada nuevo que decir.
- Se mide: el binario cuenta cuánto se inyectó y cuántas veces, y el benchmark compara con y sin esta reinyección (`status`).

**Qué eventos existen varía por agente y cambia rápido.** Codex lista `SessionStart`, `Stop`, `UserPromptSubmit`, `PreCompact`, `PostCompact` y `SessionEnd`; para Claude Code y Cursor hay que verificar al implementar cada adaptador qué hook sirve antes de editar y tras compactar (el hook de inicio de Claude Code distingue el arranque, la reanudación y la compactación según su documentación, a confirmar). Si un agente no tiene el evento, ese momento queda cubierto solo por la guía del método.

## Instalación

- Bloques **gestionados e idempotentes**: marcados, se reescriben sin duplicarse.
- Al actualizar o desinstalar, borra **solo lo que coincide byte a byte** con lo instalado; lo que la persona editó se conserva.
- `beacon install` detecta qué agentes hay en el repo.
- **`--dry-run`:** muestra qué archivos crearía o cambiaría y cómo quedarían, sin tocar nada.
- **Copia de seguridad:** antes de modificar un archivo existente guarda una copia en `.git/beacon/backups/` (`storage`). Así una instalación que sale mal se deshace y no depende de que el archivo estuviera en git.
- **Solo toca archivos del repo** (`CLAUDE.md`, `AGENTS.md`, `.codex/`, `.cursor/`, la configuración de permisos del proyecto). Lo que haga falta en la configuración de usuario del agente lo **imprime** para que la persona decida.
- Termina con un **resumen** de lo cambiado y de lo que el agente pide revisar, por ejemplo los hooks nuevos de Codex, que hay que confiar desde `/hooks`.

## Bloque de guía del método

Es el texto que se instala en el archivo de instrucciones de cada agente. Usa los comandos del CLI; cuando se instale el servidor MCP, Anchor reemplaza los comandos por los nombres de las herramientas.

```markdown
<!-- BEGIN BEACON (managed block: do not edit inside) -->
## Beacon: project memory with checked promises

Beacon keeps the project's goal, promises, decisions and open questions in `.beacon/`. Use the `beacon` CLI (add `--json`); never edit those files by hand.

1. **Start.** At the start of a session run `beacon brief`: active goal, blockers, red or drifting promises, work in progress.
2. **Orient.** Before editing code run `beacon context <paths>` for the files you will touch, and respect what it returns. If the request pulls away from the goal, say so in one line.
3. **Don't guess.** If you cannot decide something without inventing it, ask the user (one question at a time) or record it as a question with `beacon record`.
4. **Record what lasts.** Record decisions, questions and new promises (they enter as `proposed`; state a concrete result with concrete inputs and outputs) as they happen; pass long text with `beacon record --json -`. Small changes need no record. If work is too big to resume from the request plus `git diff`, create a `work` with the user's request verbatim.
5. **Promises are the contract.** Tests cite the id of the promise they prove. Never edit, weaken or delete an active promise, a runner or a visual baseline to make something pass: Beacon will ask the user for a yes, and you must wait for it.
6. **Check before you finish.** Run `beacon check` on the capabilities you touched. Report each promise's real state (Holds, Attested, Unverified, Violated, drift). Never call something Holds unless a check passed.
7. **The person delivers.** Never commit, push or open a PR yourself. End with: promises touched, changes to project state, what is pending.

When Beacon blocks something, its message says exactly what to do. Follow it.
<!-- END BEACON -->
```

## Agentes

**Primera ola:** Claude Code, Codex, Cursor. **Segunda ola:** OpenCode, Gemini CLI, Copilot.

| Agente | Instrucciones | MCP (Fase 5) | Notas |
|---|---|---|---|
| Claude Code | `CLAUDE.md`; lee `AGENTS.md` solo si no hay `CLAUDE.md` (desde v2.1.277, no en Bedrock, Vertex ni Foundry) | Sí | Poner el bloque en `AGENTS.md` y que `CLAUDE.md` lo importe con `@AGENTS.md` |
| Codex | `AGENTS.md` (tope de 32 KiB en el primer turno) | `config.toml` | Hooks soportados, ver abajo |
| Cursor | `AGENTS.md`, `CLAUDE.md`, `.cursor/rules/*.mdc` | `mcp.json` (tope de 40 herramientas) | `hooks.json` |
| OpenCode | `AGENTS.md` primario, `CLAUDE.md` de respaldo; campo `instructions` | Sí | — |
| Gemini CLI | `GEMINI.md` (configurable) | Sí | Hooks desde v0.26 |
| Copilot | `.github/copilot-instructions.md` | Sí | — |

Las capacidades de cada agente cambian rápido: verificar contra su documentación al implementar cada adaptador, incluido su mecanismo de permisos (paso 3 de "Qué instala"), cómo pasar `BEACON_ACTOR` a sus comandos (paso 4) y qué eventos de hook existen (sección "Hooks y reinyección"). Además, el bloque de guía se prueba **agente por agente**, porque el cumplimiento de las mismas instrucciones varía entre agentes (`research`, E8).

### Hooks de Codex

Verificado el 2026-11-22 en la documentación de Codex (developers.openai.com/codex/hooks):

- Se declaran en `<repo>/.codex/hooks.json` o en tablas `[hooks]` de `<repo>/.codex/config.toml`; también existen las versiones de usuario en `~/.codex/`.
- Eventos útiles para Beacon: `SessionStart` (con matcher `startup|resume`), `Stop`, y para la reinyección `PreCompact` y `PostCompact`. También existen `UserPromptSubmit` y `SessionEnd`. No verifiqué si hay un evento antes de editar archivos.
- Los comandos corren con el directorio de la sesión como cwd. Conviene resolver la ruta desde la raíz de git (`git rev-parse --show-toplevel`), porque Codex puede lanzarse desde un subdirectorio.
- Codex pide revisar y confiar los hooks nuevos o modificados desde `/hooks`. Instalar o actualizar el hook de Beacon exige ese paso, y `beacon install` debe avisarlo.
- `SessionStart` puede correr antes de que el servidor MCP esté listo, así que el hook llama al **CLI** (`beacon brief`), no a la herramienta MCP.
- Un usuario reportó (issue #35863 del repo de Codex) que en Codex Desktop un hook de `SessionStart` aparece en la interfaz pero no se ejecuta. Es coherente con la regla de que los hooks son recordatorio y no garantía (`architecture`).