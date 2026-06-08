# Typex 类型映射契约

`typex.Mapper` 将数据库 native type 映射为 `schema.LogicalType`，用于后续 Schema、生成规则和 writer 策略共享稳定类型语义。

Mapper 不打开数据库连接，也不执行 metadata scanning。无法识别 native type 时，应返回 unknown 或配置化 fallback logical type，并保留原始 native type 文本。

`NativeType.Raw` 可保留数据库特定元数据，但不能作为业务主路径替代标准 logical type 字段。

## 占位与后续 spec

当前 type mapper 只定义接口边界和 fake 映射，完整数据库类型矩阵仍是占位能力；具体映射实现应由后续 spec 随真实 adapter 验证。
