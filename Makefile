.PHONY: gen api build test tidy clean swagger web-dev web-install web-build web-preview regen

# GORM Gen — 根据 PO 模型生成类型安全的 Query 代码
gen:
	go run cmd/gen/main.go

# goctl API — 根据 .api 文件生成 handler/logic/types
api:
	goctl api go -api api/entry.api -dir . -style goZero

# 构建
build:
	go build -o bin/aipivot.exe .

# 测试
test:
	go test ./...

# 依赖整理
tidy:
	go mod tidy

# 清理生成产物
clean:
	@if exist bin rd /s /q bin
	@if exist web\dist rd /s /q web\dist

# Swagger 文档生成 — 基于 .api 文件生成 OpenAPI 2.0 JSON（需 goctl >= 1.8.2）
swagger:
	@if not exist docs\swagger mkdir docs\swagger
	goctl api swagger --api api/entry.api --dir docs/swagger --filename aipivot
	@echo "Swagger doc generated: docs/swagger/aipivot.json"

# 一键重新生成所有代码（Gen + API）
regen: gen api tidy
	@echo "All code regenerated."

# ==================== 前端（web/）====================

# 安装前端依赖
web-install:
	cd web && npm install

# 启动前端开发服务器（Vite，端口 5173，自动代理 /api → 后端 8888）
web-dev:
	cd web && npm run dev

# 构建前端生产包
web-build:
	cd web && npm run build

# 预览前端生产构建
web-preview:
	cd web && npm run preview
