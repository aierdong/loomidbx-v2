# Brief: phase-02-project-model

## Problem
用户需要把多张表组织成一次可复用的生成任务，配置每张表的行数、清空策略、执行顺序和关系值来源。当前缺少 Project 层领域模型来承接这些执行配置。

## Current State
`docs/data-model.md` 定义了 `Project`、`ProjectTable`、`ProjectTableRelation`，并明确字段规则归属 Schema 层，表级行数、清空策略和执行选项归属 Project 层。`execution_order` 设计为保存时预计算快照。

## Desired Outcome
完成后，系统具备 Project 任务组织模型，能表达 Project 基本信息、参与表、表级行数/清空策略/执行顺序、运行时关系实例化和值来源策略，为 Phase 3 执行引擎和 Phase 7 Project API 提供输入。

## Approach
先定义领域实体、枚举和值对象，并把复杂逻辑以可测试验证规则表达出来。行数状态机和关系实例化边界在模型/设计中明确，但拓扑排序和真实行数计算留给 Phase 3。

## Scope
- **In**: `Project`、`ProjectTable`、`ProjectTableRelation` 模型，`rel_value_source` 枚举，行数字段约束表达，清空策略，执行顺序快照，关系实例化配置，基础验证和测试。
- **Out**: 拓扑排序算法、执行计划构建、生成引擎、写入数据库、Project API、UI Project 页面。

## Boundary Candidates
- Project 只保存任务组织和执行配置，不保存字段生成规则副本。
- `execution_order` 是模型字段和契约，计算实现可留到 Phase 3 或服务层。
- `ProjectTableRelation` 是从 Schema 层 `TableRelation` 实例化的运行时配置。

## Out of Boundary
- 不实现启动任务。
- 不实现预览执行顺序服务。
- 不处理执行历史写入。

## Upstream / Downstream
- **Upstream**: `phase-02-relation-model`、`phase-02-field-generation-rule-model`。
- **Downstream**: `phase-02-generation-job-model`、Phase 3 generation-engine、Phase 7 project-api、Phase 8 projects-page。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 3 dependency-graph-and-topological-sort、row-count-planning。

## Constraints
必须保持字段规则与 Project 执行规则分离。Project 模型应能表达 Parent/Child、BaseTable/JoinTable 的行数状态机，但不要提前实现执行引擎。
