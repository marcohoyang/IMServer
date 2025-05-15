# 使用官方的最小化运行时作为基础镜像
FROM ubuntu:22.04

# 创建应用的工作目录
WORKDIR /app

# 将可执行文件复制到容器中
COPY ../../build/dbproxy /app/dbproxy

# 如果有配置文件，也一并复制

COPY ../../config.yaml /app/config.yaml

# 暴露应用程序所需的端口（如果有的话）
EXPOSE 50001

RUN chmod +x /app/dbproxy

# 设置环境变量（如果有需要的话）
ENV APP_ENV=production

# 定义启动命令
CMD ["./dbproxy"]