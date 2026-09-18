# SmileX-Deep-Study 一键构建控制
GO    ?= go
PNPM  ?= pnpm
BIN   := deep-study
GOOS  ?= linux
GOARCH ?= amd64

ifeq ($(OS),Windows_NT)
# Windows 下 make 默认用 cmd.exe 执行配方，强制改用 Git Bash。
# 不能写 SHELL := bash.exe：PATH 查找可能命中 WSL 的 bash（System32/WindowsApps），
# 这里从 git.exe 的位置推导 Git 安装目录，锁定 Git 自带的 bash。
GIT_EXE  := $(subst \,/,$(firstword $(shell where git)))
GIT_DIR  := $(dir $(GIT_EXE))
GIT_BASH := $(firstword $(wildcard \
  $(abspath $(GIT_DIR)bash.exe) \
  $(abspath $(GIT_DIR)../bin/bash.exe) \
  $(abspath $(GIT_DIR)../usr/bin/bash.exe) \
  $(abspath $(GIT_DIR)../../bin/bash.exe) \
  $(abspath $(GIT_DIR)../../usr/bin/bash.exe) \
  $(ProgramFiles)/Git/bin/bash.exe \
  $(LOCALAPPDATA)/Programs/Git/bin/bash.exe))
ifeq ($(GIT_BASH),)
$(error 未找到 Git Bash：请安装 Git for Windows，或改用 Git Bash 终端运行 make)
endif
SHELL := $(GIT_BASH)
EXE := .exe
endif

.PHONY: help dev build cross clean

help: ## 显示可用目标
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  make %-8s %s\n", $$1, $$2}'

dev: ## 一键调试：Go 后端 :5574 + Vite 前端 :5573（Ctrl+C 同时退出）
	@test -d web/node_modules || (cd web && $(PNPM) install)
	@mkdir -p web/dist && touch web/dist/.gitkeep
	@tmp=$$(mktemp -d); \
	$(GO) build -o "$$tmp/$(BIN)$(EXE)" ./server/cmd/server || exit 1; \
	"$$tmp/$(BIN)$(EXE)" & pid=$$!; \
	trap 'kill $$pid 2>/dev/null; rm -rf "$$tmp"' EXIT INT TERM; \
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
