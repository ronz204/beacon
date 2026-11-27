# Packs

Las rules y skills propias de Beacon, que Anchor instala en cada agente. Aprobado el 2026-11-27; **sin construir** (Fase 4b, `status`). Lo no confirmado dice **Propuesta**.

## Qué son

Un **pack** es un grupo con nombre de rules y skills. Moldean el comportamiento del agente más allá del bloque de guía (`anchor`), con una diferencia de fondo respecto de otras colecciones: **están atadas al estado y a los comandos de Beacon**.

- Cada skill termina en un registro (`decision`, `question`, promesa `proposed`) y tiene una **compuerta**: un comando que decide el binario, no el agente (P2).
- Cada rule sale de una plantilla que el binario rellena con los valores del stack, así que sirve en cualquier lenguaje.
- Las señales de Beacon apuntan a la skill que resuelve el problema, con una línea (`weaver`).

| Pack | Contenido | Por defecto |
|---|---|---|
| `core` | Rules `record`, `pledge`, `frugal`, `locale` y `vendor`; las tres ligadas al stack (`format`, `errors`, `secure`) cuando hay binding; las nueve skills | Sí |
| `git` | Rules `commit` y `branch`, y los `deny` de `git push` y `git pull` en los permisos | Solo con `--pack git` |
| `arch` | Rule `domain`: dominio puro que solo mira hacia adentro | Solo con `--pack arch` |

`git` y `arch` quedan aparte porque son opiniones fuertes y `git` cambia permisos. Todo lo demás se instala junto.

## Qué carga y cuánto pesa

| Pieza | Carga | Tope |
|---|---|---|
| Rule con `paths:` | Solo al tocar archivos que coinciden | 40 líneas |
| Skill | Bajo demanda; solo su descripción ocupa contexto | 60 líneas; descripción de unos 120 caracteres (cifra a medir) |

`domain` es la única excepción de tamaño documentada (87 líneas hoy), porque carga solo al tocar el dominio.

## Nombres y anatomía

- **Nombres:** toda rule y toda skill tiene un nombre de **seis letras** en minúsculas, como los del proyecto (spider, weaver, ledger, pledge). Los packs tienen nombres cortos porque se escriben en `--pack`.
- **Prefijo:** todo lo instalado lleva `beacon-` (`.claude/rules/beacon-*.md`, `.claude/skills/beacon-*/SKILL.md`); lo que la persona escriba sin ese prefijo nunca se toca.
- **Sin choques:** una skill no puede llamarse como un comando de Beacon, porque se invoca como `/beacon-<nombre>`. Una rule sí puede.
- **Skill:** cuándo se usa, pasos, compuerta, alarmas (3 a 5 líneas) y cierre. **Rule:** título, qué gobierna, viñetas con su porqué y no-objetivos.
- Una prueba comprueba el largo del nombre, el choque con comandos y los topes.
- Se escriben en inglés, como la guía del método; la persona recibe respuestas en su idioma.

## Cómo se adaptan al stack

El miedo es que cada lenguaje exija reescribir cada rule. Se separa **la intención** (agnóstica) del **binding** (los valores que cambian).

1. **Universal:** archivo fijo (`record`, `frugal`).
2. **Paramétrica:** plantilla con huecos `{{clave}}` que el **binario** rellena con `text/template` (biblioteca estándar) a partir de `packs:` en `config.yaml` y del escaneo de `init`: idioma, tipos de commit, archivos de test o manifiestos del stack (que entran en el `paths:` de la rule). Sin agente: el resultado es determinista, así que instalar, actualizar y desinstalar son idempotentes.
3. **Ligada al stack:** `format`, `errors`, `secure` y `domain` necesitan un binding del stack. Go viene incluido. Si no hay binding, esas rules se omiten, y `doctor` lo avisa con el `hint` de `/beacon-tailor`. El agente **no escribe la rule**: `tailor` propone el binding y la persona lo aprueba.

**Por qué no que el agente reescriba la rule:** cada proyecto acabaría con una rule irreproducible, no habría forma de actualizarla sin pisar lo editado, y se rompería la separación entre lo que decide el modelo y lo que decide el código.

**`packs:` en `config.yaml` es de nivel 3** (`deltas`, `ledger`): decide qué reglas y skills recibe el agente y cambia pocas veces, solo al configurar el proyecto. Su forma exacta es **Propuesta** y se fija en el paso K2 del plan:

```yaml
packs:
  enabled: [core, git]
  skip: [refute]
  binding:
    stack: go
    language: {code: en, ui: es}
```

## Las rules (11)

| Rule | Nivel | Qué hace | Carga |
|---|---|---|---|
| `record` | Universal | Nunca editar los registros a mano; usar `beacon record` | Al tocar `.beacon/**` |
| `pledge` | Paramétrica | El test cita el id de su promesa; nunca debilitar ni borrar una promesa para que algo pase | Archivos de test del stack |
| `frugal` | Universal | La escalera de lo más barato primero y el cierre de lo omitido; los atajos con límite se anotan con `beacon record` | Al escribir código |
| `locale` | Paramétrica | Un idioma para identificadores y errores, otro para la UI | Código y UI del stack |
| `vendor` | Paramétrica | Una dependencia nueva se justifica con la escalera y queda como `decision` con sus alternativas descartadas | Solo manifiestos (`go.mod`, `package.json`, `pyproject.toml`...) |
| `format` | Ligada al stack | Vertical sobre horizontal, sin comentarios, datos de prueba ficticios | Código del stack |
| `errors` | Ligada al stack | Códigos estables, mensajes fijos, detalle estructurado | Código de errores |
| `secure` | Ligada al stack | Validar en los bordes de confianza y mantener los secretos fuera del código; lo crítico va además como `deny`, no solo en prosa | Handlers, CLI y lectores de datos |
| `commit` | Paramétrica | Conventional Commits, atómicos, por fase | Siempre |
| `branch` | Paramétrica | `<tipo>/<slug>` corto | Siempre |
| `domain` | Ligada al stack | Dominio puro, contratos hacia adentro | Dominio, según el binding |

## Las skills (9)

| Skill | Qué hace | Compuerta |
|---|---|---|
| `elicit` | Entrevista de una pregunta a la vez; cada respuesta se vuelve `decision`, `question` o promesa `proposed` | No se construye mientras `beacon ready` muestre un trabajo bloqueado por una pregunta abierta |
| `ensure` | La promesa primero, luego el test rojo y el verde | `beacon check <id>`; el test cita el id |
| `locate` | Causa raíz antes de arreglar; tras 3 arreglos fallidos se registra una `question`; la regresión queda como promesa | `beacon check` pasa de rojo a verde |
| `finish` | Ninguna afirmación de "listo" sin evidencia fresca; luego una promesa manual a la vez | `beacon check` y `beacon attest` |
| `digest` | Escribir para que otra persona entienda: qué, por qué y cómo; cada afirmación cita un id o un `archivo:línea` | Cada id citado resuelve con `beacon get` |
| `tailor` | Ajusta los packs al stack: propone el binding como cambio de `config.yaml`, no escribe rules | `beacon record` con el sí de la persona y `beacon doctor` |
| `review` | Revisa el diff contra promesas y `nongoals`; también cubre recibir feedback: verificar antes de aplicar | `beacon diff` y `beacon verify` |
| `refute` | Un revisor con contexto limpio ataca una decisión no trivial, hasta 3 ciclos; el resultado es una `decision` con sus `rejected` | `beacon record`; el tope de ciclos es prosa, no se puede comprobar |
| `divide` | Parte un `work` en tareas con un resultado comprobable y sus `blocks` | `beacon doctor` valida la jerarquía y `beacon ready` muestra qué sigue |

`elicit` y `digest` se escriben desde cero; si la persona encuentra sus skills antiguas de Delta, se ajustan.

**Invocación (Propuesta):** `ensure`, `locate` y `finish` pueden ser invocables por el modelo; las demás las invoca la persona o una señal. Falta confirmar si desactivar la invocación del modelo quita también la descripción del contexto.

## Señales que apuntan a skills

El texto de la skill no se inyecta: solo una línea con su nombre, para no saturar. **Condición de P7:** un mensaje solo nombra una skill si está instalada, y `doctor` lo comprueba.

| Señal | Mensaje |
|---|---|
| Pregunta abierta que bloquea trabajo | `/beacon-elicit` |
| Promesa con ejecutor aún sin test | `/beacon-ensure` |
| Promesa `Violated` | `/beacon-locate` |
| Promesas manuales pendientes tras cambiar código | `/beacon-finish` |
| Cambio de nivel 3 o diff grande | `/beacon-review` |
| `work` grande sin tareas | `/beacon-divide` |
| Decisión desactualizada (`stale`) | `/beacon-refute` |
| Packs sin binding para el stack | `/beacon-tailor` (lo dice `doctor`) |

## Instalación y experiencia de uso

Se aplica P4: cero comandos que aprender, valor en la primera sesión, ninguna ceremonia.

- **Cero preguntas por defecto:** `beacon install` instala `core`, detecta el stack y sigue. Lo opcional se pide con `--pack git` o `--pack arch`.
- **Quién da el sí:** `packs:` es nivel 3. Una persona que ejecuta `install` en una terminal ya lo da, y se sella a su nombre (`ledger`, Actor); un agente sin terminal recibe el código 2 y pide el sí con `--approve`.
- **Se ve antes de hacerse:** `--preview` muestra qué archivos crearía y cómo quedarían; la instalación termina con un resumen de lo instalado y de cómo quitarlo.
- **Se quita por piezas o entero:** `--skip <pieza>` (**Propuesta**: hoy `--skip` solo acepta `hooks` y `reinject`) y `--uninstall`, que borra solo lo que coincide byte a byte.
- **Actualizar es volver a instalar:** es idempotente. Lo que la persona editó se conserva y `doctor` avisa de que quedó atrás.
- **Todo visible en `doctor`** (**Propuesta**): packs instalados y su versión, piezas editadas o desactualizadas, rules omitidas por falta de binding, skills que una señal nombra sin estar instaladas.
- **Lo propio no se toca:** lo que no lleva el prefijo `beacon-` queda intacto.
- **Windows:** la comparación byte a byte exige saltos de línea LF. `install` agrega a `.gitattributes` las líneas `text eol=lf` de los archivos de los packs, como ya hace con `.beacon/**` (`anchor`).
- **Sin sorpresas:** los hooks nunca fallan la sesión y toda instalación deja una copia de seguridad en `.git/beacon/backups/` (`anchor`, `layout`). Cada bloqueo dice qué hacer (P7).

## Medición

El benchmark de fricción (`survey` 7.6) suma la condición E: C más `core`. La ablación de una pieza se hace con `--skip`. Una rule o skill que no mejore ninguna métrica de proceso, integridad o fricción se borra. **El piloto del benchmark aún no se ha corrido**, así que hasta entonces los packs son una hipótesis (`status`).

## Compilar rules en comprobaciones

Lo que el binario o el CI pueden comprobar no debería quedarse en prosa: el formato de commit (hook `commit-msg` de Snares), el nombre de rama (pre-commit), el largo de línea (ejecutor con un linter) y los imports del dominio (ejecutor). Cada una se vuelve una promesa con ejecutor, y la rule se reduce al juicio. `secure` va con reglas `deny` y hooks. Lo que sigue siendo juicio (el idioma de los mensajes, el diseño) queda como prosa. Es el último paso del plan.

## Límites

- Skills y rules son recordatorio, no garantía: la obediencia decae durante la sesión (`survey`, E7 y E8). Las garantías siguen en las capas 1 y 2 de `design`.
- Las compuertas de `refute` (tope de ciclos) y de `locate` (tope de intentos) son prosa y no se pueden comprobar.
- `paths:` en rules es de Claude Code; los demás agentes lo resuelven en la segunda ola (`anchor`).
