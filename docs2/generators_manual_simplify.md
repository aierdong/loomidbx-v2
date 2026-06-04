# 数据生成器设计

生成器（Generator）是指为某种数据类型（如 integer）生成模拟数据的代码模块。

## 一、生成器索引

| 生成器类型 | 名称 | 输出数据类型 | 作用 |
|-----------|------|-------------|------|
| `distribute_int` | 整数分布生成器 | integer | 基于概率分布（正态、指数、幂律等）生成随机整数 |
| `sql_expression` | SQL 表达式生成器 | 任意 | 通过 SQL 表达式生成数据 |
| `python_expression` | Python 表达式生成器 | 任意 | 通过 Python 表达式生成数据 |

---

## 二、数据类型与生成器类型关系

| 列数据类型 | 可用的生成器类型 |
|-----------|----------------|
| integer | distribute_int, sql_expression, python_expression |
| text | sql_expression, python_expression |
| float | sql_expression, python_expression |
| boolean | sql_expression, python_expression |
| datetime | sql_expression, python_expression |

---

## 三、生成器详情

### 基础配置

> 每个生成器（除 constant、enums、foreign_key、dict_table、external_data_source、ai_generator 之外）的配置包含以下"基础配置"：
> - **type_format**：类型格式化配置。
> - **stringify**：字符串模板配置。将任意生成器的输出最终转换为字符串。
> - **null_percentage**：生成空值的概率（可选，默认 0，取值范围 0~1）。
> - **unique**：列数据是否唯一，它仅对"unique"约束列有效。
> - **seed**：随机种子。
>
> **不适用基础配置的生成器**：
> - **foreign_key** — 输出由引用列的值决定

```json
{
  "type_format": {},
  "stringify": {
    "template": "ID-${value}", 
    "padding": {
      "length": 5, 
      "char": "0", 
      "direction": "left"
    }
  }, 
  "null_percentage": 0.1, 
  "unique": false,
  "seed": null
}
```

---

### 3.1 `distribute_int` — 整数分布生成器

**输出类型**：integer

基于概率分布生成整数序列，支持正态分布、指数分布、幂律分布等。

#### 完整配置结构

```json
{
  "distribution": {
    "normal": {
      "mean": 50,
      "std_dev": 15
    }
  },
  "clamp_min": null,
  "clamp_max": null,
  "seed": null,
  "type_format": {
    "radix": "decimal", 
    "upper_case": false, 
    "thousands_sep": true
  }, 
  "stringify": {
    "template": "ID-${value}", 
    "padding": {
      "length": 5, 
      "char": "0", 
      "direction": "left"
    }
  }, 
  "null_percentage": 0.1, 
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| distribution | object | 是 | 分布配置，必须包含且仅包含以下分布类型之一 | -- |
| clamp_min | integer | 否 | 输出截断最小值，生成值小于此值时强制替换为 clamp_min | null（不截断） |
| clamp_max | integer | 否 | 输出截断最大值，生成值大于此值时强制替换为 clamp_max | null（不截断） |
| type_format | object | 否 | 整数格式化配置 | -- |
| type_format.radix | string | 否 | 数字进制：`decimal`（十进制）、`hex`（十六进制）、`octal`（八进制）、`binary`（二进制） | "decimal" |
| type_format.upper_case | boolean | 否 | 是否使用大写（用于十六进制） | false |
| type_format.thousands_sep | boolean | 否 | 是否使用千分位分隔符（仅在十进制模式下有效） | false |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |

#### 分布类型参数速查

`distribution` 对象中的 key 即为分布类型名称，必须包含且仅包含以下类型之一：

| 分布类型（key） | 参数名 | 类型 | 必填 | 说明 | 默认值 |
|---|---|---|---|---|---|
| `normal`（正态分布） | mean | number | 否 | 均值 | 0 |
| | std_dev | number | 否 | 标准差（必须 > 0） | 1 |
| `log_normal`（对数正态分布） | mean | number | 否 | 对数均值 | 0 |
| | std_dev | number | 否 | 对数标准差（必须 > 0） | 1 |
| `exponential`（指数分布） | lambda | number | 是 | 率参数 λ（必须 > 0） | -- |
| `power_law`（幂律分布） | alpha | number | 是 | 幂律指数（必须 > 1） | -- |
| | x_min | number | 是 | 最小值（必须 > 0） | -- |
| `poisson`（泊松分布） | lambda | number | 是 | 期望 λ（必须 > 0） | -- |
| `binomial`（二项分布） | n | integer | 是 | 试验次数（必须 > 0） | -- |
| | p | number | 是 | 成功概率，范围 [0,1] | -- |

**示例：指数分布**
```json
{
  "distribution": {
    "exponential": { "lambda": 0.5 }
  },
  "clamp_min": 0,
  "clamp_max": 9999
}
```

### 3.24 `sql_expression` — SQL 表达式生成器

**输出类型**：任意（根据表达式结果）

允许用户通过 SQL 表达式生成数据，表达式中可引用其他列的值。这对于生成复杂的、依赖于其他字段的计算值非常有用，例如订单编号、组合地址等。

#### 完整配置结构

```json
{
  "expression": "CONCAT('ORD-', DATE_FORMAT(${order_time}, '%Y%m%d'), LPAD(${seq}, 4, '0'))",
  "null_percentage": 0.05,
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| expression | string | 是 | SQL 表达式字符串。表达式中可以使用 `${column_name}` 引用当前表中的其他列。 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

> **引用限制**：`${column_name}` 只能引用同表中的**非表达式列**（即未配置 `sql_expression` 或 `python_expression` 生成器的列）。若引用了另一个表达式列，配置校验阶段将报错。计算列的执行规则见专题 7。

---

### 3.25 `python_expression` — Python 表达式生成器

**输出类型**：任意（根据表达式结果）

允许用户通过 Python 表达式生成数据，表达式中可引用其他列的值。表达式中仅支持 Python 标准库函数，以确保安全性和可预测性。适用于需要进行复杂逻辑处理或数据转换的场景。

#### 完整配置结构

```json
{
  "expression": "${name}.upper() + '_' + str(${id})",
  "null_percentage": 0.05,
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| expression | string | 是 | Python 表达式字符串。表达式中可以使用 `${column_name}` 引用当前表中的其他列。仅支持 Python 标准库函数。 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

> **引用限制**：`${column_name}` 只能引用同表中的**非表达式列**（即未配置 `sql_expression` 或 `python_expression` 生成器的列）。若引用了另一个表达式列，配置校验阶段将报错。计算列的执行规则见专题 7。

---


## 附录：数据类型速查

| 数据类型 | 说明 |
|---------|------|
| integer | 整数类型 |
| text | 字符串类型 |
| float | 浮点数类型 |
| boolean | 布尔类型 |
| datetime | 日期时间类型 |
