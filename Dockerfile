# 第一阶段：构建 Go 应用
FROM golang:1.23 AS builder

# 设置国内 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 先复制依赖文件，利用 layer cache（改业务代码不会触发重新下载依赖）
COPY go.mod go.sum ./
RUN go mod download

# 复制源码并构建（关闭 CGO）
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
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

COPY --from=builder /app/tool-agent ./tool-agent

EXPOSE 8080

ENTRYPOINT ["./tool-agent"]
