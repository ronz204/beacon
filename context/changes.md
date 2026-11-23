# Cambios

Qué cambios al estado del proyecto necesitan a la persona, y cómo se detectan.

## Para qué sirven los niveles

No todo cambio merece pedir permiso, solo los que podrían **debilitar lo que ya se prometió**. Lo que se evita es el vicio típico de los agentes: un test falla y, en vez de arreglar el código, el agente edita o borra la promesa para que "pase".

## Niveles

| Nivel | Qué es | Ejemplo | Qué pasa |
|---|---|---|---|
| 1 | Mantenimiento | Responder una pregunta abierta, agregar detalle a una guide, marcar una guide como revisada, cambiar `review`, `limits` o `agents` en `config.yaml`, anotar `rejected` o `nongoals` | El agente lo hace |
| 2 | Agregar | Una promesa nueva (`proposed`), una decisión nueva | El agente lo hace y aparece en el resumen de la sesión |
| 3 | Debilitar o cambiar lo verificado | Editar el texto de una promesa activa, borrarla, cambiar su check, cambiar la meta, cambiar los `runners` o el `schema` de `config.yaml`, **actualizar una captura base de un ejecutor visual** | Beacon lo detecta y pregunta "¿sí o no?" antes de guardarlo |

Pasar una promesa de `proposed` a `active` también requiere un sí de la persona.

## Regla mecánica del nivel 3

*Cambiar el texto o el check de una promesa `active`, sacarla de `active`, borrarla, cambiar la meta, cambiar los `runners` o el `schema` de `config.yaml`, o modificar, añadir o borrar un archivo que cubra el `baseline` de un ejecutor es nivel 3.*

Es conservadora: puede pedir un sí por un typo, pero nunca deja pasar un debilitamiento.

**Atajo:** si el texto **normalizado** (sin mayúsculas ni puntuación, espacios colapsados) queda idéntico, el cambio es de nivel 1. Un typo con letras cambiadas sí pregunta.

### Capturas base

Un ejecutor visual declara en `config.yaml` los globs de sus capturas base (`baseline`, `records`). Para el binario, esos archivos son parte de lo prometido: una regresión visual pasa porque la captura nueva coincide con la base, así que regrabar la base es la forma más fácil de "arreglar" una regresión sin arreglarla.

- La detección es mecánica: un diff de git que toque un archivo cubierto por algún `baseline` es nivel 3, sin mirar la imagen.
- Se aprueba como cualquier otro nivel 3 (ver abajo), y la persona ve en el PR qué capturas cambiaron.
- Es conservador: una captura nueva por una pantalla nueva también pide el sí. Si eso resulta demasiado, se mide en la fricción y se afina el `baseline`.

## Quién detecta

El binario, no el modelo:

- Cuando el cambio pasa por una operación del agente, compara el registro antes y después.
- En el pre-commit y en el CI compara contra la versión de git, así que también se detectan las ediciones manuales.

## Ejemplo

La promesa dice *"tras 5 intentos fallidos la cuenta se bloquea 15 minutos"*. Para pasar un test, el agente la cambia a *"tras 10 intentos"*. El binario ve que cambió el texto de una promesa `active`: nivel 3. El cambio no se guarda hasta que la persona diga sí.

## Aprobación

No hay un sistema propio de aprobaciones; se reutiliza lo que git ya hace.

- **En solo:** la persona dice sí en el chat y el agente lo transmite con `beacon record --approve`. El binario guarda `approved: {by, at, via}` en el registro afectado: `by` es la persona (`human:<user.name de git>`) y `via` es el agente que transmitió el sí (`records`, sección Actor). Para una captura base, que no es un registro, el sí queda en el registro de la capability o promesa afectada.
- **En equipo:** Sentry exige en el PR dos cosas (`sentry`): que cada cambio de nivel 3 traiga un `approved.by` humano, y que el PR tenga una review aprobada de una persona con permiso sobre `.beacon/records/capabilities/**` (CODEOWNERS o review obligatoria del hosting). Las capturas base se protegen igual: la review de la persona sobre esos archivos es la garantía.
- **Qué garantiza y qué no:** el campo `approved` es un **marcado, no una prueba**, porque el agente corre con el mismo usuario del sistema y podría escribir un sí que la persona no dio. Lo que sí es una garantía en equipo es la revisión del PR por una persona: ahí se ve el diff clasificado por nivel. En uso personal la protección es la confianza y que todo queda a la vista.

## Riesgo y revisión independiente

- **El riesgo lo calcula el código.** Es **alto** si hay un cambio de nivel 3 (incluye los ejecutores, el esquema y las capturas base) o se toca una zona con promesas en rojo o con deriva. Si no, es **bajo**.
- **Riesgo bajo:** no pasa nada.
- **Riesgo alto:** `beacon verify` corre los checks relacionados y pide **revisión independiente**: un agente con contexto limpio que solo ve las promesas y el diff, o una persona.
- **El agente puede subir el riesgo, nunca bajarlo.**
- **La revisión nunca hace commit ni push:** la persona es dueña de la entrega.

## Los ejecutores son código sensible

Los `runners` de `config.yaml` contienen lo que el CI ejecuta. Cambiarlos, igual que el `schema`, es de nivel 3, y en CI se usan los `runners` de la rama base hasta que alguien apruebe el cambio (`sentry`). Las demás claves de `config.yaml` no ejecutan nada y son de nivel 1, para no gastar síes en cambios inofensivos.

## Medir la fricción

Si el nivel 3 pide demasiado, se desactiva: un sistema que obliga a ceremonia para cambiar una coma se desinstala en tres días. Por eso la fricción se mide (`engine`).