
`编译可执行文件`
./build.sh

`用dockersfile创建dbproxy image`
docker build -f Dockerfile.dbproxy  -t dbproxy.v1.0 .

`用dockersfile创建imserver image`
docker build -f Dockerfile.im  -t imserver.v1.0 .

`启动所有服务（默认前台运行）`
 docker-compose up

`后台运行`
 docker-compose up -d

`查看状态`
 docker-compose ps

`重新build image`
 sudo docker-compose build --no-cache app-name


`架构图`
 ![alt text](image.png)

graph TD
    subgraph Application Network
        im-server-1 -->|访问| dbproxy
        im-server-1 <-->|访问| redis-pubsub
        im-server-2 -->|访问| dbproxy
        im-server-2 <-->|访问| redis-pubsub
        dbproxy -->|访问| mysql
        dbproxy -->|访问| redis-cache
    end

    subgraph Proxy Network
        nginx-proxy -->|代理请求至| im-server-1
        nginx-proxy -->|代理请求至| im-server-2
    end

    mysql-.->|持久化存储| mysql_data
