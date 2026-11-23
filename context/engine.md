# Engine

El binario `beacon`: CLI y servidor MCP sobre Spider. Toda decisión determinista vive aquí (`architecture`).

## Interfaces

- **CLI: la interfaz principal.** Se construye primero y todo lo demás se apoya en él: los hooks, Sentry y los agentes sin MCP lo llaman directamente.
- **Servidor MCP: después** (Fase 5), como un envoltorio fino sobre las mismas operaciones. Las garantías no cambian de interfaz (`architecture`).
- **Los hooks** ejecutan el CLI e inyectan el resultado en la sesión donde el agente lo permita: `beacon brief` al empezar y tras compactar el contexto, y `beacon context --hook` antes de editar un archivo (`anchor`). Así el agente recibe el estado sin tener que acordarse de pedirlo, que es lo que más ataca el riesgo central (`architecture`).
- **Payloads largos:** `beacon record --json -` lee el registro por stdin, para evitar problemas de quoting en la shell.
- **Permisos:** el agente suele pedir aprobación para cada comando nuevo, así que Anchor instala la regla que permite `beacon` (`anchor`).
- **Actor:** cada comando toma el actor de `BEACON_ACTOR` o, si no existe, de `human:<user.name de git>` (`records`).

## Salida

Sin `--json`, la salida es texto breve para personas. Con `--json`, **todos los comandos devuelven el mismo sobre**:

```json
{"ok": true, "data": {}, "signals": [], "hint": null}
```

- `data`: el resultado del comando.
- `signals`: las señales de Spider que aplican (`spider`).
- `hint`: el siguiente paso exacto cuando algo bloquea, o `null` (P7).
- Si `ok` es `false`, `data` es `null` y aparece `error: {code, message}`.

**Códigos de salida:** `0` ok, `1` error de uso (argumentos o registro inválido), `2` bloqueado, a la espera de un sí o de una respuesta de la persona, o con un recurso ocupado (por ejemplo una tarea ya reclamada).

**Idioma:** los mensajes, los `hint` y los errores van **en inglés**, igual que las claves y la guía del método (`architecture`).

## CLI

| Comando | Qué hace |
|---|---|
| `beacon brief` | Meta activa, bloqueos y rojos, trabajo en curso, índice de guides, con tope de líneas (`spider`) |
| `beacon context <paths>` | Todo lo que aplica a esos archivos, con señales. Con `--hook`: lo mismo pero solo lo que esta sesión todavía no recibió, con tope de `limits.hook` líneas, y sin salida si no hay nada nuevo (`anchor`) |
| `beacon record` | Crea o actualiza un registro con validación y niveles (`changes`); acepta el registro por stdin con `--json -`; con `--approve` registra el sí de la persona. Al proponer una promesa, responde con la guía de redacción (`contracts`) |
| `beacon close <id>` | Cierra un registro con motivo y lo archiva (`storage`) |
| `beacon get <id>` | Detalle de un registro |
| `beacon find <texto>` | Búsqueda de registros |
| `beacon check [capability]` | Ejecuta los checks (`contracts`). **Sin argumentos, solo los de las capabilities relacionadas con lo que cambió** contra la rama base o sin commitear (por sus `paths`); con un id, esa capability; con `--all`, todo |
| `beacon status` | Estado general: rojos, deriva, bloqueos, pendientes de sí |
| `beacon ready` | Trabajo y tareas sin bloqueos, por rama |
| `beacon claim <task>` | Reclama una tarea (consultivo): escribe el claim local en `.git/beacon/` y `owner` y `branch` en el registro. Si otro worktree o rama ya la reclamó, sale con código 2 y dice cuál (`workflow`, `storage`) |
| `beacon diff` | Cambios de estado desde un commit, por nivel, incluidas las capturas base (`changes`) |
| `beacon verify` | Checks relacionados más revisión independiente (riesgo alto) |
| `beacon attest [--evidence <ruta o enlace>]` | Atestaciones manuales al cerrar la sesión; `--evidence` guarda la referencia a lo revisado (`contracts`) |
| `beacon stale` | Guides y decisiones desactualizadas |
| `beacon render` | Vistas Markdown |
| `beacon fmt` | Formato canónico de los registros |
| `beacon doctor` | Valida el estado, ayuda con conflictos, avisa de guides largas y limpia claims de ramas que ya no existen |
| `beacon migrate` | Migra el esquema |
| `beacon stats` | Métricas locales de fricción |
| `beacon sentry ci` | Todo lo que corre el CI en una llamada (`sentry`) |
| `beacon install [--dry-run]` | Instala en los agentes con copia de seguridad; `--dry-run` solo muestra qué cambiaría (`anchor`) |

## Servidor MCP

Meta: menos de 10 herramientas, por la degradación con muchas (`research`, E5). Ocho:

`beacon_brief`, `beacon_context`, `beacon_record`, `beacon_close`, `beacon_get`, `beacon_find`, `beacon_check`, `beacon_status`.

Lo que el CLI tiene aparte se pliega así, sin agregar herramientas:

- `ready` va dentro de `beacon_brief` y `beacon_status`: es una vista.
- `claim` es `beacon_record` sobre una tarea (`owner` y `branch`): es una actualización de un registro, y el servidor también escribe el claim local.
- `diff` y `stale` salen en `beacon_status`.
- Las aprobaciones y las atestaciones (con su `evidence`) se guardan con `beacon_record`.
- `verify`, `fmt`, `doctor`, `migrate`, `render`, `stats`, `sentry` e `install` quedan solo en el CLI y en CI.
- `context --hook` no tiene equivalente MCP: los hooks llaman al CLI (`anchor`).

El servidor es **residente**: carga el grafo una vez y recarga solo lo cambiado (`storage`). Queda un hueco libre bajo el tope de 10. Devuelve el mismo sobre que el CLI.

## Mensajes (P7)

Cada bloqueo dice exactamente qué hacer, y lo que dice funciona. Un mensaje solo nombra un comando si ejecutarlo resuelve el bloqueo (`principles`). Ese texto va en `hint`.

## Fricción

El binario registra **localmente** en `.git/beacon/` (`storage`) cuándo un bloqueo se resolvió *en banda* (siguiendo el mensaje), *fuera de banda* (por otra vía) o terminó en *callejón sin salida*. También cuenta los síes pedidos por sesión y cuánto texto inyectan los hooks. `beacon stats` lo muestra. **Cero telemetría:** nada sale de la máquina.

## Render

`beacon render` genera vistas Markdown bajo demanda para personas y para agentes sin MCP (resumen del proyecto, una capability con sus promesas, preguntas abiertas). **No se versionan**: se regeneran desde los registros.