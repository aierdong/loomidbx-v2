# Storage 模块

`internal/storage` 是 `phase-01-local-storage-strategy` 的本地存储基础设施边界，已从工程骨架占位演进为当前 spec 的实现边界。它只消费配置系统已经解析完成的 `dataDir` 和运行模式，不重新解析环境变量、用户目录或配置文件路径；后续 spec 只能在这些边界上追加业务迁移和仓储能力。

## 本模块拥有

- 本地数据目录布局：主业务 SQLite 文件为 `loomidbx.db`，并在数据目录下预留 `migrations/`、`tmp/` 和 `backups/`。
- 初始化入口：`Bootstrapper.Initialize` 创建布局目录、打开主 SQLite 文件、执行基础迁移并返回脱敏诊断视图。
- 迁移基础设施：`internal/storage/migration` 定义递增编号迁移、排序、成功记录和失败不落成功记录的语义。
- 数据分类策略：普通配置、结构化本地业务数据和敏感信息分别归属配置系统、SQLite 和 SecretStore 边界。
- 敏感信息边界：`SecretRef` 可以持久化为引用，`SecretValue` 不得进入普通配置或普通业务表；默认 `UnavailableSecretStore` 返回不可用错误。
- 稳定错误码：路径、打开、迁移和 secret store 失败均返回可诊断且脱敏的错误。

## 文件布局

```text
<dataDir>/
  loomidbx.db
  migrations/
  tmp/
  backups/
```

`dataDir` 必须来自 `internal/config` 输出。测试必须使用 `t.TempDir()` 或同等隔离目录，不能读写真实用户数据目录。

## 数据分类

- 普通配置：主题、语言、数据目录、开发或测试选项，继续由 `internal/config` 管理。
- 结构化本地业务数据：连接元数据、Schema 缓存、字段规则、Project 配置和执行历史，由后续规格通过 SQLite 迁移和 Repository 扩展。
- 敏感信息：数据库密码、token、密钥和类似凭据，只能通过 secret store 接口与凭据引用表达，不能明文写入普通配置或普通业务表。

## 当前非目标

本模块当前不创建完整连接、Schema、字段规则、Project 或执行历史业务表，不实现平台 keychain、加密文件、云同步、UI 设置页、目标数据库写入或生成数据保存。后续业务规格追加迁移时必须复用现有迁移记录和错误语义。
