# Go 钱包服务

使用 Go 构建的简单 REST API 和 gRPC 钱包服务，采用内存存储。
需求文档地址： [需求文档](https://gist.github.com/imwithye/c31dc1a76dbcbf8b221813a2c8ac26e2)

## 功能特性

- 创建具有唯一 ID 的钱包
- 查询钱包余额
- 钱包间转账
- 线程安全的内存存储
- 完整的转账验证
- **同时支持 REST API 和 gRPC**
- **Docker 支持**

## 前置要求

- Go 1.21 或更高版本

## 安装

1. 克隆仓库
2. 安装依赖：
   ```bash
   go mod download
   ```

## 启动服务器

```bash
go run ./cmd/server
```

服务器将启动于:
- REST API: `http://localhost:8080`
- gRPC API: `localhost:9090`

## API 接口

### 1. 创建钱包

**请求：**
```bash
curl -X POST http://localhost:8080/wallets
```

**响应：**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 2. 获取钱包

**请求：**
```bash
curl http://localhost:8080/wallets/{wallet_id}
```

**响应：**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "balance": 0
}
```

### 3. 转账

**请求：**
```bash
curl -X POST http://localhost:8080/wallets/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "from_wallet_id": "550e8400-e29b-41d4-a716-446655440000",
    "to_wallet_id": "650e8400-e29b-41d4-a716-446655440001",
    "amount": 100
  }'
```

**响应：**
```json
{
  "message": "转账成功",
  "from": "550e8400-e29b-41d4-a716-446655440000",
  "to": "650e8400-e29b-41d4-a716-446655440001",
  "amount": 100
}
```

## gRPC API

服务同时支持 gRPC API（端口 9090），使用相同的业务逻辑。

### 使用 grpcurl 测试 gRPC

```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出所有服务
grpcurl -plaintext localhost:9090 list

# 列出服务方法
grpcurl -plaintext localhost:9090 list wallet.WalletService

# 创建钱包
grpcurl -plaintext -d '{}' localhost:9090 wallet.WalletService/CreateWallet

# 获取钱包
grpcurl -plaintext -d '{"id": "your-wallet-id"}' localhost:9090 wallet.WalletService/GetWallet

# 转账
grpcurl -plaintext -d '{
  "from_wallet_id": "wallet-id-1",
  "to_wallet_id": "wallet-id-2",
  "amount": 100
}' localhost:9090 wallet.WalletService/Transfer
```

### 使用 Go 客户端调用 gRPC

```go
import (
    "context"
    "google.golang.org/grpc"
    "interview/proto"
)

func main() {
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := proto.NewWalletServiceClient(conn)

    // 创建钱包
    resp, err := client.CreateWallet(context.Background(), &proto.CreateWalletRequest{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Wallet ID:", resp.Id)
}
```

## Docker 支持

### 使用 Docker 构建

```bash
# 构建镜像
docker build -t wallet-service .

# 运行容器
docker run -p 8080:8080 -p 9090:9090 wallet-service
```

### 使用 Docker Compose

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 错误响应

API 返回相应的 HTTP 状态码：

- `400 Bad Request` - 无效输入（负数金额、同一钱包转账等）
- `404 Not Found` - 钱包不存在
- `405 Method Not Allowed` - 请求方法错误
- `500 Internal Server Error` - 服务器错误

**错误响应格式：**
```json
{
  "error": "错误信息"
}
```

## 测试示例

```bash
# 启动服务器
go run ./cmd/server &

# 创建两个钱包
WALLET1=$(curl -s -X POST http://localhost:8080/wallets | jq -r '.id')
WALLET2=$(curl -s -X POST http://localhost:8080/wallets | jq -r '.id')

echo "钱包 1: $WALLET1"
echo "钱包 2: $WALLET2"

# 查询余额（应该为 0）
curl http://localhost:8080/wallets/$WALLET1
curl http://localhost:8080/wallets/$WALLET2

# 以下请求会失败 - 余额不足
curl -X POST http://localhost:8080/wallets/transfer \
  -H "Content-Type: application/json" \
  -d "{\"from_wallet_id\":\"$WALLET1\", \"to_wallet_id\":\"$WALLET2\", \"amount\":100}"
```

## 项目结构

```
interview/
├── cmd/
│   └── server/
│       └── main.go           # 服务器入口（REST + gRPC）
├── internal/
│   ├── handler/
│   │   └── wallet.go         # HTTP 处理器
│   ├── grpc/
│   │   └── wallet.go         # gRPC 处理器
│   ├── service/
│   │   └── wallet.go         # 业务逻辑（共享）
│   ├── model/
│   │   └── wallet.go         # 数据模型
│   └── storage/
│       └── memory.go         # 内存存储
├── proto/
│   ├── wallet.proto          # Protocol Buffers 定义
│   ├── wallet.pb.go          # 生成的 protobuf 代码
│   └── wallet_grpc.pb.go     # 生成的 gRPC 代码
├── Dockerfile                # Docker 镜像定义
├── docker-compose.yml        # Docker Compose 配置
├── go.mod
└── README.md
```

## 架构说明

- **REST 和 gRPC 共享业务逻辑**: `internal/service/wallet.go` 中的 `WalletService` 被 `internal/handler/wallet.go` (REST) 和 `internal/grpc/wallet.go` (gRPC) 共同使用
- **端口配置**:
  - REST API: 8080
  - gRPC API: 9090
