# Phase 6：外部数据源生成器

## 阶段目标

实现从外部数据源取数的生成器能力，例如静态列表、CSV、HTTP API、外部数据库或其他可复用数据源，并将其纳入统一生成器框架。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [external-data-generators-design.md](../external-data-generators-design.md) | 外部取数生成器设计 | 必读 |
| [generator-extensibility-design.md](../generator-extensibility-design.md) | 生成器扩展机制 | 必读接口与生命周期 |
| [generators_check_rules.md](../generators_check_rules.md) | 参数与数据校验 | 只读外部数据源相关规则 |
| [data-model.md](../data-model.md) | 字段规则配置 | 只读外部源配置表达 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 实现数据库数据源生成器时 |
| [api-contract.md](../api-contract.md) | 需要管理外部数据源配置 API 时 |

## 本阶段核心任务

- 定义外部数据源配置模型。
- 实现外部数据加载、缓存、采样和错误处理策略。
- 将外部数据源生成器注册到统一 Registry。
- 支持预览和校验外部源配置。
- 处理认证、超时、缺失字段、类型不匹配等错误。

## 非目标

- 不实现 AI 生成。
- 不实现完整 UI 管理界面，除非当前 spec 明确要求。
- 不把外部数据源逻辑写入执行引擎核心。

## Spec-Kit 建议拆分

- external-source-config-model
- static-list-source-generator
- csv-source-generator
- http-source-generator
- database-source-generator
- external-source-validation-and-preview

## Context Budget

本阶段以 `external-data-generators-design.md` 为中心。只有实现数据库外部源时才阅读数据库方言抽象文档。
