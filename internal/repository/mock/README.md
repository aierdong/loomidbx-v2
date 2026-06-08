# Mock Repository 替身

本目录提供服务层单元测试使用的仓储替身。替身只模拟当前已经实现的 repository 契约，不声明连接、Schema、Project 或执行历史等业务仓储已经完成。

## 使用规则

- 使用 `NewRepositories` 注入测试需要的 `storage.StorageDiagnostics`。
- 当诊断视图 `Ready=false` 时，替身应返回与真实仓储一致的 `REPOSITORY_STORAGE_UNAVAILABLE`。
- 调用尚未实现的业务仓储能力时，替身应返回 `REPOSITORY_NOT_IMPLEMENTED`，不能返回空成功。
- 测试替身不得读写真实 SQLite 文件或真实用户数据目录。
