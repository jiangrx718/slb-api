#!/bin/bash

IMAGE_PREFIX="slb-api"
ARCH_SUFFIX="amd64"

# 如果传了参数，就用参数作为版本号，否则获取本地最大版本号
if [[ -n "$1" ]]; then
  VERSION="$1"
else
  VERSION=$(docker images --format "{{.Repository}}:{{.Tag}}" \
    | grep "^${IMAGE_PREFIX}:" \
    | grep "${ARCH_SUFFIX}" \
    | sed -E "s/^${IMAGE_PREFIX}:([0-9]+\.[0-9]+\.[0-9]+)-${ARCH_SUFFIX}$/\1/" \
    | sort -Vr \
    | head -n 1)
fi

if [[ -z "$VERSION" ]]; then
  echo "❌ 未找到符合格式的镜像版本（slb-api:x.y.z-amd64）"
  exit 1
fi

FULL_IMAGE="${IMAGE_PREFIX}:${VERSION}-${ARCH_SUFFIX}"

echo "✅ 即将运行镜像: ${FULL_IMAGE}"

# 停止并删除旧容器（如存在）
docker rm -f slb-api 2>/dev/null

# 启动容器
docker run -d \
  --name slb-api \
  --restart=always \
  -p 8083:8083 \
  -v /etc/localtime:/etc/localtime:ro \
  -v /etc/timezone:/etc/timezone:ro \
  -v /Users/jiang/go/src/private/slb/slb-api/config/app.yml:/app/config/app.yml \
  --network common-app-net \
  "$FULL_IMAGE"
