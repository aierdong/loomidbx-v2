# Phase 7：API 与服务层

## 阶段目标

实现 LoomiDBX 的资源 API、服务层编排和前后端契约，包括连接、Schema、字段规则、Project、生成任务、生成器元数据、执行历史等能力。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [api-contract.md](../api-contract.md) | API 契约来源 | 必读 |
| [data-model.md](../data-model.md) | 请求/响应模型来源 | 只读对应资源模型 |
| [user-stories.md](../user-stories.md) | 行为需求 | 只读当前 API 对应故事 |
| [engine-4-observability.md](../engine-4-observability.md) | 执行历史与状态 | 实现任务/历史 API 时阅读 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [generator-extensibility-design.md](../generator-extensibility-design.md) | 实现生成器元数据 API 时 |
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 实现连接验证/Schema 扫描 API 时 |

## 本阶段核心任务

- 实现资源服务层。
- 实现 API 请求/响应 DTO。
- 实现连接验证、Schema 扫描、字段规则保存、Project 管理、生成任务启动与查询。
- 实现生成器元数据查询 API。
- 实现错误码、校验错误和业务错误响应。
- 衔接领域模型、生成器框架和执行引擎。

## 非目标

- 不在 API 层实现业务核心算法。
- 不在 API 层直接写复杂生成器逻辑。
- 不实现 UI 页面。
- 不扩大首发 API 范围到 AI 能力，除非后续明确纳入。

## Spec-Kit 建议拆分

- connection-api
- schema-introspection-api
- field-rule-api
- project-api
- generator-metadata-api
- generation-job-api
- execution-history-api
- error-response-contract

## Context Budget

API 阶段按资源拆分 spec。每个 spec 只阅读对应资源的模型和契约，不要一次实现整个 API 面。
