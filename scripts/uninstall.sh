#!/bin/bash
# CloudProbe 完整卸载脚本
# 清除 Docker 容器、数据卷、项目目录

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

INSTALL_DIR="/www/wwwroot/cloudprobe"

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  CloudProbe 卸载工具${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "${RED}警告: 此操作将删除以下内容：${NC}"
echo "  1. Docker 容器: cloudprobe-dashboard, cloudprobe-postgres, cloudprobe-redis"
echo "  2. Docker 数据卷: postgres_data, redis_data"
echo "  3. 项目目录: ${INSTALL_DIR}"
echo "  4. 所有监控数据、配置、用户数据"
echo ""
echo -e "${YELLOW}数据删除后不可恢复！${NC}"
echo ""

read -rp "确认卸载 CloudProbe？[yes/N] " CONFIRM
if [ "$CONFIRM" != "yes" ] && [ "$CONFIRM" != "YES" ]; then
    echo -e "${GREEN}已取消卸载${NC}"
    exit 0
fi

echo ""
echo -e "${YELLOW}[1/4] 停止并删除 Docker 容器...${NC}"
if [ -d "${INSTALL_DIR}" ]; then
    cd "${INSTALL_DIR}"
    docker-compose -f docker-compose.bt.yml down --volumes --remove-orphans 2>/dev/null || true
    docker rm -f cloudprobe-dashboard cloudprobe-postgres cloudprobe-redis 2>/dev/null || true
    echo -e "${GREEN}   Docker 容器已删除${NC}"
else
    docker rm -f cloudprobe-dashboard cloudprobe-postgres cloudprobe-redis 2>/dev/null || true
    echo -e "${GREEN}   Docker 容器已删除${NC}"
fi

echo -e "${YELLOW}[2/4] 删除 Docker 数据卷...${NC}"
docker volume rm -f cloudprobe_postgres_data cloudprobe_redis_data 2>/dev/null || true
docker volume prune -f 2>/dev/null || true
echo -e "${GREEN}   数据卷已清理${NC}"

echo -e "${YELLOW}[3/4] 删除项目目录...${NC}"
if [ -d "${INSTALL_DIR}" ]; then
    rm -rf "${INSTALL_DIR}"
    echo -e "${GREEN}   目录已删除: ${INSTALL_DIR}${NC}"
else
    echo -e "${YELLOW}   目录不存在，跳过${NC}"
fi

echo -e "${YELLOW}[4/4] 清理残留...${NC}"
# 删除可能残留的 Go 安装（可选）
rm -f /tmp/go.tar.gz /tmp/gen_hash.go
echo -e "${GREEN}   临时文件已清理${NC}"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  CloudProbe 已完全卸载${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo "端口状态检查:"
ss -tlnp | grep -E ':809[0-9]' || echo -e "${GREEN}   809x 端口已释放${NC}"
echo ""
echo -e "${CYAN}如需重新安装，执行:${NC}"
echo "  bash <(curl -fsSL https://raw.githubusercontent.com/dennis926/cloudprobe/main/scripts/deploy-bt.sh)"
