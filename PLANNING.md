# Plan de construcción: Fase 2 (Weaver)

Este plan cubre solo la fase en curso: en qué orden se construye y qué archivos salen. El detalle funcional vive en la KB (`.claude/delta/`), y todas las fases, en `status`.

La numeración de pasos y decisiones sigue la del plan de la Fase 1 (Spider), ya cerrada, para que las referencias de la KB y de los commits sigan valiendo: por eso los pasos empiezan en el 23 y las decisiones en D5. Ese plan está en el historial de git (`git show 967f5aa:PLANNING.md`).

> **Estado: en curso.** Pasos 23 a 25 hechos; sigue el paso 26. Las decisiones D5 a D20 están cerradas con la persona y registradas en la KB.

## 1. Panorama

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
- `router` llama a **Weaver** para todo lo que clasifica, ejecuta o escribe (`record`, `close`, `claim`, `check`, `diff`, `brief`, `status`). `brief` y `status` van por Weaver porque necesitan los pendientes de sí (Spider detecta qué cambió y Weaver lo clasifica, `deltas`).
- Weaver solo ve los **tipos públicos de `spider.go`**, nunca los de `spider/internal` (D15).

## 2. Decisiones

D5 a D14 se cerraron con la persona el 2026-11-25, y D15 a D20, propuestas en el plan, se aprobaron el mismo día.

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

### D15. Frontera pública de Spider con tipos planos

- **Decisión:** `spider.go` expone structs públicos sin lógica (`RecordView`, `StateChange`, `ProofVerdict`, `RunnerConfig`...), definidos en un segundo archivo del paquete raíz, `views.go`. Ningún tipo de `spider/internal` cruza la frontera.
- **En Weaver:** `driver/facade` convierte esos structs en valores de su dominio pasando por sus parsers, como cualquier dato que entra de fuera (`code-domain-layer`).
- **Por qué:** con alias a tipos internos, la API de Spider quedaría atada a su dominio.
- **Consecuencia:** la regla del `README` "cada pieza se expone por su archivo de entrada" pasa a decir "por su paquete raíz". Se ajusta en el paso 23.

### D16. Forma del payload

```json
{"type": "capability", "record": {"id": "bc-7f3a", "name": "...", "promises": []}}
```

- `type` es uno de los siete tipos de registro, o `config`. Las claves de `record` son las del YAML (`ledger`), incluidos los campos desconocidos, que se conservan.
- **Lo pone el binario, nunca el payload:** el `id` de un registro o una promesa nuevos (aleatorio, `layout`), `created`, `author` y `approved`. En un registro existente se conservan los valores guardados; si el payload trae otros, se ignoran y la respuesta lo avisa.
- **Una guide** lleva además `body` con el Markdown. Se escribe como frontmatter más cuerpo.
- **Los valores se leen por el tipo del esquema**, igual que en el YAML: una fecha o una versión llega como string.
- `beacon record -` lee de stdin: el `-` es el argumento posicional y `--json` sigue siendo solo el formato de salida. Así, el `beacon record --json -` de `weaver` funciona sin cambios.

### D17. Nivel de cada cambio

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

### D18. Las promesas no se borran por omisión

- **Decisión:** un payload de capability que omite una promesa existente se rechaza con código 1 (`promise_omitted`). El `hint` dice que se cierre con `beacon close <promesa> --reason <texto>`.
- **Por qué:** P5. Lo que deja de valer se cierra con motivo y se archiva; no desaparece.

### D19. Cuándo un cambio de nivel 3 está aprobado

Sirve para la señal "Pendiente de sí" (`brief`, `status`, `diff`), sobre todo con ediciones a mano que no pasan por `record`.

- **Regla del sí fresco:** un cambio que pide sí, entre la base y el árbol de trabajo, está aprobado si en ese mismo rango cambió el `approved` del elemento afectado: la promesa, el goal o `config.yaml`.
- **Capturas base:** cuentan como aprobadas si en el mismo rango cambió el `approved` de alguna capability con una promesa que usa ese ejecutor.
- **Es un marcado, no una prueba** (`deltas`): la garantía en equipo sigue siendo la review del PR.

### D20. Cómo se corren los checks

- **Una invocación por combinación distinta de ejecutor y parámetros.** Si dos promesas comparten `{runner, pkg, target}`, el ejecutor corre una vez y el veredicto vale para las dos.
- **En serie**, en orden de id. El paralelismo se evalúa con la medición del paso 37.
- **Directorio:** la raíz del repo. **Entorno:** el del proceso, sin cambios.
- **Salida:** en modo texto, la del ejecutor pasa tal cual a stderr. Con `--json`, `data` incluye las últimas 40 líneas de cada invocación que falló, para que el agente vea el error sin volver a correrla.
- **Promesas manuales:** se saltan y se listan como pendientes de `attest` (Fase 4).
- **Promesas `proposed`:** también se comprueban; su check es un dato como cualquier otro.

## 3. Estructura objetivo

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

## 4. Bloque 4: Spider para Weaver

**Rama:** `feat/spider-writes`, desde `main` actualizado. **Forma de trabajo:** primero el test que falla, luego el código mínimo, y al final `go test ./...` en verde.

### Paso 23. Fachada pública (hecho)

- **Objetivo:** que Weaver y `router` usen todo lo que Spider ya calcula sin tocar su `internal`.
- **Construye:** `views.go` con los tipos planos (D15); en `spider.go`, `Brief(ctx, pending)`, `Context(ctx, paths)`, `HookContext(ctx, session, paths)`, `Ready(ctx, branch)`, `Analyze(ctx)`, `Get(ctx, id)`, `Find(ctx, text)`, `StateChanges(ctx, revision)` y `Runners(ctx)`; y `action/find.go`.
- **`find`:** busca una subcadena, sin distinguir mayúsculas ni acentos, en los campos de texto (`statement`, `name`, `intent`, `text`, `what`, `why`, `request`, `title`) de los registros activos; con `--archived`, también en `archive/`. Devuelve como máximo 20 resultados, con el aviso `N more`.
- **`Brief(ctx, pending)`** recibe las líneas de pendientes de sí ya armadas por Weaver y las pone en su prioridad (`spider`).
- **Cuidado:** `Analyze` lo usan `brief`, `status` y `check`. Se arma una sola vez por comando y se reutiliza.
- **Hecho cuando:** un test externo (`spider_test`) usa cada operación sin importar nada de `spider/internal`, y el `README` dice "por su paquete raíz".
- **Cómo quedó:**
  - `spider.go` expone `Brief`, `Context`, `HookContext`, `Ready`, `Analyze`, `Find`, `Get` y `StateChanges`; las lecturas devuelven sus fallas como `Finding`, igual que `doctor`.
  - `Get` y `StateChanges` entregan cada registro como JSON con la forma del payload (D16). Una guide trae `id` (sacado del nombre del archivo) y `body`.
  - `Runners` pasa al paso 26, junto con el parseo de `runners` en `config.yaml`, que es donde se necesita.
  - Hasta el paso 26 no hay `results.json`: la fachada usa un lector sin resultados, así que toda promesa automática sale `Unverified` con `not-run-here`.
  - En una señal, `related` y `relation` se omiten cuando no aplican, en vez de ir como `null`.
  - El plegado de acentos pasó al kernel (`kernel.FoldText`), porque lo usan `find` y los slugs de archivo.
  - Si `config.yaml` falla, los límites quedan en sus valores por defecto y la falla se reporta, para que `brief` no quede con tope 0.

### Paso 24. Escritura de registros (hecho)

- **Objetivo:** crear y reemplazar registros con las mismas garantías que la lectura.
- **Construye:** `driver/record/payload.go`, `driver/record/writer.go` y `action/save.go`.
- **Flujo:**
  1. El payload JSON (D16) se convierte en nodos YAML (JSON es YAML válido) y pasa por **los mismos `decode<Tipo>`** que un archivo. No hay un segundo parser.
  2. El binario pone el `id` (si es nuevo, con `IDIssuer`), `created` y `author`, y conserva `approved` (D16). Las promesas nuevas sin `id` reciben uno.
  3. Se valida el **grafo completo** con el registro reemplazado. Una referencia rota o un ciclo rechazan la escritura (`graph_invalid`, código 1, con detalles).
  4. Se escribe con el formato canónico de `fmt`, escritura atómica y candado sobre el archivo.
- **Dos fases separadas:** `Prepare(ctx, payload)` devuelve el antes y el después como `StateChange` sin escribir nada, y `Save(ctx, prepared, stamps)` escribe. Weaver clasifica entre las dos.
- **`get --json`** devuelve la forma del payload, así que `get` → editar → `record` es un ciclo cerrado.
- **Cuidado:** el id nunca cambia. Un payload con un `id` que no existe da `record_missing` y no crea nada (salvo en una tarea; ver "Cómo quedó").
- **Hecho cuando:** `get` seguido de `record` con el mismo payload no cambia ningún byte, un campo desconocido sobrevive, y un payload inválido falla con el mismo código y la misma ruta de campo que el archivo equivalente.
- **Cómo quedó:**
  - El payload se lee con el parser YAML estricto y se **fusiona nodo por nodo** sobre el archivo existente: lo que no cambió conserva su nodo original, con sus comentarios. Las promesas se emparejan por id.
  - `Prepare` devuelve el antes y el después como `Record`, los campos ignorados y si el registro queda igual; `Save` sella los `approved` pedidos y escribe. `spider.Approval` es `kernel.ApprovalInput`.
  - Un payload inválido da `payload_invalid`, y en `details.findings` va cada falla con el mismo código que daría el archivo (por ejemplo `question_field_missing`). `record_missing`, `promise_omitted` y `promise_unknown` (una promesa con un id que el registro no tiene) llegan tal cual.
  - **El archivo no se renombra al actualizar**, aunque cambie el título: un nombre estable no ensucia el diff, y el slug de los registros existentes no siempre sale del título (el del goal de ejemplo no sale de `statement`). Solo un registro nuevo toma su slug del título.
  - **Una tarea nueva** sin `id` toma el siguiente hijo libre de su work (`bc-91cd.2`). Es el único tipo que acepta un `id` nuevo explícito, porque su jerarquía sale del id (`ledger`); para crear una subtarea se manda `bc-91cd.1.1`.
  - En una guide, `body` ausente conserva el cuerpo actual, para que un cambio de frontmatter no borre el texto por accidente.
  - Escribir compara el archivo con lo leído en `Prepare`, bajo el candado `.git/beacon/records.lock`. Si alguien lo cambió en medio, falla con `record_changed`.
  - El payload de `config` llega en el paso 26, junto con su clave `approved` (D13).

### Paso 25. Archivo y `close` (hecho)

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
- **Cómo quedó:**
  - Los registros archivados entran al grafo como nodos cerrados: resuelven referencias y jerarquía, pero no bloquean ni dan señales. `Load` lee `archive/` solo si el grafo reporta una referencia o una tarea padre que falta; `Prepare` lo lee siempre, para que un id nuevo nunca repita uno archivado.
  - `close` va en dos fases, como `record`: `PrepareClose` devuelve el antes, el después (la capability sin la promesa, si se cierra una promesa) y la línea que se archivará; `Close` la escribe. El sello de `--approve` solo puede ir sobre el elemento que se cierra, y queda en su línea archivada.
  - Archivar y borrar del registro ocurren bajo el mismo candado que `record` (`.git/beacon/records.lock`), después de comprobar que el archivo no cambió desde `PrepareClose`.
  - `get` devuelve `archived: true` cuando el registro está archivado, y `find` recibe si debe buscar también en el archivo. El payload de `record` acepta la clave `archived` para que el ciclo `get` → `record` siga cerrado.
  - Errores nuevos: `guide_not_archivable` y `task_has_open_children`, con las subtareas abiertas en `details.subtasks`.
  - **Límite conocido:** se archiva primero y se borra después. Si el proceso muere en medio, el registro queda activo y archivado a la vez, y `doctor` lo reporta como id repetido (`graph_node_repeated`). Se arregla borrando el archivo de `records/`. Si pasa en la práctica, se agrega a `doctor` una reparación automática.

### Paso 26. Resultados, huella y `config.yaml`

- **Objetivo:** que los resultados de los checks existan de verdad y que la deriva no mienta con cambios sin commitear (D9, D12).
- **Construye:**
  - `driver/states/results.go`: el `ProofReader` real y su escritura.
  - `ProofResult` con `Footprint` y `Reason`. `VerifyPromise` trata `timeout` como un fallo.
  - En `driver/commit`: la huella (los archivos de unos `paths` cambiados respecto de un commit, con el SHA-256 de su contenido) y los archivos de `baseline` en `StateChanges`.
  - En `driver/record/config.go`: `runners` parseados (`run` y `baseline`) y la clave `approved` (D13).
  - En `action/drift.go`: hay deriva si la huella actual difiere de la guardada.
- **`RecordResults(ctx, verdicts)`** en `spider.go`: recibe `{promesa, outcome, reason}`, calcula `HEAD` y la huella **una vez por capability**, y fusiona en `results.json` bajo candado.
- **Cuidado:** la huella se calcula sobre los archivos del diff que ya se obtuvo en lote (`layout`), sin una consulta a git por promesa. Un archivo borrado entra en la huella como borrado.
- **Hecho cuando:** (a) un check con cambios sin commitear no muestra deriva; (b) editar después uno de esos archivos sí la muestra; (c) commitear exactamente lo comprobado no la muestra; (d) dos escritores concurrentes no pierden resultados.

## 5. Bloque 5: dominio de Weaver

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

## 6. Bloque 6: comandos

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

## 7. Bloque 7: cierre

### Paso 36. Ramas largas

- **Objetivo:** probar con casos reales la hipótesis de `method`.
- **Casos (e2e en `branch_test.go`):**
  1. Una rama cambia una promesa mientras `main` cambia otra: `diff` en la rama muestra solo la suya.
  2. `main` debilita una promesa con `approved` y la rama se rebasa: el cambio de `main` no aparece como pendiente en la rama.
  3. Las dos ramas cambian la misma promesa: git da conflicto en ese archivo y `doctor` lo reporta.
  4. Una rama vieja reclama una tarea que `main` ya cerró: `claim` falla con `record_missing` y el `hint` sugiere `beacon get <id>`, que la muestra archivada.
- **Hecho cuando:** los cuatro casos pasan, o su comportamiento queda escrito en `method` y ese archivo deja de decir "no está resuelto".

### Paso 37. Medición y KB

- **Medir:** con 500 capabilities, cuánto tardan `record` (incluida la validación del grafo completo), `status` y el costo propio de `check` sin contar el ejecutor. Se anota en `layout`, junto a las mediciones anteriores.
- **Sincronizar:** `status` marca la Fase 2 como cerrada, y la KB recoge lo que cambie al implementar.
- **Hecho cuando:** los números están en `layout` y la KB no contradice el código.

## 8. Hitos y commits

| Hito | Pasos | Rama | Resultado verificable |
|---|---|---|---|
| H6. Spider para Weaver | 23 a 26 | `feat/spider-writes` | `get` → `record` idempotente; archivo y huella probados |
| H7. Dominio de Weaver | 27 y 28 | `feat/weaver-cli` | Tabla de niveles e invocaciones en verde |
| H8. Lecturas y escritura | 29 a 32 | `feat/weaver-cli` | `get`, `find`, `context`, `ready`, `record`, `close` y `claim` en e2e |
| H9. Checks y estado | 33 a 35 | `feat/weaver-cli` | `check`, `diff`, `brief` y `status` en e2e |
| H10. Fase 2 cerrada | 36 y 37 | `feat/weaver-cli` | Ramas largas probadas y tiempos en `layout` |

## 9. Riesgos de la Fase 2

| Riesgo | Qué lo mitiga |
|---|---|
| En Windows, matar un ejecutor no mata a sus hijos (un `pnpm` que lanza `node`) | `cmd.Cancel` más `WaitDelay`, y un test con un ejecutor que lanza un hijo. Si no alcanza, un job object de Windows en `driver/runner`, sin tocar el dominio |
| Validar el grafo completo en cada `record` es lento en repos grandes | Se mide en el paso 37; la carga ya ronda los 70 ms con 500 capabilities |
| El payload completo de una capability grande es largo para el agente | `get --json` da la base exacta; si molesta, se reconsidera el parche (D7) con datos de fricción |
| El agente escribe un `approved` que la persona no dio | El `approved` del payload se ignora (D16) y el sello solo sale de `--approve`. Sigue siendo un marcado (`deltas`) |
| La huella cuesta con muchos archivos sin commitear | Se calcula solo sobre los archivos del diff en lote que caen en los `paths` (paso 26) |
| Estrechar `paths` oculta deriva sin pedir sí | Riesgo aceptado en D17 y anotado en `status` |
| El tope fijo de 10 minutos se queda corto en suites grandes | D11: se vuelve configurable si la fricción lo pide |
