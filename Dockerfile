# ===== 构建阶段 =====
# 始终在本机架构上运行编译器（BUILDPLATFORM），配合 TARGETOS/TARGETARCH
# 支持 docker buildx --platform linux/amd64 跨平台编译
# 注意：不要加 `# syntax=docker/dockerfile:1`，部分国内镜像加速器拉不到该前端镜像，
# 而本文件用到的 cache mount / TARGETARCH 内置 BuildKit 前端已原生支持
FROM --platform=$BUILDPLATFORM golang:1.23 AS builder

ARG TARGETOS
ARG TARGETARCH
# 构建时注入的版本号：CI 传 --build-arg VERSION=<commit>，本地构建默认 docker
ARG VERSION=docker

# 设置国内 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 先复制依赖文件，利用 layer cache（改业务代码不会触发重新下载依赖）
COPY go.mod go.sum ./
# cache mount 让模块缓存跨构建持久化，go.mod 变动时也无需全量重新下载
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制源码并构建（关闭 CGO 产出静态二进制）
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w -X main.Version=${VERSION}" -trimpath -o web-app .

# ===== 运行阶段 =====
# 仍用 alpine:3.19：本机镜像加速器拉不到新 tag（403）、直连 Docker Hub 超时，
# 3.19 为本地已有缓存；网络恢复后建议升级到更新的 alpine 小版本
FROM alpine:3.19 AS runtime

# ca-certificates: HTTPS 调用需要（如 Sentry）
# tzdata: 容器内时区统一为 Asia/Shanghai
# 创建非 root 用户，并预创建日志目录（应用按配置写 ./logs/*.log）
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && addgroup -S app && adduser -S -G app app \
    && mkdir -p /app/logs && chown -R app:app /app

WORKDIR /app

COPY --from=builder /app/web-app ./web-app

# 以非 root 身份运行，减小攻击面
USER app

EXPOSE 8080

ENTRYPOINT ["./web-app"]
