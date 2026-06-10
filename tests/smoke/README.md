# Smoke 验证落位

本目录用于记录最轻量的应用存活验证。当前阶段不依赖真实数据库凭据、Schema、Project 配置、生成数据、用户 SQL 或远端账号数据。

Domain-only 规格可以在本目录记录运行时替代证据，例如领域包级 `go test`、项目级 `task test` 或 `go test ./...`。这类记录只证明对应领域包可编译、测试可执行、合同验证通过；不暗示已经完成端到端 liveness、真实数据库访问、执行引擎、API/UI 工作流或桌面启动验证。

后续 Wails dev/build 闭环稳定后，可在这里补充启动桌面壳、加载根页面或调用 bootstrap health 的 smoke 验证说明。
