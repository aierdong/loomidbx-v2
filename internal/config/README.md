# Config 模块

本目录从工程骨架占位演进为 `phase-01-config-system` 的应用级配置模块，负责配置模型、默认值、路径解析、环境覆盖、加载、校验、保存和设置视图契约；后续 spec 只能在各自边界内复用这些输出。

配置系统输出的 `SettingsView.paths.dataDir` 与 `SettingsView.development.mode` 可供 `phase-01-local-storage-strategy` 复用。SQLite 业务存储、schema、迁移、Repository 和目标数据库访问都属于相邻 spec 范围，不在本模块创建或管理。

本模块不要求用户提供数据库凭据或远端账号数据。普通配置文件不得保存数据库密码、账号 token、LLM API key、Schema、Project、生成数据或用户 SQL。
