# Repository 模块

`internal/repository` 是服务层访问本地存储的接口边界，已从工程骨架占位演进为 `phase-01-local-storage-strategy` 的最小契约。服务层应依赖本目录的接口和工作单元，不直接打开 SQLite 文件、不自行执行迁移，也不读取本地路径细节；后续 spec 只能在该契约上扩展具体业务仓储。

## 本模块拥有

- `Repositories`：服务层可见的仓储集合接口，当前只提供存储诊断读取和未实现业务仓储的稳定错误入口。
- `Factory`：从已初始化的 storage 诊断视图构造仓储集合。
- `UnitOfWork`：最小工作单元边界，用于未来事务接入，当前不暴露底层连接。
- `RepositoryError`：未实现、未迁移、存储不可用等稳定错误语义。
- `mock/`：服务层单元测试使用的仓储替身，必须遵守真实接口的错误语义。

## 使用规则

后续服务层需要本地数据时，应通过 `UnitOfWork.Do` 获取 `Repositories`。业务仓储尚未由对应领域规格实现时，必须返回 `REPOSITORY_NOT_IMPLEMENTED` 或 `REPOSITORY_NOT_MIGRATED`，不能静默创建表或返回空成功。

## 当前非目标

本模块当前不实现连接、Schema、字段规则、Project 或执行历史仓储，不定义完整业务表字段，不处理 Wails binding 或 UI 状态，也不实现安全凭据存储。SQLite 具体连接和迁移执行仍由 `internal/storage` 管理。
