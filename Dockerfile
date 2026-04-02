# 1. 编译阶段
FROM golang:1.24-alpine AS builder

# [关键] 设置国内代理加速下载
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

# 只有 go.mod 或 go.sum 变化时才会重新下载依赖
COPY go.mod go.sum ./

# [优化] 增加 -x 可以看到下载进度，如果卡住能知道死在哪个包
RUN go mod download -x

COPY . .

# 编译静态二进制文件
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gvb-server .

# 2. 运行阶段
FROM alpine:3.20

WORKDIR /app

# 安装必要的运行环境
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 从编译阶段拷贝二进制文件
COPY --from=builder /app/gvb-server /app/gvb-server

# [注意] 确保当前目录下有 settings.docker.yaml
COPY settings.docker.yaml /app/settings.yaml

EXPOSE 8080

# 启动命令
ENTRYPOINT ["/app/gvb-server"]