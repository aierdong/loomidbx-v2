# Phase 3：数据生成执行引擎

## 阶段目标

实现从 Project 配置到实际生成与写入的核心执行流程，包括执行生命周期、依赖排序、行数规划、生成上下文、批处理写入和基础执行状态。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [engine-1-architecture.md](../engine-1-architecture.md) | 总体架构与生命周期 | 必读 |
| [engine-2-topology.md](../engine-2-topology.md) | 拓扑排序与行数预计算 | 必读 |
| [engine-3-execution.md](../engine-3-execution.md) | 生成与写入流程 | 必读 |
| [data-model.md](../data-model.md) | Project、表、字段、关系模型 | 只读执行相关模型 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [engine-4-observability.md](../engine-4-observability.md) | 需要实现执行历史、日志、进度、错误报告时 |
| [database-dialect-abstraction-design.md](../database-dialect-abstraction-design.md) | 需要接入真实数据库写入时 |
| [generator-extensibility-design.md](../generator-extensibility-design.md) | 需要调用生成器契约时 |

## 本阶段核心任务

- 定义执行生命周期。
- 从 Project 构建执行计划。
- 按外键/依赖关系进行拓扑排序。
- 计算各表生成行数。
- 构建生成上下文。
- 调用生成器产出行数据。
- 批量写入目标数据库或写入适配层。
- 记录最小执行状态和错误信息。

## 非目标

- 不实现所有内置生成器。
- 不实现 UI 进度界面。
- 不实现 AI 生成。
- 不实现复杂可观测性，除非 spec 属于 Phase 9 或明确要求。

## Spec-Kit 建议拆分

- execution-lifecycle
- dependency-graph-and-topological-sort
- row-count-planning
- generation-context
- batch-generation-loop
- batch-writer-adapter
- execution-result-and-error-model

## Context Budget

执行引擎阶段重点阅读 `engine-1/2/3`。`engine-4` 只在实现历史、日志、进度、追踪时阅读。
