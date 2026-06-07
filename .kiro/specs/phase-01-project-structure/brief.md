# Brief: phase-01-project-structure

## Problem

LoomiDBX 需要从 greenfield 状态启动开发。开发者需要一个清晰的 Wails + Go + Vue3 工程骨架，避免后续领域模型、数据库适配、生成引擎、服务层和 UI 页面混杂在一起。

## Current State

当前仓库主要包含规划文档与 agent 指南，尚未建立实际应用源码结构。`docs/agent/01-architecture-bootstrap.md` 明确要求先建立项目骨架、模块边界、测试工具链和数据库方言抽象的落位目录。

## Desired Outcome

完成后，仓库应具备可运行或可构建的基础 Wails 应用结构，明确 Go 后端、Vue3 前端 + PrimeVue（Unstyled Mode） + Tailwind、Wails binding、领域层、服务层、适配器层、配置与测试目录的职责边界。后续 spec 能在既定目录中增量添加功能。

## Approach

采用 Wails 官方工程形态作为基础：Go 后端提供 App/Facade 和 service/domain 包，Vue3 前端提供页面、组件、store 和 API client，Wails 负责生成 binding 和运行时事件。该 spec 只建立骨架、命令和边界，不实现业务能力。

## Scope

- **In**:
  - 初始化或规范 Wails + Go + Vue3 工程目录。
  - 定义 Go 后端包结构，例如 app/facade、domain、service、repository、adapter、config、internal 等边界。
  - 定义 Vue3 前端目录，例如 pages、components、stores、api、router、types。
  - 定义 PrimeVue（Unstyled Mode）和 Tailwind 的目录和样式约定。
  - 建立基础启动、构建、格式化、测试命令占位。
  - 明确后续领域模型、数据库方言、执行引擎、生成器和服务层的落位。
- **Out**:
  - 不实现完整配置系统。
  - 不实现数据库连接或 Schema 扫描。
  - 不实现生成执行引擎。
  - 不实现具体 UI 页面。
  - 不实现业务 API 或 Wails 绑定细节，除非作为最小样例。

## Boundary Candidates

- Wails App/Facade 只做前后端桥接，不承载复杂业务逻辑。
- Go service/domain 层承载业务规则，adapter/repository 层承载外部系统和本地存储访问。
- Vue API client 封装 Wails 生成函数，页面不直接散落调用底层 binding。

## Out of Boundary

- 真实数据库方言实现。
- 本地 SQLite schema 设计和迁移实现。
- 登录、设置、项目、Schema 等业务页面。
- 生成器接口、执行引擎和生成算法。

## Upstream / Downstream

- **Upstream**: `docs/agent/README.md`、`docs/phase.md`、`docs/agent/01-architecture-bootstrap.md`、`docs/product_outline.md`、`docs/api-contract.md`。
- **Downstream**: `phase-01-config-system`、`phase-01-local-storage-strategy`、`phase-01-database-dialect-interface`、`phase-01-test-tooling` 以及后续所有业务 spec。

## Existing Spec Touchpoints

- **Extends**: 无。
- **Adjacent**: `phase-01-config-system`、`phase-01-test-tooling`、`phase-01-database-dialect-interface`。

## Constraints

使用 Golang + Vue3 + Wails。API 契约应以 Go Service / Wails Binding / Vue API Client 的方式落地，不应强制引入本地 HTTP 服务。首批实现应保持最小骨架，不跨阶段实现业务功能。
