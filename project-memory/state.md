# 项目状态

最后核对日期：2026-11-22。以下描述仓库内容，实际部署状态需要单独核实。

## 已确认

- 项目是 Go API 网关，后端按 router、controller、service、model 分层，前端包含 default 和 classic 两套主题。来源：[项目约定](../AGENTS.md)。
- 用户于 2026-11-22 确认维护新版前端和新版模型广场。用户提供的系统设置截图显示这两项已选中，截图反映当时所示实例的设置。来源：[维护决策](decisions/maintenance.md)。
- Web 路由加载两套前端构建目录并使用主题感知文件系统。来源：[web-router.go](../router/web-router.go)。
- default 主题的模型广场入口会依据 `model_square_theme` 选择 catalog 或 classic 定价组合。catalog 卡片由入口文件中的 `CatalogModelCard` 渲染。来源：[pricing/index.tsx](../web/default/src/features/pricing/index.tsx)。
- 仓库中的初始设置为前端主题 classic、模型广场 catalog。来源：[theme.go](../setting/system_setting/theme.go)。
- GitHub Actions 工作流处理镜像构建与推送。来源：[docker-image-ghcr.yml](../.github/workflows/docker-image-ghcr.yml)。

## 待核实

- 实际运行实例的主题、模型广场设置、部署版本和镜像状态。
- 既有架构及维护决策的历史背景和负责人。现有规则与代码可以确认当前约束，无法单独证明当初的决策过程。
