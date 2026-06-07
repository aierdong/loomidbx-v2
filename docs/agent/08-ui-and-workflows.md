# Phase 8：UI 与用户工作流

## 阶段目标

基于 UI DSL 和 API 契约实现 LoomiDBX 的主要界面与用户工作流，包括登录、首页、项目管理、Schema 管理、设置等页面。

## 必须阅读

| 文档 | 用途 | 阅读方式 |
|---|---|---|
| [agentic-ui-dsl.md](../agentic-ui-dsl.md) | UI DSL 解释规则 | UI 阶段首次必读 |
| [api-contract.md](../api-contract.md) | 前后端交互 | 只读当前页面需要的 API |
| [user-stories.md](../user-stories.md) | 用户行为 | 只读当前页面相关故事 |
| [ui-dsl/login.dsl.yaml](../ui-dsl/login.dsl.yaml) | 登录页 | 只在实现登录页时阅读 |
| [ui-dsl/index.dsl.yaml](../ui-dsl/index.dsl.yaml) | 首页 | 只在实现首页时阅读 |
| [ui-dsl/projects.dsl.yaml](../ui-dsl/projects.dsl.yaml) | 项目管理 | 只在实现项目页时阅读 |
| [ui-dsl/schema.dsl.yaml](../ui-dsl/schema.dsl.yaml) | Schema 管理 | 只在实现 Schema 页时阅读 |
| [ui-dsl/settings.dsl.yaml](../ui-dsl/settings.dsl.yaml) | 系统设置 | 只在实现设置页时阅读 |

## 可选阅读

| 文档 | 触发条件 |
|---|---|
| [data-model.md](../data-model.md) | 需要定义 UI 状态模型时 |
| [generators_manual.md](../generators_manual.md) | 实现字段规则配置 UI 时 |
| [settings_login.md](../settings_login.md) | 实现登录/设置细节时 |

## 本阶段核心任务

- 建立 UI 路由、布局和页面骨架。
- 按页面实现 DSL 中定义的意图结构。
- 接入 API 契约和本地状态管理。
- 实现表单校验、加载状态、错误状态和空状态。
- 实现用户从连接、Schema 配置、Project 配置到启动生成的主要路径。

## 非目标

- 不根据视觉细节过度还原 prototype。
- 不在 UI 层实现核心生成算法。
- 不一次实现所有页面。
- 不绕过 API/服务层直接操作底层引擎，除非架构明确允许。

## Spec-Kit 建议拆分

- app-shell-and-routing
- login-page
- home-page
- projects-page
- schema-management-page
- settings-page
- field-rule-editor
- generation-job-progress-view

## Context Budget

每个 UI spec 只读取一个 DSL 文件和它需要的 API 片段。不要一次读取所有 UI DSL。
