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
  - 已完成批次：Phase 1 的工程骨架与基础架构，以及 Phase 2 的核心领域模型与 Schema。
  - 下一批 spec：Phase 3 的数据生成执行引擎，包括执行生命周期、依赖图与拓扑排序、行数规划、生成上下文、批量生成循环、批量写入适配和执行结果/错误模型。
- **Out**:
  - AI 生成能力不进入首发主线。
  - 不在首期实现所有数据库方言，优先为 MySQL 和 PostgreSQL 验证抽象。
  - Phase 3 不实现所有内置生成器、UI 进度界面、完整 API 层或复杂可观测性。
  - 不把 Wails binding 当成业务逻辑层；业务逻辑应保留在 Go service/domain/engine 层。

## Constraints

- 必须遵循 `docs/agent/README.md` 的上下文控制规则：按当前 Phase 和当前 spec 读取最小必要文档，不跨阶段实现未来能力。
- API 契约在 Wails 形态下应理解为传输无关服务契约：Go Service / Facade + Wails Binding + Vue API Client，而不是必须启动本地 HTTP 服务。
- 数据库差异必须下沉到 Adapter、Dialect、Introspector、TypeMapper、Capabilities；业务层和 engine 层避免按数据库类型硬编码分支。
- 本地隐私边界：数据库连接信息、Schema、表字段、生成规则、Project 配置、生成数据和用户 SQL 不应上传到远端账号服务。
- 首期数据库方言抽象以 MySQL 和 PostgreSQL 为优先验证目标，并为 Oracle、SQL Server、SQLite、ClickHouse、TiDB、Hive 预留扩展空间。

## Boundary Strategy

- **Why this split**: Phase 用于控制开发顺序和阅读范围；spec 用于可独立实现和验收的小边界。Phase 1 已稳定 Wails 工程结构、配置、本地存储、数据库抽象接口和测试工具链。Phase 2 已完成连接、Schema 树、约束/字段、关系、字段规则、Project 和执行任务配置等持久化领域模型。Phase 3 在此基础上把执行引擎拆成生命周期、计划排序、行数计算、上下文、批处理、写入适配和结果错误表达，避免把调度、算法、生成循环和数据库写入混入单一大 spec。
- **Shared seams to watch**:
  - Go 后端包结构与 Wails binding 的边界。
  - 本地配置文件、SQLite 存储和敏感连接信息的分工。
  - 数据库方言接口与 Phase 2 Schema 持久化模型的映射边界：`internal/dbx/schema` 是扫描快照，Phase 2 domain/storage 模型是本地业务持久化表达，两者不要混同。
  - 字段生成规则归属 Schema 层，Project 只保存表级执行配置和运行时关系覆盖。
  - Phase 2 可以定义模型、枚举、验证、序列化和本地存储迁移，但不实现具体生成器、执行引擎、API 端点或 UI 表单。
  - Phase 3 engine 可以定义最小生成器调用接口和写入适配接口，但不实现 Phase 4/5 的完整生成器注册表和内置生成器集合。
  - 执行进度事件、历史查询 API 和 UI 展示属于后续 Phase 7/8/9；Phase 3 只保留必要状态、结果和错误边界。

## Specs (dependency order)

- [x] phase-01-project-structure -- 建立 Wails + Go + Vue3 的工程骨架、目录结构、模块边界、基础命令和后续模块落位约定。Dependencies: none
- [x] phase-01-config-system -- 建立应用配置模型、默认值、环境/开发配置覆盖、配置加载与保存接口。Dependencies: phase-01-project-structure
- [x] phase-01-local-storage-strategy -- 定义本地 SQLite / 配置文件的数据目录、迁移策略、Repository 落位和敏感信息处理边界。Dependencies: phase-01-project-structure, phase-01-config-system
- [x] phase-01-database-dialect-interface -- 定义数据库 Adapter、Dialect、Introspector、TypeMapper、Capabilities 的最小接口和 mock 实现边界。Dependencies: phase-01-project-structure
- [x] phase-01-test-tooling -- 建立 Go/Vue/Wails 相关的格式化、lint、单元测试、构建验证命令和最小样例测试。Dependencies: phase-01-project-structure
- [x] phase-02-connection-model -- 定义连接配置领域模型、数据库类型枚举、敏感字段边界、连接参数序列化与基础校验。Dependencies: phase-01-project-structure, phase-01-config-system, phase-01-local-storage-strategy, phase-01-database-dialect-interface
- [x] phase-02-database-schema-model -- 定义 DbCatalog、DbSchema 等数据库层级模型，以及从 introspection 快照到本地 Schema 树的统一表达。Dependencies: phase-02-connection-model, phase-01-database-dialect-interface
- [x] phase-02-table-field-constraint-model -- 定义 DbTable、DbColumn、TableConstraint 等表、字段和基础约束模型，并覆盖主键、唯一、非空、默认值等表达。Dependencies: phase-02-database-schema-model
- [x] phase-02-relation-model -- 定义 ForeignKey 与 TableRelation 关系模型，覆盖 Parent/Child、BaseTable/JoinTable、倍数范围和物理/逻辑关系标记。Dependencies: phase-02-table-field-constraint-model
- [x] phase-02-field-generation-rule-model -- 定义 GeneratorConfig 字段级规则模型、输出类型映射、配置状态和参数 JSON 边界。Dependencies: phase-02-table-field-constraint-model, phase-02-relation-model
- [x] phase-02-project-model -- 定义 Project、ProjectTable、ProjectTableRelation 等任务组织模型，明确行数状态机、清空策略、执行顺序快照和关系实例化配置。Dependencies: phase-02-relation-model, phase-02-field-generation-rule-model
- [x] phase-02-generation-job-model -- 定义 GenerationJob / ExecutionTask / ExecutionTableResult 等生成任务配置与执行历史领域模型，先完成状态、结果和快照表达，不实现执行引擎。Dependencies: phase-02-project-model
- [x] phase-03-execution-lifecycle -- 定义执行任务从创建、预检、运行、取消、失败到完成的生命周期、状态流转和 engine 入口。Dependencies: phase-02-generation-job-model
- [x] phase-03-dependency-graph-and-topological-sort -- 从 Project 表、外键和逻辑关系构建依赖图，并输出可执行的表级拓扑顺序与循环依赖错误。Dependencies: phase-03-execution-lifecycle, phase-02-relation-model, phase-02-project-model
- [x] phase-03-row-count-planning -- 基于 Project 表级配置、关系倍数和执行策略计算每张表的目标行数与不可满足场景。Dependencies: phase-03-dependency-graph-and-topological-sort
- [x] phase-03-generation-context -- 定义执行期上下文、表/字段规则快照、已生成键值引用和生成器调用所需的只读数据边界。Dependencies: phase-03-row-count-planning, phase-02-field-generation-rule-model
- [x] phase-03-batch-generation-loop -- 实现按执行计划和批次调度字段生成、行组装、关系值填充和批次状态推进的主循环。Dependencies: phase-03-generation-context
- [x] phase-03-batch-writer-adapter -- 定义批量写入目标数据库的 engine 侧适配接口、事务/清空策略边界和写入结果汇总。Dependencies: phase-03-batch-generation-loop, phase-01-database-dialect-interface
- [x] phase-03-execution-result-and-error-model -- 汇总执行结果、表级结果、错误分类、失败范围和最小历史记录边界，供后续 API/UI/可观测性阶段复用。Dependencies: phase-03-batch-writer-adapter

## Later Phase Backlog

### Phase 0：项目最小背景

- project-brief
- glossary
- scope-and-non-goals

### Phase 1：项目骨架与基础架构

- phase-01-project-structure
- phase-01-config-system
- phase-01-local-storage-strategy
- phase-01-database-dialect-interface
- phase-01-test-tooling

### Phase 2：核心领域模型与 Schema

- phase-02-connection-model
- phase-02-database-schema-model
- phase-02-table-field-constraint-model
- phase-02-relation-model
- phase-02-field-generation-rule-model
- phase-02-project-model
- phase-02-generation-job-model

### Phase 3：数据生成执行引擎（下一批）

- phase-03-execution-lifecycle
- phase-03-dependency-graph-and-topological-sort
- phase-03-row-count-planning
- phase-03-generation-context
- phase-03-batch-generation-loop
- phase-03-batch-writer-adapter
- phase-03-execution-result-and-error-model

### Phase 4：生成器框架与注册表

- phase-04-generator-interface
- phase-04-generator-definition-schema
- phase-04-generator-registry
- phase-04-generator-parameter-validation
- phase-04-generator-metadata-query
- phase-04-generator-contract-tests

### Phase 5：内置生成器实现

- phase-05-string-generators
- phase-05-number-generators
- phase-05-boolean-generators
- phase-05-datetime-generators
- phase-05-enum-generators
- phase-05-id-generators
- phase-05-relation-generators
- phase-05-computed-field-generators

### Phase 6：外部数据源生成器

- phase-06-external-source-config-model
- phase-06-static-list-source-generator
- phase-06-csv-source-generator
- phase-06-http-source-generator
- phase-06-database-source-generator
- phase-06-external-source-validation-and-preview

### Phase 7：API 与服务层

- phase-07-connection-api
- phase-07-schema-introspection-api
- phase-07-field-rule-api
- phase-07-project-api
- phase-07-generator-metadata-api
- phase-07-generation-job-api
- phase-07-execution-history-api
- phase-07-error-response-contract

### Phase 8：UI 与用户工作流

- phase-08-app-shell-and-routing
- phase-08-login-page
- phase-08-home-page
- phase-08-projects-page
- phase-08-schema-management-page
- phase-08-settings-page
- phase-08-field-rule-editor
- phase-08-generation-job-progress-view

### Phase 9：测试、可观测性与验收

- phase-09-unit-test-strategy
- phase-09-generator-contract-tests
- phase-09-engine-integration-tests
- phase-09-api-contract-tests
- phase-09-ui-workflow-tests
- phase-09-execution-history-and-progress
- phase-09-error-reporting-and-logging
- phase-09-release-acceptance-checklist
