# Brief: phase-03-generation-context

## Problem

批量生成需要稳定的执行期上下文来读取 Project、Schema、字段规则、关系引用和已生成键值。没有上下文边界时，生成循环可能直接依赖存储层、Project 服务或具体数据库对象，导致 engine 难以测试和复用。

## Current State

Phase 2 已定义 Schema、字段、约束、关系、字段生成规则和 Project 配置等持久化模型。Phase 3 前序 spec 将提供生命周期、执行顺序和行数计划，但尚未定义执行期间供生成器和关系填充使用的上下文快照。

## Desired Outcome

完成后，engine 具备 GenerationContext：包含执行任务快照、表级计划、字段规则快照、关系引用、已生成主键/唯一键缓存和生成器调用所需的只读输入，为批量生成循环提供一致的数据访问边界。

## Approach

把执行期上下文设计为从 Phase 2 模型派生的只读快照和受控运行态集合。上下文不直接访问 UI、Wails binding 或本地存储；生成循环通过上下文读取规则和记录已生成引用，后续 generator framework 通过最小接口接入。

## Scope

- **In**:
  - 执行任务、Project、Schema、字段规则和 row count plan 的上下文快照。
  - 表级、字段级和关系级查找接口。
  - 已生成主键、外键候选值和关系引用的运行态缓存边界。
  - 供生成器调用的最小输入结构。
- **Out**:
  - 不实现完整生成器注册表或内置生成器。
  - 不实现跨执行任务的全局缓存。
  - 不把生成数据内容持久化到本地历史中。

## Boundary Candidates

- Generation context snapshot：负责执行期只读配置。
- Runtime reference store：负责保存当前执行中的键值和关系引用。
- Generator call input：负责隔离 Phase 4 生成器契约的接入点。

## Out of Boundary

- 真实数据库读取已有父表数据的高级策略。
- 外部数据源生成器上下文。
- UI 预览状态和表单草稿。

## Upstream / Downstream

- **Upstream**: phase-03-row-count-planning、phase-02-field-generation-rule-model、phase-02-table-field-constraint-model、phase-02-relation-model。
- **Downstream**: phase-03-batch-generation-loop、phase-04-generator-interface、phase-05-relation-generators。

## Existing Spec Touchpoints

- **Extends**: 无；该 spec 将 Phase 2 领域模型转化为 engine 执行期视图。
- **Adjacent**: phase-02-field-generation-rule-model 提供字段规则表达；Phase 4 后续定义完整生成器接口。

## Constraints

GenerationContext 不应直接依赖 Wails、Vue 或具体数据库驱动。上下文中的错误和日志不得泄露敏感连接信息或大批量生成数据内容。
