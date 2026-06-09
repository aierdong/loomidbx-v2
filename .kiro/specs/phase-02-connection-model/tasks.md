# Implementation Plan

- [ ] 1. 建立连接领域包基础结构
- [x] 1.1 创建 `internal/domain/connection` 包和核心文件 (P)
  - 新增 `connection.go`、`database_type.go`、`credential.go`、`params.go`、`validation.go` 的包声明和最小类型骨架。
  - 为所有导出类型、字段、常量和枚举值添加符合项目规则的 Go 注释。
  - 完成后，包可以被 `go test` 编译，不依赖 Wails、Vue 或数据库驱动。
  - _Requirements: 1.5, 2.4_
  - _Boundary: ConnectionDomainScaffold_

- [ ] 2. 实现数据库类型枚举
- [ ] 2.1 定义 `DatabaseType` 稳定字符串值 (P)
  - 定义 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、ClickHouse、TiDB、Hive 的枚举常量和 JSON 字符串值。
  - 明确 `DatabaseType` 是连接领域层的业务连接类型，不直接替代 `internal/dbx/core.DBType`。
  - 提供 `IsKnown`、`String` 等基础方法，未知值应被明确识别。
  - 完成后，所有数据库类型可被稳定序列化和反序列化。
  - _Requirements: 2.1, 2.2_
  - _Boundary: DatabaseType_
  - _Depends: 1.1_

- [ ] 2.2 表达首期优先支持与网络地址需求 (P)
  - 增加 `IsPrimarySupported` 或等价方法表达 MySQL/PostgreSQL 的首期优先验证目标。
  - 增加 `RequiresNetworkAddress` 或等价方法区分 SQLite 与网络数据库的主机校验需求。
  - 不在本任务中扩展 `internal/dbx/core.DBType` 到 8 种业务类型；如需要映射辅助函数，只允许把 MySQL/PostgreSQL 映射到现有 `core.DBType`，其余类型返回明确不支持结果。
  - 完成后，预留数据库类型可被表达但不会被误判为真实适配能力完成。
  - _Requirements: 2.3, 2.4, 5.2_
  - _Boundary: DatabaseType_
  - _Depends: 2.1_

- [ ] 3. 实现连接聚合和值对象
- [ ] 3.1 定义 `ConnectionID` 与 `Connection` 聚合
  - 定义连接 ID、名称、数据库类型、主机、端口、初始数据库、用户名、凭据引用和扩展参数字段。
  - 添加稳定 JSON 标签，避免使用连接名称作为唯一身份。
  - 完成后，连接基础信息可以作为独立领域对象创建、序列化和传递。
  - _Requirements: 1.1, 1.2, 1.5, 6.1_
  - _Boundary: ConnectionAggregate_
  - _Depends: 2.1_

- [ ] 3.2 定义连接默认值和规范化行为
  - 为缺少可选字段的连接提供安全默认值或零值解释。
  - 对名称、主机、用户名、参数键等用户输入定义必要的 trim/normalize 行为。
  - 完成后，反序列化缺少可选字段不会产生未定义状态。
  - _Requirements: 1.3, 6.2, 6.3, 6.4_
  - _Boundary: ConnectionAggregate_
  - _Depends: 3.1_

- [ ] 4. 实现敏感凭据与扩展参数边界
- [ ] 4.1 定义 `CredentialRef` 敏感凭据引用 (P)
  - 定义凭据引用 ID、凭据类型、`Provider`、`Key` 和可选元数据字段，不包含明文密码或令牌字段。
  - 确保 `Provider` / `Key` 可无损映射到 `internal/storage.SecretRef`；业务 ID、类型和元数据不得替代 provider/key 语义。
  - 确保 JSON 输出只包含引用信息，不输出秘密值。
  - 完成后，连接模型可以引用敏感凭据而不持有明文，并为后续安全存储集成保留转换边界。
  - _Requirements: 3.1, 3.2, 3.4, 5.5_
  - _Boundary: CredentialRef_
  - _Depends: 1.1_

- [ ] 4.2 定义 `ConnectionParams` 扩展参数模型 (P)
  - 支持字符串、数字、布尔和 JSON 兼容结构的参数值表达。
  - 保留参数键和值的稳定序列化格式，不在领域层解释为驱动行为。
  - 完成后，普通扩展参数可以 JSON 往返，并保持可校验。
  - _Requirements: 4.1, 4.2, 4.4, 4.5, 6.1, 6.2_
  - _Boundary: ConnectionParams_
  - _Depends: 1.1_

- [ ] 4.3 实现敏感参数键识别
  - 实现对 `password`、`token`、`secret`、`credential` 等敏感键片段的大小写不敏感识别。
  - 在参数校验中标记敏感参数，不把敏感值写入错误消息。
  - 完成后，扩展参数中的敏感键可被下游安全存储策略识别。
  - _Requirements: 3.3, 3.4, 5.5_
  - _Boundary: ConnectionParams, CredentialRef_
  - _Depends: 4.1, 4.2_

- [ ] 5. 实现基础校验与错误结构
- [ ] 5.1 定义字段级校验错误模型
  - 定义 `ValidationError`、错误码常量和校验结果集合。
  - 每个错误包含字段路径、错误码和安全错误消息。
  - 完成后，服务层和 UI 可复用校验结果而不解析自由文本。
  - _Requirements: 5.1, 5.4, 5.5_
  - _Boundary: Validation_
  - _Depends: 3.1_

- [ ] 5.2 实现 `Connection.Validate` 基础规则
  - 校验连接名称、数据库类型、主机、端口和扩展参数键。
  - 同一次校验返回所有可检测问题，不只返回第一个错误。
  - 完成后，无效连接能产生字段级错误集合，且错误中不包含敏感值。
  - _Requirements: 1.3, 1.4, 4.3, 5.1, 5.2, 5.3, 5.4, 5.5_
  - _Boundary: Validation_
  - _Depends: 2.2, 3.2, 4.3, 5.1_

- [ ] 6. 增加连接模型单元测试
- [ ] 6.1 测试数据库类型枚举行为 (P)
  - 覆盖所有已知数据库类型、未知类型、首期优先支持判断和网络地址需求。
  - 覆盖 `DatabaseType` 与当前 `internal/dbx/core.DBType` 能力边界不混淆的行为：预留类型不能被当作已有 Adapter 能力。
  - 完成后，数据库类型字符串值变化会被测试捕获。
  - _Requirements: 2.1, 2.2, 2.3, 2.4_
  - _Boundary: DatabaseTypeTests_
  - _Depends: 2.2_

- [ ] 6.2 测试连接校验行为 (P)
  - 覆盖名称为空、未知数据库类型、端口越界、网络数据库缺少主机、扩展参数空键和多错误返回。
  - 完成后，基础校验规则和错误集合行为具备回归保护。
  - _Requirements: 1.3, 1.4, 4.3, 5.1, 5.2, 5.3, 5.4, 5.5_
  - _Boundary: ValidationTests_
  - _Depends: 5.2_

- [ ] 6.3 测试序列化、反序列化和敏感字段排除 (P)
  - 覆盖连接 JSON 往返、缺少可选字段、未知字段兼容、普通扩展参数复杂结构和敏感凭据引用。
  - 断言 `CredentialRef` 序列化字段可恢复 `Provider` / `Key`，并可映射到 `storage.SecretRef` 的同名语义。
  - 断言序列化输出和错误消息都不包含明文密码、令牌或敏感参数值。
  - 完成后，连接模型满足本地持久化和后续 Wails 合同的稳定性要求。
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 4.1, 4.2, 4.4, 6.1, 6.2, 6.3, 6.4, 6.5_
  - _Boundary: SerializationTests_
  - _Depends: 4.3, 5.2_

- [ ] 7. 运行领域模型验证
  - 执行连接领域包相关 `go test`，优先限定到 `internal/domain/connection` 或等价包路径。
  - 如项目已有格式化或 lint 命令，运行与本包相关的最小验证命令。
  - 完成后，记录通过的验证命令和任何非本规格导致的剩余问题。
  - _Requirements: 6.5_
  - _Boundary: ValidationRun_
  - _Depends: 6.1, 6.2, 6.3_
