# Phase 4：生成器框架与注册表

## 阶段目标

实现所有生成器共享的基础契约、注册表、定义元数据、参数 Schema、校验挂载和查询能力，为后续内置生成器和外部数据源生成器提供稳定扩展点。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [generator-extensibility-design.md](../generator-extensibility-design.md) | 生成器扩展机制 | 必读 |
| [generators_manual.md](../generators_manual.md) | 生成器清单与定义 | 重点读通用定义模型，不要一次实现全部生成器 |
| [generators_check_rules.md](../generators_check_rules.md) | 校验规则 | 重点读通用校验框架和字段约束相关规则 |
| [data-model.md](../data-model.md) | 字段规则模型 | 只读 Field/Rule/Constraint 相关部分 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [engine-3-execution.md](../engine-3-execution.md) | 需要确认执行引擎如何调用生成器时 |
| [api-contract.md](../api-contract.md) | 需要暴露生成器元数据查询接口时 |

## 本阶段核心任务

- 定义 Generator 接口。
- 定义 GeneratorContext、GeneratorResult、GeneratorDefinition。
- 定义参数 Schema 和默认值表达。
- 实现 GeneratorRegistry。
- 支持按 key/type/category 查询生成器定义。
- 接入基础校验规则。
- 提供注册与重复注册错误处理。
- 提供最小 MockGenerator 或示例生成器用于测试。

## 非目标

- 不实现完整内置生成器清单。
- 不实现外部数据源生成器。
- 不实现 UI 配置表单。
- 不实现 AI 生成器。
- 不让生成器直接访问数据库。

## Spec-Kit 建议拆分

- generator-interface
- generator-definition-schema
- generator-registry
- generator-parameter-validation
- generator-metadata-query
- generator-contract-tests

## Context Budget

本阶段只处理框架和契约。阅读 `generators_manual.md` 时只抽取通用定义和分类，不要把所有具体生成器实现纳入当前 spec。
