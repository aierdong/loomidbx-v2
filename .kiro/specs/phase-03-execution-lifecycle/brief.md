# Brief: phase-03-execution-lifecycle

## Problem

执行引擎需要一个清晰的入口和生命周期，才能把 Phase 2 中的 Project、GenerationJob 与 ExecutionTask 模型转化为可运行的生成流程。没有生命周期边界时，预检、启动、运行、取消、失败和完成状态容易分散到 service、binding 或未来 UI 中。

## Current State

Phase 2 已定义 GenerationJob / ExecutionTask / ExecutionTableResult 等任务配置与执行历史领域模型，但尚未实现 engine 侧的任务启动、状态流转、预检入口或执行控制。当前项目还没有真正从 Project 构建执行流程的 engine 编排层。

## Desired Outcome

完成后，后端应具备执行引擎的最小生命周期抽象：可以接收执行任务快照，完成预检入口、状态初始化、运行/失败/取消/完成流转，并为后续依赖排序、行数规划、生成上下文、批量生成和结果汇总阶段提供稳定入口。

## Approach

采用 engine 包内的生命周期协调器作为 Phase 3 的起点。该 spec 只定义执行入口、状态机、预检结果和生命周期事件边界，不提前实现完整生成器注册表、UI 进度事件或真实数据库写入。

## Scope

- **In**:
  - ExecutionTask / GenerationJob 快照到 engine 执行输入的映射边界。
  - 执行状态机和允许的状态流转。
  - 执行预检入口、启动、取消、失败、完成的最小抽象。
  - 生命周期层与后续计划、生成、结果汇总组件的接口边界。
- **Out**:
  - 不实现依赖图算法、行数计算、批量生成循环或真实写入逻辑。
  - 不实现 UI 进度页面、Wails runtime events 或 API 层。
  - 不实现 Phase 4/5 的完整生成器框架和内置生成器。

## Boundary Candidates

- Engine lifecycle coordinator：负责执行入口和状态流转。
- Execution precheck model：负责表达是否可以进入运行阶段。
- Execution control interface：为取消和失败传播保留边界。

## Out of Boundary

- 生成器参数 Schema、生成器注册表、内置生成器实现。
- 前端进度展示和服务层 API 契约。
- 复杂日志、追踪和可观测性。

## Upstream / Downstream

- **Upstream**: phase-02-generation-job-model、phase-02-project-model。
- **Downstream**: phase-03-dependency-graph-and-topological-sort、phase-03-row-count-planning、phase-03-batch-generation-loop、phase-07-generation-job-api、phase-08-generation-job-progress-view。

## Existing Spec Touchpoints

- **Extends**: 无；该 spec 建立新的 Phase 3 engine 生命周期边界。
- **Adjacent**: phase-02-generation-job-model 提供任务和结果模型；phase-07-generation-job-api 后续暴露服务契约。

## Constraints

生命周期实现必须留在 Go 后端 engine/service 层，不能把状态规则放入 Wails binding 或 Vue 页面。状态和错误信息不得包含数据库密码、用户 SQL 或生成数据内容等敏感信息。
