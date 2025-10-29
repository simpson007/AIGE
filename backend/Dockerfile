# ---------- 构建阶段 ----------
FROM golang:1.24-alpine AS builder

# 多源备份 + 强制 IPv4
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache git gcc musl-dev

# 2) 装构建工具
RUN apk add --no-cache git gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o aige-server .

# ---------- 运行阶段 ----------
FROM alpine:latest

# 同样换源 + 装运行时依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates sqlite-libs

# 创建用户
RUN addgroup -g 1000 aige && \
    adduser -D -u 1000 -G aige aige

WORKDIR /app
COPY --from=builder /app/aige-server .
RUN mkdir -p /app/data && chown -R aige:aige /app

USER aige
EXPOSE 8182

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8182/health || exit 1

CMD ["./aige-server"]