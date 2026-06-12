# Research Document

## Feature: phase-03-batch-generation-loop

## 研究目标

本研究用于支撑 `phase-03-batch-generation-loop` 的设计与任务拆分，确认批量生成主循环在 Phase 3 engine 中的职责边界、上游输入契约、下游接缝、错误安全要求和未来能力隔离。

## 已阅读上下文

### 目标功能简述

- `brief.md` 明确本功能负责 engine 批量生成主循环。
- 主循环需要按拓扑计划和 `RowCountPlan` 调度表与批次。
- 主循环需要按字段计划组装行值，关系字段必须优先使用 `GenerationContext` / `RuntimeReferenceStore` 的引用。
- 主循环需要调用最小 generator invoker 和 writer seam。
- 主循环需要向 lifecycle 交付进度、失败和安全错误摘要。

### Steering 结论

- `roadmap.md` 将本功能放在 Phase 3，依赖 `phase-03-generation-context`，职责是按执行计划和批次调度字段生成、行组装、关系值填充和批次状态推进。
- `product.md` 强调约束优先于展示和随机性，依赖表先生成，外键必须优先使用关系引用，无法安全生成时必须给出明确原因。
- `tech.md` 将 engine 生成路径定义为计划、排序、行数、上下文、生成器调用、批量写入和状态记录；generator 不应直接访问数据库。
- `structure.md` 将执行生命周期、计划、上下文、批处理和结果归入 `internal/engine`，将 generator registry 和 built-ins 归入后续 generator 层。

## 上游规格契约

### phase-03-execution-lifecycle

- 生命周期负责接收执行输入、聚合预检、驱动阶段切换和统一失败语义。
- 生命周期可以通过下游 generation port 调用生成阶段，但不拥有批量生成循环、generator registry 或 writer 内部。
- 生成失败需要映射为 lifecycle 兼容安全错误，不修改 lifecycle 状态枚举。

### phase-03-dependency-graph-and-topological-sort

- 输出稳定 `ExecutionPlan`，核心字段为 `ProjectID`、`OrderedTables`、`Edges` 和 `Warnings`。
- `OrderedTables` 中每个 `PlannedTable` 提供 `ProjectTableID`、`TableID` 和 `ExecutionOrder`。
- 批量生成主循环必须消费该顺序，不得重建依赖图或重新排序。

### phase-03-row-count-planning

- 输出稳定 `RowCountPlan`，核心字段为 `ProjectID`、`Tables` 和 `Warnings`。
- `PlannedRowCount` 提供 `ProjectTableID`、`TableID`、`ExecutionOrder`、`TargetRows` 和来源摘要。
- 批量生成主循环必须使用最终 `TargetRows`，不得重新解释 `ProjectTable.RowCount` 或倍率约束。

### phase-03-generation-context

- 输出 `GenerationContext`，包含任务、Project、表、字段、关系和 `RuntimeReferenceStore`。
- `ContextTable` 提供 `ProjectTableID`、`TableID`、`ExecutionOrder`、`TargetRows`、行数来源和字段视图。
- `ContextField` 提供字段身份、顺序、逻辑类型、约束摘要、规则摘要和是否需要生成。
- `GeneratorCallInput` 是后续 generator 接入的最小调用输入。
- `ContextReferenceReader` / `RuntimeReferenceStore` 负责当前执行内引用读取与记录。
- 生成上下文本身不实现 batch loop、generator registry 或 writer adapter。

## 设计发现

### 需要新增的 engine 子包

建议新增 `internal/engine/batch` 包，集中表达批量生成主循环，原因：

- 职责位于 engine，独立于 lifecycle 状态机、plan、rowcount、gencontext、generator registry 和 writer adapter。
- 批次调度、行组装和接缝调用具有状态推进性质，不适合放入 domain 或 generator 包。
- 可通过接口约束下游 generator invoker 与 writer seam，保证后续实现可替换。

### 输入模型边界

主循环的最小输入应包含：

- 执行任务和 Project 身份。
- `plan.ExecutionPlan`。
- `rowcount.RowCountPlan`。
- `gencontext.GenerationContext`。
- 批次配置，例如批次大小。
- `GeneratorInvoker` 接口。
- `BatchWriter` 接口。
- 可选进度 sink 或返回式进度摘要。

主循环应先对齐三类计划/上下文：ProjectTable 集合、TableID、ExecutionOrder、TargetRows 和 Project/Task 边界。

### 调度策略

- 表级顺序只来自 `ExecutionPlan.OrderedTables`。
- 每张表的目标行数只来自 `RowCountPlan` / `GenerationContext` 对齐结果。
- 批次范围应使用确定性 `[start, end)` 行序区间。
- 零行表应产生完成统计，但不调用 generator 或 writer。
- 本阶段不做跨表并行或高级调度，避免破坏引用可用性。

### 行和字段值模型

需要区分至少三类字段输出：

- 普通具体值：由引用读取或 generator invoker 返回。
- 显式空值：字段值应写入 NULL 或等价语义。
- 省略/默认值：writer seam 可据此让数据库默认值或自动生成语义生效。

该区分不能折叠为 Go 零值，否则 writer seam 无法可靠表达默认值和可空语义。

### 关系字段优先级

外键和关系字段的处理顺序必须固定：

1. 判断字段是否由关系引用覆盖或要求当前执行引用。
2. 如果关系候选引用可用，使用引用值。
3. 如果引用缺失且字段不能安全跳过，返回阻断错误。
4. 只有非关系优先字段，或关系语义允许普通生成时，才调用 generator invoker。

该规则来自产品约束和 generation context 要求，避免出现随机生成外键导致引用不一致。

### 引用记录时机

为了避免失败批次污染下游引用，引用可分为批次内暂存与提交后可见：

- 行组装阶段识别主键、唯一键和关系候选字段。
- writer seam 成功后将该批次引用记录为可供后续表使用。
- writer seam 失败时丢弃本批次暂存引用或不将其暴露为可用引用。

如果现有 `RuntimeReferenceStore` 只支持直接记录，则 batch 包需要通过受控调用顺序或轻量 pending buffer 保持上述语义，不应改变存储持久化边界。

### Writer seam 边界

Writer seam 只接收已经组装好的批次：任务/表身份、批次范围、字段集合和行集合。它负责后续真实写入、事务、清空策略、驱动错误转换和重试策略。主循环不得打开数据库连接或执行 SQL。

### 进度与结果

需要输出可被 lifecycle 消费的摘要：

- 表开始/完成。
- 批次开始/完成/失败。
- 已完成行数和目标行数。
- 失败范围：表、批次、行、字段。
- 安全错误：错误码、阶段、字段路径、安全消息、阻断标记。

本规格不直接推送 Wails/UI/API 事件，也不持久化执行历史。

## 关键决策

1. **新增 `internal/engine/batch` 包**：主循环作为 engine 独立包，依赖上游计划和上下文输出。
2. **接口化 generator 与 writer**：只定义最小 invoker/writer seam，不实现 registry、built-ins 或真实写入。
3. **严格计划对齐**：生成前校验 ExecutionPlan、RowCountPlan、GenerationContext 的表集合、顺序和目标行数。
4. **关系优先**：关系/外键字段先查引用，缺失时安全失败，不默默退化为随机生成。
5. **writer 成功后引用可见**：避免失败写入批次的生成值成为下游引用。
6. **安全错误模型**：公开错误只包含安全位置和摘要，不包含 SQL、连接信息、规则参数或生成值。

## 风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| 上游 GenerationContext 类型尚未实现 | 实现阶段可能需要适配实际命名 | 设计使用概念契约，任务要求先接入已实现类型再完善测试 |
| 引用记录提交语义与 RuntimeReferenceStore 能力不匹配 | 失败批次可能污染下游引用 | 在 batch 包内引入 pending 引用缓冲或仅在 writer 成功后调用记录接口 |
| 字段默认/省略语义表达不足 | writer 无法区分 NULL 与 DEFAULT | 在 batch value model 中定义显式状态枚举 |
| generator invoker 原始错误泄露敏感数据 | 安全边界破坏 | 错误映射层统一输出固定安全消息和字段路径 |
| 批次大小边界不清 | 生成阶段可能出现无限循环或内存过大 | 输入校验固定正数与安全上限，异常阻断 |

## 不采用方案

- **在 lifecycle 中直接实现主循环**：会混淆状态机与生成调度职责，降低后续测试与替换性。
- **在 generator 包中实现批次调度**：generator 应只负责生成值或 registry，不能拥有表级执行计划和 writer seam。
- **让 writer seam 决定关系引用**：writer 只负责写入，关系值必须在 engine 上下文中确定。
- **生成失败时继续依赖下游表**：会破坏依赖引用完整性；当前阶段采用依赖感知的部分成功模型，只跳过依赖失败表或失败引用的下游表，允许无依赖关系的后续表继续执行。
- **实现并行生成**：高级并行会引入引用可见性和事务语义复杂度，留给后续规格。
