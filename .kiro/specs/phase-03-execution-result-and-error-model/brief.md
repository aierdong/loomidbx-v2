# Brief: phase-03-execution-result-and-error-model

## Problem

执行完成或失败后，用户和后续 API/UI 需要知道哪些表成功、哪些批次失败、失败原因是什么、影响范围如何。没有统一结果和错误模型时，生命周期、生成循环、写入适配和历史记录会各自表达错误，难以复盘和重试。

## Current State

Phase 2 已定义 GenerationJob / ExecutionTask / ExecutionTableResult 等领域模型，但 Phase 3 的运行时错误分类、表级结果汇总、批次失败范围和最小历史记录边界尚未落地。

## Desired Outcome

完成后，engine 能够输出统一的 ExecutionResult：包含任务状态、表级结果、行数统计、批次统计、错误分类、失败范围和诊断信息，并能映射回 Phase 2 执行历史模型供后续服务/API/UI 使用。

## Approach

在 Phase 3 收尾处统一运行时结果和错误模型。该 spec 汇总生命周期、生成循环和 writer adapter 的结果，定义错误分类与安全消息边界，保持对后续 Phase 7/8/9 的契约友好但不提前实现完整 API 或可观测性。

## Scope

- **In**:
  - ExecutionResult、TableExecutionResult、BatchResult 和 EngineError 的运行时表达。
  - 错误分类：预检错误、计划错误、生成错误、关系引用错误、写入错误、取消错误。
  - 失败范围：任务级、表级、批次级、字段级的最小表达。
  - 结果到 Phase 2 执行历史模型的映射边界。
- **Out**:
  - 不实现完整执行历史 API、UI 进度页或日志查询。
  - 不实现复杂 tracing、metrics 或远端遥测。
  - 不保存生成数据内容作为历史记录。

## Boundary Candidates

- Runtime execution result：负责 engine 内部汇总。
- Engine error taxonomy：负责错误分类和安全消息。
- History mapper：负责与 Phase 2 执行历史模型对接。

## Out of Boundary

- 用户可视化错误报告页面。
- 自动重试和失败恢复策略。
- 远端日志上传或遥测。

## Upstream / Downstream

- **Upstream**: phase-03-batch-writer-adapter、phase-03-batch-generation-loop、phase-03-execution-lifecycle。
- **Downstream**: phase-07-execution-history-api、phase-07-error-response-contract、phase-08-generation-job-progress-view、phase-09-error-reporting-and-logging。

## Existing Spec Touchpoints

- **Extends**: phase-02-generation-job-model 的执行历史表达在运行时结果中的映射。
- **Adjacent**: phase-03-execution-lifecycle 负责状态流转；该 spec 负责结果和错误汇总。

## Constraints

失败要可追踪，但错误信息不能泄露数据库密码、令牌、用户 SQL 或生成数据内容。Phase 3 只保留最小历史和诊断边界，复杂可观测性留给 Phase 9。
