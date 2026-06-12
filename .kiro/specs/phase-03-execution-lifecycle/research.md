# Research & Design Decisions

## Summary

- **Feature**: `phase-03-execution-lifecycle`
- **Discovery Scope**: Extension / Integration-focused discovery
- **Key Findings**:
  - 现有 `internal/domain/execution` 已定义 GenerationJob、ExecutionTask、ExecutionTableResult、状态枚举、安全错误快照和字段级校验，可作为 lifecycle 输入与结果映射的上游合同。
  - 项目 steering 要求业务规则保留在 Go 后端 service/domain/engine 层，Wails binding 与 Vue 不承载状态流转和执行控制规则。
  - Phase 3 roadmap 明确执行生命周期只建立 engine 入口和状态边界，下游依赖排序、行数规划、生成上下文、批量生成和写入适配由后续 specs 接入。

## Research Log

### 现有执行领域模型

- **Context**: 本规格依赖 `phase-02-generation-job-model`，需要确认可复用的任务模型、状态与错误合同。
- **Sources Consulted**:
  - `E:/git/loomidbx-v2/.kiro/specs/phase-02-generation-job-model/design.md`
  - `E:/git/loomidbx-v2/internal/domain/execution/generationjob.go`
  - `E:/git/loomidbx-v2/internal/domain/execution/executiontask.go`
  - `E:/git/loomidbx-v2/internal/domain/execution/executiontableresult.go`
  - `E:/git/loomidbx-v2/internal/domain/execution/status.go`
  - `E:/git/loomidbx-v2/internal/domain/execution/errorsnapshot.go`
  - `E:/git/loomidbx-v2/internal/domain/execution/validation.go`
- **Findings**:
  - `GenerationJob` 已聚合 `ExecutionTask` 与 `[]ExecutionTableResult`，明确不承载生命周期算法。
  - `ExecutionTaskStatus` 当前只有 `RUNNING`、`SUCCESS`、`PARTIAL_FAILED`、`FAILED`，未包含 lifecycle 内部需要的 initialized、prechecked、cancelled 等细粒度状态。
  - `ExecutionErrorSnapshot` 已强调安全错误摘要，不应回显凭据、用户 SQL 或生成数据。
- **Implications**:
  - lifecycle 需要新增 engine 内部状态模型，不能直接扩大 Phase 2 持久化枚举语义。
  - lifecycle 最终快照可以映射到 Phase 2 领域模型，但内部运行时状态应保持在 engine 包内。

### Steering 与架构边界

- **Context**: 执行生命周期容易被 service、binding、UI 分散实现，需要确认责任归属。
- **Sources Consulted**:
  - `E:/git/loomidbx-v2/.kiro/steering/product.md`
  - `E:/git/loomidbx-v2/.kiro/steering/tech.md`
  - `E:/git/loomidbx-v2/.kiro/steering/structure.md`
  - `E:/git/loomidbx-v2/.kiro/steering/roadmap.md`
- **Findings**:
  - Go 后端承载领域模型、服务层、执行引擎和业务规则。
  - 执行前预检由服务层和执行引擎统一完成；UI 和 Wails binding 不应实现核心生成算法或状态规则。
  - Phase 3 engine 可以定义最小生成器调用接口和写入适配接口，但不实现完整注册表、内置生成器集合或真实写入。
- **Implications**:
  - 本规格应把 lifecycle coordinator、state machine、precheck result、control token 与 downstream seams 放入 `internal/engine/lifecycle` 或等价 engine 子包。
  - 后续 service/facade 只调用 lifecycle 入口，不复制状态流转规则。

### 下游接缝与最小抽象

- **Context**: 生命周期必须为后续 Phase 3 specs 留出入口，但不能提前实现后续算法。
- **Sources Consulted**:
  - `E:/git/loomidbx-v2/.kiro/specs/phase-03-execution-lifecycle/brief.md`
  - `E:/git/loomidbx-v2/.kiro/steering/roadmap.md`
- **Findings**:
  - 下游 specs 包括依赖图与拓扑排序、行数规划、生成上下文、批量生成循环、批量写入适配、执行结果与错误模型。
  - 本规格只定义执行入口、状态机、预检结果和生命周期事件边界。
  - 下游组件失败必须回到统一失败语义和安全错误摘要。
- **Implications**:
  - 需要定义小型 `Prechecker`、`PlannerHook`、`GenerationHook`、`ResultHook` 或合并后的 runner seam，用于测试生命周期语义。
  - seam 的默认实现应为 no-op 或 test fake，不包含实际依赖图、生成器注册、批处理或数据库写入逻辑。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| Engine lifecycle coordinator + ports | 在 engine 内部集中状态机、预检、控制语义，并通过小接口接入下游阶段 | 状态规则集中、易单测、符合后端 owns business rules | 需要明确接口不能膨胀为后续算法 | Selected |
| Service-only orchestration | 由 service 直接编排预检、状态和下游调用 | 入口靠近业务用例 | 容易把 engine 规则散落到 service/facade，后续复用困难 | Rejected |
| 扩展 Phase 2 domain status | 直接在 ExecutionTaskStatus 增加更多生命周期状态 | 可减少新状态模型 | 会污染持久化/历史模型，与 Phase 2 边界冲突 | Rejected |
| 提前实现完整 pipeline | 一次性实现计划、生成、写入和结果 | 功能闭环完整 | 跨越多个 roadmap specs，超出当前边界 | Rejected |

## Design Decisions

### Decision: engine 内部状态机独立于历史状态枚举

- **Context**: Phase 2 执行历史状态只描述任务结果和历史快照，不适合表达 initialized、prechecking、cancelling 等运行时控制细节。
- **Alternatives Considered**:
  1. 复用 `ExecutionTaskStatus` 并增加新枚举。
  2. 在 engine lifecycle 中定义内部 `LifecycleState` 并提供最终映射。
- **Selected Approach**: 使用 engine 内部状态机表达 `INITIALIZED`、`PRECHECKING`、`READY`、`RUNNING`、`CANCELLING`、`CANCELLED`、`FAILED`、`COMPLETED` 等生命周期状态；只在最终快照需要时映射到 Phase 2 历史模型。
- **Rationale**: 保持 Phase 2 持久化合同稳定，避免把运行时控制语义写入历史模型。
- **Trade-offs**: 多一个状态模型，但边界更清晰、测试更直接。
- **Follow-up**: 实现时需提供非法流转测试，避免服务层绕过状态机。

### Decision: 预检结果统一汇总但不执行下游真实能力

- **Context**: 预检需要阻止不安全任务启动，但依赖排序、行数规划和写入检查属于后续 specs。
- **Alternatives Considered**:
  1. 本规格实现所有预检逻辑。
  2. 本规格只校验输入并定义可组合预检结果。
- **Selected Approach**: lifecycle 提供预检入口、结果结构、阻断/警告分类和下游 precheck seam；当前只实现快照与状态前置条件检查。
- **Rationale**: 满足当前可测试能力，同时允许后续 specs 插入更丰富预检。
- **Trade-offs**: 当前预检能力有限，需要后续 specs 补齐业务规则。
- **Follow-up**: 后续 dependency graph、row count、generation context 和 result/error specs 需要复用此结果结构。

### Decision: 下游接缝采用最小 port，不引入重型 pipeline 框架

- **Context**: 生命周期需要调用后续计划、生成、结果汇总，但当前不能实现完整 pipeline。
- **Alternatives Considered**:
  1. 引入完整 pipeline / middleware abstraction。
  2. 使用最小接口和 fake/no-op 实现验证生命周期。
- **Selected Approach**: 定义最小 downstream seam：计划、生成、结果汇总各自返回阶段结果或安全错误；当前实现只用于触发生命周期状态变化。
- **Rationale**: 避免过早抽象，遵循 roadmap 小边界拆分。
- **Trade-offs**: 后续 specs 可能需要扩展 seam 字段，届时触发 revalidation。
- **Follow-up**: 接缝扩展必须保持生命周期状态规则不变。

### Decision: 生命周期错误默认安全化

- **Context**: 产品和 brief 都要求错误不包含数据库密码、用户 SQL 或生成数据内容。
- **Alternatives Considered**:
  1. 直接透传下游错误消息。
  2. 使用安全错误摘要和敏感标记过滤。
- **Selected Approach**: lifecycle 对外只暴露错误码、阶段、字段路径和安全消息；下游原始错误不得出现在公开错误结果中。
- **Rationale**: 符合本地隐私边界，并与 Phase 2 `ExecutionErrorSnapshot` 一致。
- **Trade-offs**: 开发排查需要内部安全日志或调试机制，但复杂可观测性不属于本规格。
- **Follow-up**: 后续错误模型和可观测性 specs 可扩展内部诊断，但不能改变对外安全边界。

## Risks & Mitigations

- 状态语义与 Phase 2 历史状态混淆 — 使用独立 lifecycle 状态，并在设计和测试中验证不修改 Phase 2 枚举。
- 下游 seam 过度设计 — 仅保留预检、计划、生成、结果汇总阶段的最小输入输出，不包含依赖图、批次、writer adapter 或 result aggregator 细节。
- 取消语义被实现为真实 goroutine 中断 — 当前只定义可观察取消意图和控制 token，不实现并发调度或强制中断。
- 错误消息泄露敏感信息 — 对生命周期错误统一安全化，并用单元测试覆盖敏感标记过滤。

## References

- `E:/git/loomidbx-v2/.kiro/specs/phase-03-execution-lifecycle/brief.md` — feature scope and constraints.
- `E:/git/loomidbx-v2/.kiro/steering/roadmap.md` — Phase 3 dependency order and boundaries.
- `E:/git/loomidbx-v2/.kiro/steering/tech.md` — backend ownership, engine standards, privacy and error rules.
- `E:/git/loomidbx-v2/.kiro/steering/structure.md` — module boundaries and dependency direction.
- `E:/git/loomidbx-v2/.kiro/specs/phase-02-generation-job-model/design.md` — upstream execution domain model contract.
