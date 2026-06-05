# 数据生成器设计

生成器（Generator）是指为某种数据类型（如 integer）生成模拟数据的代码模块。

## 一、生成器索引

| 生成器类型 | 名称 | 输出数据类型 | 作用 |
|-----------|------|-------------|------|
| `distribute_int` | 整数分布生成器 | integer | 基于概率分布（正态、指数、幂律等）生成随机整数 |
| `dict_table` | 字典表生成器 | 任意 | 从另一个表中拉取指定列的数据作为生成值 |
| `external_data_source` | 外部数据源生成器 | 任意 | 从外部数据源（JSON/CSV 文件、外部API）拉取维度数据 |
---

## 二、数据类型与生成器类型关系

| 列数据类型 | 可用的生成器类型 |
|-----------|----------------|
| integer | distribute_int, dict_table, external_data_source |
| text | dict_table, external_data_source |
| float | dict_table, external_data_source |
| boolean | dict_table, external_data_source |
| datetime | dict_table, external_data_source |

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
> - **dict_table** — 输出由字典表的列决定
> - **external_data_source** — 输出由外部数据源决定；不支持 type_format、stringify 及 seed，但支持 null_percentage 和 unique

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

## 附录：数据类型速查

| 数据类型 | 说明 |
|---------|------|
| integer | 整数类型 |
| text | 字符串类型 |
| float | 浮点数类型 |
| boolean | 布尔类型 |
| datetime | 日期时间类型 |
