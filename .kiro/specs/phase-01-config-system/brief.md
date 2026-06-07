# Brief: phase-01-config-system

## Problem

LoomiDBX 需要在桌面端保存和加载应用运行配置，包括语言、主题、本地数据目录、开发环境选项以及后续账号/LLM 等配置入口。没有配置系统会导致各模块各自读取环境或硬编码默认值。

## Current State

项目尚无源码级配置实现。产品文档提到系统配置，Phase 1 要求建立配置管理、环境变量和本地数据目录约定，但不应在此阶段实现完整设置 UI 或远端账号能力。

## Desired Outcome

完成后，应用具备统一配置模型、默认值、配置文件路径约定、加载/保存接口和开发环境覆盖机制。后续本地存储、设置页和服务层可以复用同一配置入口。

## Approach

在 Go 后端建立配置包，定义 AppConfig、默认配置、配置加载顺序和保存接口。配置系统优先服务本地应用启动和基础设置，不处理业务数据，不直接承担 SQLite schema 管理。

## Scope

- **In**:
  - 定义应用配置模型和默认值。
  - 定义配置文件路径和数据目录发现策略。
  - 支持开发/测试环境覆盖。
  - 提供加载、校验、保存接口。
  - 为 Vue 设置页和 Go service 预留读取接口。
- **Out**:
  - 不实现完整设置页面。
  - 不实现账号登录注册流程。
  - 不保存数据库连接密码的最终加密方案，除非只定义接口边界。
  - 不设计所有业务表或本地 SQLite schema。

## Boundary Candidates

- 配置系统负责应用级设置和路径发现。
- 本地存储策略负责 SQLite、业务数据、迁移和 Repository。
- 敏感信息处理可在此定义接口，但最终存储策略应与 local-storage-strategy 协调。

## Out of Boundary

- UI 表单和设置页面交互。
- 远端账号服务。
- 数据库 Schema 缓存、Project、执行历史等业务数据持久化。

## Upstream / Downstream

- **Upstream**: `phase-01-project-structure`、`docs/agent/01-architecture-bootstrap.md`、`docs/product_outline.md`、`docs/api-contract.md`。
- **Downstream**: `phase-01-local-storage-strategy`、`settings-page`、服务层初始化、测试配置。

## Existing Spec Touchpoints

- **Extends**: 无。
- **Adjacent**: `phase-01-project-structure`、`phase-01-local-storage-strategy`、`phase-01-test-tooling`。

## Constraints

配置系统必须适合桌面应用，不依赖远端服务启动。涉及用户数据库连接、Schema、生成配置和生成数据的内容不得上传远端。实现时保持与 Wails 启动生命周期兼容。
