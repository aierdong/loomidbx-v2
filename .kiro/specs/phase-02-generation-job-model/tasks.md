# Implementation Plan

- [ ] 1. 建立 `生成任务与执行历史模型` 领域包
- [x] 1.1 创建核心模型文件 (P)
  - 创建 `生成任务与执行历史模型` 所需 Go 文件，并保持 domain 层职责。
  - 为导出类型、字段、常量和枚举值添加项目要求的 Go 注释。
  - 完成后，包可独立编译，不依赖 Wails、Vue 或真实数据库驱动。
  - _Requirements: 1, 2_
  - _Boundary: DomainScaffold_

- [ ] 2. 实现核心模型与枚举
- [x] 2.1 实现核心实体和值对象
  - 实现 GenerationJob, ExecutionTask 的稳定身份、父级引用、字段和 JSON 标签。
  - 确保模型不包含 out-of-scope 的服务、API、UI 或执行字段。
  - 完成后，下游规格可以通过稳定合同消费模型。
  - _Requirements: 1, 3_
  - _Boundary: DomainModels_
  - _Depends: 1.1_

- [x] 2.2 实现枚举、状态和值对象 (P)
  - 实现 ExecutionTableResult, JobStatus 相关枚举或值对象。
  - 枚举使用稳定字符串值，并能识别未知值。
  - 完成后，类型和状态可以安全序列化。
  - _Requirements: 2, 3_
  - _Boundary: DomainEnums_
  - _Depends: 1.1_

- [ ] 3. 实现基础校验
- [x] 3.1 定义字段级校验错误结构
  - 定义字段路径、错误码和安全消息。
  - 支持一次返回多个校验问题。
  - 完成后，下游服务和 UI 可以复用结构化错误。
  - _Requirements: 4_
  - _Boundary: Validation_
  - _Depends: 2.1, 2.2_

- [ ] 3.2 实现模型校验规则
  - 覆盖必填字段、父级引用、枚举合法性、范围和唯一性规则。
  - 不访问数据库、执行引擎或 UI 状态。
  - 完成后，无效模型返回可诊断且安全的错误集合。
  - _Requirements: 3, 4_
  - _Boundary: Validation_
  - _Depends: 3.1_

- [ ] 4. 增加测试
- [ ] 4.1 覆盖序列化和枚举测试 (P)
  - 测试 JSON 往返、缺省字段兼容和枚举字符串稳定性。
  - 完成后，破坏字段名或枚举值的变化会被捕获。
  - _Requirements: 2, 5_
  - _Boundary: SerializationTests_
  - _Depends: 2.1, 2.2_

- [ ] 4.2 覆盖校验和边界测试 (P)
  - 测试基础校验、多错误返回、上游引用边界和 out-of-scope 未被实现。
  - 完成后，领域模型满足当前规格验收要求。
  - _Requirements: 1, 3, 4, 5_
  - _Boundary: ValidationTests_
  - _Depends: 3.2_

- [ ] 5. 运行最小验证
  - 执行当前领域包相关 `go test`。
  - 如项目已有格式化或 lint 命令，运行与本包相关的最小验证命令。
  - 完成后，记录验证结果和非本规格导致的剩余问题。
  - _Requirements: 5_
  - _Boundary: ValidationRun_
  - _Depends: 4.1, 4.2_
