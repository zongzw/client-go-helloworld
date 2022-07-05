binary_build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags '-s -w --extldflags "-static -fpic"' -o test-standalone-linux; \
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 \
	go build -ldflags '-s -w --extldflags "-static -fpic"' -o test-standalone-darwin
