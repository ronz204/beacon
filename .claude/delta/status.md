# Estado

Dónde está el proyecto: qué está decidido, qué sigue abierto, qué riesgos hay y en qué orden se construye.

## Decisiones vigentes

El detalle de cada una vive en su archivo; aquí solo el mapa.

| Tema | Decisión | Archivo |
|---|---|---|
| Naturaleza | Método que comprueba con código, agnóstico al agente | `vision` |
| Nombre | **Beacon**, con carpeta `.beacon/` y prefijo de ids `bc-`. Se mantiene pese a la colisión con el producto de Asymptote Labs (ver Riesgos) | `layout` |
| Componentes | Spider, Weaver, Anchor, Snares; módulos en un binario y un repo | `design` |
| Nombres de componentes | Metáfora de la telaraña: Spider, Weaver, Anchor, Snares (Engine y Sentry quedan descartados) | `design` |
| Interfaz | **Solo CLI**, sin servidor MCP. Se reconsidera únicamente si un agente importante no tiene terminal o si se mide latencia real que moleste en repos grandes | `weaver` |
| Salida `--json` | Un sobre uniforme `{ok, data, signals, hint}`; códigos de salida 0 (ok), 1 (error de uso) y 2 (bloqueado o recurso ocupado) | `weaver` |
| `beacon check` | Sin argumentos, solo lo relacionado con los cambios; con un id o con `--all` | `weaver` |
| Idioma de los mensajes | Inglés fijo, igual que las claves y la guía del método | `design` |
| Lenguaje y distribución | Go, binario estático, MIT, sin telemetría; Releases, `go install`, Homebrew y Scoop | `design` |
| Dependencias | `go.yaml.in/yaml/v4` y `spf13/cobra` v1.10.2 para el CLI. Los globs se implementan a mano sobre `regexp`, porque la gramática es pequeña y `doublestar` es inconsistente con `**` dentro de llaves | `layout`, `weaver` |
| Carpetas raíz | `source/` para el código y `testing/` para las pruebas que cruzan el binario. `source/` no protege como `internal`: solo el `internal/` de cada pieza es privado para Go | `README` |
| Fuente de verdad | Archivos en git, cargados en memoria al arrancar; sin caché persistente | `layout` |
| Layout | `records/` (7 carpetas en plural), `archive/` mensual y `config.yaml` | `layout` |
| Sello repetido | `approved` admite `n` (`n: 2`, `n: 3`...), que el binario añade solo cuando un sello sería idéntico al guardado (mismo actor, día y `via`), para que un segundo sí del mismo día se reconozca | `ledger`, `deltas` |
| Estado local | En `.git/beacon/`, no en `.beacon/`: resultados de checks, claims locales, métricas de fricción, lo ya inyectado por los hooks y copias de seguridad. Escritura atómica porque los worktrees lo comparten | `layout` |
| Formato | YAML estricto con `beacon fmt`; guides en Markdown con frontmatter | `layout`, `guides` |
| Comillas | `beacon fmt` entrecomilla solo lo que otra herramienta podría reinterpretar | `layout` |
| Librería YAML | `go.yaml.in/yaml/v4` v4.0.0, porque `gopkg.in/yaml.v3` está archivada | `layout` |
| Ids | Aleatorios (`bc-xxxx`, 4 hex de `crypto/rand`), con reintento, crecimiento e hijos jerárquicos | `layout` |
| Esquema | `schema: 1`, `beacon migrate`, campos desconocidos preservados; esquema por tipo definido | `layout`, `ledger` |
| Campos | Una sola palabra, en inglés. Comunes: `id`, `created`, `author`, `closed` con `reason`, `approved`; sin `updated`. Opcionales nuevos: `nongoals` (goal y capability), `rejected` (decision), `evidence` (dentro de `attested`) | `ledger` |
| Relaciones | Campos tipados (`implements`, `blocks`, `parent`, `supersedes`, `serves`, `origin`, `refers`), sin lista genérica | `ledger` |
| Jerarquía de tareas | `parent` apunta siempre al work raíz y la tarea padre sale del id; sin huérfanas, bloqueo heredado, trabajo en las hojas y sin `done` antes que las hijas | `ledger` |
| Globs | Estilo doublestar (`*`, `**`, `?`, `{a,b}`), desde la raíz del repo y sin negaciones; las llaves se expanden antes de comparar, con un máximo de 64 combinaciones | `ledger` |
| Actor | `BEACON_ACTOR` (lo pone Anchor) o `human:<user.name de git>`; `approved: {by, at, via}`. Es un marcado, no una prueba | `ledger`, `deltas` |
| `config.yaml` | Claves: `schema`, `runners` (con `baseline` opcional en los visuales), `review`, `limits` (`guide`, `brief`, `hook`), `agents`. Nivel 3 solo para `runners` y `schema` | `ledger`, `deltas` |
| Carga en memoria | Validada con benchmark en Go: 500 archivos en unos 100 ms, 5.000 en unos 0,5 a 0,8 s | `layout` |
| Consultas a git | En lote (una por commit base y una de frescura), nunca una por promesa; de 5 a 60 ms | `layout`, `spider` |
| Tipos de registro | goal, capability (con promesas), decision, question, guide, work, task | `ledger` |
| Paths | Pista gruesa y opcional, pueden solaparse | `ledger` |
| Guides | Metadatos livianos de una palabra; el `brief` es mínimo, el resto bajo demanda | `guides` |
| Tope del `brief` | `limits.brief` líneas; se recorta por prioridad fija y con aviso | `spider` |
| Frescura | `review` (duración) y `reviewed` (fecha); plazos por tipo de 6 y 12 meses, valores iniciales; también se marca con 10 o más commits sobre sus paths desde la revisión | `guides` |
| Deriva | Compara el commit del último check con el árbol de trabajo, así que incluye cambios sin commitear | `spider`, `pledge` |
| Spider | Estructura activa con señales; el grafo vive en el contexto `thread` | `spider`, `design` |
| Frontera de `diff` | Spider detecta qué cambió y Weaver lo clasifica por nivel | `spider`, `deltas`, `weaver` |
| Checks | Ejecutores en `config.yaml` más parámetros; sin shell | `pledge` |
| Redacción de promesas | Guía, no regla: resultado observable con entradas y salidas concretas, una cosa por promesa, EARS opcional. `beacon record` la devuelve al proponer una promesa | `pledge`, `anchor` |
| Resultados de checks | Locales en `.git/beacon/results.json`, recalculados en CI; sin resultado local es Unverified con razón `not-run-here` | `pledge` |
| Estados | Holds, Attested, Unverified, Violated; `observed` nunca cuenta como Holds | `pledge` |
| Promesas manuales | `attest` solo al cerrar sesión y solo si cambió el código; `--evidence` guarda la referencia a lo revisado | `pledge` |
| UI | Beacon no trae navegador. Cuatro casos: E2E y accesibilidad como ejecutor (Holds), regresión visual contra una captura base (Holds solo como "igual a la base"), "se ve bien" manual con evidencia `observed` (Attested). **Actualizar una captura base es nivel 3** | `pledge`, `deltas` |
| Niveles de cambio | 1, 2 y 3 con regla mecánica y atajo por texto normalizado; incluye las capturas base | `deltas` |
| Aprobación | Sí en el chat transmitido por el agente; en equipo, además la review del PR por una persona (la garantía real) | `deltas`, `snares` |
| Riesgo | Lo calcula el código; el agente solo puede subirlo | `deltas` |
| CI | Comando genérico `beacon snares ci` y una plantilla de GitHub Actions que escucha `pull_request` y `merge_group`; otras plataformas después | `snares` |
| Flujo | Pasos 0 a 9; tamaño por reanudabilidad | `method` |
| Arranque | Escaneo mecánico, todo `proposed`, preguntas de una en una; guide de arquitectura como `draft` con un esqueleto corto inspirado en arc42 | `method` |
| Claims | Consultivos, en dos sitios: local en `.git/beacon/` (visible entre worktrees, avisa al instante) y `owner` y `branch` en el registro (avisan al equipo); `ready` por rama | `method`, `layout` |
| Guía del método | Bloque de unas 20 líneas en inglés, siete reglas, versión CLI | `anchor` |
| Hooks y reinyección | Donde el agente tenga el evento: `beacon brief` al empezar y tras compactar, y `beacon context --hook` antes de editar (una vez por archivo y sesión, con tope `limits.hook`). Los eventos de cada agente se verifican al implementar | `anchor` |
| Instalación | `beacon install` con `--dry-run`, copia de seguridad en `.git/beacon/backups/`, solo archivos del repo y resumen final | `anchor` |
| Agentes | Ola 1: Claude Code, Codex, Cursor. Ola 2: OpenCode, Gemini CLI, Copilot | `anchor` |
| Hooks de Codex | Soportados (`SessionStart`, `Stop`, y `PreCompact` y `PostCompact` para la reinyección); el hook llama al CLI | `anchor` |
| Idioma | Claves, guía del método y mensajes del binario en inglés; contenido de los registros en cualquier idioma | `design` |
| Fricción | Medida solo en local | `weaver` |
| `beacon status` | Entra en la Fase 2, porque el aviso del `brief` lo nombra (P7) | `weaver` |
| `record` | Registro completo por stdin, sin parches; el sí se transmite reenviando el mismo payload con `--approve`, sin estado pendiente | `weaver`, `deltas` |
| Aprobación de `config.yaml` | Clave `approved` sellada por `record --approve` | `ledger`, `deltas` |
| Rama base | `git merge-base HEAD` con `origin/HEAD`, `main` o `master`; `--base` la fuerza | `weaver` |
| Ejecución de checks | Tope fijo de 10 minutos por invocación (`timeout` es `Violated`); con un rojo, `check` sale con código 2. El tope se vuelve configurable si la fricción lo pide | `weaver` |
| `results.json` | Lo guarda Spider, con commit y huella de lo no commiteado, para que un check con cambios sin commitear no muestre deriva | `layout`, `pledge` |
| Archivo | Adelantado a la Fase 2, porque `close` lo necesita; una línea JSON por registro cerrado | `layout` |
| Detalle de Weaver | Aprobado el 2026-11-25: frontera pública de Spider con tipos planos, forma del payload, nivel de cada cambio, promesas que no se borran por omisión, regla del sí fresco y cómo se corren los checks (D15 a D20) | `PLANNING.md` |
| Benchmark de fricción | Protocolo aprobado: condiciones A (sin Beacon), B (solo guía), C (completo) y D (C sin reinyección); piloto de 10 tareas, 3 repeticiones y un agente (90 sesiones); métricas de proceso, integridad, fricción, resultado y costo. El arnés se construye pronto y el benchmark se corre en la Fase 6 | `survey` 7.6 |

## Abierto

**No queda nada de diseño abierto.** Las propuestas R1 a R9 que salieron de la investigación del 2026-11-22 se aprobaron ese mismo día y ya están aplicadas en los archivos de arriba (`survey` las llama "propuestas": ese texto es el de la fecha de la investigación). Lo que sigue son detalles de implementación que se resuelven al construir cada fase:

- Qué hook sirve en cada agente para reinyectar antes de editar y tras compactar, y cómo identificar la sesión (R1).
- El nombre y el comportamiento exactos del evento `merge_group` de GitHub (`snares`).
- Si el nivel 3 de las capturas base resulta demasiado ruidoso; se mide en la fricción y se afina el `baseline`.

## Diferido

Fuera del orden de construcción hasta que se cumpla su condición.

| Qué | Por qué se difiere | Se retoma si |
|---|---|---|
| **Render:** `beacon render`, vistas Markdown bajo demanda para personas (resumen del proyecto, una capability con sus promesas, preguntas abiertas), sin versionar | No es necesario por ahora: los slugs hacen legible `.beacon/`, y `get`, `find` y `status` resumen el estado (`layout`, `weaver`) | Una persona no logra entender el estado con eso, por ejemplo al revisar un PR o al presentar el proyecto |

## Investigación

**Hecha el 2026-11-22** (`survey`, sección 7): PRD, ADR, arc42 y C4; OpenSpec; flujos de equipo; protocolo del benchmark de fricción; Spec Kit, Superpowers y Kiro; Gentle-AI; verificación de UI.

**Lo que queda, y no bloquea la construcción:**

- **Correr el benchmark de fricción.** Solo está diseñado.
- **Gentle-AI en profundidad:** no se leyeron los documentos enlazados desde el README (Review and Architecture, Telemetry, Agents).
- **Probar las herramientas de UI:** Playwright MCP, `agent-browser` y ProofShot se leyeron, no se usaron.
- **Ramas largas en equipo:** probadas en la Fase 2 con casos reales (`method`); sin investigación externa.
- **Verificar al implementar:** los hooks de cada agente, el evento `merge_group` y las fuentes de nivel bajo.

## Riesgos

| Riesgo | Nota |
|---|---|
| **El agente no llama a las herramientas** | El central: las garantías cubren promesas, no que la memoria se use. Los hooks que inyectan el `brief` y el contexto lo reducen (`design`, `anchor`), pero la obediencia decae dentro de la sesión (E7) y los hooks pueden no ejecutarse |
| Sobreingeniería | Se construye todo sin recorte, así que hay que vigilar el alcance |
| Evidencia débil | Fuentes de vendedores, benchmarks sintéticos, hipótesis de cognición; ningún estudio muestra que las specs reduzcan defectos (E9) y Beacon tampoco lo ha mostrado (`survey`) |
| Arranque lento | Con el CLI como única interfaz cada llamada carga el estado, y los hooks de reinyección llaman más veces. Medido con el binario completo y 500 capabilities con 10 promesas: `context` 105 ms, `brief` y `status` 230 ms, `check` unos 0,4 a 0,5 s propios; los repos reales pueden ser más lentos (`layout`) |
| Ruido de los hooks | Inyectar antes de cada edición puede cansar o gastar contexto; por eso hay tope, una vez por archivo y sesión, y se mide (`anchor`) |
| Marcado del actor no infalible | El agente corre con el mismo usuario del sistema; la garantía real es la review del PR por una persona (`deltas`) |
| Fatiga del nivel 3 | Demasiados síes se desactivan (`deltas`); las capturas base pueden sumar (ver Abierto) |
| Registro que se pudre | Promesas y guides sin revisar |
| Calidad de las promesas | Mal redactadas no protegen nada; la guía de redacción ayuda pero no se puede imponer por código (`pledge`) |
| Paths gruesos | Contexto impreciso o deriva no detectada |
| Seguridad | El CI ejecuta código del repo (`snares`) |
| Conflictos en equipo | Mismo archivo editado por dos personas (`layout`); conflictos semánticos que solo ve el CI sobre la fusión (`snares`) |
| Agentes cambiantes | Los adaptadores envejecen rápido (`anchor`) |
| Colisión de nombre | Beacon coincide con el producto de Asymptote Labs (open source, Go, memoria para agentes de código, con financiación), incluido el comando `beacon`; en npm los nombres `beacon` y `beacon-cli` ya están tomados. Se decidió mantener el nombre, así que hay riesgo de confusión y de choque con ese comando en la máquina de quien use ambas herramientas |
| Resultados local vs CI | Una promesa puede pasar en un clon y fallar en CI; la fuente de verdad es el CI (`pledge`) |

## Orden de construcción

Sin recorte de producto; el orden sigue las dependencias. El plan detallado de la Fase 2 (Weaver), cerrada, está en `PLANNING.md`; el de la Fase 1 (Spider) queda en el historial de git.

| Fase | Contenido |
|---|---|
| 0 | **Cerrada:** nombre, benchmarks de carga y de git, esquema, guía del método, decisiones de interfaz y las propuestas de la investigación |
| 1 | **Cerrada el 2026-11-23. Spider:** esquema (con `nongoals`, `rejected` y `evidence`), parser YAML estricto, `fmt`, ids, grafo en memoria, validación, `doctor`, globs |
| 2 | **Cerrada el 2026-11-26. Weaver (CLI):** `--json` con el sobre uniforme y códigos 0, 1 y 2; `brief`, `status`, `context` con `--hook`, `get`, `find`, `ready`, `record` con stdin y `--approve`, `close` con archivo, `claim` con claim local, `check` con ejecutores y `results.json`, y `diff` con niveles (incluidas las capturas base) y pendientes de sí; ramas largas probadas y tiempos medidos |
| 3 | **Anchor sobre el CLI:** guía del método, hooks con reinyección, permisos y marca del actor para Claude Code, luego Codex y Cursor; `beacon install` con `--dry-run` y copias; guía de redacción de promesas; arnés del benchmark; uso del propio Beacon en su desarrollo |
| 4 | **Snares:** pre-commit, `beacon snares ci` y plantilla de GitHub Actions (con `merge_group`); arranque, `attest` con `--evidence`, `verify`; ejecutores de ejemplo para UI |
| 5 | Frescura de guides, `stale` |
| 6 | Segunda ola de agentes, `stats`, correr el benchmark de fricción |