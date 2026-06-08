# Dialect SQL 方言契约

`dialect.Dialect` 提供最小 SQL 方言原语：identifier quoting、placeholder 和 batch insert statement 构建。

`BuildInsert` 只返回 SQL 文本和参数数组，不执行 SQL、不管理事务、不持久化写入结果，也不实现 COPY、LOAD DATA、upsert 自动化或数据库特定 writer 策略。

当方言不能构建请求的 SQL 形式时，应返回 typed unsupported dialect operation error，调用方可用 `errors.Is` 匹配错误类别。

## 占位与后续 spec

当前 dialect 只定义接口契约，真实数据库 SQL 生成规则仍是占位能力；数据库特定方言实现应由后续 spec 承担。
