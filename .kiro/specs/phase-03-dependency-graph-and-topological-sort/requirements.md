# Requirements Document

## Introduction

`phase-03-dependency-graph-and-topological-sort` 定义 Phase 3 执行引擎中的表级依赖图与拓扑排序边界。Phase 2 已提供 `ForeignKey`、`TableRelation`、`Project`、`ProjectTable` 和 `ProjectTableRelation` 等领域快照，`phase-03-execution-lifecycle` 已定义生命周期预检与 planner 接缝，但当前仍缺少 engine 侧从 Project 表集合和关系集合构建可执行依赖顺序的能力。

本规格面向 Go 后端 engine/service 层，要求系统能够基于 Project 执行表、物理外键、逻辑关系和 Project 关系实例构建表级依赖图，输出稳定拓扑执行顺序，并将循环依赖、缺失表、重复边、未知关系类型、未知值来源等问题转化为可被生命周期预检消费的安全结果。

## Boundary Context

- **In scope**: Project 表节点建模、物理外键依赖边、逻辑关系依赖边、Project 关系实例依赖边、边来源与外部值来源摘要表达、确定性拓扑排序、循环依赖与缺失节点诊断、生命周期 planner/precheck 接缝输出、安全错误摘要。
- **Out of scope**: 行数规划、倍率推导、生成上下文、字段级生成器选择、生成器注册表、批量生成循环、writer adapter、事务、真实数据库写入、SQL 执行或校验、API/UI/Wails 事件、复杂循环拆解或延迟约束写入策略。
- **Adjacent expectations**: 上游 `phase-03-execution-lifecycle` 聚合本规格输出的预检结果和 planner 结果；下游 `phase-03-row-count-planning` 使用稳定执行顺序进行行数规划；数据库差异只能通过已抽象的 schema/relation/capability 快照进入，不由本规格按数据库类型硬编码分支。

## Requirements

### Requirement 1: Project 表节点输入边界

**Objective:** As a 后端开发人员, I want 系统能够从 Project 执行表快照建立稳定图节点, so that 后续依赖边和拓扑排序具有明确的执行边界。

#### Acceptance Criteria

1. When 调用方提交 Project 执行表集合时, the 执行计划构建器 shall 为每个有效 ProjectTable 创建一个表级依赖图节点。
2. When ProjectTable 节点通过基础校验时, the 执行计划构建器 shall 保留 ProjectTable 身份、Schema Table 身份和稳定排序所需的最小字段。
3. If ProjectTable 缺少 ProjectTable 身份、Project 引用或 Schema Table 引用, then the 执行计划构建器 shall 返回字段级阻断预检错误而不是生成可排序计划。
4. If 同一 Project 输入中出现重复 Schema Table 节点且无法唯一映射执行表, then the 执行计划构建器 shall 返回安全的重复节点错误。
5. The 执行计划构建器 shall 不读取 UI、Wails binding、Vue 页面状态或真实数据库连接作为节点判断依据。

### Requirement 2: 物理外键依赖边构建

**Objective:** As a 后端开发人员, I want 系统能够从物理外键快照构建依赖边, so that 被引用表在外键子表之前生成。

#### Acceptance Criteria

1. When 物理外键的引用表和外键所在表都存在于 Project 执行表集合中时, the 执行计划构建器 shall 创建从被引用表到外键所在表的依赖边。
2. When 物理外键被转换为依赖边时, the 执行计划构建器 shall 保留边来源类型、来源身份、上游表和下游表摘要。
3. If 物理外键引用的父表或子表无法映射到 Project 节点, then the 执行计划构建器 shall 返回字段级预检问题并说明缺失节点边界。
4. If 多个物理外键形成相同上游和下游节点关系, then the 执行计划构建器 shall 对排序边进行去重并保留可诊断的来源摘要。
5. The 执行计划构建器 shall 不按数据库类型硬编码外键方向或排序规则。

### Requirement 3: 逻辑关系与 Project 关系依赖边构建

**Objective:** As a 后端开发人员, I want 系统能够从逻辑关系和 Project 关系实例构建执行内依赖, so that 应用定义关系与物理关系一起参与排序。

#### Acceptance Criteria

1. When 逻辑关系的父表和子表都存在于 Project 执行表集合中时, the 执行计划构建器 shall 创建从父表到子表的依赖边。
2. When Project 关系实例的值来源依赖当前执行中的父表时, the 执行计划构建器 shall 创建从父 ProjectTable 到子 ProjectTable 的依赖边。
3. Where Project 关系实例的值来源来自数据库查询且父 ProjectTable 不在当前 Project 中, the 执行计划构建器 shall 不创建执行内依赖边并 shall 保留安全的外部来源摘要。
4. If Project 关系实例要求当前执行父表但 ParentProjectTableID 缺失或无法映射, then the 执行计划构建器 shall 返回阻断预检错误。
5. If 逻辑关系类型或 Project 关系值来源未知，或关系缺少排序所需的节点身份, then the 执行计划构建器 shall 返回安全预检问题而不是执行 SQL、行数计算或生成数据推断。

### Requirement 4: 稳定拓扑排序输出

**Objective:** As a 后端开发人员, I want 系统输出确定性的表级拓扑执行顺序, so that 相同 Project 快照在不同运行中得到一致计划。

#### Acceptance Criteria

1. When 依赖图无阻断错误且可排序时, the 拓扑排序器 shall 输出每个 ProjectTable 的稳定执行顺序。
2. When 多个节点同时没有未满足依赖时, the 拓扑排序器 shall 使用确定性稳定键选择下一个节点。
3. When 排序完成时, the 拓扑排序器 shall 确保每条依赖边的上游节点出现在下游节点之前。
4. If 输入图包含孤立节点, then the 拓扑排序器 shall 将孤立节点纳入稳定顺序而不是丢弃。
5. The 拓扑排序器 shall 不计算目标行数、批次数量、字段生成顺序或写入策略。

### Requirement 5: 循环、缺失和不可排序诊断

**Objective:** As a 后端开发人员, I want 系统能够在执行前诊断不可排序图, so that 生命周期预检可以阻止无法安全执行的任务。

#### Acceptance Criteria

1. If 依赖图存在循环依赖, then the 拓扑排序器 shall 返回阻断预检错误并提供安全的循环节点摘要。
2. If 依赖边引用缺失节点, then the 执行计划构建器 shall 返回字段级阻断预检错误并保留相关边来源摘要。
3. If 去重后的依赖图仍存在不可排序状态, then the 拓扑排序器 shall 返回统一的不可排序安全错误。
4. If 关系类型未知、值来源未知或不满足排序前置条件, then the 执行计划构建器 shall 将该问题表达为阻断错误或非阻断警告并说明安全原因。
5. The 执行计划构建器 shall 不尝试自动拆解循环、延迟约束写入或修改关系配置来绕过不可排序问题。

### Requirement 6: 生命周期预检与 planner 接缝集成

**Objective:** As a 后端开发人员, I want 依赖计划结果能够被生命周期预检和 planner 接缝消费, so that 执行生命周期可以统一处理计划成功和失败。

#### Acceptance Criteria

1. When 生命周期请求计划阶段预检时, the 依赖计划组件 shall 返回包含通过状态、阻断错误和非阻断警告的预检结果。
2. When 生命周期请求 planner 阶段结果时, the 依赖计划组件 shall 返回排序成功结果或安全失败结果。
3. If 依赖计划存在阻断错误, then the 生命周期预检 shall 能够阻止执行进入运行状态。
4. Where 后续行数规划、生成上下文或批量生成循环需要表顺序时, the 依赖计划组件 shall 提供稳定拓扑结果作为输入边界。
5. The 依赖计划组件 shall 不修改 lifecycle 内部状态枚举、Phase 2 历史状态枚举或 Project 持久化字段语义。

### Requirement 7: 安全错误与敏感信息边界

**Objective:** As a 后端开发人员, I want 依赖图和排序错误保持安全摘要边界, so that 计划失败可诊断但不会泄露数据库密码、用户 SQL 或生成数据内容。

#### Acceptance Criteria

1. If 节点、边、循环或排序失败返回错误, then the 依赖计划组件 shall 只暴露错误码、阶段、字段路径和安全消息。
2. If 关系来源包含 SQL 文本、连接详情、数据库密码或生成数据示例, then the 依赖计划组件 shall 不在公开错误或警告中包含原始内容。
3. When 下游生命周期聚合依赖计划错误时, the 依赖计划组件 shall 提供与生命周期安全错误模型兼容的摘要字段。
4. The 依赖计划组件 shall 通过测试确认公开消息不会包含敏感 SQL、连接字符串、密码或生成数据内容。
5. The 依赖计划组件 shall 不把原始下游错误载荷透传给 API、UI、Wails binding 或历史模型。

### Requirement 8: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定依赖图排序边界, so that 后续 Phase 3 能力可以独立接入而不重写排序规则。

#### Acceptance Criteria

1. The 依赖计划组件 shall 通过单元测试覆盖节点构建、物理外键边、逻辑关系边、Project 关系边、边去重和稳定排序。
2. The 依赖计划组件 shall 通过测试覆盖循环依赖、缺失节点、未知关系类型、未知值来源以及不可排序诊断。
3. The 依赖计划组件 shall 通过接缝测试证明计划结果可以被 lifecycle planner/precheck 边界消费。
4. The 依赖计划组件 shall 通过边界测试确认不依赖 Wails、Vue、前端 API、真实数据库驱动或数据库类型硬编码。
5. The 依赖计划组件 shall 通过边界测试确认没有实现行数规划、生成上下文、生成器注册表、批量生成循环、writer adapter 或真实写入行为。
