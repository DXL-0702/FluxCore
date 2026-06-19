# FluxCore 项目规划

> 当前项目处于规划阶段。所有任务状态均为 **未开始**，需经逐步讨论确认后推进。

## 项目定位

FluxCore 的功能落点不是传统项目管理工具，而是面向个人开发者与多项目并发开发者的 **Git 原生研发状态记录系统**。

核心目标是把开发者已经自然产生的 `commit`、`push`、README 变更、分支、任务 ID 与开发意图，自动转化为项目状态、任务进度、活动日志、实时仪表盘与多项目上下文。

产品边界应保持清晰：FluxCore 不替代 Jira、Linear、Notion 或 CI 平台，而是让代码流自动变成项目流。

## 技术栈总览（已确定）

| 层级 | 技术选型 | 职责 |
| :--- | :--- | :--- |
| CLI | Go (Cobra) | Git Hook 注入、本地上下文管理、项目切换 |
| 后端 | Go (Gin) | REST API、Webhook 处理、WebSocket 广播 |
| ORM | GORM | 数据持久化，SQLite/PostgreSQL 动态切换 |
| 数据库 | SQLite / PostgreSQL | 开发 SQLite（零配置），生产可切 PostgreSQL |
| 消息 | Redis | WebSocket Pub/Sub 实时事件分发 |
| 前端 | React 19 + Vite 6 | SPA 仪表盘 |
| 样式 | Tailwind CSS | 原子化 CSS |

## 开发路线

v0.1 开发前需先确认 [ARCHITECTURE.md](ARCHITECTURE.md) 中的架构基线，再进入具体代码实现。

### 阶段一 — 基础设施与无感绑定

> 目标：搭建基础环境，打通 CLI ↔ 后端连接，实现项目本地绑定。

**后端**
- [ ] 初始化 Go 模块，引入 Gin 框架
- [ ] 配置 GORM，实现基于 `DB_TYPE` 环境变量的 SQLite/PostgreSQL 动态切换
- [ ] 定义 `Project`、`Repository`、`User`、`Config` 基础数据模型
- [ ] 实现 `POST /api/projects`、`GET /api/projects` 与项目仓库绑定接口
- [ ] 添加健康检查与基础认证中间件

**CLI**
- [ ] 实现 `fluxcore init` — 在项目中初始化 `.fluxcore/` 配置目录
- [ ] 实现 `fluxcore link` — 读取 `.git/config`，生成本地项目映射
- [ ] 添加 `fluxcore status` — 显示当前项目绑定状态

**前端**
- [ ] 搭建 React + Vite + Tailwind 基础框架
- [ ] 构建项目列表视图与空状态页
- [ ] 实现项目创建弹窗

### 阶段二 — Git 事件驱动与自动化日志

> 目标：实现“推送即记录”，建立 FluxCore 的核心数据来源。

**后端**
- [ ] 开发本地事件接收接口，接收 CLI 从 Git Hook 上报的 commit 事件
- [ ] 开发 Commit Message 解析器（正则匹配 `#TaskID`、Conventional Commit 类型）
- [ ] 从解析结果自动创建结构化日志条目
- [ ] 实现 `Event` 模型，作为 commit、task、README、CI、intent 等后续事件的统一事实表
- [ ] 实现 `Task` 模型与状态机（Open → In Progress → Testing → Done）

**CLI**
- [ ] 通过 `fluxcore init` 自动注入 `post-commit` Hook
- [ ] 添加 `fluxcore log` — 终端查看近期活动

**前端**
- [ ] 构建项目详情页，展示任务列表与日志时间轴
- [ ] 添加提交类型标签（feat / fix / refactor / chore）

### 阶段三 — README 驱动同步与实时推送

> 目标：引入 Redis，实现文档变更即时反映。

**后端**
- [ ] 集成 Redis，建立 WebSocket 广播管线
- [ ] 监听推送中的 `README.md` 变更，解析 Front Matter（`status`、`version`、`description`）
- [ ] 自动从 README 元数据更新项目卡片

**前端**
- [ ] 接入 WebSocket，实现项目卡片与日志流的无刷新实时更新
- [ ] 添加实时状态指示器（在线 / 同步中 / 离线）
- [ ] 实现事件到达 Toast 通知

### 阶段四 — 多项目并发与开发态势监控

> 目标：解决多项目、多分支、多任务并行开发下的上下文切换、状态判断与风险识别问题。

**CLI**
- [ ] 实现 `fluxcore switch <project>` — 自动切换工作目录并恢复上下文
- [ ] 添加 `fluxcore dashboard` — 从终端打开 Web UI

**前端**
- [ ] 构建全局仪表盘，聚合所有项目实时动态流
- [ ] 集成 CI 状态展示（构建成功/失败徽标）
- [ ] 添加项目级筛选、排序与搜索
- [ ] 构建 Workstream Radar（开发流雷达）视图

### 阶段五 — 智能化与生态完善

> 目标：引入 AI 辅助，完善部署与通知生态。

- [ ] **AI 集成** — 分析 `git diff` 自动生成工作日报摘要
- [ ] **插件系统** — 支持自定义 Webhook 处理器与事件处理器
- [ ] **一键部署** — 提供 Docker Compose 一键部署方案
- [ ] **通知推送** — 可选的 Slack / Discord / 邮件关键事件推送

## 产品创新方向

### Intent Snapshot（开发意图快照）

> 核心定位：FluxCore 不只记录“做了什么”，还应保留“为什么做”。该方向作为阶段二之后、阶段五 AI 能力之前的增强层，不影响当前 MVP 推进顺序。

**概念说明**
- 开发者可在开始一个任务、分支或阶段时，记录一条轻量开发意图，例如：`fluxcore intent "重构认证模块，拆分 token 校验和用户加载逻辑"`。
- FluxCore 后续自动关联该意图与当前分支、commit、diff、README 变更、任务状态和活动日志。
- Web UI 展示时，不只呈现零散 commit，而是形成“开发意图 → 实际变更 → 当前状态 → 阶段总结”的连续上下文。

**正向价值**
- 强化 FluxCore 的产品差异化：从 commit 日志看板升级为个人开发记忆系统。
- 适配多项目并发场景：开发者切回旧项目时，可快速理解上次修改的目标、范围与未完成事项。
- 为后续 AI 摘要提供高质量上下文：AI 不只基于 diff 推断，还能结合明确的开发意图生成更准确的日报、阶段总结和风险提示。

**潜在风险**
- 若交互设计过重，可能重新变成手动填报，违背“无感记录”的核心理念。
- 需要清晰定义 intent 与 branch、commit、task 的关联规则，否则容易出现上下文错配。
- 早期不宜引入复杂 AI 判断，应先以轻量命令和确定性关联规则落地。

### Workstream Radar（开发流雷达）

> 核心定位：FluxCore 不只展示多项目动态，而是识别多并发开发流的状态、风险和恢复入口。

**概念说明**
- 每个活跃项目、分支、任务或开发意图都可抽象为一个 `Workstream`。
- 核心对象从 `Project` 扩展为 `Project → Branch → Task → Intent → Event`。
- 全局视图不只是项目卡片墙或 commit 活动流，而是回答“现在最应该关注哪个开发流，为什么”。

**状态模型**
- `Active`：近期有 commit、push、README 或任务状态变化。
- `Idle`：一段时间无进展，但仍存在未关闭任务、未合并分支或未完成 intent。
- `Blocked`：存在失败 CI、冲突风险、缺少后续动作或关键状态卡住。
- `Diverged`：分支明显落后主干，继续开发或合并前需要同步。
- `Ready to Resume`：可恢复开发，并能展示上次意图、最后变更与下一步建议。
- `Ready to Merge`：功能完成度高，测试或合并是下一动作。

**正向价值**
- 从“展示日志”升级到“判断并发状态”，强化多项目场景下的产品价值。
- 帮助开发者在项目切换时快速恢复上下文，减少多项目并行带来的认知成本。
- 与 Intent Snapshot 天然结合，可判断实际变更是否仍符合最初开发意图。
- 为后续 AI 能力提供明确落点：解释风险、总结上下文、建议下一步。

**潜在风险**
- 状态规则需要足够确定，避免因误报导致用户不信任雷达结果。
- 早期不宜引入复杂评分系统，否则会拖慢 MVP。
- UI 需要保持克制，避免变成复杂运维大屏或噪声密集的信息墙。

## 目标目录结构

```
FluxCore/
├── cli/                  # CLI 工具 (Go + Cobra)
│   ├── cmd/              # 命令定义
│   ├── internal/         # Hook 注入、配置管理
│   └── main.go
├── server/               # 后端服务 (Go + Gin)
│   ├── api/              # HTTP 路由处理
│   ├── model/            # GORM 数据模型
│   ├── service/          # 业务逻辑层
│   ├── ws/               # WebSocket Hub
│   ├── db/               # 数据库连接 (SQLite/PG)
│   └── main.go
├── web/                  # 前端 (React + Vite)
│   ├── src/
│   │   ├── components/   # 可复用 UI 组件
│   │   ├── pages/        # 路由页面
│   │   ├── hooks/        # 自定义 React Hooks
│   │   └── lib/          # API 客户端、WebSocket 客户端
│   └── index.html
├── migrations/           # 数据库迁移文件
├── docker-compose.yml    # 生产部署配置
└── .fluxcore/            # 本地配置 (已 gitignore)
```

## 下一步建议

当前所有阶段均处于 **未开始** 状态。建议从 **阶段一** 入手，按以下顺序推进：

1. **先确定技术选型细节**：例如 Go 模块路径命名、前端包管理器选用 npm/pnpm/yarn、数据库迁移工具选择。
2. **再并行启动后端框架搭建与前端脚手架**：两者无强依赖，可独立进行。
3. **CLI 模块随后接入**：需等后端有基础 API 后才具备实际联调意义，但框架可先行搭建。
