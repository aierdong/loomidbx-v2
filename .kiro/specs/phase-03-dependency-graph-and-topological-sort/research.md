# Research Document

## Summary

`phase-03-dependency-graph-and-topological-sort` 是 Phase 3 执行引擎中紧接生命周期后的计划层规格。调研确认 Phase 2 已提供表、外键、逻辑关系和 Project 关系快照，但没有 engine 侧可执行依赖图和稳定拓扑排序。本规格应新增纯 Go engine 计划包，输出可被 `phase-03-execution-lifecycle` 的 planner/precheck 接缝消费的计划结果和安全预检问题。

核心决策：以 ProjectTable 为节点，以“上游依赖表先生成”为边方向，物理外键使用 `ReferencedTableID -> TableID`，逻辑关系使用 `ParentTableID -> ChildTableID`，Project 关系实例使用 `ParentProjectTableID -> ChildProjectTableID`。所有错误只暴露阶段、错误码、字段路径和安全消息，不泄露 SQL、连接详情或生成数据。

## Research Log

### Brief and Roadmap Review

- `brief.md` 明确问题：生成数据必须先满足外键和逻辑关系约束，子表不能早于父表生成。
- 期望结果：从 Project 配置和 Schema 关系构建表级依赖图，输出稳定拓扑执行顺序，并在循环依赖、缺失表或不可排序场景中返回明确预检错误。
- Roadmap 将本规格置于 `phase-03-execution-lifecycle` 之后、`phase-03-row-count-planning` 之前，说明本规格只负责计划顺序，不负责行数、上下文、生成循环或写入。

### Upstream Lifecycle Touchpoints

- `phase-03-execution-lifecycle` 拥有生命周期状态机、预检聚合、安全错误、下游 planner/generation/writer ports。
- 生命周期设计明确把依赖图和拓扑排序列为后续规格，并要求 planner hook 只通过阶段成功、阶段失败和安全错误驱动生命周期状态。
- 本规格的计划结果应能作为 lifecycle planner seam 的实现输入/输出，但不应反向修改 lifecycle 内部状态枚举或历史模型枚举。

### Existing Domain Model Review

- `internal/domain/schema.ForeignKey`：`TableID` 是外键所在子表，`ReferencedTableID` 是被引用父表，图边方向应为 `ReferencedTableID -> TableID`。
- `internal/domain/schema.TableRelation`：`ParentTableID` / `ChildTableID` 表达逻辑或物理关系方向，图边方向应为 `ParentTableID -> ChildTableID`。
- `internal/domain/schema.RelationType`：支持 `PARENT_CHILD` 和 `JOIN_TABLE`；JOIN_TABLE 分支仍表达 base/parent 到 join/child 的依赖。
- `internal/domain/project.ProjectTable`：Project 内的可执行表节点；`ExecutionOrder` 是快照，不由 Phase 2 计算。
- `internal/domain/project.ProjectTableRelation`：Project 级关系实例；`ParentProjectTableID` 可为空，表示上游表不在当前 Project。
- `internal/domain/project.RelationValueSource`：`FROM_EXECUTION` 依赖当前执行内父表，`FROM_DB_QUERY` 可来自数据库查询，`MERGED` 同时使用当前执行和数据库查询；本规格不能执行或解析 SQL。

### Steering Review

- Product steering 强调约束正确性优先于展示效果，依赖表应先生成，循环和缺失关系应在预检中被识别。
- Tech steering 要求业务/engine 逻辑不得按数据库类型硬编码分支，数据库差异应由 adapter/dialect/capabilities 抽象进入。
- Structure steering 指定 engine 拥有依赖排序，Wails binding 和 Vue 不承载业务规则。

## Architecture Pattern Evaluation

### Option A: Extend lifecycle package with graph logic

- 优点：接入 planner seam 简单。
- 缺点：生命周期包会吸收计划算法，破坏 Phase 3 分层；后续行数规划、上下文和批量循环更难替换。
- 结论：不采用。

### Option B: Create independent engine planning package

- 优点：计划层与 lifecycle 状态控制解耦；可被 lifecycle planner port 调用；便于后续 row-count planning 复用拓扑结果。
- 缺点：需要定义清晰输入/输出和错误边界。
- 结论：采用。

### Option C: Store computed order directly in Phase 2 domain model

- 优点：可复用 `ProjectTable.ExecutionOrder` 字段。
- 缺点：会把运行时算法推入 Phase 2 领域模型，并可能改变持久化语义。
- 结论：不采用；本规格只输出计划结果，是否持久化由后续服务/API 规格决定。

## Design Decisions

1. **节点以 ProjectTable 为执行边界**
   - Graph node 使用 ProjectTable identity 和 schema TableID 双重引用。
   - 同一 Project 内 TableID 重复或 ProjectTable identity 缺失应作为预检阻断问题。

2. **边方向固定为 parent/dependency -> child/dependent**
   - 物理外键：`ReferencedTableID -> TableID`。
   - Schema 逻辑关系：`ParentTableID -> ChildTableID`。
   - Project 关系实例：`ParentProjectTableID -> ChildProjectTableID`。
   - 该方向直接对应拓扑输出的生成顺序。

3. **Project 关系值来源只影响依赖判断，不触发 SQL 行为**
   - `FROM_EXECUTION`：父 ProjectTable 必须存在，并形成当前执行依赖边。
   - `FROM_DB_QUERY`：可没有父 ProjectTable，当前规格不形成执行内依赖边，但保留外部来源警告/元数据。
   - `MERGED`：如父 ProjectTable 存在则形成依赖边；SQL 只作为外部来源存在，不被执行或解析。

4. **稳定拓扑排序使用确定性队列**
   - 入度为 0 的节点按稳定键排序。
   - 稳定键优先采用现有 `ExecutionOrder`，再使用 ProjectTable ID、TableID 等确定性字段。
   - 相同输入应输出相同顺序。

5. **错误与 lifecycle 安全摘要对齐**
   - 图构建、排序和预检错误输出安全摘要：code、stage、fieldPath、safeMessage。
   - 不输出原始 SQL、连接字符串、密码或生成数据内容。

6. **不引入数据库类型分支**
   - 所有关系输入来自 domain snapshot 或后续 adapter/capability 层整理后的模型。
   - 本规格不按 MySQL/PostgreSQL/SQLite 等名称分支。

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Project 关系与 Schema 关系重复产生重复边 | 可能造成诊断噪声和排序不稳定 | 使用 canonical edge key 去重，同时保留来源摘要 |
| 外部父表缺失被误判为阻断 | `FROM_DB_QUERY` 场景可能被错误阻止 | 依据 RelationValueSource 区分执行内依赖与外部值来源 |
| 循环诊断泄露内部输入内容 | 安全边界被破坏 | 循环错误只输出节点/字段路径和安全消息，不输出 SQL 或连接详情 |
| 排序规则漂移影响历史快照 | 后续规格难以复用 | 在设计和测试中固定稳定键与确定性输出 |
| 计划层提前实现行数规划 | 规格越界 | 边界测试禁止 row count planning、generator registry、batch loop 和 writer 行为 |

## References

- `.kiro/specs/phase-03-dependency-graph-and-topological-sort/brief.md`
- `.kiro/steering/roadmap.md`
- `.kiro/steering/product.md`
- `.kiro/steering/tech.md`
- `.kiro/steering/structure.md`
- `.kiro/specs/phase-03-execution-lifecycle/requirements.md`
- `.kiro/specs/phase-03-execution-lifecycle/design.md`
- `.kiro/specs/phase-02-relation-model/requirements.md`
- `.kiro/specs/phase-02-relation-model/design.md`
- `.kiro/specs/phase-02-project-model/requirements.md`
- `.kiro/specs/phase-02-project-model/design.md`
