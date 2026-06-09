# Brief: phase-02-database-schema-model

## Problem
用户需要浏览连接下的数据库、Schema 和表结构。Phase 1 已有 `internal/dbx/schema` 扫描快照模型，但还没有本地持久化用的 `DbCatalog`、`DbSchema` 领域模型，也没有明确快照模型到业务模型的映射边界。

## Current State
`internal/dbx/schema` 已表达 introspection 返回的 canonical database/namespace/table snapshot。`docs/data-model.md` 定义了 `DbCatalog` 和 `DbSchema` 的持久化结构，以及 MySQL 无 Schema 时使用空字符串隐式 Schema 的规则。

## Desired Outcome
完成后，系统具备连接下数据库层级和 Schema 层级的领域表达，可以稳定承接扫描结果、记录扫描时间，并为表/字段模型提供父级容器。

## Approach
把 `internal/dbx/schema` 视为扫描输入快照，Phase 2 模型定义本地业务持久化表达。优先实现 `DbCatalog`、`DbSchema`、标识符/名称规则、扫描时间和唯一性约束表达，并补充映射/验证测试。

## Scope
- **In**: `DbCatalog`、`DbSchema` 领域模型、Catalog/Schema 名称规则、隐式空 Schema 约定、扫描时间字段、与连接模型的关联、基础验证和序列化测试。
- **Out**: 表、字段、约束细节；真实 introspection 扫描；Schema 重扫差异计算；API 和 UI 树形展示。

## Boundary Candidates
- Introspection snapshot 与本地业务模型分离。
- Catalog/Schema 模型只表达层级容器，不承载字段规则或 Project 配置。
- 无 Schema 数据库统一映射为空字符串 Schema，避免上层按数据库类型分支。

## Out of Boundary
- 不实现扫描服务或数据库访问。
- 不实现结构变更影响分析。
- 不实现数据库对象树 UI。

## Upstream / Downstream
- **Upstream**: `phase-02-connection-model`、`phase-01-database-dialect-interface`。
- **Downstream**: `phase-02-table-field-constraint-model`、Phase 7 schema-introspection-api、Phase 8 schema-management-page。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 1 database-dialect-interface 中的 `internal/dbx/schema`。

## Constraints
必须保留数据库原始层级信息和扫描时间，避免丢失后续排错能力。不要把 `internal/dbx/schema` 快照模型直接当作持久化业务模型。
