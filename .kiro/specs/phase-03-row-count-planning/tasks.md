# Implementation Plan

> 任务边界约定：每个实现任务继承最近章节的边界；子任务上的显式 `_Boundary:` 和 `_Depends:` 行用于收窄该章节边界，并在存在时优先生效。

## 0. 对齐 Phase 2 domain 行数基础类型

- [ ] 0.1 调整 Phase 2 ProjectTable 行数字段类型
  - 将 `internal/domain/project` 中 ProjectTable 的 RowCount 字段及相关校验、序列化和测试从 `int` 对齐为 `int64`
  - 保持 nil、显式 0 和正数的既有业务语义不变，不在 rowcount 输入映射中引入无意义的类型转换
  - 确认 Phase 2 domain 相关测试通过后，再让 rowcount 包基于统一的 `int64` 行数类型建模
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

## 1. 建立 rowcount 包基础模型

- [ ] 1.1 创建行数规划输入和输出模型
  - 在 `internal/engine/rowcount` 中定义 `RowCountInput`、`RowCountTableInput`、`RowCountPlan`、`PlannedRowCount`、`RowCountSource` 等基础类型
  - 为导出类型和字段添加 Go 注释，保持 engine 内部模型不依赖 UI/API DTO
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 6.4_

- [ ] 1.2 创建行数规划错误和结果模型
  - 定义 `RowCountIssue`、错误码、阶段枚举、`RowCountResult` 和预检兼容结果字段
  - 实现通过状态、阻断错误和警告的聚合规则
  - _Requirements: 6.1, 6.2, 6.3, 7.1, 7.3_

- [ ] 1.3 创建约束和目标范围模型
  - 定义 `RowCountConstraint`、约束来源类型、倍率范围和目标范围结构
  - 表达 Parent/Child 与 BaseTable/JoinTable 的上游到下游约束方向
  - _Requirements: 3.1, 3.2, 3.3, 4.3_

## 2. 实现输入映射与拓扑对齐

- [ ] 2.1 实现 Project 表与拓扑计划对齐
  - 根据 dependency `ExecutionPlan.OrderedTables` 建立行数节点
  - 校验拓扑表能映射到 ProjectTable，校验 Project 表集合没有超出拓扑边界的执行表
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 2.2 实现表级行数配置解释
  - 将非空非负配置标记为显式目标
  - 将显式 `0` 标记为零行目标
  - 将空配置标记为动态候选
  - 对负数和安全整数边界返回字段级阻断问题
  - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 2.3 (P) 添加输入映射单元测试
  - 覆盖有效拓扑对齐、缺失 ProjectTable、额外 ProjectTable、显式正数、显式零、动态空值和负数配置
  - 验证输出顺序严格沿用拓扑执行顺序
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 2.2, 2.3, 2.4, 8.1_
  - _Boundary: 仅测试 Row Count Input Mapper，不实现约束求解_

## 3. 实现关系倍率约束构建

- [ ] 3.1 实现 Schema TableRelation 约束构建
  - 从 Parent/Child 关系创建父表到子表的目标行数约束
  - 从 BaseTable/JoinTable 关系创建基础表到关联表的目标行数约束
  - 对缺失端点返回安全预检问题
  - _Requirements: 3.1, 3.2_

- [ ] 3.2 实现 ProjectTableRelation 约束构建和优先级
  - 使用 Project 关系实例倍率覆盖同一执行关系的 Schema 默认倍率
  - 处理当前执行内父子表约束，外部 DB 查询来源只保留安全摘要而不执行 SQL
  - _Requirements: 3.3, 7.2, 7.5_

- [ ] 3.3 实现倍率合法性校验
  - 校验倍率非负、`multiplierMin <= multiplierMax` 和合法零值组合
  - 对非法倍率返回字段级阻断错误
  - 确保不按数据库产品名称硬编码关系方向或倍率语义
  - _Requirements: 3.4, 3.5, 5.4_

- [ ] 3.4 (P) 添加关系约束构建测试
  - 覆盖 Parent/Child、BaseTable/JoinTable、Project 关系覆盖、缺失端点、负倍率、min/max 冲突和非法零值组合
  - 验证 SQL 或连接信息不会出现在公开问题消息中
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 7.2, 7.4, 8.1, 8.2_
  - _Boundary: 仅测试约束构建和安全摘要，不实现目标行数推导_

## 4. 实现目标行数推导与约束校验

- [ ] 4.1 实现显式目标校验
  - 对已有显式目标的下游表校验其是否落入上游倍率约束范围
  - 保持显式零与动态空值的差异
  - _Requirements: 2.2, 4.2, 5.2, 5.3_

- [ ] 4.2 实现动态目标推导
  - 在父表目标已知且约束范围唯一可确定时推导动态下游目标
  - 将推导结果标记为关系推导来源并保留安全来源摘要
  - _Requirements: 4.1, 4.3, 6.4_

- [ ] 4.3 实现多约束范围合并和冲突诊断
  - 合并多个约束对同一动态表产生的范围
  - 对范围为空、显式目标不满足、不可满足零值和非唯一动态范围返回安全阻断错误
  - _Requirements: 4.3, 4.4, 5.1, 5.2, 5.4_

- [ ] 4.4 实现安全整数边界和不可规划诊断
  - 对乘法溢出、范围溢出和无法确定目标行数返回统一不可规划错误
  - 确保规划失败不输出部分成功 `RowCountPlan`
  - _Requirements: 2.4, 5.4, 6.2_

- [ ] 4.5 (P) 添加约束求解单元测试
  - 覆盖唯一动态推导、显式目标范围校验、多约束合并、范围冲突、零父表冲突、固定零下游、溢出和非唯一动态范围
  - 验证输出目标行数非负且顺序稳定
  - _Requirements: 4.1, 4.2, 4.3, 4.4, 5.2, 5.3, 5.4, 8.1, 8.2_
  - _Boundary: 仅测试 rowcount evaluator，不接入 generation context 或 batch loop_

## 5. 实现 planner 协调入口和 lifecycle 接缝

- [ ] 5.1 实现 Row Count Planner 入口
  - 串联输入映射、约束构建、约束求解和安全问题聚合
  - 成功时返回按拓扑顺序排列的 `RowCountPlan`
  - 失败时返回 `Passed=false`、阻断错误和空计划
  - _Requirements: 6.1, 6.2, 6.3, 6.4_

- [ ] 5.2 实现 precheck 兼容结果转换
  - 提供 lifecycle precheck 可聚合的通过状态、阻断错误和警告字段
  - 保持字段方向与 lifecycle / dependency plan 安全错误一致
  - _Requirements: 6.1, 6.3, 7.1, 7.3_

- [ ] 5.3 (P) 添加 lifecycle 接缝测试
  - 使用 fake lifecycle precheck / planner 调用 row count planner
  - 验证成功结果可被后续 generation context 读取，阻断错误可阻止进入生成阶段，warnings 不阻止成功
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 8.3_
  - _Boundary: 仅验证接缝形状，不修改 lifecycle 状态机_

## 6. 固定安全错误与敏感信息边界

- [ ] 6.1 实现安全消息映射
  - 为输入、配置、倍率、冲突、零值和不可规划错误提供固定安全消息
  - 禁止将原始 SQL、连接字符串、密码、生成数据或下游错误载荷写入 `SafeMessage`
  - _Requirements: 7.1, 7.2, 7.5_

- [ ] 6.2 (P) 添加安全错误测试
  - 使用包含 SQL、连接字符串、密码和生成数据示例的关系来源构造失败场景
  - 验证公开错误和警告只包含安全字段和固定安全消息
  - _Requirements: 7.1, 7.2, 7.4, 7.5_
  - _Boundary: 仅测试 rowcount 错误公开面，不引入 API/UI/Wails 类型_

## 7. 添加边界和未来能力隔离测试

- [ ] 7.1 添加禁止依赖边界测试
  - 扫描 `internal/engine/rowcount` 导入，确认不依赖 Wails、Vue、frontend API、facade、store 或真实数据库 driver
  - _Requirements: 1.5, 3.5, 8.4_

- [ ] 7.2 添加禁止数据库类型硬编码测试
  - 扫描 rowcount 源码，确认不按 MySQL、PostgreSQL、SQLite、Oracle、SQLServer 等产品名称分支行数规则
  - _Requirements: 3.5, 8.4_

- [ ] 7.3 添加未来能力隔离测试
  - 确认 rowcount 包未实现 dependency graph、topological sort、generation context、generator registry、batch loop、writer adapter、transaction 或 real write 行为
  - _Requirements: 1.5, 4.5, 5.5, 6.5, 8.5_

- [ ] 7.4 添加持久化语义保护测试
  - 验证 rowcount planner 不写回 `ProjectTable.RowCount`、ExecutionOrder、关系倍率配置或 lifecycle / execution 状态枚举
  - _Requirements: 2.5, 6.5, 8.5_

## 8. 完成验证和整理

- [ ] 8.1 运行 rowcount 包测试
  - 执行 `go test ./internal/engine/rowcount/...` 并修复失败
  - 确认单元测试、接缝测试和边界测试全部通过
  - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5_

- [ ] 8.2 运行相关 engine 测试
  - 执行相关 engine 包测试，确认 rowcount 新包不破坏 lifecycle 或 dependency plan 包
  - 保持 `go.mod` 不因本规格新增第三方依赖而变化
  - _Requirements: 6.5, 8.4, 8.5_

- [ ] 8.3 复核公开模型和任务覆盖
  - 确认每个需求都有实现任务和测试任务覆盖
  - 确认 `RowCountPlan` 可作为后续 generation context 和 batch generation loop 的稳定输入
  - _Requirements: 6.4, 8.1, 8.2, 8.3, 8.4, 8.5_
