# Brief: phase-01-test-tooling

## Problem

LoomiDBX 同时包含 Go 后端、Vue3 前端和 Wails 桌面集成。若没有统一的测试、格式化、lint 和构建命令，后续每个 spec 难以提供可靠验证证据。

## Current State

当前仓库尚未建立源码级测试工具链。Phase 1 明确要求建立测试框架、lint/format/build 命令，为后续领域模型、数据库适配、生成器、API 和 UI 开发提供基础。

## Desired Outcome

完成后，项目具备最小可用的验证命令集合：Go 单元测试、Vue 单元测试或类型检查、格式化/lint、Wails 构建或检查命令，以及至少一个后端和前端样例测试。后续 spec 能明确引用这些命令作为验收依据。

## Approach

基于 `phase-01-project-structure` 的工程形态选择轻量测试工具链。Go 使用标准 `go test` 起步；Vue3 使用项目选择的包管理器和测试/类型检查工具；Wails 使用官方构建或检查命令验证集成。首批只建立工具链，不追求完整业务覆盖。

## Scope

- **In**:
  - 定义 Go 后端测试命令和样例测试。
  - 定义 Vue3 前端类型检查、lint/format 和样例测试策略。
  - 定义 Wails 构建/检查命令。
  - 在 README 或开发文档中记录常用验证命令。
  - 确保后续 spec 能引用最小验证路径。
- **Out**:
  - 不补齐业务模块完整测试。
  - 不实现生成器契约测试全量集合。
  - 不实现 E2E UI 自动化全套流程。
  - 不引入重型可观测性平台。

## Boundary Candidates

- 工具链 spec 负责验证基础设施。
- 各业务 spec 负责自己的单元/集成测试用例。
- Phase 9 负责跨模块测试、可观测性和最终验收补齐。

## Out of Boundary

- 业务测试覆盖率目标。
- 真实数据库集成测试矩阵。
- 发布流水线和安装包签名。

## Upstream / Downstream

- **Upstream**: `phase-01-project-structure`、`docs/agent/01-architecture-bootstrap.md`。
- **Downstream**: 后续所有实现 spec，特别是 generator-contract-tests、engine-integration-tests、api-contract-tests、ui-workflow-tests。

## Existing Spec Touchpoints

- **Extends**: 无。
- **Adjacent**: `phase-01-project-structure`、`phase-01-config-system`、`phase-01-database-dialect-interface`。

## Constraints

测试工具链应轻量、可本地运行，并适合 agent 在每次改动后运行最小相关验证。不要为了工具链引入与项目规模不匹配的复杂 CI/CD 或重型平台。
