# Brief: phase-01-database-dialect-interface

## Problem

LoomiDBX 需要连接和操作多种数据库。不同数据库在元数据查询、类型系统、SQL 方言、事务能力、外键能力和批量写入方式上差异明显。如果业务层直接拼接 SQL 或按数据库类型写分支，后续扩展会变得困难。

## Current State

`docs/database-dialect-abstraction-design.md` 已提出 Adapter、Dialect、Introspector、TypeMapper、Capabilities 的抽象方向，并明确首期优先 MySQL 和 PostgreSQL。但项目尚未建立代码级接口和测试替身。

## Desired Outcome

完成后，Go 后端拥有数据库方言抽象的最小接口集合，业务层可以依赖统一接口进行连接、能力查询、Schema introspection 和写入计划构建的后续扩展。首批至少应有 mock/fake adapter 支持测试，不要求实现完整数据库驱动。

## Approach

在 Go 后端 adapter/db 或类似目录中定义 Adapter、Dialect、Introspector、TypeMapper、Capabilities 等接口和值对象。用能力模型表达数据库差异，保留原始类型和元数据字段，为后续 MySQL/PostgreSQL 实现留出扩展点。

## Scope

- **In**:
  - 定义 Adapter、Dialect、Introspector、TypeMapper、Capabilities 的最小接口。
  - 定义连接配置、连接测试、Schema 扫描结果和类型映射的接口边界。
  - 定义 SQL 标识符引用、占位符、批量插入等 Dialect 能力的接口雏形。
  - 提供 mock/fake 实现用于测试和后续服务层开发。
  - 记录首期 MySQL/PostgreSQL 优先、其他数据库预留的扩展约定。
- **Out**:
  - 不实现完整 MySQL/PostgreSQL introspection。
  - 不实现真实批量写入器。
  - 不实现执行引擎。
  - 不实现 UI 连接页面。
  - 不实现所有数据库能力差异。

## Boundary Candidates

- Adapter 是数据库类型的统一入口。
- Introspector 只负责元数据扫描到统一 Schema 的转换。
- Dialect 只负责 SQL 语法差异。
- TypeMapper 负责 Native Type 到 Logical Type 的映射。
- Capabilities 作为运行时能力协商依据，业务层避免硬编码数据库类型。

## Out of Boundary

- 完整 canonical schema 模型。
- 真实数据库连接管理 UI。
- 数据生成和写入执行流程。
- 复杂 ORM 或跨数据库迁移工具。

## Upstream / Downstream

- **Upstream**: `phase-01-project-structure`、`docs/database-dialect-abstraction-design.md`、`docs/agent/01-architecture-bootstrap.md`。
- **Downstream**: `database-schema-model`、`schema-introspection-api`、`batch-writer-adapter`、`engine-integration-tests`。

## Existing Spec Touchpoints

- **Extends**: 无。
- **Adjacent**: `phase-01-local-storage-strategy`、`connection-model`、`database-schema-model`。

## Constraints

首期不追求抹平所有数据库差异。业务层应通过 capabilities 做决策，不应写 `if dbType == ...` 的硬编码主路径。必须保留 NativeType、LogicalType 和 Raw 元数据的扩展空间。
