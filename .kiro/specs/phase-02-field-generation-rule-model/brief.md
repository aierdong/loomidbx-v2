# Brief: phase-02-field-generation-rule-model

## Problem
用户需要为字段保存生成规则。规则必须归属 Schema 层并可被多个 Project 复用，同时需要承接未来生成器注册表、参数校验和结构变更状态提示。

## Current State
`docs/data-model.md` 定义 `GeneratorConfig` 与 `DbColumn` 一对一绑定，包含 `generator_name`、`data_mapping_type`、`params` 和 `config_status`。Phase 4 以后才会实现完整生成器契约和参数校验。

## Desired Outcome
完成后，系统有字段级生成规则的领域模型，能够表达绑定字段、生成器标识、输出逻辑类型、参数 JSON 和配置状态，并清楚声明与未来生成器框架的接口边界。

## Approach
先定义稳定的规则存储形态和基础验证：字段绑定唯一、生成器名称非空、输出类型枚举受控、参数保留为 JSON/RawMessage 边界、状态支持 ACTIVE 与 NEEDS_REVIEW。复杂参数 schema 校验留到 Phase 4。

## Scope
- **In**: `GeneratorConfig` 模型，`data_mapping_type` 枚举，`config_status` 枚举，参数 JSON 表达，字段一对一绑定规则，基础验证、序列化测试。
- **Out**: 具体生成器实现、生成器注册表、参数 schema 校验、预览生成、字段智能推荐、API/UI 表单。

## Boundary Candidates
- 字段规则归属 Schema 层，不归属 Project 层。
- 参数结构由未来生成器定义，本 spec 只保存和传递。
- 结构重扫后状态标记机制可以定义字段，但不实现完整 diff 流程。

## Out of Boundary
- 不实现 Phase 4 生成器接口。
- 不实现 Phase 5 内置生成器。
- 不生成样本预览数据。

## Upstream / Downstream
- **Upstream**: `phase-02-table-field-constraint-model`、`phase-02-relation-model`。
- **Downstream**: `phase-02-project-model`、Phase 4 generator-registry-and-contract、Phase 7 field-rule-api、Phase 8 field-rule-editor。

## Existing Spec Touchpoints
- **Extends**: 无。
- **Adjacent**: Phase 4 generator-interface、generator-parameter-validation。

## Constraints
字段规则与表执行规则必须分离。模型不能假设具体生成器参数结构，也不能提前引入未来生成器实现依赖。
