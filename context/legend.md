# Referencia

## Glosario

| Término | Significado |
|---|---|
| **PAD (Promise-Anchored Development)** | Desarrollo anclado a promesas, el método de Beacon: el agente cambia el código, pero lo prometido se comprueba fuera del modelo y no se suelta sin un sí humano. Beacon es la herramienta; PAD es el método (`vision`) |
| **Contrato vivo** | Lo que una parte del sistema promete, verificado de forma continua, que no se debilita sin un sí (`vision`) |
| **Promesa** | Garantía de una capability con su `check` (`pledge`) |
| **Promise ratchet (trinquete de promesas)** | Regla por la que agregar una promesa `proposed` es libre, pero activarla, editarla, debilitarla o sacarla de `active` exige un sí humano (`deltas`) |
| **Promise ledger (libro de promesas)** | Las promesas de todas las capabilities, versionadas en git junto al código. Los resultados de los checks quedan fuera, en `.git/beacon/` (`ledger`, `pledge`) |
| **Capability** | Una parte del sistema con intención, `paths` y promesas (`ledger`) |
| **Check** | Datos que apuntan a un ejecutor, o `manual` (`pledge`) |
| **Ejecutor (runner)** | Comando declarado en `config.yaml` que corre un check |
| **Holds / Attested / Unverified / Violated** | Estados de verificación (`pledge`) |
| **observed** | Evidencia de un agente; nunca cuenta como Holds |
| **Deriva (proof decay, prueba vencida)** | Promesa en Holds o Attested cuyo código cambió desde su última comprobación: no está rota, pero ya no está demostrada (`pledge`) |
| **Brief** | Resumen mínimo que entra al empezar una sesión (`guides`) |
| **Context** | Lo que aplica a unos archivos (`spider`) |
| **Guide** | Contexto orientativo en Markdown (`guides`) |
| **Work / Task** | Trabajo grande y sus nodos (`method`) |
| **Claim** | Reserva consultiva de una tarea (`method`) |
| **Nivel 1, 2, 3** | Clasificación de cambios por riesgo (`deltas`) |
| **Spider, Engine, Render, Anchor, Sentry** | Componentes (`design`) |
| **ODD** | Organic Driven Development, el método de guía de Gentle-AI según su README actual |
| **RDD** | Receipt-Driven Development, el flujo de revisión de Gentle-AI donde el binario decide el siguiente paso |
| **SDD** | Desarrollo guiado por specs; Böckeler distingue spec-first, spec-anchored y spec-as-source (`survey`) |
| **Harness engineering** | Diseño de los controles y verificaciones alrededor de un agente; es la categoría en la que entra Beacon |
| **Fitness function** | Comprobación objetiva y continua de una propiedad del sistema; antecedente de los checks de promesas |
| **Living documentation** | Documentación que se mantiene al día con el software; antecedente de los contratos vivos |
| **Promise Theory** | Teoría de Mark Burgess sobre agentes autónomos que declaran su comportamiento como promesas aún no verificadas |
| **ADR / MADR** | Registro de una decisión (contexto, opciones, resultado, estado); MADR es su plantilla en Markdown |
| **EARS** | Sintaxis de requisitos del tipo "WHEN … THE SYSTEM SHALL …", usada por Kiro |
| **Worktree** | Copia de trabajo adicional del mismo repo de git, con su propia rama |
| **MCP** | Protocolo para exponer herramientas a agentes |

## Fuentes

Confianza según `survey`. Las citas vienen de documentación de proyectos, blogs y estudios; varias se recordaron de memoria y conviene revisarlas antes de citarlas fuera de esta KB. Las marcadas "ronda 2026-11-22" se consultaron ese día.

| Tema | Fuente | Confianza |
|---|---|---|
| Cognición extendida en arañas | Japyassú y Laland (2017), hipótesis debatida | Media |
| Contexto curado frente a completo | Hindsight; estudios de "context rot" | Media |
| Archivos de contexto | Gloaguen et al., "Evaluating AGENTS.md" (arXiv 2602.11988, feb 2026; ETH Zurich). Ronda 2026-11-22: abstract y resultados | Media-alta |
| Decaimiento de instrucciones en la sesión | McMillan, "Instruction Adherence in Coding Agent Configuration Files" (arXiv 2605.10039, mayo 2026). Ronda 2026-11-22: solo abstract | Media-baja |
| Cumplimiento de instrucciones en agentes | Ding et al., OctoBench (ACL 2026, arXiv 2601.10343). Ronda 2026-11-22: solo abstract | Media |
| Efecto de las specs en defectos | Hill, ESEM 2026 (LIPIcs vol. 394, art. 84). Ronda 2026-11-22: paper leído | Media |
| Taxonomía y problemas de SDD | Böckeler, "Understanding Spec-Driven-Development: Kiro, spec-kit, and Tessl" (martinfowler.com, oct 2025); Thoughtworks Technology Radar vol. 33 (nov 2025). Ronda 2026-11-22 | Media-alta |
| Guía de `CLAUDE.md` | Documentación de Anthropic | Alta |
| Interfaz frente a sustrato | LangSmith, memoria de agentes de Anthropic, Letta frente a Mem0 | Baja a media |
| Cantidad de herramientas MCP | Blogs, guía de OpenAI, límite de Cursor | Media |
| Specs que se pudren | Estudio de 2026 sobre 470.795 archivos de spec (solo resumen) | Baja a media |
| OpenSpec | Repo Fission-AI/OpenSpec, esquema `spec-driven` y guías de terceros (ronda 2026-11-22) | Media |
| Spec Kit | README del repo github/spec-kit y documentación de Microsoft (ronda 2026-11-22). No confirmé el comando `spec-drift` | Media |
| Kiro | Documentación y blog de AWS, artículos de DZone y AWS Builder Center; charla de Al Harris (ronda 2026-11-22) | Media |
| Superpowers | Repo obra/superpowers y material de terceros (ronda 2026-11-22) | Media |
| BMAD, Tessl | Documentación de cada proyecto; Tessl solo a través de Böckeler | Media |
| ODD y Gentle-AI | README actual de Gentle-AI (ronda 2026-11-22; no se leyeron los documentos enlazados); escritos de sus autores | Media |
| Engram, Beads | Documentación de cada proyecto | Media |
| ADR y MADR | adr.github.io/madr (MADR 4.0.0), bliki de Martin Fowler (ronda 2026-11-22) | Alta |
| arc42 | arc42.org y docs.arc42.org (ronda 2026-11-22) | Alta |
| C4 | A través de arc42 y blogs; no se leyó la fuente original | Media |
| PRD y specs para agentes | Osmani (ene 2026), Makerkit, Augment (ronda 2026-11-22) | Baja a media |
| Worktrees y equipos de agentes | Blogs y guías de 2026 (Augment, MindStudio, dev.to); Collide y Vibsync son vendedores | Baja |
| Verificación de UI | Artículos sobre Playwright MCP (Argos, DZone), repos de ProofShot y agent-browser, Simon Willison (ronda 2026-11-22); herramientas no probadas | Media-baja |
| Agentes (rutas y límites) | Documentación de cada agente, cambia rápido | Media |
| Nombres PAD, promise ratchet, proof decay y promise ledger | Búsquedas de la ronda 2026-11-22: no se halló uso de esas frases exactas, lo que no prueba que sean inéditas | Media-baja |
| Antecedentes de esos nombres | Test ratchet (Steve Kinney), EA-Graph (arXiv 2608.04278), Commitment Ledger (Towards Data Science). Ronda 2026-11-22: solo resultados de búsqueda | Media-baja |
| Métodos vecinos | ACDD (PyPI), agent-spec (docs.rs), Guarantee-Driven Development (dev.to). Ronda 2026-11-22: solo resúmenes | Baja a media |
| Harness engineering | Fowler, "Harness engineering for coding agent users" (martinfowler.com); Osmani; LangChain. Ronda 2026-11-22: solo resultados de búsqueda | Media |
| Fitness functions | Thoughtworks Technology Radar; InfoQ. No se leyó *Building Evolutionary Architectures*. Ronda 2026-11-22 | Media |
| Living documentation | Adzic, *Specification by Example*, a través de InfoQ y resúmenes. Ronda 2026-11-22 | Media |
| Promise Theory | Burgess y Bergstra, a través de Wikipedia y arXiv 2604.10505. Ronda 2026-11-22: solo resultados de búsqueda | Media |
