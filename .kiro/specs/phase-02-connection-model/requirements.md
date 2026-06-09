# Requirements Document

## Introduction

`phase-02-connection-model` 为 LoomiDBX 建立数据库连接配置的领域表达。目标用户包括需要保存和复用数据库连接的市场/售前、数据库开发测试人员和软件开发人员。当前 Phase 1 已完成配置系统、本地存储策略和数据库方言接口边界，但还没有连接业务模型，导致后续 Schema 扫描、Project 组织和连接 API 缺少稳定输入。

本规格要求系统能够表达连接名称、数据库类型、网络地址、初始数据库、用户名、敏感凭据引用和扩展参数，并对这些信息进行基础校验和安全边界划分。

## Boundary Context

- **In scope**: 连接领域模型、数据库类型枚举、连接参数和值对象、敏感字段边界、基础校验、序列化/反序列化期望。
- **Out of scope**: 连接管理 UI、Wails API、连接 CRUD 服务、真实数据库连接验证、具体密码加密实现、Schema 扫描。
- **Adjacent expectations**: 本规格依赖 Phase 1 的本地存储策略和数据库方言接口；下游 Schema 模型和 API 规格应复用本规格定义的连接身份、数据库类型和敏感信息边界。

## Requirements

### Requirement 1: 连接基础信息表达

**Objective:** As a 数据库开发或测试人员, I want 系统稳定表达可复用的数据库连接配置, so that 后续扫描 Schema、组织 Project 和执行生成任务可以引用同一连接定义。

#### Acceptance Criteria

1. When 用户创建或加载连接配置时, the 系统 shall 表达连接名称、数据库类型、主机、端口、初始数据库、用户名和扩展参数。
2. When 连接配置被用于下游流程时, the 系统 shall 提供稳定的连接标识用于引用该连接，而不是依赖连接名称作为唯一身份。
3. If 连接名称为空或只包含空白字符, then the 系统 shall 将该连接配置判定为无效。
4. If 数据库类型不属于系统支持或预留的数据库类型集合, then the 系统 shall 将该连接配置判定为无效。
5. The 系统 shall 区分连接业务字段与应用全局配置字段，避免把连接模型混入普通应用配置。

### Requirement 2: 数据库类型与方言能力对齐

**Objective:** As a 开发人员, I want 连接模型使用统一的数据库类型枚举, so that 后续数据库适配、Schema 扫描和 UI 选择项可以保持一致。

#### Acceptance Criteria

1. When 连接配置声明数据库类型时, the 系统 shall 使用稳定的数据库类型枚举表达 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、ClickHouse、TiDB 和 Hive。
2. When 下游组件读取数据库类型时, the 系统 shall 提供可序列化的字符串值用于持久化和前后端传输。
3. If 数据库类型暂未实现真实适配器, then the 系统 shall 仍可表达该类型的配置边界，但不得暗示真实连接能力已经可用。
4. The 系统 shall 保持数据库类型表达与 Phase 1 数据库方言接口的扩展方向一致。

### Requirement 3: 敏感凭据边界

**Objective:** As a 用户, I want 数据库密码和令牌不会作为普通明文字段保存, so that 本地隐私和凭据安全边界得到保护。

#### Acceptance Criteria

1. When 连接配置包含密码、令牌或其他凭据时, the 系统 shall 使用敏感凭据引用或加密后结果表达这些值，而不是把明文作为普通业务字段保存。
2. When 连接配置被序列化为普通业务数据时, the 系统 shall 排除明文密码、明文令牌和其他明文敏感参数。
3. If 扩展参数包含被识别为敏感的键名, then the 系统 shall 将该参数纳入敏感字段边界处理。
4. The 系统 shall 允许下游服务识别哪些字段属于敏感凭据边界，以便交给本地安全存储策略处理。

### Requirement 4: 扩展连接参数表达

**Objective:** As a 高级用户, I want 保存数据库连接的可扩展参数, so that 不同数据库的 SSL、超时、字符集和连接选项可以在不破坏核心模型的情况下表达。

#### Acceptance Criteria

1. When 用户需要配置数据库特定连接选项时, the 系统 shall 支持以键值参数表达非敏感扩展选项。
2. When 扩展参数被保存或加载时, the 系统 shall 保持参数键、值和基础类型的可序列化表达。
3. If 扩展参数键名为空或只包含空白字符, then the 系统 shall 将该扩展参数判定为无效。
4. If 扩展参数包含复杂结构, then the 系统 shall 通过 JSON 边界表达该结构，并保留反序列化校验能力。
5. The 系统 shall 不把扩展参数解释为数据库驱动专属行为，真实连接行为由后续适配器或服务规格负责。

### Requirement 5: 基础校验与错误表达

**Objective:** As a 开发人员, I want 连接模型提供明确的基础校验结果, so that 服务层和 UI 可以复用同一组错误原因。

#### Acceptance Criteria

1. When 连接配置缺少必填字段时, the 系统 shall 返回包含字段名称和错误原因的校验结果。
2. If 主机为空且数据库类型需要网络地址, then the 系统 shall 将该连接配置判定为无效。
3. If 端口超出有效 TCP 端口范围, then the 系统 shall 将该连接配置判定为无效。
4. When 多个字段同时无效时, the 系统 shall 返回所有可检测的基础校验问题，而不是只返回第一个问题。
5. The 系统 shall 不在校验错误中泄露明文密码、令牌或敏感扩展参数值。

### Requirement 6: 序列化兼容性

**Objective:** As a 开发人员, I want 连接模型具备稳定的序列化和反序列化行为, so that 本地 SQLite、配置文件和后续 Wails 契约可以安全复用。

#### Acceptance Criteria

1. When 连接配置被序列化时, the 系统 shall 生成稳定字段名和可持久化格式。
2. When 已保存连接配置被反序列化时, the 系统 shall 恢复数据库类型、连接基础信息、敏感凭据引用和扩展参数。
3. If 反序列化输入缺少可选字段, then the 系统 shall 使用安全默认值而不是产生未定义状态。
4. If 反序列化输入包含未知字段, then the 系统 shall 不因未知字段破坏已知字段的恢复能力。
5. The 系统 shall 通过测试覆盖连接配置的序列化、反序列化、基础校验和敏感字段排除行为。
