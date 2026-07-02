# syntax=docker/dockerfile:1

# ===== Build stage =====
FROM golang:1.23-alpine AS builder

WORKDIR /build

# 先复制依赖文件，利用 layer cache
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并静态编译（CGO 关闭以便在 alpine 上运行）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=docker" \
    -trimpath \
    -o tool-agent .

# ===== Runtime stage =====
FROM alpine:3.20

# ca-certificates: HTTPS 调用需要
# tzdata: 容器内时间 zone
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone

WORKDIR /app

COPY --from=builder /build/tool-agent ./tool-agent
COPY config/app.yml /app/config/app.yml
EXPOSE 8080

ENTRYPOINT ["./tool-agent"]
