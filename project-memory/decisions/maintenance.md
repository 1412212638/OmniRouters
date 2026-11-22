# 维护决策与约束

最后核对日期：2026-11-22。下列条目概括现有项目约定，详细要求和例外以来源为准。历史决策理由待核实。

## 镜像构建

默认采用源码级本地工作，由 GitHub Actions 在推送后构建镜像。工作流也支持标签和手动触发。这样安排已写入项目约定，原始决策背景待核实。来源：[AGENTS.md](../../AGENTS.md)、[工作流](../../.github/workflows/docker-image-ghcr.yml)。

## 前端入口

修改前先追踪路由、主题和实际渲染组件。default 和 classic 各有代码路径，需要根据任务确定受影响的界面。模型广场 catalog 使用 `CatalogModelCard`，另有独立的 classic 定价组合。该约定有助于避免修改未被目标界面使用的组件。来源：[AGENTS.md](../../AGENTS.md)、[pricing/index.tsx](../../web/default/src/features/pricing/index.tsx)、[web-router.go](../../router/web-router.go)。

## 数据库和计费

数据库变更需要兼容 SQLite、MySQL 和 PostgreSQL。计费工作须遵循输入边界、额度饱和处理和审计等现有约束。表达式计费任务先阅读专门设计文档。来源：[AGENTS.md](../../AGENTS.md)、[表达式设计](../../pkg/billingexpr/expr.md)、[额度计算](../../common/quota_math.go)。

## 变更记录

源码级变更推送前需记录原因、范围、保留行为、验证、限制及提交和推送状态。来源：[AGENTS.md](../../AGENTS.md)、[上游同步日志](../../docs/upstream-sync-log.md)。
