# 专题 5-1：数据生成执行引擎设计 — 总体架构与执行生命周期

> 本文档是专题 5（数据生成执行引擎设计）的第一部分，聚焦于架构总览与执行生命周期。  
> 依赖文档：产品大纲、数据模型（专题 2）、生成器设计（专题 4）  
> 关联专题：专题 5-2（拓扑排序与行数预计算）、专题 5-3（生成与写入引擎）、专题 5-4（可观测性与历史）

---

## 0. 文档范围与设计边界

本引擎负责将一个已配置完毕的 Project 实际"跑起来"，完成从读取配置到写入目标数据库的全流程。引擎本身是一个本地运行的后端服务模块，不涉及以下内容：

- Schema 扫描与结构发现（专题 3 / Schema 模块）
- 生成器内部实现细节（专题 4）
- 计算字段的求值逻辑（专题 7）
- AI 生成（专题 8）
- API 契约（专题 9）

引擎的核心职责可以用一句话描述：**按照依赖顺序，逐表生成数据、逐批写入目标库，并将结果完整记录下来。**

---

## 1. 设计原则

在所有具体设计决策之前，先确立以下原则，后续章节的取舍均以此为准：

**约束优先于数据量**：若要保证约束合法（FK 引用、唯一性、非空），宁可减少生成行数，也不能写入会导致数据库报错的数据。这是产品大纲 §7 规则"约束优先于展示效果"的引擎级映射。

**批量生成，边生成边写入**：引擎不把所有行先生成完再批量写入，而是生成一批、写入一批，以控制内存占用并使进度可观测。

**串行表执行（v1）**：表按拓扑顺序串行执行，不做跨表并行，这简化了 FK Pool 的同步问题和错误传播逻辑。

**失败传播策略可配置，但依赖安全优先**：`失败即终止` 是用户可选的执行策略，而不是引擎的强制原则。若某张表写入失败且没有依赖它的子表，则视为局部失败：在用户未启用 `失败即终止` 时，Project 可继续处理后续无依赖风险的表，整体任务最终标记为 `PARTIAL_FAILED`。若失败表存在子表依赖，则必须终止当前 Project 的后续处理，并将受影响的下游依赖表标记为 `SKIPPED`，因为继续执行会产生连锁约束风险；该规则不受用户是否启用 `失败即终止` 影响。已成功写入的数据不回滚（代价过高），用户下次执行前可勾选 `truncate_before`。

**执行顺序预计算**：拓扑排序在用户保存 Project 时执行，结果写入 `ProjectTable.execution_order`，引擎直接按该字段顺序执行，不在运行时重新计算依赖图（数据模型 D-08）。

---

## 2. 组件总览

引擎由 9 个功能内聚的组件构成，各组件职责如下：

```
┌──────────────────────────────────────────────────────────────────────┐
│                          ExecutionEngine (Facade)                    │
│  协调所有组件，是引擎的唯一对外入口                                   │
└────────┬─────────────────────────────────────────────────────────────┘
         │
         ├──► ConfigLoader         从本地 DB 加载 Project/ProjectTable/
         │                         ProjectTableRelation/GeneratorConfig
         │
         ├──► PreflightValidator   执行前验证：配置完整性、行数计算、
         │                         JoinTable 唯一性容量校验（D-10）
         │
         ├──► TableExecutor        单表执行器：生成行、写入、更新 FK Pool
         │       │
         │       ├──► RowBuilder   列级数据生成，处理表达式列依赖
         │       │       └──► GeneratorRegistry  生成器分发
         │       │
         │       ├──► DBWriter     批量 INSERT，捕获写入的 PK 值
         │       │
         │       └──► FKPool       跨表 FK 值池，供子表取样
         │
         ├──► ProgressEmitter      向前端推送进度事件
         │
         └──► HistoryManager       持久化 ExecutionTask / ExecutionTableResult
```

### 各组件职责说明

**ConfigLoader**
从本地数据库（LoomiDBX 的应用存储）加载完整的执行上下文：Project 配置、所有 ProjectTable（按 `execution_order` 排序）、ProjectTableRelation、以及每个参与表的所有列和对应 GeneratorConfig。同时加载 DbTable/DbColumn 的结构元数据，供 DBWriter 构造 INSERT 语句。

**PreflightValidator**
在实际执行之前进行的"飞前检查"。这是唯一允许与目标数据库建立连接并发起查询的非写入阶段。职责包括：
- 校验所有参与列是否都有有效的 GeneratorConfig（`config_status = ACTIVE`）
- 执行 `rel_source_sql` 查询以确定父表的存量记录数
- 根据 D-05(docs/engine-2-topology.md:L197) 矩阵计算每张表的预估行数
- 执行 D-10(docs/engine-2-topology.md:L309) JoinTable 唯一性容量校验
- 输出 `ExecutionPlan`（含每表精确/估算行数、D-10 警告、阻塞性错误）

**TableExecutor**
单表执行的最小单元，负责一张表从"开始"到"完成"的全过程：计算本表实际行数（依赖 FK Pool 的最终状态）、分批生成、分批写入、捕获 PK 更新 FK Pool、更新执行状态。

**RowBuilder**
给定一组列配置，生成一行数据（`Map<column_name, value>`）。内部处理列依赖顺序（表达式列必须在其引用列之后生成）、外键值的从 FK Pool 取样、以及 `null_percentage` 应用。

**GeneratorRegistry**
生成器注册表，将 `generator_name`（如 `distribute_int`、`sql_expression`）映射到对应的生成器实现。每个生成器实现接口：`generate(config: GeneratorConfig, context: RowContext) → value`。

**FKPool**
跨表传递主键值的内存数据结构。父表每完成一批写入，将成功写入的 PK 值追加到 FK Pool 中；子表在生成行时从 FK Pool 中取样。对于大规模父表，FK Pool 支持容量上限配置和抽样降级（见专题 5-3 §3.2）。

**DBWriter**
封装目标数据库的写入操作，屏蔽不同数据库（MySQL/PostgreSQL/Oracle 等）的 INSERT 差异。核心接口：`write_batch(table, rows) → InsertResult`，返回实际写入行数和数据库分配的 PK 值（用于 Auto-Increment 场景）。

**ProgressEmitter**
将执行进度事件推送到前端 UI。桌面应用场景下推荐使用进程内事件总线或 IPC 机制；若前后端分离则使用 SSE（Server-Sent Events）。事件类型详见专题 5-4。

**HistoryManager**
在执行开始时创建 `ExecutionTask` 记录，为每张表创建 `ExecutionTableResult` 记录，随着执行进展持续更新两者的状态字段。

---

## 3. 执行生命周期

一次完整的生成任务由以下 4 个阶段构成：

```
Phase 1: 装配（Assemble）
  ↓
Phase 2: 预检（Preflight）
  ↓
Phase 3: 执行（Execute）
  ↓
Phase 4: 收尾（Finalize）
```

### Phase 1：装配（Assemble）

**触发**：用户在 Project 界面点击"执行"。

**步骤**：
1. `ConfigLoader` 加载完整执行上下文。
2. 按 `ProjectTable.execution_order` 升序排列表列表。
3. `HistoryManager` 创建 `ExecutionTask` 记录（状态 `RUNNING`，`started_at = NOW()`）。
4. `HistoryManager` 为每张表创建 `ExecutionTableResult` 记录（状态 `PENDING`）。

装配阶段不访问目标数据库，不做任何生成操作。

### Phase 2：预检（Preflight）

**触发**：装配完成后自动触发。

**步骤**：
1. **配置完整性校验**：遍历所有参与表的所有列，检查是否每列都有 `config_status = ACTIVE` 的 GeneratorConfig（或该列允许为 NULL / 有默认值，见下方"不可生成字段"规则）。任何阻塞性错误（blocking error）均需在此阶段报告，任务不得进入执行阶段。
2. **rel_source_sql 预查询**：对所有 `parent_project_table_id IS NULL` 的 `ProjectTableRelation` 记录（即父表不在本 Project 中的情况），执行 `rel_source_sql` 统计目标库的存量父记录数，作为行数计算的基础。
3. **行数预计算**：按照 D-05 (./engine-2-topology.md:L309) 矩阵为每张表计算预估行数（详见专题 5-2 §2）。
4. **JoinTable 容量校验（D-10）**：检测 JoinTable 的倍数上限是否超过 BaseTable 的有效记录总数，如有超限则自动修正并记录 Warning。
5. **生成并展示 ExecutionPlan**：将每张表的预估行数、执行顺序、Warning 和错误汇总成一份计划报告，在前端展示给用户。

**两种处理结果**：
- 若存在**阻塞性错误**（配置缺失、rel_source_sql 执行失败等）：更新 `ExecutionTask.status = FAILED`，任务终止，用户需修复配置后重新触发。
- 若无阻塞性错误（可能有 Warning）：向用户展示 ExecutionPlan，等待用户确认后进入执行阶段。

**"不可生成字段"规则**的引擎级映射（产品大纲 §7）：
- 字段只有在 `is_nullable = TRUE` 或 `default_value IS NOT NULL` 时，才允许没有 GeneratorConfig。
- 若字段 `is_nullable = FALSE` 且 `default_value IS NULL` 且无 GeneratorConfig → 阻塞性错误。
- 若字段 `is_nullable = TRUE` 且无 GeneratorConfig → 该列写入 NULL（无需报错，可选 Warning 提示用户）。

### Phase 3：执行（Execute）

**触发**：用户在预检报告中点击"确认执行"。

**主循环（按 execution_order 串行）**：

```
skipped_tables = Set()
should_stop_project = false

for each ProjectTable pt in sorted_tables:

  // 0. 若已因上游失败被标记跳过，则不再执行
  if pt in skipped_tables:
    continue

  // 0.1 若已触发 Project 级终止，则跳过剩余未执行表
  if should_stop_project:
    ExecutionTableResult[pt].status = SKIPPED
    ProgressEmitter.emit(TableSkipped, pt)
    continue

  // 1. 标记开始
  ExecutionTableResult.status = RUNNING
  ProgressEmitter.emit(TableStarted, pt)

  // 2. TRUNCATE（如需）
  if pt.truncate_before:
    DBWriter.truncate(pt.table)
    // 注意：TRUNCATE 顺序问题见专题 5-3 §4.5

  // 3. 计算本表实际行数
  actual_row_count = RowCountResolver.resolve(pt, fk_pool, execution_plan)

  // 4. 执行生成写入
  result = TableExecutor.execute(pt, actual_row_count, fk_pool)

  // 5. 处理结果
  if result.status == SUCCESS:
    ExecutionTableResult.rows_written = result.rows_written
    ExecutionTableResult.status = SUCCESS
    FKPool.add(pt.table_id, result.generated_pks)
    ProgressEmitter.emit(TableCompleted, pt, SUCCESS)

  else (FAILED):
    ExecutionTableResult.rows_written = result.rows_written  // 记录已写入数
    ExecutionTableResult.status = FAILED
    ExecutionTableResult.error_message = result.error
    ProgressEmitter.emit(TableCompleted, pt, FAILED)

    downstream_pts = get_dependents(pt)

    if downstream_pts is not empty:
      // 失败表存在子表依赖：必须终止，避免连锁约束风险
      for each downstream_pt in downstream_pts:
        ExecutionTableResult[downstream_pt].status = SKIPPED
        ProgressEmitter.emit(TableSkipped, downstream_pt)
        skipped_tables.add(downstream_pt)

      should_stop_project = true

    else if execution_options.fail_fast:
      // 用户启用“失败即终止”：即使是局部失败，也终止剩余处理
      should_stop_project = true

    else:
      // 无子表依赖且未启用“失败即终止”：视为局部失败
      // Project 继续执行后续无依赖风险的表，最终汇总为 PARTIAL_FAILED
      continue
```

**关于"失败后是否继续执行"的策略**：

当一张表失败时，引擎按以下优先级处理：

1. **失败表存在子表依赖**：必须终止当前 Project 的后续处理，并将受影响的下游依赖表标记为 `SKIPPED`。这是依赖安全规则，不受用户是否启用 `失败即终止` 影响。
2. **失败表没有子表依赖，且用户启用 `失败即终止`**：立即终止剩余处理。此时失败本身是局部的，但用户选择了全局中止策略。
3. **失败表没有子表依赖，且用户未启用 `失败即终止`**：视为局部失败，记录该表 `FAILED`，继续执行后续无依赖风险的表；最终任务通常汇总为 `PARTIAL_FAILED`。

实现方式：维护 `skipped_tables` 集合和 `should_stop_project` 标志。`skipped_tables` 用于避免执行已受上游失败影响的依赖表；`should_stop_project` 用于表达用户主动选择的失败即终止，或依赖链失败导致的强制终止。

### Phase 4：收尾（Finalize）

**步骤**：
1. 根据所有 `ExecutionTableResult.status` 计算 `ExecutionTask.status`：
   - 全部 `SUCCESS` → `SUCCESS`
   - 存在 `FAILED` 或 `SKIPPED`，但有至少一个 `SUCCESS` → `PARTIAL_FAILED`
   - 全部 `FAILED` 或无任何 `SUCCESS` → `FAILED`
2. 写入 `ExecutionTask.ended_at = NOW()`。
3. `ProgressEmitter.emit(TaskCompleted, task_status)`。

---

## 4. 状态流转图

### ExecutionTask 状态流转

```
          [用户点击执行]
               │
               ▼
           RUNNING
          /         \
    (全部成功)    (部分/全部失败)
        │                │
        ▼                ▼
     SUCCESS      PARTIAL_FAILED / FAILED
```

### ExecutionTableResult 状态流转

```
[装配阶段创建]
      │
      ▼
   PENDING
      │
[该表开始执行]
      │
      ▼
   RUNNING
    /    \
(成功)  (失败)    (前置依赖失败)
  │       │              │
  ▼       ▼              ▼
SUCCESS  FAILED        SKIPPED
```

---

## 5. 关键数据结构

以下是引擎内部流转的核心数据结构（伪代码定义，非具体语言实现）：

```
// 预检阶段的输出
ExecutionPlan {
  task_id: bigint
  tables: TablePlan[]
  warnings: Warning[]
  errors: BlockingError[]
}

TablePlan {
  project_table_id: bigint
  table_name: string
  execution_order: int
  estimated_rows: int          // 精确值或估算值
  estimated_rows_range: [int, int]  // 当行数由随机倍数决定时给出范围
  truncate_before: bool
  d10_warnings: D10Warning[]   // JoinTable 倍数自动修正的警告
}

// FK Pool 存储结构
FKPool {
  // table_id → [pk_value...]
  store: Map<bigint, PkValue[]>

  add(table_id, pks: PkValue[]): void
  sample(table_id, n: int, without_replacement: bool): PkValue[]
  count(table_id): int
}

// 单表执行结果
TableExecutionResult {
  rows_written: int
  generated_pks: PkValue[]   // 用于填充 FK Pool
  status: SUCCESS | FAILED
  error: string | null
}
```

---

## 6. 与数据模型的对应关系

| 引擎概念 | 数据模型实体 | 说明 |
|---------|------------|------|
| ExecutionPlan | 无（运行时对象）| 预检输出，不持久化 |
| 执行主记录 | `ExecutionTask` | 整个任务的状态与时间记录 |
| 单表执行结果 | `ExecutionTableResult` | 含快照字段，历史不依赖表存活 |
| 表执行顺序 | `ProjectTable.execution_order` | 预计算，引擎直接读取 |
| FK 值来源策略 | `ProjectTableRelation.rel_value_source` | 决定子表 FK 值取自何处 |
| 行数决策 | `ProjectTable.row_count` + D-05 矩阵 | NULL 或具体值的语义见专题 5-2 |

---

*下一篇：专题 5-2 — 拓扑排序与行数预计算*
