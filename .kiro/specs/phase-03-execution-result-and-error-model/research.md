# Research Document

## Summary

本研究确认 `phase-03-execution-result-and-error-model` 应作为 Phase 3 engine 的收尾边界：它不拥有生命周期状态机、规划算法、批次生成或数据库写入，而是消费这些阶段已经定义的安全 result/issue 形态，归一化为运行时 `ExecutionResult`、`TableExecutionResult`、`BatchResult` 和 `EngineError`，并提供到 Phase 2 execution domain 模型的纯内存映射。

核心设计选择是新增 `internal/engine/result` 包，采用 mapper/aggregator 模式隔离上游阶段模型与下游历史模型。该包只保存安全统计、状态、错误分类和失败范围，不保存 SQL、参数、连接信息、规则参数或生成数据内容。

## Research Log

| Source | Findings | Impact |
|--------|----------|--------|
| `.kiro/specs/phase-03-execution-result-and-error-model/brief.md` | 明确需要统一执行结果、表级结果、批次失败范围、错误分类和 Phase 2 历史映射；排除 API/UI/日志/遥测/重试 | 设定本规格 owned boundary 和 non-goals |
| `.kiro/steering/roadmap.md` | Phase 3 最后一个 spec，依赖 batch writer adapter；Phase 3 只保留必要状态、结果和错误边界 | result 包应位于 writer 之后，服务后续 Phase 7/8/9 |
| `.kiro/steering/product.md` | 本地隐私优先，失败需可追踪，但连接信息、schema、规则、生成数据和用户 SQL 不应上传 | 强化安全错误和敏感过滤测试 |
| `.kiro/steering/tech.md` | Go backend owns engine/result/history；Wails 只是桥接；错误需可诊断但不泄密 | 包结构放在 backend engine，不引入 Wails/Vue |
| `.kiro/steering/structure.md` | Domain、engine、adapter 依赖方向必须清晰 | result 包可依赖 engine upstream 和 domain execution，但 domain 不依赖 engine |
| `phase-03-execution-lifecycle` | lifecycle 负责入口、预检、状态机、取消和阶段切换，提供安全错误 | 本规格不重写状态机，只消费状态/错误快照 |
| `phase-03-dependency-graph-and-topological-sort` | planner 输出 `ExecutionPlan`、`PlanResult`、`PlanIssue` | planner issue 映射为 planning/precheck EngineError |
| `phase-03-row-count-planning` | rowcount 输出 `RowCountPlan`、`RowCountResult`、`RowCountIssue` | row count target 和不可规划错误进入 task/table result |
| `phase-03-generation-context` | context 输出表/字段快照、引用存储和 `ContextIssue` | context 和 reference issue 映射为 context/reference EngineError |
| `phase-03-batch-generation-loop` | batch 包已有 `BatchResult`、`BatchIssue`、progress/stats 和 writer seam failure | 当前 spec 的 runtime `BatchResult` 应作为最终结果摘要，映射 batch 包结果而不复制主循环职责 |
| `phase-03-batch-writer-adapter` | writer 输出 `WriteResult`、`WriterIssue`、partial accepted、statement count 和安全错误 | writer issue/result 直接驱动批次结果和部分失败状态 |
| `phase-02-generation-job-model` | domain execution 定义 `ExecutionTaskStatus`、`ExecutionTableStatus`、`ExecutionErrorSnapshot` | 需要纯内存 history mapper，不实现 repository/API/migration |

## Architecture Pattern Evaluation

### Option A: 将结果模型放入 lifecycle 包

- **Pros**: lifecycle 已有任务状态，调用链近。
- **Cons**: 会让 lifecycle 状态机吸收 planner/batch/writer/history 映射细节，扩大生命周期边界；容易形成 lifecycle 与所有阶段的强耦合。
- **Decision**: 拒绝。lifecycle 继续负责状态流转，本规格独立负责最终结果汇总。

### Option B: 将结果模型放入 batch 或 writer 包

- **Pros**: batch/writer 已有批次统计和错误。
- **Cons**: 预检、规划、上下文、取消和历史映射不是 batch/writer 的职责；会导致 writer 了解 Phase 2 历史模型。
- **Decision**: 拒绝。batch/writer 结果通过 mapper 进入 result 包。

### Option C: 新增 `internal/engine/result` 包

- **Pros**: 清晰表达 Phase 3 收尾边界；能汇总所有上游结果；可单向依赖 Phase 2 execution domain 做历史映射；便于边界测试。
- **Cons**: 需要定义少量适配 DTO 或 mapper，避免上游包反向依赖。
- **Decision**: 采用。

## Design Decisions

### Decision 1: Runtime result 与 upstream batch result 分离

`phase-03-batch-generation-loop` 已有 `BatchResult` 用于主循环阶段结果；本规格仍定义最终运行时 `BatchResult`，但语义限定为 execution result 内的批次摘要。实现时应通过明确命名或包限定区分，例如 `result.BatchResult` 与 `batch.BatchResult`。这样可以避免将主循环内部错误模型直接暴露给历史/API。

### Decision 2: EngineError 是公开安全错误，不保存 raw cause

`EngineError` 仅包含 code、category、stage、scope、fieldPath、safeMessage、blocking、occurredAt 和可选 source summary。任何 raw error、SQL、参数、DSN、密码、令牌、规则参数或生成值只允许在上游内部处理，不进入 result 包公开模型。

### Decision 3: History mapper 是纯内存转换

本规格只提供从 runtime result 到 `internal/domain/execution` 模型的 mapper，包括 `ExecutionTaskStatus`、`ExecutionTableStatus`、rows written 和 `ExecutionErrorSnapshot`。它不创建 repository、不写数据库、不处理迁移、不定义 API DTO。

### Decision 4: 状态推导以阻断范围和接受范围为准

任务成功要求所有计划表成功且无 blocking errors；写入后发生失败则为 canonical `PARTIAL_FAILED`；写入前任务级阻断为 canonical `FAILED`；取消不新增 runtime 任务状态，而是根据已接受范围使用 `FAILED` 或 `PARTIAL_FAILED`，并通过 cancellation error snapshot 保留原因。

### Decision 5: 保持 deterministic ordering

错误、表结果和批次结果应按阶段顺序、拓扑顺序、批次 index、statement index 和输入顺序确定性输出，便于测试、历史快照和后续 UI 展示。

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 与 batch 包已有 `BatchResult` 命名混淆 | 实现和 API 消费者误用模型 | 使用包名限定、文档说明和 mapper 测试区分阶段结果与最终结果 |
| 上游 issue 字段尚未实现或后续变更 | result mapper 难以直接依赖具体类型 | 允许先定义同构 adapter DTO 和 mapper 函数；上游类型稳定后再集成 |
| Phase 2 无取消状态 | 取消语义无法作为独立任务状态持久化 | 不新增平行 runtime 任务状态；取消且无成功写入使用 `FAILED`，部分取消使用 `PARTIAL_FAILED`，并在安全错误快照中保留 cancellation code |
| 敏感信息通过 safe message 泄露 | 违反本地隐私边界 | 固定消息模板、敏感样本测试、禁止 raw error 透传 |
| result 包过度实现观测/历史能力 | 跨阶段范围膨胀 | 边界测试禁止 API/UI/log/query/metrics/retry/recovery 相关实现 |

## References

- `.kiro/specs/phase-03-execution-result-and-error-model/brief.md`
- `.kiro/steering/roadmap.md`
- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`
- `.kiro/specs/phase-03-execution-lifecycle/requirements.md`
- `.kiro/specs/phase-03-execution-lifecycle/design.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/requirements.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/design.md`
- `.kiro/specs/phase-03-row-count-planning/requirements.md`
- `.kiro/specs/phase-03-row-count-planning/design.md`
- `.kiro/specs/phase-03-generation-context/requirements.md`
- `.kiro/specs/phase-03-generation-context/design.md`
- `.kiro/specs/phase-03-batch-generation-loop/requirements.md`
- `.kiro/specs/phase-03-batch-generation-loop/design.md`
- `.kiro/specs/phase-03-batch-writer-adapter/requirements.md`
- `.kiro/specs/phase-03-batch-writer-adapter/design.md`
- `.kiro/specs/phase-02-generation-job-model/requirements.md`
- `.kiro/specs/phase-02-generation-job-model/design.md`
