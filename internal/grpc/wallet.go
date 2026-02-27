package grpc

import (
	"context"

	"interview/internal/service"
	"interview/proto"
)

// WalletServer 实现 gRPC WalletService 接口
type WalletServer struct {
	proto.UnimplementedWalletServiceServer
	service *service.WalletService
}

// NewWalletServer 创建一个新的 gRPC 钱包服务器
func NewWalletServer(service *service.WalletService) *WalletServer {
	return &WalletServer{
		service: service,
	}
}

// CreateWallet 创建一个余额为零的新钱包
func (s *WalletServer) CreateWallet(ctx context.Context, req *proto.CreateWalletRequest) (*proto.CreateWalletResponse, error) {
	wallet, err := s.service.CreateWallet()
	if err != nil {
		return nil, err
	}
	return &proto.CreateWalletResponse{
		Id: wallet.ID,
	}, nil
}

// GetWallet 根据 ID 获取钱包信息
func (s *WalletServer) GetWallet(ctx context.Context, req *proto.GetWalletRequest) (*proto.GetWalletResponse, error) {
	wallet, err := s.service.GetWallet(req.Id)
	if err != nil {
		return nil, err
	}
	return &proto.GetWalletResponse{
		Id:      wallet.ID,
		Balance: wallet.Balance,
	}, nil
}

// Transfer 从一个钱包向另一个钱包转账
func (s *WalletServer) Transfer(ctx context.Context, req *proto.TransferRequest) (*proto.TransferResponse, error) {
	response, err := s.service.Transfer(req.FromWalletId, req.ToWalletId, req.Amount)
	if err != nil {
		return nil, err
	}
	return &proto.TransferResponse{
		Message: response.Message,
		From:    response.From,
		To:      response.To,
		Amount:  response.Amount,
	}, nil
}
