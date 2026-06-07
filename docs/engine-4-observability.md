# 专题 5-4：数据生成执行引擎设计 — 可观测性、执行历史与补充关键问题

> 本文档是专题 5 的第四部分，聚焦于运行期进度监控、历史记录管理，以及若干无法归入前三篇的关键设计问题。  
> 依赖文档：数据模型（专题 2）§9 执行历史  
> 关联专题：专题 5-1（总体架构）、专题 5-2（拓扑排序）、专题 5-3（生成与写入）

---

## 1. 运行期可观测性

### 1.1 设计目标

用户点击"确认执行"后，需要清晰地知道：
- 任务整体进展到哪了（正在执行哪张表）；
- 每张表写入了多少行、还剩多少；
- 有没有失败，失败原因是什么；
- 任务何时结束。

可观测性的设计需要**避免两个极端**：更新太频繁（每行都通知，前端卡顿）或太稀疏（每张表只通知一次，大表感觉卡住）。

### 1.2 进度事件模型

引擎定义 7 类进度事件，形成完整的任务生命周期通知体系：

```
// 事件基类
Event {
  event_type: string
  task_id: bigint
  timestamp: datetime
}

// 任务开始
TaskStarted extends Event {
  event_type: "task_started"
  total_tables: int
  total_estimated_rows: int
}

// 单表开始
TableStarted extends Event {
  event_type: "table_started"
  project_table_id: bigint
  table_name: string
  execution_order: int
  estimated_rows: int
  truncate_before: bool
}

// 批次完成（最高频事件）
BatchCompleted extends Event {
  event_type: "batch_completed"
  project_table_id: bigint
  rows_written_so_far: int   // 本表累计写入行数
  estimated_rows: int        // 本表预估总行数（用于进度条计算）
  batch_duration_ms: int     // 本批次耗时（用于速率显示）
}

// 单表完成
TableCompleted extends Event {
  event_type: "table_completed"
  project_table_id: bigint
  table_name: string
  rows_written: int
  status: "SUCCESS" | "FAILED"
  error_message: string | null
  duration_ms: int
}

// 单表跳过
TableSkipped extends Event {
  event_type: "table_skipped"
  project_table_id: bigint
  table_name: string
  reason: string   // 例如："前置依赖表 orders 执行失败"
}

// D-10 警告（在 Preflight 完成、执行开始前发出）
ValidationWarning extends Event {
  event_type: "validation_warning"
  warning_type: "D10_MULTIPLIER_CORRECTED" | "FK_POOL_SAMPLED" | ...
  table_name: string
  message: string
}

// 任务结束
TaskCompleted extends Event {
  event_type: "task_completed"
  status: "SUCCESS" | "PARTIAL_FAILED" | "FAILED"
  total_rows_written: int
  duration_ms: int
}
```

### 1.3 事件发送频率控制

`BatchCompleted` 是最高频事件，可能每秒发出数十次（高速写入时）。需要在发送频率和实时性之间平衡：

```
// 频率限流策略
const MIN_EMIT_INTERVAL_MS = 200   // 每张表至多每 200ms 发送一次 BatchCompleted

last_emit_time: Map<table_id, timestamp>

function emit_batch_completed(event):
  now = current_time_ms()
  last = last_emit_time.get(event.project_table_id, 0)
  
  if (now - last) >= MIN_EMIT_INTERVAL_MS:
    ProgressEmitter.emit(event)
    last_emit_time[event.project_table_id] = now
  // 否则丢弃本次事件（下次超过间隔时会发送最新状态）
```

**特殊情况：最后一个批次强制发送**，确保用户看到最终写入数。

### 1.4 进度事件的传输机制

LoomiDBX 是桌面端应用，引擎（后端）和 UI（前端）之间的通信方式取决于桌面框架：

| 框架 | 推荐机制 | 说明 |
|------|---------|------|
| Electron（主进程↔渲染进程）| IPC + `ipcMain.emit` | 进程内事件，无网络开销 |
| Tauri（Rust 后端↔WebView）| Tauri Events API | 原生跨进程事件 |
| 本地 HTTP 服务（前后端分离）| SSE（Server-Sent Events）| 单向推送，比 WebSocket 轻量 |

无论哪种机制，引擎侧的 `ProgressEmitter` 只负责"发布"事件，具体传输层由外部注入的适配器实现，保持引擎核心的传输无关性。

### 1.5 前端进度展示设计建议

基于上述事件模型，前端可以渲染以下信息：

```
任务执行中
══════════════════════════════════════════════════════

整体进度  [████████░░░░░░░░░░░░]  3 / 5 张表完成

✓ customers       1,000 行    完成  (1.2s)
✓ orders          5,000 行    完成  (4.8s)
● order_items    正在写入...
    [██████████████░░░░░░]  8,432 / 12,500  67%  ≈ 2,100 行/秒
○ order_payment_rel          等待中
○ reviews                    等待中

══════════════════════════════════════════════════════
已运行 6.2s  |  已写入 14,432 行
```

进度条的计算：`progress = rows_written_so_far / estimated_rows`（当 `estimated_rows` 为估算值时，进度条可能超过 100%，需处理 > 100% 的情况，例如锁定显示 99% 直到完成）。

写入速率：`rate = batch_size / batch_duration_ms × 1000`（行/秒），用滑动窗口取最近 3 批次的平均速率，避免抖动。

---

## 2. 任务取消

### 2.1 取消机制设计

用户需要能在执行过程中随时取消任务（例如发现参数配置有误）。

```
// 取消令牌（Cancellation Token）
CancellationToken {
  is_cancelled: bool = false
  cancel(): void { is_cancelled = true }
}
```

取消令牌注入到 `TableExecutor` 中，在每个批次生成/写入的循环间隙检查：

```
while rows_remaining > 0:
  if cancel_token.is_cancelled:
    // 当前批次若已提交则保留，若未提交则回滚
    throw CancellationError()
  
  batch = RowBuilder.build(batch_size)
  DBWriter.write(batch)
  ...
```

### 2.2 取消后的状态处理

- 当前表：`rows_written` 记录已写入数量，`status = FAILED`，`error_message = "用户取消"`。
- 后续所有表：`status = SKIPPED`。
- `ExecutionTask.status = FAILED`，`ended_at = NOW()`。
- ProgressEmitter 发出 `TaskCompleted` 事件（状态 FAILED）。

**取消不做已写入数据的回滚**（与写入失败的处理一致），用户需在下次执行前手动决定是否 TRUNCATE。

---

## 3. 执行历史

### 3.1 记录写入时机

执行历史分散在整个任务生命周期中逐步写入，而非在最后一次性写入：

| 写入时机 | 写入内容 |
|---------|---------|
| Phase 1 装配完成后 | 创建 `ExecutionTask`（status=RUNNING, started_at=NOW） |
| 每张表进入执行前 | 创建 `ExecutionTableResult`（status=PENDING） |
| 每张表开始执行时 | 更新 `ExecutionTableResult.status = RUNNING` |
| 每批次写入成功后 | 更新 `ExecutionTableResult.rows_written`（实时更新） |
| 每张表完成时 | 更新 `ExecutionTableResult.status / error_message` |
| 任务结束时 | 更新 `ExecutionTask.status / ended_at` |

实时更新 `rows_written` 的意义：即使程序崩溃，历史记录中也能看到当时写入了多少行，便于诊断问题。

### 3.2 ExecutionTableResult 的 execution_order 字段

`ExecutionTableResult.execution_order`（数据模型中的字段）在创建记录时从 `ProjectTable.execution_order` 复制快照，确保即使 ProjectTable 后续重新排序，历史记录中的执行顺序仍反映当时的实际顺序。

### 3.3 历史列表查询

用户在历史记录页面可以查看：

```
// 查询某 Project 的执行历史
SELECT et.*,
       COUNT(etr.id) AS table_count,
       SUM(etr.rows_written) AS total_rows_written,
       COUNT(CASE WHEN etr.status = 'FAILED' THEN 1 END) AS failed_tables
FROM ExecutionTask et
LEFT JOIN ExecutionTableResult etr ON et.id = etr.execution_task_id
WHERE et.project_id = ?
GROUP BY et.id
ORDER BY et.started_at DESC
LIMIT 20
```

历史详情（展开某次执行）：

```
// 某次执行的所有表结果
SELECT etr.*
FROM ExecutionTableResult etr
WHERE etr.execution_task_id = ?
ORDER BY etr.execution_order ASC
```

### 3.4 失败记录的可读性

`ExecutionTableResult.error_message` 应面向用户可读，而非原始 DB 异常堆栈：

```python
# 错误信息的三层转化
def format_error(db_exception: DBException) -> str:
    
    # 层 1：识别常见错误类型并给出友好描述
    if is_fk_violation(db_exception):
        return f"外键约束违反：列 {extract_column()} 的值在关联表中不存在。" \
               f"请检查 FK Pool 是否为空或关联关系配置是否正确。"
    
    if is_unique_violation(db_exception):
        return f"唯一约束违反：列 {extract_column()} 存在重复值。" \
               f"请检查该列的生成器是否启用了 unique=true，或扩大取值范围。"
    
    if is_null_violation(db_exception):
        return f"非空约束违反：列 {extract_column()} 生成了 NULL 值，" \
               f"但该列不允许为空。请检查 GeneratorConfig 配置。"
    
    # 层 2：其他已知类型
    if is_connection_error(db_exception):
        return f"目标数据库连接中断：{db_exception.message}"
    
    # 层 3：未知错误附带原始信息
    return f"写入失败：{db_exception.message}（原始错误：{db_exception.raw}）"
```

### 3.5 重跑（Re-run）

用户可基于一次历史执行记录快速发起重跑：

```
// 重跑逻辑：复用同一 Project 配置，创建新的 ExecutionTask
function rerun(execution_task_id):
  old_task = load(ExecutionTask, execution_task_id)
  // 触发 ExecutionEngine.start(old_task.project_id)
  // 与普通执行完全相同，project_id 不变
```

重跑不会改变原有历史记录，而是创建一条全新的 `ExecutionTask`。

---

## 4. 补充关键问题

### 4.1 生成种子（Seed）与可复现性

生成器配置中包含 `seed` 参数，用于固定随机数序列，使得相同配置可以重现完全一致的数据。

**实现策略**：

```
// 每列的有效 seed = 用户配置的 seed（若非 null）或 全局任务 seed + column_id 的哈希偏移
effective_seed(col_id, user_seed, task_seed):
  if user_seed is not null:
    return user_seed
  else:
    return task_seed ^ hash(col_id)   // XOR 确保不同列得到不同 seed
```

`task_seed`：每次执行任务时生成一个随机整数作为任务级种子，存入 `ExecutionTask`（需在数据模型中扩展此字段，或记录在日志中）。若用户希望完全复现，需在下次执行前指定相同的 task_seed。

**JoinTable 的 seed 问题**：JoinTable 的行构建依赖随机取样（为每个 B1 记录随机选 B2），这个随机性也需要纳入 seed 管理，确保可复现。

**当前版本的最小化实现**：不要求 100% 可复现，仅保证在 `seed` 非 null 的列上输出稳定。完整可复现作为未来增强。

### 4.2 目标库连接管理

引擎执行期间需要持续使用目标数据库连接，需注意以下问题：

**连接复用**：整个任务执行过程中使用单一长连接（或小连接池），而非每批次新建连接：

```python
# 任务开始时建立连接
connection = ConnectionPool.get(project.connection_id, pool_size=2)

# 所有 DBWriter 操作复用此连接
# 任务结束时归还
connection.close()
```

**连接超时**：大表执行时间可能超过数据库的 `wait_timeout` 或 `idle_timeout`。需要：
- 设置较长的连接超时（或禁用自动断开）。
- 每隔 N 批次发送一次 keepalive 查询（如 `SELECT 1`）。

**目标库与本地库的连接隔离**：注意区分"本地应用数据库"（存储 ProjectTable、GeneratorConfig 等）和"目标数据库"（数据写入目的地）。两者使用独立的连接，绝对不能混用。

### 4.3 表并发执行的架构预留

v1 采用串行执行，但以下设计决策为 v2 并行执行预留了空间：

**已预留的设计**：
- FK Pool 设计为共享数据结构（加锁或用 thread-safe 实现即可支持并发写入）。
- 拓扑排序结果含"层级（level）"标注（同层表无依赖，可并行）。
- 每张表的执行逻辑完全封装在 `TableExecutor` 中，天然可分发到线程/协程。

**v2 并行执行的核心问题**：
- 同层的多张表同时写入目标库，需要连接池（pool_size = 并发表数）。
- 进度事件来自多个并发执行器，前端需按 `project_table_id` 分组显示。
- 任一表失败时，是否取消同层正在执行的其他表（需要 CancellationToken 广播）。

### 4.4 Schema 变更与配置过期的运行时处理

预检阶段已捕获 `config_status = NEEDS_REVIEW` 的列（作为 Warning 或 BlockingError），但以下情况可能在预检之后、执行中途发生：

**场景**：A 表正在执行写入，此时用户（或外部进程）对目标库的 B 表结构做了变更（如加了 NOT NULL 列），当引擎轮到执行 B 表时，INSERT 会因为 schema 不匹配而报错。

**处理方式**：这属于写入失败，按常规失败流程处理（记录 error_message、标记 FAILED/SKIPPED），无需特殊处理。建议在 error_message 中提示"建议重新扫描 Schema 并核查相关配置"。

### 4.5 行数为 0 的特殊情形

某些情形下，计算出的实际行数为 0：
- Parent 表 `row_count = 0`（用户有意为之）。
- Child 表的 FK Pool 为空（父表一行都没写成功，或 rel_source_sql 查无数据）。

**处理策略**：
- 行数为 0 时，跳过生成和写入（不执行 INSERT），直接将 `ExecutionTableResult.rows_written = 0, status = SUCCESS`。
- **不视为失败**，但需发出 Warning 提示用户注意。
- 对于 FK Pool 为空导致的行数 = 0，Warning 消息应明确说明原因（"父表 XXX 无有效记录，本表生成了 0 行"）。

### 4.6 ClickHouse 和 Hive 的特殊适配

ClickHouse 和 Hive 不是传统 OLTP 数据库，在 INSERT 行为上有差异：

**ClickHouse**：
- 支持批量 INSERT，但不支持事务（无 ROLLBACK）。
- 每次 INSERT 实际写入一个 Part，频繁小 INSERT 会触发大量 Part 合并，影响性能。
- 建议对 ClickHouse 目标将 `BATCH_SIZE` 提高到 10,000 ~ 50,000。
- 无 AUTO_INCREMENT 概念，PK 捕获使用查询方案（见专题 5-3 §3.4）。

**Hive**：
- 传统 Hive 不支持行级 INSERT，只支持 `INSERT INTO SELECT ...`。
- 需要先将数据写入临时文件（CSV/Parquet），再通过 `LOAD DATA` 或 `INSERT OVERWRITE` 写入。
- 建议对 Hive 目标实现独立的 `HiveDBWriter`，替换标准批量 INSERT 逻辑。
- 无事务支持，失败处理退化为记录已写入的文件批次。

### 4.7 执行引擎的幂等性

同一 Project 执行两次，会产生两批数据（除非勾选了 `truncate_before`）。这是预期行为，引擎本身不做去重。

但以下情况需要特别注意：
- 若表有 AUTO_INCREMENT PK，两次执行的数据 PK 值不会冲突（递增分配），完全安全。
- 若表有手动生成的 UNIQUE 整数 PK（如 `distribute_int` 加 `unique = true`），第二次执行的值池仍从头生成，可能与第一次的值冲突。
  - 对策：在开始生成前查询目标表的 `MAX(pk_column)`，将生成器的起始值设为 `MAX + 1`（仅适用于整数 PK，需在生成器 config 中支持 `start_from_db_max` 选项）。
  - 这是一个"未来增强"特性，当前版本文档中作为 Known Limitation 注明。

### 4.8 整合清单：产品规则到引擎行为的映射

| 产品大纲规则（§7） | 引擎实现位置 |
|---|---|
| 字段生成规则与表执行规则分离 | ConfigLoader 分别加载 GeneratorConfig 和 ProjectTable，互不交叉 |
| 外键字段优先遵从关联关系 | RowBuilder 中 FK 列优先从 FK Pool 取值，不调用普通生成器 |
| 约束优先于展示效果 | D-10 自动修正、唯一性去重、FK Pool 保证 FK 合法性 |
| 不可生成字段必须有依据 | Preflight 配置完整性校验（§5-1 Phase 2）|
| 依赖表先生成 | 拓扑排序 + execution_order 串行执行 |
| 生成失败应中止当前批次 | 批次级事务回滚 + TableExecutor 失败中止 |
| 结构变更必须提示影响范围 | GeneratorConfig.config_status = NEEDS_REVIEW 在 Preflight 中报 Warning/BlockingError |

---

## 5. 未解决问题与后续专题入口

以下问题在本引擎设计中留有接口但未完整展开，待后续专题明确：

| 问题 | 待定内容 | 相关专题 |
|------|---------|---------|
| 计算字段（表达式列）的求值 | sql_expression 的跨数据库 SELECT 构造细节；python_expression 的安全沙箱实现 | 专题 7 |
| AI 生成器的接入 | AI 生成器如何嵌入 GeneratorRegistry；AI 调用失败的降级策略 | 专题 8 |
| 外部数据源（dict_table、external_data_source）的加载与缓存 | 外部数据的懒加载时机、缓存策略、内存占用上限 | 专题 6 |
| 各数据库的类型映射 | 不同数据库 `data_type` 字符串到引擎内部 `data_mapping_type` 的映射规则 | 专题 10 |
| rel_source_sql 的安全性与校验 | 是否限制 SQL 只能是 SELECT；如何防止用户写入危险 SQL | 专题 9/11 |
| 大规模执行的性能基准 | BATCH_SIZE 调优；ClickHouse/Hive 的写入吞吐基准 | 专题 10 |

---

*专题 5（数据生成执行引擎设计）完。共 4 篇子文档：*  
*5-1 总体架构与执行生命周期 | 5-2 拓扑排序与行数预计算 | 5-3 生成与写入引擎 | 5-4 本篇*
