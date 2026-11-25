# Investigación

Lo que sabemos de otras herramientas y de la evidencia, **solo lo que sostiene una decisión actual**. Las fuentes están en `legend`.

**Confianza:** casi todo viene de documentación de proyectos, blogs o estudios de vendedores. No hay evaluaciones independientes de ODD, Engram ni Beads. Sirve para orientar, no para blindar decisiones. La ronda del 2026-11-22 (sección 7) añade cuatro estudios con método explícito, pero ninguno prueba lo que Beacon promete.

## 1. Herramientas relacionadas

| Herramienta | Qué es | Peso | ¿Verifica con código? | Debilidad principal |
|---|---|---|---|---|
| **Spec Kit** (GitHub) | Constitución → spec → plan → tareas → implementar → converger, una rama por spec; extensiones y bundles | Pesado | No que se describa: `converge` y `analyze` son prompts de slash command | Muchos archivos repetitivos; la revisión pesa más que el código (sección 7) |
| **BMAD** | Un equipo ágil simulado con muchas personas de agente | Muy pesado | No | Curva alta; coordinación en el usuario |
| **OpenSpec** | `specs/` (verdad actual por capability) y `changes/` (propuestas con deltas ADDED, MODIFIED y REMOVED) que se fusionan al archivar | Medio | Tiene un paso `verify` opcional; no pude confirmar su mecanismo | El más cercano en filosofía; su estado es Markdown, la fusión de deltas es una skill (prompt) y requiere Node |
| **Kiro** (AWS) | IDE y CLI propios: requirements en EARS, design, tasks, steering y hooks | Pesado | Parcial (según una charla de AWS, genera tests de propiedades desde EARS) | Desproporcionado en tareas pequeñas; atado a su entorno |
| **Tessl** | Registro de specs; el código se regenera | Variable | Parcial | Apuesta extrema por la spec; salida no determinista al regenerar |
| **Superpowers** | Skills que fuerzan brainstorming, plan y TDD, con una aprobación humana antes de implementar | Medio | TDD | Sin estado duradero del proyecto; el método vive en prompts |
| **ODD / Gentle-AI** | Configurador de ecosistema (memoria, skills, hooks) para 17 agentes; ODD (guía) y RDD (revisión por recibos donde el binario decide el siguiente paso) | Ligero | Solo la transición de la revisión | El README no detalla qué es código y qué es prompt; el soporte varía por agente |
| **Engram** | Memoria persistente de sesiones con servidor MCP | Mínimo | No | Guarda historia, no verdad; casi todo lo decide el modelo; 27 herramientas MCP |
| **Beads** | Gestor de tareas con grafo de dependencias | Ligero | El grafo sí | Solo tareas; en versiones nuevas el estado va separado de git |
| **AGENTS.md** | Estándar abierto de archivo de instrucciones | Mínimo | No | Solo texto, sin estructura ni comprobación |
| **Beacon** (Asymptote Labs) | Capa de memoria y observabilidad que registra sesiones de agentes y aprende de ellas | — | No | Adyacente: observa lo que pasó, no comprueba lo que se promete. Comparte nombre con este proyecto (ver `status`, Riesgos) |

Engram y Beads se pueden usar junto a Beacon, pero Beacon no depende de ellos.

## 2. Ideas que usamos

| De | Idea | Dónde vive |
|---|---|---|
| ODD | El tamaño se decide por reanudabilidad; solo el trabajo grande deja registro | `method` |
| ODD | Pasar trabajo por referencia, nunca parafraseado | `method` |
| ODD | Preguntar solo decisiones de producto reales, de una en una; nunca pedir permiso para seguir | `method` |
| ODD | Delegar solo con un motivo con nombre | `method` |
| ODD | Un test por regla pedida y uno por comportamiento existente tocado | `method` |
| ODD | Estado visible: Working, Checking, Ready, Needs your decision | `method` |
| Gentle-AI | Fuente canónica renderizada en la configuración nativa de cada agente, con bloques gestionados idempotentes | `anchor` |
| Gentle-AI | Al actualizar, borrar solo lo que coincide con lo instalado; lo editado se conserva | `anchor` |
| Gentle-AI | Adaptadores sin decisiones: toda decisión en un binario | `anchor` |
| Gentle-AI | El agente puede subir el riesgo, nunca bajarlo; la revisión no toma el control de la entrega | `deltas` |
| Gentle-AI | Medir la fricción: en banda, fuera de banda, callejón sin salida | `weaver` |
| Engram | Revisión por antigüedad según el tipo | `guides` |
| Engram | Revelado progresivo: primero un resumen, luego el detalle | `spider` |
| Engram | Formato What / Why / Where para decisiones | `ledger` |
| Beads | Grafo de dependencias tipado y "qué está listo" calculado por código | `spider`, `ledger` |
| Beads | Ids por hash para que las ramas no choquen; hijos jerárquicos | `layout` |
| Beads | `--json` en todo; instrucciones al agente en pocas líneas | `weaver`, `anchor` |
| Beads | Claim al tomar una tarea | `method` |
| Beads | El estado viaja en git, en el mismo commit que el código | `layout` |
| Graphiti | Los hechos se cierran, no se borran | `tenets` |

## 3. Lo que evitamos

| Error | Visto en |
|---|---|
| Pipelines de fases obligatorias para todo | Spec Kit, BMAD |
| Documentos enormes por adelantado | Spec Kit, Kiro |
| Método que solo existe como texto en el prompt | ODD, Superpowers |
| Memoria que acumula historia sin distinguir lo vigente | Engram |
| Decenas de herramientas MCP | Engram |
| Contratos de revisión gigantes y versionados | La revisión de Gentle-AI |
| Estado de tareas separado del código en git | Beads (versiones nuevas) |
| Telemetría activa por defecto | Beads |
| Un solo archivo grande con todo el estado (conflictos de merge) | Beads (versión clásica) |
| Que el agente edite Markdown a mano como fuente de verdad | Spec Kit, ODD |
| Un proceso de tamaño fijo sin importar la tarea | Kiro y Spec Kit (Böckeler, sección 7) |

## 4. El hueco que atacamos

Lo que nadie hace junto:

1. **Un método ligero que se hace cumplir con código.** ODD es ligero pero no se hace cumplir; los pesados tienen estructura pero no ligereza.
2. **Promesas duraderas que se comprueban de forma continua.** Nadie ejecuta checks de forma continua y avisa de la deriva. Spec Kit está añadiendo pasos de convergencia, pero los describe como prompts (sección 7).
3. **Estado del proyecto estructurado y vigente.** Las memorias guardan historia; los documentos se pudren.
4. **Orientación de producto.** Ninguna herramienta conoce la meta ni avisa cuando el trabajo se aleja.
5. **Agnóstico con garantías.** Gentle-AI es agnóstico, pero lo determinista solo llega a una parte de los agentes y no está detallado.

| | Pesado (SDD, BMAD) | Ligero (ODD, AGENTS.md) | **Beacon** |
|---|---|---|---|
| Ceremonia | Alta | Baja | **Baja** |
| Se hace cumplir | Parcial | No | **Sí, con código** |
| Estado duradero | Documentos que se pudren | Notas o nada | **Registros estructurados y vigentes** |
| Comprobación | Al final, manual | Autoinformada | **Continua, por código** |
| Producto | No | No | **Meta activa** |
| Agentes | Uno o pocos | Muchos | **Cualquiera con terminal** |

En la taxonomía de Böckeler (sección 7), Beacon apunta a **spec-anchored con comprobación**, el nivel que las herramientas probadas dejan más vago.

## 5. Lecciones de los autores de Gentle-AI

- "Las buenas decisiones humanas no sobreviven al contacto con un agente a menos que algo las haga cumplir." (`tenets` P2)
- "Un guard que nadie llama no es un guard." Un control documentado y testeado puede no estar conectado a nada.
- "La sobreingeniería es más difícil de quitar que un bug." Su primera revisión la sobreconstruyó el agente al inferir requisitos que nadie pidió.
- "Un mensaje solo puede nombrar un comando si ejecutarlo resuelve el bloqueo." (`tenets` P7)
- "Apunta el instrumento hacia ti mismo primero." (`tenets` P8)
- Un sistema que obliga a ceremonia para cambiar una coma se desinstala en tres días. (`deltas`)

## 6. Evidencia que sostiene decisiones

| # | Hecho | Confianza | Consecuencia |
|---|---|---|---|
| E1 | **Menos contexto, pero preciso, rinde más.** En el sistema Hindsight el mismo modelo de 20B pasó de 39,0 % con todo el historial a 83,6 % con una porción curada. Estudios de "context rot" muestran que el rendimiento baja al crecer el contexto aunque no se llene la ventana. Otro midió que el cumplimiento de restricciones cayó de 73 % en el turno 5 a 33 % en el 16 | Media (papers y blogs) | `context` devuelve solo lo que aplica a esos archivos; el `brief` es mínimo (`spider`, `guides`) |
| E2 | **Los archivos de contexto en prosa ayudan poco.** En un estudio de ETH Zurich (Gloaguen et al., ICLR 2026) los generados por un modelo tienden a bajar la tasa de éxito (cerca de 3 % en promedio) y suben el costo de 20 a 23 %; los escritos por humanos mejoran ~4 % en promedio. Los autores recomiendan omitir los generados y describir solo requisitos mínimos. Releído en el paper el 2026-11-22 | Media-alta (estudio revisado) | El bloque que instala Anchor es de 20 a 30 líneas (`anchor`) |
| E3 | **La guía de Anthropic para `CLAUDE.md`** apunta a menos de 200 líneas y lo trata como contexto, no como mecanismo de cumplimiento (para eso están los hooks) | Alta (documentación oficial) | Lo que debe cumplirse lo hace cumplir el binario o Snares, no el texto |
| E4 | **La interfaz importa más que el sustrato.** LangSmith muestra al agente un filesystem sobre una base de datos; la memoria de los agentes gestionados de Anthropic se monta sobre un filesystem; Letta superó a Mem0 en LoCoMo con un archivo y grep (Letta mide a su competidor) | Baja a media (vendedores) | El agente ve operaciones sobre registros; que sean archivos sirve para versionar y compartir (`layout`) |
| E5 | **Demasiadas herramientas MCP degradan al agente.** La precisión de selección baja desde unas 20 a 50 herramientas visibles; OpenAI recomienda mantenerse muy por debajo de 20; Cursor limita a 40. Kiro responde con "Powers", que activan herramientas MCP solo cuando la conversación las necesita | Media (blogs que coinciden) | Beacon no expone herramientas MCP: su única interfaz es el CLI (`weaver`) |
| E6 | **Las specs se pudren y se comprueban solo al final.** Un estudio de 2026 sobre 470.795 archivos de spec concluye que el detalle de implementación hace que se pudran y que la comprobación spec-código se hace al terminar el cambio, no de forma continua (solo se leyó el resumen) | Baja a media (muestra sesgada hacia vendedores) | Los contratos vivos (`vision`, `pledge`) |
| E7 | **La obediencia a las instrucciones decae dentro de la sesión, y reorganizar el archivo no lo arregla.** Estudio factorial de 1.650 sesiones de Claude Code (16.050 funciones, dos bases TypeScript, sobre todo Sonnet 4.6): cada función adicional generada se asoció con ~5,6 % menos de probabilidad (odds) de cumplir. Tamaño, posición, arquitectura de archivos y contradicciones no mostraron efecto detectable. El decaimiento se halló durante el análisis (exploratorio) y la tarea era una anotación trivial | Media-baja (un autor, exploratorio, un modelo principal; solo leí el abstract) | Reinyectar contexto durante la sesión y hacer cumplir con código, no pulir el prompt (`anchor`; propuesta R1 en `status`) |
| E8 | **Resolver la tarea no es cumplir las reglas del proyecto.** OctoBench (ACL 2026, ocho modelos): hay una brecha sistemática entre resolver la tarea y cumplir las instrucciones del scaffold; las restricciones de skills son un cuello de botella persistente frente a las de memoria; el cumplimiento varía entre Claude Code, Kilo y Droid | Media (solo leí el abstract) | Las garantías no pueden depender del prompt; el bloque de guía se prueba agente por agente (`anchor`) |
| E9 | **No hay evidencia de que las specs reduzcan defectos.** Estudio de ESEM 2026 (Hill): 88.052 PRs fusionados de 119 repos de código abierto; ninguna de las cinco hipótesis se sostuvo tras emparejar por complejidad, porque tener una spec funciona como indicador de tarea difícil. Límites del propio autor: el 97,5 % de las "specs" eran referencias a tickets, no artefactos de herramientas SDD, así que las herramientas no quedaron probadas; sugiere que las specs aún pueden servir como registro de intención | Media (un autor, observacional, análisis post hoc; paper leído) | No prometer menos defectos por tener specs: el valor es lo comprobado. Beacon tampoco tiene esa evidencia, y de ahí el benchmark propio (sección 7.6) |
| E10 | **Taxonomía y problemas de SDD** (Böckeler, Thoughtworks; el Radar de nov 2025 lo pone en "Assess"). Tres niveles: spec-first, spec-anchored y spec-as-source. Todas las herramientas que probó son spec-first y dejan vago el mantenimiento a largo plazo. Problemas: revisar Markdown verboso cuesta más que revisar código, el tamaño del proceso no se ajusta a la tarea, el agente a veces ignora instrucciones y duplica código existente, y el brownfield es difícil | Media-alta (análisis de práctica de una autora) | Confirma el hueco: promesas vivas y ancladas, y proceso proporcional (`method`, P3) |

## 7. Ronda de investigación del 2026-11-22

Cubre los siete temas pendientes. Lo que cambia en el diseño está como propuesta en `status`, no como decisión.

### 7.1 Contexto de proyecto: PRD, ADR, arc42, C4

| Fuente | Qué aporta | Equivalente en Beacon | Cambio propuesto |
|---|---|---|---|
| **ADR** (Nygard; MADR 4.0.0, sept 2024) | Título, estado (proposed, accepted, rejected, deprecated, o superseded con enlace al sucesor), contexto, **opciones consideradas con pros y contras** y resultado. Sus autores insisten en que las opciones descartadas son clave para entender la decisión | `decision`: `what`, `why`, `where`, `supersedes`, `status` | Falta registrar las alternativas descartadas (R4) |
| **arc42** | Doce secciones (metas, restricciones, contexto, estrategia, bloques, vista de ejecución, despliegue, conceptos transversales, decisiones, calidad, riesgos y deuda, glosario). Lema de viajar ligero: quitar lo que nadie necesita; docs-as-code | Metas: `goal`. Bloques: `capability`. Decisiones: `decision`. Calidad: promesas. Restricciones, conceptos, riesgos y glosario: guides de tipo architecture, convention y domain | Ninguno en el esquema. Sirve como esqueleto opcional de la guide de arquitectura que produce el arranque (`method`) |
| **C4** | Diagramas por niveles (contexto, contenedores, componentes, código), con modelo como código | Una capability se parece a un contenedor o componente, pero describe comportamiento, no estructura | No se adopta. Unas vistas Markdown (diferidas, `status`) podrían emitir un diagrama Mermaid de capabilities y relaciones |
| **PRD para agentes** (Osmani, Makerkit, Augment) | Poco pero preciso: qué y por qué antes que cómo; criterios de aceptación **comprobables con entradas y salidas concretas**; **qué no se incluye** (non-goals); casos borde (permisos, límites de tasa), que el agente no construye si nadie los pide | Los criterios de aceptación son las promesas, con la ventaja de que se ejecutan | Falta registrar lo que **no** se hace (R5) y una guía de redacción de promesas (R6) |

Las guías de PRD son blogs y material de vendedores: orientan, no prueban.

### 7.2 OpenSpec, Spec Kit, Kiro y Superpowers

| Herramienta | Qué hace según su documentación | Para Beacon |
|---|---|---|
| **OpenSpec** | `openspec/specs/<capability>/spec.md` es la verdad vigente; cada cambio vive en `openspec/changes/<nombre>/` con una delta spec (ADDED, MODIFIED, REMOVED) y `tasks.md`. Al archivar se fusionan las deltas y el cambio pasa a `changes/archive/AAAA-MM-DD-nombre/`. El esquema de artefactos es configurable (proposal, specs, design, tasks) y el agente consulta la CLI para saber el estado. Paquete npm, Node 20.19 o superior, MIT, más de 30 herramientas | Es el paralelo más cercano: capability ≈ spec, archivo con fecha ≈ `archive/`, delta ≈ nuestros niveles. Beacon cambia el sustrato (YAML y Go), comprueba con código y no necesita carpeta de cambios aparte |
| **Spec Kit** | CLI en Python (`specify`); constitución, `specify`, `clarify`, `plan`, `tasks`, `analyze`, `implement` y `converge` (repetir hasta converger); una rama por spec; extensiones y bundles; la extensión de bugs (assess, fix, test) no usa spec. El README describe los comandos como archivos de prompt y no describe ninguna verificación determinista en `converge`, `analyze` ni los de bugs | Confirma que la comprobación en el mercado la hace el agente. Un blog menciona un comando `spec-drift` de mayo de 2026 que no aparece en el README y no pude verificar |
| **Kiro** | `.kiro/` con `steering/` (contexto fijo), `specs/<feature>/` (requirements en EARS, design, tasks, a veces `tasks.meta.json`) y `hooks/`. Hay specs de bugfix, y los Powers cargan herramientas MCP bajo demanda | EARS (`WHEN … THE SYSTEM SHALL …`) es una forma útil de redactar promesas comprobables (R6). Los Powers confirman el costo de tener muchas herramientas MCP visibles (E5) |
| **Superpowers** | Skills: brainstorming, plan, TDD, desarrollo con subagentes y revisión; la ceremonia escala con la tarea, pero la aprobación humana antes de implementar no; un hook de `SessionStart` carga la skill inicial; las instrucciones del usuario pesan más que las skills. Funciona en Claude Code, Codex, Cursor, Gemini CLI, OpenCode y Copilot | Confirma el patrón de hook de inicio (el nuestro inyecta `beacon brief`). Su método es prompt, sin código que lo haga cumplir |

### 7.3 Gentle-AI: lo que dice su README actual

Solo leí el README; no los documentos enlazados (Review and Architecture, Telemetry, Agents).

- **Componentes:** Engram (memoria), librería de skills, Context7 (opcional), CodeGraph (grafo de símbolos solo lectura), lista de bloqueo de seguridad (`~/.ssh`, `.env`, credenciales), **copias de seguridad de la configuración antes de cada escritura**, `doctor` de solo lectura, personas, asignación de modelos.
- **Agentes:** 17 integraciones en la lista (el badge dice 18).
- **Flujos:** el README describe ODD (trabajo pequeño en línea, el sustancial con un documento recuperable) y **RDD**, donde el binario lee el estado de los archivos y devuelve el siguiente paso válido, la evidencia se ata a un candidato congelado, la profundidad de revisión sale de una evaluación de riesgo de solo lectura y como máximo hay una corrección acotada. Commit, push y release siguen siendo humanos. No menciona fases SDD.
- **Instalación:** Homebrew, curl, `go install` y Scoop en Windows. No describe la idempotencia de actualizar y desinstalar, ni detalla la telemetría.
- **Para Beacon:** el reparto es el mismo que el nuestro (el binario decide, el agente propone) y el README tampoco itemiza qué parte es código. De ahí salen dos ideas baratas: copias de seguridad al instalar y Scoop como canal de distribución (R8).

### 7.4 Flujos de equipo y sesiones paralelas

**Qué se encontró** (casi todo blogs y vendedores, poca evidencia rigurosa):

- La práctica común es una tarea, una rama, un worktree y una sesión de agente. Los worktrees evitan los choques de archivos durante el trabajo y los trasladan al merge.
- Los archivos "hotspot" (rutas, configuraciones, registros) siguen generando conflictos (guía de Augment, vendedor).
- Hay conflictos que Git no ve: dos cambios que fusionan limpio pero dejan el proyecto roto, como renombrar una función en un archivo y llamarla con el nombre viejo en otro (descripción de Collide, vendedor; plausible por naturaleza).
- Un análisis sobre el equipo de agentes de Claude Code señala que varios agentes comparten un `MEMORY.md` sin bloqueos ni estrategia de merge (blog).

**Qué ya cubre el diseño:** un archivo por registro, ids aleatorios y listas ordenadas (`layout`) reducen los conflictos de texto; `.git/beacon/` se comparte entre worktrees (`layout`).

**Huecos que salen de cruzar esto con el diseño** (inferencia mía, no de una fuente):

1. Un claim se guarda como `owner` y `branch` en un registro versionado. En un worktree vecino no se ve hasta que alguien lo commitea y fusiona, así que el claim no avisa a las sesiones paralelas de la misma máquina (R2).
2. `.git/beacon/` lo comparten todos los worktrees, así que dos sesiones pueden escribir a la vez. La escritura tiene que ser atómica (R2).
3. Un conflicto semántico solo lo detecta el CI si corre sobre el resultado de la fusión. En GitHub eso significa también el evento de la cola de merge (R3).

**Sin investigar a fondo:** ramas largas (qué pasa con una promesa cambiada en la rama principal mientras la rama vive). Hipótesis: `beacon diff` compara contra el punto de origen de la rama, no contra la punta de la principal. Se prueba en la Fase 2 con casos reales.

### 7.5 Verificación de UI

Solo documentación y artículos; no probé las herramientas.

- **Playwright MCP:** el agente maneja un navegador real, usa como representación principal el árbol de accesibilidad y reserva las capturas para diagnóstico; las verificaciones registran el código `expect` equivalente.
- **CLIs dedicadas:** `agent-browser` (Vercel Labs) es un wrapper de Playwright pensado para agentes; Simon Willison apunta que los agentes funcionan bien con CLIs dedicadas, lo que respalda que Beacon sea solo CLI. **ProofShot** (código abierto, agnóstico) graba video, recoge errores de consola y servidor y empaqueta la evidencia para que la revise una persona.
- **Regresión visual en CI** (Argos, capturas de Playwright): compara una captura base con la nueva por viewport y detecta cambios de layout que los tests funcionales no ven; exige un renderizado determinista.
- **Límite reconocido en las fuentes:** que el agente mire la pantalla ayuda a iterar, y las capturas comparadas detectan cambios en el tiempo, pero ninguna de las dos prueba que la interfaz sea correcta.

Cómo encaja en los estados de Beacon (propuesta R7):

| Promesa de UI | Cómo se comprueba | Estado |
|---|---|---|
| Comportamiento (un botón abre algo, un formulario muestra un error) | Test E2E con Playwright como ejecutor | Holds |
| Estructura y accesibilidad (roles, nombres accesibles) | Aserciones sobre el árbol de accesibilidad | Holds |
| "No cambió visualmente" | Captura contra una base guardada en git | Holds solo en el sentido de "igual a la base aprobada" |
| "Se ve bien" | El agente deja evidencia (captura o video): es `observed`; la persona atesta | Attested, con la evidencia enlazada |

Beacon no trae navegador: la documentación de `pledge` incluye ejemplos de ejecutores (Playwright, regresión visual).

### 7.6 Benchmark propio de fricción: protocolo

Lo único que estudios como E7 y E8 enseñan para diseñarlo es **qué medir y cómo**. El benchmark no se ha corrido.

- **Pregunta:** ¿Beacon reduce los incumplimientos de promesas y de proceso sin un costo excesivo, y cuánta fricción causa?
- **Condiciones:** A) sin Beacon; B) Beacon con solo el bloque de guía; C) Beacon completo (guía, hooks, CLI y Snares local); opcionalmente D) C sin reinyección a mitad de sesión, para medir R1.
- **Tareas** (piloto de 10, luego 20 a 30), sobre dos repos (uno propio con promesas reales y uno público): una feature pequeña, un bug, un cambio que cruza capabilities, una tarea que **tienta a debilitar una promesa** (un test falla por una promesa vieja), una con una pregunta abierta que bloquea, y sesiones largas (15 funciones o más) para medir el decaimiento.
- **Agentes:** Claude Code y Codex como mínimo. Al menos 3 repeticiones por celda, porque la salida es no determinista (Böckeler). El piloto de 10 tareas, 3 condiciones y 3 repeticiones con un agente son 90 sesiones, con costo real en tokens.
- **Métricas:**
  1. Proceso: ¿corrió `brief` al empezar?, ¿corrió `context` antes de la primera edición de cada archivo? Se mide por paso de la sesión, como en E7.
  2. Integridad: intentos de debilitar una promesa y qué capa los detuvo; lo que llega a git sin un sí debe ser cero.
  3. Fricción: bloqueos en banda, fuera de banda y callejones sin salida, y cantidad de síes pedidos por sesión.
  4. Resultado: tarea resuelta (con tests ocultos) y promesas violadas al final.
  5. Costo: tokens, tiempo y latencia del binario (E2 muestra que el contexto extra cuesta más de 20 %).
- **Análisis:** tasas con intervalos de confianza; no afirmar causalidad con N pequeño; publicar tareas y registros, y reportar también cuando Beacon empata o pierde.
- **Sesgos a vigilar:** quien diseña es quien evalúa; las tareas pueden favorecer a Beacon; un solo modelo principal.