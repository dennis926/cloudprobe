#!/bin/bash
# CloudProbe 宝塔面板一键部署脚本
# 适用环境: Linux + 已安装宝塔面板 (8C8G+)
# 功能: 端口交互检测 -> 自动安装Go -> 编译检查 -> Docker构建 -> 生成部署文档

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

PROJECT_NAME="cloudprobe"
INSTALL_DIR="/www/wwwroot/${PROJECT_NAME}"
GITHUB_REPO="https://github.com/dennis926/cloudprobe.git"
GITEE_REPO="https://gitee.com/den7hon/cloudprobe.git"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  CloudProbe 宝塔面板一键部署脚本${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 检测网络环境选择下载源
echo -e "${YELLOW}[1/9] 检测网络环境...${NC}"
if curl -s --max-time 3 https://github.com > /dev/null 2>&1; then
    REPO_URL="${GITHUB_REPO}"
    GOPROXY_URL="https://proxy.golang.org,direct"
    echo -e "${GREEN}   海外网络 -> GitHub + Go Proxy${NC}"
else
    REPO_URL="${GITEE_REPO}"
    GOPROXY_URL="https://goproxy.cn,direct"
    echo -e "${GREEN}   国内网络 -> Gitee + goproxy.cn${NC}"
fi

# 检查 Docker
echo -e "${YELLOW}[2/9] 检查 Docker 环境...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}   Docker 未安装${NC}"
    echo "   宝塔面板 -> 软件商店 -> Docker -> 安装"
    exit 1
fi
if ! docker compose version &> /dev/null 2>&1 && ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}   Docker Compose 未安装${NC}"
    exit 1
fi
echo -e "${GREEN}   Docker $(docker --version | awk '{print $3}') OK${NC}"

# 克隆代码
echo -e "${YELLOW}[3/9] 获取代码...${NC}"
if [ -d "${INSTALL_DIR}" ]; then
    cd "${INSTALL_DIR}"
    git config pull.rebase false 2>/dev/null
    git pull origin main
else
    git clone "${REPO_URL}" "${INSTALL_DIR}"
    cd "${INSTALL_DIR}"
fi
echo -e "${GREEN}   代码就绪${NC}"

# 创建数据目录
echo -e "${YELLOW}[4/9] 创建数据目录...${NC}"
mkdir -p "${INSTALL_DIR}/data" "${INSTALL_DIR}/config"

# ==================== 交互式端口检测 ====================
echo -e "${YELLOW}[5/9] 检查端口占用...${NC}"

# 检测端口是否被占用的函数
check_port() {
    local port=$1
    ss -tlnp 2>/dev/null | grep -q ":${port} "
}

# 获取占用端口的进程信息
get_port_info() {
    local port=$1
    ss -tlnp 2>/dev/null | grep ":${port} " | head -1
}

# 查找下一个可用端口
find_next_port() {
    local port=$1
    while check_port "$port"; do
        port=$((port + 1))
    done
    echo "$port"
}

CP_PORT=8090

if check_port "$CP_PORT"; then
    PORT_INFO=$(get_port_info "$CP_PORT")
    echo -e "${RED}   警告: 端口 ${CP_PORT} 已被占用${NC}"
    echo -e "${CYAN}   占用信息: ${PORT_INFO}${NC}"
    echo ""
    echo -e "${YELLOW}   请选择处理方式：${NC}"
    echo -e "   1) 输入自定义端口（推荐）"
    echo -e "   2) 使用自动检测的可用端口"
    echo ""
    read -rp "   请输入选项 [1/2，默认2]: " PORT_CHOICE
    PORT_CHOICE=${PORT_CHOICE:-2}

    if [ "$PORT_CHOICE" = "1" ]; then
        while true; do
            read -rp "   请输入新端口(1024-65535): " USER_PORT
            if ! [[ "$USER_PORT" =~ ^[0-9]+$ ]] || [ "$USER_PORT" -lt 1024 ] || [ "$USER_PORT" -gt 65535 ]; then
                echo -e "${RED}   输入无效，请输入 1024-65535 之间的数字${NC}"
                continue
            fi
            if check_port "$USER_PORT"; then
                PORT_INFO=$(get_port_info "$USER_PORT")
                echo -e "${RED}   端口 ${USER_PORT} 也被占用了: ${PORT_INFO}${NC}"
                echo -e "${YELLOW}   请重新输入${NC}"
                continue
            fi
            CP_PORT=$USER_PORT
            break
        done
    else
        CP_PORT=$(find_next_port 8090)
        echo -e "${GREEN}   自动使用可用端口: ${CP_PORT}${NC}"
    fi
else
    echo -e "${GREEN}   端口 ${CP_PORT} 可用${NC}"
fi

# 替换 docker-compose 中的端口
if [ "${CP_PORT}" != "8090" ]; then
    sed -i "s/0.0.0.0:8090:8000/0.0.0.0:${CP_PORT}:8000/" docker-compose.bt.yml
fi

# 获取服务器外网IP
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || curl -s icanhazip.com 2>/dev/null || echo "你的服务器IP")

# ==================== 交互式管理员初始化 ====================
echo ""
echo -e "${YELLOW}[6/9] 管理员账号初始化${NC}"
echo -e "${CYAN}----------------------------------------${NC}"

read -rp "是否自动初始化管理员账号？[Y/n] " INIT_ADMIN
INIT_ADMIN=${INIT_ADMIN:-Y}

if [ "$INIT_ADMIN" = "Y" ] || [ "$INIT_ADMIN" = "y" ]; then
    read -rp "请输入管理员用户名 [默认: admin]: " ADMIN_USER
    ADMIN_USER=${ADMIN_USER:-admin}

    read -rsp "请输入管理员密码 [默认: admin]: " ADMIN_PASS
    echo ""
    ADMIN_PASS=${ADMIN_PASS:-admin}

    cat > "${INSTALL_DIR}/.env" << EOF
CP_ADMIN_USER=${ADMIN_USER}
CP_ADMIN_PASS=${ADMIN_PASS}
EOF

    echo -e "${GREEN}   已配置自动初始化: ${ADMIN_USER} / ${ADMIN_PASS}${NC}"
else
    # 手动模式：不写入 .env，等 Docker 启动后通过 SQL 直接创建
    echo -e "${YELLOW}   已选择手动模式，将在部署完成后创建${NC}"
fi
echo -e "${CYAN}----------------------------------------${NC}"

# 安装 Go + 快速编译检查
echo -e "${YELLOW}[7/9] Go 编译检查...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}   安装 Go 1.22...${NC}"
    GO_VERSION="1.22.0"
    GO_ARCH=$(uname -m)
    case "$GO_ARCH" in
        x86_64)  GO_SUFFIX="amd64" ;;
        aarch64) GO_SUFFIX="arm64" ;;
        *)
            echo -e "${RED}   不支持架构: $GO_ARCH${NC}"
            exit 1
            ;;
    esac
    wget -q -O /tmp/go.tar.gz "https://go.dev/dl/go${GO_VERSION}.linux-${GO_SUFFIX}.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
fi

export GOPROXY="${GOPROXY_URL}"
echo "   $(go version)"

echo -e "${YELLOW}   下载依赖...${NC}"
go mod tidy

echo -e "${YELLOW}   编译 dashboard + agent...${NC}"
set +e
BUILD_OUTPUT=$(CGO_ENABLED=0 go build -o /dev/null ./cmd/dashboard 2>&1)
BUILD_EXIT=$?
if [ $BUILD_EXIT -eq 0 ]; then
    BUILD_OUTPUT=$(CGO_ENABLED=0 go build -o /dev/null ./cmd/agent 2>&1)
    BUILD_EXIT=$?
fi
set -e
if [ $BUILD_EXIT -eq 0 ]; then
    echo -e "${GREEN}   编译通过${NC}"
else
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  编译失败！错误信息如下：${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    echo "$BUILD_OUTPUT"
    echo ""
    echo -e "${RED}========================================${NC}"
    echo -e "${RED}  请把上方错误信息复制给我，我来修复${NC}"
    echo -e "${RED}========================================${NC}"
    echo ""
    exit 1
fi

# Docker 构建
echo -e "${YELLOW}[8/9] Docker 构建...${NC}"
docker-compose -f docker-compose.bt.yml up -d --build

# 等待服务就绪
echo -e "${YELLOW}[9/10] 等待服务启动...${NC}"
sleep 5
for i in {1..30}; do
    if docker-compose -f docker-compose.bt.yml logs dashboard 2>/dev/null | grep -q "Server starting"; then
        echo -e "${GREEN}   Dashboard 启动成功${NC}"
        break
    fi
    [ $i -eq 30 ] && echo -e "${YELLOW}   启动可能仍在进行，稍后检查日志${NC}"
    sleep 2
done

# ==================== 手动模式：创建管理员账号 ====================
if [ "$INIT_ADMIN" = "n" ] || [ "$INIT_ADMIN" = "N" ]; then
    echo ""
    echo -e "${YELLOW}========================================${NC}"
    echo -e "${YELLOW}  手动创建管理员账号${NC}"
    echo -e "${YELLOW}========================================${NC}"

    while true; do
        read -rp "请输入管理员用户名: " MANUAL_USER
        if [ -z "$MANUAL_USER" ]; then
            echo -e "${RED}用户名不能为空，请重新输入${NC}"
            continue
        fi
        break
    done

    while true; do
        read -rsp "请输入管理员密码: " MANUAL_PASS
        echo ""
        if [ -z "$MANUAL_PASS" ]; then
            echo -e "${RED}密码不能为空，请重新输入${NC}"
            continue
        fi
        read -rsp "请再次确认密码: " MANUAL_PASS2
        echo ""
        if [ "$MANUAL_PASS" != "$MANUAL_PASS2" ]; then
            echo -e "${RED}两次输入的密码不一致，请重新输入${NC}"
            continue
        fi
        break
    done

    echo -e "${YELLOW}正在生成密码哈希...${NC}"
    export PATH=$PATH:/usr/local/go/bin

    cat > /tmp/gen_hash.go << 'GOEOF'
package main
import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
    "os"
)
func main() {
    pass := os.Getenv("TEMP_PASS")
    h, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
    if err != nil {
        fmt.Println("ERROR:", err)
        return
    }
    fmt.Println(string(h))
}
GOEOF

    export TEMP_PASS="$MANUAL_PASS"
    HASH=$(go run /tmp/gen_hash.go)
    unset TEMP_PASS
    rm -f /tmp/gen_hash.go

    echo -e "${YELLOW}正在写入数据库...${NC}"
    docker exec -i cloudprobe-postgres psql -U cpuser -d cloudprobe -c "
    INSERT INTO users (username, password, role, status, created_at, updated_at)
    VALUES ('${MANUAL_USER}', '${HASH}', 'admin', 'active', NOW(), NOW())
    ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password;
    "

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}管理员账号创建成功！${NC}"
        ADMIN_USER=$MANUAL_USER
        ADMIN_PASS=$MANUAL_PASS
    else
        echo -e "${RED}管理员账号创建失败，请手动执行 reset-password.sh${NC}"
    fi
fi

# 设置文档中显示的账号（如果用户跳过初始化则显示默认值）
DOC_ADMIN_USER=${ADMIN_USER:-admin}
DOC_ADMIN_PASS=${ADMIN_PASS:-admin}

# 生成部署信息文档
echo -e "${YELLOW}[10/10] 生成部署信息文档...${NC}"
DEPLOY_TIME=$(date '+%Y-%m-%d %H:%M:%S')
DEPLOY_INFO_FILE="${INSTALL_DIR}/deploy-info.txt"

cat > "${DEPLOY_INFO_FILE}" << EOF
========================================
  CloudProbe 部署信息
========================================

部署时间: ${DEPLOY_TIME}
安装目录: ${INSTALL_DIR}

----------------------------------------
  访问信息
----------------------------------------
外网访问: http://${SERVER_IP}:${CP_PORT}
内网访问: http://127.0.0.1:${CP_PORT}

管理员账号:
  用户名: ${DOC_ADMIN_USER}
  密码:   ${DOC_ADMIN_PASS}

----------------------------------------
  服务状态
----------------------------------------
Dashboard 端口: ${CP_PORT}
Dashboard 容器: cloudprobe-dashboard
PostgreSQL 容器: cloudprobe-postgres (内部端口 5432)
Redis 容器:      cloudprobe-redis (内部端口 6379)

----------------------------------------
  防火墙配置建议
----------------------------------------
请确保云服务器安全组/防火墙放行以下端口:
  TCP ${CP_PORT}  (CloudProbe Dashboard)
  TCP 80          (HTTP, 用于宝塔 Nginx 反代)
  TCP 443         (HTTPS, 用于 SSL 访问)

----------------------------------------
  宝塔面板反代配置（推荐）
----------------------------------------
1. 宝塔面板 -> 网站 -> 添加站点
2. 输入你的域名
3. 设置 -> 反向代理 -> 添加目标 URL: http://127.0.0.1:${CP_PORT}
4. 申请 SSL 证书，开启强制 HTTPS

----------------------------------------
  常用命令
----------------------------------------
查看日志:
  cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml logs -f dashboard

查看状态:
  cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml ps

停止服务:
  cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml down

重启服务:
  cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml restart

更新到最新版:
  cd ${INSTALL_DIR} && git pull origin main && bash scripts/deploy-bt.sh

----------------------------------------
  全新重新安装（如已卸载）
----------------------------------------
海外网络:
  bash <(curl -fsSL https://raw.githubusercontent.com/dennis926/cloudprobe/main/scripts/deploy-bt.sh)

国内网络:
  bash <(curl -fsSL https://gitee.com/den7hon/cloudprobe/raw/main/scripts/deploy-bt.sh)

----------------------------------------
  Agent 安装命令
----------------------------------------
在被监控服务器上执行:

国内服务器 (WebSocket):
  curl -fsSL http://${SERVER_IP}:${CP_PORT}/api/v1/agent/install | bash

海外服务器 (gRPC):
  curl -fsSL http://${SERVER_IP}:${CP_PORT}/api/v1/agent/install | bash -s -- -m grpc

----------------------------------------
  项目信息
----------------------------------------
GitHub: https://github.com/dennis926/cloudprobe
Gitee:  https://gitee.com/den7hon/cloudprobe

========================================
EOF

echo -e "${GREEN}   部署信息已保存到: ${DEPLOY_INFO_FILE}${NC}"

# ==================== 部署完成 ====================
echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${CYAN}访问地址: http://${SERVER_IP}:${CP_PORT}${NC}"
echo -e "${CYAN}管理员账号: ${DOC_ADMIN_USER} / ${DOC_ADMIN_PASS}${NC}"
echo ""
echo -e "${YELLOW}重要提示:${NC}"
echo -e "  1. 请在云服务器安全组中放行 TCP ${CP_PORT} 端口"
echo -e "  2. 部署信息文档: ${DEPLOY_INFO_FILE}"
echo -e "  3. 建议配置域名 + SSL 通过宝塔 Nginx 反代访问"
echo ""
