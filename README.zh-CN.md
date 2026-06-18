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
API_TOKEN=local-dev-token go run .

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
