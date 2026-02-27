package model

import "time"

// Wallet 表示用户的钱包，包含唯一ID和余额
type Wallet struct {
	ID      string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Balance int64     `json:"balance" example:"1000"`
	Created time.Time `json:"created" example:"2024-01-01T00:00:00Z"`
}

// TransferRequest 表示钱包间转账的请求
type TransferRequest struct {
	FromWalletID string `json:"from_wallet_id" example:"550e8400-e29b-41d4-a716-446655440000" binding:"required"`
	ToWalletID   string `json:"to_wallet_id" example:"660e8400-e29b-41d4-a716-446655440001" binding:"required"`
	Amount       int64  `json:"amount" example:"100" binding:"required,gt=0"`
}

// WalletResponse 表示获取钱包时的响应
type WalletResponse struct {
	ID      string `json:"id"`
	Balance int64  `json:"balance"`
}

// CreateWalletResponse 表示创建钱包时的响应
type CreateWalletResponse struct {
	ID string `json:"id"`
}

// ErrorResponse 表示错误响应
type ErrorResponse struct {
	Error string `json:"error"`
}

// TransferResponse 表示转账成功的响应
type TransferResponse struct {
	Message string `json:"message"`
	From    string `json:"from"`
	To      string `json:"to"`
	Amount  int64  `json:"amount"`
}
