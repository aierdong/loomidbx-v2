# Research Log

## Feature

`phase-03-generation-context`

## Research Scope

本研究围绕 Phase 3 执行引擎的 GenerationContext 边界展开，目标是确认它如何消费 lifecycle 执行输入、dependency `ExecutionPlan`、rowcount `RowCountPlan` 以及 Phase 2 Project / Schema / Rule / Relation 快照，并为后续 batch generation loop 与 generator framework 提供稳定只读上下文和运行态引用存储。

## Sources Reviewed

- `.kiro/specs/phase-03-generation-context/brief.md`
- `.kiro/steering/roadmap.md`
- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`
- `.kiro/specs/phase-03-execution-lifecycle/requirements.md`
- `.kiro/specs/phase-03-execution-lifecycle/design.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/requirements.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/design.md`
- `.kiro/specs/phase-03-row-count-planning/requirements.md`
- `.kiro/specs/phase-03-row-count-planning/design.md`
- `internal/domain/execution/generationjob.go`
- `internal/domain/execution/executiontask.go`
- `internal/domain/project/project.go`
- `internal/domain/project/projecttable.go`
- `internal/domain/project/projecttablerelation.go`
- `internal/domain/schema/dbtable.go`
- `internal/domain/schema/dbcolumn.go`
- `internal/domain/rule/generatorconfig.go`

## Key Findings

### 1. GenerationContext sits after lifecycle, topology and row-count planning

The roadmap defines Phase 3 as an ordered chain: lifecycle -> dependency graph/topological sort -> row-count planning -> generation context -> batch generation loop -> writer adapter -> result/error model. Therefore GenerationContext must consume upstream outputs rather than re-run their algorithms.

Implications:

- `ExecutionPlan.OrderedTables` is the only source of execution order.
- `RowCountPlan.Tables` is the only source of target row counts.
- Context building should fail if topology and row-count outputs disagree.
- Later batch generation should consume context lookups and reference store APIs, not domain repositories.

### 2. Lifecycle owns execution state, not context

The lifecycle design defines execution input, precheck aggregation, state transitions and downstream ports. It explicitly excludes generation context, batch loop and writer behavior.

Implications:

- GenerationContext should expose lifecycle-compatible result fields: `Passed`, `BlockingErrors`, `Warnings` and optional context pointer.
- Context errors must use the same safe shape as lifecycle, dependency plan and rowcount issues: code, stage, field path, safe message, blocking flag.
- Context construction must not modify lifecycle state or Phase 2 execution history state enums.

### 3. Dependency plan owns sorting, not context

The dependency plan design outputs `ExecutionPlan` with `OrderedTables []PlannedTable`, each containing `ProjectTableID`, `TableID` and `ExecutionOrder`. It explicitly excludes row count and generation context.

Implications:

- GenerationContext should align table snapshots to `ExecutionPlan.OrderedTables`.
- It may build indexes and lookup maps, but must not build dependency graph nodes/edges or run topological sort.
- Boundary tests should scan for sort/graph implementation leakage in `internal/engine/gencontext`.

### 4. RowCountPlan owns target row counts, not context

The row count design outputs `RowCountPlan.Tables []PlannedRowCount`, each containing `ProjectTableID`, `TableID`, `ExecutionOrder`, `TargetRows` and `Source`. It explicitly names generation context as a downstream consumer.

Implications:

- Context must preserve row count targets and source summaries.
- Context must not reinterpret nullable `ProjectTable.RowCount` or relationship multipliers.
- Context precheck should reject plan mismatches between `ExecutionPlan` and `RowCountPlan`.

### 5. Phase 2 domain models are pure snapshots

Relevant Phase 2 models already exist as pure Go domain packages:

- `execution.GenerationJob` wraps `ExecutionTask` and table result history records.
- `project.Project` stores Project identity and connection reference.
- `project.ProjectTable` stores table-level execution configuration including nullable `RowCount`, truncate flag and execution order snapshot.
- `project.ProjectTableRelation` stores Project relation instances, multiplier snapshots and optional SQL source text.
- `schema.DbTable` and `schema.DbColumn` store schema structure and field constraints.
- `rule.GeneratorConfig` stores field-level generator selection, mapping type, params and config status.

Implications:

- GenerationContext should copy or derive read-only engine snapshots from these models.
- It should not expose mutable pointers to domain objects if later code could mutate persisted semantics.
- SQL text and rule params may contain sensitive content and cannot be reflected in safe error messages.

### 6. Runtime references are execution-local and non-persistent

The feature brief requires generated primary key, foreign key candidate and relation reference caches. Product and steering documents classify generated data content as private.

Implications:

- Runtime reference store should be scoped to one execution context instance.
- It should allow recording and querying generated references for downstream relation filling.
- It should avoid cross-task/global cache and local persistence.
- Public errors must never include raw generated values.

### 7. Minimal generator call input is allowed, full generator framework is not

Steering says Phase 3 engine can define minimal generator call interfaces, while Phase 4 owns full generator registry and Phase 5 owns built-in generators.

Implications:

- This spec can define a `GeneratorCallInput` value and read-only accessor interface.
- It must not select generator implementations, validate generator parameter schema, invoke random generators or implement a registry.
- It should be stable enough for Phase 4 to adapt without exposing engine internals.

## Design Decisions

### Decision 1: Use a dedicated engine subpackage for context

- **Decision**: Define GenerationContext under `internal/engine/gencontext`.
- **Rationale**: `context` is a common Go package name and could be confused with `context.Context`; `gencontext` is short, engine-scoped and explicit.
- **Alternatives considered**:
  - `internal/engine/context`: rejected due to name collision risk with Go standard library conventions.
  - `internal/generator/context`: rejected because the context is owned by Phase 3 engine, not Phase 4 generator framework.

### Decision 2: Context input is an aggregated immutable snapshot boundary

- **Decision**: Define `ContextInput` with task, Project, ProjectTables, DbTables, DbColumns, GeneratorConfigs, relations, ExecutionPlan and RowCountPlan.
- **Rationale**: A single input boundary makes alignment and precheck deterministic and testable.
- **Consequence**: Services are responsible for loading snapshots before calling the engine; context construction does not call repositories.

### Decision 3: Plan alignment is fail-fast and blocking

- **Decision**: Mismatches among execution task ProjectID, Project snapshot, `ExecutionPlan`, `RowCountPlan` and ProjectTable snapshots return blocking context issues.
- **Rationale**: A partial or inconsistent context would lead to unsafe generation and hard-to-debug relation errors.
- **Consequence**: No partial `GenerationContext` should be returned on blocking input errors.

### Decision 4: Expose read-only views and indexes, not mutable domain objects

- **Decision**: Context types such as `ContextTable`, `ContextField`, `ContextRelation`, `ContextRule` and `ContextRowTarget` represent engine snapshots.
- **Rationale**: Later batch loop and generator code should not mutate Project, Schema, Rule or plan objects.
- **Consequence**: Implementation may copy required fields and build maps by stable IDs.

### Decision 5: RuntimeReferenceStore stores values but sanitizes public diagnostics

- **Decision**: Define a runtime store for execution-local generated references, with public errors that never include raw values.
- **Rationale**: Relation generation needs access to generated keys, but generated data content is private.
- **Consequence**: Tests must confirm sensitive generated data examples are absent from issue messages.

### Decision 6: GeneratorCallInput is minimal and generator-agnostic

- **Decision**: Define a minimal call input containing task/table/field identity, row index, target rows, logic type, constraint summary, rule summary and reference accessor.
- **Rationale**: It enables batch loop and future generator framework integration without implementing registry or built-ins.
- **Consequence**: No generator selection, random generation, parameter schema validation or batch scheduling belongs in this spec.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Context starts duplicating topology or row-count logic | Phase 3 algorithms diverge | Boundary tests for graph/sort/rowcount solver terms and imports |
| Context exposes mutable domain pointers | Batch loop mutates persistence snapshots accidentally | Use copied read-only context structs and lookup methods |
| Reference errors leak generated values | Privacy boundary violation | Fixed safe messages and sensitive-content tests |
| Generator call input becomes a registry | Premature Phase 4 implementation | Keep call input as data-only plus read-only accessors; test absence of registry/built-in generators |
| External DB query relation source is executed too early | Violates scope and DB boundary | Store only safe source summary and future capability marker |

## Open Questions for Implementation

- Whether `GeneratorConfig.ConfigStatus` severities are already expressive enough for warning vs blocking decisions, or whether this package should define a small mapping table from status values to issue severity.
- Whether generated reference values should use `any` for early engine integration or a stricter `ReferenceValue` wrapper. The design recommends a wrapper to centralize redaction and comparison behavior.
- Whether unique-key references should be grouped by single column only at first or allow composite key groups. The design recommends a key-scope model that can represent both without implementing unique capacity planning.

## Conclusion

GenerationContext should be implemented as a dedicated Phase 3 engine package that builds immutable execution-time snapshots and an execution-local reference store from already-approved upstream plans and Phase 2 domain snapshots. It must provide lifecycle-compatible safe results, stable lookups and minimal generator call input while staying strictly out of topology sorting, row-count solving, generator registry, batch loop, writer, UI/API and real database access boundaries.
