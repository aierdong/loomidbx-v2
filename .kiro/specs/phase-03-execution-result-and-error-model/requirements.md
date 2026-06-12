# Requirements Document

## Introduction

`phase-03-execution-result-and-error-model` 定义 Phase 3 执行引擎收尾处的统一运行时结果与安全错误模型。上游 `phase-03-execution-lifecycle`、`phase-03-dependency-graph-and-topological-sort`、`phase-03-row-count-planning`、`phase-03-generation-context`、`phase-03-batch-generation-loop` 和 `phase-03-batch-writer-adapter` 已分别表达状态流转、预检/规划问题、生成上下文、批次生成和写入结果；Phase 2 已定义 `GenerationJob`、`ExecutionTask`、`ExecutionTableResult` 和 `ExecutionErrorSnapshot` 等执行历史领域模型。当前仍缺少 engine 侧统一汇总执行结果、表级结果、批次结果、错误分类、失败范围和最小历史映射的稳定边界。

本规格面向 Go 后端 engine/result 接缝层，要求系统能够将 lifecycle、planner、rowcount、generation context、batch loop 和 writer adapter 的安全结果归一化为统一 `ExecutionResult`，包含任务状态、表级状态、行数统计、批次统计、错误分类、失败范围和诊断摘要，并提供到 Phase 2 执行历史模型的纯内存映射边界。该边界不得实现完整执行历史 API、UI 进度页、日志查询、复杂 tracing/metrics/远端遥测、生成数据持久化、自动重试或失败恢复策略。

## Boundary Context

- **In scope**: engine 侧 `ExecutionResult`、`TableExecutionResult`、`BatchResult`、`EngineError`、错误码/类别/阶段/失败范围模型；上游 safe issue/result 到统一错误模型的映射；任务/表/批次/字段级失败范围表达；直接复用 Phase 2 canonical 任务/表状态的成功、失败、部分失败和跳过结果汇总规则；取消通过 cancellation `EngineError` 与安全摘要表达；Phase 2 `ExecutionTask`、`ExecutionTableResult`、`ExecutionErrorSnapshot` 的最小历史映射；安全诊断和敏感信息过滤测试。
- **Out of scope**: 执行历史查询 API、Wails binding、Vue 进度页、日志查询 UI、复杂 observability pipeline、远端遥测、生成数据内容持久化、SQL/参数持久化、自动重试、断点续跑、补偿写入、失败恢复策略、真实数据库事务实现、writer adapter 内部实现。
- **Adjacent expectations**: lifecycle 负责状态机和阶段切换；plan/rowcount/gencontext/batch/writer 负责各自阶段的安全 issue/result；本规格负责统一运行时结果和历史模型映射；后续 Phase 7/8/9 可以复用本规格的安全合同实现 API/UI/可观测性。

## Requirements

### Requirement 1: 统一执行结果输入边界

**Objective:** As a 后端开发人员, I want result model 接收生命周期和各执行阶段的最小安全输出, so that engine 收尾阶段只汇总已归一化的执行信息而不重新运行阶段逻辑。

#### Acceptance Criteria

1. When 调用方提交任务身份、Project 身份、lifecycle 快照、计划结果、行数计划结果、上下文结果、批次生成结果、writer 结果和取消状态时, the system shall 构造统一执行结果汇总输入。
2. When 汇总输入通过基础校验时, the system shall 保留 TaskID、ProjectID、阶段状态、表身份、批次范围、行数统计、写入统计和上游安全错误所需的最小字段。
3. If 汇总输入缺少必需身份、任务/Project 身份不一致、表身份无法对应、阶段顺序非法或统计为负数, then the system shall 返回阻断结果汇总错误而不是生成部分可信结果。
4. If 上游阶段仅提供安全 issue 而未提供完整阶段结果, then the system shall 仍能基于安全 issue 形成任务级失败结果和错误摘要。
5. The system shall 不重新执行 lifecycle、依赖图构建、拓扑排序、行数规划、生成上下文构建、批次生成、writer adapter、真实数据库写入或 UI/API 查询。

### Requirement 2: 任务级状态和完成语义

**Objective:** As a 后端开发人员, I want ExecutionResult 直接使用 canonical 任务级成功、失败和部分失败语义，并用安全错误表达取消, so that lifecycle 和后续历史/API 可以一致理解执行结论。

#### Acceptance Criteria

1. When 所有计划内表均成功完成且没有阻断错误时, the system shall 将执行结果标记为成功并汇总总目标行数、接受行数、写入行数和批次数。
2. When 任一任务级预检、规划、上下文或生命周期错误在写入前阻断执行时, the system shall 将执行结果标记为失败并保留任务级失败范围。
3. When 部分表或批次已经成功接受但后续表、批次或 writer 失败时, the system shall 将执行结果标记为部分失败并保留已接受范围和失败范围。
4. When 执行被取消时, the system shall 记录 cancellation `EngineError` 和安全的已完成/未完成摘要，并根据是否存在已接受范围将任务状态设为 canonical `FAILED` 或 `PARTIAL_FAILED`。
5. The system shall 不实现自动重试、断点续跑、幂等键、补偿写入或失败恢复策略来改变任务级结论。

### Requirement 3: 表级结果汇总

**Objective:** As a 后端开发人员, I want TableExecutionResult 表达每张执行表的目标、成功、失败和跳过状态, so that 用户可以知道哪些表成功、失败或因依赖失败跳过。

#### Acceptance Criteria

1. When 上游计划中存在执行表时, the system shall 按拓扑执行顺序为每张表生成表级结果摘要。
2. When 表的所有批次成功接受时, the system shall 将该表标记为成功并汇总 target rows、generated rows、accepted rows、batch count 和 statement count。
3. If 表内任一批次生成、引用提交或 writer 失败, then the system shall 将该表标记为失败并记录安全错误摘要和失败批次范围。
4. If 上游依赖表失败导致后续表未执行, then the system shall 将受影响表标记为 skipped，并使用依赖失败的安全范围而不是生成虚假批次结果。
5. The system shall 不重新排序表、不重新计算目标行数、不写回 Project 执行顺序、不读取数据库确认行数。

### Requirement 4: 批次级结果和部分接受范围

**Objective:** As a 后端开发人员, I want BatchResult 表达每个批次的范围、状态和接受统计, so that 写入失败可以定位到安全批次范围而不泄露写入内容。

#### Acceptance Criteria

1. When batch loop 或 writer adapter 返回批次成功时, the system shall 记录 batch index、start row、end row、accepted rows 和 statement count。
2. When 非事务 writer 失败且部分 statement 已执行时, the system shall 在批次结果中标记 partial accepted 并保留安全的 accepted rows 和 statement index 摘要。
3. If 批次生成失败、引用缺失、引用记录失败、dialect 失败或 executor 失败, then the system shall 将该批次标记为失败并关联统一 EngineError。
4. If 表未执行或被跳过, then the system shall 不伪造成功批次，允许批次结果为空并由表级状态表达原因。
5. The system shall 不保存批次行值、生成数据、SQL 文本、statement 参数或 driver 原始错误载荷。

### Requirement 5: EngineError 分类、阶段和失败范围

**Objective:** As a 后端开发人员, I want EngineError 统一表达错误分类、阶段、失败范围和安全诊断, so that 所有 Phase 3 阶段失败都能被一致复盘。

#### Acceptance Criteria

1. When 上游返回 lifecycle、precheck、planner、rowcount、context、generation、reference、writer、transaction、clear、dialect、executor 或 cancellation issue 时, the system shall 将其映射到统一 EngineError 分类和阶段。
2. When 错误需要定位影响范围时, the system shall 使用 task、project、project table、table、column、batch、row、statement 和 field path 等安全标识表达 failure scope。
3. If 一个阶段返回多个错误或警告, then the system shall 保持确定性顺序并区分 blocking errors 与 warnings。
4. If 错误来源未知或上游 issue 字段不完整, then the system shall 使用安全 fallback code、stage 和 message，而不是透传原始错误文本。
5. The system shall 不把数据库产品名称、SQL 内容、连接字符串、密码、令牌、规则参数原文、生成值或原始 driver 错误作为公开诊断字段。

### Requirement 6: 上游阶段结果归一化

**Objective:** As a 后端开发人员, I want result model 显式适配 Phase 3 各阶段的安全结果形态, so that upstream specs 可以独立演进且最终结果保持稳定。

#### Acceptance Criteria

1. When 接收 lifecycle precheck 或状态机错误时, the system shall 映射为任务级 lifecycle/precheck EngineError，并保持 lifecycle 状态摘要。
2. When 接收 dependency plan、rowcount 或 generation context issue 时, the system shall 映射为 planning/context EngineError，并保留可安全定位的表、关系或字段路径。
3. When 接收 batch generation loop issue 或 progress summary 时, the system shall 映射为 generation/reference/batch EngineError 和表/批次统计。
4. When 接收 writer adapter result 或 writer issue 时, the system shall 映射为 writer EngineError、批次结果和部分接受摘要。
5. The system shall 不要求上游包依赖 result 包来构造其内部 issue；允许通过 mapper 函数或同构 DTO 在 result 边界归一化。

### Requirement 7: Phase 2 执行历史模型映射边界

**Objective:** As a 后端开发人员, I want ExecutionResult 能映射到 Phase 2 执行历史领域模型, so that 后续服务/API/UI 可以复用稳定历史表达而无需理解 engine 内部细节。

#### Acceptance Criteria

1. When ExecutionResult 成功、失败、部分失败或包含取消错误时, the system shall 直接使用 Phase 2 `ExecutionTaskStatus` 的 canonical 状态；取消不得新增平行任务状态，应通过 cancellation error snapshot 和安全摘要区分。
2. When TableExecutionResult 成功、失败或跳过时, the system shall 映射到 Phase 2 `ExecutionTableStatus`、rows written、execution order 和表名/schema 名快照字段。
3. If EngineError 需要进入历史模型, then the system shall 映射为 `ExecutionErrorSnapshot` 的 code、message、fieldPath 和 occurredAt，不包含敏感原始载荷。
4. If Phase 2 当前持久化合同只能保存单字段 error message, then the system shall 提供安全摘要作为降级映射来源并保留结构化运行时错误供内存消费。
5. The system shall 不实现执行历史 repository、数据库迁移、历史查询 API、Wails DTO、Vue 页面或持久化事务。

### Requirement 8: 安全诊断和敏感信息边界

**Objective:** As a 后端开发人员, I want result/error model 过滤所有上游错误和诊断内容, so that 结果可诊断但不会泄露凭据、SQL、规则参数或生成数据。

#### Acceptance Criteria

1. If 输入汇总、状态推导、表/批次汇总、上游 issue 映射或历史映射失败, then the system shall 只暴露错误码、类别、阶段、字段路径、安全消息、阻断标记、发生时间和安全范围摘要。
2. If 上游 issue 或 error payload 包含 SQL、DSN、连接字符串、密码、令牌、规则参数、参数值、生成值或 driver 原始错误, then the system shall 用固定安全消息替代公开文本。
3. When 公开错误需要诊断位置时, the system shall 使用 ID、索引、枚举阶段和字段路径，不使用原始数据值或 SQL 片段。
4. The system shall 通过测试确认 ExecutionResult、TableExecutionResult、BatchResult、EngineError 和 ExecutionErrorSnapshot 的公开字段不包含敏感样本。
5. The system shall 不把原始下游错误载荷透传给 lifecycle、API、UI、Wails binding、历史模型、日志或观测管道。

### Requirement 9: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定 result/error 边界, so that 后续 API/UI/可观测性阶段可以独立接入而不重写执行引擎结果规则。

#### Acceptance Criteria

1. The system shall 通过单元测试覆盖执行结果输入校验、任务状态推导、表级汇总、批次汇总、错误映射和历史映射。
2. The system shall 通过测试覆盖成功、预检失败、规划失败、上下文失败、生成失败、引用失败、writer 失败、非事务部分接受、取消和跳过场景。
3. The system shall 通过接缝测试证明结果可以消费 lifecycle、planner、rowcount、gencontext、batch 和 writer 的安全 issue/result，并可映射到 Phase 2 execution domain 模型。
4. The system shall 通过边界测试确认不依赖 Wails、Vue、前端 API、store、facade、真实数据库 driver、连接管理、凭据存储或数据库产品名称硬编码。
5. The system shall 通过边界测试确认没有实现执行历史查询 API、UI 进度页、日志查询、复杂 tracing/metrics/远端遥测、生成数据持久化、自动重试、断点续跑、补偿写入或恢复策略。
