# Snares

Git hooks y CI. Es la capa más fuerte: no depende de que el agente coopere (`design`).

## Pre-commit

- Valida el formato y el esquema de los registros (`layout`).
- Compara contra la versión de git y detecta cambios de nivel 3, incluidas las ediciones manuales y las capturas base (`deltas`).
- Corre los checks relacionados con lo que cambió (`pledge`).

## CI

Snares no se ata a una plataforma: el binario trae **un comando genérico, `beacon snares ci`**, que cualquier CI puede correr, y se entrega con **una plantilla de GitHub Actions** (la primera plataforma). Otras plataformas, como GitLab, son otra plantilla y no tocan el binario.

`beacon snares ci` hace, en una sola llamada:

1. **Recalcula los checks desde cero**: es la fuente compartida de verdad (`pledge`).
2. **Usa los `runners` de `config.yaml` de la rama base** hasta que un cambio a los ejecutores esté aprobado. Un PR no puede ejecutar ejecutores que él mismo introduce.
3. **Verifica aprobaciones:** todo cambio de nivel 3 debe traer un `approved.by` humano (`deltas`).
4. **Comenta en el PR** con `beacon diff` clasificado por nivel, que es lo que debe leer el revisor.
5. Falla si hay una promesa en rojo o un nivel 3 sin aprobar.

**El CI necesita el historial.** `beacon diff` compara contra la rama base y `stale` usa fechas, así que no sirve un clon superficial. En GitHub Actions hay que usar `fetch-depth: 0`, porque el checkout por defecto trae solo el último commit; la plantilla ya lo incluye.

### También sobre la fusión

Hay cambios que fusionan limpio en git pero dejan el proyecto roto: una persona renombra una función y otra escribe código nuevo que la llama por el nombre viejo. Un CI que solo comprueba la rama del PR no lo ve; solo lo ve uno que corre sobre **el resultado de la fusión** (`layout`).

- La plantilla de GitHub Actions escucha `pull_request` y también `merge_group`, el evento de la **cola de merge**, para que `beacon snares ci` corra sobre el commit ya fusionado antes de que llegue a la rama principal.
- En ese evento la base de la comparación es la que trae el propio evento, no la punta de la rama del PR.
- La cola de merge es opcional en GitHub: sin ella, el evento no se dispara y la plantilla sigue funcionando para PRs. **El nombre y el comportamiento exactos del evento se verifican al escribir la plantilla** (viene de mi conocimiento de GitHub, no de una fuente consultada).

## Equipos

- **La garantía real es la revisión del PR por una persona.** Quién puede aprobar lo define el hosting: CODEOWNERS o review obligatoria sobre `.beacon/records/capabilities/**` y sobre las capturas base de los ejecutores visuales. El campo `approved` de los registros es un marcado (`deltas`).
- `.gitattributes` marca `.beacon/**` como `linguist-generated` para que GitHub colapse los diffs (`layout`).

## Seguridad

El CI ejecuta código del repo, por eso los ejecutores son pocos, sin shell, con parámetros que no pueden empezar con `-`, y su cambio es de nivel 3 (`pledge`, `deltas`).
