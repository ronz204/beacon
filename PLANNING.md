# Plan: los packs de rules y skills

Plan de ejecución de los packs descritos en `KETSET.md` (qué es y por qué) y en `.claude/delta/packs.md` (el diseño vigente). Aquí va el cómo, el orden y la verificación. Rama: `feat/add-skills-and-rules`.

**Estado:** propuesta y plan aprobados por la persona dueña el 2026-11-27, con el nombre `packs`. D5 quedó en nivel 3 y los packs se redujeron a tres. K0 se cerró el mismo día, D16 y D17 cerraron las dos preguntas que quedaban, y K1 y K2 se hicieron ese día (commits `5a2a167` y `dc66f2e`). Los planes anteriores (Fases 1 a 4) están en el historial de git: `PLANNING.md` hasta el commit `7895362`.

---

## 1. Panorama

### Qué se construye

Anchor aprende a instalar dos cosas nuevas en el agente, además del bloque de guía, los hooks y el permiso:

- **Rules** con `paths:` que cargan solo al tocar los archivos que les corresponden.
- **Skills** que cargan bajo demanda, y a las que las señales de Beacon apuntan con una línea.

Todo se renderiza con una plantilla y un binding, para que sirva en cualquier stack.

### Qué existirá al terminar

- `beacon install --pack <nombre>` con `--preview`, `--uninstall`, copias de seguridad y borrado solo de lo que nadie editó. `--skip` acepta el nombre de una pieza.
- 11 rules y 9 skills en tres packs (`core`, `git`, `arch`).
- Una sección `packs:` en `.beacon/config.yaml` (nivel 3) y un binding de Go incluido.
- Señales de `brief`, `check` y `doctor` que nombran la skill que resuelve el problema.
- Una condición nueva del benchmark de fricción para medir los packs.
- La KB y el README al día, sin documentos huérfanos.

### Qué no se hace

- Instalador de skills de terceros, lenguaje de plantillas propio, skills que dupliquen comandos de Beacon (`KETSET.md`, sección 15).
- Cursor, Codex y los demás agentes antes de que Claude Code esté probado a fondo (K11).

## 2. Decisiones

D1 a D12 quedaron aprobadas el 2026-11-27, con los ajustes que se indican. D13 a D18 son nuevas de ese día (D18 llegó con K1).

### D1. Los packs viven en Anchor

Anchor ya instala con marcas idempotentes, copias y `--preview`. No se crea una pieza nueva. Los fuentes van en `source/anchor/internal/assets/packs/<pack>/` (`rules/` y `skills/`; `bindings/` llega con K8).

### D2. Nombres de seis letras y prefijo `beacon-`

Todo nombre de rule y de skill es una palabra de seis letras en minúsculas. Los packs tienen nombres cortos a propósito, porque se escriben en `--pack`. Una skill no puede llamarse como un comando de Beacon. Los archivos instalados llevan el prefijo `beacon-`. Una prueba lo comprueba.

### D3. Tres packs, solo `core` por defecto

| Pack | Contenido |
|---|---|
| `core` | Rules `record`, `pledge`, `frugal`, `locale`, `vendor`, y `format`, `errors`, `secure` cuando hay binding del stack; las nueve skills |
| `git` | Rules `commit`, `branch` y los `deny` de `git push` y `git pull` |
| `arch` | Rule `domain` |

Antes eran cinco. `flow` y `code` se retiraron: las rules con `paths:` no cargan hasta que se tocan sus archivos, y medir una pieza no necesita un pack (se hace con `--skip` en el arnés).

### D4. El binario renderiza con `text/template`; el binding vive en `config.yaml`

Sin agente en el renderizado, para que sea determinista e idempotente. Biblioteca estándar, sin dependencias nuevas. Los huecos son claves del binding; si una rule necesita más que sustitución y secciones opcionales, no es agnóstica.

### D5. `packs:` en `config.yaml` es de nivel 3 (decidido)

Una rule no ejecuta nada, pero decide qué recibe el agente, y la sección cambia pocas veces. Para `config.yaml` solo existen el nivel 1 y el 3 (el 2 es para promesas y decisiones nuevas), así que la propuesta inicial de nivel 2 no aplicaba. La regla mecánica de `deltas` ya se amplió en la KB. Cuando una persona ejecuta `install` en una terminal, el comando es el sí y se sella a su nombre (`ledger`, Actor); sin terminal, un agente recibe el código 2 y pide el sí con `--approve`. Decidido el 2026-11-27.

### D6. `tailor` propone el binding; no escribe rules

El cambio a `packs:` entra por `beacon record` y espera el sí de la persona (el mismo camino que ya usa `config.yaml`). El agente nunca reescribe una rule.

### D7. Las señales nombran una skill solo si está instalada

Mismo criterio que P7. El mensaje es una línea con el nombre; el texto de la skill no se inyecta. `doctor` comprueba que cada skill nombrada exista.

### D8. Esqueleto y topes

- Skill: cuándo, pasos, compuerta, alarmas, cierre. Hasta 60 líneas. Descripción de unos 120 caracteres (cifra a medir).
- Rule: el formato de las actuales. Hasta 40 líneas, con `domain` como única excepción documentada.
- Una prueba comprueba los topes y la lista de excepciones.

### D9. Aceptación contra este repo

Renderizar los packs con el binding de Beacon debe reproducir las seis rules actuales de `.claude/rules/`, salvo diferencias decididas y anotadas. Si alguna no encaja, esa se queda a mano. Decidido el 2026-11-27.

### D10. Primero solo Claude Code

La segunda ola de agentes (K11) trae sus propios formatos. `paths:` en rules es de Claude Code.

### D11. Los packs se miden

El benchmark de fricción suma la condición E (C más `core`). La ablación de una pieza se hace con `--skip`. Lo que no mejore una métrica se borra. El piloto no se ha corrido, así que hasta entonces los packs son una hipótesis.

### D12. Un commit por paso, y los docs aparte

Se sigue `git-commit-policy`: un commit por fase de trabajo, archivos agrupados por concern, sin `git add -A`. Los documentos de los packs, la KB y el código van en commits separados.

**Excepción de rama:** por preferencia de la persona (2026-11-27), todo el plan va en `feat/add-skills-and-rules`, aunque `git-branch-policy` prefiere ramas cortas. Es una excepción anotada, no un cambio de la política.

### D13. Los archivos instalados llevan LF y `.gitattributes`

La huella de D18 ignora los saltos de línea, así que CRLF no cuenta como edición. Pero Claude Code podría no leer un encabezado con CRLF (sin confirmar) y la rule cargaría siempre. `install` agrega a `.gitattributes` las líneas `text eol=lf` de `.claude/rules/beacon-*` y `.claude/skills/beacon-*/**`, como ya hace con `.beacon/**`.

### D14. Experiencia de desarrollo

Criterio de aceptación de K1, K2 y K7: cero preguntas por defecto, `--preview` antes de escribir, resumen final de lo instalado y de cómo quitarlo, actualizar volviendo a instalar sin pisar lo editado, y todo visible en `doctor`. El detalle está en `KETSET.md`, sección 8, y en `packs.md`.

### D15. `install` crece sin romper lo existente

`--pack <nombre>` es nuevo. `--skip` hoy acepta `hooks` y `reinject`; pasa a aceptar también el nombre de una pieza. Los comportamientos actuales de `install` no cambian. `--approve` es nuevo y solo se usa cuando un agente sin terminal cambia `packs:` (D5). `--skip <pieza>` se guarda en `packs.skip`; `hooks` y `reinject` siguen valiendo solo para esa ejecución.

### D16. El modelo invoca 7 skills; `tailor` y `refute` solo la persona

El modelo puede invocar `elicit`, `ensure`, `locate`, `finish`, `digest`, `review` y `divide`. `tailor` y `refute` llevan `disable-model-invocation: true`: `tailor` cambia la configuración y `refute` lanza un revisor aparte que gasta tokens. Sus señales van dirigidas a la persona. Si el modelo intenta invocar una skill desactivada sin que la persona la nombre, Claude Code lo bloquea (documentación de Claude Code). Restringir no ahorra contexto: el listado usa hasta el 1% de la ventana, y nueve descripciones de unos 120 caracteres pesan poco. Los mensajes de las señales se redactan como hechos, no como órdenes.

### D17. La forma de `packs:`

```yaml
packs:
  enabled: [core, git]
  skip: [refute]
  stacks: [go]
  language:
    code: en
    ui: es
  commit:
    min: 40
    max: 50
```

- `install` escribe todo explícito la primera vez, para que dos clones obtengan las mismas rules. `stacks` es una lista (admite monorepos) con los nombres que ya detecta `init`: `go`, `bun`, `node` y `python`. `commit` es opcional y por defecto vale 40 y 50.
- Al cargar se valida la forma: nombres en minúsculas, números positivos y `min` sin pasar de `max`. Un pack o una pieza que no existen fallan al instalar (`pack_unknown`, `piece_unknown`) y `doctor` los cruzará en K7; un stack desconocido no falla, solo no trae binding.
- El orden canónico es `enabled`, `skip`, `stacks`, `language`, `commit`, con las listas en línea.
- **Dónde toca el código (K2):** decodificar y validar en `source/spider/internal/driver/record/config.go`; el orden en `canonical.go` (`configKeys` y las listas en línea); y el nivel 3 en `source/weaver/internal/domain/deltas` (`configGuardedFields`, un `ReasonPacksChanged` nuevo y `configReasons`). El CI solo trata `runners_changed` aparte; un cambio de `packs:` debe pasar por `unapproved_change`.

### D18. La huella decide qué es de Beacon

Cada archivo instalado lleva en el encabezado una línea `beacon: <huella>`: los primeros 16 caracteres del SHA-256 del contenido sin esa línea y con saltos LF. `install` actualiza un archivo si la huella coincide, y lo conserva si no coincide o si no la tiene. `--uninstall` borra solo los que coinciden. Se eligió sobre comparar con lo que se escribiría hoy porque una versión nueva de Beacon puede actualizar un archivo viejo que nadie tocó, y los archivos propios con el mismo nombre nunca se pisan.

## 3. Estructura objetivo

```text
source/anchor/
├─ internal/
│  ├─ assets/
│  │  └─ packs/
│  │     └─ <pack>/
│  │        ├─ rules/    # <nombre>.md o <nombre>.md.tmpl
│  │        └─ skills/   # <nombre>/SKILL.md
│  ├─ domain/blocks/      # la huella de los archivos gestionados (K1)
│  ├─ driver/catalog/     # lee y renderiza los assets de los packs (K1)
│  └─ action/             # plan y apply, junto a lo existente
└─ ...
```

En K1 la huella vive en el contexto `blocks` y el catálogo en `driver/catalog`. El contexto de dominio `packs` (nombres, packs y binding validados sin leer archivos, `code-domain-layer`) llega con K2, cuando `packs:` necesita validarse; la lectura de `config.yaml` irá en `driver`, como en el resto de Anchor.

## 4. Pasos

Cada paso deja el repositorio funcionando. El tamaño es relativo: S pequeño, M mediano, L grande.

| Paso | Tamaño | Contenido | Verificación | Commit sugerido |
|---|---|---|---|---|
| **K0** | S | Aprobar `KETSET.md`, este plan y la KB, y resolver las preguntas que bloqueaban (hecho el 2026-11-27) | La persona dice sí (hecho) | `docs(packs): record packs proposal and its plan` |
| **K1 (hecho)** | L | Mecánica de instalación de rules y skills en Anchor, solo Claude Code: plantilla, instalar, `--preview`, `--uninstall`, copia de seguridad, prefijo `beacon-`, `--pack`, `--skip <pieza>` y las líneas de `.gitattributes` (D13). Con una rule y una skill mínimas de prueba | Pruebas de `install` con archivos de referencia; idempotencia; desinstalar borra solo lo idéntico; D14 | `feat(anchor): install pack rules and skills` |
| **K2 (hecho)** | M | Sección `packs:` con la forma de D17: decodificar y validar, orden canónico y nivel 3; detección de stack con el escaneo de `init`; el sello del sí desde una terminal y el código 2 sin ella, con `--approve` (D5) | Pruebas del esquema; `fmt` estable; un `packs:` inválido falla al cargar; un cambio sin sí se bloquea en `record` y se comprueba que el CI lo trate por `unapproved_change` | `feat(anchor): add pack config and stack scan` |
| **K3 (hecho)** | M | Rules universales y paramétricas (`record` se hizo en K1): `pledge`, `frugal`, `locale`, `vendor`, `commit`, `branch`; `frugal` y `locale` con los `paths:` del binding, `commit` y `branch` sin ellos | La prueba de D9 sobre las rules que ya existen; la prueba de nombres y topes (D2, D8); una prueba renderiza cada rule y comprueba que su encabezado se lee (si no, carga siempre) | `feat(anchor): add core and git pack rules` |
| **K4 (hecho)** | M | Skills `elicit`, `ensure`, `finish`: las que ya tienen comandos que las respaldan; `elicit` se escribe desde cero. Antes, un spike en el proyecto de pruebas: ¿un hook que nombra una skill hace que el modelo la invoque?, y ¿una skill con `disable-model-invocation` conserva su descripción en `/context`? | Una sesión real en el proyecto de pruebas por skill | `feat(anchor): add elicit ensure finish skills` |
| **K5 (hecho)** | M | Skills `locate`, `digest`, `tailor`; `digest` se escribe desde cero | Ídem; `tailor` produce un cambio de `config.yaml` que pasa por `record` | `feat(anchor): add locate digest tailor skills` |
| **K6** | M | Skills `review`, `refute`, `divide` | Ídem | `feat(anchor): add review refute divide skills` |
| **K7** | M | Solo si el spike de K4 funciona (si no, los mensajes piden a la persona que ejecute el comando): señales que apuntan a skills en `brief`, `check` y `context`, y la comprobación en `doctor` (avisos de piezas editadas, rules omitidas y skills no instaladas) | Una prueba por mensaje: nombra una skill solo si está instalada (D7) | `feat(weaver): point signals to pack skills` |
| **K8** | L | Rules ligadas al stack: `format`, `errors`, `secure`, `domain`; binding de Go; revisar la excepción de tamaño de `domain` | Renderizado contra `golang-style` y `domain-layer` (D9) | `feat(anchor): add go stack-bound pack rules` |
| **K9** | M | Condición E del benchmark, ablación con `--skip` y primera medición | Resultado del piloto; si no mejora, se recorta | `test(benchmark): add pack condition to harness` |
| **K10** | L | Rules compiladas en comprobaciones: `commit-msg`, nombre de rama, linter de largo de línea, import del dominio | Cada una con su promesa y su ejecutor | `feat(snares): compile pack rules into checks` |
| **K11** | L | Segunda ola: Cursor, Codex y el resto, con su formato de reglas. Un commit por agente | Por agente, según su documentación | `feat(anchor): add pack support for cursor` |
| **K12** | S | Cierre: pasar a la KB lo que cambió al construir, revisar `packs.md`, `anchor.md`, `weaver.md` y `status.md`, actualizar el README, y retirar `KETSET.md` | La KB no cita nada inexistente; `KETSET.md` ya no tiene hechos que `packs.md` no tenga | `docs(kb): close the packs phase` |

### Dependencias

- K1 antes de todo. K2 antes de K3, que necesita los huecos.
- K4 a K6 se pueden hacer en cualquier orden después de K1; K4 primero porque depende de comandos ya existentes.
- K7 necesita las skills de K4 a K6 instaladas para poder nombrarlas.
- K9 necesita al menos K4, K5 y K6. K10 y K11 son independientes entre sí.
- K12 es el último y depende de que lo construido esté estable.

### Lo que K1 a K5 dejaron y sus límites

K1 se hizo el 2026-11-27 (commit `5a2a167`), K2 el mismo día (commit `dc66f2e`), y K3, K4 y K5 también, con las pruebas unitarias y de integración y `golangci-lint` en verde. Los atajos tienen un límite conocido y un paso para dejarlos:

- **Resueltos en K2:** `--skip <pieza>` ahora se guarda en `packs.skip` y retira lo ya instalado (la conciliación); `enabled` hace explícito qué packs hay; el binding sale de `config.yaml`.
- **Resuelto en K3:** `stacks` ya se lee. El catálogo traduce cada stack a rutas (`Binding.Code`, `Tests`, `Manifests`) y las rules `frugal`, `locale` (código), `pledge` (pruebas) y `vendor` (manifiestos) cargan solo al tocar esos archivos; sin stack conocido cargan siempre. `commit` y `branch` no llevan `paths:`.
- **Las rutas por stack viven en una tabla del catálogo** (`go`, `bun`, `node`, `python`), no en `config.yaml`. Se mueve a la config si alguien necesita otras rutas.
- **El `deny` de `git push` y `git pull` llegó con la revisión posterior a K4:** el pack `git` lo añade a `.claude/settings.json` (Bash y PowerShell) y `--uninstall` lo quita. Límite: si se quita el pack `git` de `packs:` el `deny` se queda, porque la conciliación solo retira archivos; se quita a mano o con `--uninstall`.
- **D9 (aceptación contra este repo) se cumple a medias:** las rules de los packs son versiones condensadas a 40 líneas o menos, no copias; `code-change-size` se redujo (quita el cierre largo y los no-objetivos) y `branch` omite la variante `phase-<n>`. La comparación línea a línea se cierra en K8 con `golang-style` y `domain-layer`.
- **Los packs y los skips solo crecen con las flags.** Quitar uno se hace editando `packs:` con `beacon record` (nivel 3); el siguiente `install` concilia.
- **`doctor` no cruza aún los nombres de `packs:` con los packs que existen.** Llega con K7.
- **`language.ui` y `commit` no los escribe `install`.** Los añade `tailor` (hecho en K5) o la persona.
- **Sin proyecto Beacon, mandan las flags** y no se guarda nada (el modo de K1: `core` por defecto).
- **Cambio de comportamiento:** la primera instalación en un proyecto Beacon, ejecutada por un agente sin terminal, sale con código 2 hasta recibir `--approve`. Es D5; dos pruebas de integración existentes pasan ahora `--approve`.
- **Resuelto en K4:** al quitar el `SKILL.md` de una skill se borra su carpeta si queda vacía (solo bajo `skills/`).
- **La prueba de choque con los comandos de Beacon** (D2) es de integración y lee la lista de comandos del binario real, así que no se desactualiza.
- **Spike de K4 (una corrida por caso, `claude -p` en un proyecto de prueba con `core` instalado):** las tres skills se listan con su sello en el encabezado. Con un hook `UserPromptSubmit` cuyo `additionalContext` dice que la skill `beacon-elicit` atiende el caso, el modelo abrió su `SKILL.md` (con `Read`, no con la herramienta `Skill`); sin el hook no lo abrió. Es una señal de que el texto empuja, no una prueba. K5 lo midió con `tailor` (ver abajo).
- **Spike de K5 (una corrida por caso):** `beacon-tailor`, con `disable-model-invocation`, no aparece en el listado del modelo y sí responde a `/beacon-tailor`; las demás skills sí aparecen. Con `refute` en K6 se repite. El flujo de `tailor` se comprobó a mano: `beacon record --file` con `{"type":"config",...}` sale con código 2 (`packs_changed`), con `--approve` guarda, y `beacon install` vuelve a renderizar la rule `locale` con el idioma nuevo.
- **`tailor` no se probó dentro de una sesión completa con una persona.** Solo se comprobó que se invoca y que su flujo funciona con los comandos. Las demás skills tampoco se probaron en una sesión de trabajo completa: solo que se listan y que la señal las abre. El benchmark (K9) decide si se quedan.
- **Rules por `paths:`, comprobado en una sesión real (`claude -p`, una corrida por caso):** al empezar solo cargaron `commit` y `branch`; tras leer un `.go` cargaron además `frugal` y `locale`, y `pledge` y `vendor` siguieron fuera. Un encabezado vacío (`---` y `---`) es válido.
- **CRLF (corregido tras K4):** con `autocrlf`, un clon en Windows dejaba las plantillas con CRLF y las rules instaladas heredaban el CRLF que D13 quería evitar. Ahora el catálogo normaliza a LF al renderizar y `.gitattributes` fija LF para los packs.
- **La prueba de topes no tiene excepciones.** `domain` entra con la suya en K8.

### Cómo se escribe cada skill y cada rule

1. Se parte del esqueleto de D8 y del origen indicado en `KETSET.md`.
2. Se escribe el texto en inglés, sin copiar el de las fuentes.
3. Se comprueba contra los topes y la prueba de nombres.
4. Las skills se prueban en una sesión real; lo observado se anota en el commit.

## 5. Riesgos

| Riesgo | Qué hacer |
|---|---|
| **K1 es más grande de lo previsto** | Se corta en dos commits: plantilla y packs, luego instalación y desinstalación. No se avanza a K3 sin K1 idempotente. |
| El renderizado se convierte en un lenguaje | D4: solo sustitución y secciones opcionales. Más que eso manda la rule al nivel 3 de adaptación. |
| Las rules actuales no se reproducen | Se ajusta la plantilla o se deja la rule fuera de los packs. No se fuerza. |
| Skills de baja calidad | Cada una se prueba en una sesión real antes de su commit; el benchmark decide si se queda. |
| Saturar al agente | Topes, `paths:` y la condición E. |
| Medir sin línea base | El piloto no se ha corrido. K9 espera a que se corra o lo corre junto con la Fase 6. |
| Un hook no logra que el modelo invoque la skill | Spike en K4; si no funciona, K7 pide a la persona que ejecute el comando. |
| Un `paths:` mal escrito vuelve la rule permanente | Una prueba renderiza cada rule y comprueba su encabezado (K3). |
| Cambiar `--skip` rompe a quien lo usa | D15: lo existente no cambia; solo se añade. Una prueba lo cubre. |
| Agentes cambiantes | K11 verifica cada agente contra su documentación, como ya exige `anchor.md`. |
| La KB y este plan se desincronizan | K12 revisa todo; mientras tanto, cada paso que cambie una decisión actualiza `packs.md` en el mismo commit. |

## 6. Preguntas que bloquean pasos

No queda ninguna pregunta abierta. Quedan dos verificaciones al construir, ambas en el spike de K4: si un hook puede hacer que el modelo invoque una skill por su nombre, y si una skill desactivada conserva su descripción en el contexto.

Resueltas el 2026-11-27: el nivel de `packs:` (3), los tres packs, la fase (4b), el sello del sí de `install` (ejecutarlo en una terminal es el sí), las skills `elicit` y `digest` (se escriben desde cero y la persona las ajusta si encuentra las de Delta), las rules de este repo como salida de los packs (D9), trabajar todo en una sola rama, la invocación de skills (D16) y la forma de `packs:` (D17).

## 7. Hitos

| Hito | Se alcanza cuando | Pasos |
|---|---|---|
| **Instala** | `beacon install --pack core --preview` y su desinstalación funcionan en Claude Code (alcanzado el 2026-11-27) | K1 y K2 |
| **Rules** | Las rules de este repo se reproducen desde los packs | K3 y K8 |
| **Skills** | Las nueve skills probadas en sesiones reales | K4 a K6 |
| **Guía** | Las señales nombran la skill que resuelve | K7 |
| **Medido** | El benchmark compara los packs con y sin ellos | K9 |
| **Cerrado** | La KB y el README al día, `KETSET.md` retirado | K12 |

## 8. Fuera de esta rama

- Lo de `DRAFTING.md` (la segunda dimensión de la telaraña) es independiente y no se toca aquí.
- El piloto del benchmark de fricción es de la Fase 6.
- Terminar la Fase 4 (Beacon en su propio repo y la medición y KB) no depende de esta rama.
