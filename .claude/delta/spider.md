# Spider

La memoria de Beacon. No es un diario al que se le pregunta: es una estructura que empuja al agente.

## El principio

La idea viene de una hipótesis de psicología sobre cognición extendida en arañas tejedoras (Japyassú y Laland, 2017; es una hipótesis debatida): la telaraña no sería solo una herramienta, sino parte de cómo la araña "piensa", porque opera sobre ella, siente en ella las vibraciones y la reconfigura. Se usa como **principio de diseño**, no como afirmación científica: la memoria no se consulta, se habita.

Una estructura así:

- **Se percibe sin preguntar:** las señales llegan al agente cuando toca una zona.
- **Condiciona la acción:** una pregunta abierta bloquea el trabajo relacionado; un cambio de nivel 3 no se guarda sin un sí.
- **Se modifica al usarse:** lo que el agente descubre deja rastro (`origin`) y la próxima sesión encuentra la tela cambiada.
- **Propaga la tensión por las relaciones:** tocar un hilo (`blocks`, `serves`, `supersedes`, `refers`) afecta a los vecinos.

## Qué hace Spider

Carga los registros (`layout`, `ledger`), arma el grafo de relaciones en memoria y calcula, **sin modelo**, lo que el resto del sistema necesita. Engine expone estos cálculos como comandos y herramientas (`engine`).

## Qué calcula

| Cálculo | Qué devuelve | Cómo |
|---|---|---|
| `brief` | Meta activa con sus `nongoals`, bloqueos y rojos, trabajo en curso, índice de guides | Lee goal, preguntas abiertas, estado de las promesas, works activos y guides; con tope de líneas (ver abajo) |
| `context <paths>` | Todo lo que aplica a esos archivos: capabilities cuyos `paths` coinciden (con sus `nongoals`), sus promesas con estado, decisiones con esos `where` (con sus alternativas `rejected`), guides con `applies` coincidente o `always`, preguntas relacionadas y las señales | **Recuperación por relación:** coincidencia de globs y recorrido de relaciones, no búsqueda de texto |
| `context --hook <paths>` | Lo mismo, pero solo lo que la sesión todavía no recibió, con tope de `limits.hook` líneas | Igual que `context`, más el registro local de lo ya inyectado (`layout`, `anchor`) |
| `ready` | Trabajo y tareas sin bloqueos, por rama | Recorre `blocks` y `parent`; al cerrar un bloqueo se desbloquean los dependientes; tiene en cuenta los claims locales (`method`) |
| `drift` | Promesas que pasaron pero cuyo código cambió desde el último check | Compara el commit del último resultado (`pledge`) con los cambios de git bajo sus `paths` |
| `stale` | Decisiones y guides que superaron su plazo de revisión o cuyo código cambió mucho | `reviewed` más `review`, y cambio de código desde la última revisión (`guides`) |
| `diff` | Qué cambió en el estado desde un commit, incluidas las capturas base, **sin clasificar** | Compara los registros actuales con los de git y los archivos de `baseline`; Engine asigna el nivel (`deltas`, `engine`) |

**Revelado progresivo:** primero un resumen (`brief`), luego lo que aplica a lo que se toca (`context`), y solo si hace falta el detalle de un registro (`get`).

Los `nongoals` y los `rejected` aparecen en `context` porque son lo que más ayuda a que el agente no construya lo que nadie pidió ni vuelva a proponer lo ya descartado.

## Cómo consulta git

Spider consulta git **en lote**, nunca una vez por promesa (`layout` tiene las mediciones):

- **Deriva:** una consulta `git diff --name-only <base>..HEAD` por cada commit base distinto entre los resultados guardados (suelen ser pocos). El cruce con los `paths` se hace en memoria.
- **Frescura:** una consulta `git log --since=<la fecha de revisión más antigua> --name-only`, y el conteo de cambios por glob también en memoria.

## El tope del brief

El `brief` tiene un tope de `limits.brief` líneas (60 por defecto, `ledger`). La meta activa siempre entra. Si el resto no cabe, se recorta por **prioridad fija**:

1. Bloqueos y rojos.
2. Trabajo en curso.
3. Pendientes de sí (los aporta Engine).
4. Deriva y la línea de promesas sin comprobar aquí.
5. Índice de guides (lo primero en recortarse).

Lo que no cabe se resume en una línea (`N more, run beacon status`). Es determinista y no pierde lo crítico. La inyección antes de editar sigue el mismo criterio con su propio tope, `limits.hook`.

## Señales

Spider las emite sin que se le pregunte, dentro de `brief` y `context`:

| Señal | Cuándo aparece | Efecto |
|---|---|---|
| **Bloqueo** | Una pregunta abierta con `blocks` hacia un trabajo o tarea | Ese trabajo no aparece en `ready` |
| **Rojo** | Una promesa en `Violated` bajo los `paths` que se tocan | Aparece destacada en `context` |
| **Deriva** | Una promesa que pasó, pero su código cambió desde el último check | Deja de verse como vigente hasta volver a comprobar |
| **Sin alcance** | Una promesa sin `paths` | No puede detectar deriva |
| **Sin comprobar aquí** | Una promesa automática sin resultado en este clon (`not-run-here`) | Se resume en una línea del `brief` |
| **Desactualizado** | Una guide o decisión pasada de su plazo de `review` o con mucho cambio de código | Aparece en `stale` y en el `brief` |
| **Tensión por relación** | Tocar un hilo | Se muestran los vecinos afectados |
| **Pendiente de sí** | Un cambio de nivel 3 sin aprobar. Engine lo clasifica y Spider lo muestra | Bloquea el guardado o el PR |