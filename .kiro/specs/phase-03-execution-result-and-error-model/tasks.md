# Implementation Plan

> 任务边界约定：每个实现任务继承最近章节的边界；子任务上的显式 `_Boundary:` 和 `_Depends:` 行用于收窄该章节边界，并在存在时优先生效。

## 1. 建立 result 包基础模型

- [ ] 1.1 创建运行时结果和状态模型
  - 在 `internal/engine/result` 中定义 `ExecutionResult`、`TableExecutionResult`、`BatchResult`、canonical 任务/表状态字段、批次摘要状态和统计字段
  - `ExecutionResult.Status` / `TableExecutionResult.Status` 直接使用 `execution.ExecutionTaskStatus` / `execution.ExecutionTableStatus`，不定义平行 runtime 任务/表状态枚举
  - 为导出类型和字段添加 Go 注释，保持 runtime result 不依赖 UI/API DTO
  - _Requirements: 1.1, 1.2, 2.1, 3.1, 4.1_

- [ ] 1.2 创建 EngineError 和失败范围模型
  - 定义 `EngineError`、错误码、错误类别、阶段枚举、`FailureScope`、blocking/warning 标记和发生时间字段
  - 确保公开错误只包含安全消息、字段路径和安全 ID/索引
  - _Requirements: 5.1, 5.2, 5.3, 8.1, 8.3_

- [ ] 1.3 创建结果汇总输入模型
  - 定义 `ResultInput`、阶段摘要、表摘要输入、批次摘要输入和上游 issue 同构 DTO
  - 支持成功阶段结果、失败 issue-only 结果和取消摘要输入
  - _Requirements: 1.1, 1.2, 1.4, 6.5_

- [ ] 1.4 (P) 添加基础模型单元测试
  - 覆盖 canonical 状态字段稳定值、零值安全、统计字段非负约束和 JSON 往返
  - 验证模型字段不包含行值、SQL、参数、连接字符串或生成数据字段
  - _Requirements: 1.2, 4.5, 8.4, 9.1_
  - _Boundary: 仅测试 result 基础模型，不实现上游阶段 mapper_

## 2. 实现输入校验和安全消息过滤

- [ ] 2.1 实现 ResultInput 基础校验
  - 校验 TaskID、ProjectID、时间、表身份、批次范围和统计值合法性
  - 对身份不一致、统计为负数或非法范围返回 aggregation stage 阻断错误
  - _Requirements: 1.1, 1.2, 1.3_

- [ ] 2.2 实现敏感信息 sanitizer
  - 提供固定安全消息模板和敏感样本检测辅助函数
  - 过滤 SQL、DSN、连接字符串、密码、令牌、规则参数、statement 参数、生成值和 raw driver error
  - _Requirements: 5.4, 5.5, 8.1, 8.2, 8.5_

- [ ] 2.3 (P) 添加输入校验测试
  - 覆盖有效输入、缺少身份、Project 不一致、表身份缺失、负统计、非法批次范围和 issue-only 失败输入
  - 验证非法输入不会生成可信部分成功结果
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 9.1_
  - _Boundary: 仅测试 result input validator，不调用 lifecycle、planner、batch 或 writer_

- [ ] 2.4 (P) 添加 sanitizer 测试
  - 使用包含 SQL、DSN、password、token、规则参数、statement 参数、生成值和 driver 原始错误的样本
  - 验证 `EngineError`、runtime result 和 history snapshot 公开字段均不包含敏感样本
  - _Requirements: 5.5, 8.1, 8.2, 8.4, 8.5_
  - _Boundary: 仅测试公开错误过滤，不引入日志或 observability pipeline_

## 3. 实现上游 issue/result 映射

- [ ] 3.1 实现 lifecycle 和 precheck mapper
  - 将 lifecycle 状态、预检问题、取消问题和状态机失败映射为 task-level `EngineError`
  - 保留安全阶段、blocking 标记和任务范围
  - _Requirements: 2.2, 2.4, 5.1, 6.1_

- [ ] 3.2 实现 plan、rowcount 和 context mapper
  - 将 dependency plan、row count planning 和 generation context issue 映射为 planning/context/reference 错误类别
  - 保留安全表、关系、字段路径和 source summary
  - _Requirements: 5.1, 5.2, 6.2_

- [ ] 3.3 实现 batch generation mapper
  - 将 batch loop stats、progress、`BatchIssue` 和引用失败映射为表/批次/字段级结果输入和 `EngineError`
  - 区分 generation、reference 和 batch 阶段错误
  - _Requirements: 3.3, 4.3, 5.1, 6.3_

- [ ] 3.4 实现 writer mapper
  - 将 writer `WriteResult`、`WriterIssue`、partial accepted、accepted rows、statement count 和 failed statement index 映射为批次结果输入
  - 区分 writer、transaction、clear、dialect 和 executor 阶段错误
  - _Requirements: 4.1, 4.2, 4.3, 5.1, 6.4_

- [ ] 3.5 (P) 添加 mapper 单元测试
  - 覆盖 lifecycle、precheck、planner、rowcount、context、generation、reference、writer、transaction、clear、dialect、executor 和 unknown issue
  - 验证错误排序稳定、blocking/warning 区分正确且 fallback 不透传 raw error
  - _Requirements: 5.1, 5.3, 5.4, 6.1, 6.2, 6.3, 6.4, 9.2_
  - _Boundary: 仅测试 mapper 输入输出，不实现上游阶段内部逻辑_

## 4. 实现任务、表和批次汇总

- [ ] 4.1 实现 BatchResult 构建
  - 根据批次输入构造成功、失败和 partial accepted 批次摘要
  - 保留 batch index、start row、end row、accepted rows、statement count 和 failed statement index
  - _Requirements: 4.1, 4.2, 4.3, 4.4_

- [ ] 4.2 实现 TableExecutionResult 构建
  - 按拓扑执行顺序构造表级结果，汇总 target/generated/accepted rows、batch count 和 statement count
  - 对失败表关联 primary blocking error，对未执行依赖表标记 skipped
  - _Requirements: 3.1, 3.2, 3.3, 3.4_

- [ ] 4.3 实现任务级状态推导
  - 推导 `execution.ExecutionTaskStatusSuccess`、`execution.ExecutionTaskStatusFailed` 和 `execution.ExecutionTaskStatusPartialFailed`
  - 写入前阻断返回 `FAILED`，写入后阻断返回 `PARTIAL_FAILED`，取消按已接受范围使用 `FAILED` 或 `PARTIAL_FAILED` 并记录 cancellation error
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 4.4 实现 ResultAggregator 协调入口
  - 串联输入校验、上游映射、批次构建、表构建、状态推导、统计汇总和错误排序
  - 确保失败时返回安全 `ExecutionResult` 而不是 panic 或 raw error
  - _Requirements: 1.1, 1.3, 2.1, 3.1, 4.1, 5.3_

- [ ] 4.5 (P) 添加汇总场景测试
  - 覆盖全部成功、预检失败、规划失败、上下文失败、生成失败、引用失败、writer 失败、非事务部分接受、取消和 skipped 表
  - 验证总统计、表统计、批次统计和错误范围一致
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.3, 9.2_
  - _Boundary: 仅测试最终汇总，不执行真实 generation 或 write_

## 5. 实现 Phase 2 历史模型映射

- [ ] 5.1 实现任务历史状态校验与映射
  - 校验 runtime result 已使用 Phase 2 `ExecutionTaskStatus` canonical 状态，并原样写入历史 domain model
  - 处理取消场景的 cancellation error snapshot：无已接受范围保持 `FAILED`，存在已接受范围保持 `PARTIAL_FAILED`
  - _Requirements: 7.1_

- [ ] 5.2 实现表历史映射
  - 将 `TableExecutionResult` 映射为 Phase 2 `ExecutionTableResult`
  - 填充 rows written、execution order、table/schema name snapshots、status 和时间字段
  - _Requirements: 7.2_

- [ ] 5.3 实现错误快照映射
  - 将 `EngineError` 映射为 `ExecutionErrorSnapshot` 的 code、message、fieldPath 和 occurredAt
  - 提供单字段 error message 的安全降级摘要来源
  - _Requirements: 7.3, 7.4, 8.1_

- [ ] 5.4 (P) 添加历史映射测试
  - 覆盖 task success、task failed、task partial failed、取消错误快照、table success、table failed 和 table skipped
  - 验证 Phase 2 domain 模型校验通过，错误快照不包含敏感原始载荷
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 9.3_
  - _Boundary: 仅测试纯内存 domain mapping，不实现 repository、migration、API 或 Wails DTO_

## 6. 固定接缝兼容和确定性输出

- [ ] 6.1 实现 fake upstream result/issue 构造器
  - 提供 fake lifecycle、planner、rowcount、context、batch 和 writer 输入构造器
  - 支持成功、阻断、warning、partial accepted、取消错误和 skipped 场景
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 9.3_

- [ ] 6.2 实现错误和结果排序规则
  - 按阶段、执行顺序、批次 index、statement index、字段路径和输入序号稳定排序
  - 避免 map iteration 导致结果和测试不稳定
  - _Requirements: 3.1, 5.3, 9.1_

- [ ] 6.3 (P) 添加接缝兼容测试
  - 使用 fake upstream 构造完整 lifecycle -> plan -> rowcount -> context -> batch -> writer 链路输入
  - 验证 result aggregator 不要求上游包反向依赖 result 包，输出可被 history mapper 消费
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 9.3_
  - _Boundary: 仅验证接缝形状和 mapper 兼容，不修改上游阶段包_

- [ ] 6.4 (P) 添加确定性输出测试
  - 多次使用不同 map 插入顺序构造输入，验证 table、batch、error 和 warning 输出顺序一致
  - 验证同一失败场景映射出稳定 code、stage、fieldPath 和 scope
  - _Requirements: 3.1, 5.3, 9.1_
  - _Boundary: 仅测试排序和稳定输出，不引入持久化排序规则_

## 7. 添加边界和未来能力隔离测试

- [ ] 7.1 添加禁止外部层依赖测试
  - 扫描 `internal/engine/result` 导入，确认不依赖 Wails、Vue、frontend API、store、facade、真实数据库 driver、连接管理或凭据存储
  - _Requirements: 1.5, 3.5, 7.5, 9.4_

- [ ] 7.2 添加禁止数据库类型硬编码测试
  - 扫描 result 源码，确认不按 MySQL、PostgreSQL、SQLite、Oracle、SQLServer 等产品名称分支结果或错误处理
  - _Requirements: 5.5, 9.4_

- [ ] 7.3 添加未来能力隔离测试
  - 确认 result 包未实现 execution history API、UI progress、log query、metrics、tracing、telemetry、generated data persistence、retry、resume、compensation 或 recovery 行为
  - _Requirements: 2.5, 7.5, 9.5_

- [ ] 7.4 添加上游阶段边界保护测试
  - 确认 result 包未实现 lifecycle 状态机、dependency graph、topological sort、row count solver、generation context builder、batch generation loop、writer adapter、transaction 或 real write 行为
  - _Requirements: 1.5, 3.5, 9.4, 9.5_

## 8. 完成验证和整理

- [ ] 8.1 运行 result 包测试
  - 执行 `go test ./internal/engine/result/...` 并修复失败
  - 确认模型、mapper、aggregator、history、sanitizer、seam 和 boundary 测试全部通过
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ] 8.2 运行相关 engine/domain 测试
  - 执行相关 engine 和 domain execution 包测试，确认 result 新包不破坏 Phase 2 domain 或 Phase 3 upstream 包
  - 保持 `go.mod` 不因本规格新增第三方依赖而变化
  - _Requirements: 6.5, 7.5, 9.3, 9.4, 9.5_

- [ ] 8.3 复核公开模型和任务覆盖
  - 确认每个需求都有实现任务和测试任务覆盖
  - 确认 `ExecutionResult` 直接使用 canonical execution statuses，并可作为后续 history API、error response contract、progress view 和 observability 规格的稳定安全输入
  - _Requirements: 1.1, 7.1, 8.1, 9.1, 9.2, 9.3, 9.4, 9.5_
