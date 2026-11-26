# Plan de construcción: Fase 2 (Weaver)

Este plan cubre solo la fase en curso: en qué orden se construye y qué archivos salen. El detalle funcional vive en la KB (`.claude/delta/`), y todas las fases, en `status`.

La numeración de pasos y decisiones sigue la del plan de la Fase 1 (Spider), ya cerrada, para que las referencias de la KB y de los commits sigan valiendo: por eso los pasos empiezan en el 23 y las decisiones en D5. Ese plan está en el historial de git (`git show 967f5aa:PLANNING.md`).

> **Estado: en curso.** Bloques 4, 5 y 6 cerrados y paso 36 hecho; sigue el paso 37, con la rama `feat/weaver-cli`. Las decisiones D5 a D20 están cerradas con la persona y registradas en la KB.

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
  - Se archiva primero y se borra después. Si el proceso muere en medio, el registro queda activo y archivado a la vez; como la línea archivada prueba que se pidió el cierre, `doctor` lo termina (borra el registro o saca la promesa de su capability) y lo reporta como `close_completed`.

### Paso 26. Resultados, huella y `config.yaml` (hecho)

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
- **Cómo quedó:**
  - `RecordResults` recibe veredictos `{promesa, outcome, reason}`, hace una consulta a git (cambios desde `HEAD`) y calcula la huella una vez por capability. Una promesa desconocida da `record_missing`.
  - La huella es el SHA-256 de las líneas `ruta NUL sha256-del-contenido` (o `deleted`) de los archivos cambiados que caen en los `paths`, ordenadas por ruta. Vacía si no hay ninguno.
  - La deriva pasa a ser "la huella actual difiere de la guardada". Con una huella guardada vacía es idéntica a la regla anterior, y solo se calcula cuando algún archivo de los `paths` cambió.
  - Sin git, la fachada sigue usando un lector sin resultados.
  - En Windows, un antivirus o el indexador pueden retener un archivo un instante: soltar un candado y borrar un registro reintentan como ya lo hacía el renombrado, y un candado que no se pudo soltar se reporta en vez de ignorarse. Era la causa más probable de una falla intermitente que no se pudo reproducir en más de 40 corridas.
  - `config.yaml` parsea `runners` (`run` no vacío y `baseline` opcional) y `approved`; un ejecutor sin `run` da `record_runner_invalid`. `Runners()` los expone, y `Config()` devuelve el `config` con la forma del payload.
  - `StateChanges` usa los `baseline` reales e incluye los cambios en `config.yaml` (tipo `config`).
  - El payload `{"type": "config"}` se fusiona como un registro, con `approved` como única clave del binario, y el sello usa el id `config`. Si no cambia nada, el archivo queda igual byte por byte; si cambia, se escribe con el formato canónico de `config.yaml`, que también aplica `fmt` (`layout`).

## 5. Bloque 5: dominio de Weaver

**Rama:** `feat/weaver-cli`, desde `main` con el Bloque 4 fusionado.

### Paso 27. Pieza Weaver y contexto `deltas` (hecho)

- **Objetivo:** que el nivel de un cambio sea una función pura y probada.
- **Construye:** `weaver.go` (solo el constructor), `driver/facade`, `driver/double` y el contexto `deltas`.
- **`ClassifyChange(StateChange) (ChangeLevel, bool)`:** devuelve el nivel y si pide sí, según la tabla de D17.
- **`NormalizeText`:** pasa a minúsculas, quita la puntuación y colapsa los espacios (`deltas`). Un cambio que desaparece al normalizar es de nivel 1.
- **`StampApproval(actor, human, today)`:** si el actor es una persona, `by` es el actor y no hay `via`; si es un agente, `by` es `human:<user.name de git>` y `via` es el agente. Sin `user.name`, falla con `identity_missing`.
- **`IsFreshlyApproved`:** la regla del sí fresco (D19).
- **Cuidado:** el dominio recibe el `user.name` y la fecha como datos; no lee git ni el reloj (`code-domain-layer`).
- **Hecho cuando:** cada fila de D17 tiene un caso, y el ejemplo de `deltas` (de "5 intentos" a "10 intentos") da nivel 3.
- **Cómo quedó:**
  - Los datos entran por `ParseStateChange`, que valida con los parsers del dominio. `facade` decodifica el JSON de Spider a campos planos (cada valor como texto canónico) y a promesas con su `check`, su ejecutor y su `approved`.
  - `ClassifyElements` devuelve un hallazgo por elemento afectado (promesa, registro, `config` o captura base) con su nivel y si pide sí; `ClassifyChange` es su máximo. Los pasos 30 y 34 usan los hallazgos para sellar `approved` y para preguntar `IsFreshlyApproved` por elemento.
  - Casos que D17 no nombra: una promesa nueva ya `active` es nivel 2 y pide sí, como pasar de `proposed` a `active`; cerrar una promesa `proposed` y editarla sin activarla son nivel 1; borrar un goal pide sí.
  - `NormalizeText` cambia la puntuación por un espacio y no quita símbolos: `1.5` y `15` siguen siendo distintos, porque la regla nunca debe dejar pasar un debilitamiento. El atajo vale para el texto de promesas y para `statement` y `success`; un `check` se compara exacto.
  - `StateChange` de Spider trae ahora `runners` (los ejecutores cuyo `baseline` cubre una captura base) y `closed` (id y `approved` de cada línea nueva de `archive/`). Sin lo primero, `IsFreshlyApproved` no sabía qué `approved` autoriza una captura; sin lo segundo, un cierre aprobado con `close --approve` aparecía como pendiente.
  - `weaver.New(project)` solo arma la fachada; `driver/double` trae `Changes` y `Executor` en memoria.

### Paso 28. Contexto `checks` (hecho)

- **Objetivo:** que armar una invocación sea seguro antes de ejecutar nada.
- **Construye:** el contexto `checks` y el contrato `Executor`, con `Run(ctx, argv, dir) (ExecutionResult, error)`.
- **`BuildInvocation(spec, params)`:** sustituye `{param}` en cada elemento de `run`, sin shell.
  - Falla si falta un parámetro que la plantilla usa (`param_missing`), o si sobra uno que no usa (`param_unused`), para atrapar typos.
  - Un parámetro que empieza con `-` se rechaza de nuevo aquí, aunque el registro ya lo validó (`pledge`).
- **`GroupInvocations`:** agrupa las promesas por invocación idéntica (D20).
- **`CheckOutcome`:** `passed` o `failed`, con razón `exit_code` o `timeout` (D11).
- **Hecho cuando:** no se puede construir ningún argv con un parámetro inyectado como opción, y dos promesas con el mismo check comparten invocación.
- **Cómo quedó:**
  - `ParseRunnerSpec` rechaza un `run` vacío, con elementos en blanco o con un `{param}` en el programa (primer elemento), para que un dato no elija qué se ejecuta.
  - `BuildInvocation` y `ParsePromiseCheck` comparten la regla del valor plano: no vacío, sin control y sin `-` inicial (también tras recortar espacios). Los fallos son `param_missing`, `param_unused` y `param_invalid`. Solo `{nombre}` válido es un parámetro; otras llaves pasan tal cual.
  - `GroupInvocations` ordena por id de promesa y agrupa por ejecutor más parámetros; el grupo sale en el orden de su primera promesa.
  - `OutcomeOf(ExecutionResult)` da `passed`, o `failed` con `timeout` o `exit_code`. Un programa que no arranca es `runner_unavailable`, error del contrato `Executor` para el paso 33.

## 6. Bloque 6: comandos

Cada comando tiene un test e2e que corre el binario real sobre un repo temporal (`testing/utils`).

### Paso 29. Salida, señales y lecturas (hecho)

- **Objetivo:** completar el sobre `--json` y poner en marcha las lecturas puras.
- **Construye:** el código 2 (`ExitBlocked`) en `output`, `signals.go` y los handlers de `get`, `find`, `context` (con `--hook`) y `ready`.
- **Señal en JSON:** `{"kind": "red", "record": "bc-c3d4", "related": null, "relation": null}`, sacada del `Signal` de Spider. En texto, una línea por señal.
- **`ready`** toma la rama actual de git. Con HEAD separado, usa el commit y lo avisa.
- **`context --hook`** necesita un id de sesión, que recibe con `--session <id>`. De dónde lo saca cada agente se decide en la Fase 3 (abierto en `status`).
- **Hecho cuando:** los cuatro comandos devuelven el sobre con `signals` llenas, y un archivo sin dueño recibe en `context` solo la meta y las guides `always`.
- **Cómo quedó:**
  - `output.Printer` tiene `Done(Result)` (datos, señales, `hint` y código de salida) y `Failure` acepta un código de salida; `ExitBlocked` es 2. En texto, una línea `signal: <kind> <registro> [<relación> <relacionado>]` por señal.
  - `get`, `find` y `ready` van directo a Spider. `context` convierte las rutas a la raíz del repo (`Project.RepoPath`) y con `--hook --session <id>` imprime solo lo nuevo, sin salida si no hay nada. Sin `--session` falla con `session_required`.
  - `ready` toma la rama con `Project.Branch` (`git symbolic-ref`); con HEAD separado usa el commit y avisa en una línea y en el `hint`.
  - Solo `context` entrega `signals` en el sobre: `get`, `find` y `ready` no tienen señales que dar, y sus `signals` van vacías.
  - Un error de Spider se traduce a `code`, `message`, `details` y un `hint` por código (`handle/failure.go`); un código sin `hint` propio manda a `beacon doctor`.

### Paso 30. `record` (hecho)

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
- **Cómo quedó:**
  - `action.Recorder` hace el flujo: identidad, `Prepare`, `ClassifyElements`, y si algún hallazgo pide sí y no hay `--approve`, `approval_required` (código 2, con los cambios en `details.changes`) sin escribir. Con `--approve`, sella cada elemento que pide sí (la promesa, el registro o `config`) y llama a `Save`. Un registro sin cambios no se escribe.
  - `driver/identity` resuelve el actor: `BEACON_ACTOR` si existe, y si no `human:<user.name>`; sin nombre, `identity_missing`. Un agente necesita además el nombre solo al sellar.
  - `ClassifyElements` solo informa lo que cambió: una promesa intacta o una capability sin cambios propios no salen en `changes`.
  - La respuesta es `{type, id, file, saved, unchanged, changes: [{target, id, level, needsYes, reason, before, after}], ignored}`. Cada cambio que pide sí trae su motivo (`promise_text_changed`, `goal_changed`, `runners_changed`...) y el texto antes y después, para que el agente pueda mostrarle a la persona qué aprueba; en texto salen como líneas `-` y `+`. La guía de redacción de promesas sigue en la Fase 3.
  - Un commit que ya no existe (por ejemplo una atestación tras un rebase) ya no es un fallo del estado: la promesa se marca con deriva y el comando sigue (`ledger.ErrCommitUnknown`). Un fallo real de git sí se reporta.
  - `errorFailure` mapea cada código de Spider y de Weaver a su `hint`; los códigos de la tabla de arriba coinciden con los del plan.

### Paso 31. `close` (hecho)

- **Construye:** el handler y el `action/close.go` de Weaver, que clasifica antes de archivar.
- **Hecho cuando:** cerrar una promesa `active` pide sí, cerrar una pregunta no lo pide, y cerrar una guide da `guide_not_archivable` con su `hint`.
- **Cómo quedó:**
  - `action.Closer` sigue el mismo flujo que `record`: `PrepareClose`, `ClassifyElements` sobre el antes y el después, `approval_required` (código 2, con motivo y texto) y, con `--approve`, un solo sello sobre el elemento que se cierra. Cerrar una capability con promesas `active` estampa la capability, y ese sello cubre sus promesas en `IsFreshlyApproved` (`closed` trae ahora los ids de las promesas de la línea archivada).
  - `--reason` vacío falla en el router con `reason_required`. `guide_not_archivable` y `task_has_open_children` llegan con su `hint` (borrar el archivo; terminar o cerrar las subtareas de `details.subtasks`).
  - Los `hint` aceptan marcadores `{clave}` que se rellenan con el `details` del error (por ejemplo `{task}` y `{branch}`).

### Paso 32. `claim` (hecho)

- **Flujo:** leer el claim local → si otra rama o worktree tiene la tarea, salir con código 2 (`claim_taken`) y un `hint` que diga cuál → si no, escribir el claim local y después `owner` y `branch` en la tarea con `Save` (nivel 1).
- **`--force`:** la persona puede tomar una tarea reclamada (`method`: un claim no es un candado). El agente solo lo usa si la persona lo pidió, y la guía de Anchor lo dirá (Fase 3).
- **Cuidado:** si el registro no se puede escribir, se deshace el claim local, para que los dos sitios no queden desalineados.
- **Hecho cuando:** si dos worktrees reclaman la misma tarea, el segundo recibe código 2, y `ready` en el segundo ya no la ofrece.
- **Cómo quedó:**
  - Spider expone `Claims`, `SaveClaim`, `RemoveClaim` y `Worktree`. `action.Claimer` toma el actor, la rama y el worktree actuales, y rechaza con `claim_taken` (código 2) si el claim local es de otra rama u otro worktree, **o** si el registro ya dice `claimed` en otra rama (ese caso viene de un merge o de otra máquina). El `hint` nombra la rama que la tiene.
  - Escribe el claim local y después `status: claimed`, `owner` y `branch` en la tarea con el mismo `Prepare` y `Save` de `record` (nivel 1, sin sí). Si el registro no se puede escribir, devuelve el claim local a como estaba, incluido el titular anterior si era un `--force`.
  - `doctor` libera el claim local de un worktree o una rama borrados (probado con un claim real y `git worktree remove`); el claim del registro no se libera solo (`method`).
  - `--force` reemplaza al titular y lo informa en `replaced`. Reclamar de nuevo en la misma rama no cambia nada.
  - Errores nuevos, todos con código 1 salvo `claim_taken`: `claim_not_task`, `claim_task_done`, `claim_task_archived` (su `hint` manda a `beacon get <id>`) y `claim_branch_detached`.

### Paso 33. `check` (hecho)

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
- **Cómo quedó:**
  - Spider expone `CheckTargets` (cada capability con sus promesas y su `check`) y `TouchedCapabilities(revision)` (las que tocan los archivos cambiados, incluidos los no commiteados).
  - `action.Checker` elige las capabilities (sin argumentos por la base de D10, con un id, o `--all`), valida **todos** los ejecutores y parámetros antes de correr nada, agrupa las invocaciones (D20), las corre en serie y entrega los veredictos a `RecordResults`. Después hace un `Analyze` para las señales y el estado de cada promesa. Las promesas manuales se saltan y se listan en `manual`.
  - `driver/runner` usa `os/exec` sin shell, directorio en la raíz del repo y el entorno del proceso, y guarda solo los últimos 64 KB de salida. En texto, la salida pasa en vivo a stderr; con `--json`, `invocations[].output` trae las últimas 40 líneas de cada invocación que falló. Al agotarse los 10 minutos mata el árbol de procesos (en Windows, `taskkill /T`; en Unix, el grupo de procesos), y `WaitDelay` evita que un hijo huérfano retenga la salida.
  - `driver/branch` resuelve la base con `git merge-base HEAD <base>` (`--base` la fuerza; un valor que empieza con `-` se rechaza).
  - Un programa que no arranca no es una promesa en rojo: la invocación queda `unavailable`, no se guarda ningún veredicto (la promesa sigue `Unverified`) y el comando falla con `runner_unavailable` (código 1), cuyo `hint` nombra el programa. Si además hay rojos, gana el código 2 y la invocación solo se lista.
  - Errores: `base_missing`, `runner_missing` (antes de correr nada), `capability_missing`, `param_missing`/`param_unused`/`param_invalid` y `state_invalid`.
  - **El tope de 10 minutos se prueba en el ejecutor** (con un tope reducido, un proceso colgado y otro que lanza un hijo) y en la acción (un `timeout` se guarda como `failed`), no en el e2e: D11 dice que el tope no se configura, así que el binario no tiene una perilla para reducirlo.

### Paso 34. `diff` y pendientes de sí (hecho)

- **Construye:** `action/pending.go` y `action/diff.go`.
- **`diff`:** toma `StateChanges` desde la base (o desde `--since <rev>`), clasifica cada cambio con `ClassifyChange`, agrupa por nivel y marca cuáles piden sí y cuáles ya tienen el sí fresco (D19).
- **`pending`:** los cambios que piden sí y todavía no tienen el sí fresco. Lo usan `diff`, `brief` y `status`, y alimenta la señal "Pendiente de sí".
- **Hecho cuando:** una edición a mano del texto de una promesa `active` aparece como pendiente, y deja de aparecer después de `record --approve`.
- **Cómo quedó:**
  - `diff` toma `StateChanges` desde la base (o `--since <rev>`), clasifica cada cambio con `ClassifyElements` y marca cuáles piden sí y cuáles ya tienen el sí fresco. Sale agrupado por nivel, de mayor a menor, con `pending` aparte y una señal `pending_yes` por pendiente. Sale con código 0: informa.
  - `pending` (`PendingOf`) son los hallazgos que piden sí y no están aprobados. Una captura base se aprueba cuando cambió el `approved` de alguna promesa que usa su ejecutor; un cierre, cuando su línea archivada trae `approved`.
  - **Hueco que cerré:** una edición a mano ya está en disco, así que `record --approve` con el mismo payload veía "sin cambios" y no estampaba nada. Ahora, si el registro no cambia y hay `--approve`, `record` estampa los pendientes de ese archivo; para una captura base, las promesas de ese registro que usan el ejecutor. Sin `--approve` no hace nada.
  - **Sello repetido:** el sí fresco se reconoce porque el `approved` cambió. Si el mismo actor aprueba otro cambio del mismo elemento el mismo día, por el mismo agente, el sello saldría idéntico y el cambio seguiría pendiente. Por eso `approved` admite una clave `n` (`n: 2`, `n: 3`...) que el binario añade solo cuando el sello nuevo sería igual al guardado. Un `approved` normal no cambia de forma. Lo escribe `Draft.Stamp` y lo leen el mapper y `fmt`.
  - **Árbol de procesos en Unix:** el test del ejecutor se corrió también en Linux (WSL), compilado con `GOOS=linux`; el grupo de procesos (`Setpgid` y `kill -pgid`) mata al hijo que retiene la salida.

### Paso 35. `brief` y `status` (hecho)

- **`brief`:** el `brief` de Spider con los pendientes en su prioridad (paso 23). El aviso `N more, run beacon status` ya cumple P7.
- **`status`:** rojos, deriva, bloqueos y pendientes de sí, sin tope de líneas. Sale siempre con código 0: informa, no bloquea.
- **Hecho cuando:** un `brief` recortado nombra `beacon status`, y `status` muestra todo lo que el `brief` recortó.
- **Cómo quedó:**
  - Weaver arma una línea por pendiente de sí (`action.LineOf`: `promise bc-c3d4 promise_text_changed: <texto nuevo>`) y se la entrega a Spider, que la pone en su prioridad como `Needs approval: ...`. Sin base (`base_missing`) no hay pendientes y el comando sigue.
  - `status` es el mismo `brief` **sin tope** (`Project.Status`), así que muestra todo lo que el `brief` recortó por construcción: bloqueos, rojos, trabajo en curso, pendientes de sí, deriva, promesas sin comprobar en esta máquina y el índice de guides. En texto, una primera línea cuenta rojos, deriva, bloqueos y pendientes de sí. Sale siempre con código 0.
  - El resultado del `brief` de Spider trae ahora las señales con las que se armó, y el sobre las entrega junto con una señal `pending_yes` por pendiente. `data` lleva solo `lines` y `omitted`.
  - Un estado inválido falla con `state_invalid` (código 1) y el `hint` de siempre, igual que las demás lecturas.
  - **Sin base, se avisa:** si no hay rama base (`base_missing`) o no hay git, los pendientes no se pueden calcular. El `brief` y `status` siguen, pero el sobre lleva un `hint` ("no base branch was found, so pending yeses were not checked; run beacon diff --base <branch>") que en texto sale como una línea final. Con base, `hint` es `null`.
  - `status` suma a `data` un `counts` (`red`, `drift`, `blocked`, `pending_yes`, siempre las cuatro claves). Las secciones ya viajan en cada línea (`lines[].section`), así que no hace falta una estructura aparte.
  - **Costo:** con 500 capabilities, `brief` tarda 210 ms (la mitad es calcular los pendientes con git). Se quitó una consulta de más al resolver la base (`rev-parse` antes de `merge-base`). `context --hook` no usa git y sigue en unos 75 ms. Los números están en `layout`.

## 7. Bloque 7: cierre

### Paso 36. Ramas largas (hecho)

- **Objetivo:** probar con casos reales la hipótesis de `method`.
- **Casos (e2e en `branch_test.go`):**
  1. Una rama cambia una promesa mientras `main` cambia otra: `diff` en la rama muestra solo la suya.
  2. `main` debilita una promesa con `approved` y la rama se rebasa: el cambio de `main` no aparece como pendiente en la rama.
  3. Las dos ramas cambian la misma promesa: git da conflicto en ese archivo y `doctor` lo reporta.
  4. Una rama vieja reclama una tarea que `main` ya cerró: `claim` falla con `claim_task_archived` y el `hint` sugiere `beacon get <id>`, que la muestra archivada.
- **Hecho cuando:** los cuatro casos pasan, o su comportamiento queda escrito en `method` y ese archivo deja de decir "no está resuelto".
- **Cómo quedó:** los cuatro casos pasan en `branch_test.go` y `method` quedó actualizado.
  - **Caso 3 con un cambio:** `doctor` reportaba el conflicto como "YAML inválido" en una línea al azar. Ahora Spider reconoce los tres marcadores de git (`<<<<<<<`, `=======` y `>>>>>>>`, juntos y en ese orden) en un registro o una guide, falla con `record_conflict` en la línea del primero, y el `hint` de `doctor` dice cómo resolverlo. Un solo marcador suelto no cuenta.
  - **Caso 4 distinto del plan:** una rama vieja todavía ve la tarea abierta, así que puede reclamarla; el conflicto aparece al fusionar `main` (modificada contra borrada). El fallo `claim_task_archived` ocurre cuando la rama ya fusionó `main`. Los dos caminos tienen test.
  - **Límite anotado:** un debilitamiento commiteado sin sí en `main` no aparece en `diff` de una rama que lo hereda; lo cubre el CI (`snares`, Fase 4). Se puede auditar a mano con `beacon diff --since <commit de origen>`, y hay un test que lo prueba.
  - Una guide que documente un conflicto con los tres marcadores en el margen izquierdo se lee como conflicto; el mensaje de `record_conflict` ya dice que se sangren.
  - `doctor` distingue los registros con CRLF (`record_line_endings`) de los demás no canónicos, y su `hint` manda a `.gitattributes` y a `beacon fmt`.
  - El repo de prueba de e2e lleva ahora `.gitattributes` con `.beacon/** text eol=lf`, como lo hará `install` (Fase 3): sin eso, en Windows `git checkout` dejaba los registros con CRLF y `doctor` los marcaba como no canónicos.

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
