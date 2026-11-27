# KETSET: los packs propios de Beacon (rules y skills)

**Estado:** propuesta aprobada por la persona dueña el 2026-11-27, con el nombre `packs`. No hay código. Las decisiones ya están sincronizadas en la KB (`.claude/delta/packs.md` y los archivos que lo citan); este documento conserva el razonamiento y se retira al cerrar el plan (paso K12). Rama de trabajo: `feat/add-skills-and-rules`. El plan de ejecución está en `PLANNING.md`.
**Origen:** idea de la persona dueña del proyecto, 2026-11-27, tras revisar las skills de Matt Pocock. Ese mismo día se investigaron otras colecciones y la persona adoptó todas las candidatas.
**Nombre:** `packs`, decidido el 2026-11-27 (antes se llamaba "kit"). Un pack es un grupo con nombre de rules y skills que Beacon instala; el conjunto de todos los packs es lo que describe este documento.

## Decisiones del 2026-11-27

| Tema | Decisión |
|---|---|
| Nombre | `packs` |
| Alcance | 11 rules y 9 skills; se adoptaron todas las candidatas |
| Agrupación | Tres packs: `core` por defecto, `git` y `arch` a petición (antes eran cinco) |
| `packs:` en `config.yaml` | Nivel 3. Antes se propuso nivel 2, que no existe para configuración |
| Fase | Propuesta: Fase 4b, entre la 4 y la 5 (`status`) |
| Nombres | Rules y skills de seis letras; los packs, cortos |

---

## 1. La idea en una frase

> Beacon instala en cada proyecto unos pocos packs de rules y skills propias, y lo que distingue a estos packs del resto es que **está atado al estado y a los comandos de Beacon**: sus skills terminan en registros, sus rules salen de un molde que el binario rellena, y sus señales apuntan a la skill que resuelve el problema.

## 2. Por qué

- Hoy Anchor instala un bloque de guía (7 reglas), hooks y un permiso. Eso es todo lo que moldea al agente. Rules y skills son dos herramientas del harness que Beacon aún no usa.
- Las rules de este repo (`.claude/rules/`) son buenas pero están escritas a mano para este proyecto y para Go. Otro proyecto no las puede usar sin reescribirlas.
- Las skills populares (Pocock: 27) existen en gran parte porque **no hay estado**: `handoff`, `triage`, `to-spec`, `wayfinder` suplen lo que Beacon ya tiene (registros, `work` y `task` con `blocks`, `ready`, `brief`). Copiarlas sería duplicar.
- Lo que sí vale de ellas son los comportamientos: interrogar de a una pregunta, confirmar el punto de prueba antes de escribir tests, un bucle de diagnóstico con compuertas. Todo eso termina hoy en Markdown que nadie verifica.

**Lo que lo hace nuestro:** el comportamiento se conserva, pero la salida es un registro de Beacon y la compuerta es un comando que decide el binario (P2: el modelo propone, el código decide).

## 3. Qué entra en los packs

| Pieza | Carga | Tope |
|---|---|---|
| Bloque de guía (ya existe) | Siempre | No crece |
| Rules con `paths:` | Solo al tocar archivos que coinciden | 40 líneas cada una |
| Skills | Bajo demanda; solo su descripción ocupa contexto | 60 líneas cada una; descripción de unos 120 caracteres (cifra a medir) |

Todo va con prefijo `beacon-` para no chocar con lo del usuario: `.claude/rules/beacon-*.md` y `.claude/skills/beacon-*/SKILL.md`.

### Una convención de nombres

Siguiendo la regla de `code-domain-layer` y los nombres del proyecto (spider, weaver, anchor, snares, ledger, pledge), **todo nombre de rule y de skill es una palabra en minúsculas de seis letras**. Los packs tienen nombres cortos a propósito, porque se escriben en `--pack`. Una skill no puede llamarse como un comando de Beacon (`verify`, `close`, `brief`, `attest`...), porque se invoca como `/beacon-<nombre>`; una rule sí puede, porque no se invoca.

### Una anatomía común

Para que todas pesen parecido, cada pieza sigue un esqueleto.

- **Skill:** *cuándo* se usa, *pasos*, *compuerta* (un comando de Beacon que decide), *alarmas* (3 a 5 líneas) y *cierre*.
- **Rule:** el formato de las actuales: título, qué gobierna, secciones de viñetas con su porqué, y no-objetivos.

Los propios packs se verifican: una prueba comprueba el largo del nombre, que no choque con un comando y que no pase del tope de líneas.

## 4. La variabilidad: no es tanta si se parte en tres niveles

El miedo es que cada lenguaje exija reescribir cada rule. Se evita separando **la intención** (agnóstica, nuestra) del **binding** (los valores que cambian, pequeños y estructurados).

1. **Universal:** un archivo fijo, sin huecos.
2. **Paramétrica:** una plantilla con huecos `{{clave}}`. **El binario la rellena** con `text/template` de la biblioteca estándar, a partir de una sección `packs:` de `.beacon/config.yaml` y de lo que ya detecta `beacon init` (por ejemplo, los archivos de test o los manifiestos de dependencias de cada stack, que entran en el `paths:` de la rule). Sin agente: el resultado es determinista, así que `install`, `--preview`, actualizar y desinstalar siguen siendo idempotentes, byte a byte.
3. **Ligada al stack:** aquí entra la idea de que el agente adapte, pero **acotada**. El agente no escribe la rule. Propone el *binding* (unas pocas claves) leyendo el repo; el cambio a `config.yaml` entra por `beacon record` y espera el sí de la persona; el binario renderiza la rule. Así el agente aporta juicio donde hace falta y la rule resultante sigue siendo auditable. Para los stacks con binding ya incluido (Go primero) no hay adaptación, solo selección. Sin binding, esas rules se omiten y `doctor` lo dice.

**Por qué no la otra opción (que el agente reescriba la rule entera):** cada proyecto acabaría con una rule distinta e irreproducible, no habría forma de actualizarla sin pisar lo editado, y rompe la separación entre lo que decide el modelo y lo que decide el código.

## 5. Las rules (11)

| Rule | Sale de | Nivel | Qué hace | Carga |
|---|---|---|---|---|
| `record` | Regla 1 de la guía | Universal | Nunca editar los registros a mano; usar `beacon record`. | Al tocar `.beacon/**` |
| `pledge` | Regla 5 de la guía y `deltas` | Paramétrica | El test cita el id de su promesa. Nunca debilitar ni borrar una promesa para que algo pase. | Archivos de test del stack |
| `frugal` | `code-change-size` | Universal | La escalera de lo más barato primero, con el cierre de lo omitido. Los atajos con límite se anotan con `beacon record`. | Al escribir código |
| `locale` | `code-english-only` | Paramétrica | Un idioma para identificadores y errores, otro para la UI. | Código y UI del stack |
| `vendor` | `code-change-size` y la decisión de dependencias de `status.md` | Paramétrica | Una dependencia nueva se justifica con la escalera y queda como `decision` con sus alternativas descartadas. | Solo manifiestos (`go.mod`, `package.json`, `pyproject.toml`...) |
| `format` | `code-golang-style` | Ligada al stack | Vertical sobre horizontal, sin comentarios, datos de prueba ficticios. | Código del stack |
| `errors` | Parte de `code-domain-layer` | Ligada al stack | Códigos estables, mensajes fijos, detalle estructurado. | Código de errores |
| `secure` | `security-and-hardening` de Osmani | Ligada al stack | Validar en los bordes de confianza y mantener los secretos fuera del código. Lo crítico va además como reglas `deny`, no solo en prosa. | Handlers, CLI y lectores de datos |
| `commit` | `git-commit-policy` | Paramétrica | Conventional Commits, atómicos, por fase. | Siempre, como hoy |
| `branch` | `git-branch-policy` | Paramétrica | `<tipo>/<slug>` corto. | Siempre, como hoy |
| `domain` | `code-domain-layer` | Ligada al stack | Dominio puro que solo mira hacia adentro. | Dominio, según el binding |

`domain` pesa 87 líneas hoy. Al sacar `errors` baja, pero seguirá sobre el tope de 40: es la única excepción documentada, porque carga solo al tocar el dominio. Se revisa en el paso K8.

## 6. Packs

`beacon install --pack <nombre>` y `--preview` para ver qué se escribiría. Sin `--pack`, se instala `core`.

| Pack | Contenido | Por defecto |
|---|---|---|
| `core` | Rules `record`, `pledge`, `frugal`, `locale` y `vendor`; las ligadas al stack (`format`, `errors`, `secure`) cuando hay binding; las nueve skills | Sí |
| `git` | Rules `commit` y `branch`, más `deny` de `git push` y `git pull` en los permisos | Solo con `--pack git` |
| `arch` | Rule `domain` | Solo con `--pack arch` |

**Por qué tres y no cinco.** La primera propuesta separaba `flow` (`review`, `refute`, `divide`) y `code` (las rules de código) para cuidar el contexto. Ese argumento es débil: las rules con `paths:` solo cargan al tocar sus archivos y nueve descripciones de skills pesan poco. Lo que sí justifica un pack aparte son las opiniones fuertes: `git` impone un estilo y cambia permisos, y `arch` impone una arquitectura. Medir `flow` se resuelve en el arnés del benchmark, no en la instalación.

Para quitar una pieza suelta: `--skip <pieza>`. **Ojo:** hoy `--skip` solo acepta `hooks` y `reinject`, así que habría que extenderlo.

## 7. Las skills (9)

Las dos que se conocían de Delta reaparecen: la entrevista para dejar specs claros (`elicit`) y la de escritura técnica (`digest`). No se encontraron en el historial de git, así que su contenido hay que tomarlo de la memoria de la persona o de donde las guarde.

| Skill | Qué hace | Compuerta | Inspiración |
|---|---|---|---|
| `elicit` | Entrevista de una pregunta a la vez. Cada respuesta se vuelve `decision`, `question` o promesa `proposed`. | No se construye mientras `beacon ready` muestre un trabajo bloqueado por una pregunta abierta | `grill-me` de Pocock, `interview-me` de Osmani, `brainstorming` de Superpowers y la skill de Delta |
| `ensure` | La promesa primero, luego el test rojo y el verde. | `beacon check <id>`; el test cita el id de la promesa | `tdd` de Pocock y `test-driven-development` de Superpowers |
| `locate` | Causa raíz antes de arreglar. Tras 3 arreglos fallidos se registra una `question` y se replantea. La regresión queda como promesa. | `beacon check` pasa de rojo a verde | `systematic-debugging` de Superpowers y `diagnosing-bugs` de Pocock |
| `finish` | Ninguna afirmación de "listo" sin evidencia fresca; cada tipo de afirmación se liga al comando que la prueba. Luego, una promesa manual a la vez. | `beacon check` y `beacon attest` | `verification-before-completion` de Superpowers |
| `digest` | Escribir para que otra persona entienda: qué, por qué y cómo, con ejemplos. Cada afirmación cita un id de registro o un `archivo:línea`. | Cada id citado resuelve con `beacon get` | Skill técnica de Delta y `wait-what` de Pocock |
| `tailor` | Ajusta los packs al stack. Propone el binding como un cambio de `config.yaml`; no escribe rules. | `beacon record` con el sí de la persona y `beacon doctor` | Original |
| `review` | Revisa el diff contra promesas y `nongoals`. También cubre recibir feedback: verificar antes de aplicar, sin acuerdo de cortesía. | `beacon diff` y `beacon verify` | `code-review-and-quality` de Osmani, `code-review` de Pocock y `receiving-code-review` de Superpowers |
| `refute` | Un revisor con contexto limpio ataca una decisión no trivial. Máximo 3 ciclos. El resultado es una `decision` con sus `rejected`. | `beacon record` de la decisión. El tope de ciclos es prosa, no se puede comprobar | `doubt-driven-development` de Osmani |
| `divide` | Parte un `work` en tareas, cada una con un resultado comprobable y sus `blocks`. | `beacon doctor` valida la jerarquía y `beacon ready` muestra qué sigue | `writing-plans` de Superpowers, `to-tickets` de Pocock y `planning-and-task-breakdown` de Osmani |

Invocación (a confirmar): `ensure`, `locate` y `finish` pueden ser invocables por el modelo, porque aparecen en el trabajo normal. Las demás se invocan por la persona o por una señal.

## 8. Experiencia de desarrollo

Es el criterio de diseño que más pesa en la instalación y en el uso diario. Se aplica P4: cero comandos que aprender para el usuario, valor en la primera sesión, ninguna ceremonia.

| Qué | Cómo |
|---|---|
| **Cero preguntas por defecto** | `beacon install` instala `core`, detecta el stack y sigue. Lo opcional se pide con `--pack git` o `--pack arch`. |
| **Se ve antes de hacerse** | `--preview` muestra los archivos y cómo quedarían. Al terminar, un resumen de lo instalado y de cómo quitarlo. |
| **Se quita por piezas o entero** | `--skip <pieza>` y `--uninstall`, que borra solo lo que coincide byte a byte. |
| **Actualizar es volver a instalar** | Idempotente. Lo que la persona editó se conserva y `doctor` avisa de que quedó atrás. |
| **Todo visible en `doctor`** | Packs instalados y su versión, piezas editadas o desactualizadas, rules omitidas por falta de binding, skills nombradas por una señal pero no instaladas. |
| **Lo propio no se toca** | Lo que la persona escribe sin el prefijo `beacon-` queda intacto. |
| **No hay que recordar nombres** | La persona no tiene que memorizar skills: las señales las nombran cuando hacen falta (sección 9.1). |
| **Windows** | La comparación byte a byte exige LF. `install` agrega a `.gitattributes` las líneas `text eol=lf` de los archivos de los packs, como ya hace con `.beacon/**`. Esto se vio con el aviso de git de CRLF en este mismo repo. |
| **Sin sorpresas** | Los hooks nunca fallan la sesión; toda instalación deja copia de seguridad; cada bloqueo dice qué hacer (P7). |
| **Se mide** | El benchmark de fricción ya cuenta callejones sin salida, síes pedidos por sesión y texto inyectado. El resultado decide qué se queda. |

## 9. Lo nuevo y potente

### 9.1 Las señales apuntan a la skill

P7 dice que un mensaje de Beacon lleva a una salida. Hoy las salidas son comandos. Con los packs, también pueden ser skills. El texto de la skill no se inyecta: solo una línea con su nombre. Así el contexto no se satura y la skill se descubre justo cuando hace falta.

| Señal | Mensaje |
|---|---|
| Una pregunta abierta bloquea trabajo | `/beacon-elicit` |
| Una promesa con ejecutor aún sin test | `/beacon-ensure` |
| Una promesa `Violated` | `/beacon-locate` |
| Promesas manuales pendientes tras cambiar código | `/beacon-finish` |
| Un cambio de nivel 3 o un diff grande | `/beacon-review` |
| Un `work` grande sin tareas | `/beacon-divide` |
| Una decisión desactualizada (`stale`) | `/beacon-refute` |
| Los packs sin binding para el stack | `/beacon-tailor` (lo dice `doctor`) |

**Condición de P7:** un mensaje solo nombra una skill si está instalada; `doctor` lo comprueba.

### 9.2 Reglas que se compilan en comprobaciones

La tesis de Beacon es que lo que se puede verificar con código no debe depender de la obediencia. Aplicada a los packs: una rule que el binario o el CI pueden comprobar no debería quedarse en prosa.

| Rule | Comprobación posible |
|---|---|
| `commit` | Hook `commit-msg` de Snares |
| `branch` | Pre-commit de Snares |
| `format` (largo de línea) | Ejecutor con un linter |
| `domain` (el dominio solo importa el kernel) | Ejecutor que revisa los imports |
| `secure` | Reglas `deny` y hooks, no prosa |
| `locale`, juicio de diseño | No es verificable: queda como prosa |

Cada una se vuelve una promesa con ejecutor, y la rule en prosa se reduce a lo que de verdad es juicio. Es la fase más ambiciosa y se deja para el final.

### 9.3 Se mide con el benchmark que ya existe

El arnés de `testing/benchmark/` tiene las condiciones A, B, C y D. Se añade **E: C más `core`**. La ablación de una pieza se hace con `--skip` en el arnés. Si una rule o skill no mejora ninguna métrica de proceso, integridad o fricción, se borra. Importante: el arnés está construido pero **el piloto aún no se ha corrido** (Fase 6), así que hoy no hay línea base contra la cual medir.

### 9.4 Esta rama se prueba con este repo

Criterio de aceptación: renderizar los packs con el binding de Beacon debe **reproducir las seis rules actuales de `.claude/rules/`**, salvo las diferencias que se decidan. Si no las reproduce, la plantilla está mal o la rule era demasiado específica.

## 10. Lo aprendido de otras colecciones

Se leyeron 6 de las 20 skills de Superpowers, 3 de las 26 de Osmani, y 4 de las de Pocock. De Anthropic solo la lista. El detalle vive en `survey`, sección 8.

**Adoptado:**

- **La tabla afirmación → evidencia** de `verification-before-completion`. En Beacon ya existe como Holds, Attested, Unverified y Violated, y lo decide el binario; Superpowers lo impone con prosa.
- **El "hard gate" de `brainstorming`** (no se construye hasta aprobar el diseño) equivale a `blocks`: un trabajo con pregunta abierta no sale en `ready`.
- **"Tras 3 arreglos fallidos, replantea"** de `systematic-debugging`, que aquí termina en una `question` registrada.
- **Los ciclos acotados de `doubt-driven-development`**, con el resultado como `decision` y sus `rejected`.
- **La anatomía común** de Osmani, adaptada con la compuerta como paso central.
- **Verificar el feedback antes de aplicarlo** de `receiving-code-review`, dentro de `review`.

**Descartado:**

- Las tablas de racionalizaciones: son prosa y el reporte las ve débiles solas.
- El tamaño: las skills de Superpowers rondan 1.400 a 2.300 palabras, muy por encima del tope de 60 líneas.
- Las skills que Beacon ya cubre con estado: `handoff`, `triage`, `to-spec`, `wayfinder`, `setup`, `finishing-a-development-branch`.

**Sobre saturar:** las descripciones de todas las skills se cargan en cada sesión. Hay guías que recomiendan pocas y bien acotadas, y un paper (SkillReducer, arXiv) reporta que recortar mejora la calidad. Las cifras son de blogs y de un paper no revisado, así que es una dirección y no un dato firme. Con 9 descripciones cortas estamos lejos del problema que describen (decenas de skills), y se mide con E.

## 11. Estructura propuesta

```text
source/anchor/internal/assets/packs/
├─ rules/        # plantillas .md.tmpl con paths: y huecos {{clave}}
├─ skills/       # <nombre>/SKILL.md
└─ bindings/     # go.yaml; ts.yaml y python.yaml después
```

- Va dentro de Anchor, que ya sabe instalar con marcas, copias de seguridad y `--preview`. No se crea una pieza nueva.
- Las rules y skills se escriben en inglés, como la guía del método; la persona recibe respuestas en su idioma.
- Cada archivo instalado lleva la versión del CLI en su marca para saber si está desactualizado.

## 12. Plan

Está en `PLANNING.md`: decisiones, pasos K0 a K12, verificación de cada uno y commits sugeridos.

## 13. Riesgos

| Riesgo | Qué hacer |
|---|---|
| **Saturar al agente** | Topes de tamaño, `paths:`, señales de una línea, packs y la condición E del benchmark. |
| La obediencia decae (`survey`, E7 y E8) | Skills y rules son recordatorio, no garantía. Lo importante sigue en promesas con ejecutor (9.2). |
| Las plantillas se vuelven un mini lenguaje | Solo sustitución y secciones opcionales con `text/template`; si hace falta más, la rule no es agnóstica y pasa al nivel 3. |
| El agente propone un binding malo | Entra por `record` y exige un sí; el binding es corto y se revisa de un vistazo. |
| `paths:` en rules es de Claude Code | Los otros agentes lo resuelven en K11; el tope de 32 KiB de `AGENTS.md` en Codex limita meter todo ahí. |
| Choques con rules o skills del usuario | Prefijo `beacon-` y borrado solo de lo que coincide byte a byte. |
| Los saltos de línea rompen la comparación byte a byte en Windows | `.gitattributes` con `text eol=lf` para los archivos de los packs. |
| Los packs crecen más rápido que su valor | Todo lo que no mejore una métrica se borra; los packs son parte del benchmark. |
| Algunas compuertas son prosa (`refute`, `locate`) | Se dicen como tales; no se vende como verificado lo que no lo es. |
| **Evidencia débil** | Nada demuestra aún que estos packs mejoren resultados, igual que `code-change-size` admite sobre sí mismo. Esta propuesta es una hipótesis hasta correr K9. |
| Licencia | El repo de Pocock es MIT; se reescribe desde las ideas, no se copia texto. |

## 14. Preguntas abiertas

1. **¿Qué contenían las skills de Delta?** Para `elicit` y `digest` conviene partir de ellas.
2. **¿Las rules de este repo (para construir Beacon) pasan a ser la salida de los packs?** Mi propuesta: sí, con 9.4 como criterio. Si no encaja, se mantienen aparte.
3. **¿Qué skills son invocables por el modelo?** Propuesta en la sección 7. Falta confirmar si desactivar la invocación del modelo quita también la descripción del contexto.
4. **¿Cómo se sella el sí cuando la persona ejecuta `install` y cambia `packs:`?** Al ser nivel 3, hay que fijar en K2, con `ledger` y `deltas`, cómo se reconoce el sí de quien escribe el comando en una terminal.
5. **¿Cuál es la forma exacta de `packs:`?** Se fija en K2.

Resueltas el 2026-11-27: el nivel de `packs:` (3), la agrupación en tres packs y la fase (4b).

## 15. Fuera de alcance

- Un registro de skills de terceros ni un instalador de skills ajenas.
- Un lenguaje de plantillas propio.
- Skills que dupliquen comandos de Beacon.
- Cualquier cambio de código en esta rama hasta que se cierre K0.

## 16. Fuentes

- Pocock: <https://www.aihero.dev/skills> y <https://github.com/mattpocock/skills>
- Superpowers: <https://github.com/obra/superpowers>
- Osmani: <https://github.com/addyosmani/agent-skills>
- Anthropic: <https://github.com/anthropics/skills>
- Rules por path: <https://ytrofr.github.io/claude-code-guide/docs/guide/part4-context-engineering/02-rules-system.html>
- Saturación de skills: <https://developer.microsoft.com/blog/stop-skillmaxxing-save-your-tokens/> y <https://arxiv.org/html/2603.29919>
- Lo de Beacon: `.claude/delta/` (`packs`, `anchor`, `status`, `deltas`, `weaver`) y las seis rules de `.claude/rules/`.
