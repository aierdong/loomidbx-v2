# Implementation Plan

> 任务边界约定：每个实现任务继承最近章节的边界；子任务上的显式 `_Boundary:` 和 `_Depends:` 行用于收窄该章节边界，并在存在时优先生效。

## 1. 建立 gencontext 包基础模型

- [ ] 1.1 创建上下文输入和快照模型
  - 在 `internal/engine/gencontext` 中定义 `ContextInput`、执行任务快照、Project 快照、表快照、字段快照、表约束快照、规则快照和关系快照基础类型
  - 为导出类型和字段添加 Go 注释，保持 engine 内部模型不依赖 UI/API DTO
  - _Requirements: 1.1, 1.2, 2.1, 2.4, 6.4_

- [ ] 1.2 创建 GenerationContext 输出模型
  - 定义 `GenerationContext`、`ContextTable`、`ContextField`、`ContextRule`、`ContextRelation`、`ContextRowTarget` 等只读视图
  - 表达拓扑顺序、目标行数、字段规则摘要和关系安全摘要
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 6.4_

- [ ] 1.3 创建上下文错误和结果模型
  - 定义 `ContextIssue`、错误码、阶段枚举、`ContextResult` 和预检兼容结果字段
  - 实现通过状态、阻断错误和警告的聚合规则
  - _Requirements: 6.1, 6.2, 6.3, 7.1, 7.3_

## 2. 实现输入对齐与只读快照构建

- [ ] 2.1 实现 Project 与计划边界校验
  - 校验 ExecutionTask、Project、ExecutionPlan 和 RowCountPlan 的 ProjectID 一致
  - 校验每个拓扑表能映射到 ProjectTable、DbTable 和 RowCount target
  - 对缺失或越界快照返回字段级阻断问题
  - _Requirements: 1.1, 1.2, 1.4, 2.1_

- [ ] 2.2 实现 ExecutionPlan 与 RowCountPlan 对齐
  - 按顺序比较 `ExecutionPlan.OrderedTables` 与 `RowCountPlan.Tables`
  - 校验 ProjectTableID、TableID 和 ExecutionOrder 完全一致
  - 失败时返回安全阻断问题且不输出部分上下文
  - _Requirements: 1.3, 2.3, 6.2_

- [ ] 2.3 实现只读上下文构建器
  - 从输入快照构建 `GenerationContext`、表视图、字段视图、表约束视图、规则视图、关系视图和行数视图
  - 确保输出表顺序严格沿用拓扑计划，目标行数严格来自 RowCountPlan
  - 不写回 Project、Schema、Rule、ExecutionTask、ExecutionPlan 或 RowCountPlan
  - _Requirements: 2.1, 2.3, 2.4, 2.5, 6.4, 6.5_

- [ ] 2.4 (P) 添加输入对齐和快照构建测试
  - 覆盖有效上下文构建、ProjectID 不一致、缺失 ProjectTable、缺失 DbTable、缺失必需 TableConstraintSnapshot、缺失 RowTarget、计划顺序不一致和 TableID 不一致
  - 验证上下文输出顺序严格沿用拓扑执行顺序，目标行数来自 RowCountPlan
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.3, 6.2, 8.1, 8.2_
  - _Boundary: 仅测试 Generation Context Input Mapper 和 Snapshot Builder，不实现 batch loop 或 generator registry_

## 3. 实现查找接口与字段生成计划视图

- [ ] 3.1 实现上下文查找索引
  - 支持按 ProjectTableID、TableID、ColumnID 和 ExecutionOrder 查询表、字段和目标行数
  - 对未找到场景提供确定性 false 返回或安全问题
  - 保持返回值为只读快照或副本语义
  - _Requirements: 2.2, 2.3, 6.4_

- [ ] 3.2 实现字段约束摘要构建
  - 为主键、唯一键、外键、非空、默认值、可空和字段逻辑类型建立字段级约束摘要
  - PRIMARY 和 UNIQUE 摘要必须来自表约束快照；DbColumn 的派生标记只用于交叉校验，不替代表约束输入
  - 保留字段顺序、字段身份、表身份和后续 generator call input 所需最小元数据
  - _Requirements: 3.1, 3.2, 5.1_

- [ ] 3.3 实现字段规则可用性解释
  - 将 GeneratorConfig 的 generator 名称、映射类型、参数快照和配置状态映射为 `ContextRule`
  - 对必须生成但缺少可用规则的字段返回字段级阻断错误
  - 对需要复核或不可用规则按状态返回警告或阻断错误
  - _Requirements: 2.4, 3.3, 3.4, 7.1_

- [ ] 3.4 (P) 添加查找和字段计划测试
  - 覆盖按 ID 和执行顺序查找、字段顺序稳定、目标行数查询、规则摘要、主键/唯一表约束、外键/非空/默认值/可空约束摘要
  - 覆盖缺失规则、不可用规则和需要复核规则的阻断或 warning 行为
  - _Requirements: 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 8.1, 8.2_
  - _Boundary: 仅测试只读上下文视图，不调用 generator 或执行参数 Schema 校验_

## 4. 实现运行态引用存储

- [ ] 4.1 定义引用键和值模型
  - 定义 `ReferenceScope`、`ReferenceKind`、`ReferenceValue`、`ReferenceQuery` 和关系引用摘要
  - 支持按任务、ProjectTable、Column、Relation 和 RowIndex 表达当前执行内引用范围
  - _Requirements: 4.1, 4.2, 4.5_

- [ ] 4.2 实现引用记录能力
  - 支持记录已生成主键、唯一键、外键候选值和关系引用
  - 保证引用记录绑定当前 GenerationContext 的 TaskID，不跨执行任务共享
  - _Requirements: 4.1, 4.5_

- [ ] 4.3 实现引用候选查询和缺失诊断
  - 支持按父表字段、关系和引用类型查询外键候选集合
  - 对尚未记录的上游引用返回安全缺失引用问题
  - 确保公开错误不包含原始生成值
  - _Requirements: 4.2, 4.4, 5.3, 7.1, 7.2_

- [ ] 4.4 实现外部 DB 来源安全摘要
  - 对 `FROM_DB_QUERY` 或 merged 外部来源只保留安全摘要和未来能力标记
  - 禁止执行 SQL、解析 SQL 或读取真实数据库
  - _Requirements: 4.3, 7.2, 7.5_

- [ ] 4.5 (P) 添加引用存储测试
  - 覆盖主键记录、唯一值记录、关系引用记录、外键候选查询、缺失引用问题和 TaskID 隔离
  - 验证 SQL、连接字符串、密码、令牌和生成值不会出现在公开问题消息中
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 5.3, 7.2, 7.4, 8.1, 8.2_
  - _Boundary: 仅测试当前执行内内存引用存储，不实现真实数据库读取或跨任务缓存_

## 5. 实现最小 GeneratorCallInput

- [ ] 5.1 定义 generator 调用输入模型
  - 定义 `GeneratorCallInput`、`ContextReferenceReader` 和只读引用访问器
  - 包含任务身份、Project 身份、表身份、字段身份、行序、目标行数、逻辑类型、约束摘要和规则摘要
  - _Requirements: 5.1, 5.2_

- [ ] 5.2 实现调用输入构造器
  - 根据 ProjectTableID、ColumnID 和 RowIndex 从 GenerationContext 构造单字段调用输入
  - 对表、字段、规则或目标行数缺失返回字段级阻断问题
  - _Requirements: 5.1, 5.4_

- [ ] 5.3 实现只读引用访问边界
  - 在 GeneratorCallInput 中只暴露候选引用读取能力
  - 禁止 generator call input 暴露引用写入、store、facade、Wails runtime、数据库连接或可变领域对象
  - _Requirements: 5.2, 5.3_

- [ ] 5.4 (P) 添加 GeneratorCallInput 测试
  - 覆盖有效调用输入、行序和目标行数传递、规则摘要传递、关系候选读取、缺失表、缺失字段、缺失规则和缺失目标行数
  - 验证调用输入不选择 generator、不调用生成算法、不组织批量循环
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 8.1, 8.2_
  - _Boundary: 仅测试 generator call input 形状，不实现 generator registry 或 built-in generator_

## 6. 实现 planner / lifecycle 接缝结果

- [ ] 6.1 实现 Generation Context Builder 入口
  - 串联输入边界校验、计划对齐、快照构建、字段计划构建、引用存储初始化和安全问题聚合
  - 成功时返回完整 `GenerationContext`
  - 失败时返回 `Passed=false`、阻断错误和空 Context
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ] 6.2 实现 precheck 兼容结果转换
  - 提供 lifecycle precheck 可聚合的通过状态、阻断错误和警告字段
  - 保持字段方向与 lifecycle / dependency plan / rowcount 安全错误一致
  - _Requirements: 6.1, 6.3, 7.1, 7.3_

- [ ] 6.3 (P) 添加 lifecycle 和 batch 接缝测试
  - 使用 fake lifecycle precheck / generation seam 调用 context builder
  - 使用 fake batch loop 读取表顺序、目标行数、字段视图和 generator call input
  - 验证阻断错误可阻止进入生成阶段，warnings 不阻止成功
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 8.3_
  - _Boundary: 仅验证接缝形状，不修改 lifecycle 状态机，不实现 batch generation loop_

## 7. 固定安全错误与敏感信息边界

- [ ] 7.1 实现安全消息映射
  - 为输入对齐、快照缺失、规则缺失、规则不可用、引用缺失和调用输入构造失败提供固定安全消息
  - 禁止将原始 SQL、连接字符串、密码、令牌、规则参数原文、生成数据或下游错误载荷写入 `SafeMessage`
  - _Requirements: 7.1, 7.2, 7.5_

- [ ] 7.2 (P) 添加安全错误测试
  - 使用包含 SQL、连接字符串、密码、令牌、规则参数和生成数据示例的输入构造失败场景
  - 验证公开错误和警告只包含安全字段和固定安全消息
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  - _Boundary: 仅测试 gencontext 错误公开面，不引入 API/UI/Wails 类型_

## 8. 添加边界和未来能力隔离测试

- [ ] 8.1 添加禁止依赖边界测试
  - 扫描 `internal/engine/gencontext` 导入，确认不依赖 Wails、Vue、frontend API、facade、store 或真实数据库 driver
  - _Requirements: 1.5, 5.2, 8.4_

- [ ] 8.2 添加禁止数据库类型硬编码测试
  - 扫描 gencontext 源码，确认不按 MySQL、PostgreSQL、SQLite、Oracle、SQLServer 等产品名称分支上下文规则
  - _Requirements: 8.4_

- [ ] 8.3 添加未来能力隔离测试
  - 确认 gencontext 包未实现 dependency graph、topological sort、row count solver、generator registry、built-in generator、batch loop、writer adapter、transaction 或 real write 行为
  - _Requirements: 1.5, 3.5, 5.5, 8.5_

- [ ] 8.4 添加持久化语义保护测试
  - 验证 context builder 不写回 `ProjectTable.RowCount`、ExecutionOrder、GeneratorConfig、ExecutionTask、ExecutionPlan、RowCountPlan 或 lifecycle / execution 状态枚举
  - _Requirements: 2.5, 6.5, 8.5_

## 9. 完成验证和整理

- [ ] 9.1 运行 gencontext 包测试
  - 执行 `go test ./internal/engine/gencontext/...` 并修复失败
  - 确认单元测试、接缝测试和边界测试全部通过
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 9.2 运行相关 engine 测试
  - 执行相关 engine 包测试，确认 gencontext 新包不破坏 lifecycle、dependency plan 或 rowcount 包
  - 保持 `go.mod` 不因本规格新增第三方依赖而变化
  - _Requirements: 1.5, 6.5, 8.4, 8.5_

- [ ] 9.3 复核公开模型和任务覆盖
  - 确认每个需求都有实现任务和测试任务覆盖
  - 确认 `GenerationContext` 可作为后续 batch generation loop 和 Phase 4 generator interface 的稳定输入
  - _Requirements: 6.4, 8.1, 8.2, 8.3, 8.4, 8.5_
