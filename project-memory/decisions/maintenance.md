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

## 分组定价的模型规则

新版分组定价将模型规则按计费组折叠显示。按模型名称搜索时自动展开命中的组，各组独立分页，每页 15 条。编辑与页面保存使用完整规则集合，覆盖通用倍率、客户专属倍率和有效期。2026-11-28 已在本地用模拟配置验证分页、搜索和筛选状态下的完整保存，线上界面待部署验证。来源：[计费设置入口](../../web/default/src/features/system-settings/billing/section-registry.tsx)、[模型规则编辑器](../../web/default/src/features/system-settings/models/group-ratio-visual-editor.tsx)、[保存逻辑](../../web/default/src/features/system-settings/models/ratio-settings-card.tsx)。

## 管理员资金入账记录

管理员详情在同一行显示入账额度和实付金额。实付金额按当前页的在线充值流水匹配同一用户的成功个人充值订单，易支付采用下单时的两位小数金额。历史订单缺少币种快照，显示订单记录的数值。赠送、调账以及缺少可靠实付数据的渠道订单显示 `—`。累计付费入账仍按入账额度计算。2026-11-28 已在本地 SQLite 核对折扣订单、分页及访问边界，并用模拟订单验证桌面和窄屏弹窗。线上待部署验证。来源：[充值下单](../../controller/topup.go)、[资金入账查询](../../model/wallet_funding.go)、[管理端弹窗](../../web/default/src/features/wallet/components/funding-ledger-dialog.tsx)。

## 渠道已使用量清零

新版渠道列表和卡片的「已使用」支持单渠道清零，使用 `channel.operate` 权限并记录管理审计。清零在数据库中同时归零计数并递增渠道用量版本。启用批量写入时，每个增量在入队时读取共享数据库版本，批次落库时只写入版本仍匹配的渠道。任一新版节点上的待写或已经取出的旧版本批次都会失效。直接写入按数据库更新顺序生效。渠道余额、用户结算及历史日志独立于该计数。2026-11-28 已在本地用两个 SQLite 连接和模拟浏览器页面验证；线上多节点和 MySQL、PostgreSQL 待核验。部署时需完成迁移并更新所有写入节点。来源：[渠道计数与版本](../../model/channel_used_quota.go)、[渠道权限路由](../../router/channel-router.go)、[新版渠道列表](../../web/default/src/features/channels/components/channels-columns.tsx)。

## 个人余额预警

新版个人资料按站点当前余额单位显示和编辑预警阈值，保存时转换回 `quota_warning_threshold` 内部配额。用户只修改其他偏好时保留已有阈值精度。结算后低于预警阈值的请求可以再次触发提醒。个人钱包和订阅预警使用同一通知事件限频，余额与订阅额度预警默认按用户在首次发送尝试起 24 小时内允许 1 次。其他通知类型默认 10 分钟内允许 2 次。`NOTIFY_LIMIT_COUNT` 和 `NOTIFICATION_LIMIT_DURATION_MINUTE` 仍可覆盖默认值。Redis 跨节点共享限频，内存模式按进程计数。2026-11-28 已核对模拟前端、并发限频、窗口到期和启动默认配置，线上配置待核验。来源：[个人预警设置](../../web/default/src/features/profile/components/tabs/notification-tab.tsx)、[结算预警](../../service/quota.go)、[限频实现](../../service/notify-limit.go)、[默认配置](../../common/init.go)。

## 变更记录

源码级变更推送前需记录原因、范围、保留行为、验证、限制及提交和推送状态。来源：[AGENTS.md](../../AGENTS.md)、[上游同步日志](../../docs/upstream-sync-log.md)。

## 使用日志的输入输出 JSON

使用日志记录经过白名单筛选的输入参数和输出元数据。新增参数时按字段和值类型选择性记录，保留用户关闭记录的能力。请求提示词、媒体地址、密钥和生成内容不进入该摘要。管理员和用户视图使用同一份脱敏摘要。2026-11-24 已核对[后端实现](../../model/call_io_log.go)和[回归测试](../../model/call_io_log_test.go)。线上版本尚未独立核对。

新版通用使用日志的详情使用右侧抽屉。请求轨迹标签集中展示可独立展开的输入、输出 JSON，以及基于现有日志字段生成的请求元数据。管理员保留请求审计和原始数据标签。上下条切换以打开详情时的同页结果为准。2026-11-24 已核对[通用日志列表](../../web/default/src/features/usage-logs/components/usage-logs-table.tsx)、[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)及[请求轨迹](../../web/default/src/features/usage-logs/components/dialogs/request-trace-panel.tsx)，并在本地模拟数据中验证。线上版本尚未核对。

Token 明细的总量按不重叠的上游用量口径计算。Anthropic 的输入、输出与独立缓存读写相加，OpenAI 风格的缓存属于输入子集时只算输入加输出。缓存写入 TTL 子项不重复累加。OpenAI Chat Completions 转 Anthropic Messages 时，`input_tokens` 扣除缓存读写后再分别填入 Anthropic 缓存字段。流式首包若尚无上游 usage，则 `message_start.input_tokens` 为 0，避免把含缓存的本地估算报成未缓存输入。最终 usage 在 `message_delta` 中提供。成功结束的流要调用幂等 finalizer，以便在上游未发送 `finish_reason` 时结束仍打开的内容块并补发 `message_delta`、`message_stop`。失败流不合成成功终止事件。2026-11-26 已核对[响应转换](../../relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp.go)、[回归测试](../../relaykit/relayconvert/internal/oai_chat/to_claude_messages_resp_test.go)和[转换黄金样例](../../relaykit/relayconvert/testdata/golden/stream/openai_to_claude.golden.json)。流尾收束由 [OpenAI relay handler](../../relay/channel/openai/helper.go) 调用。2026-11-26 使用本机缓存的 Go 1.26.10 离线通过了 11 个定向转换用例和 [Claude 流收尾用例](../../relay/channel/openai/claude_stream_test.go)。整包测试遇到 Responses 请求测试的编译错误及 OpenAI Responses 流生命周期测试的断言失败。

2026-11-26 对线上 `/v1/messages` 使用 `cache_control.type=ephemeral` 和稳定长前缀重复发起请求。`deepseek-v4.1-flash` 与 `doubao-seed-2-1-pro` 在非流式响应及流式 `message_delta` 均观察到 `cache_read_input_tokens > 0`。`glm-5.3` 的非流式响应也观察到缓存读取。命中样本中的 `input_tokens`、`cache_read_input_tokens` 与 `cache_creation_input_tokens` 之和等于响应内的 `billing_usage.openai_usage.prompt_tokens`。

`glm-5.3` 的流式样本生成了文本，但 `message_delta` 中的 `cache_read_input_tokens` 和 `output_tokens` 均为 0，需核对实际渠道和上游原始 SSE 才能定位原因。`doubao-seed-2-1-lite` 的两种模式本轮均为 `cache_read_input_tokens=0`。四个模型的线上样本均为 `cache_creation_input_tokens=0`，非零写入仍待覆盖。

2026-11-26 用户切换 `glm-5.3` 上游渠道后，使用同一稳定长前缀复测线上 `/v1/messages`。非流式重复请求和两次流式 `message_delta` 均命中缓存，典型样本为 `input_tokens=19`、`cache_read_input_tokens=10816`、`cache_creation_input_tokens=0`、`output_tokens=3`，输入字段合计与 `billing_usage.openai_usage.prompt_tokens=10835` 一致。流式两次都有文本、`message_delta` 和 `message_stop`，`message_start` 的用量仍为 0。另一组流式样本有一次缓存读取为 0，仍正常返回非零输出用量。测试表明渠道切换后旧的流式用量异常未复现。公开响应无渠道标识，管理员日志查询未授权，旧渠道原始 SSE 和两次实际路由仍待核对。

2026-11-26 用户再次切换 `glm-5.3` 渠道后，沿用上一轮测试前缀请求线上 `/v1/messages`，两次非流式与两次流式均返回 HTTP 200。首次非流式 `input_tokens=595`、`cache_read_input_tokens=10240`，重复非流式和两次流式 `message_delta` 都为 `input_tokens=0`、`cache_read_input_tokens=10835`。每条最终用量的输入三项之和都与 `billing_usage.openai_usage.prompt_tokens=10835` 一致。两次流式分别返回 `output_tokens=2` 和 `15`，均有文本、`message_delta` 和 `message_stop`。`message_start` 用量仍全部为 0，本轮也未观察到非零缓存写入。该结果来自网关响应，实际路由渠道和上游原始 SSE 仍需日志确认。

2026-11-26 用户确认，`glm-5.3` 渠道 1 配置为火山类型，渠道 3 是从渠道 1 切换为 OpenAI 类型后复测的配置。渠道 1 非流式样本有缓存读取，流式样本的最终缓存读取和输出用量为 0。渠道 2、3 的非流式和流式复测均有正确拆分缓存的样本。渠道 1 的 Base URL、出站请求和原始上游 SSE 仍待核对。

2026-11-26 核对[火山方舟 Chat API](https://docs.volcengine.com/docs/ark/chat-api?lang=zh)与[火山适配器](../../relay/channel/volcengine/adaptor.go)。普通火山路径将 `/v1/messages` 转发为 Chat Completions，现在会保留 `stream_options.include_usage=true` 来请求上游在流尾提供用量。特殊 Base URL 按原生 Messages 处理。共享 OpenAI 流处理器在上游省略 usage 时，从实际收到的文本和工具调用估算输出，工具调用估算计入总量。缓存读取和写入的准确值依赖上游实报。[请求](../../relay/channel/volcengine/adaptor_test.go)和[流响应](../../relay/channel/openai/claude_stream_test.go)定向回归已通过。渠道 1 使用的具体分支、可能存在的空或部分 usage 对象、以及修复后的线上表现仍待复测。

账单导出中心按选择的账期和文件语言异步生成个人或组织账单，支持简体中文和英文，旧任务默认英文。导出包包含 L0 PDF、L1 XLSX 汇总、L2 资源汇总 CSV 和 L3 模型账单明细 CSV，并保留七天。导出只读取已授权范围内的账务记录。单次导出最多覆盖 12 个月，财务流水与压缩归档仍有上限。用量日志分批读取并按日和报表维度汇总，25,000 条上限作用于聚合组，不再按请求条数限制。L3 按所选账期、模型和计费单位汇总，逐次请求可在使用日志查看。2026-11-25 已用本地 Go 1.26.10 和模拟 SQLite 数据核对[导出服务](../../service/billing_export.go)、[日志读取](../../model/billing_export_data.go)和[导出渲染](../../pkg/billexport/csv_rows.go)。生产数据库和亿级记录耗时尚未验证。

账单导出取提交时的全站显示模式、汇率和额度基数。USD、人民币、自定义货币的换算与钱包一致，实际订阅 Tokens 保持原单位，旧版原始额度模式单列。历史订单的币种缺失时按笔数列示。L1 分开列示用量、钱包入账调整和订单，L2 按入账月份汇总，L3 按模型列示计费用量、扣费、退款及净消耗。2026-11-25 已用模拟 SQLite 账目核对三种货币模式和跨级净消耗，见[导出显示口径](../../service/billing_export_display.go)、[汇总账单](../../pkg/billexport/xlsx.go)和[CSV 账单](../../pkg/billexport/csv_rows.go)。

管理员可在默认前端账单导出页搜索并选择用户，个人账单历史、生成和下载均以该用户为范围。任务记录同时保存发起管理员 ID，后台执行时重新验证管理员身份。组织账单仍按组织所有者或管理员权限校验。2026-11-25 已核对[控制器](../../controller/billing_export.go)、[任务授权](../../service/billing_export.go)和[账单导出页](../../web/default/src/features/billing-exports/index.tsx)。

异步任务按任务 ID 关联提交日志与结算日志。后端在差额记录中保存 `pre_consumed_quota` 和 `actual_quota`，日志自身的额度是差额。无差额时既有通用结算仍只写服务日志。新版抽屉按记录类型显示这些金额。2026-11-24 已核对[后端任务计费](../../service/task_billing.go)和[详情组件](../../web/default/src/features/usage-logs/components/dialogs/details-dialog.tsx)。

新版模型详情和通用使用日志共用阶梯文案，优先展示可解析的计费条件。日志匹配继续读取原始档位名，复杂条件只展示能可靠从现有详情推导出的信息。2026-11-24 已核对[展示逻辑](../../web/default/src/features/pricing/lib/tier-display-label.ts)，并用本地模拟数据验证两个入口。线上版本尚未核对。
