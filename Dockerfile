FROM golang as builder

WORKDIR /app
COPY . /app

# 编译Lustre CSI Driver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/lustre-csi-driver ./cmd/main.go

# 创建最终的运行镜像
## yum install libnl3
## docker run -v /root:/mnt --privileged=true --net=host -itd --name rockylinux rockylinux:8  /bin/bash
#ctr -n k8s.io run -t --privileged --detach docker.io/library/lustre-client:v1.0.0 lustre-client1 /bin/bash
#ctr -n k8s.io task exec --exec-id exec-1 -t lustre-client1 /bin/bash
#ctr -n k8s.io i import lustre-csi-driver.tar
#FROM   rockylinux:8
#WORKDIR /app
#COPY ./wistor-client-2.15.2_4.18.0-425.3.1_ofed5.8.3_0718.bin /app
#制作镜像
#docker commit my-rockylinux my-rockylinux-image
#打包
#docker save -o my-rockylinux-image.tar my-rockylinux-image
#加载
#docker load -i my-rockylinux-image.tar
#ctr -n k8s.io images import my-rockylinux-image.tar
# ctr -n k8s.io images list
# ctr -n k8s.io tasks ls

# docker run -itd --privileged --name test   lustre-csi-driver:v1.0.0   /bin/bash

FROM rockylinux:8.7

WORKDIR /app

COPY ./rpm /app

RUN sed -i 's|^mirrorlist=|#mirrorlist=|g' /etc/yum.repos.d/Rocky* \
    && sed -i 's|^#baseurl=http://dl.rockylinux.org/\$contentdir|baseurl=https://mirrors.aliyun.com/rockylinux|g' /etc/yum.repos.d/Rocky* \
    && yum clean all \
    && yum makecache

# Install dependencies (including libnl3, dracut, and findutils)
RUN yum install -y kmod dracut libnl3 findutils

# Install Lustre client RPMs (ensure dependencies are resolved)
RUN rpm -ivh --nodeps ./kmod-lustre-client-2.15.5-1.el8.x86_64.rpm \
    && rpm -ivh --nodeps ./lustre-client-2.15.5-1.el8.x86_64.rpm



# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/lustre-csi-driver /bin/lustre-csi-driver

# 设置执行入口
ENTRYPOINT ["/bin/lustre-csi-driver"]

# docker build -t lustre-csi-driver:v1.0.0 .
