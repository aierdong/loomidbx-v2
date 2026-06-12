# Brief: phase-03-batch-generation-loop

## Problem

执行引擎需要按照计划逐表、逐批生成行数据，并在关系字段、非空字段、默认值和字段规则之间做最小协调。没有批量生成循环时，生命周期、排序、行数和上下文都无法形成可执行闭环。

## Current State

Phase 3 前序 spec 将定义执行入口、拓扑顺序、行数计划和 GenerationContext。项目尚未实现按批次调用生成器、组装行、填充关系字段和推进表级状态的主循环。

## Desired Outcome

完成后，engine 能够根据执行计划和上下文按表生成批次数据，调用最小生成器接口，组装字段值，优先填充关系字段，并把批次交给写入适配边界。循环应能报告表级/批次级进度和失败范围。

## Approach

实现 engine 侧的 batch generation loop，先以接口和 mock 生成器支撑流程闭环，避免提前实现 Phase 4/5 的完整生成器框架。循环关注调度、行组装、关系值引用和错误传播，写入由后续 writer adapter spec 接管。

## Scope

- **In**:
  - 按拓扑顺序和 row count plan 逐表调度。
  - 按批次生成行数据和字段值。
  - 外键/关系字段优先从 GenerationContext 的引用中取值。
  - 对不可生成字段、生成器失败和关系引用缺失返回明确错误。
  - 将生成批次传递给写入适配接口。
- **Out**:
  - 不实现完整内置生成器集合。
  - 不实现真实数据库批量写入细节。
  - 不实现 UI 进度页面或 API 推送。

## Boundary Candidates

- Batch loop runner：负责表和批次调度。
- Row assembler：负责字段值组装和关系值填充。
- Generator invoker interface：负责隔离后续生成器注册表。

## Out of Boundary

- 高级并发调度和跨表并行生成。
- 外部数据源生成器。
- 写入事务、清空策略和数据库错误适配。

## Upstream / Downstream

- **Upstream**: phase-03-generation-context、phase-03-row-count-planning、phase-03-dependency-graph-and-topological-sort。
- **Downstream**: phase-03-batch-writer-adapter、phase-03-execution-result-and-error-model、phase-04-generator-interface。

## Existing Spec Touchpoints

- **Extends**: 无；该 spec 建立 engine 主执行循环。
- **Adjacent**: phase-02-field-generation-rule-model 定义字段规则；Phase 4/5 后续提供真实生成器实现。

## Constraints

外键字段必须优先遵从关联关系，不能作为普通随机字段处理。遇到无法安全生成的数据时应返回错误或预检/运行诊断，不能静默降级。
