.PHONY: all build run clean test

# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# 二进制文件名
BINARY_NAME=narutoscript-next
BINARY_WINDOWS=$(BINARY_NAME).exe

# 主入口
MAIN_PATH=./cmd/main.go

# 构建
build:
	$(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME) $(MAIN_PATH)

# Windows 构建
build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_WINDOWS) $(MAIN_PATH)

# macOS 构建
build-darwin:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

# Linux 构建
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME)-linux $(MAIN_PATH)

# 所有平台
build-all: build-windows build-darwin build-linux

# 前端构建
build-frontend:
	cd web && npm install && npm run build

# 运行
run:
	$(GOCMD) run $(MAIN_PATH)

# 测试
test:
	$(GOTEST) -v ./...

# 清理
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_WINDOWS)
	rm -f $(BINARY_NAME)-*

# 安装依赖
deps:
	$(GOGET) ./...

# 开发模式
dev:
	$(GOCMD) run $(MAIN_PATH)