#!/bin/bash

# 检查是否提供了token参数
if [ $# -ne 1 ]; then
    echo "请提供一个token作为参数"
    exit 1
fi

token=$1

# 设置测试的目标URL
url="http://localhost:8080/api/user/friends"

# 设置并发数和请求总数
concurrency=100
total_requests=10000
# 没加redis之前，10000条需要13.5s  重启后 5.49s   5.42s
# 加redis后，10000条需要9.4s       重启后 2.87s   3.05s
# dbProxy直接返回， 10000条需要 8s
# 不调用rpc直接返回， 10000条需要 7.43s
# 直接返回，  10000条需要 7.23s

# 执行压力测试，添加包含token的Cookie
go-wrk -c $concurrency -n $total_requests -H "Cookie: token=Bearer $token" $url