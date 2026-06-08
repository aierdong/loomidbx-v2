# Research & Design Decisions

## Summary
- **Feature**: `phase-01-config-system`
- **Discovery Scope**: Extension
- **Key Findings**:
  - 上游 `phase-01-project-structure` 已创建 Wails + Go + Vue 工程骨架，并把 `internal/config` 明确标注为由本 spec 接管的占位模块。
  - Steering 要求本地桌面优先、Wails facade 只做薄入口、业务规则在 Go service/domain 层，配置系统不能上传用户数据库、Schema、生成配置或用户 SQL。
  - 配置系统需要和后续本地存储策略分工：本 spec 定义应用配置与路径发现，SQLite schema、迁移、Repository 和敏感凭据最终存储由相邻 spec 承担。

## Research Log

### 上游工程骨架与落位
- **Context**: 本 spec 依赖 `phase-01-project-structure`，需要复用现有目录与调用边界。
- **Sources Consulted**: `.kiro/specs/phase-01-project-structure/design.md`、`internal/config/README.md`、`app.go`、`frontend/src/api/README.md`。
- **Findings**:
  - `internal/config` 当前只是占位，说明默认值、环境变量、用户设置和校验由本 spec 负责。
  - `App` facade 目前只暴露 `BootstrapStatus`，复杂规则不应写入 facade。
  - 前端已建立 `frontend/src/api/` 作为封装 generated bindings 的唯一入口样例。
- **Implications**: 配置实现应新增 Go 配置包和服务边界，并可增加薄 facade 方法；前端只需要契约和 API client 落位，不实现完整设置页面。

### 项目隐私与配置范围
- **Context**: 配置系统会接触本地路径、未来账号/LLM 入口和可能的密钥状态，必须避免突破隐私边界。
- **Sources Consulted**: `.kiro/steering/product.md`、`.kiro/steering/tech.md`、`docs/api-contract.md`。
- **Findings**:
  - 数据库连接信息、Schema、字段规则、Project 配置、生成数据和用户 SQL 默认不得上传远端。
  - 普通应用配置、本地业务数据、敏感信息需要明确分层。
  - Settings API 文档建议返回 `apiKeyConfigured` 这类布尔状态，不返回明文密钥。
- **Implications**: 本 spec 只允许保存普通应用配置；敏感输入需要通过接口边界标注或交给后续安全存储，不写入普通配置文件。

### 加载顺序与测试隔离
- **Context**: Phase 1 要求建立配置管理、环境变量和本地数据目录约定。
- **Sources Consulted**: `.kiro/specs/phase-01-config-system/brief.md`、`.kiro/steering/roadmap.md`、`.kiro/steering/structure.md`。
- **Findings**:
  - 默认值、配置文件、环境/开发覆盖是最小可解释的加载顺序。
  - 测试配置需要隔离路径，避免污染真实用户配置。
  - 本地存储策略会依赖配置系统输出的数据目录。
- **Implications**: 设计需要把路径发现、加载来源报告、校验错误和保存行为作为公共契约，而不是散落到后续 SQLite 或 UI 实现中。

## Architecture Pattern Evaluation

| Option | Description | Strengths | Risks / Limitations | Notes |
|--------|-------------|-----------|---------------------|-------|
| 单包配置服务 | 在 `internal/config` 内集中放模型、默认值、路径、加载、保存和校验 | 简单、符合 Phase 1 范围、易测试 | 包内职责需要清晰拆分文件，避免变成杂物箱 | 选用，当前 spec 不需要更重架构 |
| 配置 Repository + Service 分层 | 配置文件读写放 repository，服务层合成配置 | 和后续业务存储风格一致 | 对单个配置文件过重，会提前引入本地存储策略细节 | 暂不选用，后续可按需要拆分 |
| 前端直接管理设置文件 | 前端页面读写本地配置 | UI 实现快速 | 违反后端拥有本地能力和 Wails 边界，难以保护隐私 | 拒绝 |

## Design Decisions

### Decision: 配置包采用文件内分责的轻量服务
- **Context**: 当前只有工程骨架，占位模块已经存在，配置系统需要可实现但不应过度抽象。
- **Alternatives Considered**:
  1. 完整 repository/service/facade 分层。
  2. 单个配置包内提供模型、loader、store、validator 和 service。
- **Selected Approach**: 在 `internal/config` 内定义清晰文件职责；由 `Service` 组合 `PathResolver`、`FileStore`、`Loader` 和 `Validator`，对上提供稳定方法。
- **Rationale**: 满足强类型、测试和后续 facade 接入，同时避免在本地存储策略前提前设计数据库式 repository。
- **Trade-offs**: 包内需要通过文件职责和测试约束维持边界。
- **Follow-up**: 如果后续敏感存储或多配置源变复杂，再把 store 接口迁出到更通用基础设施层。

### Decision: 加载顺序固定为默认值到文件到环境覆盖
- **Context**: 配置需要支持无文件启动、用户持久化和开发/测试覆盖。
- **Alternatives Considered**:
  1. 环境变量优先于配置文件并写回文件。
  2. 默认值、配置文件、环境覆盖合成最终配置，环境覆盖不自动写回。
- **Selected Approach**: 使用默认值作为基础，叠加配置文件，再叠加环境覆盖；保存时只保存用户可变项，不把临时覆盖写回。
- **Rationale**: 行为可解释，适合测试隔离，避免开发变量污染用户配置。
- **Trade-offs**: 调试时需要返回来源信息帮助理解最终值。
- **Follow-up**: 实现阶段需要为覆盖来源提供可测试的 metadata。

### Decision: 普通配置与敏感信息严格分离
- **Context**: 设置契约会涉及未来 LLM API key 或账号入口，但本 spec 不实现最终安全存储。
- **Alternatives Considered**:
  1. 在普通配置文件中保存密钥。
  2. 普通配置只保存非敏感字段和 `configured` 状态，敏感值交给后续安全存储边界。
- **Selected Approach**: 配置模型包含敏感入口的状态和非敏感元数据，不保存明文密钥。
- **Rationale**: 符合产品隐私边界，也和 Settings API 中 `apiKeyConfigured` 的契约一致。
- **Trade-offs**: 后续安全存储实现前，写入密钥只能返回未支持或转交边界错误。
- **Follow-up**: 与 `phase-01-local-storage-strategy` 对齐安全存储接口。

## Risks & Mitigations
- 配置路径与本地存储路径边界混淆 — 在设计和任务中明确本 spec 只输出路径与数据目录，SQLite schema 和迁移归属本地存储策略。
- 环境覆盖污染用户配置 — 明确覆盖不自动写回，并在保存契约中只持久化用户可变项。
- 未来设置页误以为账号/LLM 已完成 — 在模型和契约中使用占位状态与未支持错误，避免展示为完整能力。

## References
- `.kiro/steering/product.md` — 产品范围与隐私边界。
- `.kiro/steering/tech.md` — Wails、Go service、配置与持久化边界。
- `.kiro/steering/structure.md` — 后端分层、前端 API client 和模块落位规则。
- `.kiro/specs/phase-01-project-structure/design.md` — 上游工程骨架与 `internal/config` 占位边界。
- `docs/api-contract.md` — Settings 契约、Wails facade 和前端 API client 建议。

