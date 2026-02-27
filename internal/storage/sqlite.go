package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
	"github.com/google/uuid"
	"interview/internal/model"
)

// SQLiteStorage 使用 SQLite 数据库存储钱包数据
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage 创建一个新的 SQLite 存储实例
// DSN 是 SQLite 数据库文件路径，例如 "./wallet.db"
func NewSQLiteStorage(dsn string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("打开 SQLite 数据库失败: %w", err)
	}

	// 启用 WAL 模式以提高并发性能
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, fmt.Errorf("设置 WAL 模式失败: %w", err)
	}

	// 创建表
	if err := createTable(db); err != nil {
		return nil, fmt.Errorf("创建表失败: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

// createTable 创建钱包表
func createTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS wallets (
		id TEXT PRIMARY KEY,
		balance INTEGER NOT NULL DEFAULT 0,
		created DATETIME NOT NULL
	);
	`
	_, err := db.Exec(query)
	return err
}

// Create 创建一个余额为零的新钱包
func (s *SQLiteStorage) Create() (*model.Wallet, error) {
	id := uuid.New().String()
	now := time.Now()

	query := `INSERT INTO wallets (id, balance, created) VALUES (?, ?, ?)`
	_, err := s.db.Exec(query, id, 0, now)
	if err != nil {
		return nil, fmt.Errorf("插入钱包失败: %w", err)
	}

	return &model.Wallet{
		ID:      id,
		Balance: 0,
		Created: now,
	}, nil
}

// Get 根据ID获取钱包
func (s *SQLiteStorage) Get(id string) (*model.Wallet, error) {
	query := `SELECT id, balance, created FROM wallets WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var wallet model.Wallet
	err := row.Scan(&wallet.ID, &wallet.Balance, &wallet.Created)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("钱包不存在: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("查询钱包失败: %w", err)
	}

	return &wallet, nil
}

// Update 更新现有钱包
func (s *SQLiteStorage) Update(wallet *model.Wallet) error {
	query := `UPDATE wallets SET balance = ? WHERE id = ?`
	result, err := s.db.Exec(query, wallet.Balance, wallet.ID)
	if err != nil {
		return fmt.Errorf("更新钱包失败: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取影响行数失败: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("钱包不存在: %s", wallet.ID)
	}

	return nil
}

// GetAll 返回所有钱包
func (s *SQLiteStorage) GetAll() []*model.Wallet {
	query := `SELECT id, balance, created FROM wallets ORDER BY created DESC`
	rows, err := s.db.Query(query)
	if err != nil {
		return []*model.Wallet{}
	}
	defer rows.Close()

	wallets := []*model.Wallet{}
	for rows.Next() {
		var wallet model.Wallet
		if err := rows.Scan(&wallet.ID, &wallet.Balance, &wallet.Created); err != nil {
			continue
		}
		wallets = append(wallets, &wallet)
	}

	return wallets
}

// TransferAtomically 在两个钱包间执行原子转账操作
// 使用数据库事务确保原子性
func (s *SQLiteStorage) TransferAtomically(fromID, toID string, amount int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	// 检查源钱包是否存在并获取余额
	var fromBalance int64
	err = tx.QueryRow(`SELECT balance FROM wallets WHERE id = ?`, fromID).Scan(&fromBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("一个或两个钱包不存在")
	}
	if err != nil {
		return fmt.Errorf("查询源钱包失败: %w", err)
	}

	// 检查目标钱包是否存在
	var toBalance int64
	err = tx.QueryRow(`SELECT balance FROM wallets WHERE id = ?`, toID).Scan(&toBalance)
	if err == sql.ErrNoRows {
		return fmt.Errorf("一个或两个钱包不存在")
	}
	if err != nil {
		return fmt.Errorf("查询目标钱包失败: %w", err)
	}

	// 检查余额是否足够
	if fromBalance < amount {
		return fmt.Errorf("余额不足")
	}

	// 执行转账
	_, err = tx.Exec(`UPDATE wallets SET balance = balance - ? WHERE id = ?`, amount, fromID)
	if err != nil {
		return fmt.Errorf("更新源钱包余额失败: %w", err)
	}

	_, err = tx.Exec(`UPDATE wallets SET balance = balance + ? WHERE id = ?`, amount, toID)
	if err != nil {
		return fmt.Errorf("更新目标钱包余额失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
