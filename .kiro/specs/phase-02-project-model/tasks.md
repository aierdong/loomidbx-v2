# Implementation Plan

- [x] 1. 建立 Project 领域包基础
  - 在现有 Go 模块内建立 `Project 任务组织模型` 的领域包位置，保留纯 domain 边界。
  - 准备核心模型、枚举、校验和测试文件承载点，并遵守 Go 注释规则。
  - 完成后，包能够在不依赖 Wails、Vue、store、service、engine、generator、真实数据库驱动或 `internal/dbx` 的情况下被 `go test` 发现。
  - _Requirements: 1.1, 1.4, 3.4, 5.4_
  - _Boundary: DomainScaffold_

- [ ] 2. 实现 Project 配置领域模型
- [x] 2.1 实现 Project 聚合根合同
  - 表达 Project 的稳定身份、目标连接、名称、说明和时间字段。
  - 为所有导出类型和字段补充符合项目规则的 Go 注释。
  - 完成后，Project 可通过 lower camelCase JSON 字段完成创建、加载和序列化。
  - _Requirements: 1.1, 1.2, 3.1, 3.2, 5.1, 5.2_
  - _Boundary: ProjectModel_
  - _Depends: 1_

- [x] 2.2 (P) 实现 ProjectTable 表级配置合同
  - 表达 Project 内目标表引用、可空行数、清空策略和执行顺序快照。
  - 明确 Project 采用两阶段创建，ProjectTable 只能引用已持久化 Project，`projectId` 必须大于 0。
  - 保留 `rowCount` 的 nil、0 和正数三种语义，不推导表角色矩阵。
  - 完成后，ProjectTable 可稳定序列化，并且不包含字段生成规则或运行时执行状态。
  - _Requirements: 1.1, 1.2, 2.1, 2.4, 3.1, 3.2, 5.1, 5.2, 5.4_
  - _Boundary: ProjectTableModel_
  - _Depends: 1_

- [x] 2.3 (P) 实现 ProjectTableRelation 与取值来源合同
  - 表达关系实例化快照、上下游 ProjectTable 引用、倍数范围、取值来源和 SQL 文本字段。
  - 明确 ProjectTableRelation 只能引用已持久化 Project，并保护 `parentProjectTableId` 与取值来源的组合语义。
  - 稳定 `FROM_EXECUTION`、`FROM_DB_QUERY`、`MERGED` 枚举字符串，并能识别未知枚举。
  - 完成后，ProjectTableRelation 可稳定序列化，并且只保存关系配置快照、不执行 SQL。
  - _Requirements: 1.1, 1.2, 2.1, 2.2, 2.4, 3.1, 3.2, 5.1, 5.2, 5.4_
  - _Boundary: ProjectRelationModel, RelationValueSource_
  - _Depends: 1_

- [ ] 3. 实现字段级基础校验
- [x] 3.1 定义 Project 校验问题合同
  - 提供字段路径、错误码、安全消息和必要 severity 的 JSON 兼容结构。
  - 保证校验结果能够一次返回多个问题，不 panic、不回显敏感 SQL 内容。
  - 完成后，下游 service/UI 可按稳定 JSON 形状消费字段级错误。
  - _Requirements: 1.3, 4.1, 4.2, 4.4, 4.5, 5.3_
  - _Boundary: ProjectValidation_
  - _Depends: 2.1, 2.2, 2.3_

- [x] 3.2 实现模型基础校验规则
  - 覆盖 ID 形状、Project 两阶段创建约束、必填字符串、控制字符、时间顺序、行数、执行顺序、倍数范围和枚举合法性。
  - 覆盖 `relValueSource`、`relSourceSql` 与 `parentProjectTableId` 的组合规则，以及同一 Project 中重复表引用的集合级检查。
  - 完成后，无效 Project、ProjectTable 和 ProjectTableRelation 会返回稳定字段路径和错误码，不访问外部资源。
  - _Requirements: 1.3, 2.3, 3.3, 4.1, 4.2, 4.4, 4.5, 5.3_
  - _Boundary: ProjectValidation_
  - _Depends: 3.1_

- [x] 3.3 实现 JSON presence 诊断入口
  - 提供可区分字段缺失与 Go 零值的 JSON 解码辅助能力。
  - 覆盖 `truncateBefore`、nullable ID、nullable rowCount 和枚举字段的缺失诊断。
  - 完成后，缺少必填 JSON 字段会返回字段级 presence issue，而不会被误判为合法零值。
  - _Requirements: 1.3, 4.3, 5.3_
  - _Boundary: ProjectJSONValidation_
  - _Depends: 3.2_

- [ ] 4. 增加单元与边界测试
- [x] 4.1 覆盖模型、枚举和 JSON 序列化测试
  - 覆盖 Project、ProjectTable、ProjectTableRelation 的 JSON 往返和 lower camelCase 字段名。
  - 覆盖 rowCount 的 nil、0、正数语义，以及三种 RelationValueSource 枚举字符串稳定性。
  - 完成后，破坏字段名、枚举值或 nullable 语义的改动会导致相关 Go 测试失败。
  - _Requirements: 1.2, 1.5, 2.2, 2.5, 3.2, 3.5, 5.1, 5.2, 5.5_
  - _Boundary: SerializationTests_
  - _Depends: 2.1, 2.2, 2.3_

- [ ] 4.2 覆盖校验、presence 和 out-of-scope 边界测试
  - 覆盖多错误返回、必填字段缺失、非法引用形状、非法范围、非法枚举、SQL_REQUIRED、PARENT_REQUIRED 和重复表引用。
  - 覆盖 Project 包未吸收服务、API、UI、数据库访问、执行算法、字段规则或运行时状态。
  - 完成后，当前规格所有基础校验和边界承诺都有可执行测试保护。
  - _Requirements: 1.3, 1.4, 1.5, 2.3, 2.4, 2.5, 3.3, 3.4, 3.5, 4.1, 4.2, 4.3, 4.4, 4.5, 5.3, 5.4, 5.5_
  - _Boundary: ValidationBoundaryTests_
  - _Depends: 3.3_

- [ ] 5. 运行最小验证
  - 对当前领域包运行 Go 格式化和相关 `go test`。
  - 若项目级验证命令已存在，仅运行与本包相关的最小验证，不扩大到后续阶段能力。
  - 完成后，Project 领域包相关测试通过，任何非本规格遗留问题被清楚记录。
  - _Requirements: 1.5, 2.5, 3.5, 4.5, 5.5_
  - _Boundary: ValidationRun_
  - _Depends: 4.1, 4.2_
