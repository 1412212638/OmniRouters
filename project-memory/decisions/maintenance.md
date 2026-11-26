# 维护决策与约束

最后核对日期：2026-11-22。下列条目概括现有项目约定，详细要求和例外以来源为准。历史决策理由待核实。

## 镜像构建

默认采用源码级本地工作，由 GitHub Actions 在推送后构建镜像。工作流也支持标签和手动触发。这样安排已写入项目约定，原始决策背景待核实。来源：[AGENTS.md](../../AGENTS.md)、[工作流](../../.github/workflows/docker-image-ghcr.yml)。

## 前端入口

用户于 2026-11-22 确认，日常维护范围是新版前端 `web/default/` 和新版模型广场 catalog。修改前追踪路由及实际渲染组件，并在新版界面验证。catalog 卡片使用 `CatalogModelCard`，`components/model-card.tsx` 属于经典广场的独立定价组合。经典前端和经典广场只在用户明确要求或共享代码依赖确有需要时涉及。来源：[AGENTS.md](../../AGENTS.md)、[pricing/index.tsx](../../web/default/src/features/pricing/index.tsx)、[web-router.go](../../router/web-router.go)。

## 控制台侧边栏

2026-11-25 已在本地模拟预览核对个人与组织工作区。新版控制台将个人充值和订阅套餐归入财务，个人入口集中在账户与支持，组织空间的管理入口归入组织管理。账单导出入口显示为对账中心。菜单按 URL 读取原有管理员模块开关和个人可见性设置，展示分组使用独立标识。线上设置尚未核对。来源：[侧边栏分组](../../web/default/src/hooks/use-sidebar-data.ts)、[可见性配置](../../web/default/src/hooks/use-sidebar-config.ts)。

## 数据库和计费

数据库变更需要兼容 SQLite、MySQL 和 PostgreSQL。计费工作须遵循输入边界、额度饱和处理和审计等现有约束。表达式计费任务先阅读专门设计文档。来源：[AGENTS.md](../../AGENTS.md)、[表达式设计](../../pkg/billingexpr/expr.md)、[额度计算](../../common/quota_math.go)。

## 变更记录

源码级变更推送前需记录原因、范围、保留行为、验证、限制及提交和推送状态。来源：[AGENTS.md](../../AGENTS.md)、[上游同步日志](../../docs/upstream-sync-log.md)。

## 使用日志的输入输出 JSON

使用日志记录经过白名单筛选的输入参数和输出元数据。新增参数时按字段和值类型选择性记录，保留用户关闭记录的能力。请求提示词、媒体地址、密钥和生成内容不进入该摘要。管理员和用户视图使用同一份脱敏摘要。2026-11-24 已核对[后端实现](../../model/call_io_log.go)和[回归测试](../../model/call_io_log_test.go)。线上版本尚未独立核对。

新版通用使用日志的详情使用右侧抽屉。请求轨迹标签集中展示可独立展开的输入、输出 JSON，以及基于现有日志字段生成的请求元数据。管理员保留请求审计和原始数据标签。上下条切换以打开详情时的同页结果为准。2026-11-24 已核对[通用日志列表](../../web/default/src/features/usage-logs/components/usage-logs-table.tsx)、[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)及[请求轨迹](../../web/default/src/features/usage-logs/components/dialogs/request-trace-panel.tsx)，并在本地模拟数据中验证。线上版本尚未核对。

Token 明细的总量按不重叠的上游用量口径计算。Anthropic 的输入、输出与独立缓存读写相加，OpenAI 风格的缓存属于输入子集时只算输入加输出。缓存写入 TTL 子项不重复累加。OpenAI Chat Completions 转 Anthropic Messages 时，`input_tokens` 扣除缓存读写后再分别填入 Anthropic 缓存字段。流式首包若尚无上游 usage，则 `message_start.input_tokens` 为 0，避免把含缓存的本地估算报成未缓存输入；最终 usage 在 `message_delta` 中提供。2026-11-26 已核对[响应转换](../../relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go)、[回归测试](../../relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go)和[转换黄金样例](../../relaykit/relayconvert/testdata/golden/stream/openai_to_claude.golden.json)。

账单导出中心按选择的账期和文件语言异步生成个人或组织账单，支持简体中文和英文，旧任务默认英文。导出包包含 L0 PDF、L1 XLSX 汇总、L2 资源汇总 CSV 和 L3 模型账单明细 CSV，并保留七天。导出只读取已授权范围内的账务记录。单次导出最多覆盖 12 个月，财务流水与压缩归档仍有上限。用量日志分批读取并按日和报表维度汇总，25,000 条上限作用于聚合组，不再按请求条数限制。L3 按所选账期、模型和计费单位汇总，逐次请求可在使用日志查看。2026-11-25 已用本地 Go 1.26.10 和模拟 SQLite 数据核对[导出服务](../../service/billing_export.go)、[日志读取](../../model/billing_export_data.go)和[导出渲染](../../pkg/billexport/csv_rows.go)。生产数据库和亿级记录耗时尚未验证。

账单导出取提交时的全站显示模式、汇率和额度基数。USD、人民币、自定义货币的换算与钱包一致，实际订阅 Tokens 保持原单位，旧版原始额度模式单列。历史订单的币种缺失时按笔数列示。L1 分开列示用量、钱包入账调整和订单，L2 按入账月份汇总，L3 按模型列示计费用量、扣费、退款及净消耗。2026-11-25 已用模拟 SQLite 账目核对三种货币模式和跨级净消耗，见[导出显示口径](../../service/billing_export_display.go)、[汇总账单](../../pkg/billexport/xlsx.go)和[CSV 账单](../../pkg/billexport/csv_rows.go)。

管理员可在默认前端账单导出页搜索并选择用户，个人账单历史、生成和下载均以该用户为范围。任务记录同时保存发起管理员 ID，后台执行时重新验证管理员身份。组织账单仍按组织所有者或管理员权限校验。2026-11-25 已核对[控制器](../../controller/billing_export.go)、[任务授权](../../service/billing_export.go)和[账单导出页](../../web/default/src/features/billing-exports/index.tsx)。

异步任务按任务 ID 关联提交日志与结算日志。后端在差额记录中保存 `pre_consumed_quota` 和 `actual_quota`，日志自身的额度是差额。无差额时既有通用结算仍只写服务日志。新版抽屉按记录类型显示这些金额。2026-11-24 已核对[后端任务计费](../../service/task_billing.go)和[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)。

新版模型详情和通用使用日志共用阶梯文案，优先展示可解析的计费条件。日志匹配继续读取原始档位名，复杂条件只展示能可靠从现有详情推导出的信息。2026-11-24 已核对[展示逻辑](../../web/default/src/features/pricing/lib/tier-display-label.ts)，并用本地模拟数据验证两个入口。线上版本尚未核对。
