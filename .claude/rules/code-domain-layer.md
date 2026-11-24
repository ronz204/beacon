---
paths:
  - "internal/kernel/**"
  - "internal/*/internal/domain/**"
---

# Domain Layer Conventions (Go)

This governs the domain layer: how bounded contexts are laid out and what they may import, the files inside a context, how each building block is written in Go, and how contexts work together without knowing each other. It holds for any data source, because the domain only sees the contracts it declares. The layout of `action` and `driver` is out of scope.

---

## Bounded contexts

- A context is one Go package at `internal/<piece>/internal/domain/<context>/`, because a Go folder is a package and a concept split across folders would need hand-written re-exports.
- Its name is one lowercase word, six letters when a natural word exists. No underscores, hyphens or glued words, and never a name repeated from the standard library when it would be imported beside it.
- A context imports only `internal/kernel` and the pure standard library, never another context, `action`, `driver`, `system`, runtime config or a third-party library.
- Pure standard library means packages such as `context`, `errors`, `maps`, `regexp`, `slices`, `strings`, `time` and `unicode`. Forbidden: `os`, `io/fs`, `net`, `os/exec`, `log`, `math/rand`, `crypto/rand` and calling `time.Now`.
- Domain types carry no `yaml` or `json` struct tags, because a tag is a storage format inside the domain.
- `internal/kernel` is the shared kernel. It sits outside every piece's own `internal` folder because every piece must import it, and it holds only what every context needs. Two contexts wanting the same thing is not enough: each keeps its own until a third needs it.
- The package is the public surface and capitalization is the visibility gate. There are no barrels. Exported names are unique across the context and carry their full domain noun (`CapabilityID`, no bare `ID`).

## Sub-slices

- A sub-slice is a cohesive concept: typically one aggregate root plus what exists only to serve it. It is a file prefix inside the context, one short word, never a folder. Split on two concepts, never on size or technical layer.
- Files are `<subslice>_<kind>.go`, prefix always kept, and only the kinds needed exist. There is no catch-all `contracts` file.
- A `<kind>` or `<role>` is never `test` or a Go OS or architecture name (`linux`, `windows`, `js`, `wasm`, `arm`), because Go reads those filename suffixes as build constraints and would silently exclude the file.
- Tests sit beside the code as `<subslice>_<kind>_test.go`, in the external package `<context>_test` unless they must reach an unexported rule.

| Kind | Holds |
|---|---|
| `entity` | The root entity and the state struct its restore function accepts |
| `parts` | Entities owned by the root |
| `vos` | Value objects, typed ids, closed literal sets |
| `policies` | Pure functions over several values or entities: inputs in, verdict out |
| `events` | Domain events the root emits |
| `errors` | The sub-slice's error alias, codes and fixed messages |
| `repository` | The entity's data-source interface: `Find…`, `List…` and, only when the process stores the entity, `Save…` |
| `<role>` | One interface the domain needs from outside (`locator`, `broadcaster`), one file each |

## Contracts

- A contract is a small Go interface in the sub-slice it serves (one every context shares, in the kernel), so every dependency points inward. No `I` prefix or `Interface` suffix. It uses domain language, never a technology's.
- A contract method that may do I/O takes `context.Context` first and returns `error` last.
- Its implementation lives in the piece's `driver` and is named for what it reaches. The in-memory one doubles as the test double. A domain package never references a concrete driver.
- The implementation turns every technology failure into the sub-slice's domain error, keeping the cause with `Unwrap`, so `action` can handle a failure without knowing what caused it.
- The domain has no query side. A shape built for one command or report is assembled in `action` from entities.
- Data entering from outside is decoded in `driver`, every field is parsed with the domain's parsers and the entity is rebuilt through its restore function, so drifted data fails on read.
- The mapper sits beside the driver and may go both ways when the process writes the source. Whatever the domain does not model, such as unknown YAML fields, stays in the driver and is written back unchanged.

## Building blocks

- A rule lives with the entity or value object it constrains. Anything no domain rule requires (a wrapper, a variant of a base, extra indirection) is removed.
- Go has no inheritance, so there is no base entity. The kernel offers only what generics express cleanly, such as `kernel.Error[C]`.

| Piece | Go shape |
|---|---|
| Value object | `type CapabilityName struct{ value string }`, with `ParseCapabilityName(string) (CapabilityName, error)` and `String()` |
| Identifier | A value object per id kind. The format is the issuer's; ids come from an injected generator |
| Literal set | `type PromiseStatus string`, typed constants, `PromiseStatuses()` and `ParsePromiseStatus` |
| Error | `type CapabilityError = kernel.Error[CapabilityCode]`, with closed `snake_case` codes and fixed messages |
| Entity, part | A struct with one unexported `state` field, a `Restore<Entity>` function, intention-named factories and `SameAs` |
| Event | A struct with `Type`, `EntityID`, `OccurredAt` and `Payload` of plain values |

- A value object wraps its value in an unexported field, because a defined type such as `type Name string` can be built by any conversion and would bypass the parser. It stays comparable and usable as a map key.
- A parser is the only constructor of a value object: it trims, normalizes, validates and fails with the sub-slice's domain error, passing bounds as structured details. The zero value is invalid and never crosses a boundary as a real value.
- A literal set is the one exception to the wrapper, because constants need a primitive type. Values from outside always pass through its parser, and code outside the context never converts to it.
- An error carries a stable `snake_case` code. Messages are fixed per code, in English, never interpolated; variable data travels in structured details. Callers inspect it with `errors.As`.
- Creation uses intention-named factories taking `(input <Name>Input, now time.Time)`, never positional fields. The domain never generates ids or reads the clock. A part is created only by an operation of its root.
- `Restore<Entity>` trusts its input apart from copying slices and maps. Identity is `SameAs`, never `==`.
- Go cannot freeze a struct, so entity state is unexported and read through accessors named for the field (`Name()`, never `GetName()`). An accessor returning a slice or map returns a copy. A derived fact is a method named as a computation.
- Entities use value receivers. A state change returns a new value through a private `with` that clones slices and maps and never changes the id.
- An event is named `<entity>.<past_tense_verb>`. An operation that raises events returns `(next, events, error)` and never dispatches; one that raises none returns `(next, error)`, or the bare value when it cannot fail.

## Collaboration between contexts

- A context refers to another's data by id only, as its own identifier value object, and never imports its types. A type two contexts both seem to need is a kernel primitive or two separate types.
- A change that must cause work elsewhere travels as an event of plain values: the operation returns it, `action` dispatches it once the entity is stored, and a listener calls the receiving context. Dispatch is in-process and consistency across contexts is eventual.
- To read what another context owns, the consumer declares its own role interface in its own words, returning only what its rule decides on, never the owner's entity.
- A flow that needs several contexts is an `action` workflow, never logic inside one of them.

---

## Non-goals

- Does not decide which contexts, sub-slices or aggregate boundaries exist; that belongs to the project's design documents.
- Does not govern the layout of `action` or `driver`.
