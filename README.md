
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