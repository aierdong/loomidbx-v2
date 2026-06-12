# Implementation Plan

> 任务边界约定：每个实现任务继承最近章节的边界；子任务上的显式 `_Boundary:` 和 `_Depends:` 行用于收窄该章节边界，并在存在时优先生效。

## Tasks

- [ ] 1. 创建 writer 包基础模型与安全错误类型
  - 新增 `internal/engine/writer` 包的写入输入、策略、结果、错误码、阶段、字段路径和安全范围模型。
  - 定义 `WriteInput`、`WriteRequest`、`WriterStrategy`、`WriteResult`、`WriterIssue` 和安全 scope 结构。
  - 实现安全错误构造器，确保公开消息只包含错误码、阶段、字段路径、安全消息、阻断标记和安全范围。
  - _Requirements: 1.1, 1.2, 8.1, 8.3_
  - _Boundary: 不引入 UI/API/Wails/store/facade/driver 依赖，不修改 lifecycle 或 batch 状态模型。_
  - _Depends: 无_

- [ ] 2. 实现 batch payload 到 write request 的输入校验
  - 校验 TaskID、ProjectID、ProjectTableID、TableID、批次范围、列集合和行值列一致性。
  - 校验 adapter、dialect、capabilities、executor 和 writer strategy 的必需边界。
  - 对非法输入返回阻断 `WriterIssue`，不调用 dialect、clear、transaction 或 executor seam。
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - _Boundary: 不重新组装生成行，不读取 GenerationContext、store、facade 或凭据存储。_
  - _Depends: 1_

- [ ] 3. 实现 capability-first 写入策略校验
  - 基于 `Capabilities` 校验事务能力、批量插入能力、最大批量行数和最大参数数。
  - 对事务必需但不支持、批量写入不支持、批次或参数超限返回安全阻断错误。
  - 覆盖允许降级和禁止降级的策略分支，但不按数据库类型分支。
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_
  - _Boundary: 不包含 mysql/postgres/sqlite 等数据库产品名称业务分支，不自行重切上游批次。_
  - _Depends: 1, 2_

- [ ] 4. 实现字段值状态映射模型
  - 定义 writer 内部 `WriteRow`、`WriteValue`、`WriteValueState` 和列元数据映射结构。
  - 将 batch loop 的 present/null/omitted_default 状态转换为 writer 内部状态。
  - 确保 null 与 omitted/default 不被折叠为 Go 零值或普通 nil。
  - _Requirements: 3.1, 3.2, 3.3, 3.5_
  - _Boundary: 不修改生成值，不解释 generator 参数，不公开字段原始值。_
  - _Depends: 1, 2_

- [ ] 5. 实现 insert request 分组和默认值表达边界
  - 将字段状态映射为 dialect insert request 的列集合和行值集合。
  - 对异构 omitted/default 字段集合按稳定 key 分组，或在无法安全表达时返回阻断错误。
  - 为 dialect 不支持的默认值表达保留安全错误路径。
  - _Requirements: 3.1, 3.2, 3.3, 3.4_
  - _Boundary: 不手写数据库特定 SQL，不把所有缺省语义折叠为普通零值。_
  - _Depends: 4_

- [ ] 6. 定义并实现 clear seam 协调器
  - 定义 `ClearExecutor`、`ClearScope`、`ClearResult` 和表级清空状态跟踪。
  - 在 ProjectTable 当前 writer session 的首批写入前按策略调用 clear seam。
  - 对 clear seam 失败或能力缺失返回表级阻断错误。
  - _Requirements: 4.1, 4.3, 4.4_
  - _Boundary: 不实现 TRUNCATE/DELETE SQL，不关闭外键检查，不跨表重排清空，不写回 Project。_
  - _Depends: 1, 3_

- [ ] 7. 保持上游表和批次写入顺序
  - 确保 writer adapter 按调用顺序处理批次，不重新排序表、批次或 statement group。
  - 为 clear-before-write、dialect-build、execute 的调用顺序提供可断言记录点。
  - 覆盖多批次同表只清空一次、后续批次直接写入的测试。
  - _Requirements: 4.1, 4.2, 4.5_
  - _Boundary: 不实现跨表高级清空排序或依赖图重排。_
  - _Depends: 6_

- [ ] 8. 实现 dialect insert request 构建包装
  - 将 insert request group 转换为 DBX `dialect.InsertRequest` 或等价接口输入。
  - 调用 `Dialect.BuildInsert` 并保留 statement 顺序。
  - 将 dialect unsupported 或 typed error 映射为安全 `WriterIssue`。
  - _Requirements: 5.1, 5.3_
  - _Boundary: 不拼接数据库特定 SQL，不解释 SQL 文本，不导入真实数据库 driver。_
  - _Depends: 5_

- [ ] 9. 定义并实现 statement executor seam
  - 定义 `StatementExecutor` 和 `StatementResult`，执行 dialect 返回的 statement。
  - 按确定性顺序移交 statement，汇总 accepted rows 和 statement count。
  - 对 executor 失败返回安全批次错误并保留 statement index scope。
  - _Requirements: 5.2, 5.4, 5.5, 6.2_
  - _Boundary: 不打开连接，不管理连接池，不输出 SQL 或参数值。_
  - _Depends: 8_

- [ ] 10. 定义并实现 transaction seam 协调
  - 定义 `TransactionFactory`、`Transaction`、事务内 executor 获取和 begin/commit/rollback 边界。
  - 在事务策略启用且能力支持时包裹 statement 执行。
  - 任一 statement 失败时请求 rollback 并返回安全失败结果。
  - _Requirements: 6.1, 6.3_
  - _Boundary: 不实现真实 driver transaction，不实现 savepoint 或跨批次事务恢复。_
  - _Depends: 3, 9_

- [ ] 11. 实现非事务部分接受摘要
  - 在非事务策略下记录已成功执行 statement 的 accepted rows 和 statement count。
  - executor 失败时返回 `PartialAccepted=true` 和安全阻断错误。
  - 确认部分接受摘要不包含 SQL、参数或生成值。
  - _Requirements: 6.2, 6.4, 8.1, 8.3_
  - _Boundary: 不实现补偿写入、重试、断点续跑或幂等恢复。_
  - _Depends: 9_

- [ ] 12. 实现 BatchWriterAdapter 协调入口
  - 串联输入校验、capability gate、clear coordinator、value mapper、dialect builder、transaction/executor 和 result mapper。
  - 实现与 batch loop `BatchWriter` seam 兼容的 `WriteBatch` 行为。
  - 成功时返回 accepted rows、statement count 和安全统计。
  - _Requirements: 1.1, 4.2, 5.1, 5.2, 6.2, 9.3_
  - _Boundary: 不将 writer 逻辑并入 batch generation loop，不触发 RuntimeReferenceStore。_
  - _Depends: 2, 3, 6, 8, 10, 11_

- [ ] 13. 实现数据库错误安全归一化
  - 统一过滤 dialect、executor、transaction 和 clear seam 原始错误中的 SQL、连接字符串、DSN、密码、token、参数值和生成数据。
  - 使用固定安全消息和安全 scope 生成公开 `WriterIssue`。
  - 覆盖 raw driver-like error、SQL-like error 和 credential-like error 的过滤。
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Boundary: 不透传原始下游错误载荷给 lifecycle、API、UI、Wails、历史模型或日志。_
  - _Depends: 1, 8, 9, 10, 11_

- [ ] 14. (P) 实现 fake writer 和 seam 测试替身
  - 提供 fake batch writer、fake dialect、fake executor、fake transaction、fake clear executor。
  - 支持配置成功、失败、能力缺失、事务失败、清空失败和部分接受场景。
  - 记录调用顺序、statement count、accepted rows 和最后一次安全 scope。
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  - _Boundary: 不连接真实数据库，不要求真实凭据、网络或容器。_
  - _Depends: 1_

- [ ] 15. 编写输入、capability 和映射单元测试
  - 覆盖有效/非法 payload、缺失 seam、事务/批量能力、最大行数/参数限制。
  - 覆盖 present/null/omitted_default 映射、异构省略字段分组和无法表达时阻断。
  - 确认源码和测试不按数据库产品名称决定策略。
  - _Requirements: 9.1, 2.5, 3.4_
  - _Boundary: 使用 fake DBX 能力和 fake payload，不访问真实数据库或凭据。_
  - _Depends: 3, 5, 14_

- [ ] 16. 编写清空、dialect、executor 和事务测试
  - 覆盖首批清空、非首批不重复清空、清空失败阻断。
  - 覆盖 dialect BuildInsert 成功/失败、executor 成功/失败、事务成功、失败 rollback、commit/rollback 失败安全化。
  - 覆盖非事务部分接受摘要。
  - _Requirements: 9.2, 4.1, 4.4, 5.3, 5.4, 6.3, 6.4_
  - _Boundary: 使用 fake dialect/executor/transaction/clear，不执行真实 SQL。_
  - _Depends: 6, 8, 10, 11, 14_

- [ ] 17. 编写 batch loop 和 lifecycle 接缝兼容测试
  - 验证 `BatchWriterAdapter` 可作为 batch loop writer seam 使用并返回兼容 accepted rows。
  - 使用 fake lifecycle/result 聚合器验证 writer result 的安全字段可被阶段结果消费。
  - 验证 fake writer 失败和真实 adapter 失败形态一致。
  - _Requirements: 7.1, 7.3, 7.4, 9.3_
  - _Boundary: 不修改 lifecycle 状态机，不实现执行历史持久化。_
  - _Depends: 12, 13, 14_

- [ ] 18. 编写边界测试并运行最小验证
  - 检查 `internal/engine/writer` 不导入 Wails、Vue、frontend API、store、facade、真实数据库 driver、连接管理或凭据存储包。
  - 检查源码不包含数据库产品名称业务分支和未来能力实现关键词。
  - 检查公开 `WriterIssue` 不包含敏感 SQL、连接字符串、DSN、密码、token、参数值、生成值或 raw driver error。
  - 运行 writer 包相关测试并确认通过。
  - _Requirements: 1.5, 2.5, 4.5, 5.5, 6.5, 7.5, 8.4, 8.5, 9.4, 9.5_
  - _Boundary: 不新增第三方依赖，不修改 go.mod。_
  - _Depends: 13, 15, 16, 17_
