# 第一阶段：构建 Go 应用
FROM golang:1.23 AS builder

# 设置国内 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app
COPY . .

# 下载依赖并构建（关闭 CGO）
RUN go mod download && \
    CGO_ENABLED=0 GOOS=linux go build \
        -ldflags="-s -w -X main.Version=docker" \
        -trimpath \
        -o tool-agent .

# ===== Runtime stage =====
FROM alpine:3.19 as prod

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
