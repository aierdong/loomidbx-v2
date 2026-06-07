# Phase 9：测试、可观测性与验收

## 阶段目标

补齐跨模块测试、契约测试、执行历史、错误处理、日志、进度反馈和验收流程，使 LoomiDBX 从功能实现进入可验证、可维护状态。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [engine-4-observability.md](../engine-4-observability.md) | 可观测性、执行历史、错误处理 | 必读 |
| [api-contract.md](../api-contract.md) | API 契约测试 | 只读需要验证的资源 |
| [generators_check_rules.md](../generators_check_rules.md) | 生成器校验测试 | 只读校验规则 |
| [user-stories.md](../user-stories.md) | 端到端验收路径 | 只读关键用户路径 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [engine-1-architecture.md](../engine-1-architecture.md) | 需要验证执行生命周期时 |
| [engine-2-topology.md](../engine-2-topology.md) | 需要验证依赖排序/行数规划时 |
| [engine-3-execution.md](../engine-3-execution.md) | 需要验证生成与写入时 |
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 需要做数据库兼容测试时 |

## 本阶段核心任务

- 建立单元测试、集成测试、契约测试和端到端测试分层。
- 补充生成器契约测试。
- 补充执行引擎集成测试。
- 补充 API 契约测试。
- 实现或完善执行历史、进度、错误报告和日志。
- 建立首发版本验收清单。

## 非目标

- 不重写核心业务模块。
- 不为了测试方便删除业务约束。
- 不扩大产品范围。
- 不引入重型可观测性平台，除非项目确实需要。

## Spec-Kit 建议拆分

- unit-test-strategy
- generator-contract-tests
- engine-integration-tests
- api-contract-tests
- ui-workflow-tests
- execution-history-and-progress
- error-reporting-and-logging
- release-acceptance-checklist

## Context Budget

测试阶段按模块补齐，不要一次性阅读所有专题文档。优先从当前失败或缺失覆盖的模块对应文档开始。
