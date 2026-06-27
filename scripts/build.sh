#!/bin/bash
# CloudProbe 服务器端快速编译检查脚本
# 用途: 在 Docker 构建前先快速发现编译错误，避免反复 Docker build
# 用法: bash scripts/build.sh

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_DIR"

echo -e "${YELLOW}=== CloudProbe 快速编译检查 ===${NC}"

# 检查 Go 环境
echo -e "${YELLOW}[1/4] 检查 Go 环境...${NC}"
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}   Go 未安装，正在安装 Go 1.22...${NC}"
    GO_VERSION="1.22.0"
    GO_ARCH=$(uname -m)
    case "$GO_ARCH" in
        x86_64)  GO_SUFFIX="amd64" ;;
        aarch64) GO_SUFFIX="arm64" ;;
        *) echo -e "${RED}   不支持的架构: $GO_ARCH${NC}"; exit 1 ;;
    esac
    GO_URL="https://go.dev/dl/go${GO_VERSION}.linux-${GO_SUFFIX}.tar.gz"
    wget -q -O /tmp/go.tar.gz "$GO_URL"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo -e "${GREEN}   Go $(go version | awk '{print $3}') 安装完成${NC}"
else
    echo -e "${GREEN}   $(go version)${NC}"
fi

# 下载依赖 + 生成 go.sum
echo -e "${YELLOW}[2/4] 整理依赖...${NC}"
export GOPROXY=https://goproxy.cn,direct
go mod tidy
echo -e "${GREEN}   go.sum 已生成${NC}"

# 快速编译 dashboard
echo -e "${YELLOW}[3/5] 编译 dashboard...${NC}"
CGO_ENABLED=0 go build -o /dev/null ./cmd/dashboard
echo -e "${GREEN}   dashboard 编译通过${NC}"

# 快速编译 agent
echo -e "${YELLOW}[4/5] 编译 agent...${NC}"
CGO_ENABLED=0 go build -o /dev/null ./cmd/agent
echo -e "${GREEN}   agent 编译通过${NC}"

echo ""
echo -e "${GREEN}=== 编译检查全部通过 ===${NC}"
echo ""
echo "接下来执行 Docker 构建:"
echo "  docker-compose -f docker-compose.bt.yml build --no-cache"
echo "  docker-compose -f docker-compose.bt.yml up -d"
