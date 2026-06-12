# Requirements Document

## Introduction

`phase-03-batch-generation-loop` 定义 Phase 3 执行引擎中的批量生成主循环边界。上游 `phase-03-execution-lifecycle` 提供执行入口、预检聚合和阶段切换；`phase-03-dependency-graph-and-topological-sort` 输出稳定表级 `ExecutionPlan`；`phase-03-row-count-planning` 输出稳定 `RowCountPlan`；`phase-03-generation-context` 输出执行期只读 `GenerationContext`、字段生成视图和 `RuntimeReferenceStore`。当前仍缺少 engine 侧按表、批次和行组装数据、调用最小生成器接缝、优先填充关系字段、记录运行态引用并把批次交给 writer 边界的主循环能力。

本规格面向 Go 后端 engine/service 层，要求系统能够基于 lifecycle 执行输入、`ExecutionPlan`、`RowCountPlan` 和 `GenerationContext` 按拓扑顺序调度表生成，按目标行数切分批次，为每行每字段构造生成输入或关系引用输入，组装批次行数据，调用 writer seam，并将进度、失败和安全错误摘要交回 lifecycle。该主循环不得实现完整生成器注册表、内置生成器、writer adapter 内部、真实数据库写入、UI/API/Wails 事件推送、高级并行调度、外部数据源生成器、事务或清空策略。

## Boundary Context

- **In scope**: 按拓扑计划和行数计划的表级调度；批次切分；行和字段值组装；最小 generator invoker 接口；关系字段从 `GenerationContext` / `RuntimeReferenceStore` 读取候选引用；生成后主键/唯一键/关系引用记录；批次 writer seam 调用；表/批次/行级进度和失败摘要；安全错误边界。
- **Out of scope**: 完整 generator registry、内置 generator、随机算法实现、writer adapter 内部、真实数据库连接与写入、事务和清空策略、UI/API/Wails 事件推送、高级并行调度、外部数据源 generator、依赖图重建、行数重算、生成上下文重建。
- **Adjacent expectations**: 上游 lifecycle 负责执行状态机和预检聚合；dependency plan 负责排序；row count plan 负责目标行数；generation context 负责只读快照、字段计划视图和引用存储；下游 writer adapter 负责真实批量写入语义；后续 generator framework 负责 registry、内置实现和参数 schema。

## Requirements

### Requirement 1: 批量生成输入边界与计划对齐

**Objective:** As a 后端开发人员, I want 批量生成主循环接收 lifecycle、执行计划、行数计划和生成上下文的最小输入, so that 生成阶段只消费已验证的执行期边界。

#### Acceptance Criteria

1. When 调用方提交执行任务身份、`ExecutionPlan`、`RowCountPlan`、`GenerationContext`、generator invoker 和 writer seam 时, the 批量生成主循环 shall 为每个拓扑计划中的执行表建立表级生成工作单元。
2. When 表级工作单元通过基础校验时, the 批量生成主循环 shall 保留 ProjectTable 身份、Schema Table 身份、拓扑执行顺序、目标行数、字段视图和批次调度所需的最小字段。
3. If `ExecutionPlan.OrderedTables`、`RowCountPlan.Tables` 与 `GenerationContext.Tables` 的 ProjectTable 集合、TableID、执行顺序或目标行数不一致, then the 批量生成主循环 shall 返回阻断生成错误而不是生成部分批次。
4. If 执行任务或 Project 身份与计划和上下文边界不一致, then the 批量生成主循环 shall 返回安全阻断问题并说明输入边界不一致。
5. The 批量生成主循环 shall 不重新构建依赖图、不重新计算拓扑顺序、不重新推导目标行数、不重建生成上下文、不读取 UI、Wails binding、Vue 页面状态、store/facade 或真实数据库连接作为生成判断依据。

### Requirement 2: 拓扑表调度与批次切分

**Objective:** As a 后端开发人员, I want 主循环按拓扑顺序和行数目标调度表与批次, so that 上游依赖表先生成且每张表拥有确定工作量。

#### Acceptance Criteria

1. When 生成开始时, the 批量生成主循环 shall 按 `ExecutionPlan.OrderedTables` 的稳定顺序逐表执行，确保上游表在下游表前完成。
2. When 表目标行数大于配置批次大小时, the 批量生成主循环 shall 将表工作量切分为确定性的连续批次范围。
3. When 表目标行数为零时, the 批量生成主循环 shall 跳过字段生成和 writer 调用，并报告该表零行完成。
4. If 批次大小缺失、为负、为零或超出安全边界, then the 批量生成主循环 shall 返回安全阻断错误而不是进入生成阶段。
5. The 批量生成主循环 shall 不实现跨表高级并行调度、不改变拓扑顺序、不将批次数量写回 Project、ExecutionPlan、RowCountPlan 或 GenerationContext。

### Requirement 3: 行组装与字段生成顺序

**Objective:** As a 后端开发人员, I want 主循环为每个批次稳定组装行和字段值, so that writer seam 接收一致、可测试且字段完整的批次数据。

#### Acceptance Criteria

1. When 批次执行时, the 批量生成主循环 shall 为批次范围内每个行序建立行缓冲并按上下文字段顺序处理字段。
2. When 字段视图判定需要生成且不存在关系优先填充值时, the 批量生成主循环 shall 构造最小 `GeneratorCallInput` 并调用 generator invoker 获取字段值。
3. When 字段可以通过数据库默认值、可空语义或数据库自动生成语义安全跳过时, the 批量生成主循环 shall 在行组装中保留可由 writer seam 区分的缺省或省略语义。
4. If 必须生成的字段无法构造调用输入、缺少规则或 generator invoker 返回失败, then the 批量生成主循环 shall 返回字段级阻断错误并标记失败行、字段和表。
5. The 批量生成主循环 shall 不实现具体随机生成算法、不选择 generator 实现、不校验 generator 参数 schema、不生成示例数据。

### Requirement 4: 关系字段优先填充与引用读取

**Objective:** As a 后端开发人员, I want 外键和关系字段优先使用运行态引用而不是普通随机生成, so that 下游表生成结果满足当前执行内关系约束。

#### Acceptance Criteria

1. When 字段视图表明字段参与外键或 Project 关系且需要当前执行引用时, the 批量生成主循环 shall 优先通过 `GenerationContext` 的安全引用访问器读取候选关系值。
2. When 可用关系候选引用存在时, the 批量生成主循环 shall 使用引用值填充关系字段，而不是调用普通 generator invoker 生成该字段。
3. When 关系字段允许外部 DB 查询来源但外部能力尚未接入时, the 批量生成主循环 shall 只保留安全来源摘要并返回可诊断的缺失能力或缺失引用问题，不执行 SQL 或读取真实数据库。
4. If 关系字段缺少必需上游引用且不能安全跳过, then the 批量生成主循环 shall 返回阻断错误，供 lifecycle 阻止当前表及依赖该引用的后续表继续生成。
5. The 批量生成主循环 shall 不把引用原始值写入公开错误、进度事件、历史记录或日志摘要。

### Requirement 5: 运行态引用记录

**Objective:** As a 后端开发人员, I want 主循环在生成批次后记录当前执行内的键值和关系引用, so that 后续表可以读取已生成父表候选值。

#### Acceptance Criteria

1. When 字段值成功生成并属于主键、唯一键或关系引用候选字段时, the 批量生成主循环 shall 通过 `RuntimeReferenceStore` 按任务、ProjectTable、Column 和行序维度记录引用。
2. When 批次 writer seam 成功接受批次时, the 批量生成主循环 shall 将该批次可供下游使用的引用视为当前执行内已记录候选。
3. If 批次生成成功但 writer seam 返回失败, then the 批量生成主循环 shall 不把失败批次的引用暴露为下游可用引用。
4. If 引用记录失败或引用存储返回安全问题, then the 批量生成主循环 shall 返回阻断错误并停止依赖该引用的后续表生成。
5. The 批量生成主循环 shall 不跨执行任务共享引用缓存、不持久化生成数据内容、不在公开错误中包含原始生成值。

### Requirement 6: Writer seam 调用边界

**Objective:** As a 后端开发人员, I want 主循环只通过 writer seam 移交批次行数据, so that 写入适配、事务和真实数据库行为可以独立实现。

#### Acceptance Criteria

1. When 批次行组装完成且无阻断字段错误时, the 批量生成主循环 shall 调用 writer seam 并传递任务身份、表身份、批次范围、字段列集合和行值集合。
2. When writer seam 成功返回时, the 批量生成主循环 shall 标记该批次完成并继续后续批次或表。
3. If writer seam 返回失败, then the 批量生成主循环 shall 将失败映射为安全批次错误并停止受影响表的继续生成。
4. When writer seam 需要表达省略字段、空值或默认值语义时, the 批量生成主循环 shall 传递可区分的行值状态，而不是把所有缺省语义折叠为普通零值。
5. The 批量生成主循环 shall 不打开数据库连接、不执行 SQL、不实现事务、不清空表、不重试数据库写入策略、不解释数据库驱动错误细节。

### Requirement 7: 进度、失败和 lifecycle 接缝兼容

**Objective:** As a 后端开发人员, I want 主循环输出可被 lifecycle 消费的进度和失败摘要, so that 执行生命周期可以统一处理生成阶段成功、失败和中止。

#### Acceptance Criteria

1. When 表、批次或行生成推进时, the 批量生成主循环 shall 输出包含任务、表、批次、目标行数、已完成行数和阶段的进度摘要。
2. When 生成阶段完成全部表时, the 批量生成主循环 shall 返回成功结果和按拓扑顺序汇总的表级生成统计。
3. If 任一阻断错误发生, then the 批量生成主循环 shall 返回安全失败结果，包含错误码、阶段、字段路径、阻断标记和安全消息。
4. When lifecycle 聚合生成结果时, the 批量生成主循环 shall 提供与 lifecycle precheck/generation 失败模型兼容的摘要字段。
5. The 批量生成主循环 shall 不修改 lifecycle 内部状态枚举、不直接推送 UI/API/Wails 事件、不写执行历史持久化模型。

### Requirement 8: 安全错误与敏感信息边界

**Objective:** As a 后端开发人员, I want 生成失败、引用缺失和 writer seam 失败保持安全摘要边界, so that 失败可诊断但不会泄露连接信息、SQL、规则参数或生成数据内容。

#### Acceptance Criteria

1. If 输入对齐、批次调度、字段生成、引用读取、引用记录或 writer seam 调用失败, then the 批量生成主循环 shall 只暴露错误码、阶段、字段路径、安全消息、阻断标记和安全范围摘要。
2. If generator invoker、引用访问器或 writer seam 返回原始错误载荷, then the 批量生成主循环 shall 过滤 SQL 文本、连接详情、密码、令牌、规则参数原文和生成数据示例。
3. When 公开错误需要标识失败位置时, the 批量生成主循环 shall 使用任务、ProjectTable、Table、Column、批次和行序等安全标识，而不是原始数据值。
4. The 批量生成主循环 shall 通过测试确认公开消息不会包含敏感 SQL、连接字符串、密码、令牌、规则参数原文或生成数据内容。
5. The 批量生成主循环 shall 不把原始下游错误载荷透传给 API、UI、Wails binding、历史模型或日志。

### Requirement 9: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定 batch generation loop 边界, so that 后续 writer、generator framework 和结果模型可以独立接入而不重写主循环规则。

#### Acceptance Criteria

1. The 批量生成主循环 shall 通过单元测试覆盖输入对齐、拓扑表调度、批次切分、零行表、行组装、字段生成调用和 writer seam 调用。
2. The 批量生成主循环 shall 通过测试覆盖关系字段优先、引用缺失、引用记录、generator 失败、writer 失败和安全错误诊断。
3. The 批量生成主循环 shall 通过接缝测试证明结果可以被 lifecycle generation 阶段和后续 writer adapter 边界消费。
4. The 批量生成主循环 shall 通过边界测试确认不依赖 Wails、Vue、前端 API、真实数据库 driver、store、facade 或数据库产品名称硬编码。
5. The 批量生成主循环 shall 通过边界测试确认没有实现 dependency graph、topological sort、row count solver、generation context builder、generator registry、built-in generator、writer adapter 内部、transaction、clear strategy 或 real write 行为。
