# Implementation Plan

- [ ] 1. 建立 `数据库 Schema 层级模型` 领域包
- [x] 1.1 创建核心模型文件 (P)
  - 创建 `数据库 Schema 层级模型` 所需 Go 文件，并保持 domain 层职责。
  - 为导出类型、字段、常量和枚举值添加项目要求的 Go 注释。
  - 完成后，包可独立编译，不依赖 Wails、Vue 或真实数据库驱动。
  - _Requirements: 1, 2_
  - _Boundary: DomainScaffold_

- [ ] 2. 实现核心模型与枚举
- [x] 2.1 实现核心实体和值对象
  - 实现 `DbCatalog`、`DbSchema` 的稳定身份、父级引用、时间字段和 JSON 标签，字段合同必须与 `design.md` 的 Data Models 表一致。
  - 实现 `SchemaIdentity`，确保 `schemaName` 在 JSON 中始终存在，且空字符串稳定表示隐式 Schema。
  - 为 `DbSchema` 和 `SchemaIdentity` 实现自定义 JSON 反序列化或等价解码逻辑，必须区分 `schemaName` 缺失、`null` 和 `""`。
  - 确保模型不包含 out-of-scope 的服务、API、UI 或执行字段。
  - 完成后，下游规格可以通过稳定合同消费模型。
  - _Requirements: 1, 3_
  - _Boundary: DomainModels_
  - _Depends: 1.1_

- [ ] 2.2 实现枚举、状态和值对象 (P)
  - 实现 `SchemaIssueCode`、`SchemaIssueSeverity`、`SchemaValidationMode` 和 `SchemaValidationIssue`，字段形状必须兼容现有 `ConfigIssue` / `ApiIssue` 的 `path`、`code`、`severity`、`message` 合同。
  - 枚举使用 `design.md` 中定义的稳定字符串值，并能识别未知值。
  - 明确 schema domain 不直接依赖 `internal/config` 包。
  - 完成后，类型和状态可以安全序列化。
  - _Requirements: 2, 3_
  - _Boundary: DomainEnums_
  - _Depends: 1.1_

- [ ] 3. 实现基础校验
- [ ] 3.1 定义字段级校验错误结构
  - 定义字段路径、错误码、严重级别和安全消息。
  - `path` 必须使用 JSON 字段名语义，`severity` 必须使用 `info`、`warning`、`error` 稳定值。
  - 支持一次返回多个校验问题。
  - 完成后，下游服务和 UI 可以复用结构化错误。
  - _Requirements: 4_
  - _Boundary: Validation_
  - _Depends: 2.1, 2.2_

- [ ] 3.2 实现模型校验规则
  - 覆盖必填字段、父级引用、枚举合法性、时间字段和唯一性语义表达。
  - 提供明确校验入口，至少覆盖 `SchemaValidationModeDraft` 与 `SchemaValidationModePersisted`，避免单一含糊 `Validate()` 混淆新建对象和持久化快照。
  - `DbSchema.SchemaName == ""` 必须作为隐式 Schema 合法表达，不得当作缺失字段处理。
  - 不访问数据库、执行引擎或 UI 状态。
  - 完成后，无效模型返回可诊断且安全的错误集合。
  - _Requirements: 3, 4_
  - _Boundary: Validation_
  - _Depends: 3.1_

- [ ] 4. 增加测试
- [ ] 4.1 覆盖序列化和枚举测试 (P)
  - 测试 JSON 往返、缺省字段兼容、枚举字符串稳定性和 validation issue 字段形状。
  - 测试隐式 Schema 序列化时 `schemaName` 字段存在且值为 `""`。
  - 测试 `schemaName` 缺失和 `null` 反序列化时产生字段级错误，不得被静默当作隐式 Schema。
  - 完成后，破坏字段名或枚举值的变化会被捕获。
  - _Requirements: 2, 5_
  - _Boundary: SerializationTests_
  - _Depends: 2.1, 2.2_

- [ ] 4.2 覆盖校验和边界测试 (P)
  - 测试基础校验、多错误返回、上游引用边界和 out-of-scope 未被实现。
  - 测试 draft / persisted 两种校验模式的 ID 与审计时间差异，确保新建对象和持久化快照使用不同规则。
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
