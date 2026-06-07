# Brief: phase-01-local-storage-strategy

## Problem

LoomiDBX 需要保存连接信息、Schema 扫描缓存、字段规则、Project 配置、执行历史和系统设置。若缺少统一本地存储策略，后续模块会各自定义文件或表结构，导致迁移、隐私和测试困难。

## Current State

文档建议本地应用服务主要读写本地 SQLite 和配置文件，但项目尚未建立本地数据目录、SQLite 初始化、迁移、Repository 边界和敏感信息处理策略。

## Desired Outcome

完成后，项目具备本地存储的总体策略：哪些数据进入配置文件，哪些进入 SQLite，数据目录如何确定，迁移如何组织，Repository 如何落位，敏感信息如何隔离或加密。后续领域模型和服务层可以在此基础上实现具体表结构。

## Approach

以配置系统提供的数据目录为基础，定义本地 SQLite 与配置文件的分工。建立 storage/repository/migration 的目录与接口约定，先提供最小初始化和 mock/测试能力，不在本 spec 中实现所有业务表。

## Scope

- **In**:
  - 定义本地数据目录和文件布局。
  - 定义配置文件与 SQLite 的职责分工。
  - 建立 SQLite 初始化、连接管理和迁移目录约定。
  - 定义 Repository 接口落位和测试替身策略。
  - 明确敏感连接信息的存储边界和后续加密接口。
- **Out**:
  - 不实现完整业务数据模型表。
  - 不实现所有 Repository。
  - 不实现完整账号、授权或云同步。
  - 不实现目标数据库写入。

## Boundary Candidates

- 配置文件适合保存轻量应用设置和路径配置。
- SQLite 适合保存连接元数据、Schema 缓存、字段规则、Project 和执行历史等结构化本地业务数据。
- 敏感信息需要与普通业务数据分离处理，至少预留 secret store/encryption 接口。

## Out of Boundary

- 具体 connection-model、project-model、generation-job-model 的完整字段定义。
- UI 设置页。
- 远端账号服务和遥测上报。

## Upstream / Downstream

- **Upstream**: `phase-01-project-structure`、`phase-01-config-system`、`docs/api-contract.md`。
- **Downstream**: `connection-model`、`project-model`、`generation-job-model`、`connection-api`、`execution-history-api`。

## Existing Spec Touchpoints

- **Extends**: 无。
- **Adjacent**: `phase-01-config-system`、`phase-01-database-dialect-interface`。

## Constraints

必须维护本地隐私边界：数据库连接信息、Schema、生成规则、Project 配置、生成数据和 SQL 不上传远端。首批只定义策略和最小基础设施，不提前实现所有业务表。
