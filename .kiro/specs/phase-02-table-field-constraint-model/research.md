# Research Log

## Summary

`phase-02-table-field-constraint-model` 属于 Phase 2 领域模型批次，采用轻量设计发现。核心结论是：先稳定 `DbTable`、`DbColumn`、`TableConstraint`、`ColumnLogicalType`、字段级校验和 JSON 序列化合同，不提前实现后续服务、API、UI、执行引擎或真实数据库扫描。

本次设计修订重点补齐了三个 implementation readiness 缺口：完整字段矩阵、与现有 `internal/dbx/schema` 的边界、以及可测试的基础校验合同。

## Research Log

### 主题：依赖边界

- **来源**: `.kiro/steering/product.md`、`.kiro/steering/tech.md`、`.kiro/steering/structure.md` 与上游规格 `phase-02-database-schema-model`。
- **发现**: 当前规格只能依赖上游 `DbSchema` 的稳定身份语义，不能反向吸收上游职责，也不能直接依赖 Wails、Vue、store、service、engine 或真实数据库驱动。
- **影响**: 设计将 `internal/domain/schema` 定义为纯领域模型包，校验函数保持纯内存、无外部副作用。

### 主题：现有 dbx 扫描快照模型

- **来源**: `internal/dbx/schema/model.go`、`internal/dbx/schema/constraints.go`、`internal/dbx/schema/logical_type.go`。
- **发现**: 代码库已有 `Database`、`Namespace`、`Table`、`Column`、`PrimaryKey`、`ForeignKey`、`UniqueConstraint`、`CheckConstraint`、`Index` 和 `LogicalType`，用于数据库适配与 introspection 层，保留原始元数据和方言差异。
- **影响**: 新增 `internal/domain/schema` 不直接 import `internal/dbx/schema`。后续映射规格负责从扫描快照转换为领域模型，避免 adapter 模型泄漏到 domain。

### 主题：数据模型来源

- **来源**: `docs/data-model.md` §5 Schema 结构、§6 Schema 约束、§11.2、§12 D-04。
- **发现**: `DbTable`、`DbColumn`、`TableConstraint` 已有持久化字段来源；`TableConstraint` 仅覆盖 `PRIMARY` 和 `UNIQUE`；`DbColumn.is_primary_key` 是从约束派生的冗余字段。
- **影响**: 设计补充字段矩阵，并明确 `IsPrimaryKey` 只定义字段，不实现同步流程；`ColumnIDs` 在 JSON 中使用数组，未来 repository 再映射到持久化层的逗号分隔字段。

### 主题：下游合同

- **来源**: requirements.md 1.1-5.5 与 Phase 2 后续字段规则、关系模型、Schema 缓存消费需求。
- **发现**: 下游需要稳定字段名、状态枚举、错误结构和 JSON 往返行为，不能由实现阶段临时解释。
- **影响**: 设计加入 Requirements Traceability、Stable Enums、Validation Matrix 和 JSON Serialization Contract，确保任务生成和实现有明确依据。

### 主题：范围控制

- **来源**: steering 的 Phase discipline 与产品边界。
- **发现**: Phase 2 可以定义模型、枚举、验证和序列化，但不实现未来 API/UI/执行能力，也不提前实现 ForeignKey、TableRelation、字段规则或 Project。
- **影响**: 所有 out-of-scope 内容写入设计边界、集成说明和 revalidation triggers。

## Synthesis Outcomes

- 采用 Go domain 包承载模型和值对象，包路径为 `internal/domain/schema`。
- 将 `internal/dbx/schema` 定位为扫描快照模型，将 `internal/domain/schema` 定位为稳定领域合同，两者通过后续映射规格衔接。
- 使用稳定字符串枚举保护持久化、Wails 合同和前端消费，不使用整数 ordinal。
- 使用字段级错误集合表达基础校验问题，issue JSON 形状与现有 `ConfigIssue` / `ApiIssue` 兼容但不形成依赖。
- 只校验单对象内部可判断的问题；需要聚合状态或数据库状态的唯一性、引用归属和同步流程延后到 store/service 规格。
