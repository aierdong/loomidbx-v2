# Research & Design Decisions

## Summary
- **Feature**: `phase-01-database-dialect-interface`
- **Discovery Scope**: Extension
- **Key Findings**:
  - `phase-01-project-structure` 已落地 `internal/dbx/` 及 adapter、dialect、introspect、typex、capability 子目录，本 spec 应扩展这些位置，不建立竞争性目录。
  - `docs/database-dialect-abstraction-design.md` 已给出 Adapter、Dialect、Introspector、TypeMapper、Capabilities、Canonical Schema 和 InsertPlan 的长期方向；本 spec 只落最小接口和值对象。
  - 首期需要 fake/mock 支持以服务后续服务层和测试，不要求真实 MySQL/PostgreSQL 驱动或真实数据库连接。

## Research Log

### 上游工程骨架对齐
- **Context**: 本 spec 依赖 `phase-01-project-structure`，需要确认可写入的后端落位。
- **Sources Consulted**: `.kiro/specs/phase-01-project-structure/design.md`、`requirements.md`、`tasks.md`、`internal/dbx/*/README.md`。
- **Findings**:
  - 工程骨架明确把 DBX 能力放在 `internal/dbx/` 下。
  - 现有子目录为 `adapter`、`dialect`、`introspect`、`typex`、`capability`，当前均为占位说明。
  - 骨架 spec 的非目标明确不实现真实数据库方言行为。
- **Implications**:
  - 本 spec 的文件结构计划应直接复用这些目录。
  - 新增 schema、core、fakes 等目录需要说明与现有目录的关系，避免破坏依赖方向。

### 数据库方言长期设计
- **Context**: brief 指向 `docs/database-dialect-abstraction-design.md`，需要提炼当前阶段可落地范围。
- **Sources Consulted**: `docs/database-dialect-abstraction-design.md`、`.kiro/steering/tech.md`、`.kiro/steering/structure.md`。
- **Findings**:
  - 长期目标采用统一模型、方言适配、能力协商。
  - 建议 Go 包包含 core、schema、dialect、introspect、typex、plan、writer、adapters 等层次。
  - 首期优先 MySQL/PostgreSQL，但文档中的真实 introspection、writer、依赖排序和集成测试属于后续阶段。
- **Implications**:
  - 当前 spec 设计 Adapter、Capabilities、ConnectionConfig、Schema、LogicalType、Dialect、TypeMapper 和 registry。
  - InsertPlan 只在 Dialect 的 batch insert request 边界中保留最小形态，不建立完整 writer 或执行引擎。

### 测试替身需求
- **Context**: Desired Outcome 要求首批至少有 mock/fake adapter 支持测试。
- **Sources Consulted**: brief、`.kiro/steering/tech.md` 的测试方向、`phase-01-project-structure` 的测试边界。
- **Findings**:
  - Phase 1 需要单元测试和 mock 测试先行，完整测试工具链由 `phase-01-test-tooling` 承担。
  - 测试替身不能伪装为真实数据库支持。
- **Implications**:
  - 设计中新增 `internal/dbx/fakes/`，提供可配置 fake adapter、dialect、introspector、mapper 和 registry 测试支持。
  - 测试覆盖应验证能力协商、错误类型、Schema 原始元数据保留和 fake 的 deterministic 行为。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 最小接口分层 | 在 `internal/dbx/` 下定义 core/schema/dialect/introspect/typex/capability/fakes | 对齐骨架，边界清晰，便于后续并行实现 | 需要避免一次性定义过多未来接口 | 选中 |
| 单包聚合 | 把所有接口和值对象放入 `internal/dbx` 一个 Go 包 | 初期文件少 | 后续类型、方言和扫描实现容易互相污染 | 放弃 |
| 直接实现 MySQL/PostgreSQL adapter | 立刻验证真实数据库差异 | 更接近端到端能力 | 超出 brief，依赖驱动和集成环境，扩大范围 | 放弃 |
| 引入 ORM 或 SQL builder | 借助外部库处理 SQL 差异 | 可减少 SQL 拼装工作 | 当前只需要接口边界，过早绑定依赖 | 放弃 |

## Design Decisions

### Decision: 采用 DBX 内部分层接口
- **Context**: 需要在不实现真实驱动的情况下固定数据库差异边界。
- **Alternatives Considered**:
  1. 单一 `dbx.Adapter` 包含所有类型。
  2. 按能力拆分 `core`、`schema`、`dialect`、`introspect`、`typex`、`capability`。
- **Selected Approach**: 在 `internal/dbx/` 下建立按能力分层的 Go 包，并让 adapter 聚合其他能力接口。
- **Rationale**: 匹配 steering 和长期设计，同时让后续 schema、writer、API spec 可以只依赖需要的子包。
- **Trade-offs**: 文件数量更多，但每个文件职责更窄。
- **Follow-up**: 实现时需检查 import 方向，避免 schema 依赖 adapter 或具体数据库实现。

### Decision: 当前只提供 fake，不提供真实数据库 adapter
- **Context**: brief 明确不要求完整数据库驱动，但需要支持测试。
- **Alternatives Considered**:
  1. 提供 MySQL/PostgreSQL 空壳 adapter。
  2. 提供 fake adapter 和可配置测试夹具。
- **Selected Approach**: 只提供 `fakes` 测试替身，并在文档中声明 MySQL/PostgreSQL 为优先验证目标但未完成。
- **Rationale**: 避免把不可用的生产 adapter 误认为已支持。
- **Trade-offs**: 后续真实 adapter spec 需要新增具体数据库子包。
- **Follow-up**: 后续 MySQL/PostgreSQL spec 应复用本接口并补充真实能力矩阵测试。

### Decision: Schema 模型保留 Raw 元数据
- **Context**: 不同数据库的元数据完整度差异明显，当前模型无法覆盖全部高级特性。
- **Alternatives Considered**:
  1. 只保留标准化字段。
  2. 标准化字段加 `Raw` 扩展。
- **Selected Approach**: Database、Namespace、Table、Column、Constraint、Index 和 View 均允许保留 raw metadata。
- **Rationale**: 避免首期抽象丢失排错信息，同时不急于支持所有高级能力。
- **Trade-offs**: Raw 需要避免被业务主路径滥用。
- **Follow-up**: 后续 Schema API 应决定 Raw 是否暴露给前端或只用于诊断。

## Risks & Mitigations

- 接口过大导致首期实现拖慢 — 只包含 brief 明确要求的最小能力，不引入 writer、执行引擎或真实 adapter。
- fake 被误用为生产 adapter — 使用 `fakes` 包名、README 和测试命名明确 test-only。
- 业务层继续按数据库类型硬编码 — 能力模型、registry 和测试任务要求验证 capability-first 路径。
- Raw 元数据被滥用为主路径逻辑 — 设计中声明 Raw 用于诊断和扩展，标准字段仍是主路径。

## References

- `.kiro/steering/product.md` — 产品边界与隐私原则。
- `.kiro/steering/tech.md` — 数据库兼容策略和分层后端原则。
- `.kiro/steering/structure.md` — `internal/dbx/` 目录和依赖方向。
- `.kiro/specs/phase-01-project-structure/design.md` — 上游工程骨架和 DBX 占位目录。
- `docs/database-dialect-abstraction-design.md` — 多数据库方言抽象长期设计。
