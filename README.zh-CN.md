<div align="center">

# FluxCore

**Code is Log. Push is Update.**

专为个人开发者和多项目并发场景设计的**研发效能控制台**——深度集成 Git 工作流，让每次 `git push` 自动驱动项目仪表盘更新，零手动填报、零上下文切换。

[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react&logoColor=black)](https://react.dev)
[![Vite](https://img.shields.io/badge/Vite-6-646CFF?style=flat-square&logo=vite&logoColor=white)](https://vitejs.dev)
[![SQLite](https://img.shields.io/badge/SQLite-003B57?style=flat-square&logo=sqlite&logoColor=white)](https://sqlite.org)
[![Redis](https://img.shields.io/badge/Redis-DC382D?style=flat-square&logo=redis&logoColor=white)](https://redis.io)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)

[English](README.md) · [简体中文](README.zh-CN.md)

</div>

---

## 为什么选择 FluxCore？

大多数项目管理工具迫使开发者陷入**双循环困境**——写完代码后，还要手动记录做了什么。FluxCore 彻底消灭第二个循环。

| 痛点 | 传统工具 | FluxCore |
| :--- | :--- | :--- |
| 进度追踪 | 手动填工时 / 更新工单 | 从 Commit 自动生成 |
| 项目状态 | 过时的看板和文档 | WebSocket 实时同步 |
| 多项目切换 | 终端散落、上下文丢失 | `fluxcore switch` 一键恢复 |
| README 作为项目元数据 | 与开发流程脱节 | 每次推送自动解析并同步 |

## 核心理念

- **无感记录** — 拒绝手动填报工时。通过 Git Hooks + Commit Message 解析，自动生成结构化功能日志。
- **实时反馈** — 类似 Jenkins 的即时体验。代码推送后，Web 仪表盘通过 WebSocket 毫秒级刷新。
- **上下文感知** — 无论是 CLI 还是 Web UI，都能智能识别当前项目状态与 README 元数据。

## 系统架构

```
┌─────────────────────────────────────────────────────────────────┐
│                         开发者工作流                             │
│   git commit → post-commit hook → fluxcore CLI → Backend API    │
└────────────────────────────┬────────────────────────────────────┘
                             │
              ┌──────────────▼──────────────┐
              │      Go 后端服务 (Gin)       │
              │  ┌────────┐  ┌───────────┐  │
              │  │Webhook │  │  Commit    │  │
              │  │接收器   │  │  解析器    │  │
              │  └───┬────┘  └─────┬─────┘  │
              │      │             │         │
              │  ┌───▼─────────────▼─────┐  │
              │  │    GORM (SQLite/PG)   │  │
              │  └───────────┬───────────┘  │
              │              │              │
              │  ┌───────────▼───────────┐  │
              │  │  Redis Pub/Sub + WS   │  │
              │  └───────────┬───────────┘  │
              └──────────────┼──────────────┘
                             │
              ┌──────────────▼──────────────┐
              │   React + Vite 实时仪表盘    │
              │  项目卡片、活动时间轴、CI 状态 │
              └─────────────────────────────┘
```

## 技术栈

| 模块 | 技术选型 | 说明 |
| :--- | :--- | :--- |
| **CLI** | Go (Cobra) | Git Hook 注入、本地上下文管理、项目切换 |
| **后端** | Go (Gin) | REST API、Webhook 处理、WebSocket 广播 |
| **ORM** | GORM | 数据持久化，支持 SQLite/PostgreSQL 动态切换 |
| **数据库** | SQLite / PostgreSQL | 开发环境 SQLite（零配置），生产环境可迁移至 PostgreSQL |
| **消息** | Redis | WebSocket Pub/Sub 实时事件分发 |
| **前端** | React 19 + Vite 6 | 现代化 SPA 仪表盘 |
| **样式** | Tailwind CSS | 原子化 CSS，快速构建仪表盘界面 |

## 工作流预览

```bash
# 1. 绑定项目
fluxcore link --project "my-awesome-app"

# 2. 照常开发与提交
git commit -m "feat: #101 完成支付接口"
git push

# 3. 一切自动发生
#    → CLI：终端显示推送确认
#    → Web：项目卡片版本号即时跳动，
#           时间轴出现新日志，
#           任务 #101 自动流转至"测试中"
```

## 开发路线图

### 阶段一 — 基础设施与无感绑定

> 搭建基础环境，打通 CLI ↔ 后端连接，实现项目本地绑定。

- **后端**
  - [ ] 初始化 Go 模块与 Gin 框架
  - [ ] 配置 GORM，实现基于 `DB_TYPE` 环境变量的 SQLite/PostgreSQL 动态切换
  - [ ] 定义 `Project`、`User`、`Config` 数据模型
  - [ ] 实现 `POST /api/projects` 与 `GET /api/projects` 接口
  - [ ] 添加健康检查与基础认证中间件
- **CLI**
  - [ ] 实现 `fluxcore init` — 在项目中初始化 `.fluxcore/` 配置目录
  - [ ] 实现 `fluxcore link` — 读取 `.git/config`，生成本地项目映射
  - [ ] 添加 `fluxcore status` — 显示当前项目绑定状态
- **前端**
  - [ ] 搭建 React + Vite + Tailwind 基础框架
  - [ ] 构建项目列表视图与空状态页
  - [ ] 实现项目创建弹窗

### 阶段二 — Git 事件驱动与自动化日志

> 实现"推送即记录"——FluxCore 的灵魂。

- **后端**
  - [ ] 开发 Webhook 接收器，监听 Git Push 事件
  - [ ] 开发 Commit Message 解析器（正则匹配 `#TaskID`、Conventional Commit 类型）
  - [ ] 从解析结果自动创建结构化日志条目
  - [ ] 实现 `Task` 模型与状态机（Open → In Progress → Testing → Done）
- **CLI**
  - [ ] 通过 `fluxcore init` 自动注入 `post-commit` Hook
  - [ ] 添加 `fluxcore log` — 终端查看近期活动
- **前端**
  - [ ] 构建项目详情页，展示任务列表与日志时间轴
  - [ ] 添加提交类型标签（feat / fix / refactor / chore）

### 阶段三 — README 驱动同步与实时推送

> 引入 Redis，实现文档变更即时反映。

- **后端**
  - [ ] 集成 Redis，建立 WebSocket 广播管线
  - [ ] 监听推送中的 `README.md` 变更，解析 Front Matter（`status`、`version`、`description`）
  - [ ] 自动从 README 元数据更新项目卡片
- **前端**
  - [ ] 接入 WebSocket，实现项目卡片与日志流的无刷新实时更新
  - [ ] 添加实时状态指示器（在线 / 同步中 / 离线）
  - [ ] 实现事件到达 Toast 通知

### 阶段四 — 多项目并发与全局视角

> 解决多项目切换痛点，提供全局监控能力。

- **CLI**
  - [ ] 实现 `fluxcore switch <project>` — 自动切换工作目录并恢复上下文
  - [ ] 添加 `fluxcore dashboard` — 从终端打开 Web UI
- **前端**
  - [ ] 构建全局仪表盘，聚合所有项目实时动态流
  - [ ] 集成 CI 状态展示（构建成功/失败徽标）
  - [ ] 添加项目级筛选、排序与搜索

### 阶段五 — 智能化与生态完善

> 引入 AI 辅助，完善部署与通知生态。

- [ ] **AI 集成** — 分析 `git diff` 自动生成工作日报摘要
- [ ] **插件系统** — 支持自定义 Webhook 处理器与事件处理器
- [ ] **一键部署** — 提供 Docker Compose 一键部署方案
- [ ] **通知推送** — 可选的 Slack / Discord / 邮件关键事件推送

## 项目结构

```
FluxCore/
├── cli/                  # CLI 工具 (Go + Cobra)
│   ├── cmd/              #   命令定义
│   ├── internal/         #   Hook 注入、配置管理
│   └── main.go
├── server/               # 后端服务 (Go + Gin)
│   ├── api/              #   HTTP 路由处理
│   ├── model/            #   GORM 数据模型
│   ├── service/          #   业务逻辑层
│   ├── ws/               #   WebSocket Hub
│   ├── db/               #   数据库连接 (SQLite/PG)
│   └── main.go
├── web/                  # 前端 (React + Vite)
│   ├── src/
│   │   ├── components/   #   可复用 UI 组件
│   │   ├── pages/        #   路由页面
│   │   ├── hooks/        #   自定义 React Hooks
│   │   └── lib/          #   API 客户端、WebSocket 客户端
│   └── index.html
├── migrations/           # 数据库迁移文件
├── docker-compose.yml    # 生产部署配置
└── .fluxcore/            # 本地配置 (已 gitignore)
```

## 快速开始

> **提示：** FluxCore 正在积极开发中，以下为目标安装流程。

### 前置条件

- Go 1.23+
- Node.js 20+ & npm
- Redis（阶段三起需要）

### 本地开发

```bash
# 克隆仓库
git clone https://github.com/your-username/FluxCore.git
cd FluxCore

# 启动后端
cd server
cp .env.example .env
go run main.go

# 启动前端（新终端窗口）
cd web
npm install && npm run dev

# 安装 CLI
cd cli
go install .

# 在任意项目中初始化 FluxCore
cd /path/to/your/project
fluxcore init
fluxcore link --project "my-project"
```

## 参与贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建特性分支（`git checkout -b feat/amazing-feature`）
3. 提交变更，遵循 [Conventional Commits](https://www.conventionalcommits.org/zh-Hans/) 规范
4. 推送分支（`git push origin feat/amazing-feature`）
5. 发起 Pull Request

## 开源协议

本项目基于 [MIT 协议](LICENSE) 开源。

---

<div align="center">

**FluxCore** — 让代码自己说话。

</div>
