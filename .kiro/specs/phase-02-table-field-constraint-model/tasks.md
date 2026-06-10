# Implementation Plan

- [x] 0. 执行上游规格前置门禁
- [x] 0.1 确认 `phase-02-database-schema-model` 已完成并提供上游类型
  - 检查上游规格任务必须已经完成，且 `internal/domain/schema` 中已存在 `DbSchema` 相关基础模型。
  - 检查 `internal/domain/schema` 中已存在 `SchemaValidationIssue`、`SchemaIssueCode`、`SchemaIssueSeverity` 和 `SchemaValidationMode`。
  - 如果上游任务未完成，或任一上游类型不存在，则当前规格实现必须拒绝继续并停止；不得在本规格中创建临时兼容类型或重复声明同名类型。
  - 完成后，后续任务可以安全复用上游 schema domain 合同。
  - _Requirements: 1, 3, 4, 5_
  - _Boundary: UpstreamGate_

- [x] 1. 建立 `表字段与基础约束模型` 领域包
- [x] 1.1 创建核心模型文件 (P)
  - 创建 `表字段与基础约束模型` 所需 Go 文件，并保持 domain 层职责。
  - 为导出类型、字段、常量和枚举值添加项目要求的 Go 注释。
  - 完成后，包可独立编译，不依赖 Wails、Vue 或真实数据库驱动。
  - _Requirements: 1, 2_
  - _Boundary: DomainScaffold_
  - _Depends: 0.1_

- [x] 2. 实现核心模型与枚举
- [x] 2.1 实现核心实体和值对象
  - 实现 DbTable, DbColumn 的稳定身份、父级引用、字段和 JSON 标签。
  - 确保模型不包含 out-of-scope 的服务、API、UI 或执行字段。
  - 完成后，下游规格可以通过稳定合同消费模型。
  - _Requirements: 1, 3_
  - _Boundary: DomainModels_
  - _Depends: 1.1_

- [x] 2.2 实现枚举、状态和值对象 (P)
  - 实现 TableConstraint, ColumnLogicalType 相关枚举或值对象。
  - 枚举使用稳定字符串值，并能识别未知值。
  - 完成后，类型和状态可以安全序列化。
  - _Requirements: 2, 3_
  - _Boundary: DomainEnums_
  - _Depends: 1.1_

- [x] 3. 实现基础校验
- [x] 3.1 复用上游字段级校验错误结构
  - 复用上游 `SchemaValidationIssue`、`SchemaIssueCode`、`SchemaIssueSeverity` 和 `SchemaValidationMode`，不得重新定义同名类型。
  - 仅当现有 `SchemaIssueCode` 无法表达表/字段/约束错误时，才在既有类型上追加必要常量。
  - 字段路径、错误码、严重级别和安全消息必须保持与上游 schema domain 合同一致。
  - 支持一次返回多个校验问题。
  - 完成后，下游服务和 UI 可以复用统一结构化错误，且包内不存在重复 validation 类型声明。
  - _Requirements: 4_
  - _Boundary: Validation_
  - _Depends: 2.1, 2.2_

- [x] 3.2 实现模型校验规则
  - 覆盖必填字段、父级引用、枚举合法性、范围和唯一性规则。
  - 不访问数据库、执行引擎或 UI 状态。
  - 完成后，无效模型返回可诊断且安全的错误集合。
  - _Requirements: 3, 4_
  - _Boundary: Validation_
  - _Depends: 3.1_

- [ ] 4. 增加测试
- [ ] 4.1 覆盖序列化和枚举测试 (P)
  - 测试 JSON 往返、缺省字段兼容和枚举字符串稳定性。
  - 完成后，破坏字段名或枚举值的变化会被捕获。
  - _Requirements: 2, 5_
  - _Boundary: SerializationTests_
  - _Depends: 2.1, 2.2_

- [ ] 4.2 覆盖校验和边界测试 (P)
  - 测试基础校验、多错误返回、上游引用边界和 out-of-scope 未被实现。
  - 完成后，领域模型满足当前规格验收要求。
  - _Requirements: 1, 3, 4, 5_
  - _Boundary: ValidationTests_
  - _Depends: 3.2_

- [ ] 5. 运行最小验证
  - 执行当前领域包相关 `go test`。
  - 如项目已有格式化或 lint 命令，运行与本包相关的最小验证命令。
  - 完成后，记录验证结果和非本规格导致的剩余问题。
  - _Requirements: 5_
  - _Boundary: ValidationRun_
  - _Depends: 4.1, 4.2_
