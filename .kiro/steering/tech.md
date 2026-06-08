# Technology Stack

LoomiDBX 采用 Golang + Vue 3 + Wails 的桌面应用形态。Go 后端承载领域模型、数据库适配、生成引擎、本地存储和服务层；Vue 3 前端承载桌面 UI 与工作流；Wails 负责本地桌面壳、Go 方法绑定、前后端桥接和运行时事件。

本文件记录技术决策和约束。它不列出所有依赖；新增依赖或模块应优先符合这些边界。

## Architecture

### Desktop-first Wails Architecture

项目是本地桌面应用，不默认启动公网或本地 REST 服务。API 契约应理解为传输无关的服务契约，并映射到 Wails 绑定方法和运行时事件。

推荐调用链：

```text
Vue3 Page / Store
  -> Frontend API Client
  -> Wails generated binding
  -> Go App Facade
  -> Go Service
  -> Repository / Adapter / Engine
```

Wails 绑定层只负责参数接收、调用 Facade/Service、返回 DTO 和错误转换。业务规则应保留在 Go service/domain 层。

### Layered Backend

Go 后端应按职责分层：

- **Domain / Model**：连接、Schema、表、字段、约束、关系、规则、Project、GenerationJob 等核心表达。
- **Service / Facade**：承载业务用例、事务边界、校验和前后端契约。
- **Repository / Store**：访问本地 SQLite、配置文件和其他本地持久化。
- **Database Adapter**：访问用户目标数据库，封装连接、扫描和写入差异。
- **Execution Engine**：构建执行计划、排序依赖、生成数据、批量写入和发布状态。
- **Generator Framework**：定义生成器接口、注册表、参数 Schema、校验和内置/外部生成器。

### Frontend Responsibility

Vue 3 前端负责页面渲染、用户交互、局部状态、表单体验、加载/错误/空状态和对服务契约的调用。前端不应实现核心生成算法、数据库规则判断或复杂约束推导。

前端 UI 组件库采用 PrimeVue。当前阶段使用 PrimeVue styled mode，并以官方 Aura preset 起步；主题轻量定制应优先通过 `@primeuix/themes` 的 design token 和 `definePreset` 完成。完全 unstyled mode 保留为后续 UI 深度定制目标，不作为 Phase 1 工程骨架的默认要求。

## Core Technologies

- **Backend Language**：Go。
- **Frontend Framework**：Vue 3。
- **Frontend UI Components**：PrimeVue，当前阶段采用 styled mode + 官方 Aura preset。
- **Desktop Runtime / Bridge**：Wails。
- **Local Persistence**：本地 SQLite 与配置文件组合。
- **Database Integration**：通过 Adapter、Dialect、Introspector、TypeMapper、Capabilities 封装数据库差异。

具体版本和命令以 Phase 1 工程骨架 spec 落地后的配置文件为准。本文件只记录稳定技术方向。

## Database Compatibility Strategy

数据库差异必须下沉到专门组件：

- `Adapter`：每种数据库的统一入口。
- `Dialect`：SQL 方言、占位符、标识符引用、批量插入等差异。
- `Introspector`：Schema 元数据扫描。
- `TypeMapper`：原生类型到逻辑类型的映射。
- `Capabilities`：事务、外键、批量写入、RETURNING、JSON、数组等能力描述。

业务层不应按数据库类型硬编码分支：

```go
// 避免
if dbType == "clickhouse" {
    // special case
}

// 推荐
if caps.SupportsTransaction {
    // use transaction
}
```

统一 Schema 模型应保留数据库原始信息，例如 `NativeType`、逻辑类型和原始元数据，避免抽象层丢失后续排错和兼容能力。

首期优先验证 MySQL 和 PostgreSQL，并为 Oracle、SQL Server、SQLite、ClickHouse、TiDB、Hive 等保留扩展空间。

## API and Service Contract Standards

### Contract before Transport

API 设计先定义资源、请求、响应、错误和事件，再映射到 Wails binding。HTTP 路径可以作为语义参考，但不是必须实现的网络接口。

### Backend Owns Business Rules

复杂业务规则应在 Go 后端统一实现：

- Project 保存时计算或校验执行顺序。
- Schema 重扫后标记受影响的字段规则。
- 执行前预检由服务层和执行引擎统一完成。
- 字段可生成性、约束冲突和生成器参数校验由后端负责最终判断。

### Frontend API Client as Boundary

前端页面不应直接散落调用 Wails 生成函数。应通过前端 API Client 封装调用细节，为后续 Wails、HTTP 或测试 mock 保留替换空间。

### Runtime Events for Long Tasks

长任务状态变化使用 Wails runtime events。事件命名应稳定、语义化，并按领域前缀组织，例如：

```text
execution:task-started
execution:table-started
execution:batch-completed
execution:validation-warning
execution:task-completed
```

## Persistence and Privacy

本地业务数据主要存放在 SQLite 和配置文件中。设计时需要区分：

- 普通应用配置：主题、语言、开发配置、数据目录等。
- 本地业务数据：连接、Schema 缓存、字段规则、Project、执行历史等。
- 敏感信息：数据库密码、令牌、凭据等，应使用更安全的本地存储策略，不能明文随意写入普通配置。

远端账号服务不得接收数据库连接、Schema、生成配置、生成数据和用户 SQL。

## Generator and Engine Standards

生成器通过统一接口和注册表扩展，不应直接访问数据库。执行引擎负责提供上下文、调度生成器、处理依赖关系和写入计划。

执行引擎主路径：

1. 从 Project 构建执行计划。
2. 基于关系和外键做依赖排序。
3. 计算表级行数。
4. 构建生成上下文。
5. 调用生成器产出行数据。
6. 通过写入适配层批量写入目标数据库。
7. 记录状态、错误和历史。

## Development Standards

### Context Control

开发时必须按当前 Phase 和当前 spec 读取最小必要文档。不要因为相关文档存在就默认全部读取，也不要跨阶段实现未要求能力。

### Testing Direction

测试体系应覆盖 Go 后端、Vue 前端和 Wails 集成构建。不同阶段按边界逐步补齐：

- Phase 1 建立格式化、lint、单元测试和构建验证命令。
- 领域模型和数据库抽象优先做单元测试和 mock 测试。
- 生成器需要契约测试。
- 执行引擎需要集成测试。
- API/服务层需要契约测试。
- UI 需要工作流级验证。

### Quality Rules

- 优先修复根因，不用 UI 或 binding 层掩盖后端模型问题。
- 保持服务契约稳定，DTO 与领域模型边界清晰。
- 新依赖必须服务于当前 spec，不为未来假设提前引入重型框架。
- 日志和错误信息应帮助用户和开发者定位问题，但不能泄露敏感数据。