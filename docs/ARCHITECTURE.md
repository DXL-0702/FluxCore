# FluxCore v0.1 架构基线

> 本文档用于在进入阶段一开发前固化最小架构边界。目标不是完整详细设计，而是避免核心模型、事件链路和模块职责在开发中反复返工。

## 架构目标

v0.1 的目标是跑通个人开发者本地多项目研发状态记录闭环：

```text
Git Repository → Git Hook → fluxcore CLI → Backend API → Database → Web UI
```

v0.1 优先保证本地开发链路稳定，不优先接入远程 Git 平台 Webhook。远程 Webhook 后续作为增强能力，用于获得更权威的 push、merge、CI 与远端仓库事件。

## 设计原则

- **Local-first**：优先服务个人开发者本地多项目并发场景，降低部署和接入成本。
- **Event-first**：底层以事件流记录开发事实，任务、日志、开发流状态均从事件派生。
- **Deterministic-first**：早期状态判断使用确定性规则，AI 只做总结、解释和建议，不作为底层状态真值。
- **MVP 克制**：只实现阶段一、阶段二所需的最小闭环，为 Intent Snapshot 与 Workstream Radar 预留结构，不提前实现复杂能力。

## 事件链路

### v0.1 主链路

```text
git commit
  ↓
post-commit hook
  ↓
fluxcore CLI hook handler
  ↓
POST /api/events
  ↓
Backend event service
  ↓
SQLite
  ↓
Web UI query / later WebSocket push
```

**职责边界**
- Git Hook：只负责触发 CLI，不承担业务判断。
- CLI：读取仓库路径、当前分支、commit SHA、remote URL、本地配置 token，并发送事件。
- Backend：校验 token，接收事件，解析 commit message，写入结构化数据。
- Web UI：读取后端聚合结果，不直接访问本地 Git 仓库。

### 延后链路

远程 Webhook 不进入 v0.1 主路径，后续用于补充：

- GitHub / GitLab / Gitea push 事件
- Pull Request / Merge Request 状态
- CI 构建结果
- 远端默认分支变化

这样可以避免 MVP 阶段同时处理平台适配、Webhook 鉴权、公网回调和本地服务暴露问题。

## 核心领域模型

### Project

项目实体，表示用户关心的产品、工具或研发对象。

关键字段：
- `id`
- `name`
- `description`
- `status`
- `created_at`
- `updated_at`

### Repository

Git 仓库实体。v0.1 可按一个 Project 对应一个 Repository 实现，但模型上保留多仓扩展能力。

关键字段：
- `id`
- `project_id`
- `name`
- `local_path`
- `remote_url`
- `default_branch`
- `created_at`
- `updated_at`

### Branch

仓库分支状态。用于后续判断开发流是否活跃、停滞或偏离主干。

关键字段：
- `id`
- `repository_id`
- `name`
- `head_sha`
- `base_branch`
- `last_seen_at`
- `created_at`
- `updated_at`

### Task

任务实体。v0.1 通过 commit message 中的 `#TaskID` 或后续手动创建关联。

关键字段：
- `id`
- `project_id`
- `external_ref`
- `title`
- `status`
- `source`
- `created_at`
- `updated_at`

状态建议：

```text
Open → In Progress → Testing → Done
```

### Event

事件实体。Event 是 FluxCore 的底层事实表，其他视图和状态尽量从事件派生。

关键字段：
- `id`
- `project_id`
- `repository_id`
- `branch_name`
- `event_type`
- `source`
- `commit_sha`
- `message`
- `payload`
- `occurred_at`
- `received_at`

v0.1 最小事件类型：
- `project_created`
- `repository_linked`
- `commit_observed`
- `task_detected`

后续预留事件类型：
- `readme_updated`
- `ci_status_changed`
- `intent_created`
- `branch_diverged`
- `workstream_state_changed`

### User 与 Config

v0.1 不实现完整注册登录系统。

- `User` 可保留为单用户占位模型，不进入复杂权限体系。
- `Config` 用于服务端运行配置与本地 CLI 配置映射。
- 认证使用本地单用户 token，避免过早引入账号系统。

## Workstream 预留设计

Workstream 不建议在 v0.1 作为独立强实体实现，先作为可计算视图预留。

```text
Workstream = Project + Repository + Branch + optional Task + optional Intent + latest Events
```

后续 Workstream Radar 可基于确定性规则判断：

- `Active`：近期有 commit、push、README 或任务状态变化。
- `Idle`：一段时间无进展，但仍存在未关闭任务、未合并分支或未完成 intent。
- `Blocked`：存在失败 CI、冲突风险、缺少后续动作或关键状态卡住。
- `Diverged`：分支明显落后主干，继续开发或合并前需要同步。
- `Ready to Resume`：可恢复开发，并能展示上次意图、最后变更与下一步建议。
- `Ready to Merge`：功能完成度高，测试或合并是下一动作。

v0.1 只需保证事件、分支、任务数据足以支撑后续计算，不实现完整雷达算法。

## 阶段一 API 边界

阶段一 API 只服务基础绑定与可见性。

必需接口：
- `GET /health`
- `POST /api/projects`
- `GET /api/projects`
- `POST /api/projects/:project_id/repositories`
- `GET /api/projects/:project_id/repositories`

阶段二再引入：
- `POST /api/events`
- `GET /api/projects/:project_id/events`
- `GET /api/projects/:project_id/tasks`

接口原则：
- CLI 只调用后端 API，不直接写数据库。
- Web UI 只调用后端 API，不读取 CLI 本地配置。
- 后端负责统一生成数据库 ID、校验 token、维护事件一致性。

## 数据库策略

v0.1 使用 SQLite 作为默认数据库，保持零配置启动。

PostgreSQL 从模型和 GORM 配置层面预留：

- 通过 `DB_TYPE=sqlite|postgres` 切换。
- 避免使用 SQLite 独有 SQL 特性。
- 主键、时间字段、JSON 字段需要选择兼容 GORM 的写法。

迁移策略：
- 阶段一可使用 GORM AutoMigrate。
- 进入多环境部署或 PostgreSQL 前，再引入显式迁移工具。

## 认证策略

v0.1 使用本地单用户 token：

- 服务端启动时从环境变量或配置文件读取 token。
- CLI 在 `.fluxcore/` 本地配置中保存 token。
- API 请求通过 `Authorization: Bearer <token>` 认证。

暂不实现：
- 多用户注册登录
- OAuth
- RBAC 权限模型
- 团队空间

## 模块职责

### CLI

负责：
- `fluxcore init`
- `fluxcore link`
- `fluxcore status`
- Git Hook 注入
- 本地 `.fluxcore/` 配置读写
- 事件上报

不负责：
- 任务状态计算
- Workstream 状态判断
- 数据库存储

### Backend

负责：
- REST API
- token 认证
- 领域模型持久化
- commit message 解析
- 事件写入与查询
- 后续 WebSocket 广播

不负责：
- 直接操作用户本地 Git 仓库
- 启动或控制用户编辑器、终端会话

### Web

负责：
- 项目列表
- 项目详情
- 任务与日志展示
- 后续实时事件流与 Workstream Radar

不负责：
- 直接读取本地文件系统
- 直接执行 Git 命令

## v0.1 暂不实现

- 远程 Git 平台 Webhook
- 完整多用户系统
- 完整 Workstream Radar 算法
- Intent Snapshot 命令与 UI
- AI 总结
- 插件系统
- Docker Compose 生产部署
- Slack / Discord / 邮件通知

这些能力保留在规划中，但不进入第一轮基础设施开发范围。
