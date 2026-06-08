# 测试目录

本目录记录跨层测试和 smoke 验证的落位。Go 单元测试应靠近后端包，前端 deterministic 样例验证在 `frontend/` 内执行，跨层 smoke 验证记录放在 `tests/smoke/`。

## 当前覆盖

- Go 单元测试：`go test ./...` 或 `task test:go`，覆盖 bootstrap、config、本地存储基础、数据库方言接口 mock 和工程结构边界等 deterministic 行为。
- Go 静态检查：`go vet ./...` 或 `task lint:go`。
- 前端类型检查：`npm --prefix frontend run typecheck` 或 `task lint:frontend`，覆盖 TypeScript/Vue 源码和配置。
- 前端 deterministic 样例验证：`npm --prefix frontend run test` 或 `task test:frontend`，当前使用 Vitest 覆盖 API client 结果转换边界。
- Wails 骨架验证：`task doctor` 检查本地工具链，`task build` 执行标准桌面构建；缺少 Wails 或平台依赖时可记录 doctor 诊断并运行 `task build:fallback`。

## 隐私边界

测试和验证命令不需要真实数据库凭据、Schema、生成数据、Project 配置、用户 SQL 或远端账号数据。后续业务测试若需要 fixture，必须使用非敏感 synthetic fixture。

## 延后范围

以下能力不属于 Phase 1 测试工具链，延后到对应 spec 或 Phase 9：

- 业务覆盖率目标和覆盖率 gate。
- 真实数据库集成测试矩阵。
- 生成器契约测试。
- 执行引擎集成测试。
- API 契约测试。
- UI E2E 和完整工作流自动化。
- 远端 CI、遥测或可观测性平台。
