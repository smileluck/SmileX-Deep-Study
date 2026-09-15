# SmileX-Deep-Study 一键构建控制
GO    ?= go
PNPM  ?= pnpm
BIN   := deep-study
GOOS  ?= linux
GOARCH ?= amd64

.PHONY: help dev build cross clean

help: ## 显示可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-8s %s\n", $$1, $$2}'

dev: ## 一键调试：Go 后端 :5574 + Vite 前端 :5573（Ctrl+C 同时退出）
	@test -d web/node_modules || (cd web && $(PNPM) install)
	@tmp=$$(mktemp -d); \
	trap 'kill %1 2>/dev/null; rm -rf "$$tmp"' EXIT INT TERM; \
	$(GO) build -o "$$tmp/$(BIN)" ./server/cmd/server && "$$tmp/$(BIN)" & \
	cd web && $(PNPM) dev

web-build: ## 仅构建前端（web/dist）
	cd web && $(PNPM) install --frozen-lockfile && $(PNPM) build

build: web-build ## 一键打包本机平台：前端构建 + go:embed 内嵌（含 harness 工作区脚手架资产）→ ./deep-study
	./scripts/sync-adapters.sh
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o $(BIN) ./server/cmd/server
	@echo "打包完成：./$(BIN)  （运行后访问 http://127.0.0.1:5574；启动时自动铺出 AGENTS.md/.agents/.codebuddy 等工作区文件，非空目录自动收进 deepstudy/ 子目录，二进制留在原地）"

cross: web-build ## 交叉编译：make cross GOOS=linux GOARCH=amd64 → ./deep-study-linux-amd64（含脚手架资产）
	./scripts/sync-adapters.sh
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -trimpath -ldflags="-s -w" -o $(BIN)-$(GOOS)-$(GOARCH) ./server/cmd/server
	@echo "打包完成：./$(BIN)-$(GOOS)-$(GOARCH)"

clean: ## 清理构建产物
	rm -f $(BIN) $(BIN)-* && rm -rf web/dist
