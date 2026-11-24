# Plan de construcción: Spider

> **Estado: preparación cerrada el 2026-11-23.** Todavía no hay código. Las decisiones D1 a D4 están tomadas y registradas en la KB. Los pasos 4 a 14 están hechos y el siguiente es el 15.

## 1. Panorama

### Qué es Spider

Spider es la memoria de Beacon. Lee los registros de `.beacon/`, los convierte en objetos validados, arma un grafo de relaciones en memoria y calcula, **sin modelo**, lo que el agente necesita saber: qué aplica a un archivo, qué está bloqueado, qué promesa perdió su prueba y qué sigue.

El detalle funcional vive en la KB. Este plan solo dice **en qué orden se construye y qué archivos salen**.

| Tema | Archivo de la KB |
|---|---|
| Qué calcula Spider y qué señales emite | `.claude/delta/spider.md` |
| Tipos de registro, campos y `config.yaml` | `.claude/delta/ledger.md` |
| Carpetas, YAML estricto, ids, estado local | `.claude/delta/layout.md` |
| Guides y frescura | `.claude/delta/guides.md` |
| Estados de una promesa y deriva | `.claude/delta/pledge.md` |
| Fases del proyecto | `.claude/delta/status.md` |
| Cómo se escribe el dominio en Go | `.claude/rules/code-domain-layer.md` |

### Qué existirá al terminar

| Al cerrar | Qué se puede hacer |
|---|---|
| **Bloque 1: preparar** | Reglas corregidas, decisiones cerradas y rama lista |
| **Bloque 2: fundación** (Fase 1) | `beacon fmt` y `beacon doctor` funcionan sobre un `.beacon/` real |
| **Bloque 3: cálculos** | Spider responde `brief`, `context`, `ready`, `drift`, `stale` y `diff` como librería, listo para que Engine lo exponga |

### Cómo fluye la información

```text
.beacon/records/*.yaml y *.md
        │
        ▼
driver/record      lee YAML estricto y reporta archivo y línea
        │
        ▼
domain/*           parsea cada campo y reconstruye las entidades
        │
        ▼
action/load        reúne todos los registros y errores
        │
        ▼
domain/thread      grafo: nodos, relaciones y validación (D3)
        │
        ├──────────── driver/commit    qué cambió en git
        ├──────────── ProofReader      resultados de checks (los produce Engine)
        ▼
action/*           context, ready, brief, drift, stale, diff
        │
        ▼
spider.go          la única puerta de entrada
        │
        ▼
router / engine    comandos del CLI
```

### Qué queda fuera de este plan

- **Ejecutar checks, clasificar niveles y calcular riesgo:** son de Engine.
- **`record`, `close`, `claim` como comandos completos:** son de Engine (Fase 2). Spider les da la lectura, la validación y la escritura de bajo nivel.
- **Archivo mensual (`archive/`):** Fase 5. Hasta entonces nada se archiva, así que una referencia ausente es una referencia rota. Cuando exista, la validación del grafo deberá consultarlo.
- **`render`, Anchor y Sentry:** fases posteriores. `beacon install` deberá agregar `.beacon/** text eol=lf` a `.gitattributes` (`anchor`).

## 2. Estructura objetivo

Esto es lo que quedaría **al terminar los tres bloques**. Los números indican el paso que crea cada parte. Los tests unitarios (`*_test.go`) van junto a cada archivo y no se listan.

```text
beacon/
├─ cmd/beacon/
│  └─ main.go                        # 16 · solo llama a router.Run
│
├─ internal/
│  ├─ kernel/                        # bases puras, sin I/O
│  │  ├─ error.go                    # 4  · Error[C], códigos y detalles
│  │  ├─ actor.go                    # 4  · human:<nombre>, agent:<nombre>
│  │  ├─ date.go                     # 4  · fecha YYYY-MM-DD
│  │  ├─ glob.go                     # 4  · PathGlob; PathMatcher en 15
│  │  ├─ identity.go                 # 4  · RecordID; IDIssuer en 13
│  │  ├─ text.go                     # 5  · texto obligatorio de registros
│  │  ├─ approval.go                 # 5  · approved: by, at, via
│  │  ├─ presence.go                 # 7  · IsMissing para campos opcionales
│  │  └─ failure.go                  # 10 · fallo de un archivo con su línea
│  │
│  ├─ system/                        # I/O compartido, sin lógica de Beacon
│  │  ├─ git.go                      # 21 · ejecuta git sin shell
│  │  └─ atomic.go                   # 12 · archivo temporal + renombrado
│  │
│  ├─ router/                        # CLI
│  │  ├─ router.go                   # 16 · Run(args, stdin, stdout, stderr) int
│  │  └─ internal/
│  │     ├─ handle/
│  │     │  ├─ doctor.go             # 16 · beacon doctor
│  │     │  └─ format.go             # 16 · beacon fmt
│  │     └─ output/
│  │        └─ output.go             # 16 · sobre --json y códigos de salida
│  │
│  └─ spider/
│     ├─ spider.go                   # 16 · fachada pública de la pieza
│     └─ internal/
│        ├─ domain/
│        │  ├─ ledger/               # lo que el proyecto quiere y promete
│        │  │  ├─ goal_entity.go     # 5
│        │  │  ├─ goal_vos.go        # 5
│        │  │  ├─ goal_errors.go     # 5
│        │  │  ├─ goal_policies.go   # 5
│        │  │  ├─ goal_repository.go # 10
│        │  │  ├─ capability_entity.go      # 5
│        │  │  ├─ capability_parts.go       # 5  · Promise
│        │  │  ├─ capability_vos.go         # 5  · check, estados
│        │  │  ├─ capability_errors.go      # 5
│        │  │  ├─ capability_policies.go    # 5
│        │  │  ├─ capability_repository.go  # 10
│        │  │  ├─ capability_proofs.go      # 17 · contrato ProofReader
│        │  │  └─ capability_changes.go     # 21 · contrato ChangeReader
│        │  │
│        │  ├─ memory/               # lo que el proyecto sabe
│        │  │  ├─ decision_entity.go      # 6
│        │  │  ├─ decision_vos.go         # 6
│        │  │  ├─ decision_errors.go      # 6
│        │  │  ├─ decision_policies.go    # 6
│        │  │  ├─ decision_repository.go  # 10
│        │  │  ├─ question_entity.go      # 6
│        │  │  ├─ question_vos.go         # 6
│        │  │  ├─ question_errors.go      # 6
│        │  │  ├─ question_policies.go    # 6
│        │  │  ├─ question_repository.go  # 10
│        │  │  ├─ guide_entity.go         # 6
│        │  │  ├─ guide_vos.go            # 6
│        │  │  ├─ guide_errors.go         # 6
│        │  │  ├─ guide_policies.go       # 6
│        │  │  └─ guide_repository.go     # 10
│        │  │
│        │  ├─ agenda/               # lo que el proyecto está haciendo
│        │  │  ├─ work_entity.go      # 7
│        │  │  ├─ work_vos.go         # 7
│        │  │  ├─ work_errors.go      # 7
│        │  │  ├─ work_policies.go    # 7
│        │  │  ├─ work_repository.go  # 10
│        │  │  ├─ task_entity.go      # 7
│        │  │  ├─ task_vos.go         # 7
│        │  │  ├─ task_errors.go      # 7
│        │  │  ├─ task_policies.go    # 7
│        │  │  ├─ task_repository.go  # 10
│        │  │  ├─ claim_entity.go     # 22
│        │  │  ├─ claim_errors.go     # 22
│        │  │  └─ claim_repository.go # 22
│        │  │
│        │  └─ thread/               # D3 · la tela: relaciones y señales
│        │     ├─ graph_entity.go    # 14
│        │     ├─ graph_vos.go       # 14 · tipos de relación
│        │     ├─ graph_policies.go  # 14 · relaciones válidas y ciclos
│        │     ├─ graph_errors.go    # 14
│        │     ├─ signal_vos.go      # 17 · tipos de señal
│        │     └─ signal_policies.go # 17 · cuándo aparece cada señal
│        │
│        ├─ action/                  # casos de uso
│        │  ├─ load.go               # 10 · carga todo y acumula errores
│        │  ├─ validate.go           # 14 · validación del conjunto
│        │  ├─ doctor.go             # 15
│        │  ├─ signals.go            # 17
│        │  ├─ context.go            # 18
│        │  ├─ ready.go              # 19
│        │  ├─ brief.go              # 20
│        │  ├─ drift.go              # 21
│        │  ├─ stale.go              # 21
│        │  └─ diff.go               # 21 · D4
│        │
│        └─ driver/
│           ├─ record/               # YAML en .beacon/
│           │  ├─ errors.go          # 9  · errores con archivo y línea
│           │  ├─ parser.go          # 9  · YAML estricto
│           │  ├─ store.go           # 10 · lectura y escritura
│           │  ├─ mapper.go          # 10 · YAML a dominio y de vuelta
│           │  ├─ config.go          # 11 · config.yaml
│           │  ├─ format.go          # 12 · salida canónica
│           │  ├─ naming.go          # 13 · <id>-<slug>.<ext>
│           │  └─ testdata/          # 8  · fixtures de referencia
│           ├─ random/
│           │  └─ random.go          # 13 · IDIssuer real
│           ├─ match/
│           │  └─ match.go           # 15 · PathMatcher con doublestar
│           ├─ double/
│           │  └─ double.go          # 17 · contratos en memoria para tests
│           ├─ commit/
│           │  └─ commit.go          # 21 · consultas a git en lote
│           └─ states/
│              └─ states.go          # 22 · claims.json y lo ya inyectado
│
└─ specifies/
   ├─ e2e/
   │  ├─ doctor_test.go              # 16
   │  └─ format_test.go              # 16
   └─ utils/
      ├─ binary.go                   # 16 · compila y ejecuta beacon
      └─ workspace.go                # 16 · crea un repo temporal con .beacon/
```

### Dónde va cada tipo de código

| Si el código... | Va en |
|---|---|
| Es una regla que siempre es cierta, sin I/O | `domain/<contexto>` |
| Lo necesitan tres o más contextos y es puro | `kernel` |
| Coordina varios contextos o arma una respuesta | `action` |
| Habla con YAML, git, el disco o una librería externa | `driver` |
| Es I/O genérico que también usará Engine | `system` |
| Lee argumentos o imprime | `router` |

### Quién puede importar a quién

```text
cmd/beacon → router → spider.go → action → domain → kernel
                                     │
                                     └──→ contratos del dominio
                                              ▲
                                  driver ─────┘ (los implementa)
                                     │
                                     └──→ system, librerías externas
```

Lo cablea `spider.go`: crea los drivers y se los pasa a `action`. `domain` nunca conoce a `driver`.

## 3. Decisiones tomadas

Cerradas el 2026-11-23 en el paso 2 y sincronizadas en `.claude/delta/`.

### D1. Librería YAML: `go.yaml.in/yaml/v4` v4.0.0

- **Por qué:** `gopkg.in/yaml.v3` está archivada y su README la declara sin mantenimiento. `go.yaml.in` es la continuación de la organización YAML, y v4 es la línea con desarrollo activo.
- **Comprobado en una prueba:** conserva el archivo byte a byte, detecta anclas, alias y `<<`, da líneas exactas y convierte 500 archivos a nodos en unos 41 ms.
- **Consecuencia:** no rechaza claves duplicadas al leer nodos, así que esa validación es nuestra (paso 9).
- **Afecta a:** pasos 9, 10 y 12.

### D2. Ids aleatorios

- **Decisión:** 4 caracteres hexadecimales con `crypto/rand`, reintentando si ya existen y creciendo cuando el espacio esté lleno.
- **Por qué:** un hash del contenido haría que dos ramas con el mismo título generen el mismo id.
- **Afecta a:** paso 13.

### D3. El grafo vive en `domain/thread`

- **Decisión:** un cuarto contexto que solo conoce ids, relaciones tipadas y señales. `action` traduce las entidades a nodos. El nombre viene de los "hilos" de `spider`.
- **Por qué:** la regla impide que un contexto importe a otro, y las reglas del grafo deben probarse como dominio puro.
- **Afecta a:** pasos 14 y 17.

### D4. Spider detecta, Engine clasifica

- **Decisión:** Spider entrega **qué cambió** en los registros, y Engine lo **clasifica** en nivel 1, 2 o 3.
- **Consecuencia:** la señal "Pendiente de sí" no existe hasta que Engine la provea. El `brief` deja su espacio listo.
- **Afecta a:** pasos 20 y 21.

## 4. Bloque 1: preparar (cerrado)

### Paso 1. Corregir la regla de dominio (hecho)

- **Qué:** cambiar `source/` por `internal/` en `.claude/rules/code-domain-layer.md`, tanto en sus `paths` como en el texto.
- **Detalle:** corregir también la frase que dice que el kernel vive fuera de cualquier `internal`. Ahora vive en `internal/kernel`, visible para todo el módulo.
- **Por qué primero:** con las rutas viejas, la regla no se carga al editar el dominio.
- **Hecho cuando:** ninguna línea de la regla menciona `source/`.

### Paso 2. Cerrar D1 a D4 (hecho)

- **Qué:** verificar la librería YAML, decidir las cuatro preguntas y registrar el resultado en la KB.
- **Hecho cuando:** `.claude/delta/status.md` y los archivos afectados reflejan las decisiones, y el `README.md` incluye los cambios de estructura.

### Paso 3. Crear la rama (hecho)

- **Qué:** crear `feat/spider-foundation` desde `main` actualizado.
- **Hecho cuando:** la rama existe y el árbol está limpio.

## 5. Bloque 2: fundación (Fase 1)

**Forma de trabajo en todos los pasos:** primero el test que falla, luego el código mínimo, y al final `go test ./...` en verde.

### Paso 4. Kernel (hecho)

- **Objetivo:** tener las piezas puras que todo lo demás usa.
- **Construye:** `Error[C]` con código estable y detalles; `Actor`; fecha; `PathGlob`; y formato de ids `bc-<hex>` con hijos `.1`, `.1.1`. Los contratos `IDIssuer` y `PathMatcher` se agregan cuando los usan los pasos 13 y 15.
- **Cuidado:** un id de la forma `bc-91cd.1` debe saber quién es su padre, porque la jerarquía de tareas depende de eso.
- **Hecho cuando:** cada parser acepta sus ejemplos válidos, rechaza los inválidos y falla con un código estable.

### Paso 5. Contexto `ledger` (hecho)

- **Objetivo:** modelar lo que el proyecto quiere y promete.
- **Construye:** `Goal` y `Capability`, con `Promise` como parte. Incluye `check` (ejecutor con parámetros o manual), `attested` con `evidence`, `approved`, `nongoals`, `origin` y `refers`.
- **Reglas clave:** una promesa solo nace desde su capability; un check tiene runner o `manual`, nunca ambos; los parámetros de un runner no pueden empezar con `-`.
- **Cuidado:** el estado de ciclo (`proposed`, `active`) se guarda, pero el de verificación (`Holds`, `Violated`...) **se calcula** y no vive en la entidad.
- **Hecho cuando:** ninguna capability inválida se puede construir desde fuera del paquete.

### Paso 6. Contexto `memory` (hecho)

- **Objetivo:** modelar lo que el proyecto sabe.
- **Construye:** `Decision` (con `rejected`, `supersedes`, `review`), `Question` (con `blocks`, `answer`) y `Guide` (con `kind`, `applies`, `always`, `audience`, `status`).
- **Reglas clave:** una pregunta `answered` tiene respuesta; `review` es una duración válida (`6mo`, `12mo`, `30d`).
- **Cuidado:** el id de una guide sale del nombre del archivo, no de su frontmatter.
- **Hecho cuando:** los tres tipos se construyen solo con datos válidos.

### Paso 7. Contexto `agenda` (hecho)

- **Objetivo:** modelar lo que el proyecto está haciendo.
- **Construye:** `Work` (petición textual, log, `implements`) y `Task` (`parent`, `blocks`, `owner`, `branch`).
- **Reglas clave:** el id de una tarea desciende del de su `parent`; una tarea `claimed` tiene `owner`.
- **Cuidado:** `parent` siempre apunta al work raíz, porque `ledger` lo define como task → work. Una tarea anidada como `bc-91cd.1.1` también tiene como `parent` a `bc-91cd`. La petición del work se guarda textual, sin recortar espacios.
- **Hecho cuando:** los estados cerrados y la jerarquía están probados.

### Paso 8. Fixtures de referencia (hecho)

- **Objetivo:** fijar el resultado correcto antes de escribir el parser.
- **Construye:** en `driver/record/testdata/`:
  - `valid/`: un ejemplo por cada tipo, tomados de `ledger`.
  - `invalid/`: anclas, alias, `<<`, claves duplicadas, campos con tipo incorrecto, enumeraciones inválidas.
  - `unknown/`: registros con campos que el esquema no conoce.
  - `quoting/`: `yes`, `no`, `null`, `1.0`, vacíos, valores con `:` o `#`, espacios en los bordes y caracteres de flujo.
  - `canonical/`: pares de entrada y salida esperada de `fmt`.
- **Hecho cuando:** cada regla de `layout` tiene al menos un fixture.

### Paso 9. Parser YAML estricto (hecho)

- **Objetivo:** leer YAML sin que el parser adivine tipos.
- **Construye:** `driver/record/parser.go`, que trabaja sobre el árbol de nodos.
- **Reglas clave:** rechaza anclas, alias, `<<` y claves duplicadas; `version: 1.0` se lee como el tipo que diga el esquema; todo error lleva archivo, línea y una forma de arreglarlo (P7).
- **Cuidado:** `yaml.Unmarshal` lee solo el primer documento y acepta tags propios, así que el parser debe rechazar por su cuenta varios documentos, tags personalizados y archivos vacíos.
- **Depende de:** D1.
- **Hecho cuando:** los fixtures `invalid/` de estructura YAML fallan con el error y la línea esperados. Los de esquema y dominio se cubren en el paso 10, porque necesitan el mapper.

### Paso 10. Carga de `.beacon/` (hecho)

- **Objetivo:** pasar de archivos a entidades.
- **Construye:** los contratos de repositorio de cada contexto, `store.go`, `mapper.go` y `action/load.go`.
- **Cuidado:** `Restore<Entity>` confía en su entrada, así que el mapper llama al `Validate<Entity>State` de cada tipo antes de reconstruir el registro. El cuerpo Markdown de una guide no entra a la entidad: lo guarda el driver y se lee solo cuando hace falta.
- **Reglas clave:**
  - Lectura en paralelo, un worker por núcleo.
  - **Un archivo roto no detiene la carga:** se juntan todos los errores.
  - Los campos desconocidos se quedan en el driver y se reescriben sin cambios.
  - El nombre del archivo debe coincidir con el `id` del registro.
- **Hecho cuando:** un `.beacon/` de ejemplo carga los 7 tipos, un directorio con varios errores los reporta todos y los fixtures `invalid/` restantes fallan con su error, salvo los dos de `config/`, que son del paso 11.

### Paso 11. `config.yaml` (hecho)

- **Objetivo:** leer la configuración que Spider necesita.
- **Construye:** `driver/record/config.go`.
- **Reglas clave:** lee `schema`, `limits` y `review`; aplica los valores por defecto; **rechaza un esquema más nuevo** que el del binario; conserva `runners`, `agents` y cualquier campo desconocido.
- **Hecho cuando:** `schema: 2` falla con un mensaje que dice cómo resolverlo.

### Paso 12. `fmt` canónico (hecho)

- **Objetivo:** que un registro tenga una sola forma posible en disco, para que los diffs sean pequeños.
- **Construye:** `driver/record/format.go` y `system/atomic.go`. No hay `action/format.go`, porque no coordinaría nada: `spider.go` llamará directo a `Store.Format`.
- **Reglas clave:** orden fijo de claves; listas por id; comillas solo donde otra herramienta podría reinterpretar; escritura atómica.
- **Cuidado:**
  - Los campos desconocidos van **después de los conocidos, en orden alfabético**, y los comentarios YAML se conservan junto a su clave.
  - Los ejemplos de `ledger` usan mapas en línea (`check: {runner: ...}`). Los fixtures `canonical/` deciden exactamente cuándo.
  - **Saltos de línea en Windows:** git ya está convirtiendo LF a CRLF en este repo. `fmt` debe escribir LF, y `.beacon/**` necesita `text eol=lf` en `.gitattributes`, o cada archivo parecería modificado.
- **Hecho cuando:** formatear dos veces da el mismo resultado byte a byte y ningún campo desconocido se pierde.

### Paso 13. Ids y slugs (hecho)

- **Objetivo:** crear registros nuevos con nombres estables.
- **Construye:** el contrato `IDIssuer` en `kernel`, `driver/random/random.go` y la generación de `<id>-<slug>.<ext>`.
- **Reglas clave:** el slug sale del título, en minúsculas y con guiones, y normaliza acentos (`Pagos sin fricción` da `pagos-sin-friccion`).
- **Depende de:** D2.
- **Hecho cuando:** un choque simulado reintenta, y un espacio lleno hace crecer el id.

### Paso 14. Grafo en memoria (hecho)

- **Objetivo:** convertir registros sueltos en una red navegable.
- **Construye:** `domain/thread` y `action/validate.go`.
- **Relaciones:** `implements`, `blocks`, `parent`, `supersedes`, `serves`, `origin` y `refers`, con los orígenes y destinos válidos de `ledger`.
- **Valida:**
  - Ids duplicados, incluidas las promesas.
  - Referencias a registros inexistentes.
  - Relaciones con un tipo de origen o destino inválido.
  - Ciclos en `blocks` y `parent`, que dejarían trabajo bloqueado para siempre.
  - Subtareas huérfanas: `bc-91cd.1.1` exige que exista `bc-91cd.1`.
  - Tareas `done` con subtareas `open` o `claimed`.
- **Depende de:** D3.
- **Hecho cuando:** cada error tiene un test y el recorrido de vecinos es determinista.

### Paso 15. Globs y `doctor`

- **Objetivo:** saber qué registros aplican a una ruta y diagnosticar el estado completo.
- **Construye:** el contrato `PathMatcher` en `kernel`, `driver/match/match.go` y `action/doctor.go`.
- **Reglas de globs:** `*`, `**`, `?`, `{a,b}`; desde la raíz; con `/`; sensibles a mayúsculas; sin negaciones.
- **Cuidado:** en Windows, las rutas se normalizan a `/` antes de comparar. Los corchetes se rechazan, así que `app/[id]/page.tsx` se escribe `app/*/page.tsx` hasta que haga falta un escape.
- **`doctor` en esta fase:** valida esquema, formato y grafo, y avisa de guides que superan `limits.guide`. La ayuda con conflictos se agrega al usarlo en ramas reales, y la limpieza de claims en el paso 22.
- **Hecho cuando:** `doctor` reporta todos los problemas de un `.beacon/` roto con su solución.

### Paso 16. Fachada y CLI mínimo

- **Objetivo:** cerrar la Fase 1 con algo ejecutable.
- **Construye:** `spider.go`, `router`, `cmd/beacon/main.go` y las pruebas de `specifies/`.
- **Comandos:** `beacon fmt` y `beacon doctor`, con `--json` y códigos de salida 0, 1 y 2.
- **Medición:** repetir la carga de 500 archivos y comparar con los ~100 ms de `layout`.
- **Hecho cuando:** los tests e2e corren el binario real, y el tiempo medido queda anotado aunque no alcance la referencia.

## 6. Bloque 3: cálculos de Spider

Corresponde a las Fases 2 y 5 de `status`. Se construye como librería en una rama nueva, `feat/spider-queries`. Engine solo tendrá que exponerlo.

### Paso 17. Señales

- **Objetivo:** que el estado avise sin que se le pregunte.
- **Construye:** `signal_vos.go`, `signal_policies.go`, `action/signals.go`, el contrato `ProofReader` y `driver/double`.

| Señal | Se construye en | Necesita |
|---|---|---|
| Bloqueo | 17 | Grafo |
| Sin alcance | 17 | Registros |
| Tensión por relación | 17 | Grafo |
| Rojo | 17 | `ProofReader` |
| Sin comprobar aquí | 17 | `ProofReader` |
| Deriva | 21 | Git y `ProofReader` |
| Desactualizado | 21 | Git y fechas |
| Pendiente de sí | Engine | Clasificación por niveles (D4) |

- **Hecho cuando:** cada señal tiene un caso que aparece y otro que no.

### Paso 18. `context <paths>`

- **Objetivo:** entregar todo lo que aplica a unos archivos.
- **Devuelve:** capabilities coincidentes con sus `nongoals`, promesas con estado, decisiones con sus `rejected`, guides con `applies` coincidente o `always`, preguntas relacionadas y señales.
- **Regla clave:** recuperación por globs y relaciones, nunca por búsqueda de texto. No necesita los cuerpos de las guides que no apliquen ni `archive/`.
- **Hecho cuando:** un archivo sin dueño recibe solo la meta y las guides `always`.

### Paso 19. `ready`

- **Objetivo:** saber qué trabajo se puede tomar.
- **Regla clave:** recorre `blocks` y `parent`; al cerrar un bloqueo, los dependientes quedan disponibles. El bloqueo se hereda a las subtareas, y una tarea con subtareas abiertas no aparece porque se trabaja en las hojas (`ledger`). Recibe la rama como parámetro.
- **Hecho cuando:** una pregunta abierta saca su trabajo de `ready`, y responderla lo devuelve.

### Paso 20. `brief`

- **Objetivo:** el resumen que recibe el agente al empezar.
- **Regla clave:** tope de `limits.brief` líneas. La meta activa siempre entra, y el resto se recorta en este orden fijo:
  1. Bloqueos y rojos.
  2. Trabajo en curso.
  3. Pendientes de sí (vacío hasta que exista Engine, D4).
  4. Deriva y promesas sin comprobar aquí.
  5. Índice de guides.
- **Hecho cuando:** con exceso de contenido nunca se pierde lo crítico y aparece `N more, run beacon status`.

### Paso 21. Git: `drift`, `stale` y `diff`

- **Objetivo:** cruzar el estado con la historia del código.
- **Construye:** `system/git.go`, `driver/commit`, y `action/drift.go`, `stale.go` y `diff.go`.
- **Reglas clave:**
  - Consultas **en lote**: una por commit base distinto y una de frescura. Nunca una por promesa.
  - El cruce con `paths` se hace en memoria.
  - `git` se ejecuta sin shell.
- **Depende de:** D4.
- **Hecho cuando:** los tests corren sobre repos temporales reales y 500 capabilities quedan cerca de los 150 a 200 ms de `layout`.

### Paso 22. Estado local

- **Objetivo:** compartir estado entre worktrees sin riesgo.
- **Construye:** `driver/states`, claims locales, `context --hook` y la limpieza de claims en `doctor`.
- **Reglas clave:** toda escritura en `.git/beacon/` es atómica, y la ruta sale de `git rev-parse --git-common-dir`.
- **Cuidado:** en Windows, renombrar sobre un archivo que otro proceso tiene abierto puede fallar, así que hace falta reintentar y probarlo con escritores concurrentes.
- **Hecho cuando:** varios procesos escriben a la vez sin dejar un archivo a medias, y `--hook` no repite lo ya inyectado.

## 7. Hitos y commits

Se commitea por hito revisado, no por archivo.

| Hito | Pasos | Rama | Resultado verificable |
|---|---|---|---|
| H0. Preparación | 1 a 3 | `main` y rama nueva | Regla y decisiones al día |
| H1. Dominio | 4 a 7 | `feat/spider-foundation` | Tests del dominio en verde |
| H2. Persistencia | 8 a 13 | `feat/spider-foundation` | `fmt` idempotente sobre los fixtures |
| H3. Fase 1 cerrada | 14 a 16 | `feat/spider-foundation` | `beacon doctor` y `beacon fmt` en e2e |
| H4. Consultas | 17 a 20 | `feat/spider-queries` | `context`, `ready` y `brief` probados |
| H5. Git y estado local | 21 y 22 | `feat/spider-queries` | Deriva, frescura y worktrees probados |

## 8. Riesgos

| Riesgo | Qué lo mitiga |
|---|---|
| YAML v4 es reciente (2026-11-04) | Los fixtures de los pasos 8 a 12 la prueban antes de depender de ella; v3.0.6 queda como alternativa cercana |
| CRLF en Windows genera diffs falsos | `fmt` escribe LF, más `text eol=lf` para `.beacon/**` (paso 12) |
| Renombrado atómico falla en Windows | Reintentos y prueba concurrente (paso 22) |
| El dominio en Go resulta verboso por la regla | Se evalúa al cerrar H1, antes de replicarlo en Engine |
| Spider crece antes de que Engine valide su interfaz | El Bloque 3 se construye contra los comandos definidos en `engine` |
| Sobreingeniería | Cada paso implementa solo lo que la KB pide |
| Benchmarks sintéticos | Se miden de nuevo en los pasos 16 y 21, con el código real |
