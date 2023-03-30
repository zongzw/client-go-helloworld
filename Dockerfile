# docker build -t test-standalone:latest -f Dockerfile .
FROM alpine:3.17.3

COPY test-standalone-linux /
CMD ["/test-standalone-linux", "/kube.config"]
