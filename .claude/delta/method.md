# Flujo de trabajo

Cómo trabaja el agente con Beacon. Es ligero: el binario respalda los pasos que de otro modo dependerían de que el agente obedezca.

## Pasos

| Paso | Qué hace el agente | Qué respalda el binario |
|---|---|---|
| 0. Empezar | Lee `beacon brief` | El hook de inicio lo ejecuta e inyecta el resultado donde exista, y lo repite tras compactar el contexto; si no, la guía lo pide (`anchor`) |
| 1. Autorizar | Decide si la petición pide un cambio (explicar o investigar es solo lectura) | — |
| 2. Orientar | Contrasta la petición con la meta activa; si se aleja o tiene un costo oculto, lo dice en una línea | `brief` da la meta y los `nongoals` |
| 3. Explorar | Lee el código y `beacon context <paths>` | Recuperación por relación (`spider`); un hook antes de editar inyecta lo que aplica a cada archivo, una vez por sesión (`anchor`) |
| 4. Resolver incertidumbre | Lo que no se puede decidir sin inventar se pregunta (si es de producto) o se registra como `question` | Una `question` bloquea el trabajo relacionado en `ready` |
| 5. Medir tamaño | ¿Se reanuda desde la petición y el diff? Si no, crea un `work` con la petición textual | El `work` guarda la petición y las tareas |
| 6. Implementar | Test primero; cada test cita el id de su promesa; registra decisiones (con las alternativas que descartó) y cierra preguntas al resolverlas; las promesas nuevas dicen un resultado observable con entradas y salidas concretas (`pledge`) | Validación de registros y niveles de cambio |
| 7. Comprobar | `beacon check` en las capabilities tocadas | Ejecución y veredicto por código (`pledge`) |
| 8. Cerrar | Informe: promesas tocadas y su estado, cambios en el estado del proyecto, pendientes | `beacon diff` da el resumen clasificado; los de nivel 3 esperan un sí |
| 9. Entregar | **La persona** decide commit, push y PR | Git hook y CI ejecutan `beacon check` (`snares`), también sobre la fusión en la cola de merge |

## Tamaño por reanudabilidad

- **Pequeño:** se entiende, el riesgo está contenido y se puede reanudar desde la petición más el `git diff`. Va en línea, sin `work` ni registro.
- **Grande:** falla la prueba de reanudación (varias sesiones, esperas externas, requisitos demasiado largos para sobrevivir a la compactación del contexto). Se crea un `work` con la petición textual.
- **El número de archivos nunca decide.**

## work y task

- **`work`:** el contenedor. Guarda la petición textual del usuario, el log y las promesas que implementa.
- **`task`:** el nodo del grafo, hijo de un `work`, con id jerárquico (`bc-91cd.1`), dependencias tipadas y enlace `implements` hacia una promesa (`ledger`).
- **El estado de la tarea viaja en el mismo commit que el código:** la tarea se cierra en el commit donde pasa el check de su promesa.
- Un cambio de requisito reescribe solo la promesa afectada y reabre solo su tarea.
- Los subagentes reciben el id del `work`, el de su tarea y sus promesas **por referencia**, nunca parafraseados.

## Claims y trabajo en paralelo

La práctica común con varios agentes es una tarea, una rama, un worktree y una sesión (`survey`, 7.4). Beacon se apoya en eso y lo hace visible:

- `claim` es **consultivo** y escribe en **dos sitios**: un claim local en `.git/beacon/` (`layout`) y `owner` y `branch` en el registro de la tarea.
- **El claim local avisa al instante** a las demás sesiones de la misma máquina, aunque estén en otro worktree: si la tarea ya está reclamada, `beacon claim` sale con código 2 y el `hint` dice quién la tiene.
- **`owner` y `branch` en el registro avisan al equipo** cuando la rama sube. Si dos ramas reclaman la misma tarea en máquinas distintas, git lo muestra como **conflicto en ese archivo**; el conflicto es visible, que es lo que se quiere.
- `ready` se calcula **por rama** y tiene en cuenta los claims locales.
- Un claim no es un candado: la persona puede tomar una tarea reclamada. Lo que evita es que ocurra sin enterarse.
- **Ramas largas** (qué pasa con una promesa cambiada en la rama principal mientras una rama vive) no está resuelto del todo: la hipótesis es que `beacon diff` compare contra el punto de origen de la rama (`git merge-base`, `weaver`), y se prueba con casos reales en la Fase 2 (`status`).

## Estado visible y preguntas

- El agente reporta: *Working*, *Checking*, *Ready*, *Needs your decision*. Esto último solo aparece para una decisión humana real, un stop, una denegación o un límite agotado.
- **Nunca se pide permiso para seguir con trabajo autorizado.**
- Al usuario solo se le preguntan decisiones de producto reales, **de una en una**.

## Delegación

Se delega a un subagente solo por una razón con nombre: un mapa de exploración necesario para decidir; escritores en paralelo sobre superficies disjuntas; un respaldo de contexto en sesiones largas; o un verificador independiente para un cambio de riesgo alto (`deltas`). Ante la duda, el agente trabaja en línea.

## Tests

Test primero cuando haya un test determinista ejecutable y un resultado esperado claro. **Un test RED por regla pedida**, un test de regresión por cada comando u opción existente que se toque, y no se añaden otros casos.

## Arranque en un repo existente

El arranque da valor en la primera sesión sin que el agente invente (`tenets` P1):

1. El **binario** hace un escaneo mecánico: stack, paquetes del monorepo, comandos de test (propone los **ejecutores** de `config.yaml`) y carpetas candidatas a capability. Todo queda en estado `proposed`.
2. Lo que el binario **no puede saber** (la meta, qué prometer, qué no se hace) se vuelve una `question` que el agente le hace a la persona **de una en una**.
3. **Nada pasa a `active` sin un sí.**
4. Si el proyecto necesita una guide de arquitectura, el arranque la crea como `draft` con un **esqueleto corto** inspirado en arc42 (contexto y alcance, bloques, conceptos transversales, riesgos y deuda), con los títulos y solo lo que el escaneo encontró (por ejemplo los paquetes); lo demás queda como `question`. Se quitan las secciones que nadie necesita (`survey`, 7.1).