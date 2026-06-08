# Research & Design Decisions

## Summary
- **Feature**: `phase-01-test-tooling`
- **Discovery Scope**: Extension
- **Key Findings**:
  - `phase-01-project-structure` 已建立 Wails + Go + Vue3 骨架、`Taskfile.yml` 命令入口和 bootstrap 样例测试，本 spec 应扩展这些入口而不是重建工程结构。
  - 当前前端已固定 npm、Vue 3、TypeScript、Vite、vue-tsc 和 Prettier；最小前端测试可先采用类型检查与小型 TypeScript/Vue 单元样例，避免过早引入重型 UI/E2E。
  - Wails 集成验证应通过 `doctor`、前端 build、`wails3 build` 或当前平台 fallback 表达；真实业务工作流和数据库操作不属于本 spec。

## Research Log

### 现有工程骨架与命令入口
- **Context**: 本 spec 依赖 `phase-01-project-structure`，需要确认实际源码与命令状态。
- **Sources Consulted**: `Taskfile.yml`、`README.md`、`go.mod`、`app.go`、`app_test.go`、`internal/bootstrap/`、`internal/projectstructure/`。
- **Findings**:
  - 根目录已有 Go 模块 `github.com/gerdong/loomidbx`，Go 版本为 `1.25`。
  - `Taskfile.yml` 已提供 `setup`、`doctor`、`dev`、`build`、`format`、`lint`、`test` 入口。
  - 后端已存在 bootstrap 状态测试、doctor 测试和结构边界测试。
- **Implications**:
  - 本 spec 的实现应收敛命令语义并补齐测试工具链说明，不应改动 Phase 1 工程骨架职责。
  - Go 侧可使用 `go test ./...`、`gofmt`、`go vet ./...` 作为初始标准。

### 前端工具链状态
- **Context**: 前端验证需要基于现有 package manager 和脚本。
- **Sources Consulted**: `frontend/package.json`、`frontend/tsconfig.json`、`frontend/vite.config.ts`、`frontend/src/`。
- **Findings**:
  - 前端 package manager 已固定为 `npm@11`。
  - `lint` 和 `test` 目前都映射到 `vue-tsc --noEmit`，`format` 使用 `prettier --check .`。
  - 前端已有 API client、result 类型、bootstrap store、BootstrapPage 和 PrimeVue theme 落位。
- **Implications**:
  - 设计应把前端最小验证分成 typecheck、format、lint 和 sample validation，但不强制立即引入浏览器 E2E。
  - 若引入 Vitest，应只服务于确定性的 API/type/component 边界样例，并在文档中说明依赖原因。

### Wails 与桌面集成边界
- **Context**: 桌面验证需要在缺少本地平台依赖时可诊断。
- **Sources Consulted**: `scripts/doctor.go`、`wails.json`、`Taskfile.yml`、`phase-01-project-structure/design.md`。
- **Findings**:
  - `doctor` 已检查 Go、Node.js、npm、`wails3` 和平台提示。
  - `build` 入口依赖 `doctor`，并执行前端 build 与 `wails3 build`。
  - 平台专用 fallback 任务使用前端 build 和 `go build`，适合在 Wails CLI 不可用时保留骨架级构建证据。
- **Implications**:
  - 桌面验证应明确分为环境检查、Wails build 和平台 fallback，避免把业务流程当作桌面验收。
  - 缺少 Wails 或平台依赖时，应把失败视为可诊断状态，而不是隐藏或跳过。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 扩展现有 Taskfile 命令 | 在既有 `Taskfile.yml` 中固化统一验证入口 | 与工程骨架一致，后续 spec 易引用 | 需要避免命令名语义漂移 | 选中 |
| 新增独立测试脚本目录 | 通过脚本封装所有验证逻辑 | 可承载复杂检查 | Phase 1 过重，可能重复 Taskfile | 暂不采用 |
| 立即引入完整 E2E/覆盖率平台 | 建立浏览器自动化、覆盖率 gate 和 CI | 验证能力强 | 超出当前范围，依赖重 | 拒绝 |

## Design Decisions

### Decision: 使用现有命令入口承载最小验证矩阵
- **Context**: 工程骨架已有统一命令入口，后续 spec 需要稳定引用。
- **Alternatives Considered**:
  1. 继续使用散落命令 — 简单但不可复用。
  2. 新建独立脚本体系 — 灵活但与现有 Taskfile 重复。
  3. 扩展 `Taskfile.yml` — 与骨架一致。
- **Selected Approach**: 使用 `Taskfile.yml` 中的 `doctor`、`format`、`lint`、`test`、`build` 和聚合验证任务作为公共入口，必要时补充分层子命令。
- **Rationale**: 最小化新增复杂度，并让后续 spec 能用同一命令表述验收。
- **Trade-offs**: Taskfile 成为开发体验入口；未安装 Task 的用户仍需 README 中的等价命令。
- **Follow-up**: 实现阶段需确保 README 或开发文档记录 Task 与等价原生命令。

### Decision: 前端先以类型检查和确定性样例测试为主
- **Context**: 前端业务 UI 尚未展开，完整 E2E 不适合当前阶段。
- **Alternatives Considered**:
  1. 只运行 `vue-tsc` — 轻量但样例测试证据弱。
  2. 引入 Vitest 做小型单元测试 — 可验证 API/type/store 边界。
  3. 引入 Playwright 做 E2E — 过重且业务流程不足。
- **Selected Approach**: 保留 `vue-tsc` 为强制入口；若新增测试依赖，优先选择 Vitest 级别的轻量样例，不建立 E2E。
- **Rationale**: 符合 Phase 1 “工具链先行、业务覆盖后补”的边界。
- **Trade-offs**: UI 工作流缺少端到端证据，延后到 Phase 8/9。
- **Follow-up**: 后续 UI spec 根据实际页面补充工作流测试。

### Decision: Wails 验证以 doctor 和 build 证据为边界
- **Context**: 桌面壳和平台依赖在不同机器上可用性不同。
- **Alternatives Considered**:
  1. 每次强制运行完整 `wails3 build` — 证据强但环境要求高。
  2. 只运行 `go build` 和前端 build — 环境友好但无法验证 Wails CLI。
  3. 使用 doctor + Wails build + documented fallback — 平衡严格性和可诊断性。
- **Selected Approach**: 将 `doctor` 作为前置诊断，`build` 作为标准桌面验证，平台 fallback 作为缺少 Wails 时的骨架构建证据。
- **Rationale**: 保持本地可运行，同时清晰暴露缺失前置条件。
- **Trade-offs**: fallback 不能替代完整 Wails build，文档必须说明限制。
- **Follow-up**: 发布和安装包验证由后续 release acceptance spec 处理。

## Risks & Mitigations
- 命令名称存在但覆盖范围被误读 — 在开发文档中标注 mandatory、optional、environment-dependent 和 scope。
- 前端 `test` 只做 typecheck 导致样例测试不足 — 任务中加入最小确定性前端样例验证。
- Wails CLI 不可用导致本地验证阻塞 — 保留 doctor 诊断和平台 fallback，并明确不能替代完整桌面构建。
- 工具链过早膨胀 — 禁止本 spec 引入完整 E2E、覆盖率 gate、远端 CI 或可观测性平台。

## References
- `.kiro/steering/product.md` — 产品范围、隐私边界和 Phase 纪律。
- `.kiro/steering/tech.md` — Go + Vue 3 + Wails 技术方向与测试方向。
- `.kiro/steering/structure.md` — 工程结构、依赖方向和测试落位。
- `.kiro/specs/phase-01-project-structure/design.md` — 上游骨架和命令边界。
