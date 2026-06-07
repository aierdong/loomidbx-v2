# 工程结构与边界

本文说明 Phase 1 工程骨架中各目录的职责，帮助后续 spec 判断新增代码应放入哪个边界。

## 根目录

- `go.mod`：Go 模块定义。
- `main.go`：桌面应用进程入口。
- `app.go`：前端可见的薄 facade，目前只暴露 bootstrap 状态。
- `wails.json`：Wails v3 项目配置。
- `Taskfile.yml`：统一 setup、doctor、dev、build、format、lint、test 命令入口。

## 后端模块

- `internal/bootstrap/`：骨架级状态服务，deterministic，不读取业务数据。
- `internal/domain/`：领域模型占位，由后续 spec 实现。
- `internal/service/`：应用服务占位，由后续 spec 实现业务用例编排。
- `internal/repository/`：仓储接口占位。
- `internal/config/`：配置系统占位，完整能力归属 `phase-01-config-system`。
- `internal/storage/`：本地存储占位，完整能力归属 `phase-01-local-storage-strategy`。
- `internal/dbx/`：数据库兼容抽象占位，完整能力归属 `phase-01-database-dialect-interface`。
- `internal/engine/`：生成执行引擎占位。
- `internal/generator/`：生成器框架占位。

## 前端模块

- `frontend/src/pages/`：页面/workflow。
- `frontend/src/components/`：可复用组件。
- `frontend/src/stores/`：页面状态样例。
- `frontend/src/router/`：路由配置。
- `frontend/src/api/`：封装 generated 绑定的唯一前端 API client 边界。
- `frontend/src/types/`：共享 DTO。
- `frontend/src/styles/`：global styles、tokens 和 utilities。
- `frontend/src/ui/primevue/`：PrimeVue styled mode + Aura preset 配置。

## 生成绑定

`frontend/generated/` 是 Wails generated 绑定落位。页面、组件和 store 不直接依赖该目录；应通过 `frontend/src/api/` 调用。

## 测试目录

`tests/` 保存跨层和 smoke 验证说明。Go 单元测试靠近包目录，前端类型检查通过 `frontend` 脚本执行。完整测试工具链由 `phase-01-test-tooling` 补齐。

## Deferred 能力

当前 spec 不实现完整配置、本地存储、真实数据库连接、Schema 扫描、生成引擎、生成器、业务 UI、远端账号或 AI 生成能力。相关目录只提供占位与说明，后续 spec 实现前不得把这些占位当作已完成产品功能。
