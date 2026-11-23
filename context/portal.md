# Beacon: índice

Beacon es un método de trabajo para agentes de código que comprueba con código lo que el proyecto promete. Esta KB lo describe completo y solo en su estado vigente.

## Por dónde empezar

| Si quieres... | Lee, en este orden |
|---|---|
| Entender qué es Beacon y cómo se comporta | `vision` → `tenets` → `design` → `pledge` → `deltas` → `method` |
| Entender cómo se guardan los datos | `layout` → `ledger` → `guides` → `spider` |
| Entender cómo se conecta con agentes, git y CI | `engine` → `anchor` → `sentry` |
| Saber en qué punto está el proyecto | `status` |
| Consultar el porqué o un término | `survey`, `legend` |

Los seis primeros archivos del primer camino bastan para entender el proyecto completo.

## Mapa

| Archivo | Responde a |
|---|---|
| `vision` | ¿Qué es Beacon, para quién, y qué son los contratos vivos? |
| `tenets` | ¿Qué reglas guían todas las decisiones? |
| `survey` | ¿Qué hacen otras herramientas y qué evidencia sostiene las decisiones? |
| `design` | ¿Qué componentes hay y dónde se hace cumplir el método? |
| `layout` | ¿Cómo se guarda el estado en el repo? |
| `ledger` | ¿Qué tipos de registro existen y cómo son? |
| `guides` | ¿Cómo se maneja el contexto orientativo y qué entra al `brief`? |
| `spider` | ¿Cómo funciona la memoria activa y qué señales emite? |
| `pledge` | ¿Cómo se verifican las promesas? |
| `deltas` | ¿Qué cambios necesitan aprobación y cómo se detectan? |
| `method` | ¿Cómo trabaja el agente, paso a paso, y cómo se arranca en un repo existente? |
| `engine` | ¿Qué comandos existen? |
| `anchor` | ¿Cómo se instala Beacon en cada agente? |
| `sentry` | ¿Qué se hace cumplir en git y CI? |
| `status` | ¿Qué está decidido, qué está abierto, qué riesgos hay y qué se construye primero? |
| `legend` | ¿Qué significa un término y de dónde salió una fuente? |

## Reglas de esta KB

- Cada hecho vive en **un solo archivo**; los demás lo refieren por nombre.
- Solo se escribe lo vigente.
- Lo que no tiene marca está decidido. Lo no confirmado dice **Propuesta**. Lo abierto vive en `status`.