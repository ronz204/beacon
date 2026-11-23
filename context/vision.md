# Visión

## En una frase

> **Beacon es un método de trabajo para agentes de código que no confía en que el agente lo cumpla: lo comprueba.**

## El problema

Los agentes de código (Claude Code, Codex, Cursor, OpenCode y otros) escriben rápido pero trabajan sin rumbo:

- No recuerdan qué se decidió ni por qué. Cada sesión empieza de cero o con notas sueltas.
- Donde no saben algo, inventan. Eso produce sobreingeniería y supuestos ocultos.
- Un cambio nuevo puede romper algo viejo sin que nadie lo note.
- No saben a dónde va el producto, así que construyen lo que se les pide aunque se aleje de la meta.

Las respuestas actuales caen en dos extremos: procesos pesados (muchos documentos y ceremonia, que se pudren) o guías ligeras en prompts (nada obliga al agente a cumplir, y la verificación la reporta el propio agente). El detalle está en `survey`.

## Cómo lo atacamos

- **Ligero:** se habla normal, el trabajo pequeño no tiene ceremonia y solo el trabajo grande deja registro.
- **Fiable:** lo prometido se comprueba con código, fuera del modelo. El estado del proyecto vive en registros estructurados, no en Markdown que el agente edita a mano.
- **Orientado a producto:** el agente conoce la meta activa y avisa cuando una petición se aleja de ella.
- **Agnóstico:** el núcleo es un binario (CLI primero y servidor MCP después) que cualquier agente puede usar. Las integraciones con cada agente son adaptadores finos.
- **Ecosistema propio:** la memoria y el grafo de tareas son parte de Beacon, sin depender de otras herramientas.

## Contratos vivos

Lo que cada parte del sistema promete (sus *promesas*) es un **contrato vivo**:

- Se **verifica de forma continua** por código, no solo al terminar un cambio.
- **No se borra:** cuando deja de valer se cierra con fecha y motivo.
- **No se debilita sin un sí humano:** el binario detecta el cambio, no el modelo (`deltas`).
- **Avisa cuando envejece:** si el código cambió desde la última comprobación, aparece como deriva (`pledge`).

Es la respuesta a que las specs se pudren y se comprueban solo al final (`survey`).

## Pitch

> Los agentes de código escriben rápido, pero olvidan lo que se decidió, inventan donde no saben y rompen cosas sin avisar. Las soluciones actuales son o muy pesadas o muy blandas.
>
> **Beacon es un método ligero para trabajar con cualquier agente, donde lo importante se comprueba con código.** Guarda lo que el proyecto sabe y promete: la meta, las decisiones, las promesas de cada funcionalidad con su forma de verificarse y las preguntas abiertas. El agente trabaja con naturalidad, consulta ese estado y lo actualiza mientras programa. Beacon ejecuta las comprobaciones y avisa cuando algo se rompe, envejece o se aleja de la meta. La persona solo decide cuando cambia una promesa.

## Para quién

- Personas solas y equipos.
- Cualquier agente de código que pueda ejecutar comandos o hablar MCP, y que lea un archivo de instrucciones.
- Proyecto personal sin prisa comercial: se puede apuntar alto.

## Qué no es

- No es un diario ni una memoria de conversaciones.
- No es un gestor de tareas genérico, aunque incluye un grafo de tareas.
- No es un pipeline de fases obligatorias.
- No es un conjunto de instrucciones que el agente "debería" seguir.
- No reemplaza al control de versiones: viaja dentro de git.

## Posicionamiento

> **"ODD te dice cómo trabajar. SDD te hace escribir. Beacon comprueba."**