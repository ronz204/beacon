# Weaver

El binario `beacon`: un CLI sobre Spider. Toda decisión determinista vive aquí (`design`).

## Interfaces

- **CLI: la única interfaz.** Los agentes, los hooks y Snares lo llaman directamente.
- **Sin servidor MCP.** Las garantías viven en el binario y en CI, así que un servidor no añadiría ninguna. Las condiciones para reconsiderarlo están en `status`.
- **Los hooks** ejecutan el CLI e inyectan el resultado en la sesión donde el agente lo permita: `beacon brief` al empezar y tras compactar el contexto, y `beacon context --hook` antes de editar un archivo (`anchor`). Así el agente recibe el estado sin tener que acordarse de pedirlo, que es lo que más ataca el riesgo central (`design`).
- **Payloads largos:** `beacon record --json -` lee el registro por stdin, para evitar problemas de quoting en la shell.
- **Permisos:** el agente suele pedir aprobación para cada comando nuevo, así que Anchor instala la regla que permite `beacon` (`anchor`).
- **Actor:** cada comando toma el actor de `BEACON_ACTOR` o, si no existe, de `human:<user.name de git>` (`ledger`).

## Salida

Sin `--json`, la salida es texto breve para personas. Con `--json`, **todos los comandos devuelven el mismo sobre**:

```json
{"ok": true, "data": {}, "signals": [], "hint": null}
```

- `data`: el resultado del comando.
- `signals`: las señales de Spider que aplican (`spider`), incluida "Pendiente de sí", cuyo nivel asigna Weaver.
- `hint`: el siguiente paso exacto cuando algo bloquea, o `null` (P7).
- Si `ok` es `false`, `data` es `null` y aparece `error: {code, message, details}`. `details` es opcional y lleva lo que explica el fallo, por ejemplo los hallazgos de `doctor` con archivo y línea.

**Códigos de salida:** `0` ok, `1` error de uso (argumentos o registro inválido), `2` bloqueado, a la espera de un sí o de una respuesta de la persona, con un recurso ocupado (por ejemplo una tarea ya reclamada) o con una promesa en rojo tras `beacon check`.

**Idioma:** los mensajes, los `hint` y los errores van **en inglés**, igual que las claves y la guía del método (`design`).

## CLI

| Comando | Qué hace |
|---|---|
| `beacon brief` | Meta activa, bloqueos y rojos, trabajo en curso, índice de guides, con tope de líneas (`spider`) |
| `beacon context <paths>` | Todo lo que aplica a esos archivos, con señales. Con `--hook`: lo mismo pero solo lo que esta sesión todavía no recibió, con tope de `limits.hook` líneas, y sin salida si no hay nada nuevo (`anchor`) |
| `beacon record -` | Crea o reemplaza un registro con validación y niveles (`deltas`). Lee de stdin el registro **completo**, con la misma forma que devuelve `get --json`; no hay parches. Si un cambio pide un sí, sale con código 2 sin guardar, y el agente transmite el sí reenviando el mismo payload con `--approve`. Al proponer una promesa, responde con la guía de redacción (`pledge`) |
| `beacon close <id>` | Cierra un registro con motivo y lo archiva (`layout`) |
| `beacon get <id>` | Detalle de un registro |
| `beacon find <texto>` | Búsqueda de registros |
| `beacon check [capability]` | Ejecuta los checks (`pledge`). **Sin argumentos, solo los de las capabilities relacionadas con lo que cambió** contra la rama base o sin commitear (por sus `paths`); con un id, esa capability; con `--all`, todo. Cada invocación de un ejecutor tiene un tope fijo de 10 minutos; al agotarse, la promesa queda `Violated` con razón `timeout`. Con alguna promesa en rojo sale con código 2 |
| `beacon status` | Estado general: rojos, deriva, bloqueos, pendientes de sí |
| `beacon ready` | Trabajo y tareas sin bloqueos, por rama |
| `beacon claim <task>` | Reclama una tarea (consultivo): escribe el claim local en `.git/beacon/` y `owner` y `branch` en el registro. Si otro worktree o rama ya la reclamó, sale con código 2 y dice cuál (`method`, `layout`) |
| `beacon diff` | Cambios de estado desde la rama base (o desde `--since <rev>`), incluidas las capturas base: Spider los detecta y Weaver los clasifica por nivel (`spider`, `deltas`) |
| `beacon verify` | Checks relacionados más revisión independiente (riesgo alto) |
| `beacon attest [--evidence <ruta o enlace>]` | Atestaciones manuales al cerrar la sesión; `--evidence` guarda la referencia a lo revisado (`pledge`) |
| `beacon stale` | Guides y decisiones desactualizadas |
| `beacon fmt` | Formato canónico de los registros |
| `beacon doctor` | Valida el estado, ayuda con conflictos, avisa de guides largas y limpia claims de ramas que ya no existen |
| `beacon migrate` | Migra el esquema |
| `beacon stats` | Métricas locales de fricción |
| `beacon snares ci` | Todo lo que corre el CI en una llamada (`snares`) |
| `beacon install [--dry-run]` | Instala en los agentes con copia de seguridad; `--dry-run` solo muestra qué cambiaría (`anchor`) |

**Rama base:** `check` y `diff` comparan contra `git merge-base HEAD <base>`, donde `<base>` es `origin/HEAD` si existe y, si no, `main` o `master`. `--base <ref>` la fuerza. Es el punto de origen de la rama, así que también sirve para ramas largas (`method`).

## Mensajes (P7)

Cada bloqueo dice exactamente qué hacer, y lo que dice funciona. Un mensaje solo nombra un comando si ejecutarlo resuelve el bloqueo (`tenets`). Ese texto va en `hint`.

## Fricción

El binario registra **localmente** en `.git/beacon/` (`layout`) cuándo un bloqueo se resolvió *en banda* (siguiendo el mensaje), *fuera de banda* (por otra vía) o terminó en *callejón sin salida*. También cuenta los síes pedidos por sesión y cuánto texto inyectan los hooks. `beacon stats` lo muestra. **Cero telemetría:** nada sale de la máquina.
