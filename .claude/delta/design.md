# Arquitectura

## Componentes

| Componente | Responsabilidad | Detalle en |
|---|---|---|
| **Spider** | Registros, grafo de relaciones en memoria y señales (bloqueos, rojos, deriva, frescura) | `spider`, `layout`, `ledger` |
| **Weaver** | CLI (`--json` en todo; la única interfaz), ejecución de checks, niveles de cambio, riesgo y flujo | `weaver`, `pledge`, `deltas`, `method` |
| **Anchor** | Instala Beacon en cada agente: guía del método, hooks (incluida la reinyección) y permisos | `anchor` |
| **Snares** | Git hooks y CI: ejecuta checks, también sobre la fusión, y verifica aprobaciones | `snares` |

Los nombres siguen la metáfora de la telaraña: Spider es la tela, Weaver es quien la teje y decide sobre ella, Anchor son las líneas de anclaje que la fijan al entorno (cada agente es un punto de amarre) y Snares son las trampas que atrapan lo que viola una promesa antes de que llegue a la rama principal.

### Frontera entre componentes

Se separan como **módulos con frontera limpia dentro de un solo binario y un solo repo**. Spider es una librería con interfaz estable (registros, relaciones, señales) y Weaver la consume. Si algún día Spider sirve por sí solo, se extrae entonces. Si hace falta otro componente, se agrega.

Dentro de Spider, los registros viven en los contextos `ledger`, `memory` y `agenda`, y el grafo con sus señales en `thread`, que solo conoce ids y relaciones tipadas. La estructura de carpetas está en el `README.md`.

### Interfaces

El **CLI** es la única interfaz: lo usan los agentes, los hooks y Snares. No hay servidor MCP, porque no añadiría garantías y sumaría otra superficie que mantener (`weaver`, `status`).

## Dónde se hace cumplir el método

Capas, de más fuerte a más débil:

1. **Snares (CI y git hooks):** no dependen de que el agente coopere. Un PR con una promesa violada, o una promesa debilitada sin aprobar, no pasa; y la cola de merge comprueba el resultado de la fusión.
2. **Binario:** las operaciones validan. Un cambio de nivel 3 sin confirmación no se guarda.
3. **Hooks nativos del agente:** recordatorios y reinyección donde existan. Al empezar y tras compactar el contexto ejecutan `beacon brief`; antes de editar un archivo inyectan lo que aplica a él. El agente recibe el estado sin tener que pedirlo, y se refresca durante la sesión porque la obediencia a las instrucciones decae con ella (`survey`, E7). Un extra donde existan.
4. **Guía en el prompt:** orienta el flujo. Es lo único que depende de la obediencia del modelo.

Las garantías viven en las capas 1 y 2, que funcionan con **cualquier** agente. La capa 4 solo orienta.

**Límite:** nada de esto garantiza que el agente *llame* a las herramientas. Las capas 1 y 2 aseguran que las promesas se cumplan y no se debiliten, pero que la memoria se llene y se use depende de que el agente invoque `brief`, `context` y `record` (`status`, riesgos).

**Otro límite:** el binario no puede distinguir con certeza a la persona del agente, porque corren con el mismo usuario del sistema. El marcado del actor ayuda, pero la garantía real en equipo es la revisión del PR por una persona (`deltas`).

## Qué decide el código y qué el modelo

| Código (determinista) | Modelo (propone) |
|---|---|
| Si un check pasó; qué está listo; si un cambio debilita una promesa; qué contexto aplica a un archivo; qué está desactualizado; validar la forma de un registro; clasificar el nivel de un cambio; calcular el riesgo | Qué vale la pena registrar; el texto de una promesa o decisión; qué preguntar; cómo resolver una incertidumbre; escribir el código |

## Lenguaje y distribución

- **Go:** un solo binario sin dependencias.
- **Distribución:** binario estático por GitHub Releases, `go install`, Homebrew y Scoop (Windows).
- **Licencia:** MIT.
- **Sin telemetría.**
- **Idioma:** las claves del esquema, la guía del método y los mensajes del binario (errores, bloqueos y `hint`) van en inglés; el contenido de los registros (textos, promesas, guides) puede estar en cualquier idioma.