# Brief: phase-03-row-count-planning

## Problem

Project 中的表级行数配置和表间数量关系需要在执行前转化为每张表的目标生成行数。没有行数规划时，执行引擎无法判断关系倍数是否可满足，也无法为批量生成循环提供明确的工作量。

## Current State

Phase 2 已定义 ProjectTable 的行数状态机、表级配置和关系实例化配置，但尚未实现 engine 侧的行数计算、关系倍数约束检查或目标行数计划。

## Desired Outcome

完成后，engine 能够基于拓扑顺序、Project 表配置和关系倍数计算每张表的目标行数，输出可执行的 RowCountPlan，并对缺失配置、冲突倍数、负数/零值边界和不可满足关系给出预检错误。

## Approach

在执行计划阶段引入独立的 row count planner。它接收依赖图排序结果和 Project 配置，先处理显式表级目标，再按关系倍数约束推导或校验相关表数量，最终输出表级目标行数和约束诊断。

## Scope

- **In**:
  - 表级目标行数读取、默认值处理和边界校验。
  - 基于 Parent/Child、BaseTable/JoinTable 等关系语义检查数量约束。
  - 输出每张表的目标行数、来源和诊断信息。
  - 在预检阶段暴露不可满足的行数关系。
- **Out**:
  - 不实现真实数据生成或数据库写入。
  - 不实现性能压测级的大规模分片计划。
  - 不实现 UI 表单的行数配置交互。

## Boundary Candidates

- Row count planner：负责计算和校验目标行数。
- Row count constraint evaluator：负责关系倍数和边界规则。
- Row count diagnostics：负责把不可满足原因交给生命周期预检。

## Out of Boundary

- 执行批次大小、事务大小和数据库写入吞吐优化。
- 字段级唯一性容量估算。
- 复杂统计分布或真实业务比例建模。

## Upstream / Downstream

- **Upstream**: phase-03-dependency-graph-and-topological-sort、phase-02-project-model、phase-02-relation-model。
- **Downstream**: phase-03-generation-context、phase-03-batch-generation-loop、phase-03-execution-result-and-error-model。

## Existing Spec Touchpoints

- **Extends**: 无；该 spec 使用 Project 模型中已有的行数和关系配置。
- **Adjacent**: phase-02-project-model 定义配置表达；phase-03-dependency-graph-and-topological-sort 提供执行顺序。

## Constraints

行数规划必须在生成前完成，并优先暴露不可满足约束，而不是在写入失败后才报错。该 spec 不应引入未来 UI 或生成器框架的实现细节。
