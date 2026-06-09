# Requirements Document

## Introduction

`phase-02-relation-model` 用于定义 ForeignKey 与 TableRelation，统一 Parent/Child、BaseTable/JoinTable、倍数范围、物理/逻辑关系边界。 本规格只覆盖 Phase 2 领域模型、基础校验和序列化合同，不实现未来服务、API、UI 或执行算法。

## Boundary Context

- **In scope**: 定义 ForeignKey 与 TableRelation，统一 Parent/Child、BaseTable/JoinTable、倍数范围、物理/逻辑关系边界。
- **Out of scope**: 拓扑排序、行数规划、执行时外键取值、关系编辑 UI。
- **Adjacent expectations**: 上游依赖：phase-02-table-field-constraint-model；下游规格应复用本规格定义的身份、枚举、字段名和校验结果。

## Requirements

### Requirement 1: 领域模型表达

**Objective:** As a 开发人员, I want 系统具备外键与表关系模型的领域模型表达能力, so that 上下游规格可以在清晰边界内协同。

#### Acceptance Criteria

1. When 相关数据被创建或加载时, the 系统 shall 表达 `外键与表关系模型` 所需的稳定身份、父级引用和核心字段。
2. When 下游组件消费该模型时, the 系统 shall 提供稳定 JSON 字段名和可序列化枚举值。
3. If 输入缺少必填字段或引用不合法, then the 系统 shall 返回字段级校验错误。
4. The 系统 shall 不实现超出本规格边界的服务、API、UI、数据库访问或执行算法。
5. The 系统 shall 通过单元测试覆盖模型创建、校验、枚举和序列化行为。

### Requirement 2: 枚举与状态边界

**Objective:** As a 开发人员, I want 系统具备外键与表关系模型的枚举与状态边界能力, so that 上下游规格可以在清晰边界内协同。

#### Acceptance Criteria

1. When 相关数据被创建或加载时, the 系统 shall 表达 `外键与表关系模型` 所需的稳定身份、父级引用和核心字段。
2. When 下游组件消费该模型时, the 系统 shall 提供稳定 JSON 字段名和可序列化枚举值。
3. If 输入缺少必填字段或引用不合法, then the 系统 shall 返回字段级校验错误。
4. The 系统 shall 不实现超出本规格边界的服务、API、UI、数据库访问或执行算法。
5. The 系统 shall 通过单元测试覆盖模型创建、校验、枚举和序列化行为。

### Requirement 3: 上游引用与下游合同

**Objective:** As a 开发人员, I want 系统具备外键与表关系模型的上游引用与下游合同能力, so that 上下游规格可以在清晰边界内协同。

#### Acceptance Criteria

1. When 相关数据被创建或加载时, the 系统 shall 表达 `外键与表关系模型` 所需的稳定身份、父级引用和核心字段。
2. When 下游组件消费该模型时, the 系统 shall 提供稳定 JSON 字段名和可序列化枚举值。
3. If 输入缺少必填字段或引用不合法, then the 系统 shall 返回字段级校验错误。
4. The 系统 shall 不实现超出本规格边界的服务、API、UI、数据库访问或执行算法。
5. The 系统 shall 通过单元测试覆盖模型创建、校验、枚举和序列化行为。

### Requirement 4: 基础校验

**Objective:** As a 开发人员, I want 系统具备外键与表关系模型的基础校验能力, so that 上下游规格可以在清晰边界内协同。

#### Acceptance Criteria

1. When 相关数据被创建或加载时, the 系统 shall 表达 `外键与表关系模型` 所需的稳定身份、父级引用和核心字段。
2. When 下游组件消费该模型时, the 系统 shall 提供稳定 JSON 字段名和可序列化枚举值。
3. If 输入缺少必填字段或引用不合法, then the 系统 shall 返回字段级校验错误。
4. The 系统 shall 不实现超出本规格边界的服务、API、UI、数据库访问或执行算法。
5. The 系统 shall 通过单元测试覆盖模型创建、校验、枚举和序列化行为。

### Requirement 5: 序列化与测试

**Objective:** As a 开发人员, I want 系统具备外键与表关系模型的序列化与测试能力, so that 上下游规格可以在清晰边界内协同。

#### Acceptance Criteria

1. When 相关数据被创建或加载时, the 系统 shall 表达 `外键与表关系模型` 所需的稳定身份、父级引用和核心字段。
2. When 下游组件消费该模型时, the 系统 shall 提供稳定 JSON 字段名和可序列化枚举值。
3. If 输入缺少必填字段或引用不合法, then the 系统 shall 返回字段级校验错误。
4. The 系统 shall 不实现超出本规格边界的服务、API、UI、数据库访问或执行算法。
5. The 系统 shall 通过单元测试覆盖模型创建、校验、枚举和序列化行为。
