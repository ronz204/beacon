# Registros

Modelo de datos. Cómo se guardan (carpetas, ids, formato) está en `layout`.

## Tipos

| Tipo | Qué guarda | ¿Se verifica? |
|---|---|---|
| **goal** | La meta actual del producto, su criterio de éxito y lo que no se hace | Por la persona |
| **capability** | Una parte del sistema: nombre, intención, `paths` y sus promesas dentro | A través de sus promesas |
| **promise** (dentro de una capability) | Algo que la capability garantiza, con su forma de verificarse (`pledge`) | **Sí** |
| **decision** | Qué se decidió, por qué, dónde aplica y qué alternativas se descartaron; puede reemplazar a otra | No; se revisa por antigüedad |
| **question** | Algo que todavía no se sabe; puede bloquear trabajo | No |
| **guide** | Contexto orientativo en prosa (`guides`) | No; se revisa por antigüedad |
| **work** | Un trabajo grande en curso: petición textual, log, promesas que implementa | Por sus promesas |
| **task** | Un nodo del grafo de un work, con dependencias y claim | Se cierra al pasar el check de su promesa |

No hay un tipo especial de capability para diseño de UI: una capability de UI es una capability normal con un ejecutor visual y promesas manuales (`pledge`).

## Campos comunes

**Todos los campos son de una sola palabra**, en inglés (`design`). Todos los registros YAML llevan:

| Campo | Qué es |
|---|---|
| `id` | Hash (`bc-a1b2`). En las guides sale del nombre del archivo |
| `created` | Fecha `YYYY-MM-DD`. Es fija: nunca cambia, así que no mete ruido en los diffs. Se guarda porque sacarla de git es lento y falla con renombres y clones superficiales |
| `author` | El actor que lo escribió (ver abajo) |
| `closed` y `reason` | Fecha y motivo al cerrar; solo aparecen en `archive/` (`layout`) |
| `approved` | `{by, at, via}` cuando un cambio de nivel 3 recibió un sí (`deltas`). `via` solo aparece si un agente transmitió el sí de la persona |

- Las promesas dentro de una capability solo llevan `id`, `approved`, `closed` y `reason`; el resto lo hereda de la capability y git guarda el detalle.
- **No se guarda `updated`**: sale de git, y guardarlo ensuciaría cada diff.
- Las decisiones y las guides aceptan además `review` y `reviewed` (`guides`).

### Actor

`author`, `approved.by` y `via` salen del **actor** que ejecuta el comando:

- Si existe la variable de entorno `BEACON_ACTOR` (por ejemplo `agent:claude-code`), ese es el actor. La pone Anchor en los comandos del agente (`anchor`).
- Si no, el actor es la persona: `human:<user.name de git>`.

Es un **marcado, no una prueba**: el agente corre con el mismo usuario del sistema y podría saltárselo. La garantía real en equipo es la revisión del PR por una persona (`deltas`, `sentry`).

## Relaciones

Las relaciones son **campos tipados** en el registro que corresponde, no una lista genérica. Así el YAML es más corto, el binario valida qué relación es válida en cada tipo y las listas se ordenan por id (`layout`).

| Campo | Qué conecta | ¿Bloquea? | Va en |
|---|---|---|---|
| `implements` | work o task → promesa | No | work, task |
| `blocks` | question → work o task; task → task | **Sí** | question, task |
| `parent` | task → work | Sí (la jerarquía) | task |
| `supersedes` | decision → decision | No | decision |
| `serves` | capability o work → goal | No | capability, work |
| `origin` | question o promesa → work donde se descubrió | No | question, promesa |
| `refers` | promesa → promesa de otra capability; guide → capability | No | promesa, guide |

**Un hecho vive en un solo registro y los demás lo refieren, no lo copian.** Así no hay contradicciones entre registros.

## Paths

`paths` son **una pista de alcance, no una propiedad del código**: no hace falta saber a qué capability pertenece cada línea. Lo mismo vale para `where` (decisiones) y `applies` (guides).

- **Sintaxis de globs estilo doublestar:** `*` (dentro de un segmento), `**` (varios segmentos), `?` y `{a,b}`. Siempre desde la raíz del repo, con `/` como separador y sensibles a mayúsculas.
- **Sin negaciones** (`!`): para excluir algo se acota el glob. Así es fácil razonar qué cubre qué cuando varias capabilities se solapan.
- Tan gruesos como se quiera (`apps/api/src/auth/**`).
- Pueden solaparse, y puede haber código sin ningún dueño (recibe solo la meta y las guides `always`).
- Son **opcionales**: una promesa sin `paths` funciona pero no puede avisar de deriva, y eso se muestra como una señal, no como error.
- Sirven para `beacon context <archivos>`, la deriva y la frescura de las guides (`spider`).
- En un monorepo cada capability apunta a su paquete; no hace falta un `.beacon` por paquete.

## Esquema 1

Además de los campos comunes, cada tipo tiene estos. Los `status` son enumeraciones cerradas.

| Tipo | Obligatorios | Opcionales |
|---|---|---|
| goal | `statement`, `success`, `status`: active, achieved o dropped | `nongoals` |
| capability | `name`, `intent`, `promises` | `paths`, `serves`, `nongoals` |
| promesa | `id`, `text`, `check`, `status`: proposed o active | `approved`, `attested` (solo las manuales), `origin`, `refers` |
| decision | `what`, `why`, `status`: active o superseded | `where`, `supersedes`, `rejected`, `review`, `reviewed` |
| question | `text`, `status`: open o answered | `blocks`, `answer`, `origin` |
| work | `request` (textual), `status`: active, done o dropped | `implements`, `serves`, `log` |
| task | `title`, `parent`, `status`: open, claimed o done | `blocks`, `implements`, `owner`, `branch` |
| guide | `title`, `kind` (`guides`) | `applies`, `always`, `review`, `reviewed`, `tags`, `audience`, `status` |

- `check` es `{runner, ...parámetros}` o `{manual: <pasos>}` (`pledge`).
- `attested` es `{by, at, commit}` y, si la persona lo indica, `evidence`: una ruta o un enlace a lo que se revisó (una captura, un video). Beacon no guarda el archivo, solo la referencia (`pledge`).
- `nongoals` es una lista de textos: lo que esa meta o capability **no** hace, para que el agente no lo construya sin que nadie lo pida.
- `rejected` es una lista de textos: las alternativas descartadas y por qué, para que el agente no vuelva a proponerlas (`spider`).

Los campos desconocidos se conservan al reescribir y las versiones nuevas del esquema se migran con `beacon migrate` (`layout`).

## config.yaml

```yaml
schema: 1
runners:
  test-node:
    run: ["pnpm", "--filter", "{pkg}", "test", "--", "{target}"]
  test-go:
    run: ["go", "test", "{pkg}", "-run", "{target}"]
  visual:
    run: ["pnpm", "exec", "playwright", "test", "--grep", "{target}"]
    baseline: ["tests/**/*-snapshots/**"]
review:
  convention: 6mo
  architecture: 12mo
  product: 12mo
  domain: 12mo
  runbook: 6mo
  decision: 6mo
limits:
  guide: 150
  brief: 60
  hook: 20
agents: [claude-code, codex, cursor]
```

| Clave | Qué es | Cambiarla es |
|---|---|---|
| `schema` | Versión del esquema (`layout`) | Nivel 3 |
| `runners` | Los **ejecutores**: pocos y revisados. Los registros solo pasan parámetros (`pledge`). Un ejecutor visual declara en `baseline` los globs de sus capturas base | Nivel 3 |
| `review` | Plazo de revisión por tipo de guide y de decisión (`guides`) | Nivel 1 |
| `limits` | Largo máximo en líneas de una guide, del `brief` y de lo que un hook inyecta a mitad de sesión (`spider`, `anchor`); valores iniciales que se ajustan con el uso | Nivel 1 |
| `agents` | Agentes instalados (`anchor`) | Nivel 1 |

Solo `runners` y `schema` son sensibles porque solo los ejecutores corren código en CI (`deltas`, `sentry`). Cambiar los archivos que cubre un `baseline` también es de nivel 3, pero se detecta en el diff del repo y no en `config.yaml` (`deltas`).

## Ejemplos

Los ejemplos siguen la regla de comillas de `beacon fmt` (`layout`): los valores con `:` van entrecomillados.

**Meta**

```yaml
# .beacon/records/goals/bc-0001-pagos-sin-friccion.yaml
id: bc-0001
created: 2026-11-22
author: "human:ron"
statement: Que un usuario nuevo complete su primer pago en menos de 2 minutos.
success: Tasa de pagos completados al primer intento mayor a 90 %.
nongoals:
  - No soportamos pagos recurrentes en esta versión.
status: active          # active | achieved | dropped
```

**Capability con promesas**

```yaml
# .beacon/records/capabilities/bc-7f3a-auth-login.yaml
id: bc-7f3a
created: 2026-11-22
author: "agent:claude-code"
name: Inicio de sesión
intent: Que un usuario entre con correo y contraseña sin exponer si la cuenta existe.
paths: [src/auth/**, tests/auth/**]
serves: bc-0001
nongoals:
  - No incluye inicio de sesión con redes sociales.
promises:
  - id: bc-c3d4
    text: Tras 5 intentos fallidos la cuenta se bloquea 15 minutos.
    check: {runner: test-node, pkg: api, target: auth/lockout}
    status: active        # proposed | active
    approved: {by: "human:ron", at: 2026-11-22, via: "agent:claude-code"}
  - id: bc-c4a1
    text: La pantalla de inicio de sesión se ve bien en móvil.
    check: {manual: Abrir /login en un viewport de 375 px y confirmar que no hay desbordes}
    status: active
    approved: {by: "human:ron", at: 2026-11-22}
    attested: {by: "human:ron", at: 2026-11-22, commit: 9f3c2ab, evidence: docs/evidence/login-375.png}
```

**Decisión**

```yaml
# .beacon/records/decisions/bc-a1b2-sesiones-con-jwt-corto.yaml
id: bc-a1b2
created: 2026-11-22
author: "agent:claude-code"
what: Sesiones con JWT de 15 min más refresh token.
why: Evitamos estado de sesión en servidor y limitamos el daño si se filtra un token.
where: [src/auth/**]
rejected:
  - "Sesiones en servidor: exigen estado compartido entre instancias."
  - "JWT de larga duración: un token filtrado sirve durante días."
status: active          # active | superseded
```

**Pregunta**

```yaml
# .beacon/records/questions/bc-e5f6-limite-de-intentos.yaml
id: bc-e5f6
created: 2026-11-22
author: "agent:claude-code"
text: ¿Cuántos intentos fallidos antes de bloquear, y por cuánto tiempo?
blocks: [bc-91cd]
status: open            # open | answered
```

**Work y task**

```yaml
# .beacon/records/works/bc-91cd-migrar-a-postgres.yaml
id: bc-91cd
created: 2026-11-22
author: "agent:claude-code"
request: |
  (la petición original del usuario, textual)
implements: [bc-c3d4]
status: active
log:
  - "2026-11-22: se separó el esquema en dos pasos"
```

```yaml
# .beacon/records/tasks/bc-91cd.1-crear-esquema.yaml
id: bc-91cd.1
created: 2026-11-22
author: "agent:claude-code"
title: Crear el esquema nuevo
parent: bc-91cd
implements: bc-c3d4
status: open            # open | claimed | done  (con owner y branch al reclamarla)
```

Los campos opcionales ausentes se omiten. La petición del `work` se guarda textual para que cualquier sesión pueda reanudar desde ella (`method`).