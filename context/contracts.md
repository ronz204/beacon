# Contratos vivos

Cómo se verifican las promesas. Quién puede cambiarlas está en `changes`.

## Qué es una promesa

Algo que una capability garantiza (`records`). Tiene un texto, un `check` que dice cómo verificarla y un estado de ciclo mínimo: `proposed` (entra así) o `active` (con un sí de la persona).

No hay otros estados de ciclo (`draft`, `approved`, `built`...). Lo demás (pendiente, en verde, en rojo, con deriva) se **calcula** del último check.

## Cómo redactar una promesa

Una promesa mal redactada no protege nada: el agente la cumple "en espíritu" y el check no dice nada útil. Es una **guía**, no una regla que el binario imponga (el código no puede juzgar si un texto es vago). Anchor la deja en la guía del método y en lo que `beacon record` responde al proponer una promesa (`anchor`).

- **Un resultado observable, no una intención.** "Tras 5 intentos fallidos la cuenta se bloquea 15 minutos", no "el inicio de sesión es seguro".
- **Entradas y salidas concretas:** números, estados, mensajes, roles.
- **Una sola cosa por promesa.** Si el texto lleva "y", probablemente son dos.
- **Casos borde y permisos van como promesas propias** si importan. Los agentes construyen bien el flujo principal y suelen omitir permisos, límites de tasa y errores si nadie los pide (`research`, 7.1).
- **Forma EARS opcional:** `WHEN <condición> THE SYSTEM SHALL <comportamiento>` ayuda a que el texto sea comprobable. Es una opción, no un formato obligatorio.
- Lo que la capability **no** hace va en `nongoals` (`records`), no como promesa negativa.

## Checks: ejecutores con parámetros

Un `check` no es un comando: son **datos** que apuntan a un ejecutor declarado en `config.yaml`.

```yaml
# config.yaml (código: pocos y revisados)
runners:
  test-node:
    run: ["pnpm", "--filter", "{pkg}", "test", "--", "{target}"]
```

```yaml
# promesa (datos: muchos y baratos)
check: {runner: test-node, pkg: api, target: auth/lockout}
```

- Los parámetros se pasan como **argumentos separados, sin shell**, y no pueden empezar con `-`, para que nadie inyecte una opción.
- Un monorepo de 40 paquetes usa el **mismo** ejecutor con otro `pkg`. Si los paquetes usan toolchains distintas, hay un ejecutor por ecosistema (`test-node`, `test-go`).
- Cambiar un ejecutor es el único cambio peligroso: es de nivel 3 y en CI se usan los `runners` de la rama base (`changes`, `sentry`).

## Cada test cita el id de su promesa

El test lleva el id de la promesa que prueba (por ejemplo en su nombre o en un comentario), y un lint avisa de tests que citan promesas inexistentes. Así la relación promesa-test es comprobable.

## Estados de verificación

Cuánta evidencia respalda una promesa, sin disfrazarla de más:

| Estado | Quién lo respalda | Se cae cuando |
|---|---|---|
| **Holds** | Un ejecutor pasó | El check falla |
| **Attested** | Una persona lo verificó en un commit | Cambia el código de sus `paths` (deriva) |
| **Unverified** | Nada | Es el estado por defecto |
| **Violated** | Un check falló | Se arregla |

- "No verificado" **nunca** pasa por "cumple".
- La evidencia de un agente (capturas, navegación) se registra como **`observed`**: se muestra, pero **nunca cuenta como Holds**. Sirve para que la persona confirme rápido, no para reemplazarla.
- Una promesa automática que todavía no se corrió en este clon es **Unverified** con la razón `not-run-here`. No es un quinto estado: es un detalle del mismo (ver abajo).

## Promesas manuales y de UI

Beacon **no trae un navegador**: usa los que el proyecto ya tenga a través de ejecutores. Una promesa de UI cae en uno de cuatro casos, según qué se pueda comprobar con código:

| Promesa de UI | Cómo se comprueba | Estado |
|---|---|---|
| Comportamiento (un botón abre algo, un formulario muestra un error) | Test E2E (por ejemplo con Playwright) como ejecutor | Holds |
| Estructura y accesibilidad (roles, nombres accesibles) | Aserciones sobre el árbol de accesibilidad, también como ejecutor | Holds |
| "No cambió visualmente" | Un ejecutor de regresión visual compara contra una **captura base** guardada en git | Holds, pero solo en el sentido de "igual a la base aprobada" |
| "Se ve bien" (criterio humano) | `check: {manual: "<pasos>"}`; el agente puede dejar evidencia (captura o video), que es `observed`, y la persona atesta | Attested, con la evidencia enlazada |

- **La captura base es lo aprobado.** El ejecutor visual declara en `config.yaml` qué archivos son sus bases (`baseline`, `records`). **Actualizar una captura base es un cambio de nivel 3**: sin eso, un agente podría hacer pasar una regresión simplemente regrabando la base (`changes`).
- **Manual:** queda en `Unverified` hasta que una persona la atestigüe y pase a `Attested`. `beacon attest --evidence <ruta o enlace>` guarda la referencia a lo que se revisó dentro de `attested` (`evidence`); Beacon no guarda el archivo, y la evidencia nunca sustituye el sí de la persona.
- **Para que no sea ceremonia:** la atestación **no se pide en cada cambio**. Al cerrar la sesión, `beacon attest` lista solo las promesas manuales **cuyo código cambió**, con una pregunta de sí o no por cada una.
- **Qué no prueba esto:** que el agente mire una captura ayuda a iterar y la regresión visual detecta cambios en el tiempo, pero ninguna de las dos demuestra que la interfaz sea correcta. Por eso "se ve bien" sigue siendo de la persona.
- Los ejecutores de ejemplo (`test-e2e` con Playwright y `visual`) se documentan en `config.yaml` de ejemplo (`records`). La regresión visual exige un renderizado determinista (fuentes y animaciones fijas); es un requisito del proyecto, no de Beacon.

## Deriva

Una promesa en `Holds` o `Attested` cuyo código (`paths`) cambió desde su última comprobación se muestra como **deriva**: no deja de cumplirse, pero ya no está demostrada. Se resuelve volviendo a comprobar o a atestiguar. Sin `paths` no hay detección de deriva (`records`, `spider`).

## Dónde viven los resultados

- Los resultados de ejecutar checks (qué pasó y en qué commit) **no se guardan en los registros**: cambiarían en cada ejecución y llenarían los PRs de ruido. Viven en `.git/beacon/results.json`, fuera de `.beacon/` y sin commitear (`storage`), por id de promesa. Son la base de la deriva en ese clon. Los worktrees de un mismo clon comparten ese archivo, así que **se escribe de forma atómica** (`storage`).
- **En CI se recalculan desde cero**, que es la fuente compartida de verdad.
- Las **atestaciones manuales sí se guardan en el registro** (`attested: {by, at, commit}` y, si hay, `evidence`): son una decisión humana y deben viajar con el código.
- **En un clon nuevo no hay resultados.** Las promesas automáticas quedan `Unverified` con la razón `not-run-here`, y el `brief` las resume en una sola línea ("N promesas sin comprobar en esta máquina") en vez de listarlas una por una (`spider`). Se resuelven corriendo `beacon check`.

## Cierre

Una promesa que deja de valer **no se borra**: se cierra con fecha y motivo (`closed` y `reason`) y pasa a `archive/` (`principles` P5, `storage`).