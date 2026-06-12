# Requirements Document

## Introduction

`phase-03-batch-writer-adapter` 定义 Phase 3 执行引擎中的批量写入适配边界。上游 `phase-03-batch-generation-loop` 已在 engine 侧组装表级批次行数据，并通过 writer seam 传递任务身份、表身份、批次范围、列集合和可区分的字段值状态；`phase-01-database-dialect-interface` 已定义 Adapter、Dialect、Capabilities、batch insert statement 和 fake adapter 边界。当前仍缺少 engine 侧面向数据库写入的稳定适配接口、写入策略表达、清空顺序、能力检查、错误归一化和 mock writer 支持。

本规格面向 Go 后端 engine/dbx 接缝层，要求系统能够接收 batch generation loop 输出的批次写入负载，基于数据库 Adapter / Dialect / Capabilities 完成写入前能力校验、按表清空策略排序、批量写入请求构造、writer 结果汇总和数据库错误安全归一化。该边界不得实现完整数据库方言 SQL、连接管理、凭据存储、跨数据库迁移、UI/API 执行历史查询、高级重试/断点续跑/幂等策略、复杂可观测性管道或按数据库产品名称硬编码分支。

## Boundary Context

- **In scope**: engine 侧 `BatchWriter` 适配接口；写入请求、响应、结果和错误模型；批次行值到 DBX dialect insert request 的边界映射；Project 表级清空策略在写入前的 engine seam 顺序表达；基于 `Capabilities` 的事务、批量写入、参数上限和批次限制检查；事务/非事务执行计划的能力前置判断；mock writer / fake executor 支持；数据库错误安全归一化；writer 统计和 lifecycle-compatible 摘要。
- **Out of scope**: 完整 MySQL/PostgreSQL/其他数据库 SQL 实现；真实数据库连接池和连接生命周期；凭据存储；真实事务提交/回滚实现细节；跨数据库迁移或同步；高级重试、断点续跑、幂等写入；UI/API/Wails 执行历史查询；复杂日志、指标和追踪管道；generator registry 或 batch generation loop 内部。
- **Adjacent expectations**: 上游 batch loop 负责行组装和 writer seam 调用；DBX Adapter / Dialect / Capabilities 负责数据库差异表达；后续 execution result/error model 汇总最终执行结果；后续真实 adapter 实现负责 SQL 构建细节和数据库驱动调用。

## Requirements

### Requirement 1: 批量写入适配输入边界

**Objective:** As a 后端开发人员, I want writer adapter 接收 batch loop 的最小写入负载和数据库适配能力, so that engine 写入阶段只消费已组装且已验证的批次数据。

#### Acceptance Criteria

1. When 调用方提交任务身份、Project 身份、表身份、批次范围、列集合、行值集合、writer 策略、DBX adapter/dialect/capabilities 边界和安全执行器时, the 批量写入适配器 shall 构造表级写入工作单元。
2. When 写入工作单元通过基础校验时, the 批量写入适配器 shall 保留 TaskID、ProjectID、ProjectTableID、TableID、批次范围、列顺序、字段值状态和写入策略所需的最小字段。
3. If 批次负载缺少必需身份、列集合为空但行值需要写入、行值列与列集合不一致或批次范围非法, then the 批量写入适配器 shall 返回阻断写入错误而不是调用数据库边界。
4. If 调用方提供的 adapter、dialect、capabilities 或执行器缺失, then the 批量写入适配器 shall 返回安全阻断问题并说明写入边界不完整。
5. The 批量写入适配器 shall 不重新组装生成行、不读取 GenerationContext、不重新计算拓扑顺序或行数、不访问 UI、Wails binding、Vue 页面状态、store、facade 或凭据存储作为写入判断依据。

### Requirement 2: Capabilities 驱动的写入能力校验

**Objective:** As a 后端开发人员, I want writer adapter 基于 DBX Capabilities 判断事务和批量写入可用性, so that engine 不按数据库类型硬编码写入策略。

#### Acceptance Criteria

1. When 写入策略要求事务包裹时, the 批量写入适配器 shall 检查 `Capabilities` 中的事务能力并在不支持时返回阻断错误。
2. When 写入策略要求批量插入或多行 statement 时, the 批量写入适配器 shall 检查 `Capabilities` 中的批量插入能力、最大批量行数和最大参数数限制。
3. If 批次行数或参数数量超过 capabilities 安全限制, then the 批量写入适配器 shall 返回可诊断的安全错误或要求上游重新切分，而不是自行改变 batch loop 已提交的批次语义。
4. When 能力只支持非事务或单语句写入时, the 批量写入适配器 shall 输出明确的能力来源摘要并按配置允许的降级边界执行或阻断。
5. The 批量写入适配器 shall 不包含 `mysql`、`postgres`、`sqlite` 等数据库产品名称分支来决定写入、事务、清空或错误处理策略。

### Requirement 3: 字段值状态到写入请求映射

**Objective:** As a 后端开发人员, I want writer adapter 保留字段值的 present/null/omitted/default 语义, so that dialect 和下游 executor 可以区分真实值、空值和数据库默认值。

#### Acceptance Criteria

1. When 字段值状态为 present 时, the 批量写入适配器 shall 将该字段映射为 dialect insert request 的列值并保留参数值。
2. When 字段值状态为 explicit null 时, the 批量写入适配器 shall 将该字段映射为可区分的 NULL 写入语义。
3. When 字段值状态为 omitted/default 时, the 批量写入适配器 shall 从对应行的显式值集合中省略该字段或使用 dialect 支持的默认值表达，不把它折叠为 Go 零值。
4. If 同一批次中不同的行需要不同的省略字段集合且当前 dialect/request 形态无法安全表达, then the 批量写入适配器 shall 返回安全阻断错误或按确定性规则拆分为多个 statement 请求。
5. The 批量写入适配器 shall 不解释 generator 参数、不修改生成值、不持久化生成数据内容、不在公开错误中包含原始字段值。

### Requirement 4: 清空策略和写入顺序边界

**Objective:** As a 后端开发人员, I want writer adapter 在 engine seam 表达表级清空策略和写入顺序, so that 清空与写入规则可测试且不污染 batch loop。

#### Acceptance Criteria

1. When Project 表配置要求写入前清空目标表时, the 批量写入适配器 shall 在该表第一个写入批次前执行清空 seam 或返回能力缺失问题。
2. When 多表按拓扑顺序写入时, the 批量写入适配器 shall 接收并保持上游 batch loop 的表顺序，不重新排序表或批次。
3. When 清空策略涉及外键约束、事务或级联能力时, the 批量写入适配器 shall 只通过 capabilities 和清空 seam 的安全结果判断是否允许继续。
4. If 清空 seam 失败, then the 批量写入适配器 shall 返回表级阻断错误并阻止该表后续批次写入。
5. The 批量写入适配器 shall 不实现具体 TRUNCATE/DELETE SQL、不关闭外键检查、不跨表重排清空、不把清空策略写回 Project 持久化模型。

### Requirement 5: Dialect statement 构建与 executor seam

**Objective:** As a 后端开发人员, I want writer adapter 通过 DBX Dialect 构建写入语句并通过窄 executor seam 执行, so that SQL 差异和数据库驱动调用保持在 adapter 边界。

#### Acceptance Criteria

1. When 批次负载和能力校验通过时, the 批量写入适配器 shall 构造 dialect insert request 并调用 `Dialect.BuildInsert` 或等价 DBX 写入构建接口。
2. When dialect 返回一条或多条 statement 时, the 批量写入适配器 shall 通过窄 executor seam 按确定性顺序移交 statement 和参数。
3. If dialect 不支持当前 insert 形态, then the 批量写入适配器 shall 将 typed dialect error 映射为安全写入错误。
4. If executor 返回失败, then the 批量写入适配器 shall 将数据库/驱动错误归一化为安全批次错误并停止受影响批次。
5. The 批量写入适配器 shall 不直接导入真实数据库 driver、不打开连接、不管理连接池、不拼接数据库特定 SQL、不解释 SQL 文本内容。

### Requirement 6: 事务边界与结果提交语义

**Objective:** As a 后端开发人员, I want writer adapter 明确事务能力和批次结果提交边界, so that lifecycle 可以理解写入成功、失败和部分接受范围。

#### Acceptance Criteria

1. When 写入策略启用事务且 capabilities 支持事务时, the 批量写入适配器 shall 通过 transaction seam 表达 begin/commit/rollback 边界。
2. When 批次所有 statement 执行成功时, the 批量写入适配器 shall 返回 accepted rows、statement count 摘要和成功状态。
3. If 任一 statement 在事务内失败, then the 批量写入适配器 shall 请求 rollback seam 并返回安全失败结果。
4. If 非事务模式下发生失败且部分 statement 已执行, then the 批量写入适配器 shall 返回安全的部分接受摘要和阻断错误，供后续结果模型处理。
5. The 批量写入适配器 shall 不实现高级重试、断点续跑、幂等键、补偿写入或跨批次恢复策略。

### Requirement 7: Mock writer 和可测试接缝

**Objective:** As a 后端开发人员, I want writer adapter 提供 mock/fake 实现与 deterministic 测试接缝, so that batch loop 和 lifecycle 可以无真实数据库验证写入路径。

#### Acceptance Criteria

1. When 测试需要验证 batch loop writer seam 时, the 批量写入适配器 shall 提供可配置成功、失败、能力缺失和部分接受场景的 fake writer。
2. When 测试需要验证 dialect/executor 调用时, the 批量写入适配器 shall 允许使用 fake dialect、fake executor、fake transaction 和 fake clear seam 记录调用顺序。
3. If fake writer 被配置为失败, then it shall 返回与真实 writer 相同形态的安全错误摘要。
4. When fake writer 成功时, it shall 返回 deterministic accepted rows 和 statement count，便于上游统计断言。
5. The 批量写入适配器 shall 不要求测试连接真实数据库、不要求真实凭据、不依赖网络或数据库容器。

### Requirement 8: 安全错误、数据库错误归一化与敏感信息边界

**Objective:** As a 后端开发人员, I want writer adapter 过滤数据库和下游错误中的敏感内容, so that 写入失败可诊断但不会泄露连接信息、SQL、参数或生成数据。

#### Acceptance Criteria

1. If 输入校验、能力校验、清空 seam、dialect 构建、executor 调用、transaction seam 或结果汇总失败, then the 批量写入适配器 shall 只暴露错误码、阶段、字段路径、安全消息、阻断标记和安全范围摘要。
2. If dialect、executor、transaction 或 clear seam 返回原始错误载荷, then the 批量写入适配器 shall 过滤 SQL 文本、连接详情、DSN、密码、令牌、参数值和生成数据内容。
3. When 公开错误需要定位失败位置时, the 批量写入适配器 shall 使用任务、ProjectTable、Table、Column、批次、statement index 和行序等安全标识。
4. The 批量写入适配器 shall 通过测试确认公开消息不会包含敏感 SQL、连接字符串、密码、令牌、写入参数、生成值或原始 driver 错误细节。
5. The 批量写入适配器 shall 不把原始下游错误载荷透传给 lifecycle、API、UI、Wails binding、历史模型或日志。

### Requirement 9: 边界验证与未来能力隔离

**Objective:** As a 后端开发人员, I want 本规格通过测试固定 writer adapter 边界, so that 后续真实数据库 adapter、结果模型和集成测试可以独立接入。

#### Acceptance Criteria

1. The 批量写入适配器 shall 通过单元测试覆盖输入校验、capabilities 检查、字段值状态映射、dialect request 构建、executor 调用和结果汇总。
2. The 批量写入适配器 shall 通过测试覆盖清空策略、事务成功、事务失败回滚、非事务部分接受、dialect 失败、executor 失败和安全错误诊断。
3. The 批量写入适配器 shall 通过接缝测试证明结果可以被 batch generation loop writer seam 和 lifecycle generation/writer 阶段消费。
4. The 批量写入适配器 shall 通过边界测试确认不依赖 Wails、Vue、前端 API、store、facade、真实数据库 driver、连接管理或凭据存储。
5. The 批量写入适配器 shall 通过边界测试确认没有实现完整数据库方言 SQL、跨数据库迁移、generator registry、batch generation loop 内部、高级重试/断点续跑/幂等策略、执行历史查询 API 或复杂 observability pipeline。
