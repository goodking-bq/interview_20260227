# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个 Go 语言实现的钱包服务，同时提供 REST API 和 gRPC API 两种接口。服务使用内存存储，支持钱包创建、查询和转账功能。
项目的注释，日志，错误等都应该使用中文

## 核心架构

本项目采用分层架构，关键设计是 **REST 和 gRPC 共享同一业务逻辑层**：

```
┌─────────────────┐     ┌─────────────────┐
│   REST Handler  │     │   gRPC Handler  │
│  (HTTP/8080)    │     │  (gRPC/9090)    │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
         ┌───────────▼───────────┐
         │   WalletService       │
         │  (业务逻辑层，共享)    │
         └───────────┬───────────┘
                     │
         ┌───────────▼───────────┐
         │   MemoryStorage       │
         │  (线程安全的内存存储)  │
         └───────────────────────┘
```

**关键文件说明：**
- `cmd/server/main.go` - 程序入口，同时启动 REST 和 gRPC 服务器
- `internal/service/wallet.go` - **核心业务逻辑**，被 REST 和 gRPC 共用
- `internal/handler/wallet.go` - REST HTTP 处理器
- `internal/grpc/wallet.go` - gRPC 服务器实现
- `internal/storage/memory.go` - 内存存储，使用 `sync.RWMutex` 保证并发安全
- `proto/` - Protocol Buffers 定义和生成的 Go 代码

## 常用命令

### 构建和运行
```bash
# 运行服务
go run ./cmd/server

# 构建二进制文件
go build -o server ./cmd/server

# 构建所有包
go build ./...
```

### Proto 文件生成（如果修改了 .proto 文件）
```bash
# 安装 protoc 和相关插件（只需一次）
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 生成 Go 代码
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/wallet.proto
```

### Docker
```bash
# 构建镜像
docker build -t wallet-service .

# 运行容器
docker run -p 8080:8080 -p 9090:9090 wallet-service

# 使用 docker-compose
docker-compose up -d
```

## 代码组织约定

1. **业务逻辑必须在 `service` 层实现** - REST 和 gRPC 处理器只负责协议转换，不包含业务逻辑
2. **所有注释、日志、错误信息使用中文**
3. **错误处理** - service 层定义了标准错误：`ErrWalletNotFound`、`ErrInsufficientBalance`、`ErrInvalidAmount`、`ErrSameWallet`
4. **并发安全** - `MemoryStorage.TransferAtomically()` 方法在整个转账期间持有锁，确保原子性

## API 端口

- REST API: `8080`
- gRPC API: `9090`

## Swagger 文档

项目使用 Swagger 提供 API 文档，访问地址：`http://localhost:8080/swagger/index.html`

### 重新生成 Swagger 文档
```bash
# 安装 swag 工具（只需一次）
go install github.com/swaggo/swag/cmd/swag@latest

# 生成文档
~/go/bin/swag init -g cmd/server/main.go -o docs
```

### Swagger 注解说明
- 在 Handler 方法上使用 `@Summary`、`@Description`、`@Tags` 等注解
- 在 `cmd/server/main.go` 顶部添加 API 总体信息注解
- 模型中的 `example` tag 用于显示示例值
