# Research: phase-03-row-count-planning

## Discovery Context

本研究记录用于支撑 `phase-03-row-count-planning` 的规格设计。该特性属于 Phase 3 execution engine 的扩展型能力，位于依赖拓扑排序之后、生成上下文和批量生成之前。研究重点是对齐已有 Phase 2 Project / Relation 模型、Phase 3 lifecycle 接缝和 dependency plan 输出合同，同时保持行数规划边界独立。

## Source Inputs Reviewed

- `.kiro/specs/phase-03-row-count-planning/brief.md`
- `.kiro/steering/roadmap.md`
- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`
- `.kiro/specs/phase-03-execution-lifecycle/requirements.md`
- `.kiro/specs/phase-03-execution-lifecycle/design.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/requirements.md`
- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/design.md`
- `.kiro/specs/phase-02-project-model/requirements.md`
- `.kiro/specs/phase-02-project-model/design.md`
- `.kiro/specs/phase-02-relation-model/requirements.md`
- `.kiro/specs/phase-02-relation-model/design.md`

## Existing Contracts and Findings

### Steering Findings

- Go 后端 engine 层拥有执行生命周期、依赖排序、行数规划、生成上下文、批处理和结果模型等业务规则。
- Wails binding 和 Vue 不承载业务规则，行数规划不得从 UI 或 binding 获取判断依据。
- 项目强调本地隐私边界，公开错误不得泄露密码、连接字符串、用户 SQL 或生成数据内容。
- Phase 3 roadmap 将行数规划安排在依赖图 / 拓扑排序之后、生成上下文之前。

### Upstream Lifecycle Findings

- `phase-03-execution-lifecycle` 定义统一预检结果：通过状态、阻断错误和非阻断警告。
- lifecycle 通过下游 planner / generation / writer ports 接入后续能力，但自身不实现依赖图、行数规划、生成或写入算法。
- 公开错误模型限定为 code、stage、fieldPath、safeMessage 等安全摘要字段。
- 行数规划应以 lifecycle-compatible result 的形式返回，并在存在阻断问题时阻止进入生成阶段。

### Upstream Dependency Plan Findings

- `phase-03-dependency-graph-and-topological-sort` 输出稳定 `ExecutionPlan`，其中 `OrderedTables` 按拓扑顺序包含 `ProjectTableID`、`TableID` 和 `ExecutionOrder`。
- dependency plan 已负责依赖图、边来源、循环诊断和拓扑排序；行数规划不得重复实现这些算法。
- 该规格明确将目标行数、倍率推导和不可满足行数场景排除在边界之外，并预期后续 row count planning 消费稳定排序结果。

### Project Model Findings

- `ProjectTable` 承载表级行数配置、清空策略和执行顺序快照。
- `ProjectTable.RowCount` 的空值语义是动态推导；显式 `0` 是合法的零行目标，不等同于缺失。
- Project 模型不根据关系角色决定 `rowCount` 是否必须为空，这属于后续 service / engine 规则。
- `ProjectTableRelation` 包含当前执行关系实例、父子 ProjectTable、倍率配置和值来源。
- `FROM_EXECUTION` 和 `MERGED` 需要当前执行父表；`FROM_DB_QUERY` 可以表达外部来源，但 SQL 不得出现在公开错误中。

### Relation Model Findings

- `TableRelation` 表达逻辑关系类型和倍率范围。
- 关键关系语义包括 Parent/Child 与 BaseTable/JoinTable。
- `multiplierMin == 0` 表示上游行可以没有下游行。
- `multiplierMax == 0` 仅在 `multiplierMin == 0` 时合法，表示固定零下游行。
- Relation 模型只表达约束，不实现行数规划或 JoinTable 容量推导。

## Design Decisions

### Decision 1: 独立 engine rowcount 包

行数规划应新增 `internal/engine/rowcount` 包，避免污染 lifecycle、dependency plan、domain 或 UI 层。该包接收 Project / Relation 快照和 dependency plan 输出，产出 `RowCountPlan` 与安全诊断。

**Rationale**: 与 Phase 3 roadmap 和 dependency direction 一致；便于后续 generation context 和 batch loop 复用行数结果。

### Decision 2: 消费拓扑计划而非重排图

行数规划输入必须包含上游 `ExecutionPlan.OrderedTables`，规划输出顺序与拓扑顺序一致。行数规划不重新构建依赖图、不重新做拓扑排序。

**Rationale**: 防止两个排序算法产生分歧，也保持本规格聚焦目标行数和倍率约束。

### Decision 3: 显式零与动态空值必须分离

`nil` 行数表示需要推导或默认化，显式 `0` 表示用户要求生成零行。所有模型和测试必须保留该差异。

**Rationale**: 这是 Phase 2 Project 模型的核心语义，也是零值关系诊断的前提。

### Decision 4: 约束求解采用确定性范围收敛

行数规划应先加载显式表级目标，再按拓扑顺序和关系约束收敛目标范围。动态节点只有在存在确定、唯一、可解释的结果时才输出目标行数，否则返回阻断诊断或明确默认来源。

**Rationale**: 避免隐藏随机选择和后续生成阶段不确定性；符合“生成前暴露不可满足约束”的目标。

### Decision 5: 安全错误与 lifecycle 字段方向兼容

行数错误使用独立的 `RowCountIssue`，字段方向与 lifecycle / plan 安全错误一致：code、stage、fieldPath、safeMessage、blocking。后续集成可映射到 lifecycle precheck without raw payload.

**Rationale**: 保持包独立，减少 lifecycle 破坏性耦合，同时满足预检聚合。

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| 动态行数推导规则过于隐式 | 生成工作量不可预测 | 只在确定可解释时推导，其他情况阻断或显式默认来源 |
| 与依赖图排序职责重叠 | 算法重复且结果不一致 | 输入强制使用 ExecutionPlan，边界测试禁止排序算法扩散 |
| SQL 或连接信息进入错误消息 | 泄露敏感信息 | 使用固定安全消息和敏感词测试，不透传原始来源 |
| JoinTable 语义过度扩展 | 提前实现复杂容量估算 | 本规格只处理关系倍率和目标行数，不做字段唯一性容量估算 |
| 后续 generation context 需要额外字段 | 接口破坏 | 输出保留目标行数、来源和顺序，后续通过非破坏性字段扩展 |

## Synthesis Outcome

本规格应建立最小但完整的行数规划边界：

1. `RowCountInput` 接收 Project 表、Project 关系、Schema 关系和 dependency `ExecutionPlan`。
2. `RowCountPlanner` 输出按拓扑顺序稳定排列的 `RowCountPlan`。
3. `ConstraintEvaluator` 负责 Parent/Child 和 BaseTable/JoinTable 倍率范围校验与动态推导。
4. `Diagnostics` 负责字段级、安全、lifecycle-compatible 问题表达。
5. 测试覆盖显式零、动态空值、缺失配置、非法倍率、冲突范围、不可满足零值和未来能力隔离。
