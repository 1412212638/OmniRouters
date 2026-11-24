# 维护决策与约束

最后核对日期：2026-11-22。下列条目概括现有项目约定，详细要求和例外以来源为准。历史决策理由待核实。

## 镜像构建

默认采用源码级本地工作，由 GitHub Actions 在推送后构建镜像。工作流也支持标签和手动触发。这样安排已写入项目约定，原始决策背景待核实。来源：[AGENTS.md](../../AGENTS.md)、[工作流](../../.github/workflows/docker-image-ghcr.yml)。

## 前端入口

用户于 2026-11-22 确认，日常维护范围是新版前端 `web/default/` 和新版模型广场 catalog。修改前追踪路由及实际渲染组件，并在新版界面验证。catalog 卡片使用 `CatalogModelCard`，`components/model-card.tsx` 属于经典广场的独立定价组合。经典前端和经典广场只在用户明确要求或共享代码依赖确有需要时涉及。来源：[AGENTS.md](../../AGENTS.md)、[pricing/index.tsx](../../web/default/src/features/pricing/index.tsx)、[web-router.go](../../router/web-router.go)。

## 数据库和计费

数据库变更需要兼容 SQLite、MySQL 和 PostgreSQL。计费工作须遵循输入边界、额度饱和处理和审计等现有约束。表达式计费任务先阅读专门设计文档。来源：[AGENTS.md](../../AGENTS.md)、[表达式设计](../../pkg/billingexpr/expr.md)、[额度计算](../../common/quota_math.go)。

## 变更记录

源码级变更推送前需记录原因、范围、保留行为、验证、限制及提交和推送状态。来源：[AGENTS.md](../../AGENTS.md)、[上游同步日志](../../docs/upstream-sync-log.md)。

## 使用日志的输入输出 JSON

使用日志记录经过白名单筛选的输入参数和输出元数据。新增参数时按字段和值类型选择性记录，保留用户关闭记录的能力。请求提示词、媒体地址、密钥和生成内容不进入该摘要。管理员和用户视图使用同一份脱敏摘要。2026-11-24 已核对[后端实现](../../model/call_io_log.go)和[回归测试](../../model/call_io_log_test.go)。线上版本尚未独立核对。

新版通用使用日志的详情使用右侧抽屉。请求轨迹标签集中展示可独立展开的输入、输出 JSON，以及基于现有日志字段生成的请求元数据。管理员保留请求审计和原始数据标签。上下条切换以打开详情时的同页结果为准。2026-11-24 已核对[通用日志列表](../../web/default/src/features/usage-logs/components/usage-logs-table.tsx)、[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)及[请求轨迹](../../web/default/src/features/usage-logs/components/dialogs/request-trace-panel.tsx)，并在本地模拟数据中验证。线上版本尚未核对。

Token 明细的总量按不重叠的上游用量口径计算。Anthropic 的输入、输出与独立缓存读写相加，OpenAI 风格的缓存属于输入子集时只算输入加输出。缓存写入 TTL 子项不重复累加。2026-11-24 已核对[用量日志生成](../../service/text_quota.go)和[详情展示](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)。

异步任务按任务 ID 关联提交日志与结算日志。后端在差额记录中保存 `pre_consumed_quota` 和 `actual_quota`，日志自身的额度是差额。无差额时既有通用结算仍只写服务日志。新版抽屉按记录类型显示这些金额。2026-11-24 已核对[后端任务计费](../../service/task_billing.go)和[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)。

新版模型详情和通用使用日志共用阶梯文案，优先展示可解析的计费条件。日志匹配继续读取原始档位名，复杂条件只展示能可靠从现有详情推导出的信息。2026-11-24 已核对[展示逻辑](../../web/default/src/features/pricing/lib/tier-display-label.ts)，并用本地模拟数据验证两个入口。线上版本尚未核对。
