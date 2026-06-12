# Brief: phase-03-batch-writer-adapter

## Problem

生成出的批次数据最终需要写回目标数据库，但 engine 不应直接耦合具体数据库驱动或方言分支。没有批量写入适配边界时，写入差异、事务策略、清空策略和错误处理会污染生成循环。

## Current State

Phase 1 已定义数据库 Adapter、Dialect、Capabilities 等抽象边界，Phase 2 已定义 Project 清空策略和执行任务模型。Phase 3 尚未定义 engine 侧如何把生成批次交给目标数据库写入，以及如何汇总写入结果。

## Desired Outcome

完成后，engine 具备批量写入适配接口，可以接收表级批次数据、执行清空/写入策略、返回成功计数和数据库错误。该接口应支持 mock 实现，用于 engine 测试，并为 MySQL/PostgreSQL 适配保留能力扩展点。

## Approach

在 engine 和 dbx adapter 之间建立 writer adapter seam。该 spec 定义 engine 侧写入请求/响应、事务能力判断、清空策略调用顺序和错误归一化，不把具体 SQL 拼接或数据库驱动逻辑写入 engine 主循环。

## Scope

- **In**:
  - BatchWriter 接口、写入请求、写入结果和错误边界。
  - 表级清空策略与批量写入顺序的 engine 侧表达。
  - 基于 Capabilities 的事务/批量写入能力判断边界。
  - mock writer 和最小单元测试支持。
- **Out**:
  - 不完整实现所有数据库方言写入 SQL。
  - 不实现跨数据库迁移或同步。
  - 不实现 UI/API 侧的执行历史查询。

## Boundary Candidates

- Batch writer interface：负责 engine 与数据库适配层之间的契约。
- Write strategy model：负责清空、事务和批量大小策略。
- Write result mapper：负责将数据库错误归一化为 engine 错误。

## Out of Boundary

- 数据库连接管理和凭据存储。
- 高级重试、断点续跑和幂等写入策略。
- 复杂可观测性和日志管道。

## Upstream / Downstream

- **Upstream**: phase-03-batch-generation-loop、phase-01-database-dialect-interface、phase-02-project-model。
- **Downstream**: phase-03-execution-result-and-error-model、phase-07-generation-job-api、phase-09-engine-integration-tests。

## Existing Spec Touchpoints

- **Extends**: phase-01-database-dialect-interface 的数据库能力抽象在写入场景中的使用边界。
- **Adjacent**: phase-02-project-model 定义清空策略；phase-03-batch-generation-loop 产生写入批次。

## Constraints

数据库差异必须下沉到 Adapter、Dialect 和 Capabilities，engine 不应按数据库类型写硬编码分支。错误信息应能定位失败表、批次和字段，但不得泄露敏感连接信息。
