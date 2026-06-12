# Requirements Document

## Introduction

`phase-03-row-count-planning` 定义 Phase 3 执行引擎中的表级目标行数规划边界。Phase 2 已提供 Project 表配置、Project 关系实例和 Schema 关系倍率快照，`phase-03-dependency-graph-and-topological-sort` 输出稳定拓扑顺序，`phase-03-execution-lifecycle` 提供预检和 planner 接缝，但当前仍缺少 engine 侧在生成前计算每张表目标行数、校验关系倍率约束并暴露不可满足场景的能力。

本规格面向 Go 后端 engine/service 层，要求系统能够基于 ProjectTable 行数配置、ProjectTableRelation / TableRelation 倍率约束和依赖排序结果输出稳定 `RowCountPlan`，并将缺失配置、负数边界、零值边界、倍率冲突、不可推导和不可满足关系转化为可被生命周期预检消费的安全结果。

## Boundary Context

- **In scope**: 读取 Project 表级行数配置、解释显式 / 动态 / 默认行数语义、基于拓扑顺序和关系倍率推导或校验目标行数、输出每表目标行数和来源、诊断缺失配置、冲突倍率、负数 / 零值边界和不可满足关系、生命周期 precheck / planner 兼容结果、安全错误摘要。
- **Out of scope**: 依赖图构建、拓扑排序、生成上下文、字段级生成器选择、生成器注册表、批量生成循环、writer adapter、事务、真实数据库写入、真实数据生成、SQL 执行或校验、UI/API/Wails 事件、性能压测级分片计划、字段唯一性容量估算或复杂统计分布建模。
- **Adjacent expectations**: 上游 `phase-03-dependency-graph-and-topological-sort` 提供稳定 `ExecutionPlan.OrderedTables`；上游 `phase-03-execution-lifecycle` 聚合本规格输出的预检和 planner 结果；下游 `phase-03-generation-context` 和 `phase-03-batch-generation-loop` 只消费本规格输出的目标行数，不复制推导规则。

## Requirements

### Requirement 1: 行数规划输入边界与拓扑对齐

**Objective:** As a 后端开发人员, I want 行数规划组件接收 Project 配置和拓扑计划的最小快照, so that 目标行数计算具有明确且稳定的执行边界。

#### Acceptance Criteria

1. When 调用方提交 Project 表集合、关系集合和依赖排序结果时, the 行数规划组件 shall 为每个拓扑计划中的 ProjectTable 建立一个行数规划节点。
2. When 行数规划节点通过基础校验时, the 行数规划组件 shall 保留 ProjectTable 身份、Schema Table 身份、拓扑执行顺序、原始行数配置和后续来源标记所需的最小字段。
3. If 拓扑计划引用的 ProjectTable 无法在 Project 表集合中找到, then the 行数规划组件 shall 返回字段级阻断预检错误而不是生成部分行数计划。
4. If Project 表集合中存在不属于拓扑计划的执行表, then the 行数规划组件 shall 返回安全预检问题并说明输入边界不一致。
5. The 行数规划组件 shall 不重新构建依赖图、不重新计算拓扑顺序、不读取 UI、Wails binding、Vue 页面状态或真实数据库连接作为行数判断依据。

### Requirement 2: 表级目标行数配置解释

**Objective:** As a 后端开发人员, I want 系统能够解释 ProjectTable 的显式、动态和默认行数配置, so that 后续推导能够区分用户意图和可推导空值。

#### Acceptance Criteria

1. When ProjectTable 提供非空且非负的行数配置时, the 行数规划组件 shall 将该配置作为显式目标行数并保留来源摘要。
2. When ProjectTable 的行数配置为显式 `0` 时, the 行数规划组件 shall 将目标行数解释为零行而不是动态缺失。
3. When ProjectTable 的行数配置为空且可由关系倍率推导时, the 行数规划组件 shall 将该节点标记为待推导目标行数。
4. If ProjectTable 的行数配置为负数或超出安全整数边界, then the 行数规划组件 shall 返回字段级阻断预检错误。
5. The 行数规划组件 shall 不将动态空值写回 ProjectTable，也不得修改 Phase 2 Project 持久化字段语义。

### Requirement 3: 关系倍率约束读取与方向语义

**Objective:** As a 后端开发人员, I want 系统能够从 Project 关系实例和 Schema 逻辑关系读取倍率约束, so that 父子表和关联表数量关系在生成前被校验。

#### Acceptance Criteria

1. When Parent/Child 关系的父表和子表都存在于拓扑计划中时, the 行数规划组件 shall 使用父表目标行数与 `multiplierMin` / `multiplierMax` 建立子表目标行数约束。
2. When BaseTable/JoinTable 关系的基础表和关联表都存在于拓扑计划中时, the 行数规划组件 shall 使用基础表目标行数与倍率范围建立关联表目标行数约束。
3. When ProjectTableRelation 覆盖或实例化 Schema 关系倍率时, the 行数规划组件 shall 优先使用 Project 关系实例中与当前执行相关的倍率配置。
4. If 关系倍率出现负数、`multiplierMin` 大于 `multiplierMax` 或非法零值组合, then the 行数规划组件 shall 返回安全阻断预检错误。
5. The 行数规划组件 shall 不按数据库产品名称硬编码倍率语义或关系方向。

### Requirement 4: 目标行数推导与冲突校验

**Objective:** As a 后端开发人员, I want 系统能够根据显式配置和关系倍率推导动态行数, so that 每张执行表在生成前拥有确定目标工作量。

#### Acceptance Criteria

1. When 父表目标行数已知且子表为空配置时, the 行数规划组件 shall 在倍率范围可确定时推导子表目标行数并标记为关系推导来源。
2. When 子表已有显式目标行数且父表目标行数已知时, the 行数规划组件 shall 校验子表目标行数是否落入关系倍率允许范围。
3. When 多个关系同时约束同一动态表时, the 行数规划组件 shall 合并约束范围并在存在唯一可满足结果时输出确定目标行数。
4. If 多个显式配置或关系约束无法形成可满足目标行数, then the 行数规划组件 shall 返回阻断预检错误并提供安全约束摘要。
5. The 行数规划组件 shall 不选择随机目标行数、不生成示例数据、不估算字段唯一性容量，也不创建批次数量计划。

### Requirement 5: 缺失配置、零值和不可满足关系诊断

**Objective:** As a 后端开发人员, I want 系统能够在执行前诊断行数规划失败原因, so that 生命周期预检可以阻止无法安全生成的任务。

#### Acceptance Criteria

1. If 表缺少显式行数且无法通过任何关系推导, then the 行数规划组件 shall 返回字段级阻断预检错误或按配置规则返回安全默认来源。
2. If 父表目标行数为零且关系倍率要求至少一个下游行, then the 行数规划组件 shall 返回不可满足关系阻断错误。
3. If 父表目标行数大于零且关系倍率固定为零, then the 行数规划组件 shall 允许子表零行结果并标记安全来源。
4. If 推导过程中出现除零、溢出、范围收敛为空或无法确定唯一目标行数, then the 行数规划组件 shall 返回统一的不可规划安全错误。
5. The 行数规划组件 shall 不尝试修改用户配置、拆分关系、降低倍率或绕过不可满足约束。

### Requirement 6: 生命周期预检与 planner 接缝结果

**Objective:** As a 后端开发人员, I want 行数规划结果能够被生命周期预检和 planner 阶段消费, so that 执行生命周期可以统一处理行数规划成功和失败。

#### Acceptance Criteria

1. When 生命周期请求行数规划预检时, the 行数规划组件 shall 返回包含通过状态、阻断错误和非阻断警告的预检结果。
2. When 生命周期请求 planner 阶段行数结果时, the 行数规划组件 shall 返回排序一致的 `RowCountPlan` 或安全失败结果。
3. If 行数规划存在阻断错误, then the 生命周期预检 shall 能够阻止执行进入生成阶段。
4. Where 后续生成上下文或批量生成循环需要目标行数时, the 行数规划组件 shall 提供按拓扑顺序排列的每表目标行数作为输入边界。
5. The 行数规划组件 shall 不修改 lifecycle 内部状态枚举、Phase 2 历史状态枚举或 Project 持久化字段语义。

### Requirement 7: 安全错误与敏感信息边界

**Objective:** As a 后端开发人员, I want 行数规划错误保持安全摘要边界, so that 规划失败可诊断但不会泄露数据库密码、用户 SQL 或生成数据内容。

#### Acceptance Criteria

1. If 输入、配置、关系约束或推导失败返回错误, then the 行数规划组件 shall 只暴露错误码、阶段、字段路径、安全消息和阻断标记。
2. If 关系来源包含 SQL 文本、连接详情、数据库密码或生成数据示例, then the 行数规划组件 shall 不在公开错误或警告中包含原始内容。
3. When 下游生命周期聚合行数规划错误时, the 行数规划组件 shall 提供与生命周期安全错误模型兼容的摘要字段。
4. The 行数规划组件 shall 通过测试确认公开消息不会包含敏感 SQL、连接字符串、密码或生成数据内容。
5. The 行数规划组件 shall 不把原始下游错误载荷透传给 API、UI、Wails binding 或历史模型。

### Requirement 8: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定行数规划边界, so that 后续 Phase 3 能力可以独立接入而不重写行数规则。

#### Acceptance Criteria

1. The 行数规划组件 shall 通过单元测试覆盖输入对齐、显式行数、动态空值、显式零行、倍率范围、关系推导和稳定输出顺序。
2. The 行数规划组件 shall 通过测试覆盖缺失配置、负数边界、非法倍率、倍率冲突、零值不可满足和不可规划诊断。
3. The 行数规划组件 shall 通过接缝测试证明规划结果可以被 lifecycle precheck / planner 边界消费。
4. The 行数规划组件 shall 通过边界测试确认不依赖 Wails、Vue、前端 API、真实数据库驱动或数据库类型硬编码。
5. The 行数规划组件 shall 通过边界测试确认没有实现生成上下文、生成器注册表、批量生成循环、writer adapter、事务或真实写入行为。
