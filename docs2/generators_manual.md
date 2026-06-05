# 数据生成器设计

生成器（Generator）是指为某种数据类型（如 integer）生成模拟数据的代码模块。

## 一、生成器索引

| 生成器类型 | 名称 | 输出数据类型 | 作用 |
|-----------|------|-------------|------|
| `distribute_int` | 整数分布生成器 | integer | 基于概率分布（正态、指数、幂律等）生成随机整数 |
| `sequential_int` | 整数序列生成器 | integer | 生成按照固定步长递增或递减的整数序列 |
| `uniform_int` | 整数均匀分布生成器 | integer | 在指定范围内生成均匀分布的随机整数 |
| `unix_timestamp` | 时间戳生成器 | integer | 在指定时间范围内生成随机的 Unix 时间戳（秒级/毫秒级） |
| `snowflake` | 雪花算法生成器 | integer | 基于 Twitter Snowflake 算法生成分布式唯一 64 位长整型 ID |
| `boolean` | 布尔生成器 | boolean | 生成随机布尔值，可配置 true 概率和输出格式 |
| `constant` | 常量生成器 | 任意 | 生成固定不变的常量值（字符串、数字、布尔值等） |
| `uniform_float` | 浮点数均匀分布生成器 | float | 在指定范围内生成均匀分布的随机浮点数 |
| `distribute_float` | 浮点数分布生成器 | float | 基于概率分布生成浮点数（正态、指数、对数正态等） |
| `uniform_string` | 文本均匀分布生成器 | text | 生成指定长度范围和字符集的随机字符串 |
| `regex_string` | 文本正则表达式生成器 | text | 根据正则表达式模式生成随机字符串 |
| `uniform_time` | 日期时间均匀分布生成器 | datetime | 在指定时间范围内生成均匀分布的随机时间点 |
| `sequential_time` | 日期时间序列生成器 | datetime | 按照指定时间单位和步长生成递增的时间序列 |
| `distribute_time` | 概率分布时间生成器 | datetime | 基于概率分布围绕基准时间点生成随机时间点 |
| `relative_time` | 相对时间生成器 | datetime | 生成一个围绕当前时间（数据生成时间）上下偏移的时间 |
| `idcard` | 身份证号生成器 | text | 生成符合 GB 11643-1999 标准的 18 位中国身份证号码 |
| `email` | 邮箱生成器 | text | 生成符合真实场景的电子邮件地址（中国/全球） |
| `ip` | IP 地址生成器 | text | 生成随机 IPv4（公网/局域网）和 IPv6 地址 |
| `name` | 姓名生成器 | text | 生成随机中文或英文姓名，支持指定性别 |
| `phone` | 手机号生成器 | text | 生成随机中国电话号码（手机号/固定电话） |
| `uuid` | UUID 生成器 | text | 生成 UUID v1/v3/v4/v5/v7，支持多种格式化 |
| `address` | 地址生成器 | text | 基于预置省市区县与街道数据，生成中国行政区划地址字符串 |
| `enums` | 带权重枚举值生成器 | 任意 | 从给定值列表中按权重随机选择 |
| `sql_expression` | SQL 表达式生成器 | 任意 | 通过 SQL 表达式生成数据 |
| `python_expression` | Python 表达式生成器 | 任意 | 通过 Python 表达式生成数据 |
| `foreign_key` | 外键生成器 | 任意 | 为表中的外键字段生成数据，它从参照表（父表）的指定关联列中获取已存在的值 |
| `dict_table` | 字典表生成器 | 任意 | 从另一个表中拉取指定列的数据作为生成值 |
| `external_data_source` | 外部数据源生成器 | 任意 | 从外部数据源（JSON/CSV 文件、外部API）拉取维度数据 |
| `ai_generator` | AI 生成器 | 任意 | 利用 AI 模型根据提示词生成数据 |

---

## 二、数据类型与生成器类型关系

| 列数据类型 | 可用的生成器类型 |
|-----------|----------------|
| integer | distribute_int, sequential_int, uniform_int, unix_timestamp, snowflake, constant, enums, sql_expression, python_expression, foreign_key, dict_table, external_data_source, ai_generator |
| text | uniform_string, regex_string, idcard, email, ip, name, phone, uuid, address, constant, enums, sql_expression, python_expression, foreign_key, dict_table, external_data_source, ai_generator |
| float | distribute_float, uniform_float, constant, enums, sql_expression, python_expression, foreign_key, dict_table, external_data_source, ai_generator |
| boolean | boolean, constant, sql_expression, python_expression, foreign_key, dict_table, external_data_source, ai_generator |
| datetime | uniform_time, sequential_time, distribute_time, relative_time, constant, enums, sql_expression, python_expression, foreign_key, dict_table, external_data_source, ai_generator |

---

## 三、生成器详情

### 基础配置

> 每个生成器（除 constant、enums、foreign_key、address、dict_table、external_data_source、ai_generator 之外）的配置包含以下"基础配置"：
> - **type_format**：类型格式化配置。与数据类型强绑定，改变值的呈现形态但不改变底层类型（如 hex 格式的整数在数据库里仍然是整数）。详见各生成器说明。
> - **stringify**：字符串模板配置。将任意生成器的输出最终转换为字符串。可与任意生成器组合（如 `sequential_int` + stringify 可生成 `No-0001`），但**仅当目标列为 varchar/text 时才生效**。配置校验阶段，若目标列为非字符串类型而使用了 stringify，系统应报错或自动忽略。详见§ stringify 配置。
> - **null_percentage**：生成空值的概率（可选，默认 0，取值范围 0~1），它对非空约束的列无效。
> - **unique**：列数据是否唯一，它仅对"unique"约束列有效。
> - **seed**：随机种子，指定后相同配置每次运行生成相同的数据序列；null 表示每次随机生成。
>
> **不适用基础配置的生成器**：
> - **constant** — 直接输出原始值，无格式化需求
> - **enums** — 从固定集合中选取，无格式化需求（但支持按数据类型分别配置格式，详见§ enums）
> - **foreign_key** — 输出由引用列的值决定，不使用基础配置中的任何字段
> - **address** — 有独立的输出格式控制，不使用 `type_format` 和 `stringify`，但在自身配置中保留了 `null_percentage`、`unique` 和 `seed`
> - **dict_table** — 输出由字典表的列决定
> - **external_data_source** — 输出由外部数据源决定；不支持 type_format、stringify 及 seed，但支持 null_percentage 和 unique
> - **ai_generator** — 输出由 AI 模型决定；不支持 type_format、stringify、unique 及 seed，但支持 null_percentage

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

### 基础配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| type_format | object | 否 | 类型格式化配置。与数据类型强绑定，改变值的呈现形态但不改变底层类型。详见各生成器说明。 | -- |
| stringify | object | 否 | 字符串模板配置。将任意生成器的输出转换为字符串，仅当目标列为 varchar/text 时生效。 | -- |
| stringify.template | string | 否 | 字符串模板，`${value}` 被替换为生成器的原始输出 | null |
| stringify.padding.length | number | 否 | 填充后的总长度 | null |
| stringify.padding.char | string | 否 | 填充字符 | "0" |
| stringify.padding.direction | string | 否 | 填充方向：`left`、`right` | "left" |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |
| seed | integer | 否 | 随机种子。指定后，相同配置每次运行生成相同的数据序列；null 表示每次随机生成 | null |

### stringify 配置

stringify 可与任意生成器组合，将生成器的原始输出转换为字符串。仅当目标列为 varchar/text 时生效；若目标列为非字符串类型，配置校验阶段应报错或自动忽略。

```json
{
  "stringify": {
    "template": "No-${value}",
    "padding": {
      "char": "0",
      "length": 10,
      "direction": "left"
    }
  }
}
```

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| stringify.template | string | 否 | 字符串模板，`${value}` 被替换为生成器的原始输出 | null |
| stringify.padding.length | number | 否 | 填充后的总长度 | null |
| stringify.padding.char | string | 否 | 填充字符 | "0" |
| stringify.padding.direction | string | 否 | 填充方向：`left`、`right` | "left" |

**跨类型示例**：`sequential_int` 生成器 + stringify，目标列为 varchar：
```json
{
  "generator": "sequential_int",
  "start": 1,
  "step": 1,
  "type_format": { "radix": "decimal" },
  "stringify": { "template": "No-${value}", "padding": { "char": "0", "length": 4, "direction": "left" } }
}
```
输出：`No-0001`、`No-0002`、`No-0003` ...

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

---

### 3.2 `sequential_int` — 整数序列生成器

**输出类型**：integer

生成按照固定步长递增或递减的整数序列。

#### 完整配置结构

```json
{
  "start": 1,
  "step": 2,
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
| start | integer | 是 | 序列的起始值 | -- |
| step | integer | 否 | 步长，正数递增、负数递减，范围 [-2147483648, 2147483647]，不得为 0 | 1 |
| type_format | object | 否 | 整数格式化配置（与 distribute_int 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

> **注意**：序列值不设上限截断。当序列值超出目标列数据类型的合法范围时（例如 `INT` 列的值超过 2147483647），写入将失败。执行引擎的异常处理规则见专题 11。

---

### 3.3 `uniform_int` — 整数均匀分布生成器

**输出类型**：integer

在指定范围内生成均匀分布的随机整数。

#### 完整配置结构

```json
{
  "min": 1, 
  "max": 100, 
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
| min | integer | 是 | 随机数的最小值（包含） | -- |
| max | integer | 是 | 随机数的最大值（包含），必须 >= min | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 整数格式化配置（与 distribute_int 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.4 `unix_timestamp` — 时间戳生成器

**输出类型**：integer（输出为 Unix 时间戳，即整数）

在指定的时间范围内生成随机的 Unix 时间戳，支持秒级或毫秒级，例如 1677654321 / 1704067234567。

#### 完整配置结构

```json
{
  "min": "2023-01-01T00:00:00+08:00", 
  "max": "2023-12-31T23:59:59+08:00", 
  "in_milliseconds": false, 
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
| min | string | 是 | 最小时间，ISO 8601 格式，例如 `2006-01-02T15:04:05+08:00` | -- |
| max | string | 否 | 最大时间，格式同上 | 当前时间 |
| in_milliseconds | boolean | 否 | 是否生成毫秒级时间戳 | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 整数格式化配置（与 distribute_int 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.5 `snowflake` — 雪花算法生成器

**输出类型**：integer（64 位长整型）

基于 Twitter Snowflake 算法生成分布式环境下的唯一 ID，由时间戳、机器 ID 和序列号组成。

#### 完整配置结构

```json
{
  "worker_ids": [
    1, 
    2, 
    3
  ], 
  "epoch": 1577808000000, 
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
| worker_ids | array[integer] | 是 | 机器 ID 列表，每个 ID 范围 0-1023，最多 5 个 ID | -- |
| epoch | integer | 否 | 开始时间戳（毫秒），不能大于当前时间 | 0 |
| type_format | object | 否 | 整数格式化配置（与 distribute_int 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.6 `boolean` — 布尔生成器

**输出类型**：boolean

生成随机布尔值，可配置 true 概率和多种输出格式（true/false、yes/no、1/0 等）。

#### 完整配置结构

```json
{
  "true_ratio": 0.5,
  "seed": null,
  "type_format": {
    "boolean_value": "yes_no",
    "true_value": "true",
    "false_value": "false"
  },
  "stringify": {
    "template": "-${value}-", 
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
| true_ratio | float | 否 | 生成 true 的概率，范围 [0,1]。0 全 false，1 全 true | 0.5 |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 布尔值格式化配置 | -- |
| type_format.boolean_value | string | 否 | 预设布尔值表示：`true_false`、`yes_no`、`one_zero`、`on_off`、`yes_no_short`、`on_off_short`、`true_false_short`、`custom` | "yes_no" |
| type_format.true_value | string | 否 | true 的自定义表示（仅 boolean_value=custom 时有效） | "true" |
| type_format.false_value | string | 否 | false 的自定义表示（仅 boolean_value=custom 时有效） | "false" |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.7 `constant` — 常量生成器

**输出类型**：任意（泛型）

生成固定不变的常量值，支持字符串、数字、布尔值等任意类型。

#### 完整配置结构

```json
{
  "value": "Hello, World!",
  "null_percentage": 0
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| value | any | 是 | 要生成的常量值，类型必须与列数据类型一致 |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1]（默认 0） |

---

### 3.8 `uniform_float` — 浮点数均匀分布生成器

**输出类型**：float

在指定范围内生成均匀分布的随机浮点数。

#### 完整配置结构

```json
{
  "min": 0.0,
  "max": 100.0,
  "seed": null,
  "type_format": {
    "precision": 2,
    "scientific": false,
    "thousands_sep": true
  },
  "stringify": {
    "template": "$.${value}", 
    "padding": {
      "length": 5, 
      "char": "0", 
      "direction": "right"
    }
  }, 
  "null_percentage": 0.1, 
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| min | float | 是 | 生成的最小值 | -- |
| max | float | 是 | 生成的最大值 | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 浮点数格式化配置 | -- |
| type_format.precision | integer | 否 | 保留小数位数 | 2 |
| type_format.scientific | boolean | 否 | 是否使用科学计数法 | false |
| type_format.thousands_sep | boolean | 否 | 是否使用千分位分隔符 | false |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.9 `distribute_float` — 浮点数分布生成器

**输出类型**：float

基于概率分布生成浮点数，支持正态分布、指数分布、对数正态分布等。

#### 完整配置结构

```json
{
  "distribution": {
    "normal": {
      "mean": 0,
      "std_dev": 1
    }
  },
  "clamp_min": null,
  "clamp_max": null,
  "seed": null,
  "type_format": {
    "precision": 2,
    "scientific": false,
    "thousands_sep": true
  },
  "stringify": {
    "template": "$.${value}", 
    "padding": {
      "length": 5, 
      "char": "0", 
      "direction": "right"
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
| clamp_min | float | 否 | 输出截断最小值，生成值小于此值时强制替换为 clamp_min | null（不截断） |
| clamp_max | float | 否 | 输出截断最大值，生成值大于此值时强制替换为 clamp_max | null（不截断） |
| type_format | object | 否 | 浮点数格式化配置（与 uniform_float 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |

#### 分布类型参数速查

`distribution` 对象中的 key 即为分布类型名称，与 `distribute_int` 支持相同的分布类型，参见§ 3.1 分布类型参数速查。

---

### 3.10 `uniform_string` — 文本均匀分布生成器

**输出类型**：text

生成指定长度范围内、由指定字符集组合而成的随机字符串。

#### 完整配置结构

```json
{
  "min_length": 8,
  "max_length": 16,
  "charset": {
    "sets": ["lowercase", "uppercase", "digits", "symbols"],
    "custom_chars": "★☆○●"
  },
  "seed": null,
  "type_format": {
    "case": "upper"
  },
  "stringify": {
    "template": "Name:${value}", 
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
| min_length | integer | 是 | 最小长度，必须 >= 0 | -- |
| max_length | integer | 是 | 最大长度，必须 >= min_length，不超过 65536 | -- |
| charset | object | 否 | 字符集配置 | -- |
| charset.sets | array[string] | 否 | 字符集列表，可选值：`lowercase`（小写字母 a-z）、`uppercase`（大写字母 A-Z）、`digits`（数字 0-9）、`symbols`（常用符号）、`cjk`（中文字符）、`custom`（需配合 custom_chars 使用） | ["lowercase", "digits"] |
| charset.custom_chars | string | 否 | 自定义字符集（当 sets 中包含 `custom` 时必填） | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置 | -- |
| type_format.case | string | 否 | 大小写：`none`、`upper`、`lower`、`title` | "none" |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

**字符集组合示例：**
- `["lowercase", "digits"]` — 小写字母 + 数字（默认）
- `["lowercase", "uppercase", "digits"]` — 混合大小写 + 数字
- `["lowercase", "uppercase", "digits", "symbols"]` — 包含符号的强密码字符集
- `["lowercase", "digits", "cjk"]` — 小写字母 + 数字 + 中文
- `["custom"]` — 仅使用自定义字符集

---

### 3.11 `regex_string` — 文本正则表达式生成器

**输出类型**：text

根据指定的正则表达式模式生成随机字符串。

#### 完整配置结构

```json
{
  "pattern": "[A-Z]{2}\\d{4}[a-z]{2}",
  "max_repeat": 8,
  "seed": null,
  "type_format": {
    "case": "upper"
  },
  "stringify": {
    "template": "Name:${value}", 
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
| pattern | string | 是 | 正则表达式模式，长度不超过 256 字符 | -- |
| max_repeat | integer | 否 | 量词的最大重复次数，范围 [1, 16] | 16 |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

**支持的正则特性：**
- 字符类：`[abc]`、`[a-z]`、`[0-9]`
- 预定义字符类：`\d`、`\w`、`\s`
- 量词：`?`、`*`、`+`、`{n}`、`{n,}`、`{n,m}`
- 分组：`(abc)`
- 选择：`a|b`
- 锚点：`^`、`$`

**不支持：** 命名分组、负向字符类（`\D`、`\S`、`\W`——请使用 `[^...]` 代替）、环视

---

### 3.12 `uniform_time` — 日期时间均匀分布生成器

**输出类型**：datetime

在指定的时间范围内生成均匀分布的随机时间点。

#### 完整配置结构

```json
{
  "start": "2023-01-01T00:00:00Z",
  "end": "2024-01-01T00:00:00Z",
  "seed": null,
  "type_format": {
    "time_format": "datetime24h",
    "database_type": "PostgreSQL",
    "datetime_part": "datetime"
  },
  "stringify": {
    "template": "TIME:${value}", 
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
| start | string | 是 | 开始时间，支持任意可被 Go 标准库解析的时间格式 | -- |
| end | string | 是 | 结束时间，支持任意可被 Go 标准库解析的时间格式 | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 时间格式化配置 | -- |
| type_format.time_format | string | 否 | 预定义时间格式（详见表后），与 `database_type` 二者取其一 | "rfc3339" |
| type_format.database_type | string | 否 | 数据库类型：PostgreSQL、MySQL、Oracle、Hive、SQLServer | -- |
| type_format.datetime_part | string | 否 | 日期时间部分（仅数据库格式化时有效）：`date`、`time`、`datetime` | "datetime" |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

**预定义时间格式（time_format）速查：**

| 取值 | 示例 |
|------|------|
| iso8601 | 2023-12-25T15:04:05.000Z |
| rfc3339 | 2023-12-25T15:04:05Z |
| date | 2023-12-25 |
| date_slash | 2023/12/25 |
| date_us | 12/25/2023 |
| date_uk | 25/12/2023 |
| date_cn | 2023年12月25日 |
| time24h | 15:04:05 |
| time12h | 3:04:05 PM |
| datetime24h | 2023-12-25 15:04:05 |
| datetime12h | 2023-12-25 3:04:05 PM |
| datetime_zone | 2023-12-25 15:04:05 +0800 |
| datetime_zone_name | 2023-12-25 15:04:05 CST |
| unix_seconds | 1703491445 |
| unix_millis | 1703491445000 |

> 以上 time_format 定义对 uniform_time、sequential_time、distribute_time 三个生成器通用。

---

### 3.13 `sequential_time` — 日期时间序列生成器

**输出类型**：datetime

按照指定的时间单位和步长生成递增的时间序列，到达结束时间后循环回到开始时间。

#### 完整配置结构

```json
{
  "start": "2023-01-01T00:00:00Z",
  "end": "2024-01-01T00:00:00Z",
  "unit": "day",
  "step": 7,
  "type_format": {
    "time_format": "datetime24h"
  },
  "stringify": {
    "template": "TIME:${value}", 
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
| start | string | 是 | 开始时间 | -- |
| end | string | 否 | 结束时间，到达后重新从 start 开始 | -- |
| unit | string | 是 | 时间单位：`second`、`minute`、`hour`、`day`、`month`、`year` | -- |
| step | integer | 是 | 步长，必须为正整数 | -- |
| type_format | object | 否 | 时间格式化配置（与 uniform_time 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.14 `distribute_time` — 概率分布时间生成器

**输出类型**：datetime

基于概率分布围绕基准时间点生成随机时间点，支持正态分布、指数分布、泊松分布等。

#### 工作原理

`scale_unit` 和 `scale_value` 共同决定分布的时间尺度。分布函数生成一个无量纲的随机数，该数乘以 `scale_value` 后得到时间偏移量，单位为 `scale_unit`，再叠加到 `base` 时间上。

**示例**：`scale_unit="day"`、`scale_value=30`、`distribution={"normal": {"mean": 0, "std_dev": 1}}` 时，生成的时间以 `base` 为中心，标准差为 30 天，即约 68% 的值落在 `base ± 30 天` 范围内。

#### 完整配置结构

```json
{
  "base": "2023-01-01T00:00:00Z",
  "scale_unit": "day",
  "scale_value": 30,
  "distribution": {
    "normal": {
      "mean": 0,
      "std_dev": 1
    }
  },
  "clamp_min": null,
  "clamp_max": null,
  "seed": null,
  "type_format": {
    "time_format": "datetime24h"
  },
  "stringify": {
    "template": "TIME:${value}", 
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
| base | string | 是 | 基准时间点 | -- |
| scale_unit | string | 是 | 时间偏移单位：`second`、`minute`、`hour`、`day`、`month`、`year` | -- |
| scale_value | integer | 是 | 时间偏移尺度，分布随机数乘以此值得到实际偏移量（必须为正整数） | -- |
| distribution | object | 是 | 分布配置，必须包含且仅包含以下分布类型之一，参见§ 3.1 分布类型参数速查 | -- |
| clamp_min | string | 否 | 输出截断最小时间，格式与 base 相同，生成时间早于此值时替换为 clamp_min | null（不截断） |
| clamp_max | string | 否 | 输出截断最大时间，格式与 base 相同，生成时间晚于此值时替换为 clamp_max | null（不截断） |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 时间格式化配置（与 uniform_time 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.15 `relative_time` — 相对时间生成器

**输出类型**：datetime

生成一个围绕基准时间点（默认为每行生成时的实时系统时间）上下偏移的时间。适用于需要生成与某个时间点有固定偏移量的时间数据的场景，例如订单创建时间、支付时间等。

> **说明**：此处"当前时间"指每行数据生成时的实时系统时间，而非任务开始时的时间。在批量生成场景下，每行的"当前时间"可能略有不同。

#### 完整配置结构

```json
{
  "direction": "random",
  "offset_value": 5,
  "offset_unit": "day",
  "seed": null,
  "type_format": {
    "time_format": "datetime24h"
  },
  "stringify": {
    "template": "TIME:${value}", 
    "padding": {
      "length": 5, 
      "char": "0", 
      "direction": "left"
    }
  }, 
  "null_percentage": 0.05,
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| direction | string | 否 | 偏移方向：`add`（增加）、`subtract`（减少）或 `random`（每行随机选择方向） | "random" |
| offset_value | number | 是 | 偏移量数值 | -- |
| offset_unit | string | 是 | 偏移量单位：`year`、`month`、`day`、`hour`、`minute`、`second` | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 时间格式化配置（与 uniform_time 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.16 `idcard` — 身份证号生成器

**输出类型**：text

生成符合 GB 11643-1999 标准的 18 位中国居民身份证号码，支持指定性别和年龄范围。

#### 完整配置结构

```json
{
  "gender": "both",
  "min_age": 0,
  "max_age": 120,
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "IDCARD:${value}", 
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
| gender | string | 否 | 性别：`male`、`female`、`both` | "both" |
| min_age | integer | 否 | 最小年龄，范围 [0, 120] | 0 |
| max_age | integer | 否 | 最大年龄，范围 [0, 120] | 120 |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.17 `email` — 邮箱生成器

**输出类型**：text

生成符合真实场景的电子邮件地址，支持中国/全球区域、基于姓名/随机字符等生成风格。

#### 完整配置结构

```json
{
  "region": "cn",
  "style": "name",
  "has_number": true,
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "EMail:${value}", 
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
| region | string | 是 | 邮箱区域：`cn`（中国）、`global`（全球） | -- |
| style | string | 是 | 生成风格：`name`（基于姓名）、`random`（随机字符） | -- |
| has_number | boolean | 否 | 是否在用户名后添加随机数字 | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.18 `ip` — IP 地址生成器

**输出类型**：text

生成随机 IP 地址，支持 IPv4（公网或局域网）和 IPv6（可带网络掩码）。

#### 完整配置结构

```json
{
  "ipv4": true,
  "lan_ipv4": false,
  "ipv6": false,
  "ipv6_mask": false,
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "IP:${value}", 
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
| ipv4 | boolean | 否 | 是否生成 IPv4 地址 | false |
| lan_ipv4 | boolean | 否 | 是否生成局域网 IPv4（仅 ipv4=true 时有效） | false |
| ipv6 | boolean | 否 | 是否生成 IPv6 地址 | false |
| ipv6_mask | boolean | 否 | 是否为 IPv6 生成网络掩码（仅 ipv6=true 时有效） | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

> 注意：必须至少启用 `ipv4` 或 `ipv6` 中的一个，否则配置校验不通过。

---

### 3.19 `name` — 姓名生成器

**输出类型**：text

生成随机的中文或英文姓名，支持指定性别，英文姓名可选是否包含中间名。

#### 完整配置结构

```json
{
  "language": "zh",
  "gender": "both",
  "has_middle_name": false,
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "Name:${value}", 
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
| language | string | 否 | 姓名语言：`zh`（中文）、`en`（英文） | "zh" |
| gender | string | 否 | 性别：`male`、`female`、`both` | "both" |
| has_middle_name | boolean | 否 | 是否生成中间名（仅对英文姓名有效） | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.20 `phone` — 手机号生成器

**输出类型**：text

生成随机中国电话号码，支持手机号和固定电话混合，可配置国际区号和区号。

#### 完整配置结构

```json
{
  "telephone_ratio": 0.3,
  "idd": false,
  "area_code": true,
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "TEL:${value}", 
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
| telephone_ratio | float | 否 | 固定电话比例 (0-1)，0 全手机号，1 全固定电话 | 0 |
| idd | boolean | 否 | 是否包含国际区号（+86） | false |
| area_code | boolean | 否 | 是否包含区号（idd=true 时 area_code 必须为 true） | false |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

> **配置约束**：`idd=true` 时 `area_code` 必须为 `true`，否则配置校验不通过，后端拒绝保存。

---

### 3.21 `uuid` — UUID 生成器

**输出类型**：text, uuid 类型

生成通用唯一标识符，支持 v1/v3/v4/v5/v7 五个版本和多种格式化方式。

对于不同数据库，UUID 生成器除了支持 text 数据类型外，也支持特定的 *uuid* 类型：

- PostgreSQL: 支持 `varchar / char[36] / UUID` 类型
- MySQL：支持 `varchar / char[36] / BINARY(16)` 类型
- Sql-Server：支持 `varchar / char[36] / UNIQUEIDENTIFIER` 类型
- Oracle：支持 `varchar / RAW(16)` 类型

#### 完整配置结构

```json
{
  "version": "v4",
  "uuid_format": "standard",
  "uppercase": false,
  "namespace": "default",
  "name": "example",
  "seed": null,
  "type_format": {
    "case": "none"
  },
  "stringify": {
    "template": "ID:${value}", 
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
| version | string | 是 | UUID 版本：`v1`、`v3`、`v4`、`v5`、`v7` | -- |
| uuid_format | string | 否 | 格式化：`standard`（8-4-4-4-12）、`simple`（无连字符）、`urn`、`base32` | "standard" |
| uppercase | boolean | 否 | 是否使用大写字母 | false |
| namespace | string | 否 | 用于 v3 和 v5 版本的命名空间 | "default" |
| name | string | 条件 | 用于 v3 和 v5 版本的名称（当 version 为 `v3` 或 `v5` 时必填） | -- |
| seed | integer | 否 | 随机种子，参见§ 基础配置项说明 | null |
| type_format | object | 否 | 字符串格式化配置（与 uniform_string 相同） | -- |
| stringify | object | 否 | 字符串模板配置，参见§ stringify 配置 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

**UUID 版本说明：**

| 版本 | 生成方式 | 典型用途 |
|------|---------|---------|
| v1 | 基于时间戳 + MAC 地址 | 分布式系统唯一性 |
| v3 | 基于 MD5 哈希 + 命名空间 | 确定性 UUID |
| v4 | 基于随机数 | 最高随机性，通用场景 |
| v5 | 基于 SHA1 哈希 + 命名空间 | 比 v3 更好的哈希碰撞保护 |
| v7 | 基于时间戳 | 时间有序，适合排序 |

---

### 3.22 `address` — 地址生成器

**输出类型**：text

基于预置的省/市/区县与街道数据，生成中国行政区划体系下的地址字符串。支持生成省、市、区县、街道四级标准行政区划，以及虚拟门牌号的完整地址。

#### 数据来源

地址生成器依赖以下两个内置数据集（随产品一同分发，无需用户额外配置）：

- **省市区县数据**（JSON）：覆盖全国三级行政区划（省 → 市 → 区/县），每个节点包含行政区划名称（region）和编码（code）
- **街道/乡镇数据**（CSV）：覆盖四级行政区划，每条记录包含街道编码和名称。仅在 `level` 为 `"street"` 或 `"full"` 时加载

#### 完整配置结构

```json
{
  "level": "district",
  "output_field": "region",
  "scope_codes": null,
  "house_number": {
    "include_room": true
  },
  "null_percentage": 0,
  "unique": false,
  "seed": null
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| level | string | 是 | 生成地址的级别，见下方"地址级别速查" | -- |
| output_field | string | 否 | 输出内容：`region`（地址字符串）或 `code`（行政区划编码） | "region" |
| scope_codes | array |null | 否 | 限定抽取范围的行政区划编码列表，支持省、市、区/县级编码混用，为 null 时从全国范围随机抽取 | null |
| house_number | object | 否 | 门牌号配置，仅在 `level = "full"` 时生效 | -- |
| house_number.include_room | boolean | 否 | 是否追加房间号（如"606室"）。为 false 时仅生成楼栋号（如"005号"） | true |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一，仅对 unique 约束列有效 | false |
| seed | integer | 否 | 随机种子，设置后可重现相同的生成序列 | null |

#### 地址级别速查

`level` 值决定地址的层级深度，`region` 输出为从省级起逐级拼接的完整名称路径（各级名称直接相连，无分隔符）：

| level 值 | 说明 | region 输出示例 | code 输出示例 |
|---------|------|----------------|--------------|
| `province` | 省级（一级地址） | 上海市 | 310000 |
| `city` | 地级市/直辖市辖区（二级地址） | 上海市市辖区 | 310100 |
| `district` | 区/县级（三级地址） | 上海市市辖区浦东新区 | 310115 |
| `street` | 街道/乡镇（四级地址） | 上海市市辖区浦东新区东海农场 | 310115402000 |
| `full` | 完整地址，在 street 基础上追加随机门牌号 | 上海市市辖区浦东新区外高桥保税区005号606室 | 310115501000 |

> 当 `output_field = "code"` 且 `level = "full"` 时，输出为所在街道的编码，不含门牌号。

#### 门牌号生成规则

仅在 `level = "full"` 时有效：

- **楼栋号**：在 1 ~ 999 内随机取整，格式化为 3 位零填充后追加"号"，如"005号"
- **房间号**（`include_room = true`）：楼层在 1 ~ 30 内随机取整，房间号在 1 ~ 20 内随机取整，拼接为"楼层 + 两位房间号 + 室"，如"606室"（6 楼 06 室）、"1205室"（12 楼 05 室）

#### `scope_codes` 使用说明

接受省（6 位，末 4 位为 0000）、市（6 位，末 2 位为 00）或区/县（6 位，末位非 0）级编码，可混合使用。生成时只从指定编码范围及其下级中随机抽取。

```json
// 只生成上海市或广东省的地址
{ "scope_codes": ["310000", "440000"] }

// 只生成浦东新区的地址
{ "scope_codes": ["310115"] }
```

#### 示例配置

**示例 1：生成全国区级地址名称**

```json
{
  "level": "district",
  "output_field": "region"
}
```

输出示例：`上海市市辖区浦东新区`

**示例 2：生成带房间号的完整上海地址**

```json
{
  "level": "full",
  "output_field": "region",
  "scope_codes": ["310000"],
  "house_number": {
    "include_room": true
  }
}
```

输出示例：`上海市市辖区浦东新区外高桥保税区005号606室`

**示例 3：只输出区/县编码**

```json
{
  "level": "district",
  "output_field": "code"
}
```

输出示例：`310115`

---

### 3.23 `enums` — 带权重枚举值生成器

**输出类型**：任意（泛型）

从给定的值列表中按照指定的权重随机选择值，支持字符串、数字、布尔值、时间等多种数据类型，也可从外部数据源加载数据。

#### 完整配置结构

```json
{
  "values": ["A", "B", "C", "D"],
  "weights": [0.4, 0.3, 0.2, 0.1],
  "null_percentage": 0.1,
  "unique": false,
  "text_format": {
    "case": "upper",
    "stringify": {
      "template": "T:${value}", 
      "padding": {
        "length": 5, 
        "char": "0", 
        "direction": "left"
      }
    }
  },
  "integer_format": {
    "radix": "decimal",
    "thousands_sep": true,
    "stringify": {
      "template": "No:${value}", 
      "padding": {
        "length": 5, 
        "char": "0", 
        "direction": "left"
      }
    }
  },
  "float_format": {
    "precision": 2,
    "thousands_sep": true,
    "stringify": {
      "template": "$${value}", 
      "padding": {
        "length": 5, 
        "char": "0", 
        "direction": "left"
      }
    }
  },
  "boolean_format": {
    "true_value": "是",
    "false_value": "否",
    "stringify": {
      "template": "-${value}-", 
      "padding": {
        "length": 5, 
        "char": "0", 
        "direction": "left"
      }
    }
  },
  "datetime_format": {
    "time_format": "datetime24h",
    "stringify": {
      "template": "TIME:${value}", 
      "padding": {
        "length": 5, 
        "char": "0", 
        "direction": "left"
      }
    }
  }
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| values | array | 是 | 可选值列表，所有值必须是相同类型 | -- |
| weights | array[float] | 否 | 对应的权重列表，元素必须为非负数，长度需与 values 一致。系统在生成前会自动归一化，无需加总为 1（如 `[4, 3, 2, 1]` 与 `[0.4, 0.3, 0.2, 0.1]` 效果相同）。未填写时各值等概率选取。 | 均匀分布 |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |
| text_format | object | 否 | 字符串格式化配置（包含 type_format + stringify，与 uniform_string 相同） | -- |
| integer_format | object | 否 | 整数格式化配置（包含 type_format + stringify，与 distribute_int 相同） | -- |
| float_format | object | 否 | 浮点数格式化配置（包含 type_format + stringify，与 uniform_float 相同） | -- |
| boolean_format | object | 否 | 布尔值格式化配置（包含 type_format + stringify，与 boolean 相同） | -- |
| datetime_format | object | 否 | 时间格式化配置（包含 type_format + stringify，与 uniform_time 相同） | -- |

---

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

### 3.26 `foreign_key` — 外键生成器

**输出类型**：参照列的数据类型

为表中的外键字段生成数据。它从参照表（父表）的指定关联列中获取已存在的值，以确保数据的一致性和引用完整性。此生成器通常用于处理数据库中的物理外键或逻辑外键关系。

#### 完整配置结构

```json
{
  "reference_table": "products",
  "reference_column": "product_id",
  "condition_clause": "created_at > '2020-01-01'",
  "limit": 0,
  "pick_order": "random",
  "distribution": null,
  "null_percentage": 0.05
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| reference_table | string | 是 | 参照表（父表）的名称 | -- |
| reference_column | string | 是 | 参照表中的关联列名称，该列的值将被用于生成外键数据 | -- |
| condition_clause | string | 否 | SQL 取数条件（不含 WHERE 关键字）。为空时不限制，例如 `status in ('active', 'pending')` | null（不限制）|
| order_clause | string | 否 | SQL 排序方式（不含 ORDER BY 关键字）。为空时不排序，例如 `status,created_at desc,` | null（不限制）|
| limit | integer | 否 | 从参照表加载的最大记录数。0 或 null 表示不限制，全量加载到内存 | 0 |
| pick_order | string | 否 | 父值提取方式：`ordinal`（按加载顺序循环取值）、`random`（均匀随机抽取）、`distribution`（按分布加权抽取，见下节） | "random" |
| distribution | object | 否 | 分布配置。仅在 `pick_order = "distribution"` 时有效，此时为必填。格式与 `distribute_int` 的 `distribution` 字段完全一致 | null |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1]。请注意，如果外键列不允许为空，此配置将无效。 | 0 |

#### 分布模式（`pick_order = "distribution"`）

设置为分布模式后，不同父记录被引用的频率将不再均匀，而是由 `distribution` 配置的概率形状决定。典型用途是模拟"少数记录承担大部分引用"的长尾场景，例如"20% 的客户贡献 80% 的订单"。

**算法说明**

1. 按 `condition_clause` 和 `limit` 加载 N 条父记录，按加载顺序存入数组 V，下标 0 ~ N-1
2. 每次取值时，根据 `distribution` 所描述的概率形状，在 [0, N-1] 范围内采样出一个索引 i（系统将分布的自然值域单调映射到实际的 N 个位置）
3. 取 V[i] 作为本次外键输出值

不同分布类型的效果参考：

| 分布类型 | 效果 |
|--------|------|
| `power_law` | 低编号记录被大量引用，形成强烈的长尾效果 |
| `exponential` | 低编号记录优先被引用，程度低于幂律，曲线较平缓 |
| `normal` | 中间编号记录被引用最多，两端逐渐减少 |

> **注意**：`distribution` 只控制父值的**引用频率分布**，不影响哪些父值被加载（加载逻辑由 `condition_clause`，`order_clause`和 `limit` 控制）。

**长尾场景示例**

```json
{
  "reference_table": "customers",
  "reference_column": "customer_id",
  "condition_clause": "status in ('active', 'pending')",
  "order_clause": "status, created_at desc",
  "limit": 0,
  "pick_order": "distribution",
  "distribution": {
    "power_law": {
      "alpha": 2.0,
      "x_min": 1.0
    }
  },
  "null_percentage": 0
}
```

---

### 3.27 `dict_table` — 字典表生成器

**输出类型**：任意（根据 SQL 查询结果）

从另一个表中拉取指定列的数据作为生成值。这允许用户利用现有数据库中的字典表、配置表或任何单列数据作为数据源，实现数据标准化和复用。SQL 查询必须是 `SELECT` 单列的检索语句。

#### 完整配置结构

```json
{
  "sql_query": "SELECT name FROM categories WHERE is_active = TRUE",
  "null_percentage": 0.05
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| sql_query | string | 是 | 用于从字典表中检索数据的 SQL 查询语句。该语句必须是一个 `SELECT` 单列的检索语句。 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |

---

### 3.28 `external_data_source` — 外部数据源生成器

**输出类型**：任意（根据外部数据源内容）

从外部数据源（包括内嵌 JSON/CSV 文件、用户上传的 JSON/CSV 文件、HTTP(S) API/资源）拉取维度数据。适用于需要从外部系统或文件获取主数据、标准词表或可选项列表的场景，例如从 CRM 拉取销售员信息、从公开数据集获取城市名等。

#### 完整配置结构

```json
{
  "source_type": "http_api_resource",
  "source_config": {
    "file_path": "x.json",
    "file_id": "file_id_01.json",
    "file_field": "product_name",
    "url": "https://api.example.com/products",
    "method": "GET",
    "headers": {
      "Authorization": "Bearer YOUR_TOKEN"
    },
    "body": "{}",
    "response_path": "$.data[*].product_name",
    "auth": {
      "type": "bearer_token",
      "api_key_name": "key",
      "api_key_value": "value",
      "api_key_in": "header",
      "token": "YOUR_STATIC_TOKEN",
      "token_url": "https://api.example.com/sign",
      "token_post_body": "{}",
      "username": "john",
      "password": "YOUR_PASSWORD",
      "hmac_secret": "SECRET_KEY"
    }
  },
  "null_percentage": 0.05,
  "unique": false
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| source_type | string | 是 | 外部数据源类型：`embedded`（内嵌 JSON/CSV 文件）、`user_uploaded`（用户上传 JSON/CSV 文件）、`http_api_resource`（HTTP(S) API/资源） | -- |
| source_config | object | 是 | 外部数据源的具体配置，根据 `source_type` 不同而异。 | -- |
| source_config.file_path | string | 否 | 当 `source_type` 为 `embedded` 时必填。内嵌 JSON/CSV 文件的路径。 | -- |
| source_config.file_id | string | 否 | 当 `source_type` 为 `user_uploaded` 时必填。用户上传 JSON/CSV 文件的唯一标识符。 | -- |
| source_config.file_field | string | 否 | 当 `source_type` 为 `embedded` 或 `user_uploaded` 时可用。指定从文件中提取的字段名称：JSON 文件时为字段 key，CSV 文件时为列标题名称。不填时使用文件中的第一个字段/列。 | null（第一个字段）|
| source_config.url | string | 否 | 当 `source_type` 为 `http_api_resource` 时必填。HTTP(S) API 或资源的 URL。 | -- |
| source_config.method | string | 否 | 当 `source_type` 为 `http_api_resource` 时可用。HTTP 请求方法，如 `GET`、`POST`。 | "GET" |
| source_config.headers | object | 否 | 当 `source_type` 为 `http_api_resource` 时可用。HTTP 请求头，键值对形式。 | -- |
| source_config.body | string | 否 | 当 `source_type` 为 `http_api_resource` 且 `method` 为 `POST` 或 `PUT` 时可用。HTTP 请求体。 | -- |
| source_config.response_path | string | 否 | 当 `source_type` 为 `http_api_resource` 时可用。用于从 API 响应中提取数据的 JSONPath 表达式。 | "$" |
| source_config.auth | object | 否 | 当 `source_type` 为 `http_api_resource` 时可用。HTTP 认证配置。 | -- |
| source_config.auth.type | string | 否 | 认证类型：`none`、`api_key`、`bearer_token`、`basic_auth`、`hmac_signature`、`digest_auth`。 | "none" |
| source_config.auth.api_key_name | string | 否 | 当 `auth.type` 为 `api_key` 时必填。API Key 的名称。 | -- |
| source_config.auth.api_key_value | string | 否 | 当 `auth.type` 为 `api_key` 时必填。API Key 的值。 | -- |
| source_config.auth.api_key_in | string | 否 | 当 `auth.type` 为 `api_key` 时必填。API Key 放置位置：`header` 或 `query`。 | "header" |
| source_config.auth.token | string | 否 | 当 `auth.type` 为 `bearer_token` 时必填。Bearer Token 的值。 | -- |
| source_config.auth.token_url | string | 否 | 当 `auth.type` 为 `bearer_token` 且需要动态获取 token 时必填。用于获取 token 的 URL。 | -- |
| source_config.auth.token_post_body | string | 否 | 当 `auth.type` 为 `bearer_token` 且需要动态获取 token 时必填。获取 token 的 POST 请求体。 | -- |
| source_config.auth.username | string | 否 | 当 `auth.type` 为 `basic_auth` 或 `digest_auth` 时必填。认证用户名。 | -- |
| source_config.auth.password | string | 否 | 当 `auth.type` 为 `basic_auth` 或 `digest_auth` 时必填。认证密码。 | -- |
| source_config.auth.hmac_secret | string | 否 | 当 `auth.type` 为 `hmac_signature` 时必填。HMAC 签名密钥。 | -- |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |
| unique | boolean | 否 | 生成值是否唯一 | false |

---

### 3.29 `ai_generator` — AI 生成器

**输出类型**：任意（通常为 text）

利用 AI 模型根据提示词生成数据。如果提示词为空，系统将根据列名称和数据类型自动构造一个提示词。此生成器提供了高度灵活的数据生成能力，适用于需要生成自然语言文本、创意内容或复杂结构化数据的场景。使用此生成器需要配置 LLM 连接参数。

#### 完整配置结构

```json
{
  "prompt": "生成一个关于科幻电影的标题",
  "llm_generated": 100,
  "null_percentage": 0.05
}
```

#### 配置项说明

| 配置项 | 类型 | 必填 | 说明 | 默认值 |
|--------|------|------|------|--------|
| prompt | string | 否 | 用于指导 AI 生成内容的提示词。如果为空，系统将根据列名称和数据类型自动构造提示词。 | 自动构造 |
| llm_generated | integer | 否 | LLM 实际请求生成的不重复值数量，与目标行数无关。超出部分从已生成值中随机复用。0 或 null 表示每行独立请求（不建议在大批量生成场景下使用）。 | 0 |
| null_percentage | number | 否 | 生成空值的概率，范围 [0,1] | 0 |

**LLM 连接参数**：

使用 AI 生成器必须在环境变量中包含以下参数，用于指定 LLM 连接参数（OpenAI 兼容）：

- `LOOMIDBX_BASE_URL`: LLM API 的基础 URL。
- `LOOMIDBX_API_KEY`: LLM API 的认证密钥。
- `LOOMIDBX_MODEL`: 要使用的 LLM 模型名称。

---

## 附录：数据类型速查

| 数据类型 | 说明 |
|---------|------|
| integer | 整数类型 |
| text | 字符串类型 |
| float | 浮点数类型 |
| boolean | 布尔类型 |
| datetime | 日期时间类型2. 核心术语表 (Terminology) |
