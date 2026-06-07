# LoomiDBX 工程骨架

当前仓库已建立 Wails v3 + Go + Vue 3 的最小桌面应用工程骨架，用于支撑后续 spec 按边界增量实现。

## 快速入口

- `go run ./scripts/doctor.go`：检查 Go、Node.js、npm、wails3 和平台前置条件。
- `npm --prefix frontend install`：安装前端依赖。
- `go test ./...`：运行当前 Go 骨架测试。
- `npm --prefix frontend run test`：运行前端类型检查。

如已安装 Task，可使用 `Taskfile.yml` 中的 `setup`、`doctor`、`dev`、`build`、`format`、`lint`、`test` 统一命令。

## 结构边界

- Go 后端入口：`main.go`、`app.go`、`internal/`。
- 前端工程：`frontend/`。
- Wails generated 绑定落位：`frontend/generated/`。
- 测试落位：`tests/` 与靠近 Go 包的单元测试。
- 结构说明：`docs/architecture/project-structure.md`。

## 后续 spec

配置系统、本地存储、数据库方言、测试工具链、生成引擎、生成器和业务 UI 由后续 spec 继续实现。当前工程骨架只提供落位和 bootstrap 验证，不声明这些完整能力已经完成。

## 隐私边界

工程骨架阶段不要求真实数据库凭据、Schema、生成数据、Project 配置、用户 SQL 或远端账号数据，也不会上传这些本地产品数据。
