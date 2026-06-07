# LoomiDBX Agentic Coding Guide

本目录用于指导 Agentic Coding。Agent 不应默认阅读 `docs` 下的所有专题文档，而应根据当前开发阶段和任务目标选择最小必要上下文。

## 使用原则

1. 任何新会话优先阅读本文件和 `docs/phase.md`。
2. 根据当前任务阶段阅读对应的阶段文档。
3. 如果使用 GitHub Spec-Kit，请以 `docs/phase.md` 作为 spec 划分依据，再为当前 spec 补充更细的需求、验收标准和任务拆解。
4. 只有当阶段文档或 spec 明确要求时，才阅读原始专题文档。
5. 不要跨阶段实现未要求的功能；如果发现依赖缺失，先说明缺口，再做最小必要实现。
6. 优先改动与当前阶段相关的代码和测试，避免顺手重构无关模块。

## 文档分层

```text
专题文档 Source of Truth
  ↓
phase.md 阶段划分与文档索引
  ↓
agent/*.md 阶段阅读计划
  ↓
Spec-Kit specs 任务级实现契约
  ↓
Agent 编码、测试与验证
```

## 阶段索引

| Phase | 阶段文档 | 目标 |
|---|---|---|
| 0 | [00-project-brief.md](./00-project-brief.md) | 建立项目最小背景与术语 |
| 1 | [01-architecture-bootstrap.md](./01-architecture-bootstrap.md) | 项目骨架、技术边界、基础设施 |
| 2 | [02-domain-model-and-schema.md](./02-domain-model-and-schema.md) | 核心领域模型与 Schema 表达 |
| 3 | [03-generation-engine.md](./03-generation-engine.md) | 数据生成执行引擎 |
| 4 | [04-generator-registry-and-contract.md](./04-generator-registry-and-contract.md) | 生成器契约、注册表与校验框架 |
| 5 | [05-built-in-generators.md](./05-built-in-generators.md) | 内置生成器逐类实现 |
| 6 | [06-external-data-generators.md](./06-external-data-generators.md) | 外部数据源生成器 |
| 7 | [07-api-and-service-layer.md](./07-api-and-service-layer.md) | API 契约与服务层 |
| 8 | [08-ui-and-workflows.md](./08-ui-and-workflows.md) | UI 页面与用户工作流 |
| 9 | [09-testing-and-validation.md](./09-testing-and-validation.md) | 测试、可观测性与验收 |

## 推荐 Agent Prompt 模板

```md
请基于 LoomiDBX 当前阶段实现任务：<任务名称>。

请先阅读：
1. docs/agent/README.md
2. docs/phase.md
3. docs/agent/<当前阶段文档>.md
4. 当前 Spec-Kit spec 文件

请遵守：
- 只实现当前 spec 要求的范围；
- 不主动阅读无关专题文档；
- 不跨阶段实现未来能力；
- 完成后运行与改动相关的最小测试；
- 若发现 spec 与现有设计冲突，先说明冲突并提出最小调整建议。
```
