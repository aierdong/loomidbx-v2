# Phase 1：项目骨架与基础架构

## 阶段目标

建立项目工程骨架、模块边界、配置体系、基础持久化策略、测试工具链和数据库方言抽象的最小可用基础。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [product_outline.md](../product_outline.md) | 理解产品范围 | 只读产品定位、当前范围、关键规则 |
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 数据库兼容边界 | 重点读抽象层目标、连接、Schema introspection、写入差异 |
| [api-contract.md](../api-contract.md) | 前后端/服务边界 | 只读总体资源模型和基础约定 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [data-model.md](../data-model.md) | 需要确定初始实体目录或存储模型时 |
| [settings_login.md](../settings_login.md) | 需要实现登录/设置相关基础模块时 |

## 本阶段核心任务

- 明确技术栈与运行形态。
- 建立源码目录结构与模块边界。
- 建立配置管理、环境变量和本地数据目录约定。
- 建立测试框架、lint/format/build 命令。
- 定义数据库连接与方言抽象的最小接口。
- 准备后续领域模型、引擎和生成器模块的落位目录。

## 非目标

- 不实现完整数据库扫描。
- 不实现生成执行引擎。
- 不实现所有数据库方言。
- 不实现具体生成器。
- 不实现完整 UI 页面。

## Spec-Kit 建议拆分

- project-structure
- config-system
- local-storage-strategy
- database-dialect-interface
- test-tooling

## Context Budget

本阶段主要处理工程基础设施。除非 spec 明确要求，不要阅读 `ui-dsl/*`、生成器清单、执行引擎细节和 AI 方案。
