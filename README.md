# CloudProbe - 云服务器探针系统

CloudProbe 是一款开源、自托管的一站式运维面板，集服务器监控探针、多渠道告警通知、远程SSH运维、代理协议管理于一体。

## 功能特性

- **服务器监控**：实时采集CPU、内存、磁盘、网络指标，支持90天历史数据趋势图
- **多渠道告警**：微信、QQ、飞书、邮件、Telegram，支持分级告警升级（1分钟/5分钟）
- **WebSSH终端**：浏览器内直接SSH连接目标服务器，支持命令历史和多标签页
- **代理管理（3x-ui集成）**：通过API对接3x-ui面板，统一管理入站协议、客户端、节点
- **服务检测**：端口/HTTP/SSL证书/Docker容器状态监控
- **双部署方案**：国内（WSS通信）/ 国外（gRPC通信）
- **API接口**：RESTful API，支持第三方系统对接
- **数据备份**：90天数据保留，自动备份发送到邮箱

## 技术栈

| 层面 | 技术 |
|------|------|
| 后端 | Go 1.22 + Gin + gRPC + WebSocket |
| 前端 | Vue3 + TypeScript + Vite + Element Plus + Tailwind CSS |
| 数据库 | PostgreSQL 15 + TimescaleDB |
| 缓存 | Redis 7 |
| 终端 | xterm.js |
| 图表 | ECharts |

## 快速开始

### Docker Compose 部署

```bash
# 克隆仓库
git clone https://github.com/dennis926/cloudprobe.git
cd cloudprobe

# 启动服务
docker compose up -d

# 查看日志
docker logs -f cloudprobe-dashboard
```

访问 http://localhost:8000，默认账户密码 admin/admin（请立即修改）。

### Agent安装

在目标服务器上执行：

```bash
# 国内服务器
curl -fsSL https://your-domain.com/api/v1/agents/install.sh | bash -s -- https://your-domain.com YOUR_AGENT_TOKEN domestic

# 国外服务器
curl -fsSL https://your-domain.com/api/v1/agents/install.sh | bash -s -- https://your-domain.com YOUR_AGENT_TOKEN foreign
```

## 项目结构

```
cloudprobe/
├── cmd/
│   ├── dashboard/    # Dashboard主程序
│   └── agent/        # Agent主程序
├── internal/         # 内部模块
│   ├── api/          # HTTP API处理器
│   ├── service/      # 业务逻辑层
│   ├── repository/   # 数据访问层
│   ├── model/        # 数据模型
│   ├── notify/       # 通知渠道适配器
│   ├── config/       # 配置管理
│   ├── auth/         # 认证授权
│   └── task/         # 定时任务
├── web/              # Vue3前端
├── scripts/          # 安装脚本
├── deploy/           # 部署配置
├── docs/             # 文档
└── proto/            # Protocol Buffers
```

## 开发

```bash
# 后端
go mod download
go run ./cmd/dashboard

# 前端
cd web
npm install
npm run dev
```

## 文档

- [产品需求文档 (PRD)](docs/PRD.html)
- [技术方案文档](docs/TECH.html)

## 开源协议

MIT License
