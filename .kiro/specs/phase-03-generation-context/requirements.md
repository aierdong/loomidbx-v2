# Requirements Document

## Introduction

`phase-03-generation-context` 定义 Phase 3 执行引擎中的生成上下文边界。Phase 2 已提供 ExecutionTask、Project、Schema、字段约束、字段生成规则和关系配置等持久化模型；`phase-03-dependency-graph-and-topological-sort` 输出稳定表级拓扑计划；`phase-03-row-count-planning` 输出稳定 `RowCountPlan`。当前仍缺少 engine 侧在批量生成前构建执行期只读快照、字段规则视图和运行态引用存储的能力。

本规格面向 Go 后端 engine/service 层，要求系统能够基于 lifecycle 执行输入、dependency `ExecutionPlan`、rowcount `RowCountPlan` 以及 Phase 2 Project / Schema / Rule / Relation 快照构建 `GenerationContext`，为后续 batch generation loop 和 generator framework 提供稳定的数据访问边界，同时禁止上下文直接访问 UI、Wails binding、真实数据库连接、store/facade 或跨任务全局缓存。

## Boundary Context

- **In scope**: 执行任务、Project、Schema、字段规则、拓扑计划和行数计划的只读上下文快照；表级、字段级、关系级、行数级查找；当前执行内已生成主键、唯一键、外键候选值和关系引用的运行态存储；最小生成器调用输入；上下文构建预检和安全错误摘要。
- **Out of scope**: 完整生成器注册表、内置生成器、批量生成循环、writer adapter、事务、真实数据库写入、真实数据库读取已有父表数据、跨执行任务全局缓存、外部数据源生成器上下文、UI/API/Wails DTO、执行进度事件和执行历史持久化。
- **Adjacent expectations**: 上游 lifecycle 提供执行入口和预检聚合；dependency plan 提供稳定 `ExecutionPlan.OrderedTables`；row count planning 提供稳定 `RowCountPlan.Tables`；下游 batch generation loop 只消费本规格输出的上下文和引用存储接口，不复制快照对齐与查找规则。

## Requirements

### Requirement 1: 生成上下文输入边界与上游计划对齐

**Objective:** As a 后端开发人员, I want 生成上下文接收 lifecycle、拓扑计划、行数计划和领域快照的最小输入, so that 后续批量生成拥有一致且可测试的执行期边界。

#### Acceptance Criteria

1. When 调用方提交执行任务快照、Project 表集合、Schema 表字段集合、字段规则集合、关系集合、`ExecutionPlan` 和 `RowCountPlan` 时, the 生成上下文组件 shall 为每个拓扑计划中的执行表建立一个上下文表节点。
2. When 上下文表节点通过基础校验时, the 生成上下文组件 shall 保留 ProjectTable 身份、Schema Table 身份、拓扑执行顺序、目标行数和字段规则查找所需的最小字段。
3. If `ExecutionPlan.OrderedTables` 与 `RowCountPlan.Tables` 的 ProjectTable 集合、TableID 或执行顺序不一致, then the 生成上下文组件 shall 返回字段级阻断预检错误而不是生成部分上下文。
4. If 执行任务或 Project 快照身份与拓扑计划、行数计划的 Project 边界不一致, then the 生成上下文组件 shall 返回安全阻断问题并说明输入边界不一致。
5. The 生成上下文组件 shall 不重新构建依赖图、不重新计算拓扑顺序、不重新推导目标行数、不读取 UI、Wails binding、Vue 页面状态、store/facade 或真实数据库连接作为上下文判断依据。

### Requirement 2: 只读快照与查找接口

**Objective:** As a 后端开发人员, I want GenerationContext 暴露稳定只读快照和查找接口, so that batch loop 和 generator 调用可以读取配置而不依赖领域持久化对象的可变状态。

#### Acceptance Criteria

1. When 上下文构建成功时, the 生成上下文组件 shall 提供任务、Project、表、字段、约束、字段规则、关系、拓扑计划和行数计划的只读快照视图。
2. When 调用方按 ProjectTableID、TableID、ColumnID 或执行顺序查询时, the 生成上下文组件 shall 返回确定性结果或安全的未找到问题。
3. When 调用方查询某表的目标行数时, the 生成上下文组件 shall 返回来自 `RowCountPlan` 的目标行数和来源摘要，而不是重新解释 `ProjectTable.RowCount`。
4. When 调用方查询字段规则时, the 生成上下文组件 shall 保留字段规则配置状态、生成器名称、输出映射类型和参数 JSON 的快照边界。
5. The 生成上下文组件 shall 不把只读快照写回 Project、Schema、Rule、ExecutionTask、ExecutionPlan 或 RowCountPlan 的持久化字段。

### Requirement 3: 字段级生成计划视图

**Objective:** As a 后端开发人员, I want 上下文能够表达每个执行字段的生成输入边界, so that 后续 batch loop 可以按表和字段调用生成器而不重复装配规则。

#### Acceptance Criteria

1. When 表包含字段和可用字段规则时, the 生成上下文组件 shall 为字段建立生成计划视图，包含表身份、字段身份、字段顺序、逻辑类型、约束摘要、规则摘要和是否需要生成的判定输入。
2. When 字段为主键、唯一键、外键、非空、有默认值或允许为空时, the 生成上下文组件 shall 在字段视图中保留约束摘要供后续生成和预检使用。
3. If 必须生成的字段缺少可用字段规则且不能由关系引用、默认值、可空语义或数据库自动生成语义安全跳过, then the 生成上下文组件 shall 返回字段级阻断预检错误。
4. If 字段规则配置状态不可用或需要复核, then the 生成上下文组件 shall 按规则严重性返回阻断错误或警告，并保留安全字段路径。
5. The 生成上下文组件 shall 不实现完整生成器注册表、不校验具体生成器参数 Schema、不调用内置生成器，也不生成示例数据。

### Requirement 4: 运行态引用存储边界

**Objective:** As a 后端开发人员, I want 当前执行内的已生成键值和关系引用由受控 RuntimeReferenceStore 管理, so that 外键填充可以使用上游生成结果而不泄露或持久化生成数据内容。

#### Acceptance Criteria

1. When batch loop 记录当前执行内已生成主键或唯一键值时, the 运行态引用存储 shall 按执行任务、ProjectTable、Column 和行序维度保存可查询引用。
2. When 下游表需要外键候选值时, the 运行态引用存储 shall 能按关系或父表字段返回当前执行内已记录的候选引用集合。
3. When 关系引用来自外部 DB 查询来源时, the 生成上下文组件 shall 只保留安全来源摘要和待外部能力接入标记，不执行 SQL、不读取真实数据库。
4. If 调用方尝试读取尚未记录的上游引用, then the 运行态引用存储 shall 返回安全的缺失引用问题，供 batch loop 决定是否阻断当前表生成。
5. The 运行态引用存储 shall 不跨执行任务共享缓存、不把生成数据内容写入本地历史、不把原始值写入公开错误或日志。

### Requirement 5: 最小生成器调用输入

**Objective:** As a 后端开发人员, I want GenerationContext 提供最小 generator call input, so that Phase 4 生成器接口可以稳定接入而不依赖 engine 内部模型细节。

#### Acceptance Criteria

1. When batch loop 准备生成单个字段值时, the 生成上下文组件 shall 能构造包含任务身份、表身份、字段身份、行序、目标行数、字段逻辑类型、约束摘要、规则摘要和安全引用访问器的最小调用输入。
2. When generator call input 暴露上下文访问能力时, the 生成上下文组件 shall 只暴露只读查询和当前执行引用读取能力，不暴露 store、facade、Wails runtime、数据库连接或可变领域对象。
3. When 字段需要关系值时, the generator call input shall 能读取关系候选引用或返回安全缺失引用问题。
4. If 调用输入无法构造 because 表、字段、规则或行数目标缺失, then the 生成上下文组件 shall 返回字段级阻断错误。
5. The 生成上下文组件 shall 不定义完整 generator registry、不选择具体 generator 实现、不执行随机生成算法、不组织批量循环。

### Requirement 6: 上下文构建结果与 lifecycle 接缝兼容

**Objective:** As a 后端开发人员, I want 上下文构建结果能够被 lifecycle precheck 和 generation 阶段消费, so that 执行生命周期可以统一处理上下文成功、警告和失败。

#### Acceptance Criteria

1. When 生命周期请求生成上下文预检时, the 生成上下文组件 shall 返回包含通过状态、阻断错误和非阻断警告的结果。
2. When 生命周期或 batch loop 请求上下文构建结果时, the 生成上下文组件 shall 返回完整 `GenerationContext` 或安全失败结果。
3. If 上下文构建存在阻断错误, then the lifecycle precheck shall 能够阻止执行进入生成阶段。
4. Where 后续 batch generation loop 需要执行上下文时, the 生成上下文组件 shall 提供按拓扑顺序排列的表视图和按行数计划确定的目标工作量。
5. The 生成上下文组件 shall 不修改 lifecycle 内部状态枚举、Phase 2 历史状态枚举、Project 持久化字段、ExecutionPlan 或 RowCountPlan。

### Requirement 7: 安全错误与敏感信息边界

**Objective:** As a 后端开发人员, I want 上下文错误和引用诊断保持安全摘要边界, so that 失败可诊断但不会泄露连接信息、用户 SQL、规则参数敏感值或生成数据内容。

#### Acceptance Criteria

1. If 输入对齐、快照缺失、字段规则、引用存储或调用输入构建失败返回错误, then the 生成上下文组件 shall 只暴露错误码、阶段、字段路径、安全消息和阻断标记。
2. If 关系来源或规则参数包含 SQL 文本、连接详情、密码、令牌或生成数据示例, then the 生成上下文组件 shall 不在公开错误、警告或引用缺失消息中包含原始内容。
3. When 下游 lifecycle 或 batch loop 聚合上下文错误时, the 生成上下文组件 shall 提供与 lifecycle、dependency plan 和 rowcount 安全错误模型兼容的摘要字段。
4. The 生成上下文组件 shall 通过测试确认公开消息不会包含敏感 SQL、连接字符串、密码、令牌、规则参数原文或生成数据内容。
5. The 生成上下文组件 shall 不把原始下游错误载荷透传给 API、UI、Wails binding、历史模型或日志。

### Requirement 8: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定 GenerationContext 边界, so that 后续 batch loop、writer 和 generator framework 可以独立接入而不重写上下文规则。

#### Acceptance Criteria

1. The 生成上下文组件 shall 通过单元测试覆盖输入对齐、只读快照、查找接口、字段计划视图、引用存储和最小 generator call input。
2. The 生成上下文组件 shall 通过测试覆盖缺失表、缺失字段、缺失规则、计划不一致、引用缺失、不可用规则和安全错误诊断。
3. The 生成上下文组件 shall 通过接缝测试证明上下文结果可以被 lifecycle precheck / generation 阶段和后续 batch loop 边界消费。
4. The 生成上下文组件 shall 通过边界测试确认不依赖 Wails、Vue、前端 API、真实数据库 driver、store、facade 或数据库类型硬编码。
5. The 生成上下文组件 shall 通过边界测试确认没有实现 dependency graph、topological sort、row count solver、generator registry、built-in generator、batch generation loop、writer adapter、transaction 或 real write 行为。
