# Brief: phase-02-connection-model

## Problem
用户需要保存和复用数据库连接，但当前 Phase 1 只建立了配置、本地存储和数据库方言接口边界，还没有业务层连接模型。后续 Schema 扫描、Project 和 API 都需要稳定的连接领域表达。

## Current State
Phase 1 已提供配置系统、本地存储策略、数据库 Adapter/Dialect/Introspector 接口和 mock 能力。缺口是缺少 `Connection` 领域模型、数据库类型枚举、连接参数序列化、敏感字段边界和基础校验规则。

## Desired Outcome
完成后，Go 后端有可测试的连接领域模型，能够表达连接名称、数据库类型、主机、端口、初始数据库、用户名、加密密码引用和扩展参数，并清楚区分敏感字段与普通业务字段。

## Approach
采用 Phase 2 领域模型优先的方式，在后端 domain/model 边界定义连接实体、值对象、枚举和验证逻辑。模型应能对接 Phase 1 的本地存储/secret 策略和 `internal/dbx` 数据库类型能力，但不实现真实连接 CRUD API。

## Scope
- **In**: `Connection` 领域模型、`db_type` 枚举、连接扩展参数表达、敏感字段边界、基础校验、序列化/反序列化测试。
- **Out**: 连接管理 UI、Wails API、真实数据库连接测试流程、密码加密具体实现重做、Schema 扫描。

## Boundary Candidates
- 连接业务模型与配置系统模型分离。
- 连接模型只引用数据库类型/能力边界，不直接依赖具体数据库驱动。
- 密码或凭据字段只表达安全存储引用或加密结果，不暴露明文。

## Out of Boundary
- 不实现连接列表、保存、删除和验证服务。
- 不新增真实数据库驱动。
- 不设计远端账号或授权模型。

## Upstream / Downstream
- **Upstream**: `phase-01-project-structure`、`phase-01-config-system`、`phase-01-local-storage-strategy`、`phase-01-database-dialect-interface`。
- **Downstream**: `phase-02-database-schema-model`、Phase 7 connection-api、Schema introspection API、Project 模型。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 1 local-storage-strategy、database-dialect-interface。

## Constraints
必须遵守本地隐私边界；数据库密码、令牌和敏感参数不能作为普通明文字段持久化。Go 导出 struct、字段、常量和枚举值必须按项目规则添加注释。
