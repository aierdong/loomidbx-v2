# API Client 边界

本目录是前端调用本地能力的统一入口，用于封装 Wails generated bindings 并转换错误结果。

页面、组件和 store 不应直接依赖 `frontend/generated/` 里的 generated 绑定；后续接入真实绑定时，只允许 API client 直接引用 generated 目录。
