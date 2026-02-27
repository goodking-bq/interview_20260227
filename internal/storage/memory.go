package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"interview/internal/model"
)

// MemoryStorage 提供线程安全的钱包内存存储
type MemoryStorage struct {
	mu      sync.RWMutex
	wallets map[string]*model.Wallet
}

// NewMemoryStorage 创建一个新的内存存储实例
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		wallets: make(map[string]*model.Wallet),
	}
}

// Create 创建一个余额为零的新钱包
func (s *MemoryStorage) Create() (*model.Wallet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()
	now := time.Now()
	wallet := &model.Wallet{
		ID:      id,
		Balance: 0,
		Created: now,
	}
	s.wallets[id] = wallet

	return wallet, nil
}

// Get 根据ID获取钱包
func (s *MemoryStorage) Get(id string) (*model.Wallet, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallet, exists := s.wallets[id]
	if !exists {
		return nil, fmt.Errorf("钱包不存在: %s", id)
	}

	return wallet, nil
}

// Update 更新现有钱包
func (s *MemoryStorage) Update(wallet *model.Wallet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.wallets[wallet.ID]; !exists {
		return fmt.Errorf("钱包不存在: %s", wallet.ID)
	}

	s.wallets[wallet.ID] = wallet
	return nil
}

// GetAll 返回所有钱包
func (s *MemoryStorage) GetAll() []*model.Wallet {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wallets := make([]*model.Wallet, 0, len(s.wallets))
	for _, wallet := range s.wallets {
		wallets = append(wallets, wallet)
	}

	return wallets
}

// TransferAtomically 在两个钱包间执行原子转账操作
// 此方法在整个操作期间持有锁，以防止竞态条件
func (s *MemoryStorage) TransferAtomically(fromID, toID string, amount int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fromWallet, fromExists := s.wallets[fromID]
	toWallet, toExists := s.wallets[toID]

	if !fromExists || !toExists {
		return fmt.Errorf("一个或两个钱包不存在")
	}

	if fromWallet.Balance < amount {
		return fmt.Errorf("余额不足")
	}

	// 原子执行转账
	fromWallet.Balance -= amount
	toWallet.Balance += amount

	return nil
}
