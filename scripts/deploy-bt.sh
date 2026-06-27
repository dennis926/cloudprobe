#!/bin/bash
# CloudProbe 宝塔面板一键部署脚本
# 适用环境: Linux + 已安装宝塔面板 (8C8G+)
# 功能: 自动安装Go -> 编译检查 -> Docker构建部署

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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
echo -e "${YELLOW}[1/8] 检测网络环境...${NC}"
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
echo -e "${YELLOW}[2/8] 检查 Docker 环境...${NC}"
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
echo -e "${YELLOW}[3/8] 获取代码...${NC}"
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
echo -e "${YELLOW}[4/8] 创建数据目录...${NC}"
mkdir -p "${INSTALL_DIR}/data" "${INSTALL_DIR}/config"

# 检查端口占用
echo -e "${YELLOW}[5/8] 检查端口占用...${NC}"
CP_PORT=8090
while ss -tlnp | grep -q ":${CP_PORT} "; do
    echo -e "${YELLOW}   端口 ${CP_PORT} 被占用，尝试 ${CP_PORT}+1${NC}"
    CP_PORT=$((CP_PORT + 1))
done
echo -e "${GREEN}   使用端口 ${CP_PORT}${NC}"
if [ "${CP_PORT}" != "8090" ]; then
    sed -i "s/127.0.0.1:8090:8000/127.0.0.1:${CP_PORT}:8000/" docker-compose.bt.yml
fi

# 安装 Go + 快速编译检查（秒级反馈，发现错误立即中止）
echo -e "${YELLOW}[6/8] Go 编译检查...${NC}"
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
echo -e "${YELLOW}[7/8] Docker 构建...${NC}"
docker-compose -f docker-compose.bt.yml up -d --build

# 等待服务就绪
echo -e "${YELLOW}[8/8] 等待服务启动...${NC}"
sleep 5
for i in {1..30}; do
    if docker-compose -f docker-compose.bt.yml logs dashboard 2>/dev/null | grep -q "Server starting"; then
        echo -e "${GREEN}   Dashboard 启动成功${NC}"
        break
    fi
    [ $i -eq 30 ] && echo -e "${YELLOW}   启动可能仍在进行，稍后检查日志${NC}"
    sleep 2
done

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  部署完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "访问: http://$(curl -s ifconfig.me):${CP_PORT}"
echo "账号: admin / admin"
echo ""
echo "常用命令:"
echo "  查看日志:  cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml logs -f dashboard"
echo "  停止:      cd ${INSTALL_DIR} && docker-compose -f docker-compose.bt.yml down"
echo "  更新:      cd ${INSTALL_DIR} && git pull origin main && bash scripts/deploy-bt.sh"
