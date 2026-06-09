# Brief: phase-02-generation-job-model

## Problem
执行生成任务需要记录任务配置快照、状态、结果和失败信息。Phase 2 需要先定义执行任务和历史结果的领域表达，让 Phase 3 引擎和 Phase 7 服务层有稳定契约。

## Current State
`docs/data-model.md` 定义 `ExecutionTask` 和 `ExecutionTableResult`，包含任务状态、表级状态、执行时间、写入行数、错误信息和表名/Schema 名快照。`docs/phase.md` 建议 Phase 2 拆分为 `generation-job-model`，但不实现执行引擎。

## Desired Outcome
完成后，系统有生成任务/执行历史的模型和状态枚举，能表达任务运行中、成功、部分失败、失败，以及表级 pending/running/success/failed/skipped 等结果，为后续进度事件、历史查询和错误报告提供基础。

## Approach
将本 spec 定位为“执行记录与结果模型”，不实现生成生命周期。定义 `GenerationJob` 命名兼容层或明确与 `ExecutionTask` 的命名关系，按 data-model 落地任务主记录、表结果、状态机枚举、快照字段和基础验证。

## Scope
- **In**: `ExecutionTask` / `GenerationJob` 命名边界，`ExecutionTableResult` 模型，任务状态枚举，表级状态枚举，时间字段、写入行数、错误信息、表名/Schema 快照，基础验证和测试。
- **Out**: 执行引擎生命周期、进度事件发布、批处理写入、事务/回滚策略、API、UI 进度视图。

## Boundary Candidates
- 本 spec 只定义“可记录什么”，不定义“如何执行”。
- 历史记录必须保留名称快照，不能依赖 DbTable 永远存在。
- 状态枚举需要匹配未来执行引擎和 API 错误响应。

## Out of Boundary
- 不启动任务。
- 不写入目标数据库。
- 不实现执行历史查询 API。

## Upstream / Downstream
- **Upstream**: `phase-02-project-model`。
- **Downstream**: Phase 3 execution-lifecycle、execution-result-and-error-model，Phase 7 generation-job-api、execution-history-api，Phase 9 observability/acceptance。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 3 execution-lifecycle、Phase 7 generation-job-api。

## Constraints
失败信息应可追踪但不能泄露敏感数据。状态枚举应稳定、语义清晰，并避免把 Phase 3 执行算法提前塞进模型层。
