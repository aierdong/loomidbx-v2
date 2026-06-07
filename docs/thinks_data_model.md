# 关于数据模型的思考

> 请阅读产品大纲、用户故事，帮我生成数据模型。你可以先试着列举一份大纲，确定无误后再开始设计

The user wants me to read the product outline, user story map, and user stories for LoomiDBX, then help generate a data model. They want me to first create an outline/overview of the data model, confirm it's correct, and then design the full data model.

Let me read through all three documents carefully to understand the product.

**Product Overview**:

LoomiDBX is a desktop database simulation/synthetic data generation tool that:

- Connects to databases
- Identifies database schemas
- Allows users to configure field generation rules
- Fills generated data back into target databases

**Key Modules**:

1. **Schema Module** - understand databases, define field rules
2. **Project Module** - organize generation tasks

**Key Concepts**:

1. **Connection** - database connection info
2. **Database structure** - databases, schemas, tables, fields, relationships, constraints
3. **Field generation rules** - random, distributed, enum, external data, computed, foreign key reference
4. **Project** - a collection of tables for one generation task
5. **Execution task** - running a project to generate data
6. **History records** - results of execution

**User Stories give more details**:

For connections:

- Connection name, database type, address, port, username, password, optional params (ssl, connect_timeout)
- Support for multiple connections

For schema/structure:

- Tree: database > schema > table > field
- Field attributes: sequence, name, data type, nullable, default value, primary key, comment, generator name
- Constraints: constraint name, fields, type (Primary, Unique)
- Foreign keys: FK name, FK field, referenced table, referenced field
- DDL: CREATE TABLE statement

For field rules:

- Generation method: random value, distributed value, enum value, computed value, foreign key reference
- Rule parameters per method
- Preview (5 sample rows)

For Projects:

- Project name, description
- Tables selected for generation

Good, I've already read all three documents earlier. Now let me write the outline for the data model.
Let me structure this cleanly:

- 用户与认证 (1 entity)
- 连接管理 (1 entity)
- Schema 层 / 扫描结构 (8 entities)
- 字段生成规则 (1 entity)
- Project 层 (3 entities)
- 执行与历史 (2 entities)
- 系统配置 (1 entity)

Total: 17 entities
Let me write the outline now. I'll keep it concise but structured.


好的，已通读三份文档。以下是数据模型的实体大纲，请确认后再展开完整设计。

---

## LoomiDBX 数据模型大纲

共 **7 个分组，17 个实体**（含 2 个字段关联实体）。

---

### 第一组：用户与认证

| 实体 | 说明 |
|---|---|
| `User` | 用户账号，记录注册信息和登录状态，用于产品授权验证（6.8）|

---

### 第二组：连接管理

| 实体 | 说明 |
|---|---|
| `Connection` | 数据库连接配置，含连接名、DB 类型、主机/端口、账号/密码（加密）、扩展参数（SSL、超时等）及最近验证状态 |

---

### 第三组：扫描结构（Schema 层）

| 实体 | 说明 |
|---|---|
| `DbCatalog` | 连接下的"数据库"，对应 MySQL Database / PG Database / Oracle 实例等 |
| `DbSchema` | Catalog 下的 Schema 层（PG/MSSQL 有此层；MySQL 等无 Schema 的数据库，自动生成一条隐式记录，保持结构统一）|
| `DbTable` | 表，挂在 Schema 下，含表名、注释、DDL 快照、最近扫描时间 |
| `DbColumn` | 字段，含序号、名称、数据类型、是否主键、是否可空、默认值、注释 |
| `TableConstraint` | 表级约束，含约束名和约束类型（PRIMARY / UNIQUE）|
| `ConstraintColumn` | 约束与字段的映射（一个约束可含多字段，记录顺序）|
| `ForeignKey` | 外键定义，含外键名、来源表、目标表 |
| `ForeignKeyColumn` | 外键中的字段对应关系（来源字段 → 目标字段）|

> `DbColumn.is_primary_key` 可由 `TableConstraint` 派生，保留冗余字段是为了便于字段列表直接渲染，无需联查。

---

### 第四组：字段生成规则

| 实体 | 说明 |
|---|---|
| `GeneratorConfig` | 字段与生成器的绑定，含生成器类型（枚举）、规则参数（JSON）、配置状态 |

**生成器类型枚举：**

| 类型 | 说明 |
|---|---|
| `RANDOM` | 随机值（范围、长度、格式等）|
| `DISTRIBUTED` | 分布值（正态、均匀等分布类型及参数）|
| `ENUM` | 枚举值（候选值列表 + 权重）|
| `EXTERNAL` | 外部数据库（如 Faker，人名/手机/地址等语义类型）|
| `COMPUTED` | 计算值（依赖其他字段的表达式）|
| `FK_REFERENCE` | 外键关联值（从被引用表中取值）|
| `NULL_VALUE` | 不生成（仅允许可空或有默认值的字段）|

**配置状态枚举：** `ACTIVE`（有效）/ `NEEDS_REVIEW`（结构变更后待检查）/ `INVALID`（字段已不存在）

---

### 第五组：Project（任务组织）

| 实体 | 说明 |
|---|---|
| `Project` | 生成任务容器，关联一个 Connection，含名称和描述 |
| `ProjectTable` | Project 中每张表的执行配置，含生成行数、是否清空（truncate）、执行顺序 |
| `ProjectTableRelation` | ProjectTable 间的数量倍数关系（主表 → 子表，记录倍数；实际行数在执行时计算）|

---

### 第六组：执行与历史

| 实体 | 说明 |
|---|---|
| `ExecutionTask` | 一次生成执行的主记录，含任务名称、所属 Project、开始/结束时间、整体状态 |
| `ExecutionTableResult` | 执行中每张表的结果，含表名快照（防止表被删后历史记录丢失）、写入行数、执行状态、错误信息 |

**执行状态枚举：** `RUNNING` / `SUCCESS` / `PARTIAL_FAILED` / `FAILED`

---

### 第七组：系统配置

| 实体 | 说明 |
|---|---|
| `AppSetting` | 键值对形式的系统配置，按分组管理（存储路径、LLM 配置、UI 语言/主题）；可实现为本地配置文件或 kv 表 |

---

## 核心关系一览

```
Connection
  └─1:N─> DbCatalog
            └─1:N─> DbSchema
                      └─1:N─> DbTable
                                ├─1:N─> DbColumn
                                │         └─0..1─> GeneratorConfig
                                ├─1:N─> TableConstraint ─N:M─> DbColumn (via ConstraintColumn)
                                └─1:N─> ForeignKey ─N:M─> DbColumn (via ForeignKeyColumn)

Connection
  └─1:N─> Project
            ├─1:N─> ProjectTable ──N:1──> DbTable
            │         └─N:M─> ProjectTable (via ProjectTableRelation)
            └─1:N─> ExecutionTask
                      └─1:N─> ExecutionTableResult ──N:1──> DbTable (冗余 table_name 快照)
```

---

## 几个设计决策，欢迎讨论

1. **DbCatalog / DbSchema 两层抽象**：统一兼容 MySQL（无 Schema 概念）、PostgreSQL（有 Schema）、Oracle（Schema = 用户）等不同层级差异，避免每种数据库单独处理树形结构。

2. **GeneratorConfig 与 DbColumn 1:0..1**：规则归属 Schema 层而非 Project 层（呼应产品规则"字段生成规则与表执行规则分离"），规则跨 Project 复用。

3. **ProjectTableRelation 只存倍数**：不预存计算后的行数，实际生成数量在执行时由引擎根据主表行数 × 倍数算出，保持 ProjectTable.row_count 为直接配置量。

4. **ExecutionTableResult 存 table_name 快照**：表被删或重命名后，历史记录仍保持可读。

5. **DbColumn.is_primary_key 冗余**：虽可从 TableConstraint 推导，但字段列表渲染高频读取，保留此字段避免连表查询。

---

以上 17 个实体是否完整？有以下问题想确认后再开始详细设计：

1. `User` 的授权模型你有预期方案吗（本地单用户 / 在线账号验证 / License Key）？会影响 `User` 的字段设计。
2. `ProjectTableRelation` 目前只设计了"父子倍数关系"，是否还需要支持"固定行数覆盖"（即子表独立指定行数，不依赖主表）？
3. `AppSetting` 是否需要纳入 ER 图，或者独立作为配置文件管理即可？

---

澄清：

> 1. 不需要 User。注册/登录与软件本身无关，它是向远端服务器（我）获取授权的方式，数据存储由远端服务负责。
> 2. Connection：验证状态是不需要的，它不能表示连接是否仍然有效。
> 3. 关于 GeneratorConfig，我觉得也许不必单独一个实体，它可以依附在 DbColumn 上，它包含如下几个字段：生成器名称（distribute_int、enums、email、dict_table等）、数据映射类型（枚举：text/integer/float/boolean/datetime）、规则参数。你认为 GeneratorConfig 需要单独一个实体的考虑是性能还是其它方面的考虑？此外"配置状态" 是指的什么？生成器类型与我理解的生成器名称是一回事吗？
> 4. 对于 ProjectTableRelation 的疑问：ProjectTable 间的数量倍数关系是在执行前动态计算的，是否可以将计算出的子表数量直接记录在 ProjectTable 中？
> 5. 你提到的“ProjectTableRelation 需要支持"固定行数覆盖"”，这个提示很赞！它意味着子表的生成与主表解耦。子表关联的主表ID不依赖本次生成，可以直接从表中用 SELECT ID from '主表' 读取，然后与本次生成的主表ID合并。例如我需要生成 100 条订单明细时，需要同时询问用户取多少个订单主表ID（或者由用户提供SQL，例如只取最近3条的订单）
> 6. AppSetting 不需要纳入数据模型，我将使用配置文件持久化它


---

分析反馈并重构数据模型设计方案
The user is giving me feedback on my proposed data model outline. Let me carefully analyze each point:

1. No User entity - The registration/login is handled by a remote server, so we don't need a User table in the local data model. Got it.
1. Connection - no validation status - The validation status field should be removed from Connection, as it can't reliably represent whether a connection is still valid. Makes sense.  
1. GeneratorConfig as embedded fields on DbColumn vs separate entity - The user is questioning whether GeneratorConfig needs to be a separate entity. They suggest it could just be fields on DbColumn:
  - Generator name (like distribute_int, enums, email, dict_table - these seem to be actual generator identifiers, not abstract types)
  - Data mapping type (enum: text/integer/float/boolean/datetime)
  - Rule parameters (JSON) 

The user asks:

- Why did I think GeneratorConfig needs to be a separate entity? (Performance or other reasons?)
- What does "配置状态" (config status) mean?
- Is my "生成器类型" (generator type) the same as their "生成器名称" (generator name)?

Let me think about this:

- **Why separate entity**? My original reasoning was probably for cleaner normalization and to track the config state separately. But actually, since GeneratorConfig is 1:0..1 with DbColumn (one column has at most one generator config), embedding it directly in DbColumn is fine. The only case where a separate entity would be justified is if:

  - We need versioning of configs (not mentioned)
  - Multiple configs per column (not the case)
  - The config needs to be shared across columns (not mentioned)

So the user is right - it can just be nullable fields on DbColumn.
- **配置状态 (config status):** I was thinking of states like ACTIVE/NEEDS_REVIEW/INVALID to track whether a config is still valid after schema changes. This would help with US-008 (查看结构变化影响范围). I should clarify this.

---

非常好的反馈，逐条回应如下：

---

## 逐条确认

**1. 去掉 User 实体** ✅ 确认移除。授权由远端服务负责，本地数据模型无需存储用户数据。

---

**2. Connection 去掉验证状态** ✅ 同意。连接验证是一次性行为，结果无法代表持续有效性，存下来反而会产生误导。

---

**3. GeneratorConfig：单独实体 vs 嵌入 DbColumn**

我当时设计为单独实体的原因有两个：

- **关注点分离**：DbColumn 的字段是"从数据库扫描来的元数据"，GeneratorConfig 是"用户手动配置的生成规则"，两者的写入时机和来源完全不同，放在同一张表里会混淆职责。
- **追踪配置状态**：独立实体可以在不删除 DbColumn 记录的前提下标记"这条规则失效了"——但如果列被删除时直接级联删除 DbColumn，这个优势就不存在了。

不过，你说得对——**1:0..1 的关系没有必要拆成两张表**。嵌入 DbColumn 同样可以实现，代价只是字段职责混合一些，但更简单。

**建议方案：仍然作为独立实体，但原因只剩一个**——结构扫描与规则配置的写入路径不同（扫描时更新 DbColumn，配置时写入 GeneratorConfig），分开更清晰。但如果你更倾向于简单，嵌入也完全可行。

关于你的三个追问：

**① 配置状态是什么**：这是我为"结构变更后规则可能失效"这个场景预留的字段。例如：
- 列从 `INT` 改为 `VARCHAR` → 原来配置的 `distribute_int` 可能不再适用 → 标记为 `NEEDS_REVIEW`
- 列被删除 → 若用独立实体，原规则记录孤悬 → 标记为 `INVALID`（嵌入则随列一起删除，此状态无意义）

如果采用嵌入方式，`INVALID` 状态确实不需要，只剩 `ACTIVE` 和 `NEEDS_REVIEW` 两个值。

**② 生成器类型 vs 生成器名称**：不是同一回事。我的"生成器类型"是我自行设计的抽象分组（RANDOM/ENUM/EXTERNAL…），你的"生成器名称"（`distribute_int`、`email`、`enums`、`dict_table`…）才是真正要存储的具体实现标识符。**以你的生成器名称为准**，我的抽象分类层可以删除，分组逻辑交给前端枚举或代码配置。

---

**4. ProjectTableRelation 的子表行数：预计算存入 ProjectTable vs 执行时计算**

两种方式的差别：

| | 预计算存入 ProjectTable | 执行时动态计算 |
|---|---|---|
| 优点 | 执行逻辑简单，行数已知 | 主表行数修改后自动同步 |
| 缺点 | 主表行数变化后需手动同步 | 执行前不直观显示最终数量 |

我认为**可以两者共存**：ProjectTableRelation 存倍数，ProjectTable 存用户最终确认的行数（可由倍数自动填入，也可手动覆盖）。倍数是"来源"，行数是"结果"，两者不互斥。这样执行引擎只读 ProjectTable.row_count，简单清晰。

---

**5. FK 来源策略：这是个重要的新实体字段**

你的描述非常清晰，这意味着 ProjectTableRelation 需要扩展一个 **FK 值来源策略**：

| 策略 | 含义 |
|---|---|
| `FROM_EXECUTION` | FK 值只取本次生成的主表 ID（默认行为） |
| `FROM_DB_QUERY` | FK 值从数据库中查询已有记录（用户提供 SQL） |
| `MERGED` | 合并以上两种来源 |

当策略为 `FROM_DB_QUERY` 或 `MERGED` 时，附加一个 `fk_source_sql` 字段（如 `SELECT id FROM orders ORDER BY created_at DESC LIMIT 3`）。

这也意味着：**当主表不参与本次 Project 时**，ProjectTableRelation 的 `parent_project_table_id` 可以为空，此时策略只能是 `FROM_DB_QUERY`。

---

**6. AppSetting 不纳入模型** ✅ 使用配置文件即可。

---

## 修订后的实体大纲

基于以上讨论，调整如下：

### 变化点

| 原设计 | 调整后 |
|---|---|
| 包含 `User` | 移除 |
| `Connection` 含验证状态 | 移除验证状态字段 |
| `GeneratorConfig` 独立实体 | **待你确认**：独立实体 或 嵌入 DbColumn |
| 生成器用抽象类型枚举 | 改为存具体生成器名称字符串 |
| `ProjectTableRelation` 只存倍数 | 新增 `fk_value_source`（来源策略枚举）和 `fk_source_sql` |
| `ProjectTable` 只存用户填写行数 | 新增：可由倍数自动计算，用户也可手动覆盖 |
| `AppSetting` 实体 | 移除，改为配置文件 |

修订后共 **15–16 个实体**（含或不含独立 GeneratorConfig），请确认后我开始输出完整数据模型文档。

---

还有一个问题想在开始前确认：**GeneratorConfig 是否使用独立实体？** 你可以从两个角度选：

- **独立实体**：写入路径分离更清晰，字段职责更单一，但多一张表；
- **嵌入 DbColumn**：更简单，少一次关联查询，字段命名加前缀 `generator_` 区分即可。

---
> 补充：
> 
> 1. GeneratorConfig，我同意你的设计：使用独立实体
> 2. 关于 ProjectTable 子表行数问题：子表行数可能是一个动态值，因为倍数就是一个动态值。例如：订单主表与订单明细子表，他们的倍数关系是 1 至 5，意味着一个主表记录对应 1 - 5 条子表记录。如果要在执行前直观显示最终数量，则需要一个 map 记录每个主表记录对应多少条子表记录。嗯，这样也行。如果需要动态显示生成进度，这可能是必须的。但最终行数只能由计算决定，不能手动覆盖（可能造成主表记录找不到对应子表记录）。
> 3. ProjectTable的说明：Project 中每张表的执行配置，含生成行数或倍数范围（如 0-5）、执行前是否清空（truncate）、执行顺序
> 如果你认为我的这些思考没有问题，请开始输出完整数据模型文档
---

### 规划了完整数据模型文档的结构和内容。

The user wants me to finalize and output the complete data model document. Let me review all the feedback and confirmed decisions:

- No User entity - authorization is handled by remote server
- Connection - remove validation status field
- GeneratorConfig - independent entity (not embedded in DbColumn)
- Generator naming - use concrete generator name strings (not abstract type enums)
- ProjectTable - contains row_count OR multiplier_range (like "1-5"), truncate flag, execution order
- ProjectTableRelation - stores multiplier range, FK value source strategy, fk_source_sql
- Child table row count - dynamic, cannot be manually overridden; computed at execution time and tracked per parent record in execution results
- AppSetting - not in data model, use config file
- No User entity

Let me think through the complete data model:

#### Group 1: Connection Management

**Connection**

- id (PK)
- name (connection name)
- db_type (enum: MySQL/PostgreSQL/Oracle/MSSQL/SQLite/ClickHouse/Hive)
- host
- port
- database_name (optional, some DBs need this)
- username
- password (encrypted)
- extra_params (JSON: SSL mode, connect_timeout, etc.)
- created_at
- updated_at

#### Group 2: Schema Structure (scanned from DB)

**DbCatalog**

- id (PK)
- connection_id (FK → Connection)
- catalog_name
- scanned_at
- created_at
- updated_at

**DbSchema**

- id (PK)
- catalog_id (FK → DbCatalog)
- schema_name
- scanned_at
- created_at
- updated_at

**DbTable**

- id (PK)
- schema_id (FK → DbSchema)
- table_name
- comment
- ddl_snapshot (TEXT)
- scanned_at
- created_at
- updated_at

**DbColumn**

- id (PK)
- table_id (FK → DbTable)
- ordinal_position
- column_name
- data_type (raw DB type string)
- is_nullable (bool)
- default_value (nullable)
- is_primary_key (bool, denormalized for convenience)
- comment
- created_at
- updated_at

**TableConstraint**

- id (PK)
- table_id (FK → DbTable)
- constraint_name
- constraint_type (enum: PRIMARY / UNIQUE) …

Show more

读取 skills 目录确认有无相关技能

No specific skill for markdown files. I'll write the markdown file directly.

Let me now write the complete data model document. I'll be comprehensive and thorough.

Let me think about the complete structure one more time before writing:

- Title & Scope
- Entity Overview Table
- Each entity group with field tables
- Mermaid ERD
