# 开发命令、验证与隐私说明

本文记录工程骨架阶段可用的命令入口，以及哪些验证能力延后到 `phase-01-test-tooling`。

## 命令清单

| 命令 | 入口 | 当前能力 |
| --- | --- | --- |
| setup | `task setup` 或 `npm --prefix frontend install` | 安装前端依赖。 |
| doctor | `task doctor` 或 `go run ./scripts/doctor.go` | 检查 Go 1.25+、Node.js、npm、wails3 和平台提示。 |
| dev | `task dev` 或 `wails3 dev` | 依赖本机 Wails v3 工具链；用于启动桌面开发模式。 |
| build | `task build` 或 `npm --prefix frontend run build && wails3 build` | 构建前端并交给 Wails 构建桌面应用。 |
| format | `task format` | 格式化 Go 代码并检查前端格式；完整格式化策略 deferred 到 `phase-01-test-tooling`。 |
| lint | `task lint` | 当前执行 `go vet ./...` 和前端类型检查；完整 lint 规则 deferred 到 `phase-01-test-tooling`。 |
| test | `task test`、`go test ./...`、`npm --prefix frontend run test` | 当前覆盖 Go 骨架测试和前端类型检查；覆盖率、契约测试、E2E deferred 到 `phase-01-test-tooling`。 |

## Doctor 隐私边界

`doctor` 只检查本地工具链和平台前置条件，不读取、不上传以下本地产品数据：

- 数据库凭据和连接信息。
- Schema、表、字段和约束元数据。
- 生成数据、Project 配置和字段生成规则。
- 用户 SQL。
- 远端账号数据。

## 验证范围

当前骨架阶段的验证目标是证明项目结构、命令入口、bootstrap 示例和前后端基础配置可被识别。它不要求真实数据库、本地存储、完整配置系统、Schema 扫描、生成引擎或业务 UI 已经完成。
