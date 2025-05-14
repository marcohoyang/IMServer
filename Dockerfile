# 使用最小的基础镜像
FROM scratch

# 将本地编译好的二进制文件复制到镜像中
COPY ./build/main /msg-server

# 指定容器启动时执行的命令
ENTRYPOINT ["/msg-server"]