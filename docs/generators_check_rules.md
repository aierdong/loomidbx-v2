# 校验规则

本章统一说明配置合法性规则、约束一致性规则与运行时规则，消除分散在各生成器描述中的碎片化说明。

---

## 1. 配置合法性规则（保存时检查）

保存生成器配置时，系统将验证以下规则。不满足时，系统拒绝保存并向用户提示错误原因。

#### 索引

| 编号 | 规则集 | 适用生成器 |
|------|--------|-----------|
| 1.1 | 通用规则 | 所有含对应参数的生成器 |
| 1.2 | `distribution` 特有规则 | distribute_int, distribute_float, distribute_time, foreign_key |
| 1.3 | `distribute_int` 特有规则 | distribute_int |
| 1.4 | `sequential_int` 特有规则 | sequential_int |
| 1.5 | `uniform_int` 特有规则 | uniform_int |
| 1.6 | `unix_timestamp` 特有规则 | unix_timestamp |
| 1.7 | `snowflake` 特有规则 | snowflake |
| 1.8 | `boolean` 特有规则 | boolean |
| 1.9 | `constant` 特有规则 | constant |
| 1.10 | `uniform_float` 特有规则 | uniform_float |
| 1.11 | `distribute_float` 特有规则 | distribute_float |
| 1.12 | `uniform_string` 特有规则 | uniform_string |
| 1.13 | `regex_string` 特有规则 | regex_string |
| 1.14 | `uniform_time` 特有规则 | uniform_time |
| 1.15 | `sequential_time` 特有规则 | sequential_time |
| 1.16 | `distribute_time` 特有规则 | distribute_time |
| 1.17 | `relative_time` 特有规则 | relative_time |
| 1.18 | `idcard` 特有规则 | idcard |
| 1.19 | `email` 特有规则 | email |
| 1.20 | `ip` 特有规则 | ip |
| 1.21 | `name` 特有规则 | name |
| 1.22 | `phone` 特有规则 | phone |
| 1.23 | `uuid` 特有规则 | uuid |
| 1.24 | `address` 特有规则 | address |
| 1.25 | `enums` 特有规则 | enums |
| 1.26 | `sql_expression` 特有规则 | sql_expression |
| 1.27 | `python_expression` 特有规则 | python_expression |
| 1.28 | `foreign_key` 特有规则 | foreign_key |
| 1.29 | `dict_table` 特有规则 | dict_table |
| 1.30 | `external_data_source` 特有规则 | external_data_source |
| 1.31 | `ai_generator` 特有规则 | ai_generator |

---

### 1.1 通用规则（适用所有含对应参数的生成器）

| 规则 |
|------|
| `null_percentage` 若填写，必须在 [0, 1] 范围内 |
| `seed` 若填写，必须为整数 |
| `unique` 若填写，必须为布尔值 |

### 1.2 `distribution` 特有规则（包含在 distribute_int, distribute_float, distribute_time, foreign_key 生成器中）

| 规则 |
|------|
| `distribution` 对象必须且只能包含以下类型之一：`normal`、`log_normal`、`exponential`、`power_law`、`poisson`、`binomial` |
| `normal.std_dev` 必须 > 0 |
| `log_normal.std_dev` 必须 > 0 |
| `exponential.lambda` 必须 > 0 |
| `power_law.alpha` 必须 > 1；`power_law.x_min` 必须 > 0 |
| `poisson.lambda` 必须 > 0 |
| `binomial.n` 必须为正整数；`binomial.p` 必须在 [0, 1] 范围内 |

### 1.3 `distribute_int` 特有规则

| 规则 |
|------|
| 若同时填写 `clamp_min` 和 `clamp_max`，必须满足 `clamp_min ≤ clamp_max` |
| `type_format.radix` 若填写，必须为 `decimal`、`hex`、`octal`、`binary` 之一 |
| `type_format.upper_case` 仅在 `radix = "hex"` 时有意义，其他进制下此字段被忽略（不视为错误） |

### 1.4 `sequential_int` 特有规则

| 规则 |
|------|
| `start` 为必填项 |
| `step` 不能为 0，范围 [-2147483648, 2147483647] |
| `type_format.radix` 若填写，规则同 `distribution` 中的 radix 校验 |
| `type_format.upper_case` 仅在 `radix = "hex"` 时有意义，其他进制下此字段被忽略（不视为错误） |

### 1.5 `uniform_int` 特有规则

| 规则 |
|------|
| `min` 为必填项 |
| `max` 为必填项，必须 ≥ `min` |
| `type_format` 校验规则与 `distribution` 中的 type_format 一致 |

### 1.6 `unix_timestamp` 特有规则

| 规则 |
|------|
| `min` 为必填项，必须为有效的 ISO 8601 格式字符串 |
| `max` 若填写，必须为有效的 ISO 8601 格式字符串 |
| `max` 若填写，必须 ≥ `min` |
| `in_milliseconds` 若填写，必须为布尔值 |

### 1.7 `snowflake` 特有规则

| 规则 |
|------|
| `worker_ids` 为必填项，必须为非空数组 |
| `worker_ids` 每个元素必须在 [0, 1023] 范围内 |
| `worker_ids` 数组长度 ≤ 5 |
| `epoch` 若填写，必须为整数，不能大于当前时间戳（毫秒） |

### 1.8 `boolean` 特有规则

| 规则 |
|------|
| `true_ratio` 若填写，必须在 [0, 1] 范围内 |
| `type_format.boolean_value` 若填写，必须为 `true_false`、`yes_no`、`one_zero`、`on_off`、`yes_no_short`、`on_off_short`、`true_false_short`、`custom` 之一 |
| 当 `type_format.boolean_value = "custom"` 时，`type_format.true_value` 和 `type_format.false_value` 为必填项 |

### 1.9 `constant` 特有规则

| 规则 |
|------|
| `value` 为必填项 |
| `value` 的类型必须与目标列数据类型兼容（详见 §1.4） |

### 1.10 `uniform_float` 特有规则

| 规则 |
|------|
| `min` 为必填项 |
| `max` 为必填项，必须 ≥ `min` |
| `type_format.precision` 若填写，必须为非负整数 |
| `type_format.scientific` 若填写，必须为布尔值 |
| `type_format.thousands_sep` 若填写，必须为布尔值 |

### 1.11 `distribute_float` 特有规则

| 规则 |
|------|
| 使用 `distribution` 特有规则（已在上方定义） |
| 若同时填写 `clamp_min` 和 `clamp_max`，必须满足 `clamp_min ≤ clamp_max` |
| `type_format` 校验规则与 `uniform_float` 的 type_format 一致 |

### 1.12 `uniform_string` 特有规则

| 规则 |
|------|
| `min_length` 为必填项，必须 ≥ 0 |
| `max_length` 为必填项，必须 ≥ `min_length`，且 ≤ 65536 |
| `charset.sets` 若填写，每个元素必须为 `lowercase`、`uppercase`、`digits`、`symbols`、`cjk`、`custom` 之一 |
| 当 `charset.sets` 中包含 `custom` 时，`charset.custom_chars` 为必填项 |
| `type_format.case` 若填写，必须为 `none`、`upper`、`lower`、`title` 之一 |

### 1.13 `regex_string` 特有规则

| 规则 |
|------|
| `pattern` 为必填项，长度 ≤ 256 字符 |
| `max_repeat` 若填写，必须在 [1, 16] 范围内 |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.14 `uniform_time` 特有规则

| 规则 |
|------|
| `start` 为必填项，必须为有效的时间格式（可被 Go 标准库解析） |
| `end` 为必填项，必须为有效的时间格式 |
| `end` 必须 ≥ `start` |
| `type_format.time_format` 若填写，必须为：`iso8601`、`rfc3339`、`date`、`date_slash`、`date_us`、`date_uk`、`date_cn`、`time24h`、`time12h`、`datetime24h`、`datetime12h`、`datetime_zone`、`datetime_zone_name`、`unix_seconds`、`unix_millis` 之一 |
| `type_format.database_type` 若填写，必须为 `PostgreSQL`、`MySQL`、`Oracle`、`Hive`、`SQLServer` 之一 |
| `type_format.datetime_part` 若填写，必须为 `date`、`time`、`datetime` 之一 |

### 1.15 `sequential_time` 特有规则

| 规则 |
|------|
| `start` 为必填项 |
| `unit` 为必填项，必须为 `second`、`minute`、`hour`、`day`、`month`、`year` 之一 |
| `step` 为必填项，必须为正整数 |
| `end` 若填写，必须 ≥ `start` |
| `type_format` 校验规则与 `uniform_time` 一致 |

### 1.16 `distribute_time` 特有规则

| 规则 |
|------|
| `base` 为必填项 |
| `scale_unit` 为必填项，必须为 `second`、`minute`、`hour`、`day`、`month`、`year` 之一 |
| `scale_value` 为必填项，必须为正整数 |
| 使用 `distribution` 特有规则（已在上方定义） |
| 若同时填写 `clamp_min` 和 `clamp_max`，必须满足 `clamp_min ≤ clamp_max` |
| `clamp_min`、`clamp_max` 若填写，必须为有效的时间格式 |
| `type_format` 校验规则与 `uniform_time` 一致 |

### 1.17 `relative_time` 特有规则

| 规则 |
|------|
| `offset_value` 为必填项 |
| `offset_unit` 为必填项，必须为 `year`、`month`、`day`、`hour`、`minute`、`second` 之一 |
| `direction` 若填写，必须为 `add`、`subtract`、`random` 之一 |
| `type_format` 校验规则与 `uniform_time` 一致 |

### 1.18 `idcard` 特有规则

| 规则 |
|------|
| `gender` 若填写，必须为 `male`、`female`、`both` 之一 |
| `min_age` 若填写，必须在 [0, 120] 范围内 |
| `max_age` 若填写，必须在 [0, 120] 范围内 |
| 若同时填写 `min_age` 和 `max_age`，必须满足 `min_age ≤ max_age` |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.19 `email` 特有规则

| 规则 |
|------|
| `region` 为必填项，必须为 `cn`、`global` 之一 |
| `style` 为必填项，必须为 `name`、`random` 之一 |
| `has_number` 若填写，必须为布尔值 |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.20 `ip` 特有规则

| 规则 |
|------|
| `ipv4` 和 `ipv6` 不能同时为 `false`（必须至少启用一个） |
| `lan_ipv4` 仅在 `ipv4 = true` 时有效，否则被忽略（不视为错误） |
| `ipv6_mask` 仅在 `ipv6 = true` 时有效，否则被忽略（不视为错误） |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.21 `name` 特有规则

| 规则 |
|------|
| `language` 若填写，必须为 `zh`、`en` 之一 |
| `gender` 若填写，必须为 `male`、`female`、`both` 之一 |
| `has_middle_name` 若填写，必须为布尔值（仅 `language = "en"` 时生效，中文模式下被忽略） |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.22 `phone` 特有规则

| 规则 |
|------|
| `telephone_ratio` 若填写，必须在 [0, 1] 范围内 |
| `idd` 若填写，必须为布尔值 |
| `area_code` 若填写，必须为布尔值 |
| 当 `idd = true` 时，`area_code` 必须为 `true`，否则拒绝保存 |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.23 `uuid` 特有规则

| 规则 |
|------|
| `version` 为必填项，必须为 `v1`、`v3`、`v4`、`v5`、`v7` 之一 |
| `uuid_format` 若填写，必须为 `standard`、`simple`、`urn`、`base32` 之一 |
| `uppercase` 若填写，必须为布尔值 |
| 当 `version` 为 `v3` 或 `v5` 时，`name` 为必填项 |
| `type_format.case` 校验规则与 `uniform_string` 一致 |

### 1.24 `address` 特有规则

| 规则 |
|------|
| `level` 为必填项，必须为 `province`、`city`、`district`、`street`、`full` 之一 |
| `output_field` 若填写，必须为 `region`、`code` 之一 |
| `scope_codes` 若填写，必须为非空数组，每个元素为 6 位或 12 位字符串（合法的行政区划编码） |

### 1.25 `enums` 特有规则

| 规则 |
|------|
| `values` 为必填项，必须为非空数组 |
| `values` 中的所有元素必须为相同类型，且该类型必须与目标列数据类型兼容 |
| `weights` 若填写，长度必须与 `values` 一致 |
| `weights` 若填写，每个元素必须 ≥ 0 |
| 各 format 对象（`text_format`、`integer_format`、`float_format`、`boolean_format`、`datetime_format`）中的配置项须符合对应基础生成器的 type_format 规则——但仅按当前列数据类型所对应的那个 format 对象执行校验，其余 format 对象若存在则被忽略（不视为错误） |

### 1.26 `sql_expression` 特有规则

| 规则 |
|------|
| `expression` 为必填项 |
| `expression` 中 `${column_name}` 只能引用同表中的非表达式列（即未配置 `sql_expression` 或 `python_expression` 生成器的列） |
| 若引用了另一个表达式列或引用列不存在，配置校验阶段拒绝保存并提示原因 |

### 1.27 `python_expression` 特有规则

| 规则 |
|------|
| `expression` 为必填项 |
| `expression` 中 `${column_name}` 只能引用同表中的非表达式列（即未配置 `sql_expression` 或 `python_expression` 生成器的列） |
| 若引用了另一个表达式列或引用列不存在，配置校验阶段拒绝保存并提示原因 |

### 1.28 `foreign_key` 特有规则

| 规则 |
|------|
| `reference_table` 为必填项 |
| `reference_column` 为必填项 |
| `pick_order` 若填写，必须为 `ordinal`、`random`、`distribution` 之一 |
| 当 `pick_order = "distribution"` 时，`distribution` 字段为必填项 |
| 当 `pick_order = "distribution"` 时，`distribution` 对象的校验规则与 `distribute_int.distribution` 完全一致 |

### 1.29 `dict_table` 特有规则

| 规则 |
|------|
| `sql_query` 为必填项 |
| `sql_query` 必须为 `SELECT` 单列的查询语句 |

### 1.30 `external_data_source` 特有规则

| 规则 |
|------|
| `source_type` 为必填项，必须为 `embedded`、`user_uploaded`、`http_api_resource` 之一 |
| 当 `source_type = "embedded"` 时，`source_config.file_path` 为必填项 |
| 当 `source_type = "user_uploaded"` 时，`source_config.file_id` 为必填项 |
| 当 `source_type = "http_api_resource"` 时，`source_config.url` 为必填项 |
| `source_config.auth.type` 若填写，必须为 `none`、`api_key`、`bearer_token`、`basic_auth`、`hmac_signature`、`digest_auth` 之一 |
| 当 `auth.type = "api_key"` 时，`api_key_name`、`api_key_value`、`api_key_in` 为必填项 |
| 当 `auth.type = "bearer_token"` 时，若使用静态 token 则 `token` 为必填项；若动态获取则 `token_url` 和 `token_post_body` 为必填项 |
| 当 `auth.type = "basic_auth"` 或 `"digest_auth"` 时，`username` 和 `password` 为必填项 |
| 当 `auth.type = "hmac_signature"` 时，`hmac_secret` 为必填项 |

### 1.31 `ai_generator` 特有规则

| 规则 |
|------|
| `prompt` 若填写，长度 ≤ 4096 字符 |
| `llm_generated` 若填写，必须为非负整数 |

---

## 2. 约束一致性规则（配置时警告，执行时强制）

以下规则描述生成器配置与列约束之间的冲突处理方式。冲突不阻止保存，但系统在配置阶段向用户发出警告，并在执行阶段按下表规定强制处理。

| 冲突场景 | 配置阶段行为 | 执行阶段行为 |
|---------|-----------|-----------|
| NOT NULL 列配置了 `null_percentage > 0` | 保存时显示警告 | 执行时将 `null_percentage` 强制视为 0，不生成空值 |
| 非 UNIQUE 约束列配置了 `unique: true` | 保存时显示提示（该配置无实际效果） | 执行时忽略 `unique`，不进行唯一性约束 |
| 物理外键列未使用 `foreign_key` 生成器 | 保存时显示警告（提示可能违反引用完整性） | 执行时不做额外引用检查，由数据库约束兜底 |

---

## 3. 运行时规则（执行前检查）

执行任务前，系统逐项执行以下前置检查，任一项失败则立即中止本次任务并记录错误，不写入任何数据。

| 检查项 | 适用生成器 | 失败原因示例 |
|--------|---------|-----------|
| 参照表在目标数据库中存在 | foreign_key | `reference_table` 指向的表不存在 |
| 参照列在参照表中存在 | foreign_key | `reference_column` 指向的列不存在 |
| 参照表经 `condition_clause` 筛选后有可用数据 | foreign_key | WHERE 条件过滤后结果集为空 |
| 省市区县数据文件可读 | address | 内置数据文件缺失或格式损坏 |
| 街道数据文件可读（仅 `level = "street"` 或 `"full"`） | address | 内置街道数据文件缺失或格式损坏 |
| `scope_codes` 筛选后有可用数据 | address | 指定的编码在数据集中无匹配项 |
| 可生成的唯一值总数 ≥ 需生成的行数（`unique = true` 时） | 所有支持 unique 的生成器 | 列取值空间有限，唯一值已耗尽 |
| 姓名数据文件可读 | name | 内置中文/英文姓名数据文件缺失或格式损坏 |
| 身份证地区码数据文件可读 | idcard | 内置行政区划数据文件缺失或格式损坏 |
| 邮箱域名数据可读 | email | 内置邮箱域名数据文件缺失或格式损坏 |
| 手机号号段数据文件可读 | phone | 内置手机号号段数据文件缺失或格式损坏 |
| 字典表在目标数据库中存在且可查询 | dict_table | `sql_query` 中引用的表不存在或无权限访问 |
| 字典表查询结果非空 | dict_table | `sql_query` 执行后返回 0 行数据 |
| 外部文件可读（`source_type = "embedded"`） | external_data_source | `file_path` 指向的文件不存在或不可读 |
| 外部文件可读（`source_type = "user_uploaded"`） | external_data_source | `file_id` 对应的文件不存在或已过期 |
| 外部 API 可达且返回有效响应（`source_type = "http_api_resource"`） | external_data_source | URL 不可达、HTTP 状态码非 2xx、或 `response_path` 提取结果为空 |
| LLM 连接参数已配置 | ai_generator | 环境变量 `LOOMIDBX_BASE_URL`、`LOOMIDBX_API_KEY`、`LOOMIDBX_MODEL` 未全部设置 |
| LLM API 连通性正常 | ai_generator | 使用配置的连接参数调用 LLM API 失败 |
| 正则表达式可编译 | regex_string | `pattern` 不是合法的正则表达式 |
| Python 表达式语法合法 | python_expression | `expression` 无法被 Python 解释器解析 |
| Python 表达式执行不抛出异常 | python_expression | 以样例值试运行 `expression` 时抛出异常 |

---

## 4. 生成器与列数据类型兼容性

类型不兼容时，系统在保存配置时拒绝，并提示原因。

| 生成器 | 允许的列数据类型 | 备注 |
|--------|---------------|------|
| `distribute_int` | `integer` | -- |
| `sequential_int` | `integer` | -- |
| `uniform_int` | `integer` | -- |
| `unix_timestamp` | `integer` | 输出为 Unix 时间戳（整数） |
| `snowflake` | `integer` | 输出为 64 位长整型 |
| `boolean` | `boolean` | -- |
| `constant` | 任意 | `value` 的类型必须与列数据类型一致 |
| `uniform_float` | `float` | -- |
| `distribute_float` | `float` | -- |
| `uniform_string` | `text` | -- |
| `regex_string` | `text` | -- |
| `uniform_time` | `datetime` | -- |
| `sequential_time` | `datetime` | -- |
| `distribute_time` | `datetime` | -- |
| `relative_time` | `datetime` | -- |
| `idcard` | `text` | -- |
| `email` | `text` | -- |
| `ip` | `text` | -- |
| `name` | `text` | -- |
| `phone` | `text` | -- |
| `uuid` | `text`、`uuid`（PostgreSQL）、`BINARY(16)`（MySQL）、`UNIQUEIDENTIFIER`（SQL Server）、`RAW(16)`（Oracle） | 具体可用类型取决于目标数据库 |
| `address` | `text` | -- |
| `enums` | 任意 | `values` 中的元素类型必须与列数据类型一致 |
| `sql_expression` | 任意 | 表达式计算结果类型必须与列数据类型一致；类型不一致时在执行前检查阶段中止并报错 |
| `python_expression` | 任意 | 表达式计算结果类型必须与列数据类型一致；类型不一致时在执行前检查阶段中止并报错 |
| `foreign_key` | 任意 | 执行前验证参照列类型与目标列类型是否兼容；不兼容时中止任务并记录错误 |
| `dict_table` | 任意 | 执行前验证 SQL 查询结果的列类型与目标列类型是否兼容 |
| `external_data_source` | 任意 | 执行前验证外部数据源提取值的类型与目标列类型是否兼容 |
| `ai_generator` | 任意（通常为 `text`） | AI 返回的文本按目标列类型尝试转换；转换失败时记录错误并跳过该行 |
