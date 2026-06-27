#!/bin/bash
# CloudProbe 宝塔面板一键部署脚本
# 适用环境: Linux + 已安装宝塔面板
# 配置: 8C8G 云服务器

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_NAME="cloudprobe"
INSTALL_DIR="/www/wwwroot/${PROJECT_NAME}"
GITHUB_REPO="https://github.com/dennis926/cloudprobe.git"
GITEE_REPO="https://gitee.com/den7hon/cloudprobe.git"

echo -e "${GREEN}=== CloudProbe 宝塔面板部署脚本 ===${NC}"
echo ""

# 检测网络环境选择下载源
echo -e "${YELLOW}[1/7] 检测网络环境...${NC}"
if curl -s --max-time 3 https://github.com > /dev/null 2>&1; then
    REPO_URL="${GITHUB_REPO}"
    echo -e "${GREEN}   海外网络，使用 GitHub 源${NC}"
else
    REPO_URL="${GITEE_REPO}"
    echo -e "${GREEN}   国内网络，使用 Gitee 源${NC}"
fi

# 检查 Docker
echo -e "${YELLOW}[2/7] 检查 Docker 环境...${NC}"
if ! command -v docker &> /dev/null; then
    echo -e "${RED}   Docker 未安装，请先安装 Docker${NC}"
    echo "   宝塔面板 -> 软件商店 -> 搜索 Docker -> 安装"
    exit 1
fi
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}   Docker Compose 未安装${NC}"
    echo "   安装命令: pip install docker-compose"
    exit 1
fi
echo -e "${GREEN}   Docker 环境正常${NC}"

# 克隆代码
echo -e "${YELLOW}[3/7] 克隆代码...${NC}"
if [ -d "${INSTALL_DIR}" ]; then
    echo -e "${YELLOW}   目录已存在，执行 git pull 更新...${NC}"
    cd "${INSTALL_DIR}"
    git pull origin main
else
    git clone "${REPO_URL}" "${INSTALL_DIR}"
    cd "${INSTALL_DIR}"
fi
echo -e "${GREEN}   代码准备完成${NC}"

# 创建数据目录
echo -e "${YELLOW}[4/7] 创建数据目录...${NC}"
mkdir -p "${INSTALL_DIR}/data"
mkdir -p "${INSTALL_DIR}/config"

# 检查端口占用
echo -e "${YELLOW}[5/7] 检查端口占用...${NC}"
if ss -tlnp | grep -q ':8080 '; then
    echo -e "${RED}   警告: 8080 端口已被占用${NC}"
    echo "   请修改 docker-compose.bt.yml 中的端口映射"
    echo "   例如改为: 127.0.0.1:8081:8000"
    exit 1
fi
echo -e "${GREEN}   端口 8080 可用${NC}"

# 启动服务
echo -e "${YELLOW}[6/7] 构建并启动 CloudProbe...${NC}"
docker-compose -f docker-compose.bt.yml up -d --build

echo ""
echo -e "${GREEN}   CloudProbe 服务已启动${NC}"
echo ""

# 等待数据库初始化
echo -e "${YELLOW}[7/7] 等待数据库初始化...${NC}"
sleep 5
for i in {1..30}; do
    if docker-compose -f docker-compose.bt.yml logs dashboard 2>/dev/null | grep -q "Server starting"; then
        echo -e "${GREEN}   Dashboard 启动成功${NC}"
        break
    fi
    if [ $i -eq 30 ]; then
        echo -e "${YELLOW}   启动可能还在进行中，请稍后查看日志${NC}"
    fi
    sleep 2
done

echo ""
echo -e "${GREEN}=== 部署完成 ===${NC}"
echo ""
echo "访问方式:"
echo "  1. 直接访问: http://你的服务器IP:8080"
echo "  2. 宝塔反代: 在宝塔面板中设置域名反代到 127.0.0.1:8080"
echo ""
echo "默认账号:"
echo "  用户名: admin"
echo "  密码:   admin"
echo ""
echo "常用命令:"
echo "  查看日志:   docker-compose -f ${INSTALL_DIR}/docker-compose.bt.yml logs -f dashboard"
echo "  停止服务:   docker-compose -f ${INSTALL_DIR}/docker-compose.bt.yml down"
echo "  重启服务:   docker-compose -f ${INSTALL_DIR}/docker-compose.bt.yml restart"
echo "  查看状态:   docker-compose -f ${INSTALL_DIR}/docker-compose.bt.yml ps"
echo ""
echo -e "${YELLOW}提示: 建议通过宝塔面板设置域名 + SSL 反代访问，更安全${NC}"
