# docker build -t test-standalone:latest -f Dockerfile .
FROM alpine:latest

COPY test-standalone-linux /
CMD ["/test-standalone-linux", "/kube.config"]
