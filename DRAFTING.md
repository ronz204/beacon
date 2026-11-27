# Borradores

Ideas en bruto que todavía no son decisiones. Nada de aquí está en el orden de construcción (`.claude/delta/status.md`) hasta que se mida y se decida.

---

## La segunda dimensión de la telaraña

**Estado:** borrador. Idea de la persona dueña del proyecto, desarrollada el 2026-11-27 tras comparar Spider con Graphify.

### La idea en una frase

Spider hoy teje una sola dimensión: el grafo de compromisos (metas, promesas, decisiones, preguntas, trabajo). Una segunda dimensión sería un grafo liviano del código, y lo valioso es la arista que une las dos.

### Qué problema ataca

Que el agente navegue el código más rápido y gaste menos tokens.

- Hoy el agente descubre el código con `grep` y lecturas de archivos completos. Cada exploración cuesta tokens y llamadas a herramientas.
- Spider ya le dice qué prometió el proyecto sobre una zona (`context`), pero no le dice qué otra zona del código depende de ella.
- Los `paths` de una capability son una pista gruesa y escrita a mano. Lo que queda fuera de ellos no genera ni tensión ni deriva.

**Hipótesis sin medir:** un mapa de archivos relacionados, entregado junto con los compromisos que aplican, reduce las lecturas exploratorias del agente. Graphify afirma cifras de ahorro de tokens, pero son afirmaciones de su sitio y no están verificadas.

### Qué sería

Un grafo de **archivos** con una sola clase de arista, `imports`, calculado en memoria al consultar, sin caché persistente (igual que el resto de Spider).

Lo que le daría a Spider:

1. **Tensión por dependencia en `context <paths>`:** al tocar un archivo, mostrar quién lo importa y qué capabilities (con sus promesas) cubren a esos importadores.
2. **Deriva transitiva:** una promesa deja de verse vigente si cambió algo que su código importa, aunque esté fuera de sus `paths`.
3. **Sugerencia de `paths`** en `doctor` o `record`: "esta capability importa archivos que sus `paths` no cubren".
4. **Mapa corto para navegar** (a explorar): junto al `context`, una lista acotada de archivos vecinos para que el agente lea lo justo y no busque a ciegas.

### Qué no sería

- No es un competidor de Graphify. No habría símbolos, llamadas, tipos ni flujo de datos.
- No es un servidor ni un índice persistente.
- No pretende ser exacto: no ve reflexión, inyección de dependencias ni configuración. Las señales dirían "puede estar afectada", nunca "está rota".

Graphify y Spider son complementarios: Graphify responde "qué llama a qué"; esta capa respondería "qué promesas se ven afectadas si toco esto".

### Opciones de parseo

| Opción | Costo | Cobertura |
|---|---|---|
| `go/parser` con `ImportsOnly` (biblioteca estándar) | Pocas líneas, sin dependencias nuevas | Solo Go |
| Expresiones regulares para imports de TS/JS y Python | Poco código, aproximado | Imports relativos; fallan los alias (`tsconfig`, `__init__`) |
| tree-sitter en Go puro (`gotreesitter`) | Dependencia grande con gramáticas embebidas; su propio README reporta parseo completo de 3 a 4 veces más lento que C | Muchos lenguajes |
| tree-sitter con cgo | Rompe el binario estático y el reparto por Scoop y Homebrew | Descartada |

Orden propuesto: Go con la biblioteca estándar; si sirve, regex para TS/JS y Python; tree-sitter solo si hace falta de verdad.

### Riesgos

- **Velocidad en el hook.** `context --hook` corre antes de cada edición y hoy evita incluso git. Calcular importadores inversos exige recorrer el repo. Sin medir.
- **Sin caché por decisión** ("fuente de verdad en git, cargada en memoria"). Si el escaneo pesa, la salida sería un índice en `.git/beacon/` con invalidación por fecha de modificación, que ya sería más trabajo y rompe un principio.
- **Falsa precisión** de un grafo de imports.
- **Deriva de alcance** hacia símbolos y llamadas.
- **Beacon no tiene `.beacon/` propio con capabilities**, así que validar el valor con este repo exige armar datos de ejemplo.

### Preguntas abiertas

- ¿El objetivo principal es ahorrar tokens de navegación o hacer más inteligentes las señales de deriva? Pueden pedir diseños distintos.
- ¿Cómo se mide el ahorro? Idea: la misma tarea con y sin el mapa, contando llamadas a herramientas y tokens leídos (el arnés de fricción de `testing/benchmark/` podría servir de base).
- ¿La capa vive en el hook, solo en `context` manual, o en un comando aparte?
- ¿Se versiona algo del grafo, o siempre se recalcula?

### Spike propuesto

Un día, solo Go, con un comando de prueba (por ejemplo `context --impact`):

1. Recorrer los `.go` con `parser.ImportsOnly` y armar las aristas entre archivos.
2. Dado un archivo tocado, listar sus importadores inversos y cruzarlos con los `paths` de las capabilities.
3. Medir el tiempo en este repo y en uno grande.
4. Decidir: si la señal es útil y rápida, entra en `context` y `drift`; si no, se descarta.

### Fuentes

- Comparación con Graphify (tree-sitter, MCP, grafo de llamadas e imports): <https://graphify.com/what-is-graphify>
- Runtime de tree-sitter sin cgo: <https://pkg.go.dev/github.com/odvcencio/gotreesitter@v0.13.3>
