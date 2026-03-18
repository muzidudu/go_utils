# Fiber Makefile - 在 fiber 目录下执行 make
# 用法: make [target]

.PHONY: help sync build run deploy clean test lint

help:
	@echo "Fiber 自动化命令"
	@echo ""
	@echo "  make build         - 构建应用 (输出 dist/)"
	@echo "  make run           - 本地运行"
	@echo "  make deploy        - 构建并部署到 deploy/"
	@echo "  make clean         - 清理构建产物"
	@echo "  make sync          - 同步 go.work 依赖"
	@echo "  make test          - 运行 fiber 测试"
	@echo "  make lint          - golangci-lint"
	@echo ""
	@echo "示例: make build && make deploy"

sync:
	@cd .. && go work sync

build:
	@./scripts/build.sh

run:
	@./scripts/run.sh

deploy:
	@./scripts/deploy.sh

clean:
	@./scripts/clean.sh

test:
	@go test ./...

lint:
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed, skip"
