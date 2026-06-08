## Summary
- **Feature**: `phase-01-project-structure`
- **Discovery Scope**: New Feature / Complex Integration
- **Key Findings**:
  - 当前仓库仍处于 greenfield 状态，只有规划文档、spec 与 agent 指南，尚无 `go.mod`、Wails 配置、前端工程或应用源码；本设计需要从零定义工程骨架与边界。
  - Wails 将采用 v3：官方安装文档要求 Go 1.25+、`go install github.com/wailsapp/wails/v3/cmd/wails3@latest` 与 `wails3 doctor` 验证；项目命令必须在缺少 Go、Node/npm 或 `wails3` 时给出可执行提示。
  - PrimeVue 当前阶段采用 styled mode + 官方 Aura preset 起步，通过 `@primeuix/themes` 的 design token 和 `definePreset` 做轻量定制；完全 unstyled mode 保留为后续 UI 深度定制目标。
  - Phase 1 相邻 spec 需要在本骨架上继续扩展；本 spec 只定义配置、本地存储、数据库方言、生成引擎与测试工具链的落位，不实现完整业务能力。

## Research Log

### Greenfield 仓库现状与 Phase 1 边界
- **Context**: 需求要求建立 Wails + Go + Vue3 工程骨架，当前仓库尚未包含实际应用源码。
- **Sources Consulted**:
  - `docs/agent/01-architecture-bootstrap.md`
  - `docs/phase.md`
  - `docs/product_outline.md`
  - 仓库目录检查：根目录仅包含 `.agents/`、`.codex/`、`.kiro/`、`docs/`、`AGENTS.md`
- **Findings**:
  - Phase 1 的职责是工程结构、模块边界、配置体系、基础持久化策略、测试工具链和数据库方言抽象。
  - 本 spec 对应 Phase 1 的 `project-structure`，不应实现完整配置系统、数据库连接、生成执行引擎或业务页面。
  - 产品定位是桌面端数据库模拟/合成数据生成工具，隐私边界要求数据库连接、Schema、生成配置、生成数据和用户 SQL 默认保留在本地。
- **Implications**:
  - 设计必须优先声明边界，避免骨架阶段“顺手”实现真实业务能力。
  - 文件结构需要为相邻 spec 留出稳定落位，但占位文件必须明确声明未实现业务能力。

### Wails v3 工程与开发前提
- **Context**: 用户明确要求 Wails 使用 v3，并提供官方安装文档。
- **Sources Consulted**:
  - [Wails v3 Installation](https://v3.wails.io/quick-start/installation/)
- **Findings**:
  - Wails v3 CLI 命令为 `wails3`。
  - 安装方式为 `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`。
  - 官方文档要求 Go 1.25+，并推荐执行 `wails3 doctor` 验证环境。
  - 前端包管理器可使用 npm、pnpm、yarn 或 bun；Wails 模板通常依赖 Node/npm 类工具。
  - Windows 依赖 WebView2；macOS 依赖 Xcode Command Line Tools；Linux 默认依赖 GTK4 + WebKitGTK 6.0，较旧发行版可能需要 `-tags gtk3`。
- **Implications**:
  - 项目命令入口必须把 `wails3 doctor` 作为环境诊断入口之一。
  - 启动或构建脚本在缺少工具时应给出“安装 Go 1.25+、安装 wails3、运行 wails3 doctor、安装 Node/npm”的提示。
  - 骨架设计应以 Wails App Facade + Vue API client 作为前后端调用边界。

### PrimeVue styled mode 与后续 unstyled 迁移边界
- **Context**: 用户要求关注 PrimeVue unstyled mode，同时补充当前阶段也可以采用 styled mode + 官方 preset 来降低初始工作量。
- **Sources Consulted**:
  - [PrimeVue Styled Mode](https://primevue.org/theming/styled/)
  - [PrimeVue Unstyled Mode](https://primevue.org/theming/unstyled/)
- **Findings**:
  - PrimeVue styled mode 的主题由 base 与 preset 组成；内置 preset 包括 Aura、Material、Lara 和 Nora。
  - 官方推荐通过 design token 定制主题：primitive tokens 定义基础色板，semantic tokens 表达 primary、surface、focus ring 等语义，component tokens 只用于具体组件微调。
  - styled mode 可通过 `app.use(PrimeVue, { theme: { preset: Aura, options: { prefix, darkModeSelector, cssLayer } } })` 配置；基于 Aura 使用 `definePreset` 覆盖 token 可以快速获得可用视觉基础。
  - unstyled mode 仍可作为后续更深度 UI 定制目标：它保留组件能力和可访问性，但不包含默认主题变量与 CSS 规则，需要项目自行承担完整样式实现。
- **Implications**:
  - 本 spec 当前采用 styled mode + Aura preset，减少骨架阶段样式工作量，保证 bootstrap 页面和后续基础组件具备可用视觉表现。
  - `frontend/src/ui/primevue/` 应预留 `preset.ts` / `config.ts` 等位置，用于集中管理 Aura 定制、dark mode selector、cssLayer 选项和未来 unstyled 迁移说明。
  - `frontend/src/styles/` 仍保留 global styles、项目 token 与 utility classes，但当前不承担完整 PrimeVue 组件重写职责。
  - 完全 unstyled mode 不作为本 spec 的当前实现目标；如后续切换，需要重新验证主题配置、组件样式覆盖策略和 UI 工作量。

### API 与桥接边界
- **Context**: 需求要求页面代码不直接散落依赖底层桥接细节。
- **Sources Consulted**:
  - `docs/api-contract.md`
- **Findings**:
  - LoomiDBX 使用 Wails 时推荐先定义传输无关服务契约，再由少量 Wails App Facade 方法绑定给 Vue3 前端。
  - 前端应通过 API client 封装 Wails 生成绑定函数，隐藏 transport 细节。
  - 绑定方法只负责参数接收、调用服务、返回 DTO 与错误转换，复杂业务规则属于服务层。
  - 隐私默认安全：远端服务不接收数据库连接、Schema、生成器配置、Project 配置、生成数据或用户 SQL。
- **Implications**:
  - 后端需要 `AppFacade` 或等价入口承载最小健康检查方法。
  - 前端需要 `api/` 目录封装生成绑定，不允许页面直接依赖 Wails 生成目录。
  - `generated bindings` 目录应被视为传输产物，不是业务逻辑边界。

### 数据库方言抽象落位
- **Context**: 需求要求为未来数据库兼容工作预留 adapter、dialect、introspection、type mapping 和 capability 相关位置。
- **Sources Consulted**:
  - `docs/database-dialect-abstraction-design.md`
- **Findings**:
  - 数据库差异应下沉到 `Adapter`、`Dialect`、`Introspector`、`TypeMapper`、`Capabilities`。
  - 上层业务通过统一 Schema 视图和统一写入计划工作，不直接根据数据库类型散落分支。
  - 初期不实现所有数据库，也不追求抹平所有差异。
- **Implications**:
  - 本 spec 只创建或声明 `internal/dbx/` 相关落位，不实现真实数据库行为。
  - 占位包必须用文档或 placeholder 说明“为后续 spec 保留”。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| Wails Facade + Layered Backend | Vue 页面通过 API client 调用 Wails 生成绑定；Go Facade 调用 service、domain、repository、adapter 层 | 符合 Wails 桌面应用形态；桥接边界清晰；后续业务可增量落位 | 需要严格防止页面绕过 API client、Facade 承载业务逻辑 | 选用 |
| 纯 Wails 绑定直连业务方法 | 每个页面直接调用 Wails 绑定方法 | 初期最快 | 绑定层容易变成业务层，后续难以测试和重构 | 拒绝 |
| 本地 HTTP API + Vue Client | 后端启动 HTTP 服务，前端按 REST 调用 | 与 Web API 习惯一致，可迁移到服务端 | 对桌面端骨架过重，引入端口、生命周期和安全额外复杂度 | 暂不采用 |
| Clean Architecture 完整分层 | 严格 entities/usecases/interface adapters/frameworks 分层 | 边界严谨，可测试性强 | greenfield 骨架阶段可能过度抽象 | 仅吸收依赖方向思想，不引入完整模板化复杂度 |

## Design Decisions

### Decision: 使用 Wails v3 桌面骨架作为唯一运行形态
- **Context**: 需求要求可识别、可启动或可构建的桌面应用骨架，用户明确指定 Wails v3。
- **Alternatives Considered**:
  1. Wails v2 — 文档成熟但与用户指定版本不符。
  2. Wails v3 — 与用户要求一致，使用 `wails3` CLI 和 Go 1.25+。
- **Selected Approach**: 采用 Wails v3，并把 `wails3 doctor`、`wails3 dev`、`wails3 build` 纳入命令与验证文档。
- **Rationale**: 符合用户约束，并能以 Go + Vue3 形成单一桌面应用骨架。
- **Trade-offs**: v3 文档和生态仍可能变化，任务实现时需以官方 CLI 生成结果为准。
- **Follow-up**: 实现阶段必须验证实际 `wails3 init` 产物，并调整文件名或配置名以匹配当前 CLI。

### Decision: 桥接边界采用 Frontend API Client + Go App Facade
- **Context**: 需求 4 要求页面代码不直接散落依赖底层桥接细节。
- **Alternatives Considered**:
  1. 页面直接 import Wails generated bindings。
  2. 在 `frontend/src/api/` 统一封装 generated bindings。
- **Selected Approach**: 前端页面只依赖 `frontend/src/api/`；后端只通过 `AppFacade` 暴露最小 bootstrap/health 方法。
- **Rationale**: 后续业务 API 可以在稳定边界下扩展，不受 Wails 绑定生成路径变化影响。
- **Trade-offs**: 初期多一层薄封装，但能防止页面与传输层强耦合。
- **Follow-up**: 后续 API spec 在同一 client/facade 边界内扩展具体资源方法。

### Decision: PrimeVue 当前采用 styled mode + Aura preset，后续保留 unstyled 迁移
- **Context**: 需求 3 要求前端模块和样式落位清晰；用户补充 styled mode + 官方 preset 能显著降低当前阶段工作量，完全 unstyled 可作为后续目标。
- **Alternatives Considered**:
  1. 直接使用 PrimeVue unstyled mode — 控制力最高，但骨架阶段需要承担大量基础组件样式工作。
  2. 使用 PrimeVue styled mode + Aura preset — 初期工作量低，仍可通过 design token 定制。
  3. 完全依赖默认主题且不设定定制落位 — 最快，但后续主题演进边界不清晰。
- **Selected Approach**: 当前阶段采用 PrimeVue styled mode + Aura preset，并在 `ui/primevue/` 和 `styles/` 中保留主题 preset、design token、global styles、utility classes 与后续 unstyled 迁移说明。
- **Rationale**: 骨架阶段优先交付可用、可维护的工程结构；Aura preset 提供稳定视觉基础，design token 能支持轻量品牌化定制，避免过早投入完整组件样式重写。
- **Trade-offs**: 初期对 PrimeVue styled CSS 有依赖，极致定制能力低于 unstyled mode。
- **Follow-up**: UI 深度定制阶段可评估迁移 unstyled mode 或 Volt/code ownership 模式，并重新验证样式工作量与组件覆盖范围。

### Decision: 数据库、配置、存储、生成引擎只预留目录与文档说明
- **Context**: Phase 1 相邻 spec 会分别实现配置、本地存储、数据库方言和测试工具链。
- **Alternatives Considered**:
  1. 在本 spec 中实现最小数据库连接和配置读写。
  2. 只建立命名目录、placeholder 和 README 边界。
- **Selected Approach**: 本 spec 只预留 `internal/config/`、`internal/storage/`、`internal/dbx/`、`internal/engine/`、`internal/generator/` 等目录，不声明业务能力已完成。
- **Rationale**: 满足落位要求，同时避免与相邻 spec 抢边界。
- **Trade-offs**: 骨架验证能力仅限健康检查与静态结构验证。
- **Follow-up**: 相邻 spec 以这些目录为写入边界继续实现。

## Risks & Mitigations
- Wails v3 CLI 产物结构变化 — 实现阶段以官方 `wails3 init` 当前输出为准，并保持设计中的边界不依赖某个易变生成路径。
- 占位目录被误认为已实现能力 — 每个 reserved 目录使用 README、placeholder 或包注释明确“未实现业务能力”。
- 页面绕过 API client 直接依赖 generated bindings — lint 或 review 规则在任务阶段禁止 `pages/`、`components/` 直接 import 生成绑定目录。
- PrimeVue styled mode 定制边界过浅，后续迁移 unstyled 的成本被低估 — 当前在 `ui/primevue/` 文档中记录 Aura preset 定制范围、token 覆盖点和 unstyled 迁移触发条件，避免把 styled mode 误认为最终设计系统。
- 缺少开发工具导致命令失败不可读 — setup/doctor 脚本先检查 Go、Node/npm、`wails3`，失败时输出安装链接和下一步命令。

## References
- [Wails v3 Installation](https://v3.wails.io/quick-start/installation/) — Wails v3 安装、Go 版本、`wails3 doctor` 与平台依赖。
- [PrimeVue Styled Mode](https://primevue.org/theming/styled/) — PrimeVue styled mode、Aura preset、design token、`definePreset` 与主题选项。
- [PrimeVue Unstyled Mode](https://primevue.org/theming/unstyled/) — 后续完全 unstyled 迁移时的样式责任和 pass-through 配置参考。
- `docs/agent/01-architecture-bootstrap.md` — Phase 1 工程骨架和基础架构边界。
- `docs/api-contract.md` — Wails App Facade、Frontend API Client 与传输无关契约建议。
- `docs/database-dialect-abstraction-design.md` — 未来数据库方言抽象的 Adapter、Dialect、Introspector、TypeMapper、Capabilities 落位。
