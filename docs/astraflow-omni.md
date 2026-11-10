# AstraFlow Omni 视频生成

AstraFlow 插件自 1.1.0 新增 `gemini-omni-1.2-flash`，当前版本为 1.1.1。已有的视频增强模型
`ai-mediakit-video-enhancement-std` 保留原有分辨率、FPS、按秒定价和 10 秒预扣。
如果后台已经安装过旧版 AstraFlow 插件，更新至 1.1.1 以使用最新的定价标签。

## 渠道与请求

沿用 AstraFlow 插件渠道，填写服务商基础地址和渠道密钥，在渠道中添加模型
`gemini-omni-1.2-flash`。默认上游模型名也为该名称，不会改写成文档示例的 1.1。

客户端仅替换服务商域名和 API Key，保留原生路径、方法和请求体：

```http
POST /v1beta/models/interactions
Authorization: Bearer 平台API_KEY
Content-Type: application/json
```

```json
{
  "model": "gemini-omni-1.2-flash",
  "input": "生成一段介绍故宫的视频",
  "generation_config": {
    "video_config": { "task": "text_to_video" }
  },
  "response_format": { "type": "video", "aspect_ratio": "16:9" },
  "background": true
}
```

支持 `text_to_video`、`image_to_video`、`reference_to_video`、`edit`。
`input` 可以是文字，也可以是上游多模态或多轮步骤数组；图片和视频使用原生
`type`、`data`（原始 Base64）、`mime_type`。`thought.signature`、`model_output`、
`user_input` 和其他原生配置继续传递给上游。`response_format` 可省略。

`background: true` 异步提交；省略或显式传 `false` 时采用上游同步方式，
同步完成直接保存终态并结算，不再轮询同一已完成任务。

```http
GET /v1beta/models/interactions/平台返回的任务ID
Authorization: Bearer 平台API_KEY
```

异步由后台轮询后更新本地结果。兼容上游的 `in_progress`、`processing`、
`completed`、`failed` 状态。对客户端返回平台任务 ID 和请求使用的模型名；
完成结果中的 `steps`、`usage` 和 Base64 视频保留。

## 定价

在任务插件定价中选择 AstraFlow / `gemini-omni-1.2-flash`。
配置三个价格，单位均为 **USD / 1M Tokens**：

| 价格项 | 示例价格 | 结算数据 |
| --- | ---: | --- |
| 输入 Tokens | 1.5 | `usage.total_input_tokens` |
| 文本输出 Tokens | 9 | `output_tokens_by_modality` 中 `text` 的 Tokens 总数 |
| 视频输出 Tokens | 12 | `output_tokens_by_modality` 中 `video` 的 Tokens 总数 |

可视化价格编辑器支持直接填写三个单价。表达式编辑器等价配置：

```text
tier("base", u("input_tokens") * 1.5 / 1000000 + u("text_output_tokens") * 9 / 1000000 + u("video_output_tokens") * 12 / 1000000)
```

以上是示例售价，系统不会自动替管理员写入价格。
分组和客户专属倍率继续在既有计费路径生效。
`total_output_tokens` 用于核对模态总数；`total_tokens`、思考和工具用量保留在结果里，
不额外叠加收费。输入按 `total_input_tokens` 收费，不依赖可能不完整的输入模态明细。

提交时预估输入 65,536、文本输出 4,096、视频输出 65,536 Tokens 进行预扣；
这些是预扣估值，不是生成时长或输出上限，客户端无需增加估算字段。
成功后按实际用量在原有结算机制中多退少补，失败按原有任务退款规则处理。
如果完成结果缺少用量、缺少输出模态明细、模态总数不一致或包含非法数量，
保留预扣并记录异常，不把全部输出当作视频、不按零费用结算。

模型广场当前默认卡片显示输入、文本输出、视频输出三项价格；
详情表显示同样的三个价格，并使用当前客户实际适用的倍率。

## 视频内容与边界

上游以 Base64 返回视频。平台保存原生结果，任务列表省略图片、视频、音频的
`data` 内容以避免巨大的列表响应；原生查询仍返回完整内容。
任务日志可通过既有“预览”入口播放视频。兼容的内容接口为：

```http
GET /v1/videos/平台任务ID/content
Authorization: Bearer 平台API_KEY
```

站内预览使用已有的短期媒体访问凭证。内容响应支持 Range/HEAD，且不公开缓存。
输入最多 10 张图片、3 段视频，Base64 媒体总长度最多 48 MiB；
上游响应 JSON 最大 64 MiB，仍受部署自身请求体限制和数据库容量限制。
输入视频时长等媒体规则继续由上游检查；平台不会声称已解析 Base64 视频时长。
这些较大的响应限制只适用于 AstraFlow Omni，其他插件的既有限制保持不变。

本地使用文档样例和合成响应验证，尚未进行真实上游付费调用。
