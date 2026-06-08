# Requirements Document

## Introduction
LoomiDBX 需要建立统一的本地存储策略，为数据库连接信息、Schema 扫描缓存、字段规则、Project 配置、执行历史和系统设置提供一致落位。当前工程骨架已预留 `internal/storage` 和 `internal/repository`，配置系统规格负责确定本地数据目录，但项目尚未定义本地数据文件布局、结构化业务数据与配置文件的分工、初始化与迁移行为、Repository 边界或敏感信息处理策略。完成后，后续领域模型和服务层应能在清晰边界上扩展具体表结构和仓储实现，同时保持本地隐私边界。

## Boundary Context
- **In scope**: 本地数据目录布局、配置文件与本地结构化数据的职责分工、最小存储初始化、迁移目录与执行约定、Repository 接口落位、测试替身策略、敏感连接信息隔离和后续加密接口边界。
- **Out of scope**: 不实现完整业务数据模型表、不实现所有 Repository、不实现账号/授权/云同步、不实现目标数据库写入、不交付 UI 设置页。
- **Adjacent expectations**: `phase-01-project-structure` 提供 `internal/storage` 和 `internal/repository` 落位；`phase-01-config-system` 提供数据目录发现和普通配置入口；后续 `connection-model`、`project-model`、`generation-job-model`、`connection-api`、`execution-history-api` 在本策略上定义具体模型、表和服务行为。

## Requirements

### Requirement 1: 本地数据布局与职责分工
**Objective:** As a 后续模块开发者, I want 本地配置文件、结构化业务数据和敏感信息有明确分工, so that 新模块不会各自定义互相冲突的存储位置。

#### Acceptance Criteria
1. When 本地存储策略被查阅或初始化, the LoomiDBX 本地存储层 shall 明确哪些数据属于普通配置文件、哪些数据属于本地结构化业务数据、哪些数据属于敏感信息边界。
2. The LoomiDBX 本地存储层 shall 将连接元数据、Schema 缓存、字段规则、Project 配置和执行历史归类为本地结构化业务数据。
3. The LoomiDBX 本地存储层 shall 将主题、语言、数据目录和开发选项等轻量应用设置归类为普通配置文件职责。
4. If 数据包含数据库密码、令牌或其他凭据, then the LoomiDBX 本地存储层 shall 将该数据排除在普通配置文件和普通业务表明文存储之外。

### Requirement 2: 数据目录与文件布局
**Objective:** As a 桌面应用维护者, I want 本地数据目录具有可预测的文件布局, so that 开发、测试和真实桌面运行不会互相污染。

#### Acceptance Criteria
1. When 应用需要本地存储根目录, the LoomiDBX 本地存储层 shall 使用配置系统提供的已解析数据目录作为唯一上游来源。
2. When 本地存储被初始化, the LoomiDBX 本地存储层 shall 定义主业务数据文件、迁移文件目录、临时文件目录和备份或导出预留目录的相对布局。
3. While 测试或开发隔离模式启用, the LoomiDBX 本地存储层 shall 使用隔离数据目录并避免读写真实用户数据目录。
4. If 数据目录不可创建、不可写或布局不完整, then the LoomiDBX 本地存储层 shall 返回可诊断错误并阻止后续业务存储访问。

### Requirement 3: 存储初始化与迁移策略
**Objective:** As a 后端开发者, I want 本地存储具备最小初始化和迁移约定, so that 后续业务表可以按顺序、安全地演进。

#### Acceptance Criteria
1. When 本地存储首次初始化, the LoomiDBX 本地存储层 shall 创建所需目录并准备可连接的本地结构化数据文件。
2. When 本地结构化数据文件打开, the LoomiDBX 本地存储层 shall 确认迁移状态并只应用尚未成功记录的迁移。
3. If 迁移执行失败, then the LoomiDBX 本地存储层 shall 返回失败原因并避免把失败迁移记录为成功状态。
4. The LoomiDBX 本地存储层 shall 提供迁移命名、排序和记录规则，供后续领域模型规格添加具体业务表。

### Requirement 4: Repository 边界与测试替身
**Objective:** As a 服务层实现者, I want Repository 接口、具体存储实现和测试替身有稳定边界, so that 服务层可以在不依赖具体文件细节的情况下访问本地数据。

#### Acceptance Criteria
1. When 后续服务层需要读写本地业务数据, the LoomiDBX 本地存储层 shall 提供 Repository 接口落位和实现落位的分离约定。
2. The LoomiDBX 本地存储层 shall 定义事务或工作单元边界的最小约定，避免服务层各自管理底层连接细节。
3. Where 业务 Repository 尚未由后续规格实现, the LoomiDBX 本地存储层 shall 提供 mock 或内存替身策略用于单元测试。
4. If Repository 调用遇到未实现的业务表或能力, then the LoomiDBX 本地存储层 shall 返回明确的未实现或未迁移错误，而不是静默创建不受控结构。

### Requirement 5: 敏感信息隔离与隐私保护
**Objective:** As a 项目维护者, I want 敏感连接信息与普通本地业务数据隔离, so that 本地存储不会泄露数据库凭据或误上传产品数据。

#### Acceptance Criteria
1. The LoomiDBX 本地存储层 shall 不上传数据库连接信息、Schema、生成规则、Project 配置、生成数据或用户 SQL。
2. When 连接记录需要表达凭据状态, the LoomiDBX 本地存储层 shall 只保存凭据引用、配置状态或脱敏元数据，不保存明文密码。
3. Where 安全存储或加密实现尚未完成, the LoomiDBX 本地存储层 shall 提供接口边界和不可用状态，而不是伪装成已完成加密能力。
4. If 错误、日志或迁移状态涉及敏感字段, then the LoomiDBX 本地存储层 shall 避免输出明文敏感值。

### Requirement 6: 运行时契约与验证能力
**Objective:** As a 开发者, I want 本地存储基础设施可被启动和测试验证, so that 后续规格可以可靠复用该基础能力。

#### Acceptance Criteria
1. When 应用启动需要本地存储能力, the LoomiDBX 本地存储层 shall 提供不依赖前端 UI 的初始化入口。
2. When 初始化完成, the LoomiDBX 本地存储层 shall 返回包含数据目录、业务数据文件路径和迁移状态的诊断视图。
3. If 初始化无法完成, then the LoomiDBX 本地存储层 shall 返回稳定错误码和可读原因，供服务层或 facade 转换。
4. The LoomiDBX 本地存储层 shall 提供覆盖布局、初始化、迁移、Repository 替身和敏感信息边界的测试策略。
