# DBX 数据库兼容抽象

`internal/dbx/` 当前交付数据库方言接口、值对象、内存 registry 和 fake 测试替身。它为后续 service、Schema API、writer adapter 和执行引擎提供统一边界，但不代表任何真实数据库已经可连接或可写入。

## 已交付范围

- `core`：`Adapter`、`DBType`、连接配置、连接测试结果、typed errors 和内存 adapter registry。
- `capability`：事务、保存点、外键、批量插入、bulk load、RETURNING、upsert、catalog/schema、JSON、array、UUID、enum、generated/identity column、identifier length 等能力字段。
- `schema`：database、namespace、table、view、column、约束、index、logical type 和 Raw metadata 保留结构。
- `dialect`：identifier quoting、placeholder、batch insert request 和 statement contract，只生成 SQL 文本与参数。
- `introspect`：窄连接接口和 canonical schema 扫描 contract。
- `typex`：native type 到 logical type 的 mapper contract。
- `fakes`：test-only adapter、dialect、introspector 和 mapper，可配置失败路径并记录调用。

## 明确非目标

本阶段不实现 MySQL、PostgreSQL 或其他生产 adapter；不打开真实数据库连接；不执行 SQL；不持久化连接配置；不实现 Wails binding、前端页面、Schema API、writer、事务编排、执行引擎或数据库驱动集成。

## 能力协商

业务层应优先通过 `capability.Capabilities` 选择策略，例如事务、外键、RETURNING 或批量写入能力，而不是在主路径按数据库类型硬编码分支。MySQL/PostgreSQL 示例能力仅用于验证差异表达，不表示真实 adapter 已完成。

## 隐私约束

`core.ConnectionConfig` 是调用边界，可能临时携带 password、DSN 或 options。当前 DBX contract 不保存这些字段，不在错误信息中输出敏感 DSN 片段、密码、SQL 参数值或真实用户 SQL。Raw metadata 只用于诊断和未来 adapter 扩展；对外暴露前必须重新评估隐私边界。

## 后续扩展

真实数据库支持应进入后续 adapter spec，例如 `internal/dbx/adapters/mysql` 或 `internal/dbx/adapters/postgres`。Schema API、批量 writer、执行引擎和 UI 工作流也应由各自 spec 承担，不能在本接口阶段越界实现。

## 占位与后续 spec

`internal/dbx` 已从纯占位目录升级为接口和值对象边界；仍保留部分占位模块用于后续真实数据库 adapter、Schema API、writer、执行引擎和 UI 工作流。后续 spec 应继续明确这些模块的所有权与演进路径。
