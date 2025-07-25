#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 项目配置 - 脚本与src同级，直接使用当前目录作为项目根
PROJECT_ROOT=$(realpath "$(dirname "$0")")
PROTO_DIR="$PROJECT_ROOT/src/proto"
OUTPUT_DIR="$PROTO_DIR"
BUILD_DIR="$PROJECT_ROOT/build"
DBPROXY_DIR="$PROJECT_ROOT/src/dbproxy"
ENV_FILE="$DBPROXY_DIR/.env"  # 环境文件路径
MAIN_DIR="$PROJECT_ROOT/src"

# 调试输出：显示关键路径
echo -e "${YELLOW}项目根目录: $PROJECT_ROOT${NC}"
echo -e "${YELLOW}PROTO目录: $PROTO_DIR${NC}"
echo -e "${YELLOW}DBProxy目录: $DBPROXY_DIR${NC}"
echo -e "${YELLOW}环境文件: $ENV_FILE${NC}"

# 确保目录存在
mkdir -p "$OUTPUT_DIR"
mkdir -p "$BUILD_DIR"

# 检查必要的工具
check_tool() {
    command -v "$1" >/dev/null 2>&1 || { 
        echo -e "${RED}错误：需要安装$1${NC}"
        if [ "$1" = "protoc" ]; then
            echo -e "${YELLOW}请从 https://github.com/protocolbuffers/protobuf/releases 下载安装${NC}"
        else
            echo -e "${YELLOW}执行: go install $2@latest${NC}"
        fi
        exit 1 
    }
}

echo -e "${YELLOW}检查依赖工具...${NC}"
check_tool protoc
check_tool protoc-gen-go "google.golang.org/protobuf/cmd/protoc-gen-go"
check_tool protoc-gen-go-grpc "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
check_tool swag "github.com/swaggo/swag/cmd/swag"  # 新增Swagger工具检查

# 生成gRPC代码
echo -e "${YELLOW}正在生成gRPC代码...${NC}"
if [ ! -d "$PROTO_DIR" ]; then
    echo -e "${RED}错误：proto目录 $PROTO_DIR 不存在！${NC}"
    exit 1
fi

pushd "$PROTO_DIR" >/dev/null || exit
for proto_file in *.proto; do
    if [ -f "$proto_file" ]; then
        echo -e "${GREEN}处理 $proto_file...${NC}"
        protoc --go_out=paths=source_relative:"$OUTPUT_DIR" \
               --go-grpc_out=paths=source_relative:"$OUTPUT_DIR" \
               "$proto_file"
        
        if [ $? -ne 0 ]; then
            echo -e "${RED}生成 $proto_file 失败${NC}"
            popd >/dev/null || exit
            exit 1
        fi
    fi
done
popd >/dev/null || exit

# 生成Swagger文档
echo -e "${YELLOW}生成Swagger文档...${NC}"
pushd "$MAIN_DIR" >/dev/null || exit
swag init -g main.go --output "$MAIN_DIR/docs"
if [ $? -ne 0 ]; then
    echo -e "${RED}生成Swagger文档失败${NC}"
    popd >/dev/null || exit
    exit 1
else
    echo -e "${GREEN}Swagger文档生成成功${NC}"
fi
popd >/dev/null || exit

# 格式化代码
echo -e "${YELLOW}格式化生成的代码...${NC}"
go fmt ./... >/dev/null

# 检查依赖
echo -e "${YELLOW}检查依赖...${NC}"
go mod tidy

# 构建可执行文件
echo -e "${YELLOW}构建可执行文件...${NC}"
BUILD_TARGETS=(
    "$DBPROXY_DIR/dbproxy.go"
    "$MAIN_DIR/main.go"
)

for target in "${BUILD_TARGETS[@]}"; do
    echo -e "${YELLOW}检查文件: $target${NC}"
    
    if [ -f "$target" ]; then
        output_name=$(basename "$target" .go)
        echo -e "${GREEN}构建 $target -> $BUILD_DIR/$output_name${NC}"
        
        (cd "$(dirname "$target")" && go build  -o "$BUILD_DIR/$output_name" "$(basename "$target")")
        
        if [ $? -ne 0 ]; then
            echo -e "${RED}构建 $target 失败${NC}"
            exit 1
        fi
    else
        echo -e "${RED}错误：文件 $target 不存在！${NC}"
        echo -e "${YELLOW}目录内容:${NC}"
        ls -la "$(dirname "$target")"
        exit 1
    fi
done

# 可选：运行应用程序
echo -e "${YELLOW}=========================${NC}"
echo -e "${YELLOW}构建完成!${NC}"
echo -e "${YELLOW}生成的代码位于: $OUTPUT_DIR${NC}"
echo -e "${YELLOW}可执行文件位于: $BUILD_DIR${NC}"
echo -e "${YELLOW}=========================${NC}"

echo -e "${GREEN}是否需要运行应用程序? (y/n)${NC}"
read -r run_option

if [ "$run_option" = "y" ] || [ "$run_option" = "Y" ]; then
    if [ ! -f "$BUILD_DIR/main" ]; then
        echo -e "${RED}找不到 main 可执行文件${NC}"
        exit 1
    fi

    if [ ! -f "$BUILD_DIR/dbproxy" ]; then
        echo -e "${RED}找不到 dbproxy 可执行文件${NC}"
        exit 1
    fi

    echo -e "${GREEN}正在启动服务...${NC}"

    # 使用绝对路径启动
    "$BUILD_DIR/main" &
    MAIN_PID=$!

    "$BUILD_DIR/dbproxy" &
    DBPROXY_PID=$!

    echo "服务已启动。PID: main=$MAIN_PID, dbproxy=$DBPROXY_PID"

    # 等待子进程结束（可选）
    wait
fi