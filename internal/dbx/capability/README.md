# Capability 能力模型

`capability.Capabilities` 表达数据库运行时能力和限制，是后续业务层选择策略的依据。

当前能力字段覆盖事务、保存点、外键、延迟约束、批量插入、bulk load、RETURNING、upsert、catalog/schema、JSON、array、UUID、enum、generated/identity column、identifier length、最大参数数和最大批量行数。

`MySQLExample` 与 `PostgreSQLExample` 只用于文档、测试和能力差异验证，不注册真实 adapter，也不声明当前产品已经支持连接 MySQL 或 PostgreSQL。

业务层应基于 capabilities 决策，不应在主路径按数据库类型硬编码分支。

## 占位与后续 spec

当前能力模型保留真实 adapter 能力声明的占位语义；具体数据库能力矩阵应在后续 spec 中随生产 adapter 验证。
