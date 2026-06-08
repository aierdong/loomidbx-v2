# Roadmap

## Overview

LoomiDBX 是一款桌面端数据库模拟/合成数据生成工具，面向市场/售前、数据库开发测试人员和软件开发人员，帮助用户连接数据库、扫描 Schema、配置字段生成规则、组织 Project，并将符合关系与约束的合成数据写回目标数据库。

项目采用 Golang + Vue 3 + Wails：Go 后端承载领域模型、数据库适配、生成引擎、本地存储与服务层；Vue 3 前端承载桌面 UI 与工作流；Wails 负责本地桌面壳、Go 方法绑定、前后端桥接与运行时事件。开发过程按 Phase 控制上下文，真正的实现 spec 使用 Phase 内的小边界，而不是把 Phase 本身作为大 spec。

## Approach Decision

- **Chosen**: Phase 分组 + 小粒度 spec 的增量路线。
- **Why**: `docs/phase.md` 已定义 0-9 个开发阶段和依赖关系；直接把 Phase 作为 spec 会过大，不利于实现、测试和 review。以 Phase 管理阅读范围和依赖顺序，以 Phase 内建议拆分项作为真正 spec，可以降低上下文负担，并让每个 spec 独立设计、实现和验收。
- **Rejected alternatives**:
  - 单一全量 spec：范围过大，会混合工程骨架、领域模型、执行引擎、生成器、API、UI 和测试，难以评审。
  - 10 个 Phase 直接对应 10 个 spec：仍然过粗，特别是领域模型、API、UI、生成器阶段都包含多个独立边界。
  - 一次性展开全部 60+ 子 spec：启动成本过高，早期架构经验尚未反馈到后续 spec，容易产生过早设计。

## Scope

- **In**:
  - 首发主线：数据库连接管理、Schema 扫描、字段生成规则、Project 组织、生成执行、写回目标数据库、执行历史、主要 UI 工作流。
  - 技术主线：Golang + Vue 3 + Wails 桌面应用，本地 SQLite / 配置文件，本地服务层与 Wails binding，数据库方言抽象。
  - 当前首批 spec：Phase 1 的工程骨架与基础架构，包括 `phase-01-project-structure`、`phase-01-config-system`、`phase-01-local-storage-strategy`、`phase-01-database-dialect-interface`、`phase-01-test-tooling`。
- **Out**:
  - AI 生成能力不进入首发主线。
  - 不在首期实现所有数据库方言，优先为 MySQL 和 PostgreSQL 验证抽象。
  - 不在工程骨架阶段实现完整 Schema 扫描、执行引擎、生成器或 UI 页面。
  - 不把 Wails binding 当成业务逻辑层；业务逻辑应保留在 Go service/domain 层。

## Constraints

- 必须遵循 `docs/agent/README.md` 的上下文控制规则：按当前 Phase 和当前 spec 读取最小必要文档，不跨阶段实现未来能力。
- API 契约在 Wails 形态下应理解为传输无关服务契约：Go Service / Facade + Wails Binding + Vue API Client，而不是必须启动本地 HTTP 服务。
- 数据库差异必须下沉到 Adapter、Dialect、Introspector、TypeMapper、Capabilities；业务层避免按数据库类型硬编码分支。
- 本地隐私边界：数据库连接信息、Schema、表字段、生成规则、Project 配置、生成数据和用户 SQL 不应上传到远端账号服务。
- 首期数据库方言抽象以 MySQL 和 PostgreSQL 为优先验证目标，并为 Oracle、SQL Server、SQLite、ClickHouse、TiDB、Hive 预留扩展空间。

## Boundary Strategy

- **Why this split**: Phase 用于控制开发顺序和阅读范围；spec 用于可独立实现和验收的小边界。首批只展开 Phase 1，是为了先稳定 Wails 工程结构、配置、本地存储、数据库抽象接口和测试工具链，再让后续领域模型、引擎、API、UI 有清晰落位。
- **Shared seams to watch**:
  - Go 后端包结构与 Wails binding 的边界。
  - 本地配置文件、SQLite 存储和敏感连接信息的分工。
  - 数据库方言接口与后续 Schema 模型、执行引擎写入适配器的衔接。
  - 测试工具链需要同时覆盖 Go 后端、Vue 前端和 Wails 集成构建，但不要在首批 spec 中实现业务测试全量覆盖。

## Specs (dependency order)

- [x] phase-01-project-structure -- 建立 Wails + Go + Vue3 的工程骨架、目录结构、模块边界、基础命令和后续模块落位约定。Dependencies: none
- [x] phase-01-config-system -- 建立应用配置模型、默认值、环境/开发配置覆盖、配置加载与保存接口。Dependencies: phase-01-project-structure
- [x] phase-01-local-storage-strategy -- 定义本地 SQLite / 配置文件的数据目录、迁移策略、Repository 落位和敏感信息处理边界。Dependencies: phase-01-project-structure, phase-01-config-system
- [x] phase-01-database-dialect-interface -- 定义数据库 Adapter、Dialect、Introspector、TypeMapper、Capabilities 的最小接口和 mock 实现边界。Dependencies: phase-01-project-structure
- [x] phase-01-test-tooling -- 建立 Go/Vue/Wails 相关的格式化、lint、单元测试、构建验证命令和最小样例测试。Dependencies: phase-01-project-structure

## Later Phase Backlog

### Phase 0：项目最小背景

- project-brief
- glossary
- scope-and-non-goals

### Phase 1：项目骨架与基础架构（当前批次）

- phase-01-project-structure
- phase-01-config-system
- phase-01-local-storage-strategy
- phase-01-database-dialect-interface
- phase-01-test-tooling

### Phase 2：核心领域模型与 Schema

- connection-model
- database-schema-model
- table-field-constraint-model
- relation-model
- field-generation-rule-model
- project-model
- generation-job-model

### Phase 3：数据生成执行引擎

- execution-lifecycle
- dependency-graph-and-topological-sort
- row-count-planning
- generation-context
- batch-generation-loop
- batch-writer-adapter
- execution-result-and-error-model

### Phase 4：生成器框架与注册表

- generator-interface
- generator-definition-schema
- generator-registry
- generator-parameter-validation
- generator-metadata-query
- generator-contract-tests

### Phase 5：内置生成器实现

- string-generators
- number-generators
- boolean-generators
- datetime-generators
- enum-generators
- id-generators
- relation-generators
- computed-field-generators

### Phase 6：外部数据源生成器

- external-source-config-model
- static-list-source-generator
- csv-source-generator
- http-source-generator
- database-source-generator
- external-source-validation-and-preview

### Phase 7：API 与服务层

- connection-api
- schema-introspection-api
- field-rule-api
- project-api
- generator-metadata-api
- generation-job-api
- execution-history-api
- error-response-contract

### Phase 8：UI 与用户工作流

- app-shell-and-routing
- login-page
- home-page
- projects-page
- schema-management-page
- settings-page
- field-rule-editor
- generation-job-progress-view

### Phase 9：测试、可观测性与验收

- unit-test-strategy
- generator-contract-tests
- engine-integration-tests
- api-contract-tests
- ui-workflow-tests
- execution-history-and-progress
- error-reporting-and-logging
- release-acceptance-checklist
