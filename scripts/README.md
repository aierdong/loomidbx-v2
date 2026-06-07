# 开发脚本

本目录放置仓库级开发辅助脚本。当前阶段只提供本地工具链诊断，不读取数据库连接、Schema、Project 配置、生成数据、用户 SQL 或远端账号数据。

统一入口优先通过 `Taskfile.yml` 调用；没有安装 Task 时，也可以直接运行 `go run ./scripts/doctor.go`。
