# Phase 02 数据库 Schema 层级模型 domain-only 冒烟验证记录

验证日期：2026-06-10

## 规格边界

`phase-02-database-schema-model` 只交付 Go 后端 Domain 层的数据库 Schema 层级模型、枚举、字段级校验和 JSON 序列化合同。

本 smoke 记录不启动 Wails、Vue、真实数据库、执行引擎或 API/UI 工作流，也不验证真实扫描、重扫 diff、表字段约束或桌面壳 liveness。

## 可执行验证命令

| 命令 | 用途 | 证据范围 |
| --- | --- | --- |
| `task test` | 项目测试入口，执行 Go 全量测试与前端 deterministic 样例测试。 | 项目级回归证据；不代表真实数据库、桌面启动或端到端工作流已验证。 |
| `go test ./internal/domain/schema` | 本规格直接领域包测试。 | 覆盖 `DbCatalog`、`DbSchema`、`SchemaIdentity`、枚举、字段级校验、隐式 Schema JSON 合同和 out-of-scope 边界测试。 |
| `go test ./...` | 当环境不支持 Task 时的完整 Go 测试替代证据。 | Go 后端全包回归证据；不执行前端样例测试，也不代表 Wails 桌面启动。 |

## 已执行验证结果

| 命令 | 结果 | 备注 |
| --- | --- | --- |
| `gofmt -w internal/domain/schema/*.go` | 通过 | 领域包 Go 文件格式化后无剩余差异。 |
| `go test -count=1 ./internal/domain/schema` | 通过 | 本规格领域包直接证据。 |
| `go test ./...` | 通过 | 完整 Go 测试替代证据。 |

## 结论

本规格可用领域包测试和完整 Go 测试作为 domain-only 运行时替代证据。该证据证明 schema domain 包可编译、测试可运行、字段合同和校验规则稳定；它不暗示已经完成 Wails/Vue 启动、真实数据库扫描、API/UI 工作流或执行引擎 liveness 验证。
