# Brief: phase-03-dependency-graph-and-topological-sort

## Problem

生成数据必须优先满足外键和逻辑关系约束，因此执行引擎需要先知道哪些表依赖哪些表。没有依赖图和拓扑排序时，子表可能早于父表生成，导致外键值不可用或写入失败。

## Current State

Phase 2 已定义 ForeignKey、TableRelation、ProjectTableRelation 和 Project 执行顺序快照等领域模型，但还没有 engine 侧从 Project 表集合和关系集合构建可执行依赖图的算法。

## Desired Outcome

完成后，engine 能够从 Project 配置和 Schema 关系中构建表级依赖图，输出稳定的拓扑执行顺序，并在循环依赖、缺失表或不可排序场景中返回明确的预检错误。

## Approach

在 engine 计划层定义依赖图节点、边、关系来源和拓扑排序结果。优先支持物理外键和 Phase 2 逻辑关系模型，输出供后续行数规划和批量生成循环使用的表级执行计划。

## Scope

- **In**:
  - 从 ProjectTable、ForeignKey、TableRelation / ProjectTableRelation 构建表级依赖图。
  - 标记依赖来源、父表/子表方向和外部值来源摘要。
  - 输出拓扑排序结果和稳定执行顺序。
  - 识别循环依赖、缺失节点、重复边和不可排序错误。
- **Out**:
  - 不计算行数、不生成行数据、不写入数据库。
  - 不实现复杂循环依赖拆解策略或延迟约束写入。
  - 不实现 UI 上的依赖图可视化。

## Boundary Candidates

- Dependency graph builder：负责从 Project 和关系模型构建图。
- Topological sorter：负责排序和循环检测。
- Plan validation result：负责将不可执行原因交给生命周期预检。

## Out of Boundary

- 行数倍数推导和目标行数计算。
- 字段级生成器选择和参数校验。
- 数据库特定写入策略。

## Upstream / Downstream

- **Upstream**: phase-03-execution-lifecycle、phase-02-relation-model、phase-02-project-model。
- **Downstream**: phase-03-row-count-planning、phase-03-generation-context、phase-03-batch-generation-loop。

## Existing Spec Touchpoints

- **Extends**: 无；该 spec 使用 Phase 2 已有关系模型建立 engine 计划层。
- **Adjacent**: phase-02-relation-model 定义关系语义；phase-02-project-model 定义表级执行配置和关系覆盖。

## Constraints

依赖方向必须符合产品规则：存在外键依赖时，被依赖表先生成。业务层不应按数据库类型硬编码依赖算法，数据库差异只能通过已抽象的关系和能力信息进入 engine。
