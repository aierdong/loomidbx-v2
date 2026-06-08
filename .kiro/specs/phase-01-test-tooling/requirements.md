# Requirements Document

## Introduction
LoomiDBX 同时包含 Go 后端、Vue 3 前端和 Wails 桌面集成。开发者和后续 spec 实现者需要一组稳定、轻量、可本地运行的验证入口，避免每个功能各自定义测试、格式化、lint 和构建方式。当前 `phase-01-project-structure` 已建立工程骨架和基础命令，但完整测试工具链、前后端样例测试策略、桌面集成验证边界和开发文档仍需要由本 spec 固化。完成后，仓库应具备最小可用的验证命令集合，并让后续 spec 能明确引用这些命令作为验收依据。

## Boundary Context
- **In scope**: 定义并固化后端单元测试、前端类型检查或样例测试、格式化、lint、桌面构建或检查、验证文档和最小样例测试。
- **Out of scope**: 不补齐业务模块完整测试覆盖，不实现生成器契约测试全集，不建立完整 E2E UI 自动化体系，不引入重型 CI/CD、覆盖率平台或可观测性平台。
- **Adjacent expectations**: 本 spec 依赖 `phase-01-project-structure` 提供的工程骨架和命令入口；后续业务 spec 应引用本 spec 固化的验证路径，并在自己的边界内补充业务测试。

## Requirements

### Requirement 1: 统一验证命令集合
**Objective:** As a 开发者, I want 项目提供统一的测试、格式化、lint 和构建验证入口, so that 每次改动后都能按一致方式获得最小质量反馈。

#### Acceptance Criteria
1. When 开发者查看项目命令清单, the LoomiDBX repository shall list backend test, frontend validation, format, lint, desktop build or check, and aggregate verification entry points.
2. When 开发者运行聚合验证入口, the LoomiDBX repository shall execute the current minimum backend, frontend, lint or build validations in a deterministic order.
3. If 某个验证入口依赖缺失的本地工具, then the LoomiDBX repository shall report the missing prerequisite with an actionable message.
4. The LoomiDBX repository shall keep validation commands local and shall not require real database credentials, schema data, generated data, project configuration, user SQL, or remote account data.

### Requirement 2: 后端测试与静态检查
**Objective:** As a 后端开发者, I want Go 后端具备可运行的单元测试和静态检查入口, so that 后端基础能力能在没有业务模块的情况下被验证。

#### Acceptance Criteria
1. When 后端开发者运行后端测试入口, the LoomiDBX repository shall execute all current Go package tests without requiring a target database or remote service.
2. When 后端开发者运行后端格式化入口, the LoomiDBX repository shall format or verify all tracked Go source files in the current project scope.
3. When 后端开发者运行后端静态检查入口, the LoomiDBX repository shall verify current Go packages and report actionable diagnostics.
4. Where backend sample tests are included, the LoomiDBX repository shall cover at least one deterministic backend behavior from the application skeleton.

### Requirement 3: 前端类型检查、格式化和样例验证
**Objective:** As a 前端开发者, I want Vue 前端具备类型检查、格式化和最小样例验证入口, so that 前端基础代码能在业务页面完成前保持可维护。

#### Acceptance Criteria
1. When 前端开发者运行前端验证入口, the LoomiDBX repository shall verify TypeScript and Vue source without relying on completed business workflows.
2. When 前端开发者运行前端格式化入口, the LoomiDBX repository shall verify or apply the selected formatting rules for frontend source files.
3. If 前端验证失败, then the LoomiDBX repository shall surface diagnostics that identify the failing source or configuration.
4. Where frontend sample validation is included, the LoomiDBX repository shall cover at least one deterministic frontend boundary such as API result typing, bootstrap client behavior, or component rendering readiness.

### Requirement 4: Wails 桌面集成验证
**Objective:** As a 桌面应用开发者, I want Wails 集成具备轻量构建或检查入口, so that 前后端桥接和桌面壳不会在早期失去可验证性。

#### Acceptance Criteria
1. When 开发者运行桌面检查入口, the LoomiDBX repository shall verify required local desktop tooling or report missing prerequisites.
2. When 开发者运行桌面构建入口, the LoomiDBX repository shall run the current frontend build and desktop build or a documented local fallback.
3. If Wails CLI or平台依赖不可用, then the LoomiDBX repository shall fail with a clear diagnostic rather than an unexplained command error.
4. The LoomiDBX repository shall keep desktop validation limited to skeleton integration and shall not require full business workflows or real database operations.

### Requirement 5: 开发文档与后续 spec 可引用性
**Objective:** As a 规格实现者, I want 验证命令和适用边界被清楚记录, so that 后续 spec 能复用同一套最小验收路径。

#### Acceptance Criteria
1. When 后续 spec 需要声明验收命令, the LoomiDBX repository shall provide documented command names and expected validation scope.
2. When 某个命令只覆盖当前骨架或工具链范围, the LoomiDBX repository shall label that limitation instead of implying full business coverage.
3. The LoomiDBX repository shall identify which checks are mandatory for ordinary code changes and which checks are optional or environment-dependent.
4. If 后续 spec 需要更深的业务测试, then the LoomiDBX repository shall direct that work to the owning feature spec or later testing phase.

### Requirement 6: 范围保护和轻量工具链
**Objective:** As a 项目维护者, I want 测试工具链保持轻量且边界明确, so that Phase 1 不会提前引入与项目规模不匹配的复杂测试平台。

#### Acceptance Criteria
1. The LoomiDBX repository shall prefer built-in or already selected project tooling for the initial validation pipeline.
2. If a new testing or linting dependency is introduced, then the LoomiDBX repository shall document why it is needed for the current scope.
3. The LoomiDBX repository shall not introduce full E2E automation, coverage gates, remote CI services, or observability platforms as part of this feature.
4. While business modules are still incomplete, the LoomiDBX repository shall keep sample tests focused on deterministic skeleton or contract behavior.
