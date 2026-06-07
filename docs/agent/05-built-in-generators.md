# Phase 5：内置生成器实现

## 阶段目标

在 Phase 4 的生成器框架之上，按类型逐步实现 LoomiDBX 首发需要的内置生成器，并为每类生成器补充参数校验和契约测试。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [generators_manual.md](../generators_manual.md) | 具体生成器定义 | 只读当前 spec 要实现的生成器类型 |
| [generators_check_rules.md](../generators_check_rules.md) | 生成器校验规则 | 只读当前生成器相关规则 |
| [generator-extensibility-design.md](../generator-extensibility-design.md) | 生成器契约 | 只读接口、注册、扩展约定 |
| [data-model.md](../data-model.md) | 字段类型与规则配置 | 只读字段规则相关部分 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [computed-field-research.md](../computed-field-research.md) | 实现计算字段生成器时 |
| [engine-3-execution.md](../engine-3-execution.md) | 需要确认生成器运行上下文时 |

## 本阶段核心任务

- 按字符串、数字、布尔、日期时间、枚举、ID、关系等类型拆分实现。
- 为每类生成器定义参数 Schema、默认值和校验规则。
- 确保生成结果满足字段类型、nullable、unique、范围等基本约束。
- 为每个生成器补充单元测试和契约测试。
- 将生成器注册到内置注册表。

## 非目标

- 不在一个 spec 中实现全部生成器。
- 不实现外部 HTTP/CSV/数据库取数生成器。
- 不实现 UI 配置组件。
- 不实现 AI 生成能力。

## Spec-Kit 建议拆分

- string-generators
- number-generators
- boolean-generators
- datetime-generators
- enum-generators
- id-generators
- relation-generators
- computed-field-generators

## Context Budget

每次实现只阅读一个生成器类别相关内容。不要让 Agent 同时处理整个 `generators_manual.md` 的全部清单。
