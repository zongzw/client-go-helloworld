# docker build -t test-standalone:latest -f Dockerfile .
#   因为编译出来的test-standalone-linux 不存在外部依赖，我们可以使用alpine:latest
#   作为base image，这样做出来的docker image只有40.3MB
FROM alpine:latest

COPY test-standalone-linux /
# 默认使用/kube.config 作为启动参数，详见程序实现main.go 和 docker-compose.yml 的卷部分volumes
CMD ["/test-standalone-linux", "/kube.config"]
