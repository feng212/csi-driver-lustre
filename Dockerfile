FROM golang as builder

WORKDIR /app
COPY . /app

# 编译Lustre CSI Driver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/lustre-csi-driver ./cmd/main.go

# 创建最终的运行镜像
FROM alpine
COPY --from=builder /app/lustre-csi-driver /usr/local/bin/lustre-csi-driver

ENTRYPOINT ["/usr/local/bin/lustre-csi-driver"]
