# Introspect Schema 扫描契约

`introspect.Introspector` 定义从数据库 metadata 到 canonical schema 的扫描边界，输出 `schema.Database` 及其中的 namespace、table、view、column、约束、index 和 Raw metadata。

本阶段只定义窄连接接口和 options，不实现真实 metadata SQL，不读取凭据存储，不缓存用户数据库 Schema，也不依赖 Wails runtime。

测试应通过 fake connection 或 `internal/dbx/fakes.Introspector` 验证调用路径和 deterministic schema，而不是连接真实数据库。

## 占位与后续 spec

当前 introspect 只定义扫描 contract 和 fake 验证入口，真实 metadata SQL 与 Schema 缓存仍是占位能力；具体实现应由后续 spec 承担。
