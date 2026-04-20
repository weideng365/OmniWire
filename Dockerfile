# 构建阶段 - 前端
FROM node:20-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# 构建阶段 - 后端
FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=frontend-builder /app/web/dist ./internal/packed/public
RUN CGO_ENABLED=0 go build -tags embed -ldflags="-s -w" -o omniwire main.go

# 运行阶段
FROM alpine:3.19
WORKDIR /app

# 安装 WireGuard 工具
RUN apk add --no-cache \
    wireguard-tools \
    iptables \
    ip6tables \
    iproute2 \
    curl \
    ca-certificates \
    tzdata

COPY --from=backend-builder /app/omniwire .
COPY server/manifest/config/config.example.yaml ./manifest/config/config.yaml

# 创建数据目录
RUN mkdir -p /app/data /app/logs /etc/wireguard

EXPOSE 8080 51820/udp

ENTRYPOINT ["/app/omniwire"]
