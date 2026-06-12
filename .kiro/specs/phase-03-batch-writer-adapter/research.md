# Research Document

## Research Scope

本研究为 `phase-03-batch-writer-adapter` 的设计阶段提供上下文发现和边界决策。输入包括 feature brief、roadmap、steering 文档、上游 Phase 3 规格和 Phase 1 DBX 方言接口规格。

## Reviewed Sources

- `.kiro/specs/phase-03-batch-writer-adapter/brief.md`
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
- `.kiro/specs/phase-03-generation-context/requirements.md`
- `.kiro/specs/phase-03-generation-context/design.md`
- `.kiro/specs/phase-03-batch-generation-loop/requirements.md`
- `.kiro/specs/phase-03-batch-generation-loop/design.md`
- `.kiro/specs/phase-01-database-dialect-interface/requirements.md`
- `.kiro/specs/phase-01-database-dialect-interface/design.md`

## Key Findings

### 1. Writer adapter is downstream of batch generation, not part of row assembly

`phase-03-batch-generation-loop` owns table ordering, batch slicing, row assembly, relation reference priority, generator invocation, writer seam invocation and runtime reference commit after writer success. It explicitly does not open database connections, execute SQL, implement transactions, clear tables, retry writes or interpret driver errors.

Implication: writer adapter must implement the concrete engine-side writer behind the `BatchWriter` seam, but must not move row assembly or reference recording logic out of batch loop.

### 2. Batch payload already defines the minimum writer input

The batch loop design defines the conceptual writer seam:

```go
type BatchWriter interface {
    WriteBatch(payload BatchPayload) (BatchWriteResult, error)
}

type BatchPayload struct {
    TaskID int64
    ProjectID int64
    ProjectTableID int64
    TableID int64
    Range BatchRange
    Columns []int64
    Rows []GeneratedRow
}

type FieldValueState string

const (
    FieldValuePresent FieldValueState = "present"
    FieldValueNull FieldValueState = "null"
    FieldValueOmittedDefault FieldValueState = "omitted_default"
)
```

Implication: writer adapter should preserve this shape and refine write-request/result/error models around it, not introduce a second unrelated row representation.

### 3. Database differences must stay in DBX Adapter/Dialect/Capabilities

Phase 1 DBX design establishes:

```go
type Adapter interface {
    Type() DBType
    DisplayName() string
    Capabilities() capability.Capabilities
    TestConnection(ctx context.Context, cfg ConnectionConfig) ConnectionTestResult
    Dialect() dialect.Dialect
    Introspector() introspect.Introspector
    TypeMapper() typex.Mapper
}

type Dialect interface {
    QuoteIdentifier(name string) string
    Placeholder(index int) string
    BuildInsert(req InsertRequest) ([]Statement, error)
}
```

Capabilities cover transaction, savepoint, foreign key, deferred constraint, batch insert, bulk load, returning, upsert, catalog/schema, JSON, array, UUID, enum, generated column, identity column, identifier length and limit-style constraints.

Implication: writer adapter may ask capabilities whether a strategy is allowed and may call dialect to build statements, but must not branch by database product name or hand-code MySQL/PostgreSQL SQL.

### 4. Lifecycle only consumes safe stage results

Lifecycle defines state machine, precheck aggregation and downstream ports. It maps downstream failures to lifecycle-safe errors and explicitly excludes writer adapter, transaction and real database writes.

Implication: writer adapter result must contain success stats and safe blocking errors that lifecycle/result specs can aggregate without needing raw database errors.

### 5. Clear strategy is a Project/write seam concern, not a SQL implementation in engine

The brief identifies table-level clear strategy and write ordering as in scope, while excluding full dialect SQL and connection management. Steering states database differences must remain under Adapter/Dialect/Capabilities.

Implication: the spec should model clear-before-first-write and clear seam invocation order, plus capability checks for FK/transaction implications, but should not implement TRUNCATE/DELETE SQL or FK toggles in engine.

### 6. Sensitive data boundaries are strict

Steering and upstream specs repeatedly prohibit leaking database connection info, user SQL, generation rules, generated data and raw downstream errors. Batch loop also forbids exposing raw generated/reference values in public errors.

Implication: writer adapter must sanitize raw dialect/executor/transaction/clear errors and public error fields must use safe identifiers such as task, project table, table, column, batch, row and statement index.

## Design Decisions

### Decision 1: Place writer adapter under `internal/engine/writer`

- **Chosen**: create an engine writer package that implements the writer seam and depends on DBX interfaces.
- **Rationale**: steering says engine owns batch writing and result model, while adapter owns external database differences. `internal/engine/writer` can coordinate capabilities, dialect request construction, clear seam, transaction seam and safe errors without becoming a real DB adapter.
- **Rejected**: placing this in `internal/dbx` would mix engine strategy and execution statistics into database abstraction; placing it in `internal/engine/batch` would pollute batch generation loop with SQL/transaction concerns.

### Decision 2: Introduce executor, transaction and clear seams rather than real DB calls

- **Chosen**: define narrow interfaces for statement execution, transaction boundaries and table clearing.
- **Rationale**: current spec must be testable without real databases and must not own connection management or driver imports.
- **Rejected**: direct `database/sql` usage, because connection lifecycle and true transaction behavior are out of scope.

### Decision 3: Capability-first strategy enforcement

- **Chosen**: derive transaction/batch/parameter limit decisions from `Capabilities` and configured writer strategy.
- **Rationale**: roadmap and Phase 1 require database differences to be expressed by capabilities.
- **Rejected**: `switch adapter.Type()` or product-name branches.

### Decision 4: Preserve field value states through write mapping

- **Chosen**: keep present/null/omitted-default distinct when building insert requests.
- **Rationale**: batch loop Requirement 6 requires writer seam to distinguish omitted fields, nulls and defaults.
- **Rejected**: converting all absent values to nil or Go zero values, because it would destroy database default semantics.

### Decision 5: Public errors are writer-safe summaries only

- **Chosen**: expose `WriterIssue` with code, stage, field path, safe message, blocking flag and safe scope.
- **Rationale**: aligns with lifecycle, rowcount, generation context and batch loop safe error models.
- **Rejected**: forwarding raw dialect or driver errors.

## Open Questions Deferred to Later Specs

- Exact real MySQL/PostgreSQL insert SQL generation and batching details remain in DBX adapter/dialect implementation specs.
- Real connection pool, credential loading and transaction object lifetimes remain in connection/service specs.
- Final historical persistence shape remains in `phase-03-execution-result-and-error-model` and later API/UI specs.
- Advanced retry, resume and idempotency semantics remain out of Phase 3 writer adapter scope.

## Risk Notes

- Omitted/default values may produce heterogeneous row shapes. The design must allow deterministic statement splitting or safe blocking if the dialect request cannot express them.
- Non-transactional partial success must be surfaced explicitly, because writer adapter must not invent rollback or retry behavior.
- Capabilities may not initially include every limit needed by real adapters; implementation should keep validation extensible and fail safely when limits are unknown.
