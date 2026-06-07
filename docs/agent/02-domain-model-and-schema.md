# Phase 2：核心领域模型与 Schema

## 阶段目标

建立 LoomiDBX 的核心领域模型，包括连接、数据库对象、Schema、表、字段、约束、关系、字段生成规则、Project 和生成任务配置。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [data-model.md](../data-model.md) | 领域模型来源 | 必读，重点读连接、Schema、表、字段、Project、Job 相关模型 |
| [user-stories.md](../user-stories.md) | 行为需求来源 | 只读连接管理、Schema 浏览、字段规则、Project 组织相关故事 |
| [product_outline.md](../product_outline.md) | 产品边界 | 只读功能模块和关键规则 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 需要把数据库 introspection 结果映射到领域模型时 |
| [generators_check_rules.md](../generators_check_rules.md) | 需要定义字段规则校验挂点时 |

## 本阶段核心任务

- 定义核心实体和值对象。
- 定义 Schema introspection 结果的内部表达。
- 定义字段生成规则配置模型。
- 定义 Project、ProjectTable、GenerationJob 等执行配置模型。
- 定义关系、约束、唯一性、非空等基础约束表达。
- 建立模型序列化/反序列化策略。

## 非目标

- 不实现具体生成器逻辑。
- 不实现完整执行引擎。
- 不实现 UI 表单。
- 不实现 API 端点，除非当前 spec 明确要求。

## Spec-Kit 建议拆分

- connection-model
- database-schema-model
- table-field-constraint-model
- relation-model
- field-generation-rule-model
- project-model
- generation-job-model

## Context Budget

本阶段以 `data-model.md` 为中心。不要默认阅读执行引擎和 UI DSL；只有当模型字段需要被下游消费时，才查阅对应专题文档。
