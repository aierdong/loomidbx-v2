# 计算字段设计逻辑

本文说明 `sql_expression` 与 `python_expression` 两类计算字段的设计。它们都用于在同一行内根据已有值计算目标字段，计算发生在普通字段、外键字段等基础数据生成完成之后。

---

## 一、问题范围

计算字段需要解决以下问题：

- 一个表内可能存在多个计算字段。
- 计算字段可以引用同表内其他字段的值。
- 计算字段可以引用运行期上下文、项目变量等非字段值。
- 计算字段 A 未来可能引用计算字段 B 的值。
- 当一条生成任务中包含多个计算字段时，需要确定字段求值顺序。
- SQL 与 Python 表达式都需要做合法性、安全性、性能和类型兼容性处理。

MVP 阶段只实现“计算字段引用普通字段、外键字段和受控上下文值”。计算字段互相引用只作为后续扩展预留，不在 MVP 中实现。

---

## 二、计算时机

同一张表的一行数据按以下阶段生成：

1. 生成普通字段，例如随机数、枚举、时间、地址等。
2. 生成外键字段，从已生成父表的 ID 池或引用值池中取值。
3. 生成计算字段，包括 `sql_expression` 与 `python_expression`。
4. 执行类型转换、唯一性检查、非空检查和写入前校验。

计算字段不参与表级依赖拓扑。跨表依赖仍由外键和表级 Planner 解决；计算字段只处理同一表、同一行内的值依赖。

---

## 三、引用语法

### 3.1 MVP 推荐语法

表达式中统一使用占位符引用外部值：

```text
${column_name}
${field.column_name}
${meta.row_index}
${meta.batch_index}
${vars.project_code}
```

含义如下：

| 写法 | 含义 |
|------|------|
| `${column_name}` | 同表字段的简写，等价于 `${field.column_name}` |
| `${field.column_name}` | 显式引用同表字段 |
| `${meta.row_index}` | 当前表内行序号，从 0 开始 |
| `${meta.batch_index}` | 当前批次序号，从 0 开始 |
| `${vars.name}` | Project 或任务级变量 |

MVP 阶段推荐保留 `${column_name}`，降低用户使用门槛；内部解析时统一转换成 `field.column_name`。

### 3.2 字段名歧义处理

字段名可能包含空格、点号、特殊字符或数据库保留字。为避免歧义，支持转义形式：

```text
${field["order.id"]}
${field["user name"]}
```

MVP 可先支持普通字段名和 `${field.name}`，将方括号转义作为后续增强。若字段名无法被普通占位符表示，UI 应提示用户使用可视化字段选择器插入引用，避免手写错误。

### 3.3 非字段值

非字段值不应混在字段命名空间中，必须通过命名空间区分：

- `meta`：系统运行期上下文，如行序号、批次序号、生成时间。
- `vars`：用户在 Project 或任务中定义的变量。

MVP 阶段只建议内置少量稳定值：`meta.row_index`、`meta.batch_index`、`meta.generated_at`。`vars` 可以先作为配置结构预留，等 Project 变量设计明确后再开放。

---

## 四、引用解析方案

### 4.1 MVP 方案：占位符扫描

MVP 使用占位符扫描，不直接解析 SQL 或 Python 内部语法来发现字段引用。

流程：

1. 扫描表达式中的 `${...}` 片段。
2. 将片段解析为 `field`、`meta`、`vars` 三类引用。
3. 校验引用对象是否存在、是否允许访问。
4. 将表达式编译为运行期可执行模板。

该方案实现简单，行为可解释，也能避免把 SQL/Python 的语法差异提前引入依赖分析。

### 4.2 词法分析方案

后续可以引入轻量词法分析器，把表达式拆成字符串、标识符、操作符、占位符等 token。好处是可以避免误识别字符串字面量中的 `${...}`，也能提供更准确的错误位置。

MVP 不实现完整词法分析，只要求占位符扫描器正确处理转义和未闭合占位符。

### 4.3 解析器方案

后续可以分别引入 SQL Parser 与 Python AST Parser：

- SQL Parser 用于判断表达式是否为单个标量表达式，不包含 DDL/DML、多语句、子查询等危险结构。
- Python AST Parser 用于限制语法节点，例如只允许表达式，不允许赋值、循环、函数定义、导入等语句。

MVP 对 Python 建议直接使用 AST 白名单；对 SQL 则采用“包裹成 SELECT 后由目标数据库或方言解析器验证”的方式。

---

## 五、表内求值顺序

### 5.1 字段级 DAG

表内依赖图的节点是字段，边表示“字段 A 的表达式引用了字段 B”，即 A 依赖 B，B 必须先有值。

构建流程：

1. 收集当前表所有字段配置。
2. 将每个字段加入节点集合。
3. 对每个计算字段解析占位符引用。
4. 若引用普通字段或外键字段，记录为已满足依赖。
5. 若引用另一个计算字段，记录计算字段之间的边。
6. 对计算字段子图做拓扑排序。
7. 若存在环，拒绝执行并提示环路。

### 5.2 MVP 取舍

MVP 不允许计算字段引用另一个计算字段。因此 DAG 可以退化为校验规则：

- 引用普通字段、外键字段：允许。
- 引用不存在字段：拒绝保存。
- 引用 `sql_expression` 或 `python_expression` 字段：拒绝保存，并提示“计算字段互相引用将在后续版本支持”。

虽然 MVP 不执行计算字段互引，但仍建议在配置模型中保留 `dependencies` 字段，用于记录解析出的引用列表，方便后续升级。

### 5.3 后续互引方案

后续允许计算字段互引时，SQL 与 Python 计算字段进入同一张字段级 DAG，不区分“先 SQL 后 Python”。拓扑排序结果才是唯一求值顺序。

示例：

```text
full_name = python_expression("${first_name} + ' ' + ${last_name}")
email     = python_expression("${full_name}.lower().replace(' ', '.') + '@demo.com'")
label     = sql_expression("CONCAT(${email}, '#', ${meta.row_index})")
```

求值顺序为：

```text
first_name, last_name, meta.row_index -> full_name -> email -> label
```

### 5.4 引用环问题

若后续开放计算字段互引，必须检测以下环：

```text
A -> B -> A
A -> B -> C -> A
A -> A
```

处理策略：

- 配置保存时尽量提示。
- 执行计划构建时强制检测。
- 发现环时中止执行，列出环路字段名。

MVP 只在文案和错误模型中预留，不深入实现环提取算法。

---

## 六、`sql_expression` 设计

### 6.1 表达式形态

`sql_expression` 只允许用户输入“标量 SQL 表达式”，不允许输入完整 SQL 语句。

允许示例：

```sql
CONCAT('ORD-', DATE_FORMAT(${order_time}, '%Y%m%d'), LPAD(${seq}, 4, '0'))
${price} * ${quantity}
CASE WHEN ${amount} > 1000 THEN 'VIP' ELSE 'NORMAL' END
```

不允许示例：

```sql
SELECT * FROM users
UPDATE users SET name = 'x'
${a}; DROP TABLE users
```

### 6.2 SQL 方言

SQL 表达式必须绑定目标数据库方言。配置中建议增加：

```json
{
  "expression": "CONCAT(${first_name}, ' ', ${last_name})",
  "dialect": "mysql",
  "validation_mode": "database_parse"
}
```

`dialect` 默认取当前连接的数据库类型。用户通常不需要手动选择。

### 6.3 合法性验证

多 SQL 方言下，合法性验证分为三层：

| 层级 | 验证内容 | MVP 决策 |
|------|----------|----------|
| 占位符层 | 引用是否存在、是否越权、是否引用计算字段 | 必做 |
| 结构层 | 是否为单个标量表达式，是否包含多语句、DDL、DML | 必做 |
| 方言层 | 函数名、类型转换、日期格式等是否符合目标数据库 | 执行前用目标数据库试解析 |

MVP 推荐使用目标数据库做试解析：将表达式包裹为 `SELECT <expr>`，把字段引用替换成参数或字面量样例，然后执行只读验证。

示例：

```sql
SELECT CONCAT(?, '-', ?) AS value
```

不同数据库的占位符形式由 Connector 适配。验证只允许在当前连接上执行只读 SELECT，不允许拼接用户输入为多语句。

### 6.4 执行方式

MVP 推荐 SQL 表达式按批执行，而不是逐行请求数据库。

基本思路：

1. 对每一批待写入行，准备一个只读计算 SQL。
2. 将每行被引用字段作为参数传入。
3. 数据库返回每行计算结果。
4. 引擎把结果写回行对象，再进入写入阶段。

若目标数据库难以构造批量表达式，MVP 可以降级为逐行计算，但必须设置批次超时和最大行数提示。性能敏感场景应优先建议用户使用 Python 表达式或基础生成器。

### 6.5 安全规则

`sql_expression` 的安全边界：

- 只允许单个标量表达式。
- 禁止分号和多语句。
- 禁止 DDL、DML、事务控制、权限控制语句。
- 禁止子查询和访问数据库对象。
- 字段值必须通过参数绑定传入，不允许字符串拼接。
- 数据库连接应使用当前用户已有权限，不额外提升权限。
- 验证与执行都应设置超时。

---

## 七、`python_expression` 设计

### 7.1 表达式形态

`python_expression` 只允许单个 Python 表达式，不允许完整 Python 脚本。

允许示例：

```python
${name}.upper() + '_' + str(${id})
round(${price} * ${quantity}, 2)
f"{${province}}-{${city}}"
```

不允许示例：

```python
import os
open('/etc/passwd').read()
while True: pass
```

### 7.2 合法性验证

MVP 推荐使用 Python AST 白名单校验：

1. 将占位符转换为内部变量名。
2. 调用 `ast.parse(expression, mode="eval")`。
3. 只允许表达式节点和安全操作节点。
4. 禁止 `Import`、`Call` 到未授权函数、属性访问到危险对象、下划线开头属性等。
5. 用样例上下文试运行一次，检查是否抛异常。

允许的 AST 节点可包含：

- 常量、变量名、列表、元组、字典。
- 一元/二元运算、布尔运算、比较运算。
- 条件表达式。
- 安全函数调用。
- 对普通字符串、数字、日期对象的受控方法调用。

### 7.3 标准库与外部库

MVP 只开放白名单内置函数和少量标准库能力，不开放任意标准库导入。

建议初始白名单：

| 类型 | 允许内容 |
|------|----------|
| 内置函数 | `str`、`int`、`float`、`bool`、`len`、`round`、`abs`、`min`、`max`、`sum` |
| 字符串方法 | `upper`、`lower`、`title`、`strip`、`replace`、`startswith`、`endswith`、`split`、`join` |
| 数学 | `math.floor`、`math.ceil`、`math.sqrt` |
| 日期 | `datetime.date`、`datetime.datetime`、`datetime.timedelta` 的格式化与简单运算 |

不建议在 MVP 开放 pandas、numpy、faker 等外部库。原因是：

- 打包体积增加。
- 冷启动和导入成本增加。
- 安全边界更难解释。
- 跨平台桌面端分发复杂度上升。

后续可以以“受控扩展包”的方式开放外部库，每个库必须有版本锁定、函数白名单和资源限制。

### 7.4 执行环境

Python 表达式应在受限执行环境中运行：

- `globals` 只包含白名单函数和模块。
- `locals` 只包含当前行字段值、`meta`、`vars`。
- 禁止访问 `__builtins__` 全量对象。
- 禁止 `eval`、`exec`、`compile`、`open`、`input`、`getattr`、`setattr`、`delattr`、`__import__`。
- 禁止下划线开头的属性访问。
- 每次执行设置超时或在独立 worker 中执行。

如果主程序为 Go，MVP 可选两种落地方式：

| 方案 | 优点 | 缺点 |
|------|------|------|
| 嵌入 Python 解释器 | 表达式兼容真实 Python | 打包复杂，安全隔离成本高 |
| 独立 Python Worker | 隔离性更好，可设置进程级超时 | 进程通信和部署更复杂 |

MVP 推荐独立 Python Worker，并通过 JSON Lines 或本地 RPC 与 Go 引擎通信。Worker 启动后复用，不为每行启动新进程。

### 7.5 安全规则

`python_expression` 的安全边界：

- 只允许 `eval` 模式表达式，不允许脚本。
- AST 白名单必须先于执行。
- 函数、模块、方法全部白名单化。
- 禁止文件、网络、系统命令、动态导入、反射和私有属性访问。
- 设置单表达式超时、批次超时和最大错误次数。
- Worker 崩溃时中止当前任务，并记录表达式字段名和错误摘要。

---

## 八、类型与约束处理

计算结果必须与目标列类型兼容。处理顺序如下：

1. 表达式返回原始值。
2. 引擎按目标列类型做标准转换。
3. 若配置了 `type_format` 或 `stringify`，按生成器基础规则处理。
4. 执行 `null_percentage`、`unique`、NOT NULL 等约束处理。
5. 写入前再由数据库约束兜底。

类型转换失败时：

- 预览阶段：显示样例错误。
- 保存阶段：若可以静态判断，拒绝保存。
- 执行阶段：中止当前表或当前任务，记录字段名、表达式和错误摘要。

---

## 九、性能考虑

计算字段会引入额外依赖和运行成本，尤其是 SQL 表达式的数据库往返和 Python Worker 的进程通信。

MVP 性能策略：

- 表达式在执行前编译一次，按字段缓存。
- 占位符依赖解析结果缓存到配置或执行计划中。
- Python Worker 在任务开始时启动，在任务结束时关闭。
- SQL 表达式尽量按批计算，避免逐行数据库请求。
- 对每个表达式记录执行耗时，用于运行历史和后续优化。
- 预览时只计算少量样例，不触发全量数据计算。

需要引入外部依赖时，应满足以下原则：

- 版本锁定。
- 可跨平台打包。
- 冷启动成本可接受。
- 可以禁用网络、文件和系统访问。
- 失败时错误可解释。

---

## 十、配置结构建议

### 10.1 `sql_expression`

```json
{
  "expression": "CONCAT('ORD-', DATE_FORMAT(${order_time}, '%Y%m%d'), LPAD(${seq}, 4, '0'))",
  "dialect": "mysql",
  "validation_mode": "database_parse",
  "dependencies": ["field.order_time", "field.seq"],
  "null_percentage": 0,
  "unique": false
}
```

### 10.2 `python_expression`

```json
{
  "expression": "${name}.upper() + '_' + str(${id})",
  "runtime": "python_worker",
  "allowed_profile": "mvp_safe",
  "dependencies": ["field.name", "field.id"],
  "null_percentage": 0,
  "unique": false
}
```

`dependencies` 可由系统自动生成，不要求用户手写。保存配置时若表达式变更，应重新解析并覆盖。

---

## 十一、错误提示

| 场景 | 提示方式 |
|------|----------|
| 引用字段不存在 | `字段 total_price 的表达式引用了不存在的字段 price2` |
| 引用计算字段 | `字段 email 引用了计算字段 full_name。当前版本暂不支持计算字段互相引用` |
| 存在引用环 | `计算字段存在循环依赖：A -> B -> A` |
| SQL 方言不支持 | `当前 MySQL 连接无法解析该 SQL 表达式：函数 DATE_TRUNC 不存在或语法不兼容` |
| Python 语法错误 | `Python 表达式语法错误：第 1 行第 8 列附近无法解析` |
| Python 安全限制 | `表达式使用了未允许的函数 open` |
| 执行超时 | `字段 label 的表达式计算超时，请简化表达式或减少外部依赖` |

---

## 十二、最终决策汇总

| 决策点 | 最终决策 |
|--------|----------|
| 计算时机 | 普通字段和外键字段生成完成后，再计算 SQL/Python 字段 |
| 表级与字段级依赖 | 表级依赖只处理表顺序；字段级 DAG 只处理同表同一行内的计算顺序 |
| MVP 引用能力 | 允许引用同表非计算字段、外键字段和少量 `meta` 上下文 |
| 计算字段互引 | MVP 不实现；后续用字段级 DAG + 拓扑排序 + 环检测解决 |
| 引用语法 | 使用 `${...}` 占位符；推荐 `${column}` 简写和 `${field.column}` 显式写法 |
| 非字段值 | 使用 `meta`、`vars` 命名空间，不与字段混用 |
| 引用解析 | MVP 使用占位符扫描；词法分析和完整解析器作为后续增强 |
| SQL 表达式形态 | 只允许单个标量 SQL 表达式，不允许完整 SQL 语句 |
| SQL 方言验证 | 绑定目标数据库方言；MVP 用目标数据库只读 `SELECT` 试解析 |
| SQL 安全 | 禁止多语句、DDL、DML、子查询和对象访问；字段值通过参数绑定 |
| Python 表达式形态 | 只允许单个 Python 表达式，不允许脚本 |
| Python 合法性 | 使用 AST 白名单 + 样例试运行 |
| Python 执行 | MVP 推荐独立 Python Worker，复用进程，受限上下文执行 |
| Python 库 | MVP 只开放白名单内置函数和少量标准库能力，不开放外部库 |
| 性能 | 表达式预编译、依赖缓存、Python Worker 复用、SQL 批量计算 |
| 错误处理 | 配置期尽量提示，执行前强制检查，执行期中止并记录字段级错误 |

---

*本文固定产品与执行语义，具体接口字段可在后续实现设计中根据数据模型微调。*
