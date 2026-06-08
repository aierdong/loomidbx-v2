# 开发命令、验证与隐私说明

本文记录 `phase-01-test-tooling` 固化后的命令入口、覆盖范围、环境依赖和后续 spec 引用方式。命令只验证本地源码、工具链和骨架集成，不代表完整业务覆盖率。

## 命令清单

| 命令           | 入口                  | 等价原生命令                                                            | 当前能力                                                                         | 类型                                |
| -------------- | --------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------- | ----------------------------------- |
| setup          | `task setup`          | `npm --prefix frontend install`                                         | 安装前端依赖。                                                                   | 环境准备                            |
| doctor         | `task doctor`         | `go run ./scripts/doctor.go`                                            | 检查 Go 1.25+、Node.js、npm、wails3 和平台提示；缺失时输出下一步安装或诊断命令。 | 环境依赖                            |
| dev            | `task dev`            | `wails3 dev`                                                            | 依赖本机 Wails v3 工具链；用于启动桌面开发模式。                                 | 环境依赖                            |
| format         | `task format`         | `gofmt -w $(git ls-files '*.go')` 与 `npm --prefix frontend run format` | 格式化当前 Git 跟踪的 Go 源码，并检查前端 Prettier 规则。                        | 普通改动必跑                        |
| lint           | `task lint`           | `go vet ./...` 与 `npm --prefix frontend run lint`                      | 对全部 Go package 运行静态检查；前端 lint 当前映射到 `vue-tsc --noEmit`。        | 普通改动必跑                        |
| test           | `task test`           | `go test ./...` 与 `npm --prefix frontend run test`                     | 运行 Go 单元测试和前端 Vitest deterministic 样例验证。                           | 普通改动必跑                        |
| build          | `task build`          | `npm --prefix frontend run build && wails3 build`                       | 标准桌面构建：先运行前端 build，再运行 Wails v3 build。                          | 环境依赖                            |
| build:fallback | `task build:fallback` | `npm --prefix frontend run build && go build -o bin/loomidbx .`         | 缺少 Wails CLI 或平台依赖时的骨架级构建证据；fallback 不能替代完整 Wails build。 | 可选 fallback                       |
| verify         | `task verify`         | 依次运行 doctor、format、lint、test、build 的等价命令                   | 聚合最小质量门，按确定顺序暴露失败阶段。                                         | 普通改动必跑；其中 build 受环境影响 |

## 分层命令

`Taskfile.yml` 还提供以下分层入口，供后续 spec 在局部实现时引用：

- `task format:go`：对 `git ls-files '*.go'` 返回的所有 Go 源码运行 `gofmt -w`。
- `task format:frontend`：运行 `npm --prefix frontend run format`，当前为 Prettier check。
- `task lint:go`：运行 `go vet ./...`。
- `task lint:frontend`：运行前端 lint 入口，当前为 `vue-tsc --noEmit`。
- `task test:go`：运行 `go test ./...`。
- `task test:frontend`：运行 `npm --prefix frontend run test`，当前为 Vitest 样例测试。
- `task build:frontend`：运行前端 typecheck 和 Vite build。

## 前端样例测试依赖

前端新增 Vitest 作为轻量测试 runner，只服务当前 deterministic 样例验证：API client 成功与失败结果转换。该依赖不表示已经建立 UI E2E、覆盖率 gate 或完整组件测试平台。完整 UI 工作流测试延后到 Phase 8/9 对应 spec。

## Doctor 隐私边界

`doctor` 只检查本地工具链和平台前置条件，不读取、不上传以下本地产品数据：

- 数据库凭据和连接信息。
- Schema、表、字段和约束元数据。
- 生成数据、Project 配置和字段生成规则。
- 用户 SQL。
- 远端账号数据。

## Wails build 与 fallback

标准桌面验证是 `task build`，它会先运行前端 build，再运行 `wails3 build`。如果本机缺少 Wails CLI 或平台依赖，`task doctor` 会返回可执行诊断；开发者可运行 `task build:fallback` 获得前端 build + Go build 的骨架级证据。fallback 不能替代完整 Wails build，也不验证真实桌面壳、窗口行为或业务工作流。

## 后续 spec 引用方式

后续 spec 可按改动范围引用以下验收路径：

1. Go 后端改动：至少运行 `task test:go`，涉及静态检查时运行 `task lint:go`。
2. 前端改动：至少运行 `task lint:frontend` 和 `task test:frontend`，涉及构建配置时运行 `task build:frontend`。
3. 跨层或 Wails facade 改动：运行 `task test`；具备本地 Wails 前置条件时运行 `task build`。
4. 普通改动合并前：优先运行 `task verify`。若因本机缺少 Wails 或平台依赖导致 build 阶段失败，应记录 `task doctor` 诊断并运行 `task build:fallback` 作为受限证据。

这些命令覆盖当前骨架、配置基础、数据库方言接口 mock 边界、前端 API client 样例和桌面构建入口；不覆盖真实数据库集成、生成器契约、执行引擎集成、API 契约、完整 UI E2E 或可观测性平台。历史文档中的 deferred 能力仍按本节说明延后到对应 spec。
