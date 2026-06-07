# Requirements Document

## Introduction
LoomiDBX 需要从 greenfield 状态启动开发。开发者需要一个清晰的 Wails + Go + Vue3 工程骨架，避免后续领域模型、数据库适配、生成引擎、服务层和 UI 页面混杂在一起。当前仓库主要包含规划文档与 agent 指南，尚未建立实际应用源码结构；`docs/agent/01-architecture-bootstrap.md` 明确要求先建立项目骨架、模块边界、测试工具链和数据库方言抽象的落位目录。完成后，仓库应具备可运行或可构建的基础应用结构，明确后端、前端、桌面桥接、领域层、服务层、适配器层、配置与测试目录的职责边界，使后续 spec 能在既定目录中增量添加功能。

## Boundary Context
- **In scope**: 建立基础工程骨架、源码目录、模块落位约定、前后端桥接边界、基础命令入口和最小可验证样例。
- **Out of scope**: 不实现完整配置系统、本地存储策略、真实数据库连接、Schema 扫描、生成执行引擎、具体业务页面、完整业务 API 或具体数据库方言能力。
- **Adjacent expectations**: `phase-01-config-system`、`phase-01-local-storage-strategy`、`phase-01-database-dialect-interface` 和 `phase-01-test-tooling` 应能在本 spec 建立的目录与命令约定上继续扩展；本 spec 只提供落位边界和最小样例，不替代相邻 spec 的功能实现。

## Requirements

### Requirement 1: 基础应用骨架
**Objective:** As a 开发者, I want 仓库具备可识别、可启动或可构建的基础桌面应用骨架, so that 后续功能可以在稳定工程基础上增量实现。

#### Acceptance Criteria
1. When 开发者查看仓库根目录, the LoomiDBX repository shall present a clear application structure for the desktop shell, backend source, frontend source, generated bindings, documentation, and tests.
2. When 开发者执行项目约定的启动或构建入口, the LoomiDBX repository shall provide a deterministic result that either completes successfully or reports the missing prerequisite with an actionable message.
3. If 当前环境缺少必要开发工具, then the LoomiDBX repository shall document the missing prerequisite rather than failing with an unexplained command error.
4. The LoomiDBX repository shall include a minimal application entry point that proves the backend and frontend parts belong to one application skeleton.

### Requirement 2: 后端模块落位边界
**Objective:** As a 后端开发者, I want 后端源码具备明确的模块落位边界, so that 后续领域模型、服务、存储和数据库适配能力不会混杂在同一位置。

#### Acceptance Criteria
1. When 后端开发者新增领域模型、服务、仓储、适配器、配置或内部基础设施代码, the LoomiDBX repository shall provide named locations that indicate where each kind of code should be placed.
2. When 后端开发者查看桥接入口, the LoomiDBX repository shall make clear that bridge-facing methods are entry points to application capabilities rather than the primary place for complex business rules.
3. Where future database compatibility work is added, the LoomiDBX repository shall provide a reserved location for adapter, dialect, introspection, type mapping, and capability-related code.
4. If a backend module is reserved for a later spec, then the LoomiDBX repository shall mark it as a placeholder or documented location without claiming that the business capability is already implemented.

### Requirement 3: 前端模块与样式落位边界
**Objective:** As a 前端开发者, I want 前端源码具备页面、组件、状态、路由、API client、类型和样式的落位约定, so that 后续 UI 工作流可以按一致结构扩展。

#### Acceptance Criteria
1. When 前端开发者查看前端源码, the LoomiDBX repository shall expose distinct locations for pages, reusable components, state stores, routing, API client wrappers, shared types, and styles.
2. When 前端开发者新增页面或交互入口, the LoomiDBX repository shall make clear whether the change belongs in page-level code, reusable components, state management, or API client code.
3. Where styling infrastructure is present, the LoomiDBX repository shall provide a clear place for global styles, utility classes, theme tokens, and UI library integration notes.
4. If a UI page is outside this spec, then the LoomiDBX repository shall not present a completed business workflow for that page as part of the project structure work.

### Requirement 4: 前后端调用边界
**Objective:** As a 全栈开发者, I want 前端调用后端能力时有一致的封装边界, so that 页面代码不会直接散落依赖底层桥接细节。

#### Acceptance Criteria
1. When 前端需要调用本地应用能力, the LoomiDBX repository shall provide a designated API client location that can wrap generated bridge functions.
2. When 后端暴露本地应用能力给前端, the LoomiDBX repository shall provide a designated facade or app entry location for stable frontend-facing methods.
3. The LoomiDBX repository shall document that transport-specific bindings are not the business logic boundary.
4. If no real business API is implemented in this spec, then the LoomiDBX repository shall limit any callable example to a minimal health, greeting, or bootstrap-style capability.

### Requirement 5: 后续模块扩展落位
**Objective:** As a 规格实现者, I want 后续 Phase 和相邻 spec 有明确落位, so that 配置、本地存储、数据库方言、领域模型、生成引擎和生成器可以独立实现并验收。

#### Acceptance Criteria
1. When 后续 spec 需要新增配置能力, the LoomiDBX repository shall identify where configuration-related code and documentation should be added without implementing the complete configuration system in this spec.
2. When 后续 spec 需要新增本地存储能力, the LoomiDBX repository shall identify where local persistence-related code and migrations can be added without defining the complete storage schema in this spec.
3. When 后续 spec 需要新增数据库方言能力, the LoomiDBX repository shall identify where database compatibility code belongs without implementing real database-specific behavior in this spec.
4. When 后续 spec 需要新增生成引擎或生成器能力, the LoomiDBX repository shall identify their reserved locations without exposing them as completed user-facing features.

### Requirement 6: 基础命令与开发者验证
**Objective:** As a 开发者, I want 项目提供基础开发命令和验证入口, so that 我可以确认工程骨架处于可继续开发的状态。

#### Acceptance Criteria
1. When 开发者查看项目说明或命令清单, the LoomiDBX repository shall list the supported setup, development, build, format, lint, and test entry points that exist at this stage.
2. When 开发者运行基础验证命令, the LoomiDBX repository shall verify the project skeleton without requiring completed business modules.
3. If a validation command is intentionally deferred to `phase-01-test-tooling`, then the LoomiDBX repository shall label that command as deferred or placeholder rather than reporting it as fully available.
4. The LoomiDBX repository shall keep validation expectations limited to the current project-structure scope.

### Requirement 7: 范围保护与隐私边界
**Objective:** As a 项目维护者, I want 工程骨架阶段明确非目标和隐私边界, so that 早期基础设施不会误导后续实现或引入不必要的数据暴露。

#### Acceptance Criteria
1. The LoomiDBX repository shall state that database connection details, schema metadata, generated data, project configuration, and user SQL are local product data and are not uploaded by project-structure scaffolding.
2. If a sample, placeholder, or bootstrap capability is included, then the LoomiDBX repository shall not require real database credentials, real schema data, or remote account data to validate the skeleton.
3. When later specs add business capabilities, the LoomiDBX repository shall provide enough boundary documentation for implementers to avoid treating scaffolding placeholders as completed product behavior.
4. The LoomiDBX repository shall exclude AI generation, full business workflows, and complete database operations from the project-structure scope.
