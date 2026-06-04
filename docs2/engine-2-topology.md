# 专题 5-2：数据生成执行引擎设计 — 拓扑排序与行数预计算

> 本文档是专题 5 的第二部分，聚焦于表依赖图的构建、拓扑排序算法，以及预检阶段的行数计算。  
> 依赖文档：数据模型（专题 2）§8 生成任务、§12 D-05/D-06/D-08/D-09/D-10  
> 关联专题：专题 5-1（总体架构）、专题 5-3（生成与写入引擎）

---

## 1. 依赖图的建模

### 1.1 图的节点与边

拓扑排序的输入是当前 Project 中所有参与生成的表（`ProjectTable`）和它们之间的关系（`ProjectTableRelation`）。

在引擎内部，将依赖关系建模为**有向无环图（DAG）**，边的方向表示"必须先执行"：

```
父表（Parent / BaseTable）  →  子表（Child / JoinTable）
```

具体来说：

- **节点**：每个 `ProjectTable` 对应图中一个节点，节点 ID 即 `ProjectTable.id`。
- **有效边**：对每条 `ProjectTableRelation` 记录，仅当 `parent_project_table_id IS NOT NULL` 时，建立一条从父节点指向子节点的有向边。
  - 若 `parent_project_table_id IS NULL`（父表不在本 Project 中），此 Relation 不参与图的边构建，但仍用于行数计算（通过 `rel_source_sql`）。
- **JoinTable 的多父依赖**（数据模型 D-09）：N:N 关系被拆分为多条 `PARENT_CHILD` 类型的 Relation 存储，因此 JoinTable 在图中天然是一个拥有多条入边的节点，无需特殊处理。

### 1.2 图构建伪代码

```
function build_dependency_graph(project_tables, project_table_relations):
  graph = {}   // node_id → { in_degree: int, dependents: set<node_id> }
  
  for each pt in project_tables:
    graph[pt.id] = { in_degree: 0, dependents: {} }
  
  for each ptr in project_table_relations:
    if ptr.parent_project_table_id IS NOT NULL:
      parent = ptr.parent_project_table_id
      child  = ptr.child_project_table_id
      graph[child].in_degree += 1
      graph[parent].dependents.add(child)
  
  return graph
```

### 1.3 为何 JoinTable 天然融入标准算法

假设存在 `orders`（O）、`payments`（P）和连结表 `order_payment_rel`（R），N:N 关系拆为两条 Relation：
- `O → R`（JOIN_TABLE 类型，拆分后等价于 PARENT_CHILD）
- `P → R`（同上）

则 R 在图中有 2 条入边（in_degree = 2），只有当 O 和 P 都执行完毕，R 的 in_degree 才归零，才能入队执行。这与标准 Kahn 算法完全一致，不需要为多对多关系写任何特殊分支。

---

## 2. 拓扑排序算法

### 2.1 使用 Kahn 算法

选择 Kahn 算法（BFS 变体）而非 DFS，原因是：
1. 环检测更直观（处理结束后若仍有节点未出队，说明存在环）。
2. 输出结果直接是层序（相同层级的节点可在未来版本并行执行）。

```
function topological_sort(graph):
  queue = []   // 入度为 0 的节点
  order = []   // 输出的执行顺序

  // 初始化：将所有入度为 0 的节点入队
  for each (node_id, node) in graph:
    if node.in_degree == 0:
      queue.push(node_id)

  // 注意：同一层级内的排序需要确定性（避免每次执行顺序不同导致混乱）
  // 建议按 table_name 字母序或 project_table_id 升序作为同层排序的 tie-breaker
  sort queue by table_name ASC

  while queue is not empty:
    current = queue.pop_front()
    order.append(current)

    for each child_id in graph[current].dependents:
      graph[child_id].in_degree -= 1
      if graph[child_id].in_degree == 0:
        queue.push(child_id)
        sort queue by table_name ASC  // 保持确定性

  return order
```

### 2.2 将排序结果写入 execution_order

Kahn 算法输出的是一个有序列表（`order`），按列表下标赋值：

```
for i, project_table_id in enumerate(order):
  UPDATE ProjectTable
  SET execution_order = i + 1        // 从 1 开始
  WHERE id = project_table_id
```

根据数据模型 D-08，这一步在用户**保存 Project 配置时**执行，而非在任务执行时。引擎运行阶段直接读取已持久化的 `execution_order`，无需重新排序。

### 2.3 同层并行预留

虽然 v1 采用串行执行，但建议在排序结果上附加"层级（level）"标注，以便未来并行化：

```
function topological_sort_with_level(graph):
  // 在 BFS 的每一"轮"为同轮出队的节点标记相同 level
  level = 0
  order_with_level = []
  ...
  // 同一 level 的节点之间无依赖，可安全并行
  return order_with_level
```

---

## 3. 环检测与处理

### 3.1 何时可能出现环

- 数据库物理外键不允许循环引用（DB 引擎会拒绝建表），因此从物理 FK 生成的 `TableRelation` 通常不会产生环。
- 用户手动定义的**逻辑关系**（`is_logical = TRUE`）理论上可能形成环（例如：用户错误地将 A 和 B 互相设为对方的父表）。

### 3.2 环检测机制

Kahn 算法的自带环检测：

```
if len(order) < len(graph):
  // 未能排出所有节点 → 图中存在环
  cycle_nodes = [node_id for node_id in graph if node_id NOT IN order]
  raise CyclicDependencyError(cycle_nodes)
```

环检测需要在**两个时机**都触发：
1. **保存 Project 时**（拓扑排序阶段）：发现环则拒绝保存，向用户报告哪些表形成了循环依赖，引导用户修正 TableRelation 配置。
2. **预检阶段**（作为防御性校验）：理论上不应触发（因保存时已检查），但作为安全网保留，万一数据损坏时给出明确错误。

### 3.3 错误报告

环错误应尽可能定位出环的成员表，帮助用户快速定位问题：

```
// 找到环的一种方式：在环节点集合中 DFS 找出一条回路
function find_cycle_path(cycle_nodes, graph):
  // 从 cycle_nodes 中任意一个节点出发，沿依赖边 DFS
  // 首次访问到已在当前路径中的节点，即找到一条环路径
  ...
  return ["orders", "order_items", "orders"]  // 示例：orders → order_items → orders
```

---

## 4. 行数预计算（Preflight 阶段）

这是整个预检阶段的核心，也是最复杂的部分。行数计算需要完整覆盖数据模型 D-05 描述的 6 种状态。

### 4.1 前置概念：有效父记录池大小

在计算子表行数之前，需要先确定每个"父表方向"（即每条 `ProjectTableRelation`）的**有效父记录池大小**（`parent_pool_size`）：

| `rel_value_source` | `parent_project_table_id` | `parent_pool_size` 计算方式 |
|---|---|---|
| `FROM_EXECUTION` | NOT NULL | 父表的预估生成行数（`TablePlan.estimated_rows`） |
| `FROM_DB_QUERY` | NULL | 执行 `rel_source_sql` 的 `COUNT(*)` 结果 |
| `FROM_DB_QUERY` | NOT NULL | 执行 `rel_source_sql` 的 `COUNT(*)` 结果 |
| `MERGED` | NOT NULL | 父表预估行数 + `rel_source_sql` COUNT 结果 |
| `MERGED` | NULL | `rel_source_sql` COUNT 结果（行为同 `FROM_DB_QUERY`）|

对于 `FROM_DB_QUERY` 和 `MERGED` 模式，预检阶段需连接目标库执行 `rel_source_sql`。若 SQL 执行失败，则为阻塞性错误。

### 4.2 D-05 六种角色的行数决策矩阵

以下矩阵是对数据模型 D-05 的引擎级操作化，明确了每种情况下 `estimated_rows` 和 `estimated_rows_range` 的计算方式：

---

**情形 1：表为 Parent 或 BaseTable**

```
estimated_rows = ProjectTable.row_count   // 用户显式配置，直接使用
```

展示给用户：确定值，无范围。

---

**情形 2：Child，其 Parent 在本 Project 中**

```
ptr = 关联该 Parent 的 ProjectTableRelation
parent_rows = Parent 的 estimated_rows

// 期望值（用于展示）
expected = parent_rows × (ptr.multiplier_min + ptr.multiplier_max) / 2

// 范围
min_rows = parent_rows × ptr.multiplier_min
max_rows = parent_rows × ptr.multiplier_max

estimated_rows = round(expected)
estimated_rows_range = [min_rows, max_rows]
```

展示给用户：`约 expected 行（范围 min_rows ~ max_rows）`

---

**情形 3：Child，其 Parent 不在本 Project 中**

```
ptr = 关联该 Parent 的 ProjectTableRelation
parent_pool_size = 执行 rel_source_sql 的 COUNT(*) 结果

// 理论生成行数（由存量父记录数和倍数决定）
theoretical_min = parent_pool_size × ptr.multiplier_min
theoretical_max = parent_pool_size × ptr.multiplier_max

// 用户配置了 row_count 作为上限
cap = ProjectTable.row_count

estimated_rows = min(round((theoretical_min + theoretical_max) / 2), cap)
estimated_rows_range = [min(theoretical_min, cap), min(theoretical_max, cap)]
```

展示给用户：`约 estimated_rows 行（受 row_count 上限 cap 约束）`

若 `parent_pool_size = 0`：这是一个 Warning，告知用户 rel_source_sql 查无存量数据，该表将生成 0 行。

---

**情形 4：JoinTable，所有 BaseTable 都在本 Project 中**

设 JoinTable J 有 BaseTable B1 和 B2（两个入边方向），对应 `ptr1`（B1→J）和 `ptr2`（B2→J）：

```
// 每个 BaseTable 方向的贡献
b1_rows = B1.estimated_rows
b2_rows = B2.estimated_rows

// 从 B1 的视角：每条 B1 记录贡献 [ptr1.min, ptr1.max] 条 J 记录
// ptr1 的倍数含义：每个 B1 记录对应多少个不同 B2 记录被关联进 J
expected_from_b1 = b1_rows × (ptr1.multiplier_min + ptr1.multiplier_max) / 2

// 注意：J 的实际生成逻辑是：
// 遍历 B1 的每条记录，为其随机选取 k 个不同的 B2 记录，形成 k 条 J 记录
// 因此 b2_rows 决定了 k 的上限（不能选超过 B2 总记录数个不同记录）
// D-10 校验会处理 ptr1.multiplier_max > b2_rows 的情形

// 在 D-10 修正后
effective_max = min(ptr1.multiplier_max, b2_rows)   // 经 D-10 修正

estimated_rows = round(b1_rows × (ptr1.multiplier_min + effective_max) / 2)
estimated_rows_range = [
  b1_rows × ptr1.multiplier_min,
  b1_rows × effective_max
]
```

> **注意**：JoinTable 的行数由"遍历较多一侧的 BaseTable"驱动（通常是 B1，作为"主驱动方"），ptr1 定义了每个 B1 记录关联多少个 B2 的语义。ptr2 在执行时主要用于提供 B2 的 FK 值池，而非独立驱动行数。关于 JoinTable 生成的具体执行逻辑，见专题 5-3 §2.4。

---

**情形 5：JoinTable，所有 BaseTable 都不在本 Project 中**

```
// 两个方向都通过 rel_source_sql 获取存量记录数
b1_pool = COUNT(*) from ptr1.rel_source_sql   // B1 的存量 FK 值池
b2_pool = COUNT(*) from ptr2.rel_source_sql   // B2 的存量 FK 值池

// 行为类似情形 3，以 B1 为主驱动
effective_max = min(ptr1.multiplier_max, b2_pool)   // D-10 修正
expected = b1_pool × (ptr1.multiplier_min + effective_max) / 2
cap = ProjectTable.row_count

estimated_rows = min(round(expected), cap)
```

---

**情形 6：JoinTable，部分 BaseTable 在本 Project 中**

这是最复杂的混合情形。以 B1 在 Project 中、B2 不在为例：

```
b1_rows = B1.estimated_rows           // 来自本次生成
b2_pool = COUNT(*) from ptr2.rel_source_sql  // 来自数据库存量

effective_max = min(ptr1.multiplier_max, b2_pool)   // D-10 修正
estimated_rows = round(b1_rows × (ptr1.multiplier_min + effective_max) / 2)
estimated_rows_range = [b1_rows × ptr1.multiplier_min, b1_rows × effective_max]
```

### 4.3 行数计算的执行顺序

行数计算本身也有顺序依赖：子表的估算依赖父表的估算结果。因此，**行数计算应按拓扑顺序（execution_order）逐表进行**，确保父表的 `estimated_rows` 在子表计算时已经可用。

```
for each TablePlan tp in sorted_by_execution_order:
  tp.estimated_rows = calculate_rows(tp, parent_plans, rel_source_sql_results)
```

---

## 5. JoinTable 唯一性容量校验（D-10）

D-10 是对 JoinTable 生成逻辑的安全防护，防止因倍数配置过大导致复合唯一索引冲突。

### 5.1 触发条件

同时满足以下条件时触发：
1. 表的 `relation_type = JOIN_TABLE`。
2. 该表在 `TableConstraint` 中存在复合 UNIQUE 约束或复合 PK，且约束列恰好是该 JoinTable 两个（或多个）BaseTable 方向的 FK 列（即这个约束保证了每对 `(b1_fk, b2_fk)` 的唯一性）。

### 5.2 校验逻辑

```
for each JOIN_TABLE table J:
  for each ptr pointing to J (ptr.child = J):
    // ptr 对应的 "另一侧" BaseTable 的有效记录池大小
    other_base_pool_size = get_pool_size(ptr 对应的 BaseTable, fk_source)

    if ptr.multiplier_max > other_base_pool_size:
      // 触发 D-10 修正
      effective_max = other_base_pool_size
      emit Warning {
        table: J.table_name,
        message: "multiplier_max ({ptr.multiplier_max}) 超过 {other_base.table_name} 
                  的有效记录数（{other_base_pool_size}），已自动修正为 {effective_max}",
        original_max: ptr.multiplier_max,
        corrected_max: effective_max
      }
      // 注意：此处只修正运行时计算用的 effective_max，不写回数据库
      runtime_effective_max[ptr.id] = effective_max
```

### 5.3 D-10 的边界情况

**情形：`other_base_pool_size = 0`**

若 B2 的存量或生成记录数为 0，则 JoinTable 的任何 `multiplier_min > 0` 都无法满足。此时应提升为 **Warning（若 `multiplier_min = 0`）或 BlockingError（若 `multiplier_min > 0`）**：

```
if other_base_pool_size == 0:
  if ptr.multiplier_min > 0:
    raise BlockingError("JoinTable {J} 的 BaseTable {B2} 无有效记录，
                         但 multiplier_min = {ptr.multiplier_min} > 0，
                         无法生成任何 {J} 记录。")
  else:
    emit Warning("JoinTable {J} 的 BaseTable {B2} 无有效记录，
                  {J} 将生成 0 行。")
```

---

## 6. ExecutionPlan 输出与用户展示

预检完成后，向用户展示以下信息：

### 6.1 展示格式建议

```
执行计划预览
─────────────────────────────────────────────────────
执行顺序  表名                    预估行数     操作前清空
   1     customers               1,000        否
   2     orders                  5,000        否
   3     order_items          约 12,500        否
                               (范围: 10,000 ~ 15,000)
   4     order_payment_rel    约   4,800        否
                               (范围:  4,500 ~  5,000)
─────────────────────────────────────────────────────
总计：约 23,300 行

⚠ 警告 (1)
  - order_payment_rel: multiplier_max 已从 5 自动修正为 3
    （payments 的有效记录数为 3，低于配置的倍数上限 5）

✗ 错误 (0)
─────────────────────────────────────────────────────
[确认执行]  [取消]
```

### 6.2 若存在阻塞性错误

```
执行计划预览（存在阻塞性错误，无法执行）
─────────────────────────────────────────────────────
✗ 错误 (2)
  - order_items.product_id: 字段非空且无默认值，但缺少 GeneratorConfig
  - promotions: rel_source_sql 执行失败（SQL 语法错误）

请修复以上问题后重新执行。
─────────────────────────────────────────────────────
[关闭]
```

---

## 7. 执行顺序预计算的触发时机

根据数据模型 D-08，拓扑排序在**保存 Project 配置时**触发，不在执行时重新计算。

### 7.1 触发保存拓扑排序的操作

- 用户在 Project 中添加一张新表（`ProjectTable` 新增）
- 用户在 Project 中删除一张表（`ProjectTable` 删除）
- 用户修改 `ProjectTableRelation`（增加/删除/修改关系）

任何以上操作保存后，系统：
1. 对当前 Project 的 `ProjectTable` 和有效 `ProjectTableRelation` 重新执行 Kahn 算法。
2. 若检测到环：拒绝保存，提示用户修正。
3. 若无环：更新所有 `ProjectTable.execution_order`。

### 7.2 不触发重新排序的操作

- 修改 `ProjectTable.row_count`（不影响依赖结构）
- 修改 `ProjectTable.truncate_before`（同上）
- 修改 `ProjectTableRelation.multiplier_min / multiplier_max`（不影响拓扑，只影响行数）
- 修改 `ProjectTableRelation.rel_value_source` 或 `rel_source_sql`（不影响拓扑）

### 7.3 Schema 结构变更后的处理

当目标库 Schema 重新扫描后，若某张已在 Project 中的表被删除，对应的 `ProjectTable` 应标记为"悬空"状态，并提示用户。此类情况在执行预检时也会被捕获（ConfigLoader 阶段报错）。

---

*下一篇：专题 5-3 — 生成与写入引擎*
