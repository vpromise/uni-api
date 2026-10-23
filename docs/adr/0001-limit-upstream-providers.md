# 仅保留 OpenAI 与 Anthropic 上游接入

本 fork 选择缩小供应商范围以降低维护成本，仅保留 OpenAI API Key、Codex 订阅账号和 Anthropic API Key 三种接入方式。其他供应商的接入实现从源码中完整删除，并清理其专属界面、配置和依赖；服务端同样限制可用渠道，不能仅隐藏前端选项。

三种接入方式均保留自定义 Base URL，以支持自有反向代理。供应商范围约束内置实现和渠道类型，不以官方域名白名单限制实际请求地址。

客户端文本协议保留 OpenAI Chat Completions、OpenAI Responses 和 Anthropic Messages，以及三者之间已有的请求、响应和流式转换能力。删除 Gemini 客户端接口及其专属转换代码，范围包括宿主网关和独立的 RelayKit 模块。

保留 OpenAI 现有的图片、音频、嵌入、Realtime 和 Sora 视频能力。Sora 通过 OpenAI 渠道及对应任务插件链路提供；保留其依赖的通用任务与插件框架，删除其他供应商的内置任务插件。

本轮只收缩供应商范围，用户管理、登录、API 密钥、额度、套餐、日志和统计等业务功能的去留另行决定。实现按全新安装处理，不承担已有渠道、用户、密钥或历史日志的数据兼容与迁移要求。

验收覆盖三种接入方式、保留协议的同步与流式调用及已有转换行为，并验证其他渠道类型无法创建或转发。执行受影响 Go 包的测试与构建、前端所需检查，以及 RelayKit、TokenKit 的独立模块构建；如改变数据库行为，仍须按项目规则验证真实 SQLite、MySQL 和 PostgreSQL，全新安装范围不免除数据库兼容要求。

## 验证记录

- 根模块执行 `go test ./...` 通过；最终 JSON 包装调用与渠道状态调整另以 `go test -p 2 ./controller ./model -run 'Reset|Recover|Forgot|SaveStatusState|UpdateRetainedChannel|DisableThirdPartyPlugin'` 定向验证。
- 在 `relaykit/` 和 `tokenkit/` 内分别执行 `GOWORK=off go build ./...` 与 `GOWORK=off go test ./...`，均通过。
- 前端 `bun run typecheck`、`bun run test`（195 个文件、2175 个用例）和 `bun run build` 通过；Vitest 的 `maxWorkers` 设为 2，限制本地测试资源占用。
- 全仓前端 lint 存在原有错误；对修改文件与 HEAD 的诊断对照没有新增错误（当前 36 项、原版本 37 项）。
- SQLite 3.50.4、MySQL 8.4.13、PostgreSQL 17.11（Homebrew）均通过全新主库、独立日志库及重复启动迁移检查。临时验证执行 `go run /tmp/uni-api-db-check/main.go sqlite`、`mysql`、`postgres`，每个引擎启动迁移两次并验证三种渠道记录保留。
- 设置本地 `TEST_MYSQL_DSN`、`TEST_POSTGRES_DSN` 后执行 `go test ./controller -run 'Test(ModelPricingConversionDatabaseMatrix|VendorManagementDatabaseMatrix)$' -count=1 -v`，三个数据库均通过；分别设置 `TEST_TASK_DB_DIALECT=sqlite/mysql/postgres` 执行 `TestDisableThirdPartyPluginSupportsCascadeAndForce`，验证共享渠道解绑而专用内部渠道停用，均通过。
- 临时数据库实例与数据目录在验证后停止并删除，仅保留小体积验证记录。收尾阶段测试和构建串行执行，结束后清理临时工具、构建产物和可重建的 Go 编译缓存。

## 审查记录

Standards 轴发现的渠道权限注册、通用任务路由、有效测试覆盖、透传状态以及共享插件绑定／解绑问题已修复。Spec 轴发现的 Claude 数组系统提示转换、渠道局部更新和多密钥追加问题已修复，相关行为补入现有测试文件。

凭据入口按 [OWASP Authentication Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html)、[Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html) 和 [ASVS 5.0.0](https://asvs.owasp.org/) 检查本次改动边界：服务端拒绝被移除的渠道及插件认证方式，保留 Bearer／API Key 入口和渠道敏感操作权限，错误不包含上游密钥。现有身份认证测试和插件路由／绑定回归通过；未使用真实上游凭据发起联网调用，本文不表示对整个系统进行 ASVS 合规认证。

资源收尾：清理约 4.05 GiB Go 编译缓存、约 409 MiB 临时数据库与工具缓存，以及约 66 MiB 构建产物；验证日志合计约 428 KiB。依赖下载缓存和前端依赖保留，便于后续开发。
