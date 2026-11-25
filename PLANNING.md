# Plan de construcción

Qué se construye, en qué orden y qué archivos salen. El detalle funcional vive en la KB (`.claude/delta/`), y las fases, en `status`.

| Parte | Fases | Estado |
|---|---|---|
| **I. Spider** (secciones 1 a 8) | Fase 1 y los cálculos del Bloque 3 | Cerrada |
| **II. Weaver** (secciones 9 a 17) | Fase 2 | Planificada, sin empezar |

# Parte I. Spider

> **Estado: Fase 1 cerrada el 2026-11-23.** `beacon fmt` y `beacon doctor` funcionan sobre un `.beacon/` real. Las decisiones D1 a D4 están registradas en la KB. El Bloque 3 (pasos 17 a 22) está hecho: Spider calcula todo lo que pide `spider`, como librería, y lo expone en el CLI la Parte II.

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
| **Bloque 3: cálculos** | Spider responde `brief`, `context`, `ready`, `drift`, `stale` y `diff` como librería, listo para que Weaver lo exponga |

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
        ├──────────── ProofReader      resultados de checks (los produce Weaver)
        ▼
action/*           context, ready, brief, drift, stale, diff
        │
        ▼
spider.go          la única puerta de entrada
        │
        ▼
router / weaver    comandos del CLI
```

### Qué queda fuera de este plan

- **Ejecutar checks, clasificar niveles y calcular riesgo:** son de Weaver.
- **`record`, `close`, `claim` como comandos completos:** son de Weaver (Fase 2). Spider les da la lectura, la validación y la escritura de bajo nivel.
- **Archivo mensual (`archive/`):** Fase 2, paso 25 (D6). Hasta entonces nada se archiva, así que una referencia ausente es una referencia rota. Cuando exista, la validación del grafo deberá consultarlo.
- **Anchor y Snares:** fases posteriores. `beacon install` deberá agregar `.beacon/** text eol=lf` a `.gitattributes` (`anchor`).

## 2. Estructura objetivo

Esto es lo que quedaría **al terminar los tres bloques**. Los números indican el paso que crea cada parte. Los tests unitarios (`*_test.go`) van junto a cada archivo y no se listan.

```text
beacon/
├─ cmd/beacon/
│  └─ main.go                        # 16 · solo llama a router.Run
│
├─ source/
│  ├─ kernel/                        # bases puras, sin I/O
│  │  ├─ error.go                    # 4  · Error[C], códigos y detalles
│  │  ├─ actor.go                    # 4  · human:<nombre>, agent:<nombre>
│  │  ├─ date.go                     # 4  · fecha YYYY-MM-DD
│  │  ├─ glob.go                     # 4  · PathGlob; PathMatcher en 15
│  │  ├─ identity.go                 # 4  · RecordID; IDIssuer en 13
│  │  ├─ text.go                     # 5  · texto obligatorio de registros
│  │  ├─ approval.go                 # 5  · approved: by, at, via
│  │  ├─ presence.go                 # 7  · IsMissing para campos opcionales
│  │  ├─ failure.go                  # 10 · fallo de un archivo con su línea
│  │  ├─ formatter.go                # 15 · contrato RecordFormatter
│  │  └─ describe.go                 # 15 · código, mensaje y ubicación de un fallo
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
│  │     │  ├─ format.go             # 16 · beacon fmt
│  │     │  └─ project.go            # 16 · ubicar .beacon/ y textos de hallazgos
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
│        │  │  ├─ capability_proofs.go      # 17 · contrato ProofReader y resultados
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
│           │  └─ match.go           # 15 · PathMatcher sobre regexp
│           ├─ double/
│           │  └─ double.go          # 17 · contratos en memoria para tests
│           ├─ commit/
│           │  └─ commit.go          # 21 · consultas a git en lote
│           └─ states/
│              └─ states.go          # 22 · claims.json y lo ya inyectado
│
└─ testing/
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
| Es I/O genérico que también usará Weaver | `system` |
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

### D4. Spider detecta, Weaver clasifica

- **Decisión:** Spider entrega **qué cambió** en los registros, y Weaver lo **clasifica** en nivel 1, 2 o 3.
- **Consecuencia:** la señal "Pendiente de sí" no existe hasta que Weaver la provea. El `brief` deja su espacio listo.
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

### Paso 15. Globs y `doctor` (hecho)

- **Objetivo:** saber qué registros aplican a una ruta y diagnosticar el estado completo.
- **Construye:** el contrato `PathMatcher` en `kernel`, `driver/match/match.go` y `action/doctor.go`.
- **Decisión:** los globs se implementan a mano. Las llaves se expanden (máximo 64 combinaciones) y cada alternativa se traduce a `regexp`, que corre en tiempo lineal. Una comparación con 400.000 casos contra `doublestar` v4.10.4 mostró que esa librería no es coherente con `**` dentro de llaves (`a/**` coincide con `a`, pero `a/{**,b}` no).
- **Límite conocido:** una alternativa vacía pegada a una barra, como `{a,}/b`, genera un segmento vacío y su resultado puede diferir de `doublestar`.
- **`doctor` no valida el grafo si un registro falló al cargar**, porque reportaría referencias rotas que en realidad son consecuencia del primer error.
- **Reglas de globs:** `*`, `**`, `?`, `{a,b}`; desde la raíz; con `/`; sensibles a mayúsculas; sin negaciones.
- **Cuidado:** en Windows, las rutas se normalizan a `/` antes de comparar. Los corchetes se rechazan, así que `app/[id]/page.tsx` se escribe `app/*/page.tsx` hasta que haga falta un escape.
- **`doctor` en esta fase:** valida esquema, formato y grafo, y avisa de guides que superan `limits.guide`. La ayuda con conflictos se agrega al usarlo en ramas reales, y la limpieza de claims en el paso 22.
- **Hecho cuando:** `doctor` reporta todos los problemas de un `.beacon/` roto con su solución.

### Paso 16. Fachada y CLI mínimo (hecho)

- **Objetivo:** cerrar la Fase 1 con algo ejecutable.
- **Construye:** `spider.go`, `router`, `cmd/beacon/main.go` y las pruebas de `testing/`.
- **Comandos:** `beacon fmt` y `beacon doctor`, con `--json` y códigos de salida 0, 1 y 2.
- **Medición:** repetir la carga de 500 archivos y comparar con los ~100 ms de `layout`. **Resultado:** unos 70 ms con carga más grafo, anotado en `layout`.
- **CLI:** `spf13/cobra` v1.10.2. Los errores de uso también devuelven el sobre `--json` con código 1.
- **Hecho cuando:** los tests e2e corren el binario real, y el tiempo medido queda anotado aunque no alcance la referencia.

## 6. Bloque 3: cálculos de Spider

Corresponde a las Fases 2 y 5 de `status`. Se construye como librería en una rama nueva, `feat/spider-queries`. Weaver solo tendrá que exponerlo.

### Paso 17. Señales (hecho)

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
| Pendiente de sí | Weaver | Clasificación por niveles (D4) |

- **Hecho cuando:** cada señal tiene un caso que aparece y otro que no.
- **Cómo quedó:** el estado de cada promesa (`Holds`, `Attested`, `Unverified`, `Violated`) lo calcula `ledger.VerifyPromise` a partir del resultado de su check. Las señales de grafo (bloqueo y tensión) viven en `thread`; las que dependen de promesas (rojo, sin comprobar aquí, sin alcance) se calculan en `action/signals.go`. Una pregunta abierta es un nodo `Unfinished` del grafo.
- **Pendiente:** todavía no hay `ProofReader` real, porque los resultados de checks los produce Weaver. Hasta entonces solo existe el doble en memoria, y `context` no se expone en el CLI.

### Paso 18. `context <paths>` (hecho)

- **Objetivo:** entregar todo lo que aplica a unos archivos.
- **Devuelve:** capabilities coincidentes con sus `nongoals`, promesas con estado, decisiones con sus `rejected`, guides con `applies` coincidente o `always`, preguntas relacionadas y señales.
- **Regla clave:** recuperación por globs y relaciones, nunca por búsqueda de texto. No necesita los cuerpos de las guides que no apliquen ni `archive/`.
- **Hecho cuando:** un archivo sin dueño recibe solo la meta y las guides `always`.
- **Cómo quedó:** `context` devuelve las metas activas, las capabilities cuyos `paths` coinciden (con el estado de cada promesa), las decisiones activas con `where` coincidente, las guides `always`, con `applies` coincidente o que refieren a una capability tocada, las preguntas abiertas conectadas con lo tocado o con el trabajo que lo implementa, y las señales de todo eso. Las fallas del grafo no detienen `context`: las reporta `doctor`.

### Paso 19. `ready` (hecho)

- **Objetivo:** saber qué trabajo se puede tomar.
- **Regla clave:** recorre `blocks` y `parent`; al cerrar un bloqueo, los dependientes quedan disponibles. El bloqueo se hereda a las subtareas, y una tarea con subtareas abiertas no aparece porque se trabaja en las hojas (`ledger`). Recibe la rama como parámetro.
- **Hecho cuando:** una pregunta abierta saca su trabajo de `ready`, y responderla lo devuelve.
- **Cómo quedó:** un work está listo si está `active` y nada abierto lo bloquea. Una task está lista si está `open`, o `claimed` en la misma rama que se consulta; si su work está listo; si no tiene subtareas abiertas; y si ni ella ni sus tareas padre están bloqueadas por una pregunta abierta o una tarea sin terminar. Los claims locales de `.git/beacon/` se suman en el paso 22.

### Paso 20. `brief` (hecho)

- **Objetivo:** el resumen que recibe el agente al empezar.
- **Regla clave:** tope de `limits.brief` líneas. La meta activa siempre entra, y el resto se recorta en este orden fijo:
  1. Bloqueos y rojos.
  2. Trabajo en curso.
  3. Pendientes de sí (vacío hasta que exista Weaver, D4).
  4. Deriva y promesas sin comprobar aquí.
  5. Índice de guides.
- **Hecho cuando:** con exceso de contenido nunca se pierde lo crítico y aparece `N more, run beacon status`.
- **Cómo quedó:** la meta y sus `nongoals` entran siempre, aunque superen el tope. El resto se ordena por la prioridad fija y se corta desde el final, reservando una línea para el aviso. Las promesas sin resultado se resumen en una sola línea.
- **Decidido (2026-11-24):** las guides `always` solo tienen su línea del índice en el `brief`; su cuerpo se lee bajo demanda. `guides.md` quedó alineado.
- **Pendiente:** el aviso nombra `beacon status`, que todavía no existe (Weaver). Hasta entonces ese mensaje no cumple P7.

### Paso 21. Git: `drift`, `stale` y `diff` (hecho)

- **Objetivo:** cruzar el estado con la historia del código.
- **Construye:** `system/git.go`, `driver/commit`, y `action/drift.go`, `stale.go` y `diff.go`.
- **Reglas clave:**
  - Consultas **en lote**: una por commit base distinto y una de frescura. Nunca una por promesa.
  - El cruce con `paths` se hace en memoria.
  - `git` se ejecuta sin shell.
- **Depende de:** D4.
- **Hecho cuando:** los tests corren sobre repos temporales reales y 500 capabilities quedan cerca de los 150 a 200 ms de `layout`.
- **Decidido (2026-11-24):** una guide o decisión está desactualizada si pasó su plazo o si recibió 10 o más commits sobre sus paths desde la última revisión. La deriva cuenta los cambios sin commitear.
- **Cómo quedó:** `action.Analyze` junta carga, grafo, verificación, deriva y frescura, y lo usan `context` y `brief`. `driver/commit` hace las consultas a git en lote, con `core.quotepath=off` para nombres no ASCII, y valida commits y revisiones antes de pasarlos a git. `StateChanges` entrega los registros y capturas base que cambiaron desde una revisión, con su contenido antes y después, sin clasificar.
- **Medido (2026-11-24):** con 500 capabilities, 1.000 archivos, 50 commits y cambios sin commitear, el análisis completo con git real tarda unos 100 ms y el hook unos 40 ms, porque git corre en paralelo con la carga y el hook no consulta git.
- **Pendiente:** leer los globs `baseline` de `runners` en `config.yaml`, al conectar `diff` al CLI en Weaver.

### Paso 22. Estado local (hecho)

- **Objetivo:** compartir estado entre worktrees sin riesgo.
- **Construye:** `driver/states`, claims locales, `context --hook` y la limpieza de claims en `doctor`.
- **Reglas clave:** toda escritura en `.git/beacon/` es atómica, y la ruta sale de `git rev-parse --git-common-dir`.
- **Cuidado:** en Windows, renombrar sobre un archivo que otro proceso tiene abierto puede fallar, así que hace falta reintentar y probarlo con escritores concurrentes.
- **Hecho cuando:** varios procesos escriben a la vez sin dejar un archivo a medias, y `--hook` no repite lo ya inyectado.
- **Cómo quedó:** `driver/states` guarda `claims.json` e `injected.json` en `.git/beacon/`, con escritura atómica y un candado de archivo (`system.WithFileLock`), porque la atomicidad sola no evita que dos escritores se pisen. `ready` respeta los claims locales: una tarea reclamada en otra rama no aparece. `doctor` borra los claims cuya rama o worktree ya no existe y lo avisa. `HookContext` entrega solo lo que la sesión no recibió, con tope y aviso, y registra lo entregado.
- **Pendiente:** escribir un claim (`beacon claim`) y el comando `context --hook` son de Weaver; Spider ya los soporta.

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
| El dominio en Go resulta verboso por la regla | Se evalúa al cerrar H1, antes de replicarlo en Weaver |
| Spider crece antes de que Weaver valide su interfaz | El Bloque 3 se construye contra los comandos definidos en `weaver` |
| Sobreingeniería | Cada paso implementa solo lo que la KB pide |
| Benchmarks sintéticos | Se miden de nuevo en los pasos 16 y 21, con el código real |

---

# Parte II. Fase 2: Weaver

> **Estado: planificada el 2026-11-25, sin empezar.** Las decisiones D5 a D14 se cerraron con la persona y ya están en la KB. Lo marcado como **Propuesta** se decidió al escribir el plan y se confirma al revisarlo.

## 9. Panorama

### Qué es Weaver

Weaver es el binario `beacon` sobre Spider: el CLI, la clasificación de cambios por nivel y la ejecución de checks. Spider sabe **qué hay y qué cambió**. Weaver decide **qué nivel tiene un cambio, si necesita un sí y si una promesa se cumple**, y lo entrega por el sobre `--json` con un `hint` que funciona (P7).

| Tema | Archivo de la KB |
|---|---|
| Comandos, sobre `--json`, códigos de salida | `.claude/delta/weaver.md` |
| Niveles, aprobación, riesgo | `.claude/delta/deltas.md` |
| Checks, ejecutores, estados, deriva | `.claude/delta/pledge.md` |
| `results.json`, archivo, estado local | `.claude/delta/layout.md` |
| Campos y `config.yaml` | `.claude/delta/ledger.md` |
| Claims y ramas largas | `.claude/delta/method.md` |

### Alcance

**Entra en la Fase 2** (`status`):

| Comando | Qué hace al cerrar la fase |
|---|---|
| `beacon brief` | El `brief` de Spider con la sección "Pendientes de sí" llena |
| `beacon context <paths>` | Lo que aplica a esos archivos; con `--hook`, solo lo nuevo de la sesión |
| `beacon record -` | Crea o reemplaza un registro desde stdin, con niveles; `--approve` transmite el sí |
| `beacon close <id> --reason <texto>` | Cierra y archiva |
| `beacon get <id>` | Detalle de un registro, activo o archivado |
| `beacon find <texto>` | Búsqueda de registros |
| `beacon check [capability]` | Corre los checks y guarda los resultados |
| `beacon diff` | Cambios de estado desde la base, clasificados por nivel |
| `beacon status` | Rojos, deriva, bloqueos y pendientes de sí (D5) |
| `beacon ready` | Trabajo listo en la rama actual |
| `beacon claim <task>` | Reclama una tarea en los dos sitios |

**Queda fuera:** `verify` y el cálculo de riesgo (Fase 4, con Snares), `attest` (Fase 4), `snares ci` (Fase 4), `install` y la guía de redacción en la respuesta de `record` (Fase 3), `stale` como comando y la frescura de guides (Fase 5), `migrate`, `stats` y las métricas de fricción (Fase 6).

### Qué existirá al terminar

| Al cerrar | Qué se puede hacer |
|---|---|
| **Bloque 4: Spider para Weaver** | Spider expone una fachada pública, escribe registros, archiva y guarda resultados de checks con huella |
| **Bloque 5: dominio de Weaver** | Los niveles y la construcción de invocaciones se prueban como dominio puro |
| **Bloque 6: comandos** | Los once comandos de la tabla funcionan con `--json`, señales y códigos 0, 1 y 2 |
| **Bloque 7: cierre** | Ramas largas probadas con casos reales y tiempos medidos |

### Cómo fluye la información

```text
router (handle/*)          lee argumentos y stdin, imprime el sobre
        │
        ├──────────── spider.go   lecturas puras: get, find, context, ready, fmt, doctor
        ▼
weaver.go                  la única puerta de Weaver
        │
        ▼
action/*                   record, close, claim, check, diff, brief, status
        │
        ├──→ domain/deltas     nivel de cada cambio y si pide un sí
        ├──→ domain/checks     invocación de un ejecutor y su resultado
        │
        ├──── driver/facade    adapta spider.go a los contratos de Weaver
        ├──── driver/runner    ejecuta el ejecutor sin shell, con tope de tiempo
        └──── driver/branch    resuelve la rama base con git merge-base
```

### Quién llama a quién

- `router` llama a **Spider directo** para las lecturas que no necesitan niveles ni checks (`get`, `find`, `context`, `ready`, `fmt`, `doctor`).
- `router` llama a **Weaver** para todo lo que clasifica, ejecuta o escribe (`record`, `close`, `claim`, `check`, `diff`, `brief`, `status`). `brief` y `status` van por Weaver porque necesitan los pendientes de sí (D4).
- Weaver solo ve los **tipos públicos de `spider.go`**, nunca los de `spider/internal` (D15).

## 10. Decisiones

D5 a D14 se cerraron con la persona el 2026-11-25. D15 a D20 son **Propuestas** del plan.

### D5. `beacon status` entra en la Fase 2

- **Por qué:** el `brief` recortado dice `N more, run beacon status`. Sin el comando, ese mensaje no cumple P7.
- **Costo:** bajo; reutiliza `action.Analyze` y los pendientes de sí.
- **Afecta a:** paso 35.

### D6. El archivo (`archive/`) pasa de la Fase 5 a la Fase 2

- **Por qué:** `close` debe archivar, y sin archivo un registro cerrado rompe las referencias que apuntan a él.
- **Formato:** el de `layout`, sección "Archivo de lo cerrado".
- **Consecuencia:** la Fase 5 queda con frescura de guides y `stale`.
- **Afecta a:** pasos 25 y 31.

### D7. `record` recibe el registro completo

- **Decisión:** el payload reemplaza el registro entero y tiene la misma forma que devuelve `get --json`. No hay parches.
- **Por qué:** es determinista y no hace falta definir cómo se parchean listas ni promesas. Omitir una promesa no la borra en silencio: lo rechaza D18.
- **Afecta a:** pasos 24 y 30.

### D8. El sí se transmite reenviando el mismo payload con `--approve`

- **Decisión:** un `record` con un cambio que pide sí sale con código 2 y no guarda nada. Si la persona dice sí, el agente repite **el mismo payload** con `--approve`. No hay estado pendiente ni tokens.
- **Por qué:** el sí aplica exactamente al contenido reenviado, y no queda nada que limpiar si la persona dice que no.
- **Afecta a:** pasos 30 y 31.

### D9. `results.json` guarda el commit y una huella de lo no commiteado

- **Problema:** la deriva compara el commit del último check con el árbol de trabajo. Un check corrido con cambios sin commitear y guardado solo con `HEAD` mostraría deriva en cuanto termina.
- **Decisión:** cada resultado guarda `HEAD` y la huella de los archivos de sus `paths` que difieren de `HEAD` en ese momento. Hay deriva solo si el conjunto actual de archivos cambiados desde ese commit, con su contenido, difiere de la huella. El formato está en `layout`.
- **Consecuencia:** commitear exactamente lo que se comprobó no produce deriva.
- **Afecta a:** paso 26.

### D10. La rama base sale de `git merge-base`

- **Decisión:** la base es `git merge-base HEAD <base>`, donde `<base>` es `origin/HEAD` si existe y si no `main` o `master`, en ese orden. `--base <ref>` la fuerza. Sin base, el comando sale con código 1 y el `hint` pide `--base`.
- **Por qué:** es el punto de origen de la rama, que es la hipótesis de ramas largas (`method`), y no suma claves a `config.yaml`.
- **Afecta a:** pasos 33, 34 y 36.

### D11. Tope fijo de 10 minutos por invocación de ejecutor

- **Decisión:** si se agota, las promesas de esa invocación quedan `Violated` con razón `timeout`. No se configura.
- **Cuándo se revisa:** si la fricción medida muestra suites legítimas de más de 10 minutos. Queda anotado en `status`.
- **Afecta a:** pasos 28 y 33.

### D12. Spider guarda `results.json`

- **Decisión:** Spider es dueño de todo el estado local (`claims.json`, `injected.json` y ahora `results.json`). Weaver ejecuta los checks y entrega los veredictos a `spider.go`, que calcula el commit y la huella y escribe con escritura atómica y candado.
- **Por qué:** `ProofReader` es un contrato interno de Spider, y Weaver no puede implementarlo desde fuera de `spider/internal`.
- **Afecta a:** paso 26.

### D13. `config.yaml` acepta `approved`

- **Decisión:** un cambio de `runners` o `schema` se aprueba sellando `approved: {by, at, via}` en `config.yaml`, con `beacon record --approve -` y un payload de tipo `config`.
- **Por qué:** es el mismo mecanismo que los registros, y Snares podrá exigirlo en CI (Fase 4).
- **Afecta a:** pasos 26 y 30.

### D14. `check` con alguna promesa en rojo sale con código 2

- **Decisión:** `ok: true`, los rojos en `signals` y código 2. El `hint` nombra la promesa que falló y el comando exacto para reintentar.
- **Por qué:** un script o el pre-commit lo detectan sin leer el JSON, y el agente lo lee como "bloqueado hasta arreglar el código".
- **Afecta a:** paso 33.

### D15. Frontera pública de Spider con tipos planos (**Propuesta**)

- **Decisión:** `spider.go` expone structs públicos sin lógica (`RecordView`, `StateChange`, `ProofVerdict`, `RunnerConfig`...), definidos en un segundo archivo del paquete raíz, `views.go`. Ningún tipo de `spider/internal` cruza la frontera.
- **En Weaver:** `driver/facade` convierte esos structs en valores de su dominio pasando por sus parsers, como cualquier dato que entra de fuera (`code-domain-layer`).
- **Por qué:** con alias a tipos internos, la API de Spider quedaría atada a su dominio.
- **Consecuencia:** la regla del `README` "cada pieza se expone por su archivo de entrada" pasa a decir "por su paquete raíz". Se ajusta en el paso 23.

### D16. Forma del payload (**Propuesta**)

```json
{"type": "capability", "record": {"id": "bc-7f3a", "name": "...", "promises": []}}
```

- `type` es uno de los siete tipos de registro, o `config`. Las claves de `record` son las del YAML (`ledger`), incluidos los campos desconocidos, que se conservan.
- **Lo pone el binario, nunca el payload:** el `id` de un registro o una promesa nuevos (D2), `created`, `author` y `approved`. En un registro existente se conservan los valores guardados; si el payload trae otros, se ignoran y la respuesta lo avisa.
- **Una guide** lleva además `body` con el Markdown. Se escribe como frontmatter más cuerpo.
- **Los valores se leen por el tipo del esquema**, igual que en el YAML: una fecha o una versión llega como string.
- `beacon record -` lee de stdin: el `-` es el argumento posicional y `--json` sigue siendo solo el formato de salida. Así, el `beacon record --json -` de `weaver` funciona sin cambios.

### D17. Nivel de cada cambio (**Propuesta**)

Es la regla mecánica de `deltas`, completada para los casos que la KB no nombra. El atajo del texto normalizado aplica a todo cambio de texto.

| Cambio | Nivel | ¿Pide sí? |
|---|---|---|
| Texto o `check` de una promesa `active` | 3 | Sí |
| Promesa `active` que pasa a `proposed` o se cierra | 3 | Sí |
| Promesa que pasa de `proposed` a `active` | 2 | Sí (`deltas`) |
| Promesa nueva (`proposed`) | 2 | No |
| Goal nuevo, o cambio de su `statement`, `success` o `status` | 3 | Sí |
| `nongoals` de un goal o una capability, `rejected` de una decisión | 1 | No |
| Capability nueva sin promesas activas | 2 | No |
| `name`, `intent`, `paths` o `serves` de una capability | 1 | No |
| Decisión nueva | 2 | No |
| Cualquier otro cambio en una decisión, pregunta, guide, work o task | 1 | No |
| `runners` o `schema` de `config.yaml` | 3 | Sí |
| Otras claves de `config.yaml` | 1 | No |
| Un archivo cubierto por un `baseline` añadido, cambiado o borrado | 3 | Sí |

- **Riesgo conocido:** estrechar los `paths` de una capability es nivel 1, aunque puede ocultar deriva. Si en el uso real resulta un agujero, pasa a nivel 3 (anotado en `status`).

### D18. Las promesas no se borran por omisión (**Propuesta**)

- **Decisión:** un payload de capability que omite una promesa existente se rechaza con código 1 (`promise_omitted`). El `hint` dice que se cierre con `beacon close <promesa> --reason <texto>`.
- **Por qué:** P5. Lo que deja de valer se cierra con motivo y se archiva; no desaparece.

### D19. Cuándo un cambio de nivel 3 está aprobado (**Propuesta**)

Sirve para la señal "Pendiente de sí" (`brief`, `status`, `diff`), sobre todo con ediciones a mano que no pasan por `record`.

- **Regla del sí fresco:** un cambio que pide sí, entre la base y el árbol de trabajo, está aprobado si en ese mismo rango cambió el `approved` del elemento afectado: la promesa, el goal o `config.yaml`.
- **Capturas base:** cuentan como aprobadas si en el mismo rango cambió el `approved` de alguna capability con una promesa que usa ese ejecutor.
- **Es un marcado, no una prueba** (`deltas`): la garantía en equipo sigue siendo la review del PR.

### D20. Cómo se corren los checks (**Propuesta**)

- **Una invocación por combinación distinta de ejecutor y parámetros.** Si dos promesas comparten `{runner, pkg, target}`, el ejecutor corre una vez y el veredicto vale para las dos.
- **En serie**, en orden de id. El paralelismo se evalúa con la medición del paso 37.
- **Directorio:** la raíz del repo. **Entorno:** el del proceso, sin cambios.
- **Salida:** en modo texto, la del ejecutor pasa tal cual a stderr. Con `--json`, `data` incluye las últimas 40 líneas de cada invocación que falló, para que el agente vea el error sin volver a correrla.
- **Promesas manuales:** se saltan y se listan como pendientes de `attest` (Fase 4).
- **Promesas `proposed`:** también se comprueban; su check es un dato como cualquier otro.

## 11. Estructura objetivo

Lo que **se agrega o cambia** en la Fase 2. Los números son el paso que crea cada parte, y los tests van junto a cada archivo.

```text
beacon/
├─ source/
│  ├─ router/
│  │  └─ internal/
│  │     ├─ handle/
│  │     │  ├─ get.go                # 29
│  │     │  ├─ find.go               # 29
│  │     │  ├─ context.go            # 29 · también --hook
│  │     │  ├─ ready.go              # 29
│  │     │  ├─ record.go             # 30
│  │     │  ├─ close.go              # 31
│  │     │  ├─ claim.go              # 32
│  │     │  ├─ check.go              # 33
│  │     │  ├─ diff.go               # 34
│  │     │  ├─ brief.go              # 35
│  │     │  └─ status.go             # 35
│  │     └─ output/
│  │        ├─ output.go             # 29 · código 2 y señales en el sobre
│  │        └─ signals.go            # 29 · señal a JSON y a texto
│  │
│  ├─ spider/
│  │  ├─ spider.go                   # 23 · nuevas operaciones públicas
│  │  ├─ views.go                    # 23 · tipos públicos planos (D15)
│  │  └─ internal/
│  │     ├─ domain/ledger/
│  │     │  └─ capability_proofs.go  # 26 · ProofResult con huella y razón
│  │     ├─ action/
│  │     │  ├─ find.go               # 23
│  │     │  ├─ save.go               # 24 · validar y escribir un registro
│  │     │  ├─ close.go              # 25
│  │     │  └─ drift.go              # 26 · deriva con huella (D9)
│  │     └─ driver/
│  │        ├─ record/
│  │        │  ├─ payload.go         # 24 · JSON a nodos y de vuelta
│  │        │  ├─ writer.go          # 24 · escritura canónica y atómica
│  │        │  └─ config.go          # 26 · runners, baseline y approved
│  │        ├─ archive/
│  │        │  └─ archive.go         # 25 · segmentos .jsonl
│  │        ├─ commit/
│  │        │  └─ commit.go          # 26 · huella y capturas base
│  │        └─ states/
│  │           └─ results.go         # 26 · results.json
│  │
│  └─ weaver/
│     ├─ weaver.go                   # 27 · fachada pública de la pieza
│     └─ internal/
│        ├─ domain/
│        │  ├─ deltas/               # nivel de un cambio y su aprobación
│        │  │  ├─ change_vos.go        # 27 · ChangeKind, ChangeLevel, StateChange
│        │  │  ├─ change_policies.go   # 27 · ClassifyChange, NormalizeText (D17)
│        │  │  ├─ change_errors.go     # 27
│        │  │  ├─ approval_policies.go # 27 · StampApproval, IsFreshlyApproved
│        │  │  └─ approval_errors.go   # 27
│        │  └─ checks/               # cómo se invoca un ejecutor
│        │     ├─ runner_vos.go        # 28 · RunnerName, RunnerSpec
│        │     ├─ runner_policies.go   # 28 · BuildInvocation, GroupInvocations
│        │     ├─ runner_errors.go     # 28
│        │     ├─ runner_executor.go   # 28 · contrato Executor
│        │     └─ outcome_vos.go       # 28 · CheckOutcome, OutcomeReason
│        ├─ action/
│        │  ├─ record.go             # 30
│        │  ├─ close.go              # 31
│        │  ├─ claim.go              # 32
│        │  ├─ check.go              # 33
│        │  ├─ pending.go            # 34 · pendientes de sí (D19)
│        │  ├─ diff.go               # 34
│        │  ├─ brief.go              # 35
│        │  └─ status.go             # 35
│        └─ driver/
│           ├─ facade/
│           │  └─ facade.go          # 27 · spider.go a contratos de Weaver
│           ├─ runner/
│           │  └─ runner.go          # 33 · os/exec con tope de 10 min
│           ├─ branch/
│           │  └─ branch.go          # 33 · merge-base (D10)
│           └─ double/
│              └─ double.go          # 27 · contratos en memoria para tests
│
└─ testing/
   └─ e2e/
      ├─ read_test.go                # 29 · get, find, context, ready
      ├─ record_test.go              # 30
      ├─ close_test.go               # 31
      ├─ claim_test.go               # 32
      ├─ check_test.go               # 33
      ├─ diff_test.go                # 34
      ├─ brief_test.go               # 35
      └─ branch_test.go              # 36 · ramas largas
```

### Dónde va cada tipo de código

| Si el código... | Va en |
|---|---|
| Decide el nivel de un cambio o si está aprobado | `weaver/internal/domain/deltas` |
| Arma una invocación o interpreta su resultado | `weaver/internal/domain/checks` |
| Lee o escribe `.beacon/` o `.git/beacon/` | Spider (`driver/record`, `driver/archive`, `driver/states`) |
| Ejecuta un proceso externo | `weaver/internal/driver/runner` |
| Consulta git para Weaver | `weaver/internal/driver/branch`, sobre `system.Git` |
| Traduce tipos de `spider.go` a Weaver | `weaver/internal/driver/facade` |

### Quién puede importar a quién

```text
cmd/beacon → router → weaver.go → action → domain → kernel
                │                     │
                │                     └──→ contratos de deltas y checks
                │                                  ▲
                │              driver ─────────────┘ (los implementan)
                │                 │
                │                 ├──→ spider.go (solo driver/facade)
                │                 └──→ system, os/exec
                └──→ spider.go (lecturas puras)
```

`weaver/internal/domain` nunca importa `spider`: lo que necesita de Spider llega convertido por `driver/facade`.

## 12. Bloque 4: Spider para Weaver

**Rama:** `feat/spider-writes`, desde `main` actualizado. **Forma de trabajo:** la misma de la Fase 1: primero el test que falla, luego el código mínimo, y al final `go test ./...` en verde.

### Paso 23. Fachada pública

- **Objetivo:** que Weaver y `router` usen todo lo que Spider ya calcula sin tocar su `internal`.
- **Construye:** `views.go` con los tipos planos (D15); en `spider.go`, `Brief(ctx, pending)`, `Context(ctx, paths)`, `HookContext(ctx, session, paths)`, `Ready(ctx, branch)`, `Analyze(ctx)`, `Get(ctx, id)`, `Find(ctx, text)`, `StateChanges(ctx, revision)` y `Runners(ctx)`; y `action/find.go`.
- **`find` (Propuesta):** busca una subcadena, sin distinguir mayúsculas ni acentos, en los campos de texto (`statement`, `name`, `intent`, `text`, `what`, `why`, `request`, `title`) de los registros activos; con `--archived`, también en `archive/`. Devuelve como máximo 20 resultados, con el aviso `N more`.
- **`Brief(ctx, pending)`** recibe las líneas de pendientes de sí ya armadas por Weaver y las pone en su prioridad (D4).
- **Cuidado:** `Analyze` lo usan `brief`, `status` y `check`. Se arma una sola vez por comando y se reutiliza.
- **Hecho cuando:** un test externo (`spider_test`) usa cada operación sin importar nada de `spider/internal`, y el `README` dice "por su paquete raíz".

### Paso 24. Escritura de registros

- **Objetivo:** crear y reemplazar registros con las mismas garantías que la lectura.
- **Construye:** `driver/record/payload.go`, `driver/record/writer.go` y `action/save.go`.
- **Flujo:**
  1. El payload JSON (D16) se convierte en nodos YAML (JSON es YAML válido) y pasa por **los mismos `decode<Tipo>`** que un archivo. No hay un segundo parser.
  2. El binario pone el `id` (si es nuevo, con `IDIssuer`), `created` y `author`, y conserva `approved` (D16). Las promesas nuevas sin `id` reciben uno.
  3. Se valida el **grafo completo** con el registro reemplazado. Una referencia rota o un ciclo rechazan la escritura (`graph_invalid`, código 1, con detalles).
  4. Se escribe con el formato canónico de `fmt`, escritura atómica y candado sobre el archivo.
- **Dos fases separadas:** `Prepare(ctx, payload)` devuelve el antes y el después como `StateChange` sin escribir nada, y `Save(ctx, prepared, stamps)` escribe. Weaver clasifica entre las dos.
- **`get --json`** devuelve la forma del payload, así que `get` → editar → `record` es un ciclo cerrado.
- **Cuidado:** si cambia el slug (porque cambió el título), el archivo se renombra; el id nunca cambia. Un payload con un `id` que no existe da `record_missing`, no crea nada.
- **Hecho cuando:** `get` seguido de `record` con el mismo payload no cambia ningún byte, un campo desconocido sobrevive, y un payload inválido falla con el mismo código y la misma ruta de campo que el archivo equivalente.

### Paso 25. Archivo y `close`

- **Objetivo:** cerrar sin perder historia ni romper referencias (D6, P5).
- **Construye:** `driver/archive/archive.go` y `action/close.go`. La validación del grafo (`action/validate.go`) aprende a resolver referencias contra el archivo.
- **Reglas:**
  - `close` exige `--reason`. Pone `closed` (hoy) y `reason`, añade la línea al segmento del mes de `closed` y borra el archivo de `records/`. El formato está en `layout`.
  - Una promesa se cierra por su id: sale de su capability y va sola al archivo, con el id de la capability.
  - **Las guides no se cierran** (`guide_not_archivable`): el `hint` dice que se borre el archivo, porque git guarda la historia (`layout`).
  - **Una tarea con subtareas abiertas no se cierra** (`ledger`); el `hint` lista las subtareas.
  - Cerrar algo que pide sí (una promesa `active`, un goal) sigue D8: código 2 y después `--approve`.
- **Lectura perezosa:** `archive/` solo se lee cuando una referencia no se resuelve en `records/`, o cuando `get` o `find --archived` lo piden (`layout`).
- **Cuidado:** el segmento se reescribe completo con escritura atómica y candado, para que dos `close` simultáneos no pierdan una línea.
- **Hecho cuando:** cerrar un registro referenciado deja el grafo válido, `get` lo encuentra archivado y dos `close` concurrentes dejan las dos líneas.

### Paso 26. Resultados, huella y `config.yaml`

- **Objetivo:** que los resultados de los checks existan de verdad y que la deriva no mienta con cambios sin commitear (D9, D12).
- **Construye:**
  - `driver/states/results.go`: el `ProofReader` real y su escritura.
  - `ProofResult` con `Footprint` y `Reason`. `VerifyPromise` trata `timeout` como un fallo.
  - En `driver/commit`: la huella (los archivos de unos `paths` cambiados respecto de un commit, con el SHA-256 de su contenido) y los archivos de `baseline` en `StateChanges`. Esto cierra el pendiente del paso 21.
  - En `driver/record/config.go`: `runners` parseados (`run` y `baseline`) y la clave `approved` (D13).
  - En `action/drift.go`: hay deriva si la huella actual difiere de la guardada.
- **`RecordResults(ctx, verdicts)`** en `spider.go`: recibe `{promesa, outcome, reason}`, calcula `HEAD` y la huella **una vez por capability**, y fusiona en `results.json` bajo candado.
- **Cuidado:** la huella se calcula sobre los archivos del diff que ya se obtuvo en lote (`layout`), sin una consulta a git por promesa. Un archivo borrado entra en la huella como borrado.
- **Hecho cuando:** (a) un check con cambios sin commitear no muestra deriva; (b) editar después uno de esos archivos sí la muestra; (c) commitear exactamente lo comprobado no la muestra; (d) dos escritores concurrentes no pierden resultados.

## 13. Bloque 5: dominio de Weaver

**Rama:** `feat/weaver-cli`, desde `main` con el Bloque 4 fusionado.

### Paso 27. Pieza Weaver y contexto `deltas`

- **Objetivo:** que el nivel de un cambio sea una función pura y probada.
- **Construye:** `weaver.go` (solo el constructor), `driver/facade`, `driver/double` y el contexto `deltas`.
- **`ClassifyChange(StateChange) (ChangeLevel, bool)`:** devuelve el nivel y si pide sí, según la tabla de D17.
- **`NormalizeText`:** pasa a minúsculas, quita la puntuación y colapsa los espacios (`deltas`). Un cambio que desaparece al normalizar es de nivel 1.
- **`StampApproval(actor, human, today)`:** si el actor es una persona, `by` es el actor y no hay `via`; si es un agente, `by` es `human:<user.name de git>` y `via` es el agente. Sin `user.name`, falla con `identity_missing`.
- **`IsFreshlyApproved`:** la regla del sí fresco (D19).
- **Cuidado:** el dominio recibe el `user.name` y la fecha como datos; no lee git ni el reloj (`code-domain-layer`).
- **Hecho cuando:** cada fila de D17 tiene un caso, y el ejemplo de `deltas` (de "5 intentos" a "10 intentos") da nivel 3.

### Paso 28. Contexto `checks`

- **Objetivo:** que armar una invocación sea seguro antes de ejecutar nada.
- **Construye:** el contexto `checks` y el contrato `Executor`, con `Run(ctx, argv, dir) (ExecutionResult, error)`.
- **`BuildInvocation(spec, params)`:** sustituye `{param}` en cada elemento de `run`, sin shell.
  - Falla si falta un parámetro que la plantilla usa (`param_missing`), o si sobra uno que no usa (`param_unused`), para atrapar typos.
  - Un parámetro que empieza con `-` se rechaza de nuevo aquí, aunque el registro ya lo validó (`pledge`).
- **`GroupInvocations`:** agrupa las promesas por invocación idéntica (D20).
- **`CheckOutcome`:** `passed` o `failed`, con razón `exit_code` o `timeout` (D11).
- **Hecho cuando:** no se puede construir ningún argv con un parámetro inyectado como opción, y dos promesas con el mismo check comparten invocación.

## 14. Bloque 6: comandos

Cada comando tiene un test e2e que corre el binario real sobre un repo temporal (`testing/utils`).

### Paso 29. Salida, señales y lecturas

- **Objetivo:** completar el sobre `--json` y poner en marcha las lecturas puras.
- **Construye:** el código 2 (`ExitBlocked`) en `output`, `signals.go` y los handlers de `get`, `find`, `context` (con `--hook`) y `ready`.
- **Señal en JSON:** `{"kind": "red", "record": "bc-c3d4", "related": null, "relation": null}`, sacada del `Signal` de Spider. En texto, una línea por señal.
- **`ready`** toma la rama actual de git. Con HEAD separado, usa el commit y lo avisa.
- **`context --hook`** necesita un id de sesión, que recibe con `--session <id>`. De dónde lo saca cada agente se decide en la Fase 3 (abierto en `status`).
- **Hecho cuando:** los cuatro comandos devuelven el sobre con `signals` llenas, y un archivo sin dueño recibe en `context` solo la meta y las guides `always`.

### Paso 30. `record`

- **Flujo:** leer stdin → `Prepare` → `ClassifyChange` en cada cambio → si alguno pide sí y no hay `--approve`, salir con código 2 sin escribir → con `--approve`, sellar `approved` en cada elemento afectado → `Save`.
- **Respuesta:** `data` con el id, el archivo, los cambios con su nivel y los campos del payload que se ignoraron (D16).
- **Errores y `hint`** (todos en inglés, P7):

| Código | Salida | `hint` |
|---|---|---|
| `payload_invalid` | 1 | `fix the fields listed in details and rerun` |
| `record_missing` | 1 | `run beacon find <text> to get the id` |
| `promise_omitted` | 1 | `keep the promise, or run beacon close <promise-id> --reason <text>` |
| `graph_invalid` | 1 | `fix the references listed in details and rerun` |
| `approval_required` | 2 | `ask the person to approve the changes in details; if they say yes, rerun the same command with --approve` |
| `identity_missing` | 1 | `set git config user.name, then rerun` |

- **Hecho cuando:** el ejemplo de `deltas` se bloquea con código 2, el mismo payload con `--approve` se guarda con `approved: {by, at, via}`, y un payload de config con `runners` cambiados sigue el mismo camino (D13).

### Paso 31. `close`

- **Construye:** el handler y el `action/close.go` de Weaver, que clasifica antes de archivar.
- **Hecho cuando:** cerrar una promesa `active` pide sí, cerrar una pregunta no lo pide, y cerrar una guide da `guide_not_archivable` con su `hint`.

### Paso 32. `claim`

- **Flujo:** leer el claim local → si otra rama o worktree tiene la tarea, salir con código 2 (`claim_taken`) y un `hint` que diga cuál → si no, escribir el claim local y después `owner` y `branch` en la tarea con `Save` (nivel 1).
- **`--force`:** la persona puede tomar una tarea reclamada (`method`: un claim no es un candado). El agente solo lo usa si la persona lo pidió, y la guía de Anchor lo dirá (Fase 3).
- **Cuidado:** si el registro no se puede escribir, se deshace el claim local, para que los dos sitios no queden desalineados.
- **Hecho cuando:** si dos worktrees reclaman la misma tarea, el segundo recibe código 2, y `ready` en el segundo ya no la ofrece.

### Paso 33. `check`

- **Construye:** `driver/runner`, `driver/branch` y `action/check.go`.
- **Selección:**
  - Sin argumentos: las capabilities cuyos `paths` tocan algún archivo cambiado desde la base (D10), incluidos los no commiteados.
  - Con un id: esa capability.
  - Con `--all`: todas.
- **Ejecución:** según D20, con el tope de D11. Al agotarse el tope, se mata el proceso y su árbol (en Windows, con `cmd.Cancel` y `WaitDelay`; ver Riesgos).
- **Al terminar:** `RecordResults` (paso 26) y un `Analyze` nuevo para las señales.
- **Salida:** código 0 sin rojos y código 2 con alguno (D14), con `hint: fix the code until beacon check <capability> passes; do not edit the promise`.
- **Errores:** `runner_missing` (el check nombra un ejecutor que no está en `config.yaml`), `runner_unavailable` (el programa no está instalado; la promesa queda `Unverified` con esa razón, no `Violated`) y `base_missing` (D10).
- **Hecho cuando:** en un repo con un ejecutor de prueba (un binario Go pequeño en `testdata`), una promesa pasa, otra falla con código 2, otra agota un tope reducido en el test y queda con `timeout`, y no aparece deriva justo después.

### Paso 34. `diff` y pendientes de sí

- **Construye:** `action/pending.go` y `action/diff.go`.
- **`diff`:** toma `StateChanges` desde la base (o desde `--since <rev>`), clasifica cada cambio con `ClassifyChange`, agrupa por nivel y marca cuáles piden sí y cuáles ya tienen el sí fresco (D19).
- **`pending`:** los cambios que piden sí y todavía no tienen el sí fresco. Lo usan `diff`, `brief` y `status`, y alimenta la señal "Pendiente de sí".
- **Hecho cuando:** una edición a mano del texto de una promesa `active` aparece como pendiente, y deja de aparecer después de `record --approve`.

### Paso 35. `brief` y `status`

- **`brief`:** el `brief` de Spider con los pendientes en su prioridad (paso 23). El aviso `N more, run beacon status` ya cumple P7.
- **`status`:** rojos, deriva, bloqueos y pendientes de sí, sin tope de líneas. Sale siempre con código 0: informa, no bloquea.
- **Hecho cuando:** un `brief` recortado nombra `beacon status`, y `status` muestra todo lo que el `brief` recortó.

## 15. Bloque 7: cierre

### Paso 36. Ramas largas

- **Objetivo:** probar con casos reales la hipótesis de `method`.
- **Casos (e2e en `branch_test.go`):**
  1. Una rama cambia una promesa mientras `main` cambia otra: `diff` en la rama muestra solo la suya.
  2. `main` debilita una promesa con `approved` y la rama se rebasa: el cambio de `main` no aparece como pendiente en la rama.
  3. Las dos ramas cambian la misma promesa: git da conflicto en ese archivo y `doctor` lo reporta.
  4. Una rama vieja reclama una tarea que `main` ya cerró: `claim` falla con `record_missing` y el `hint` sugiere `beacon get <id>`, que la muestra archivada.
- **Hecho cuando:** los cuatro casos pasan, o su comportamiento queda escrito en `method` y ese archivo deja de decir "no está resuelto".

### Paso 37. Medición y KB

- **Medir:** con 500 capabilities, cuánto tardan `record` (incluida la validación del grafo completo), `status` y el costo propio de `check` sin contar el ejecutor. Se anota en `layout`, junto a las mediciones de la Fase 1.
- **Sincronizar:** `status` marca la Fase 2 como cerrada, y la KB recoge las Propuestas confirmadas (D15 a D20) y lo que cambie al implementar.
- **Hecho cuando:** los números están en `layout` y la KB no contradice el código.

## 16. Hitos y commits

| Hito | Pasos | Rama | Resultado verificable |
|---|---|---|---|
| H6. Spider para Weaver | 23 a 26 | `feat/spider-writes` | `get` → `record` idempotente; archivo y huella probados |
| H7. Dominio de Weaver | 27 y 28 | `feat/weaver-cli` | Tabla de niveles e invocaciones en verde |
| H8. Lecturas y escritura | 29 a 32 | `feat/weaver-cli` | `get`, `find`, `context`, `ready`, `record`, `close` y `claim` en e2e |
| H9. Checks y estado | 33 a 35 | `feat/weaver-cli` | `check`, `diff`, `brief` y `status` en e2e |
| H10. Fase 2 cerrada | 36 y 37 | `feat/weaver-cli` | Ramas largas probadas y tiempos en `layout` |

## 17. Riesgos de la Fase 2

| Riesgo | Qué lo mitiga |
|---|---|
| En Windows, matar un ejecutor no mata a sus hijos (un `pnpm` que lanza `node`) | `cmd.Cancel` más `WaitDelay`, y un test con un ejecutor que lanza un hijo. Si no alcanza, un job object de Windows en `driver/runner`, sin tocar el dominio |
| Validar el grafo completo en cada `record` es lento en repos grandes | Se mide en el paso 37; la carga ya ronda los 70 ms con 500 capabilities |
| El payload completo de una capability grande es largo para el agente | `get --json` da la base exacta; si molesta, se reconsidera el parche (D7) con datos de fricción |
| El agente escribe un `approved` que la persona no dio | El `approved` del payload se ignora (D16) y el sello solo sale de `--approve`. Sigue siendo un marcado (`deltas`) |
| La huella cuesta con muchos archivos sin commitear | Se calcula solo sobre los archivos del diff en lote que caen en los `paths` (paso 26) |
| Estrechar `paths` oculta deriva sin pedir sí | Riesgo aceptado en D17 y anotado en `status` |
| El tope fijo de 10 minutos se queda corto en suites grandes | D11: se vuelve configurable si la fricción lo pide |
