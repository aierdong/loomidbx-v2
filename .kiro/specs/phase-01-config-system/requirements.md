# Requirements Document

## Introduction
LoomiDBX 需要为桌面端应用建立统一配置系统。开发者、后续服务层和设置页都需要从同一入口读取应用级配置，避免语言、主题、本地数据目录、开发环境选项和未来账号/LLM 配置入口分散在环境变量、硬编码默认值或业务模块内部。当前工程骨架已经为 `internal/config` 预留位置，但尚未实现源码级配置模型、默认值、加载/保存接口或开发覆盖机制。完成后，应用应具备可验证的配置默认值、路径约定、加载顺序、校验结果和保存能力，为本地存储、设置页和服务初始化提供稳定基础。

## Boundary Context
- **In scope**: 应用级配置模型、默认值、配置文件路径约定、数据目录发现策略、加载/校验/保存行为、开发/测试覆盖行为，以及面向 Go service 和 Vue 设置页的读取契约。
- **Out of scope**: 不实现完整设置页面、不实现账号登录注册、不实现本地 SQLite schema 或迁移、不保存数据库连接密码的最终加密方案、不管理 Project、Schema、执行历史等业务数据。
- **Adjacent expectations**: `phase-01-project-structure` 提供工程骨架和 `internal/config` 落位；`phase-01-local-storage-strategy` 复用本 spec 的数据目录和配置入口来确定 SQLite 等本地存储位置；后续 `settings-page` 可调用配置读取/更新契约，但 UI 表单交互不由本 spec 交付。

## Requirements

### Requirement 1: 配置模型与默认值
**Objective:** As a 后续模块开发者, I want 应用级配置具备统一模型和默认值, so that 各模块不再各自硬编码基础设置。

#### Acceptance Criteria
1. When 应用在没有用户配置文件的环境中启动, the LoomiDBX 配置系统 shall 提供完整且可校验的默认配置。
2. The LoomiDBX 配置系统 shall 覆盖应用语言、主题、本地数据目录、开发环境选项，以及未来账号和 LLM 配置入口的占位状态。
3. If 某项配置属于未来功能入口, then the LoomiDBX 配置系统 shall 表明该入口当前未完成且不得暗示账号、LLM 或远端能力已经可用。
4. The LoomiDBX 配置系统 shall 区分普通应用配置、业务数据和敏感凭据，不把数据库密码或令牌作为普通配置明文读写目标。

### Requirement 2: 配置路径与数据目录约定
**Objective:** As a 桌面应用维护者, I want 配置文件路径和本地数据目录有稳定约定, so that 配置和后续本地存储可以在不同环境中可预测地落位。

#### Acceptance Criteria
1. When 应用需要定位配置文件, the LoomiDBX 配置系统 shall 返回当前运行环境下的确定性配置文件路径。
2. When 应用需要定位本地数据目录, the LoomiDBX 配置系统 shall 返回当前运行环境下的确定性数据目录。
3. If 用户或开发环境提供有效目录覆盖, then the LoomiDBX 配置系统 shall 使用覆盖后的路径并保持配置文件路径与数据目录之间的关系可解释。
4. If 路径不可创建、不可写或格式无效, then the LoomiDBX 配置系统 shall 返回可读的校验错误，而不是继续使用不明确路径。

### Requirement 3: 加载顺序与环境覆盖
**Objective:** As a 开发者, I want 配置加载顺序清晰且支持开发/测试覆盖, so that 本地开发、测试和桌面运行可以复用同一配置入口。

#### Acceptance Criteria
1. When 配置加载开始, the LoomiDBX 配置系统 shall 按默认值、配置文件、环境覆盖的顺序合成最终配置。
2. If 配置文件不存在, then the LoomiDBX 配置系统 shall 使用默认配置并报告该状态，不把缺失文件视为启动失败。
3. If 环境覆盖提供无效值, then the LoomiDBX 配置系统 shall 拒绝该覆盖并返回具体错误。
4. While 测试或开发模式启用, the LoomiDBX 配置系统 shall 允许使用隔离的数据目录和配置文件位置，避免污染真实用户配置。

### Requirement 4: 校验、错误与隐私保护
**Objective:** As a 项目维护者, I want 配置系统在保存和加载时执行校验并保护隐私边界, so that 无效配置和敏感数据不会悄悄进入普通配置文件。

#### Acceptance Criteria
1. When 配置被加载或保存, the LoomiDBX 配置系统 shall 校验枚举值、路径值和未来入口状态。
2. If 配置内容无效, then the LoomiDBX 配置系统 shall 返回包含字段位置和原因的错误信息。
3. The LoomiDBX 配置系统 shall 不上传数据库连接信息、Schema、生成规则、Project 配置、生成数据或用户 SQL。
4. If 配置中出现敏感字段输入, then the LoomiDBX 配置系统 shall 只记录是否已配置或交给后续安全存储边界处理，不在普通配置文件中保存明文密钥。

### Requirement 5: 保存与更新行为
**Objective:** As a 设置能力实现者, I want 配置可以被安全更新和保存, so that 设置页和服务层可以在不破坏默认值的前提下持久化用户选择。

#### Acceptance Criteria
1. When 调用方读取当前配置, the LoomiDBX 配置系统 shall 返回已合成且已校验的配置视图。
2. When 调用方保存有效配置更新, the LoomiDBX 配置系统 shall 持久化用户可变部分并保留未显式修改字段的默认值或现有值。
3. If 保存过程中发生写入失败, then the LoomiDBX 配置系统 shall 返回失败原因且不得留下无法解析的配置文件。
4. When 保存完成后再次加载配置, the LoomiDBX 配置系统 shall 返回与已保存用户可变项一致的配置结果。

### Requirement 6: 服务与前端契约入口
**Objective:** As a 全栈开发者, I want Go 服务层和前端设置页拥有稳定的配置读取入口, so that 后续页面和服务初始化可以复用同一配置能力。

#### Acceptance Criteria
1. When Go 后端服务需要应用配置, the LoomiDBX 配置系统 shall 提供不依赖 Wails UI 的读取入口。
2. When 前端需要展示基础设置, the LoomiDBX 配置系统 shall 提供可由 Wails facade 暴露的设置读取契约。
3. If 前端请求更新本 spec 支持的设置项, then the LoomiDBX 配置系统 shall 返回更新后的配置视图或可展示的校验错误。
4. The LoomiDBX 配置系统 shall 保持设置契约与完整设置页面解耦，不要求本 spec 交付 UI 表单或远端账号交互。

