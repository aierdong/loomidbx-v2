# Research & Design Decisions

## Summary
- **Feature**: `phase-01-local-storage-strategy`
- **Discovery Scope**: Extension
- **Key Findings**:
  - 上游工程骨架已经创建 `internal/storage` 与 `internal/repository` 占位，本规格应接管这些目录并保持薄基础设施边界。
  - 配置系统规格已经把本地数据目录发现作为上游契约，本地存储不得再定义第二套路径来源。
  - 产品和技术 steering 明确本地隐私边界：连接信息、Schema、规则、Project、生成数据和用户 SQL 默认不上传远端。

## Research Log

### 上游规格与目录落位
- **Context**: 本规格依赖 `phase-01-project-structure` 与 `phase-01-config-system`。
- **Sources Consulted**: `.kiro/specs/phase-01-project-structure/design.md`、`.kiro/specs/phase-01-config-system/design.md`、`internal/storage/README.md`、`internal/repository/README.md`。
- **Findings**:
  - 工程骨架将本地持久化和仓储接口分别放在 `internal/storage` 与 `internal/repository`。
  - 配置系统提供 `paths.dataDir` 和开发/测试隔离能力。
  - Wails facade 应保持薄入口，业务规则留在 Go service/domain 层。
- **Implications**: 本设计以 `ConfigService` 的数据目录为唯一输入，新增 storage 初始化、迁移、诊断和 repository 支撑，不改写配置系统职责。

### 本地数据分类与隐私边界
- **Context**: brief 要求明确配置文件、SQLite 和敏感信息的分工。
- **Sources Consulted**: `.kiro/steering/product.md`、`.kiro/steering/tech.md`、`.kiro/steering/roadmap.md`。
- **Findings**:
  - 普通配置适合保存主题、语言、路径、开发模式等轻量设置。
  - 连接元数据、Schema 缓存、字段规则、Project 和执行历史属于本地结构化业务数据。
  - 数据库密码、token 和密钥不得进入普通配置文件或普通业务表明文。
- **Implications**: 设计需要引入 `SecretStore` 接口边界和凭据引用模型，但不实现平台安全存储或加密算法。

### 迁移与 Repository 策略
- **Context**: 后续领域模型需要在本地存储上增量添加表和仓储。
- **Sources Consulted**: `.kiro/steering/structure.md`、上游设计中的分层边界。
- **Findings**:
  - 服务层应依赖 repository 接口，而不是直接依赖 SQLite 连接或文件路径。
  - Phase 1 不应提前实现完整业务表，否则会越界进入领域模型规格。
  - 迁移机制必须先建立记录表和排序规则，后续规格才能安全追加迁移。
- **Implications**: 当前规格只提供存储内核、迁移 runner、最小元数据迁移和 repository 测试替身模式。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 配置文件承载全部本地数据 | 所有设置和业务数据都写入 JSON 文件 | 初始简单 | 并发、查询、迁移和结构演进困难；敏感信息边界容易混乱 | 拒绝 |
| SQLite 承载全部本地数据 | 配置、业务数据和状态全部进入单一数据库 | 查询统一 | 轻量设置不易人工诊断；启动路径与配置路径相互缠绕 | 拒绝 |
| 配置文件 + SQLite + SecretStore 边界 | 普通配置走配置系统，结构化业务数据走 SQLite，敏感信息走独立接口 | 符合 steering，便于迁移和测试 | 需要清晰边界文档和接口约束 | 采纳 |

## Design Decisions

### Decision: 数据目录只来自配置系统
- **Context**: 配置系统已经定义数据目录发现和环境覆盖。
- **Alternatives Considered**:
  1. 本地存储自行解析环境变量和默认路径。
  2. 本地存储只接受配置系统已解析的数据目录。
- **Selected Approach**: 本地存储初始化接收 `ResolvedStorageConfig` 或等价配置视图，其中数据目录由配置系统提供。
- **Rationale**: 避免路径规则分叉，保证开发/测试隔离行为一致。
- **Trade-offs**: 本地存储启动依赖配置系统先完成加载。
- **Follow-up**: 实现时测试配置系统输出的测试目录不会落入真实用户目录。

### Decision: 当前只落地最小迁移基础设施
- **Context**: 后续连接、Project 和执行历史模型会定义具体业务表。
- **Alternatives Considered**:
  1. 在本规格中预建所有业务表。
  2. 只建立迁移记录和基础数据库参数，业务表由后续规格追加。
- **Selected Approach**: 当前规格包含迁移目录、迁移记录表和 runner，不包含完整业务 schema。
- **Rationale**: 避免越过 Phase 1 边界，同时给下游规格稳定扩展点。
- **Trade-offs**: 初始化后的数据库只有基础结构，不能直接支持业务查询。
- **Follow-up**: 后续领域规格添加迁移时必须复用命名和记录规则。

### Decision: 敏感信息以接口和引用表达
- **Context**: 产品要求数据库凭据不明文进入普通配置或普通业务表。
- **Alternatives Considered**:
  1. 当前直接实现加密文件。
  2. 当前只保存明文，后续再改。
  3. 当前定义 `SecretStore` 接口和凭据引用，不实现具体安全存储。
- **Selected Approach**: 采纳接口和引用策略。
- **Rationale**: 不伪装安全能力，也不留下明文迁移负担。
- **Trade-offs**: 连接验证在后续规格中需要处理 secret store 不可用状态。
- **Follow-up**: 后续连接规格应定义凭据写入、读取和删除的完整流程。

## Risks & Mitigations
- 路径规则与配置系统重复 — 通过接口只接收已解析数据目录，禁止 storage 自行推导默认用户目录。
- 迁移越界创建业务表 — 当前只允许基础元数据迁移，业务表必须由后续规格声明。
- mock repository 与真实存储行为漂移 — 在设计中要求测试替身遵守相同接口和错误语义。
- 敏感字段进入日志 — 错误类型只携带字段名、错误码和脱敏原因，不携带原值。

## References
- `.kiro/steering/product.md` — 产品隐私边界和首发范围。
- `.kiro/steering/tech.md` — 本地 SQLite / 配置文件组合、分层后端和 Wails 调用链。
- `.kiro/specs/phase-01-config-system/design.md` — 数据目录发现和配置职责边界。
- `.kiro/specs/phase-01-project-structure/design.md` — `internal/storage`、`internal/repository` 和 facade/API client 落位。
