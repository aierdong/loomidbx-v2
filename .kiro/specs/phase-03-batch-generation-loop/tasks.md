# Implementation Plan

> 任务边界约定：每个实现任务继承最近章节的边界；子任务上的显式 `_Boundary:` 和 `_Depends:` 行用于收窄该章节边界，并在存在时优先生效。

## Tasks

- [ ] 1. 创建 batch 包基础模型与安全错误类型
  - 新增 `internal/engine/batch` 包的输入、配置、结果、进度、错误码、阶段、字段路径和安全范围模型。
  - 定义 `BatchInput`、`BatchConfig`、`BatchResult`、`BatchIssue`、`BatchProgress` 和表级统计结构。
  - `BatchResult.Status` 和表级统计状态直接复用 `execution.ExecutionTaskStatus` / `execution.ExecutionTableStatus`，不新增重复任务/表状态枚举。
  - 实现安全错误构造器，确保公开消息只包含安全摘要。
  - _Requirements: 1.1, 1.2, 7.2, 7.3, 8.1, 8.3_
  - _Boundary: 不修改 lifecycle 状态枚举，不引入 UI/API/Wails/DB/store/facade 依赖。_
  - _Depends: 无_

- [ ] 2. 实现计划与上下文输入对齐校验
  - 校验 TaskID、ProjectID、`ExecutionPlan`、`RowCountPlan` 和 `GenerationContext` 的边界一致性。
  - 校验 ProjectTable 集合、TableID、ExecutionOrder 和 TargetRows 完全对齐。
  - 对不一致输入返回阻断 `BatchIssue`，不生成部分工作单元。
  - _Requirements: 1.1, 1.2, 1.3, 1.4_
  - _Boundary: 不重建依赖图、不重算行数、不重建 GenerationContext。_
  - _Depends: 1_

- [ ] 3. 实现批次配置安全校验
  - 校验批次大小为正数且不超过安全上限。
  - 为缺失、负数、零值和超限配置返回安全阻断错误。
  - 覆盖防止无限循环和过大批次的边界测试。
  - _Requirements: 2.2, 2.4, 8.1_
  - _Boundary: 不写回 Project、ExecutionPlan、RowCountPlan 或 GenerationContext 配置。_
  - _Depends: 1_

- [ ] 4. 实现表级工作单元构建
  - 基于对齐后的计划和上下文构造 `TableWorkUnit`。
  - 保留 ProjectTableID、TableID、ExecutionOrder、TargetRows、字段视图和行数来源摘要所需字段。
  - 确保工作单元顺序与 `ExecutionPlan.OrderedTables` 完全一致。
  - _Requirements: 1.1, 1.2, 2.1_
  - _Boundary: 不修改上游计划或上下文字段。_
  - _Depends: 2, 3_

- [ ] 5. 实现确定性批次切分
  - 将每个表目标行数切分为连续、无重叠、覆盖完整目标的 `[start, end)` 批次范围。
  - 支持目标行数小于、等于和大于批次大小的情况。
  - 对零行表生成完成统计但不生成批次执行工作。
  - _Requirements: 2.1, 2.2, 2.3, 2.5_
  - _Boundary: 不实现跨表并行、高级调度或拓扑重排。_
  - _Depends: 4_

- [ ] 6. (P) 实现字段值状态和 writer 负载模型
  - 定义 `FieldValueState`、`FieldValue`、`GeneratedRow`、`BatchRows` 和 `BatchPayload`。
  - 区分具体值、显式空值和省略/默认值语义。
  - 确保 writer seam 可识别字段状态而不是依赖 Go 零值。
  - _Requirements: 3.1, 3.3, 6.1, 6.4_
  - _Boundary: 不实现 writer adapter 内部或数据库类型转换。_
  - _Depends: 1_

- [ ] 7. (P) 定义 GeneratorInvoker 接缝
  - 定义最小 `GeneratorInvoker` 接口和 `GeneratedValue` 返回模型。
  - 使用 `gencontext.GeneratorCallInput` 作为调用输入边界。
  - 将 generator 原始失败映射为安全字段错误的接口辅助函数。
  - _Requirements: 3.2, 3.4, 3.5, 8.2_
  - _Boundary: 不实现 registry、built-in generator、随机算法或参数 schema 校验。_
  - _Depends: 1_

- [ ] 8. (P) 定义 BatchWriter 接缝
  - 定义最小 `BatchWriter` 接口和 `BatchWriteResult`，只保留 `AcceptedRows` 作为 writer 接受行数语义。
  - 设计 writer seam 失败到 `BatchIssue` 的安全映射。
  - 明确 `PartialAccepted` 或接受行数小于批次行数时的安全处理边界。
  - 保证批次负载只包含任务/表身份、批次范围、列集合和行值集合。
  - _Requirements: 6.1, 6.2, 6.3, 6.5, 8.2_
  - _Boundary: 不打开数据库连接、不执行 SQL、不实现事务、清空、重试或真实写入。_
  - _Depends: 1, 6_

- [ ] 9. 实现行缓冲与普通字段生成
  - 按 `ContextField` 顺序为每行建立行缓冲。
  - 对需要普通生成的字段构造 `GeneratorCallInput` 并调用 `GeneratorInvoker`。
  - 对可空、默认值或自动生成语义字段输出对应字段值状态。
  - _Requirements: 3.1, 3.2, 3.3, 3.4_
  - _Boundary: 不选择具体 generator 实现，不生成示例数据。_
  - _Depends: 5, 6, 7_

- [ ] 10. 实现关系字段优先引用读取
  - 识别需要关系引用优先的外键或 Project 关系字段。
  - 通过 `GenerationContext` 安全引用访问能力读取候选引用。
  - 引用存在时填充引用值并跳过普通 generator 调用。
  - _Requirements: 4.1, 4.2, 4.5_
  - _Boundary: 不把引用原始值写入公开错误、进度或日志摘要。_
  - _Depends: 9_

- [ ] 11. 实现引用缺失与外部来源能力缺失诊断
  - 对必需关系候选缺失返回字段级阻断错误。
  - 对外部 DB 查询来源保留安全来源摘要和待接入能力标记。
  - 确认不会执行 SQL 或读取真实数据库。
  - _Requirements: 4.3, 4.4, 8.1, 8.3_
  - _Boundary: 不实现外部数据源 generator 或真实 DB 查询。_
  - _Depends: 10_

- [ ] 12. 实现 pending 引用收集与提交
  - 在行组装阶段识别主键、唯一键和关系候选字段并收集 pending references。
  - writer seam 整批成功后通过 `RuntimeReferenceStore` 记录引用。
  - writer seam 失败或部分成功时不提交该批次 pending references。
  - 当失败或部分成功发生在存在下游子表的父表上时，标记依赖该父表的后续表为 skipped/blocked，但允许无依赖关系的后续表继续执行。
  - _Requirements: 5.1, 5.2, 5.3_
  - _Boundary: 不跨任务共享引用缓存，不持久化生成数据内容。_
  - _Depends: 10, 11_

- [ ] 13. 实现引用记录失败处理
  - 将 `RuntimeReferenceStore` 记录失败或安全问题映射为引用作用域阻断 `BatchIssue`。
  - 标记依赖失败引用的后续表为 skipped/blocked，并允许无依赖关系的后续表继续执行。
  - 覆盖公开错误不包含原始引用值。
  - _Requirements: 5.4, 5.5, 8.1, 8.5_
  - _Boundary: 不公开生成值或引用值。_
  - _Depends: 12_

- [ ] 14. 实现 batch runner 主循环
  - 串联输入对齐、表工作单元、批次切分、行组装、writer seam 调用和引用提交。
  - 按拓扑表顺序和批次顺序推进；表级失败后记录安全失败结果，标记依赖子表为 skipped/blocked，并继续执行无依赖关系的后续表。
  - 处理零行表、generator 失败、引用缺失、writer 失败、writer 部分成功和引用提交失败路径。
  - 对存在下游子表的父表，writer 失败或部分成功后不得继续生成依赖该父表的后续表。
  - 支持最终结果直接返回 `execution.ExecutionTaskStatusSuccess`、`execution.ExecutionTaskStatusPartialFailed` 或 `execution.ExecutionTaskStatusFailed`，而不是默认首个表失败即终止整个 Project。
  - _Requirements: 2.1, 2.3, 3.4, 5.2, 6.2, 6.3, 7.2, 7.3_
  - _Boundary: 不实现并行调度、重试策略、事务或真实写入。_
  - _Depends: 5, 8, 9, 12, 13_

- [ ] 15. 实现进度 sink 与统计汇总
  - 输出表开始/完成、批次开始/完成/失败和表跳过的进度摘要。
  - 汇总 Project 与表级目标行数、成功行数、失败行数、批次数、成功表数、失败表数和跳过表数。
  - 确保结果字段可被 lifecycle generation 阶段聚合，并能通过 canonical task status 表达部分成功状态。
  - _Requirements: 7.1, 7.2, 7.4_
  - _Boundary: 不直接推送 UI/API/Wails 事件，不写执行历史持久化模型。_
  - _Depends: 14_

- [ ] 16. 实现安全错误过滤和敏感信息测试辅助
  - 统一过滤 generator、reference 和 writer 原始错误中的 SQL、连接字符串、密码、token、规则参数和生成数据内容。
  - 使用固定安全消息和安全位置标识生成公开错误。
  - 为测试提供敏感样本断言辅助。
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_
  - _Boundary: 不透传原始下游错误载荷。_
  - _Depends: 1, 7, 8, 11, 13_

- [ ] 17. 编写单元测试覆盖调度、组装、引用和 writer seam
  - 覆盖输入对齐、批次配置、拓扑调度、批次范围、零行表和行组装。
  - 覆盖关系字段优先、引用缺失、引用提交、generator 失败、writer 失败和 writer 部分成功。
  - 覆盖存在下游子表的父表在 writer 失败或部分成功后跳过依赖表，并继续执行无依赖关系的后续表。
  - 覆盖字段值状态传递、成功统计、失败统计、跳过统计和 `execution.ExecutionTaskStatusPartialFailed` 结果状态。
  - _Requirements: 9.1, 9.2_
  - _Boundary: 使用 fake invoker、fake writer 和 fake reference store，不访问真实数据库。_
  - _Depends: 14, 15, 16_

- [ ] 18. 编写 lifecycle 和 writer 接缝测试
  - 使用 fake lifecycle generation port 验证 batch result 可被聚合。
  - 使用 fake writer adapter 边界验证批次负载结构和失败传播。
  - 验证 progress sink 只输出内存摘要且不触发 UI/API/Wails。
  - _Requirements: 7.4, 9.3_
  - _Boundary: 不修改 lifecycle 状态机，不实现真实 writer adapter。_
  - _Depends: 15, 17_

- [ ] 19. 编写边界测试并运行最小验证
  - 检查 batch 包不导入 Wails、Vue、frontend API、真实数据库 driver、store 或 facade。
  - 检查源码不包含数据库产品名称硬编码和未来能力实现关键词。
  - 检查公开 `BatchIssue` 不包含敏感 SQL、连接字符串、密码、token、规则参数或生成值。
  - 运行 batch 包相关测试并确认通过。
  - _Requirements: 1.5, 2.5, 3.5, 6.5, 7.5, 8.4, 8.5, 9.4, 9.5_
  - _Boundary: 不新增第三方依赖，不修改 go.mod。_
  - _Depends: 16, 17, 18_
