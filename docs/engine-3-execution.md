# 专题 5-3：数据生成执行引擎设计 — 生成与写入引擎

> 本文档是专题 5 的第三部分，聚焦于单表的数据生成策略、FK Pool 管理、批量写入机制、失败处理与事务模型。  
> 依赖文档：数据模型（专题 2）、生成器设计（专题 4）、专题 7（计算字段）  
> 关联专题：专题 5-1（总体架构）、专题 5-2（拓扑排序与行数预计算）

---

## 1. 边生成边写入模型

### 1.1 为什么不用"全量生成再批量写入"

若一张表需要生成 100 万行，将所有行先在内存中构建完毕再一次性 INSERT，存在以下问题：
- 内存峰值极高（100 万行 × 若干列的数据对象）。
- 若写入过程中失败，已生成的大量数据全部作废，浪费计算资源。
- 用户在很长一段时间内看不到任何写入进度。

### 1.2 批次处理模型

引擎采用**固定大小批次（Fixed-Batch）**的边生成边写入模型：

```
╔══════════════════════════════════════════════════════╗
║  TableExecutor 主循环                                 ║
║                                                      ║
║  rows_remaining = actual_row_count                   ║
║  while rows_remaining > 0:                           ║
║    batch_size = min(BATCH_SIZE, rows_remaining)      ║
║    rows = RowBuilder.build(batch_size)         ──►  ║ 生成
║    result = DBWriter.write(rows)               ──►  ║ 写入
║    FKPool.add(table_id, result.pks)            ──►  ║ 积累 FK
║    rows_remaining -= result.rows_written             ║
║    ProgressEmitter.emit(BatchCompleted)              ║
╚══════════════════════════════════════════════════════╝
```

**BATCH_SIZE 的选取**：

| 场景 | 推荐值 | 说明 |
|------|--------|------|
| 默认 | 1,000 行 | 平衡内存与进度更新频率 |
| 列数较多（> 30 列）| 500 行 | 每行对象较大，减小批次 |
| 简单表（< 10 列，无表达式）| 2,000 行 | 减少 INSERT 往返次数 |
| 用户可配置（系统设置）| — | 允许高级用户覆盖默认值 |

### 1.3 批次内的内存布局

每个批次在内存中表示为一个行列表，每行是一个字典：

```
batch = [
  { "id": 1, "name": "Alice", "email": "alice@example.com", "age": 28 },
  { "id": 2, "name": "Bob",   "email": "bob@example.com",   "age": 35 },
  ...  // batch_size 行
]
```

批次完成写入并获取 PK 反馈后立即丢弃，不在内存中累积。

---

## 2. FK Pool 设计

### 2.1 核心数据结构

FK Pool 是整个引擎中跨表传递 PK 值的中央存储，贯穿整个任务的执行生命周期。

```
FKPool {
  store: Map<table_id, PkValue[]>   // table_id → PK 值列表
  
  // 追加 PK 值（父表每批写入后调用）
  add(table_id: bigint, pks: PkValue[]): void
  
  // 取样（子表生成行时调用）
  sample(table_id: bigint, n: int, without_replacement: bool): PkValue[]
  
  // 获取当前池大小
  count(table_id: bigint): int
  
  // 从 rel_source_sql 结果预填充（Preflight 阶段）
  load_from_query(table_id: bigint, sql_result: PkValue[]): void
}
```

### 2.2 FK Pool 的值来源策略映射

FK Pool 的填充方式由 `ProjectTableRelation.rel_value_source` 决定：

| `rel_value_source` | FK Pool 填充时机 | 填充来源 |
|--------------------|----------------|--------|
| `FROM_EXECUTION` | 父表每批写入成功后 | 本次写入的 PK 值 |
| `FROM_DB_QUERY` | Preflight 阶段（或执行前）| 执行 `rel_source_sql` 查询目标库 |
| `MERGED` | 两个时机都有 | 两者合并 |

**`FROM_DB_QUERY` 和 `MERGED` 的 SQL 执行时机**：

Preflight 阶段只做 `COUNT(*)` 用于行数估算。实际的 PK 值查询（`SELECT pk_column FROM ... WHERE ...`）在进入执行阶段后、该子表开始生成之前执行。这样做的原因：
- 如果父表在 Project 中（`FROM_EXECUTION`），行数是动态的，要等父表执行完才能确定。
- 若在 Preflight 时查询全量 PK 值，时间窗口较长，数据库存量可能已变化。
- 在执行前一刻查询，能获取到最新的存量，也包括了本次 Project 中其他表（若已执行完）刚写入的数据。

### 2.3 复合外键的处理

若外键涉及多列（复合 FK），FK Pool 存储的是**元组**而非单个值：

```
// 复合 FK 示例：(order_id, line_no) 共同构成外键
store: Map<relation_id, Tuple[]>

// 取样时返回完整元组
sample(relation_id, n) → [
  (order_id=1, line_no=1),
  (order_id=1, line_no=2),
  ...
]
```

注意：复合 FK Pool 以 `relation_id` 而非 `table_id` 为 Key，因为同一张父表可能被多个不同的 FK 关系引用，每个关系的引用列组合不同。

### 2.4 JoinTable 的 FK 取样逻辑

JoinTable（N:N 连结表）需要同时从多个 BaseTable 的 FK Pool 中取值，并且通常需要保证组合唯一性（复合 UNIQUE 约束）。

以 `order_payment_rel` 为例（含 `order_id`, `payment_id` 复合唯一约束）：

```
// 遍历所有 B1（orders）记录
for each order_pk in fk_pool.get_all(orders.table_id):
  
  // 对每个 order，随机决定关联多少个 payment（受 D-10 修正后的 effective_max 约束）
  k = random_int(ptr.multiplier_min, runtime_effective_max[ptr.id])
  
  // 从 B2（payments）FK Pool 中取样 k 个不同的 payment（无重复取样）
  payment_pks = fk_pool.sample(payments.table_id, k, without_replacement=true)
  
  // 构建 JoinTable 的 k 行
  for each payment_pk in payment_pks:
    row = {
      "order_id":   order_pk,
      "payment_id": payment_pk,
      // 其他列由对应的生成器生成
    }
    batch.append(row)
    if batch.size >= BATCH_SIZE:
      DBWriter.write(batch)
      batch.clear()

// 最后一个不完整批次
if batch.size > 0:
  DBWriter.write(batch)
```

**无重复取样的实现**：

对于每个 `order_pk`，从 payment_pool 中无重复取样 `k` 个。当 payment_pool 大小为 P，且对所有 order 都需要取样时，需要管理已使用的组合对集合（`used_pairs`）以防止跨 order 的重复：

```
used_pairs: Set<(order_id, payment_id)>

for each order_pk:
  available = [p for p in payment_pool if (order_pk, p) NOT IN used_pairs]
  k = min(random_int(min, effective_max), len(available))
  selected = random_sample(available, k, without_replacement=true)
  for payment_pk in selected:
    used_pairs.add((order_pk, payment_pk))
    // 构建行...
```

**内存注意**：`used_pairs` 的规模 = JoinTable 的生成行数。对于行数极大的 JoinTable，可改为在写入前做数据库唯一性检查（INSERT IGNORE 或 ON CONFLICT DO NOTHING），但会增加 DB 往返。推荐在行数 < 100 万时使用内存 Set，超过时切换为 DB 侧去重策略。

### 2.5 FK Pool 的内存上限与降级策略

若父表生成记录极多（如 1,000 万行），FK Pool 中存储的 PK 值量将非常大。

**内存估算**：
- BIGINT PK（8 字节）× 1,000 万行 = 约 80MB
- 若 PK 为 UUID 字符串（36 字节）× 1,000 万行 = 约 360MB

**应对策略**：

配置一个 `FK_POOL_MAX_SIZE`（默认 100 万），当父表记录数超过此阈值时，采用**均匀抽样降级（Reservoir Sampling）**：

```
if fk_pool.count(table_id) >= FK_POOL_MAX_SIZE:
  // Reservoir Sampling：保持 FK_POOL_MAX_SIZE 个样本
  // 新 PK 以 FK_POOL_MAX_SIZE / total_seen 的概率替换现有样本
  reservoir_sample(table_id, new_pk)
```

降级后，子表的 FK 值将从父表的一个代表性样本中取，而非全部父记录。这会使外键分布轻微偏斜，但对数据逼真性的影响在大规模场景下可接受。**系统应在此时向用户发出 Warning**。

---

## 3. 列级生成与行构建

### 3.1 列生成顺序的依赖管理

一行数据中的各列并非完全独立，表达式列（`sql_expression`、`python_expression`）需要引用同行中其他列的值。规则是：**表达式列只能引用非表达式列**（生成器文档约束），因此列依赖图是无环的两层结构：

```
层 1（并行生成）：所有非表达式列（FK 列、普通生成器列）
        ↓
层 2（顺序生成）：所有表达式列（sql_expression、python_expression）
```

**列生成顺序算法**：

```
function build_column_order(columns, generator_configs):
  layer1 = []  // 非表达式列
  layer2 = []  // 表达式列

  for each col in columns:
    cfg = generator_configs[col.id]
    if cfg.generator_name in ['sql_expression', 'python_expression']:
      layer2.append(col)
    else:
      layer1.append(col)

  // 验证表达式列中无相互引用
  for each expr_col in layer2:
    refs = parse_column_refs(cfg.params.expression)  // 提取 ${col_name}
    for ref in refs:
      if ref in [c.column_name for c in layer2]:
        raise ConfigError("表达式列 {expr_col.column_name} 引用了另一个表达式列 {ref}")

  return layer1 + layer2   // 层 1 先，层 2 后
```

### 3.2 RowBuilder 执行流程

```
function build_row(columns, generator_configs, fk_pool, used_values):

  row = {}

  // --- 层 1：非表达式列 ---
  for each col in layer1_columns:
    cfg = generator_configs[col.id]
    
    // FK 列：从 FK Pool 取值
    if col is foreign_key:
      value = fk_pool.sample(related_table_id, 1)[0]
    
    // 普通生成器列：调用生成器
    else:
      value = GeneratorRegistry.generate(cfg, row_context={})
    
    // 应用 null_percentage
    if cfg.null_percentage > 0 and col.is_nullable:
      if random() < cfg.null_percentage:
        value = NULL
    
    row[col.column_name] = value

  // --- 层 2：表达式列 ---
  for each col in layer2_columns:
    cfg = generator_configs[col.id]
    
    // 将层 1 已生成的值作为上下文传入表达式求值
    value = evaluate_expression(cfg.params.expression, context=row)
    
    // 应用 null_percentage（表达式列也支持）
    if cfg.null_percentage > 0 and col.is_nullable:
      if random() < cfg.null_percentage:
        value = NULL
    
    row[col.column_name] = value

  return row
```

### 3.3 UNIQUE 列的去重策略

对于标记了 `unique = true` 的列（对应数据库 UNIQUE 约束列），生成器必须保证每批生成的值在全表范围内唯一。

**策略 A：内存 Set 去重（推荐，行数 ≤ 500 万时）**

```
unique_tracker: Map<column_id, Set<value>>

function generate_unique(cfg, tracker):
  MAX_RETRY = 100
  for i in 0..MAX_RETRY:
    value = GeneratorRegistry.generate(cfg, {})
    if value NOT IN tracker[cfg.column_id]:
      tracker[cfg.column_id].add(value)
      return value
  // 重试 100 次仍有冲突 → 判断为唯一空间耗尽
  raise UniqueSpaceExhaustedError(cfg.column_id)
```

**策略 B：序列化生成（适用于整数类型）**

对于需要唯一整数的列，可使用内部自增序列而非随机重试：

```
sequence_tracker: Map<column_id, int>  // 初始化为当前数据库该列的 MAX 值 + 1

function generate_unique_int(cfg, tracker):
  value = tracker[cfg.column_id]
  tracker[cfg.column_id] += 1
  return value
```

**预检阶段的唯一空间预估**：

在 Preflight 阶段检测可能发生唯一空间耗尽的情况：

```
// 以 distribute_int 生成器为例
if generator = distribute_int AND col.has_unique_constraint:
  range_size = cfg.clamp_max - cfg.clamp_min  // 可用取值空间
  if estimated_rows > range_size × 0.9:       // 超过 90% 则警告
    emit Warning("列 {col.column_name} 的唯一取值空间
                 ({range_size}) 可能不足以容纳 {estimated_rows} 行")
  if estimated_rows > range_size:
    raise BlockingError("列 {col.column_name} 的唯一取值空间不足")
```

### 3.4 Auto-Increment 主键的 FK 值捕获

当父表的 PK 是数据库自动生成的（AUTO_INCREMENT / SERIAL / SEQUENCE），引擎不会主动生成 PK 值，PK 值在 INSERT 后由数据库分配。子表的 FK Pool 需要捕获这些由数据库分配的值。

**方案 A：RETURNING 语句（PostgreSQL / SQLite 支持）**

```sql
INSERT INTO customers (name, email) VALUES (...), (...)
RETURNING id;
```

DBWriter 解析 `RETURNING` 结果，将返回的 PK 值列表送入 FK Pool。

**方案 B：lastInsertId + 行数推算（MySQL）**

MySQL 的 `INSERT ... VALUES(多行)` 后，`LAST_INSERT_ID()` 返回第一条记录的自增 ID，若是连续自增，则后续 ID = first_id + 0, 1, 2, ...

```
result = execute("INSERT INTO customers (name, email) VALUES ...")
first_pk = get_last_insert_id()
batch_pks = [first_pk + i for i in range(batch_size)]
fk_pool.add(customers.table_id, batch_pks)
```

**注意**：若 MySQL 表存在并发写入（其他进程也在插入），ID 可能不连续，此方案失效。在 LoomiDBX 的使用场景下（测试/演示环境，通常单进程写入），这个假设是合理的，但应在文档中注明此限制。

**方案 C：写入后查询（兼容所有数据库的兜底方案）**

```sql
INSERT INTO customers (name, email) VALUES ...;
SELECT id FROM customers ORDER BY id DESC LIMIT {batch_size};
```

性能最差，但兼容性最好。适合作为对不支持 RETURNING 且无法保证连续自增的数据库的降级方案。

**DBWriter 的 PK 捕获策略矩阵**：

| 目标数据库 | PK 捕获方案 | 备注 |
|-----------|------------|------|
| PostgreSQL | RETURNING | 原生支持，最优 |
| SQLite | RETURNING | v3.35+ 支持 |
| MySQL | lastInsertId + 推算 | 需单进程写入保证连续性 |
| Oracle | RETURNING INTO | 原生支持 |
| MSSQL | OUTPUT INSERTED.id | 原生支持 |
| ClickHouse | 查询方案 C | ClickHouse 无标准自增 PK |
| Hive | 查询方案 C | Hive 无 AUTO_INCREMENT |

---

## 4. 批量写入与事务策略

### 4.1 事务边界的选择

三种候选方案：

**方案 X：表级大事务**
- 整张表的所有行在一个事务中提交。
- 优点：原子性强（要么全写，要么全不写）。
- 缺点：大事务持续时间长，占用锁、undo log 等资源；对百万行级别的表完全不可行。

**方案 Y：批次级独立事务**（推荐）
- 每个批次（BATCH_SIZE 行）独立开启/提交一个事务。
- 优点：内存小、进度可见、失败只影响当前批次。
- 缺点：单表写入不具备完整原子性（部分成功是可能的）。

**方案 Z：无事务（AUTO_COMMIT）**
- 每行单独提交。
- 性能最差，不推荐。

**推荐方案 Y（批次级独立事务）**，结合"失败即停止当前表"和 Project 级失败传播策略：

- 批次内写入失败 → 回滚当前批次（该批次数据不进入库）→ 停止该表 → 记录失败状态与已写入行数。
- 已成功提交的批次数据**不回滚**（符合产品大纲 §7 规则"生成失败应中止当前批次"，且避免大规模回滚的性能问题）。
- 这里的**中止**是执行调度层面的操作，不等同于回滚整个 Project 或回滚已提交批次：
  - 当前表必须中止，不再继续生成或写入后续批次。
  - 若失败表存在依赖它的子表，则必须中止当前 Project 的后续处理，并将受影响的下游依赖表标记为 `SKIPPED`，避免连锁约束风险。
  - 若失败表没有子表依赖，则是否中止整个 Project 取决于用户是否启用 `失败即终止`；未启用时仅记录局部失败，后续无依赖风险的表可以继续执行。
  - 无论是哪种中止路径，已经提交成功的历史批次都不做补偿删除或事务回滚。

### 4.2 批次失败的处理流程

```
try:
  connection.begin()
  pks = DBWriter.insert_batch(table, rows)
  connection.commit()
  fk_pool.add(table_id, pks)
  rows_written_total += len(rows)
  ProgressEmitter.emit(BatchCompleted, rows_written_total)

except DBException as e:
  connection.rollback()
  // 记录失败信息
  ExecutionTableResult.error_message = e.message
  ExecutionTableResult.rows_written = rows_written_total  // 记录已成功写入的
  ExecutionTableResult.status = FAILED
  ProgressEmitter.emit(TableFailed, e.message)
  // 抛出给上层调度器，由其按失败传播策略决定：
  // 1. 当前表中止；
  // 2. 若存在子表依赖，则 Project 强制中止并传播 SKIPPED；
  // 3. 若无子表依赖，则根据用户的 fail_fast 选项决定是否中止 Project。
  raise TableExecutionError(e)
```

### 4.3 TRUNCATE 的执行顺序问题

`truncate_before = true` 看似是简单的"写入前先清空"，但涉及 FK 约束时有陷阱：

**问题**：若 `orders` 和 `order_items` 都勾选了 `truncate_before`，执行顺序是 `orders` 先（execution_order 小），那么 TRUNCATE `orders` 时，`order_items` 中仍有引用 `orders` 的 FK 数据，TRUNCATE 可能因 FK 约束失败。

**解决方案**：TRUNCATE 阶段单独处理，按**逆拓扑顺序**（即子表先于父表）执行所有需要 TRUNCATE 的表：

```
// 阶段 0：预先执行所有 TRUNCATE（在任何生成之前）
tables_to_truncate = [pt for pt in sorted_tables if pt.truncate_before]

// 按 execution_order 降序（逆序）TRUNCATE
for pt in reverse(tables_to_truncate):
  DBWriter.truncate(pt.table)
```

这样保证子表先被清空，再清空父表，不会违反 FK 约束。

**对于 MySQL 等需要暂时禁用 FK 检查的数据库**，可在 TRUNCATE 前后执行：
```sql
SET FOREIGN_KEY_CHECKS = 0;
TRUNCATE TABLE ...;
TRUNCATE TABLE ...;
SET FOREIGN_KEY_CHECKS = 1;
```

注意：在任务结束（无论成功或失败）后都需确保重新启用 FK 检查。

---

## 5. 实际行数的运行时决定

预检阶段输出的是**估算行数**。进入执行阶段后，在各表开始生成之前，需要基于**父表实际写入行数**（FK Pool 的真实大小）重新计算**实际生成行数**（`actual_row_count`）。

```
function resolve_actual_row_count(pt, fk_pool, execution_plan):
  
  // 情形 1：Parent 或 BaseTable
  if pt.row_count IS NOT NULL and no parent in project:
    return pt.row_count

  // 情形 2：Child，Parent 在 Project 中（row_count IS NULL）
  if is_child_with_parent_in_project(pt):
    total = 0
    parent_pks = fk_pool.get_all(parent.table_id)  // 父表实际 PK 列表
    for each parent_pk in parent_pks:
      k = random_int(ptr.multiplier_min, ptr.multiplier_max)
      total += k
    return total

  // 情形 3：Child，Parent 不在 Project 中
  if is_child_without_parent(pt):
    parent_pool_size = fk_pool.count(relation.id)  // 从 DB 查询后预填充的
    total = 0
    for i in range(parent_pool_size):
      k = random_int(ptr.multiplier_min, ptr.multiplier_max)
      total += k
    return min(total, pt.row_count)  // 受 row_count 上限约束

  // 情形 4/5/6：JoinTable
  if is_join_table(pt):
    // 由 JoinTable 执行逻辑（§2.4）自然产生，不在此预设总数
    // 返回驱动 BaseTable 的 FK Pool 大小 × effective_max 作为上界
    return calculate_join_table_upper_bound(pt, fk_pool)
```

> **说明**：情形 2 和 3 中，`actual_row_count` 是对每条父记录随机抽取倍数后的累加，因此每次执行结果会略有不同（若 multiplier_min ≠ multiplier_max）。这是预期行为，与产品设计一致。

---

## 6. DBWriter 的跨数据库抽象

### 6.1 接口定义

```
interface DBWriter:
  // 批量插入，返回实际写入的行数和分配的 PK 值
  write_batch(
    table: DbTable,
    columns: DbColumn[],
    rows: Row[],
    pk_column: DbColumn | null   // null 表示无需捕获 PK（或 PK 不是 Auto-Increment）
  ) → InsertResult { rows_written: int, pks: PkValue[] }

  // 清空表
  truncate(table: DbTable): void

  // 执行任意查询（用于 rel_source_sql 和 D-10 校验）
  query(sql: string): ResultSet
  
  // 执行 rel_source_sql 并返回 FK 值列表
  fetch_fk_values(sql: string, pk_column_alias: string): PkValue[]
```

### 6.2 NULL 值处理

不同数据库对 NULL 的 INSERT 语法略有差异，DBWriter 需统一处理：

```sql
-- 包含 NULL 值的 INSERT（通用写法）
INSERT INTO users (name, email, age) VALUES ('Alice', 'alice@example.com', NULL);
```

对于列有 `DEFAULT` 值且该列未生成值（`is_nullable = FALSE` 但有 `default_value`），应在 INSERT 时**省略该列**，让数据库使用默认值，而非显式插入 NULL（显式插 NULL 会覆盖 DEFAULT）：

```python
# RowBuilder 中：不生成该列的键，INSERT 语句中不包含该列
if col.default_value IS NOT NULL and value IS NULL:
  row.pop(col.column_name, None)  # 从 row 字典中移除，让 DB 使用默认值
```

### 6.3 INSERT 语句构造

推荐使用参数化 INSERT（防 SQL 注入，同时让驱动层处理类型转换）：

```sql
-- MySQL / PostgreSQL 风格
INSERT INTO customers (id, name, email, age)
VALUES (%s, %s, %s, %s), (%s, %s, %s, %s), ...

-- Oracle 风格
INSERT ALL
  INTO customers (id, name, email) VALUES (:1, :2, :3)
  INTO customers (id, name, email) VALUES (:1, :2, :3)
SELECT * FROM dual
```

DBWriter 根据 `Connection.db_type` 选择对应的参数化格式。

---

## 7. 表达式生成器的求值

### 7.1 sql_expression 求值

`sql_expression` 允许引用同行其他列的值（`${column_name}` 语法）。由于这是 SQL 表达式，需要通过数据库引擎求值。

**方案：SELECT 求值**

对于每批次的表达式列，构造一个无 FROM 的批量 SELECT：

```sql
-- 表达式：CONCAT('ORD-', DATE_FORMAT(${order_time}, '%Y%m%d'), LPAD(${seq}, 4, '0'))
-- 批次 1000 行的求值

SELECT
  CONCAT('ORD-', DATE_FORMAT('2024-01-15', '%Y%m%d'), LPAD(1, 4, '0')),
  CONCAT('ORD-', DATE_FORMAT('2024-01-16', '%Y%m%d'), LPAD(2, 4, '0')),
  ...  -- 1000 行的值
```

或者使用 `UNION ALL` 形式以兼容更多数据库：

```sql
SELECT CONCAT('ORD-', ...) AS computed_val FROM (
  SELECT '2024-01-15' AS order_time, 1 AS seq
  UNION ALL
  SELECT '2024-01-16' AS order_time, 2 AS seq
  ...
) t
```

这需要一次额外的 DB 往返（SELECT），但可以复用当前连接，且通常比逐行求值高效。

### 7.2 python_expression 求值

`python_expression` 在引擎本地求值，无需 DB 往返。

```python
# 表达式：${name}.upper() + '_' + str(${id})
# 对批次中每行求值

import ast

def evaluate_python_expr(expression: str, row: dict) -> any:
    # 将 ${column_name} 替换为实际值
    code = re.sub(r'\$\{(\w+)\}', lambda m: repr(row[m.group(1)]), expression)
    
    # 安全求值：仅允许标准库
    allowed_builtins = {k: __builtins__[k] for k in ['str', 'int', 'float', 'len', 'round', ...]}
    result = eval(code, {"__builtins__": allowed_builtins}, {})
    return result
```

**安全注意事项**：
- 禁止 `import`、文件 I/O、网络访问等危险操作。
- 使用 `ast.parse()` 预检表达式，若含有 `Import`、`Call` 到非白名单函数等 AST 节点则在配置保存时拒绝。
- 求值超时保护（每个表达式设 1ms 超时，防止死循环）。

---

*下一篇：专题 5-4 — 可观测性、执行历史与补充关键问题 （[engine-4-observability.md](./engine-4-observability.md)）*
