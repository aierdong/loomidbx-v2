# Brief: phase-02-relation-model

## Problem
生成数据必须遵守外键和表间关系。系统需要同时表达数据库扫描出的物理外键，以及应用内定义或修正后的逻辑关系，为后续拓扑排序、行数规划和外键赋值提供基础。

## Current State
`docs/data-model.md` 定义了 `ForeignKey` 与 `TableRelation`，并统一 Parent/Child、BaseTable/JoinTable 术语。Phase 1 的 `internal/dbx/schema.ForeignKey` 只能表示扫描快照，尚未形成本地关系模型。

## Desired Outcome
完成后，系统有可持久化的外键和表关系领域模型，能够表达来源/目标字段、关系类型、倍数范围、物理/逻辑关系标记，并为 Project 关系实例化提供稳定输入。

## Approach
把物理外键和逻辑关系拆开建模：`ForeignKey` 保存扫描事实，`TableRelation` 保存生成语义关系。N:N 关系按文档拆成多条 `JOIN_TABLE` 关系，统一作为下游依赖处理。

## Scope
- **In**: `ForeignKey`、`TableRelation` 模型，`PARENT_CHILD` / `JOIN_TABLE` 枚举，来源/目标字段 ID 列表表达，倍数范围校验，物理/逻辑关系标记，基础测试。
- **Out**: 拓扑排序实现、行数规划、执行时外键取值、JoinTable 容量自动降级、关系编辑 UI。

## Boundary Candidates
- `ForeignKey` 是数据库事实；`TableRelation` 是生成语义。
- N:N 不做复杂复合宽表，按多条扁平关系表达。
- 倍数范围属于 Schema 层默认关系配置，Project 可实例化但不在本 spec 做 Project 模型。

## Out of Boundary
- 不实现依赖图算法。
- 不查询数据库存量数据。
- 不做结构变更影响分析。

## Upstream / Downstream
- **Upstream**: `phase-02-table-field-constraint-model`。
- **Downstream**: `phase-02-field-generation-rule-model`、`phase-02-project-model`、Phase 3 topology、Phase 5 relation-generators。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 1 database-dialect-interface 中的 foreign key snapshot。

## Constraints
关系术语必须使用 Parent、Child、BaseTable、JoinTable，不使用含混口语化名称。约束优先，无法表达的关系应保留可验证错误，而不是静默忽略。
