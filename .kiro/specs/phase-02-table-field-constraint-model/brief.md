# Brief: phase-02-table-field-constraint-model

## Problem
Schema 浏览、字段规则配置和生成预检都依赖表、字段、主键、唯一性、非空、默认值等基础结构表达。当前还缺少本地业务模型来承载 `DbTable`、`DbColumn` 和基础约束。

## Current State
`docs/data-model.md` 已定义 `DbTable`、`DbColumn`、`TableConstraint`，并说明 `DbColumn.is_primary_key` 是为了高频渲染保留的冗余字段。Phase 1 的 dbx schema snapshot 已覆盖 column、primary key、unique constraint 等扫描输入。

## Desired Outcome
完成后，后端有可持久化、可验证、可序列化的表、字段和基础约束模型，能够支持字段列表、约束列表和后续字段生成规则绑定。

## Approach
基于 data-model 定义 Go 领域实体和值对象，明确表/字段的唯一性、字段顺序、原生类型、逻辑类型、默认值、非空、主键冗余标记和 PRIMARY/UNIQUE 约束表达。优先补充模型验证和快照映射测试，不实现扫描流程。

## Scope
- **In**: `DbTable`、`DbColumn`、`TableConstraint` 模型，约束类型枚举，字段顺序、原生类型、逻辑类型、默认值、非空、主键冗余标记、DDL 快照字段、基础校验和测试。
- **Out**: ForeignKey、TableRelation、字段生成规则、Project 配置、真实扫描、复杂 CHECK/INDEX 持久化。

## Boundary Candidates
- `TableConstraint` 只覆盖 PRIMARY 和 UNIQUE，FOREIGN KEY 留给 relation-model。
- `DbColumn` 可以保留高频展示冗余字段，但同步规则应在设计中说明。
- 逻辑类型应兼容 Phase 1 TypeMapper 输出，但不在本 spec 实现 TypeMapper。

## Out of Boundary
- 不实现完整数据库 DDL 解析。
- 不实现字段规则推荐。
- 不实现 UI 表格展示。

## Upstream / Downstream
- **Upstream**: `phase-02-database-schema-model`。
- **Downstream**: `phase-02-relation-model`、`phase-02-field-generation-rule-model`、Phase 4/5 生成器、Phase 7 Schema API。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 1 database-dialect-interface 的 schema snapshot 和 logical type。

## Constraints
模型应覆盖 MySQL/PostgreSQL 主流字段信息，同时保留原始类型字符串。导出 Go 结构、字段、常量和枚举值必须有注释。
