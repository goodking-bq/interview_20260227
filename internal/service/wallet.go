package service

import (
	"errors"

	"interview/internal/model"
	"interview/internal/storage"
)

var (
	ErrWalletNotFound      = errors.New("钱包不存在")
	ErrInsufficientBalance = errors.New("余额不足")
	ErrInvalidAmount       = errors.New("金额必须为正数")
	ErrSameWallet          = errors.New("不能转账到同一钱包")
)

// WalletService 提供钱包操作的业务逻辑
type WalletService struct {
	storage storage.Storage
}

// NewWalletService 创建一个新的钱包服务
func NewWalletService(storage storage.Storage) *WalletService {
	return &WalletService{
		storage: storage,
	}
}

// CreateWallet 创建一个余额为零的新钱包
func (s *WalletService) CreateWallet() (*model.Wallet, error) {
	return s.storage.Create()
}

// GetWallet 根据ID获取钱包
func (s *WalletService) GetWallet(id string) (*model.Wallet, error) {
	wallet, err := s.storage.Get(id)
	if err != nil {
		return nil, ErrWalletNotFound
	}
	return wallet, nil
}

// Transfer 从一个钱包向另一个钱包转账
func (s *WalletService) Transfer(fromID, toID string, amount int64) (*model.TransferResponse, error) {
	// 验证金额
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	// 不能转到同一钱包
	if fromID == toID {
		return nil, ErrSameWallet
	}

	// 执行原子转账 - 存储层处理锁定和验证
	if err := s.storage.TransferAtomically(fromID, toID, amount); err != nil {
		if err.Error() == "一个或两个钱包不存在" {
			return nil, ErrWalletNotFound
		}
		if err.Error() == "余额不足" {
			return nil, ErrInsufficientBalance
		}
		return nil, err
	}

	return &model.TransferResponse{
		Message: "转账成功",
		From:    fromID,
		To:      toID,
		Amount:  amount,
	}, nil
}
