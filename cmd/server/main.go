package main

import (
	"fmt"
	"log"
	"net"
	"net/http"

	grpchandler "interview/internal/grpc"
	"interview/internal/handler"
	"interview/internal/service"
	"interview/internal/storage"
	"interview/proto"

	grpcpkg "google.golang.org/grpc"

	_ "interview/docs" // Swagger 文档
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title 钱包服务 API
// @version 1.0
// @description 这是一个提供钱包创建、查询和转账功能的 REST API 服务
// @termsOfService http://swagger.io/terms/

// @contact.name API 支持
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /
// @schemes http

func main() {
	// 从环境变量加载存储配置
	cfg := storage.LoadConfigFromEnv()

	// 使用工厂函数创建存储
	store, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("创建存储失败: %v", err)
	}

	// 初始化服务
	walletService := service.NewWalletService(store)

	log.Printf("使用存储类型: %s", cfg.Type)

	// 启动 REST API 服务器
	go startRESTServer(walletService)

	// 启动 gRPC 服务器
	startGRPCServer(walletService)
}

func startRESTServer(walletService *service.WalletService) {
	// 初始化处理器
	walletHandler := handler.NewWalletHandler(walletService)

	// 设置路由
	http.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			walletHandler.CreateWalletHandler(w, r)
		} else {
			http.Error(w, "不允许的请求方法", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/wallets/", func(w http.ResponseWriter, r *http.Request) {
		// 检查是否为转账接口
		if r.URL.Path == "/wallets/transfer" {
			if r.Method == http.MethodPost {
				walletHandler.TransferHandler(w, r)
			} else {
				http.Error(w, "不允许的请求方法", http.StatusMethodNotAllowed)
			}
			return
		}
		// 否则作为 GET /wallets/{id} 处理
		walletHandler.GetWalletHandler(w, r)
	})

	// Swagger 文档路由
	http.Handle("/swagger/", httpSwagger.WrapHandler)

	// 健康检查端点
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 启动服务器
	port := "8080"
	addr := fmt.Sprintf(":%s", port)
	log.Printf("REST API 服务器启动于 %s", addr)
	log.Printf("Swagger 文档地址: http://localhost%s/swagger/index.html", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("REST 服务器启动失败: %v", err)
	}
}

func startGRPCServer(walletService *service.WalletService) {
	// 创建 gRPC 监听器
	port := "9090"
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("gRPC 监听失败: %v", err)
	}

	// 创建 gRPC 服务器
	grpcServer := grpcpkg.NewServer()

	// 注册钱包服务
	walletGrpcServer := grpchandler.NewWalletServer(walletService)
	proto.RegisterWalletServiceServer(grpcServer, walletGrpcServer)

	log.Printf("gRPC 服务器启动于 :%s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("gRPC 服务器启动失败: %v", err)
	}
}
