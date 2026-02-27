package storage

import "interview/internal/model"

// Storage 定义钱包存储的抽象接口
// 不同的存储实现（内存、SQLite、Redis）都需要实现此接口
type Storage interface {
	// Create 创建一个余额为零的新钱包
	Create() (*model.Wallet, error)

	// Get 根据ID获取钱包
	Get(id string) (*model.Wallet, error)

	// Update 更新现有钱包
	Update(wallet *model.Wallet) error

	// GetAll 返回所有钱包
	GetAll() []*model.Wallet

	// TransferAtomically 在两个钱包间执行原子转账操作
	// 此方法必须确保转账操作的原子性，防止竞态条件
	TransferAtomically(fromID, toID string, amount int64) error
}
