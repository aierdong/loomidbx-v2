# Requirements Document

## Introduction

`phase-03-execution-lifecycle` 定义生成执行引擎的最小生命周期边界。Phase 2 已提供 Project、GenerationJob、ExecutionTask 和 ExecutionTableResult 等领域模型，但尚未提供 engine 入口、预检入口、运行时状态流转或启动、取消、失败、完成语义。

本规格面向 Go 后端 engine/service 层，要求系统能够接收执行任务快照，完成预检、状态初始化、运行、取消、失败和完成的生命周期表达，并为后续依赖排序、行数规划、生成上下文、批量生成和结果汇总组件预留稳定接缝。

## Boundary Context

- **In scope**: 执行任务快照到 engine 输入的映射边界、执行状态机、预检结果、启动/取消/失败/完成语义、下游计划/生成/结果汇总组件接缝、生命周期错误的敏感信息边界。
- **Out of scope**: 依赖图算法、拓扑排序、行数规划、生成上下文、批量生成循环、写入适配器真实实现、真实数据库写入、完整生成器注册表、API/UI、Wails runtime events、复杂日志与可观测性。
- **Adjacent expectations**: 上游 `phase-02-generation-job-model` 提供任务与历史模型；后续 Phase 3 specs 将复用本生命周期入口并在各自边界内补齐计划、行数、上下文、批量生成、写入适配和结果汇总能力。

## Requirements

### Requirement 1: 执行输入与快照边界

**Objective:** As a 后端开发人员, I want 系统能够接收稳定的执行任务快照, so that 后续 engine 能以明确输入进入执行生命周期。

#### Acceptance Criteria

1. When 调用方提交待执行任务快照时, the 执行生命周期 shall 校验快照包含可识别的任务身份、Project 引用和执行表结果边界。
2. When 快照通过基础校验时, the 执行生命周期 shall 创建不依赖 UI、Wails binding 或真实数据库连接的 engine 执行输入。
3. If 快照缺少必填身份或引用字段, then the 执行生命周期 shall 返回字段级预检失败结果而不是进入运行状态。
4. The 执行生命周期 shall 不从 UI、Wails binding 或 Vue 页面接收业务规则作为生命周期判断依据。
5. The 执行生命周期 shall 保留与下游计划、生成和结果汇总组件交互所需的最小上下文，不提前计算依赖顺序、行数、批次内容或最终结果。

### Requirement 2: 预检入口与结果表达

**Objective:** As a 后端开发人员, I want 系统在启动执行前提供统一预检入口, so that 无法安全运行的任务可以在进入运行阶段前被阻止。

#### Acceptance Criteria

1. When 执行请求进入预检阶段时, the 执行生命周期 shall 返回包含通过状态、阻断错误和非阻断警告的预检结果。
2. If 预检发现任务快照、生命周期状态或下游接缝前置条件不满足, then the 执行生命周期 shall 阻止任务启动并返回安全错误摘要。
3. While 预检正在进行时, the 执行生命周期 shall 保持任务未进入运行状态。
4. Where 后续规格提供依赖排序、行数规划、生成上下文或结果汇总预检能力, the 执行生命周期 shall 能够接收这些预检结果并统一汇总。
5. The 执行生命周期 shall 不在本规格内执行真实数据库查询、SQL 校验、生成器参数深度校验或写入模拟。

### Requirement 3: 状态机与允许流转

**Objective:** As a 后端开发人员, I want 系统明确执行生命周期状态和允许流转, so that 启动、取消、失败和完成语义不会分散到 service、binding 或 UI 中。

#### Acceptance Criteria

1. When 生命周期对象被初始化时, the 执行生命周期 shall 将其置于可预检但尚未运行的初始状态。
2. When 预检通过并启动执行时, the 执行生命周期 shall 按允许流转进入运行状态。
3. If 调用方请求非法状态流转, then the 执行生命周期 shall 拒绝该流转并返回状态冲突错误。
4. While 生命周期处于终态时, the 执行生命周期 shall 拒绝再次启动、完成、失败或取消该执行。
5. The 执行生命周期 shall 明确定义取消、失败和完成为互斥终态，并保留可测试的状态流转记录。

### Requirement 4: 启动、取消、失败与完成语义

**Objective:** As a 后端开发人员, I want 系统提供最小执行控制语义, so that 后续执行计划、生成循环和结果汇总组件可以复用一致的控制边界。

#### Acceptance Criteria

1. When 执行被启动时, the 执行生命周期 shall 记录启动时间、运行状态和可供下游组件读取的执行上下文。
2. When 调用方请求取消运行中的执行时, the 执行生命周期 shall 标记取消意图并使后续下游步骤能够观察该取消状态。
3. If 运行过程中发生任务级失败, then the 执行生命周期 shall 记录安全失败摘要并进入失败终态。
4. When 所有当前边界内的执行步骤成功结束时, the 执行生命周期 shall 记录完成时间并进入完成终态。
5. The 执行生命周期 shall 在取消、失败或完成后提供一致的最终快照，供后续历史记录、API 或 UI 规格消费。

### Requirement 5: 下游组件接缝

**Objective:** As a 后端开发人员, I want 系统定义生命周期层与后续计划、生成和结果汇总组件的接缝, so that 后续 Phase 3 specs 可以独立实现而不重写生命周期规则。

#### Acceptance Criteria

1. Where 下游计划组件被接入时, the 执行生命周期 shall 通过稳定接缝请求执行计划结果而不在本规格内实现依赖图或行数规划。
2. Where 下游生成组件被接入时, the 执行生命周期 shall 通过稳定接缝触发生成步骤而不在本规格内实现生成器注册表或批量生成循环。
3. Where 下游结果汇总组件被接入时, the 执行生命周期 shall 通过稳定接缝接收最终结果汇总而不在本规格内实现 result aggregator、writer adapter 或真实数据库写入。
4. If 任一下游接缝返回失败, then the 执行生命周期 shall 将失败转化为生命周期失败语义并保留安全错误摘要。
5. The 执行生命周期 shall 允许后续组件替换接缝实现，同时保持生命周期状态规则不变。

### Requirement 6: 安全错误与边界验证

**Objective:** As a 后端开发人员, I want 系统在生命周期错误和测试中保持敏感信息边界, so that 执行失败可诊断但不会泄露数据库密码、用户 SQL 或生成数据内容。

#### Acceptance Criteria

1. If 生命周期、预检或下游接缝返回错误, then the 执行生命周期 shall 只暴露错误码、阶段、字段路径和安全消息。
2. If 错误来源包含数据库密码、用户 SQL、连接详情或生成数据内容, then the 执行生命周期 shall 不在对外错误消息中包含这些原始内容。
3. The 执行生命周期 shall 通过单元测试覆盖输入校验、预检结果、合法状态流转、非法状态流转、取消、失败和完成语义。
4. The 执行生命周期 shall 通过边界测试确认 engine 生命周期包不依赖 Wails、Vue、真实数据库驱动或前端 API。
5. The 执行生命周期 shall 通过接缝测试确认本规格没有实现依赖图算法、行数规划、生成器注册表、批量生成循环或真实写入。
