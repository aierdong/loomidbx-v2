# Project Structure

LoomiDBX 当前仓库以文档、Kiro specs 和 Agent 阶段计划为主，应用源码将在 Phase 1 工程骨架中逐步落地。本文记录未来源码组织模式和边界规则，避免后续实现时把 Wails、Go、Vue、数据库适配和业务逻辑混在一起。

## Organization Philosophy

项目采用“阶段控制 + 小粒度 spec + 分层架构”的组织方式：

- Phase 用于控制开发顺序、阅读范围和依赖关系。
- Spec 用于定义可独立实现、测试和 review 的小边界。
- Go 后端按领域和基础设施分层。
- Vue 前端按页面/workflow 与可复用组件分层。
- Wails binding 是桥接层，不是业务层。

新增目录或模块应回答两个问题：

1. 它属于哪个 Phase / spec 的边界？
2. 它承担哪一层职责：domain、service、repository、adapter、engine、generator、frontend UI，还是 binding？

## Documentation and Spec Layout

### Product and Design Documents

**Location**: `docs/`

专题文档是产品、模型、引擎、生成器、API、UI 和数据库兼容方案的来源。实现时不要默认读取整个 `docs/`，应先根据 Phase 和 spec 选择最小必要上下文。

### Agent Phase Guides

**Location**: `docs/agent/`

`docs/agent/README.md` 定义 Agentic Coding 的上下文控制规则。每个 Phase 对应一个阶段文档，说明必须阅读、可选阅读、非目标和建议 spec 拆分。

实现任务时的默认阅读顺序：

1. `docs/agent/README.md`
2. `docs/phase.md`
3. 当前 Phase 的 `docs/agent/<phase>.md`
4. 当前 spec 文件
5. spec 明确要求的专题文档片段

### Steering

**Location**: `.kiro/steering/`

Steering 是长期项目记忆，只记录稳定原则和组织模式，不记录所有文件清单。核心文件为 `product.md`、`tech.md`、`structure.md`，自定义 steering 可用于记录 API、测试、安全等专题规则。

### Specs

**Location**: `.kiro/specs/`

Specs 是任务级实现契约。当前 roadmap 采用 Phase 分组 + Phase 内小粒度 spec 的方式。不要把 10 个 Phase 直接当成 10 个大 spec。

## Planned Source Layout

Phase 1 工程骨架落地时，应建立能承载以下边界的结构。具体目录名可由 spec 最终确定，但职责边界应保持稳定。

### Go Backend

**Purpose**: 承载领域模型、服务层、数据库适配、本地存储、生成器和执行引擎。

推荐组织模式：

```text
backend or app root
  cmd / main entry
  internal/
    domain/
    service/
    store/
    dbx/
    engine/
    generator/
    facade/
```

边界规则：

- `domain` 只表达核心模型和值对象，不依赖 Wails、Vue 或具体数据库驱动。
- `service` 编排业务用例和校验规则，可以依赖 domain、store、adapter、engine。
- `store` 封装本地 SQLite、配置文件和迁移策略。
- `dbx` 或类似包承载数据库方言抽象、适配器、扫描和类型映射。
- `engine` 承载执行生命周期、计划、排序、上下文、批处理和结果模型。
- `generator` 承载生成器接口、定义、注册表、校验和具体实现。
- `facade` 或 Wails app facade 暴露少量稳定方法给前端，不写业务核心逻辑。

### Database Dialect Packages

数据库兼容代码应按“公共抽象 + 数据库实现”组织。

推荐模式：

```text
internal/dbx/
  core/
  schema/
  dialect/
  introspect/
  typex/
  adapters/
    mysql/
    postgres/
```

首期优先 MySQL 和 PostgreSQL。新增数据库时优先新增 adapter 子包和能力声明，不应修改业务层分支。

### Vue Frontend

**Purpose**: 承载桌面 UI、页面状态、表单体验和服务调用封装。

推荐组织模式：

```text
frontend/src/
  app/
  pages/
  components/
  api/
  stores/
  composables/
  types/
```

边界规则：

- `pages` 按工作流或页面组织，例如登录、首页、项目、Schema、设置、执行进度。
- `components` 放可复用 UI 组件，不放核心业务算法。
- `api` 封装 Wails 绑定函数，页面不直接依赖生成的 binding 文件。
- `stores` 管理跨页面状态，避免存放后端已经负责的业务规则。
- `types` 保存前端 DTO 或契约类型，需与 Go service/facade 返回结构保持一致。

### Wails Binding Boundary

Wails 生成代码和绑定方法应视为桥接边界。

规则：

- Go App/Facade 方法命名应稳定，按资源或用例分组。
- Binding 方法不要直接访问数据库驱动或执行复杂生成逻辑。
- 前端通过 API Client 调用 binding，避免页面层散落桥接细节。
- 执行进度、日志和长任务状态通过 runtime events 传递。

## Module Boundary Rules

### Domain does not know UI or Wails

领域模型不能依赖 Wails runtime、Vue、前端 DTO 或页面概念。UI DSL 可以驱动 API 和页面实现，但不应污染领域对象。

### Service owns orchestration

服务层负责连接验证、Schema 扫描、规则保存、Project 管理、任务启动、预检和错误转换等用例。不要把这些流程分散到前端页面或 Wails binding。

### Adapter owns external differences

数据库差异、本地存储差异和外部数据源差异都应封装在 adapter/store/source 层。领域服务只依赖稳定接口和能力描述。

### Engine owns execution details

执行计划、依赖排序、行数规划、生成上下文、批次写入和执行结果属于 engine 边界。Project 服务可以启动或查询执行，但不应复制 engine 内部算法。

### Generator is plugin-like

生成器通过接口、定义和注册表扩展。内置生成器、外部数据源生成器、关系生成器和计算字段生成器都应遵循同一契约。

## Naming Conventions

### Specs

Spec 名称使用稳定的 phase 前缀和小边界名称：

```text
phase-01-project-structure
phase-01-config-system
phase-01-local-storage-strategy
phase-01-database-dialect-interface
phase-01-test-tooling
```

后续 spec 继续使用 `phase-XX-meaningful-name`，避免用过大的 `phase-03-generation-engine` 直接承载整个阶段。

### Go

Go 包名应短小、语义明确，优先按职责命名，例如 `domain`、`service`、`store`、`engine`、`generator`、`dbx`、`dialect`、`introspect`。数据库实现包使用数据库名称，例如 `mysql`、`postgres`。

### Frontend

Vue 页面和组件应按页面或功能语义命名。通用组件不要包含具体业务流程名称；业务组件应放在对应页面/workflow 附近。

### Events and Contracts

事件名称使用领域前缀和动作：

```text
execution:task-started
execution:batch-completed
execution:task-completed
```

服务方法和 DTO 名称应按资源和用例命名，避免暴露底层数据库或 Wails 实现细节。

## Import and Dependency Direction

依赖方向应保持单向：

```text
UI -> Frontend API Client -> Wails Binding -> Facade -> Service -> Domain / Store / Adapter / Engine
```

禁止反向依赖：

- domain 不依赖 service、facade、Wails 或 frontend。
- engine 不依赖 UI 页面。
- generator 不直接依赖具体数据库连接。
- adapter 不把数据库特定类型泄漏给业务层，必要时通过原始元数据字段保留。

## Context Control Rules

实现时必须遵守 `docs/agent/README.md` 的上下文预算：

- Phase 5 每个 spec 只实现一个生成器类别。
- Phase 8 每个 spec 只读取一个 UI DSL 文件。
- Phase 7 每个 spec 只实现一个资源或一组紧密相关 API。
- Phase 9 优先针对已有模块补测试，不借测试阶段扩大产品范围。
- 如果依赖缺失，先说明缺口，再做当前 spec 的最小必要实现。

## Files That Should Not Drive Product Structure

Agent 技能、编辑器配置和自动化元数据不应成为应用架构的一部分。它们可以辅助开发流程，但不要把 `.agents/`、编辑器目录或 Kiro 设置目录写入产品模块结构。