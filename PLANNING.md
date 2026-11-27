# Plan: los packs de rules y skills

Plan de ejecución de los packs descritos en `KETSET.md` (qué es y por qué) y en `.claude/delta/packs.md` (el diseño vigente). Aquí va el cómo, el orden y la verificación. Rama: `feat/add-skills-and-rules`.

**Estado:** propuesta y plan aprobados por la persona dueña el 2026-11-27, con el nombre `packs`. D5 quedó en nivel 3 y los packs se redujeron a tres. K0 se cerró el mismo día y el siguiente paso es K1. Quedan dos preguntas abiertas (sección 6), y ninguna bloquea K1. Los planes anteriores (Fases 1 a 4) están en el historial de git: `PLANNING.md` hasta el commit `7895362`.

---

## 1. Panorama

### Qué se construye

Anchor aprende a instalar dos cosas nuevas en el agente, además del bloque de guía, los hooks y el permiso:

- **Rules** con `paths:` que cargan solo al tocar los archivos que les corresponden.
- **Skills** que cargan bajo demanda, y a las que las señales de Beacon apuntan con una línea.

Todo se renderiza con una plantilla y un binding, para que sirva en cualquier stack.

### Qué existirá al terminar

- `beacon install --pack <nombre>` con `--preview`, `--uninstall`, copias de seguridad y borrado solo de lo que coincide byte a byte. `--skip` acepta el nombre de una pieza.
- 11 rules y 9 skills en tres packs (`core`, `git`, `arch`).
- Una sección `packs:` en `.beacon/config.yaml` (nivel 3) y un binding de Go incluido.
- Señales de `brief`, `check` y `doctor` que nombran la skill que resuelve el problema.
- Una condición nueva del benchmark de fricción para medir los packs.
- La KB y el README al día, sin documentos huérfanos.

### Qué no se hace

- Instalador de skills de terceros, lenguaje de plantillas propio, skills que dupliquen comandos de Beacon (`KETSET.md`, sección 15).
- Cursor, Codex y los demás agentes antes de que Claude Code esté probado a fondo (K11).

## 2. Decisiones

D1 a D12 quedaron aprobadas el 2026-11-27, con los ajustes que se indican. D13 a D15 son nuevas de ese día.

### D1. Los packs viven en Anchor

Anchor ya instala con marcas idempotentes, copias y `--preview`. No se crea una pieza nueva. Los fuentes van en `source/anchor/internal/assets/packs/` (`rules/`, `skills/`, `bindings/`).

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

Desinstalar compara byte a byte, y en Windows git convierte LF en CRLF. `install` agrega a `.gitattributes` las líneas `text eol=lf` de `.claude/rules/beacon-*` y `.claude/skills/beacon-*/**`, como ya hace con `.beacon/**`.

### D14. Experiencia de desarrollo

Criterio de aceptación de K1, K2 y K7: cero preguntas por defecto, `--preview` antes de escribir, resumen final de lo instalado y de cómo quitarlo, actualizar volviendo a instalar sin pisar lo editado, y todo visible en `doctor`. El detalle está en `KETSET.md`, sección 8, y en `packs.md`.

### D15. `install` crece sin romper lo existente

`--pack <nombre>` es nuevo. `--skip` hoy acepta `hooks` y `reinject`; pasa a aceptar también el nombre de una pieza. Los comportamientos actuales de `install` no cambian.

## 3. Estructura objetivo

```text
source/anchor/
├─ internal/
│  ├─ assets/
│  │  └─ packs/
│  │     ├─ rules/       # <nombre>.md.tmpl
│  │     ├─ skills/      # <nombre>/SKILL.md
│  │     └─ bindings/    # go.yaml
│  ├─ domain/packs/       # packs, binding y reglas de nombre (puro)
│  └─ action/             # plan y apply de los packs, junto a lo existente
└─ ...
```

El contexto `packs` es dominio puro (`code-domain-layer`): valida nombres, packs y binding sin leer archivos. La lectura de `config.yaml` y la escritura de archivos van en `driver`, como en el resto de Anchor. El diseño exacto se ajusta al leer el código de Anchor en K1.

## 4. Pasos

Cada paso deja el repositorio funcionando. El tamaño es relativo: S pequeño, M mediano, L grande.

| Paso | Tamaño | Contenido | Verificación | Commit sugerido |
|---|---|---|---|---|
| **K0** | S | Aprobar `KETSET.md`, este plan y la KB, y resolver las preguntas que bloqueaban (hecho el 2026-11-27) | La persona dice sí (hecho) | `docs(packs): record packs proposal and its plan` |
| **K1** | L | Mecánica de instalación de rules y skills en Anchor, solo Claude Code: plantilla, instalar, `--preview`, `--uninstall`, copia de seguridad, prefijo `beacon-`, `--pack`, `--skip <pieza>` y las líneas de `.gitattributes` (D13). Con una rule y una skill mínimas de prueba | Pruebas de `install` con archivos de referencia; idempotencia; desinstalar borra solo lo idéntico; D14 | `feat(anchor): install pack rules and skills` |
| **K2** | M | Sección `packs:` de `config.yaml` y detección de stack reutilizando el escaneo de `init`; el sello del sí desde una terminal y el código 2 sin ella (D5) | Pruebas del esquema; `fmt` estable; un `packs:` inválido falla al cargar; un cambio sin sí se bloquea | `feat(anchor): add pack config and stack scan` |
| **K3** | M | Rules universales y paramétricas: `record`, `pledge`, `frugal`, `locale`, `vendor`, `commit`, `branch` | La prueba de D9 sobre las rules que ya existen; la prueba de nombres y topes (D2, D8) | `feat(anchor): add core and git pack rules` |
| **K4** | M | Skills `elicit`, `ensure`, `finish`: las que ya tienen comandos que las respaldan; `elicit` se escribe desde cero | Una sesión real en el proyecto de pruebas por skill | `feat(anchor): add elicit ensure finish skills` |
| **K5** | M | Skills `locate`, `digest`, `tailor`; `digest` se escribe desde cero | Ídem; `tailor` produce un cambio de `config.yaml` que pasa por `record` | `feat(anchor): add locate digest tailor skills` |
| **K6** | M | Skills `review`, `refute`, `divide` | Ídem | `feat(anchor): add review refute divide skills` |
| **K7** | M | Señales que apuntan a skills en `brief`, `check` y `context`, y la comprobación en `doctor` (avisos de piezas editadas, rules omitidas y skills no instaladas) | Una prueba por mensaje: nombra una skill solo si está instalada (D7) | `feat(weaver): point signals to pack skills` |
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
| Cambiar `--skip` rompe a quien lo usa | D15: lo existente no cambia; solo se añade. Una prueba lo cubre. |
| Agentes cambiantes | K11 verifica cada agente contra su documentación, como ya exige `anchor.md`. |
| La KB y este plan se desincronizan | K12 revisa todo; mientras tanto, cada paso que cambie una decisión actualiza `packs.md` en el mismo commit. |

## 6. Preguntas que bloquean pasos

| Pregunta | Bloquea | Propuesta |
|---|---|---|
| Qué skills invoca el modelo | K4 | `ensure`, `locate` y `finish`; antes se verifica si desactivar la invocación quita la descripción del contexto |
| Forma exacta de `packs:` | K2 | La de `packs.md`, ajustada al esquema |

Resueltas el 2026-11-27: el nivel de `packs:` (3), los tres packs, la fase (4b), el sello del sí de `install` (ejecutarlo en una terminal es el sí), las skills `elicit` y `digest` (se escriben desde cero y la persona las ajusta si encuentra las de Delta), las rules de este repo como salida de los packs (D9) y trabajar todo en una sola rama.

## 7. Hitos

| Hito | Se alcanza cuando | Pasos |
|---|---|---|
| **Instala** | `beacon install --pack core --preview` y su desinstalación funcionan en Claude Code | K1 y K2 |
| **Rules** | Las rules de este repo se reproducen desde los packs | K3 y K8 |
| **Skills** | Las nueve skills probadas en sesiones reales | K4 a K6 |
| **Guía** | Las señales nombran la skill que resuelve | K7 |
| **Medido** | El benchmark compara los packs con y sin ellos | K9 |
| **Cerrado** | La KB y el README al día, `KETSET.md` retirado | K12 |

## 8. Fuera de esta rama

- Lo de `DRAFTING.md` (la segunda dimensión de la telaraña) es independiente y no se toca aquí.
- El piloto del benchmark de fricción es de la Fase 6.
- Terminar la Fase 4 (Beacon en su propio repo y la medición y KB) no depende de esta rama.
