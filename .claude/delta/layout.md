# Almacenamiento

## Requisitos

- El agente **no edita Markdown** como fuente de verdad: edita registros tipados con operaciones validadas.
- El código puede **consultar** el estado: qué promesas aplican a un archivo, qué preguntas están abiertas, qué está listo, qué está desactualizado.
- El estado **viaja con el código en git**, en el mismo commit, y se comparte en equipo sin servidores.
- Una persona puede **leerlo** sin herramientas especiales.
- El repo no se llena de miles de archivos ni de ruido en los PRs, y un humano puede navegarlo.

## Layout

```text
.beacon/
  records/
    goals/         bc-0001-pagos-sin-friccion.yaml
    capabilities/  bc-7f3a-auth-login.yaml
    decisions/     bc-a1b2-sesiones-con-jwt-corto.yaml
    questions/     bc-e5f6-limite-de-intentos.yaml
    guides/        bc-0a9b-arquitectura-general.md
    works/         bc-91cd-migrar-a-postgres.yaml
    tasks/         bc-91cd.1-crear-esquema.yaml
  archive/
    2026-11.jsonl
  config.yaml
```

- **Carpetas en plural**, todas con el mismo patrón. Archivos `<id>-<slug>.<ext>`: el slug hace legible el directorio y el id lo hace único.
- **`records/` y `archive/` separados.** Todo lo activo cuelga de `records/`.
- **Una capability con sus promesas por archivo.** Las promesas son pequeñas y se editan y leen juntas. Decisiones y preguntas van sueltas, una por archivo, porque se leen y se cierran de forma independiente. Los registros activos son decenas o cientos de archivos, no miles.
- **Todo lo que hay en `.beacon/` se versiona.** Lo que no se comparte vive fuera, en `.git/beacon/` (ver abajo).
- La forma de cada registro, el esquema y `config.yaml` están en `ledger`; la de las guides, en `guides`.

## Estado local

Lo que no se comparte vive en `.git/beacon/`, dentro del directorio común de git (`git rev-parse --git-common-dir`), y no en `.beacon/`:

- `results.json`: resultados de los checks por promesa y commit (`pledge`).
- `claims.json`: los claims de tareas de este clon, visibles a todos sus worktrees (ver abajo).
- Las métricas de fricción (`engine`).
- Lo que los hooks ya inyectaron en cada sesión, para no repetirlo (`anchor`).
- `backups/`: copias de los archivos que `beacon install` modificó (`anchor`).

Git no puede commitearlo por accidente, no hay nada que añadir al `.gitignore` y vive y muere con el clon. Se puede borrar sin consecuencias: los resultados se recalculan y solo se pierden las métricas locales, los claims locales y las copias.

### Varios worktrees a la vez

Los worktrees de un mismo repo comparten `.git/beacon/`, y varias sesiones de agente pueden escribir al mismo tiempo. Por eso:

- **Toda escritura en `.git/beacon/` es atómica:** se escribe un archivo temporal en el mismo directorio y se renombra sobre el definitivo, para que nadie lea un archivo a medias.
- **Los resultados van por commit y promesa** (`pledge`), así que compartirlos entre worktrees es seguro: dos worktrees en el mismo commit producen el mismo resultado.
- **Los claims se guardan también aquí** (`claims.json`), porque un claim escrito solo en un registro versionado es invisible para un worktree vecino hasta que se commitee y se fusione. `beacon claim` escribe en los dos sitios: el claim local avisa a las sesiones de esta máquina al instante, y `owner` y `branch` en el registro avisan al equipo cuando la rama sube (`method`).
- El formato exacto de estos archivos (por ejemplo un archivo por commit) se decide al construir; el requisito es la escritura atómica.
- `beacon doctor` limpia claims de ramas o worktrees que ya no existen.

## Archivo de lo cerrado

- Lo cerrado pasa a `archive/` en **segmentos mensuales compactos** (`2026-11.jsonl`): preguntas respondidas, decisiones reemplazadas, metas logradas o descartadas, works y tasks terminados, capabilities descartadas.
- Los ids siguen siendo únicos y las relaciones hacia un registro archivado se resuelven (el binario lee `archive/` solo cuando una relación o consulta lo pide).
- **Las guides no se archivan:** si una queda obsoleta se borra, y git guarda la historia.

## Formato: YAML estricto

- Los registros estructurados son **YAML**; las guides son Markdown con frontmatter YAML.
- **Librería:** `go.yaml.in/yaml/v4` (v4.0.0), la línea mantenida por la organización YAML; `gopkg.in/yaml.v3` está archivada. Se trabaja sobre nodos para conservar estilos, líneas y campos desconocidos.
- **Estricto:** sin anclas ni alias, sin claves de combinación (`<<`) y sin claves duplicadas. La librería no rechaza estas últimas al leer nodos, así que las valida el binario.
- **Sin tipos implícitos:** el parser lee cada valor por el tipo que dice el esquema (`ledger`). `beacon fmt` entrecomilla con comillas dobles **solo lo que otra herramienta podría reinterpretar**: `yes`, `no`, `on`, `off`, `true`, `false`, `null` y `~` en cualquier capitalización, números (`1.0`, `1e3`, `0x1f`, `1234567`), fechas dentro de un campo de texto, strings vacíos, valores con `:` o `#`, con espacios al inicio o al final, valores que empiezan con un indicador YAML (`*`, `&`, `!`, `%`, `@`, `` ` ``, `?`, `- `, `|`, `>`, `[`, `{`, `'`, `"`), como el glob `"**/*.go"`, o con caracteres de flujo (`,`, `{`, `}`, `[`, `]`) dentro de un valor en línea. Los campos de fecha (`2026-11-22`) y las enumeraciones (`status: active`) van sin comillas.
- **`beacon fmt`** formatea de forma canónica: orden fijo de claves, un campo por línea, listas de ids y promesas ordenadas por id. Un formato canónico da diffs pequeños y estables.
- **En línea o en bloque:** `check`, `approved`, `attested` y las listas de ids, globs y etiquetas van en línea. Las listas de texto (`nongoals`, `rejected`, `log`) y `promises` van en bloque, y conservan su orden salvo `promises`. `request` usa un bloque literal cuando tiene saltos de línea, sin alterar su contenido.
- **Orden de claves:** el de la tabla. Los campos desconocidos van después de los conocidos, en orden alfabético y sin cambios por dentro.

| Tipo | Orden |
|---|---|
| goal | `id`, `created`, `author`, `statement`, `success`, `nongoals`, `status`, `approved` |
| capability | `id`, `created`, `author`, `name`, `intent`, `paths`, `serves`, `nongoals`, `approved`, `promises` |
| promesa | `id`, `text`, `check`, `status`, `approved`, `attested`, `origin`, `refers` |
| decision | `id`, `created`, `author`, `what`, `why`, `where`, `supersedes`, `rejected`, `status`, `review`, `reviewed`, `approved` |
| question | `id`, `created`, `author`, `text`, `blocks`, `status`, `answer`, `origin`, `approved` |
| work | `id`, `created`, `author`, `request`, `implements`, `serves`, `status`, `log`, `approved` |
| task | `id`, `created`, `author`, `title`, `parent`, `blocks`, `implements`, `status`, `owner`, `branch`, `approved` |
| guide | `title`, `kind`, `applies`, `always`, `review`, `reviewed`, `tags`, `audience`, `status`, `refers` |
| `check` | `runner` y sus parámetros en orden alfabético, o `manual` |
| `approved` | `by`, `at`, `via` |
| `attested` | `by`, `at`, `commit`, `evidence` |

- El binario valida contra un esquema y, si algo no encaja, dice cómo arreglarlo.

## Ids y nombres

- **Ids aleatorios** (`bc-a1b2`): 4 caracteres hexadecimales de `crypto/rand`, estables una vez creados. Si el id ya existe se reintenta, y crecen en longitud si el espacio se llena. No se calculan del contenido, para que dos ramas con el mismo título no choquen.
- Hijos jerárquicos: `bc-91cd.1`, `bc-91cd.1.1`.
- El slug lo genera el binario a partir del título.
- El prefijo es `bc-` y la carpeta es `.beacon/`.

## Merges y conflictos

- Un conflicto ocurre solo si dos personas tocan **el mismo archivo**; `beacon doctor` ayuda a resolverlo.
- Las listas internas (promesas dentro de una capability) van **ordenadas por id**. Como los ids son aleatorios, dos personas que añaden promesas distintas casi nunca insertan en el mismo punto y git las fusiona sin conflicto.
- Los borrados son borrados de archivo: git los entiende.
- Hay conflictos que git no ve: dos cambios que fusionan limpio pero rompen el proyecto. Esos solo los detecta un CI que corra sobre el **resultado de la fusión** (`sentry`).

## Ruido en PRs y navegación humana

| Problema | Solución |
|---|---|
| Ruido en los PRs | Formato canónico (diffs pequeños y estables); `.gitattributes` con `linguist-generated` para que GitHub colapse `.beacon/**`; comentario de CI con `beacon diff` clasificado por nivel, que es lo que lee el revisor (`sentry`); menos archivos |
| Un humano no puede navegar el directorio | Navega la **vista renderizada** (`engine`); los nombres con slug hacen legible el directorio |

## Edición humana

Una persona puede editar el YAML a mano. El binario valida en el siguiente comando y en el pre-commit, y los cambios de nivel 3 se detectan igual comparando con la versión de git (`deltas`).

## Versionado del esquema

- `schema: 1` en `config.yaml` desde el primer día.
- El binario **rechaza** esquemas más nuevos que el suyo.
- Al reescribir un registro **conserva los campos desconocidos**.
- `beacon migrate` migra entre versiones de forma determinista e idempotente.

## Carga en memoria

Todo el estado está en los archivos. Al arrancar, el binario los lee y arma el grafo en memoria. Como el CLI es la única interfaz (`engine`), **cada llamada paga esa carga**.

Medición en Go con `gopkg.in/yaml.v3` v3.0.1 (2026-11-22). Cada archivo es una capability con 10 promesas; se decodifica a structs conservando campos desconocidos, se arma el índice de ids y de rutas y se resuelve una consulta de contexto. Máquina de 2 núcleos, caché de archivos caliente, mediana de 7 corridas:

| Archivos | Promesas | Tamaño | 1 hilo | 2 hilos |
|---|---|---|---|---|
| 50 | 500 | 0,1 MB | 7 ms | 4 ms |
| 500 | 5.000 | 1,1 MB | 99 ms | 79 ms |
| 1.000 | 10.000 | 2,2 MB | 190 ms | 111 ms |
| 5.000 | 50.000 | 11,2 MB | 839 ms | 542 ms |

Es de 2 a 3 veces más rápido que la medición anterior en Python con libyaml (20 ms, 210 ms y 2 s para 50, 500 y 5.000 archivos). Con decenas de capabilities el arranque no se nota; con cientos sigue por debajo de 100 ms y solo con miles de archivos se siente en cada llamada de CLI.

Esta tabla se midió con `gopkg.in/yaml.v3`, hoy archivada. En una prueba del 2026-11-23, solo de parseo a nodos, v3.0.6 y v4.0.0 tardaron prácticamente lo mismo (42 y 41 ms para 500 archivos). La carga completa con v4 se vuelve a medir al cerrar la Fase 1.

### Consultas a git

La deriva y la frescura (`spider`) dependen de git. Medición del 2026-11-22 en un repo sintético (5.000 archivos en 100 módulos, 2.000 commits, 2 núcleos, caché caliente, mediana de 5 corridas):

| Consulta | Tiempo |
|---|---|
| `git diff --name-only <base>..HEAD` (una por commit base distinto) | de 5 a 16 ms |
| `git log --since=<fecha> --name-only` | 10 ms (7 días), 27 ms (30 días), 60 ms (toda la historia, 2.000 commits) |
| `git status --porcelain` | 8 ms |
| Una consulta por capability (`git log -1 -- <path>`), 500 capabilities | 1.147 ms |

**Conclusión de diseño:** las consultas van **en lote**, no una por promesa. Una por commit base distinto más una de frescura, y el cruce con los `paths` se hace en memoria. Con 500 capabilities, el CLI completo (carga más git) queda en el orden de 150 a 200 ms.

**Límites de la medición:** archivos y repos sintéticos y uniformes, caché caliente, coincidencia de rutas por prefijo (no globs completos). Los repos reales tienen árboles y diffs más grandes, y `git status` puede tardar más.

Mitigaciones dentro del proceso:

1. **Parseo en paralelo**, un worker por núcleo: entre 1,3 y 1,7 veces más rápido con 2 núcleos.
2. **Consultas a git en lote** (ver arriba).
3. **Lectura perezosa:** `context <paths>` no necesita los cuerpos de las guides ni `archive/`, solo el índice de rutas.

Como los hooks de reinyección (`anchor`) llaman al CLI con más frecuencia, esa latencia por llamada importa más: se mide en el benchmark de fricción (`status`).