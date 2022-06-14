# 编译两个平台的二进制静态文件，即不链接任何动态链接库。
binary_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags '-s -w --extldflags "-static -fpic"' -o test-standalone-linux; \
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
	go build -ldflags '-s -w --extldflags "-static -fpic"' -o test-standalone-darwin