# Research Log

## Summary

本规格属于 Phase 2 领域模型新增能力，采用轻量设计发现。核心结论是：连接模型应落在 Go 后端 domain 边界内，只表达连接配置和值对象，不实现 CRUD 服务、Wails API 或真实连接验证；敏感凭据必须通过引用或加密结果表达，并与普通序列化数据隔离。

## Research Log

### 主题：Phase 1 边界复用

- **来源**: `.kiro/steering/roadmap.md`、`.kiro/steering/tech.md`、`.kiro/steering/structure.md`
- **发现**: Phase 1 已确立本地存储、配置系统和数据库方言接口边界；Phase 2 只能定义领域模型和持久化表达，不能实现未来 API/UI/执行能力。
- **影响**: 本规格设计 `internal/domain/connection` 作为领域模型包，并只通过数据库类型字符串与 `internal/dbx` 的方言方向保持一致。

### 主题：敏感信息边界

- **来源**: 产品隐私边界与技术持久化约束
- **发现**: 数据库密码、令牌和敏感参数默认不得上传远端，也不应作为普通业务字段明文持久化。
- **影响**: 设计使用 `CredentialRef` / `EncryptedSecret` 形式表达敏感凭据，并提供敏感参数识别方法，具体加密或安全存储由本地存储策略和后续服务实现。

### 主题：序列化与下游兼容

- **来源**: 后续 Schema、Project、API 规格依赖关系
- **发现**: 下游需要稳定引用连接身份、数据库类型和扩展参数，但不应耦合真实数据库驱动。
- **影响**: 设计中将连接 ID、数据库类型、普通参数、敏感引用作为稳定合同，并以单元测试验证 JSON 往返和敏感字段排除。

## Architecture Pattern Evaluation

- **选择**: 领域模型和值对象优先，服务和持久化适配延后。
- **理由**: 当前 spec 的边界是连接配置领域表达；过早引入 repository、service 或 Wails binding 会吸收 Phase 7/API 和 Phase 1/storage 之外的职责。
- **替代方案**: 直接设计连接 CRUD 服务。该方案被拒绝，因为它会引入 API、事务和 UI 工作流边界，超出本规格。

## Synthesis Outcomes

- 使用 `Connection` 聚合表达业务身份和基础连接参数。
- 使用 `DatabaseType` 枚举统一 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、ClickHouse、TiDB、Hive 的字符串值。
- 使用 `ConnectionParams` 表达普通扩展参数，使用 `CredentialRef` 和敏感键识别函数保护凭据边界。
- 校验逻辑应返回字段级错误集合，避免下游重复实现基础规则。
