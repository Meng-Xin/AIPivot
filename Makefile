.PHONY: gen api build test tidy clean swagger

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

# Swagger 文档生成 — 基于 .api 文件生成 OpenAPI 2.0 JSON（需 goctl >= 1.8.2）
swagger:
	@if not exist docs\swagger mkdir docs\swagger
	goctl api swagger --api api/entry.api --dir docs/swagger --filename aipivot
	@echo "Swagger doc generated: docs/swagger/aipivot.json"

# 一键重新生成所有代码（Gen + API）
regen: gen api tidy
	@echo "All code regenerated."
