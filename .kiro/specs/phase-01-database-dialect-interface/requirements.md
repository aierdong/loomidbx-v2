# Requirements Document

## Introduction

LoomiDBX 需要在 Go 后端连接和操作多种目标数据库。当前工程骨架已经为 `internal/dbx/` 预留了 Adapter、Dialect、Introspector、TypeMapper 和 Capabilities 的占位目录，但还没有代码级接口、统一值对象或测试替身。不同数据库在连接参数、元数据扫描、类型系统、SQL 方言、事务、外键和批量写入能力上存在差异；如果后续业务层直接按数据库类型写分支，会削弱扩展性并增加维护成本。完成后，后续服务层、Schema 模型、写入计划和测试可以依赖稳定的数据库方言抽象边界，而不需要先实现完整 MySQL 或 PostgreSQL 驱动。

## Boundary Context

- **In scope**: 定义数据库 Adapter、Dialect、Introspector、TypeMapper、Capabilities、连接配置、连接测试、Schema 扫描结果、类型映射、SQL 方言原语和 mock/fake 测试替身的最小接口边界。
- **Out of scope**: 不实现完整 MySQL/PostgreSQL introspection，不连接真实数据库，不执行真实批量写入，不实现执行引擎，不实现 UI 连接页面，不覆盖所有高级数据库能力差异。
- **Adjacent expectations**: `phase-01-project-structure` 已提供 `internal/dbx/` 占位目录；后续 `database-schema-model`、`schema-introspection-api`、`batch-writer-adapter` 和 `engine-integration-tests` 应基于本 spec 的稳定接口继续扩展。

## Requirements

### Requirement 1: 数据库适配统一入口
**Objective:** As a 后端开发者, I want 一个统一的数据库适配入口, so that 后续服务层可以按能力使用目标数据库而不是按数据库类型散落分支。

#### Acceptance Criteria
1. When 后端代码需要识别某种目标数据库, the LoomiDBX DBX module shall expose a stable database type identifier and adapter metadata.
2. When 后端代码需要取得数据库能力、方言、扫描器或类型映射器, the LoomiDBX DBX module shall provide a single adapter-facing entry point for those contracts.
3. When 后端代码需要验证连接配置, the LoomiDBX DBX module shall define a connection test contract that reports success or actionable failure without requiring business services to know driver details.
4. If a requested database type is unsupported, then the LoomiDBX DBX module shall return a typed unsupported-database error rather than silently choosing a fallback adapter.

### Requirement 2: 能力模型与运行时协商
**Objective:** As a 业务服务实现者, I want 数据库差异通过能力模型表达, so that 后续生成、扫描和写入流程可以基于能力做决策。

#### Acceptance Criteria
1. The LoomiDBX DBX module shall define capabilities for transaction, savepoint, foreign key, deferred constraint, batch insert, bulk load, returning, upsert, catalog, schema, JSON, array, UUID, enum, generated column, identity column, and identifier length support.
2. When 后续服务需要选择事务、外键或写入策略, the LoomiDBX DBX module shall make those choices derivable from capabilities instead of database-type conditionals.
3. If a database cannot guarantee a capability, then the LoomiDBX DBX module shall allow the adapter to declare the capability as unavailable or constrained.
4. Where MySQL and PostgreSQL priority is documented, the LoomiDBX DBX module shall keep their expected capability differences visible without requiring complete driver implementation.

### Requirement 3: Schema 扫描结果边界
**Objective:** As a Schema 功能实现者, I want 统一的 Schema 扫描输出边界, so that 后续 Schema 管理、规则配置和写入计划可以复用同一结构。

#### Acceptance Criteria
1. When an introspector scans metadata, the LoomiDBX DBX module shall describe database, namespace, table, column, primary key, foreign key, unique constraint, check constraint, index, and view information in a canonical schema result.
2. The LoomiDBX DBX module shall preserve native type, logical type, nullable, default value, ordinal position, generated or identity markers, comments, and raw metadata for scanned columns.
3. If a database exposes metadata that is not normalized by the current model, then the LoomiDBX DBX module shall retain raw metadata for future diagnostics or adapter-specific enhancement.
4. While canonical schema support is limited to the first-phase interface scope, the LoomiDBX DBX module shall not claim complete coverage for advanced database features such as partition tables, materialized views, custom types, spatial types, or expression indexes.

### Requirement 4: 类型映射边界
**Objective:** As a 生成器和 Schema 规则实现者, I want 原生数据库类型映射到统一逻辑类型, so that 后续字段生成规则可以基于稳定类型语义工作。

#### Acceptance Criteria
1. When a native database type is mapped, the LoomiDBX DBX module shall produce a logical type that preserves kind, length, precision, scale, bit width, timezone, element type, enum values, and native type where applicable.
2. If a native type cannot be confidently recognized, then the LoomiDBX DBX module shall return an unknown or explicitly configured fallback logical type while preserving the original native type.
3. Where database-specific mapping options are needed, the LoomiDBX DBX module shall provide an options boundary without forcing business services to parse native type strings themselves.
4. The LoomiDBX DBX module shall distinguish type mapping from metadata scanning, so mappers can be tested without opening a database connection.

### Requirement 5: SQL 方言原语
**Objective:** As a 写入计划实现者, I want SQL 方言差异有最小稳定原语, so that 后续批量写入和执行引擎可以生成数据库匹配的语句。

#### Acceptance Criteria
1. When SQL identifiers are rendered, the LoomiDBX DBX module shall provide a dialect contract for quoting identifiers consistently.
2. When parameterized SQL is rendered, the LoomiDBX DBX module shall provide a dialect contract for placeholders by argument index.
3. When insert statements are planned, the LoomiDBX DBX module shall define a batch insert request and statement result contract that separates SQL text from arguments.
4. If a dialect cannot build a requested SQL form, then the LoomiDBX DBX module shall return a typed dialect error that identifies the unsupported operation.
5. The LoomiDBX DBX module shall keep real execution, transaction orchestration, COPY, LOAD DATA, upsert automation, and write result persistence outside this feature boundary.

### Requirement 6: 测试替身与注册机制
**Objective:** As a 测试和服务层实现者, I want mock/fake adapter 支持, so that 后续服务可以在没有真实数据库的情况下验证能力协商和调用路径。

#### Acceptance Criteria
1. When unit tests need a database adapter, the LoomiDBX DBX module shall provide fake or mock implementations for adapter, dialect, introspector, type mapper, and capabilities.
2. When tests register adapters, the LoomiDBX DBX module shall provide an in-memory registry that returns the requested adapter or a typed missing-adapter error.
3. If fake introspection data is supplied, then the LoomiDBX DBX module shall return deterministic canonical schema results suitable for repeatable tests.
4. The LoomiDBX DBX module shall allow tests to assert calls and configured failures without opening network connections or requiring real credentials.

### Requirement 7: 范围保护、隐私与依赖对齐
**Objective:** As a 项目维护者, I want 数据库方言接口阶段明确范围和隐私边界, so that 基础接口不会误导后续实现或泄露本地产品数据。

#### Acceptance Criteria
1. The LoomiDBX DBX module shall not require real database credentials, real schema metadata, generated data, project configuration, user SQL, or remote account data to validate this feature.
2. If sample adapters or fixtures are included, then the LoomiDBX DBX module shall mark them as fake or test-only rather than production database support.
3. When this feature extends the project structure skeleton, the LoomiDBX DBX module shall use the existing `internal/dbx/` responsibility boundary rather than introducing a competing backend layout.
4. The LoomiDBX DBX module shall document that complete database drivers, schema APIs, writer adapters, execution engine behavior, and UI workflows belong to downstream specs.
