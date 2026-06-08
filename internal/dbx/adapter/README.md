# Adapter 落位说明

本目录预留给后续真实数据库 adapter 子包，例如 MySQL 或 PostgreSQL 实现。

当前统一入口定义在 `internal/dbx/core.Adapter`：每个 adapter 负责暴露数据库类型、显示名称、capabilities、connection test、dialect、introspector 和 type mapper。

本阶段不包含生产数据库 adapter，不引入真实驱动，不保存凭据，也不要求测试使用真实连接。需要测试 adapter 调用路径时，请使用 `internal/dbx/fakes`。

## 占位与后续 spec

本目录仍是生产数据库 adapter 的占位位置；MySQL、PostgreSQL 或其他真实 adapter 的所有权应由后续 spec 明确。
