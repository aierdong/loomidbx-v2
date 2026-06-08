# Config 模块

`internal/config` 是 `phase-01-config-system` 的应用级配置模块，已从工程骨架占位演进为当前 spec 的实现边界。它负责普通应用配置的模型、默认值、路径解析、环境覆盖、加载、保存、校验和设置视图契约；后续 spec 只能复用这些输出，不能把业务数据、完整设置页或远端账号能力塞回本模块。

## 本模块拥有

- 配置模型：`model.go` 定义 `AppConfig`、`UserConfig`、外观、路径、开发模式、未来账号/LLM 入口状态和隐私边界；`dto.go` 定义 facade 与前端 API client 可读取的 `SettingsView` 和更新输入。
- 默认值：`defaults.go` 提供完整默认配置。未来入口必须保持不可用或未配置状态，不能暗示账号、LLM 或远端能力已经可用。
- 路径：`paths.go` 返回确定性的配置文件路径和本地数据目录，支持桌面默认路径、开发覆盖和测试隔离目录。下游本地存储只消费 `SettingsView.paths.dataDir` 与 `SettingsView.development.mode`。
- 加载：`loader.go` 按默认值、普通配置文件、环境覆盖的顺序合成最终配置；缺失配置文件是首次启动状态，不是错误。
- 保存：`store.go` 只读写 `UserConfig` 形态的普通 JSON 配置文件；`service.go` 负责合并用户可变项、保存后重新加载，并返回设置视图或字段级错误。
- 校验：`validate.go` 校验枚举、路径、未来入口状态、隐私边界和敏感明文禁止规则。加载和保存都必须走同一类校验规则。

## 新代码放置判断

- 新增应用级配置字段、枚举或可持久化普通设置，先放在 `model.go`，再同步默认值、DTO、校验和测试。
- 新增配置文件位置、数据目录、开发/测试覆盖规则，放在 `paths.go` 或 `env.go`，不要在调用方重复解析环境变量或用户目录。
- 新增加载合成顺序、来源报告或缺失文件语义，放在 `loader.go`。
- 新增普通配置文件读写、原子保存或解析错误处理，放在 `store.go`。
- 新增字段级校验、敏感明文拦截或未来入口状态限制，放在 `validate.go`。
- 新增后端读取/更新用例，放在 `service.go`；Wails facade 只能薄调用 service，不能实现配置合成、路径解析、保存或校验规则。

## 不属于本模块

- 完整设置页、表单交互、文件选择器、账号登录、LLM 连通性测试和 UI 工作流属于后续 UI/API spec；本模块只提供可被暴露的设置契约。
- SQLite 文件创建、SQLite schema、迁移、备份、Repository、Schema 缓存、Project、执行历史和业务数据持久化属于 `phase-01-local-storage-strategy` 相邻 spec 或后续业务 spec。
- 数据库密码、账号 token、LLM API key、连接字符串等安全凭据存储属于后续安全存储边界；普通配置文件只能表达是否已配置，不能保存明文。
- 远端账号、登录注册、OAuth、授权、遥测发送、LLM 远端调用、网络上传和任何网络依赖都不由本 spec 实现。
- 目标数据库 adapter、Schema introspection、生成器、执行引擎、网络上传、生成数据、Project 配置和用户 SQL 不得从配置模块读取或上传。

## 隐私边界

普通配置、业务数据和安全凭据必须保持分离。普通配置文件不得保存数据库密码、账号 token、LLM API key、Schema、Project、生成规则、生成数据或用户 SQL；错误信息也不得回显敏感输入原值。
