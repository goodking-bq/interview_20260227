package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"interview/internal/model"
)

// RedisStorage 使用 Redis 存储钱包数据
type RedisStorage struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisStorage 创建一个新的 Redis 存储实例
// DSN 格式: "localhost:6379" 或带密码的 "localhost:6379,password=xxx"
func NewRedisStorage(dsn string) (*RedisStorage, error) {
	opt, err := parseRedisDSN(dsn)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %w", err)
	}

	return &RedisStorage{
		client: client,
		ctx:    ctx,
	}, nil
}

// parseRedisDSN 解析 Redis 连接字符串
func parseRedisDSN(dsn string) (*redis.Options, error) {
	opt := &redis.Options{
		Addr:     dsn,
		Password: "",
		DB:       0,
	}

	// 简单的解析逻辑，支持 "host:port" 和 "host:port,password=pass" 格式
	if len(dsn) > 0 {
		// 如果包含逗号，解析密码
		for i, c := range dsn {
			if c == ',' {
				opt.Addr = dsn[:i]
				// 解析密码参数
				params := dsn[i+1:]
				if len(params) > 9 && params[:9] == "password=" {
					opt.Password = params[9:]
				}
				break
			}
		}
	}

	return opt, nil
}

// walletKey 生成钱包在 Redis 中的 key
func walletKey(id string) string {
	return "wallet:" + id
}

// Create 创建一个余额为零的新钱包
func (s *RedisStorage) Create() (*model.Wallet, error) {
	id := uuid.New().String()
	now := time.Now()

	key := walletKey(id)
	// 使用 Hash 存储钱包数据
	err := s.client.HSet(s.ctx, key, map[string]interface{}{
		"id":      id,
		"balance": 0,
		"created": now.Unix(),
	}).Err()

	if err != nil {
		return nil, fmt.Errorf("创建钱包失败: %w", err)
	}

	return &model.Wallet{
		ID:      id,
		Balance: 0,
		Created: now,
	}, nil
}

// Get 根据ID获取钱包
func (s *RedisStorage) Get(id string) (*model.Wallet, error) {
	key := walletKey(id)
	data, err := s.client.HGetAll(s.ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("查询钱包失败: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("钱包不存在: %s", id)
	}

	// 解析数据
	createdUnix, _ := strconv.ParseInt(data["created"], 10, 64)
	balance, _ := strconv.ParseInt(data["balance"], 10, 64)

	return &model.Wallet{
		ID:      data["id"],
		Balance: balance,
		Created: time.Unix(createdUnix, 0),
	}, nil
}

// Update 更新现有钱包
func (s *RedisStorage) Update(wallet *model.Wallet) error {
	key := walletKey(wallet.ID)

	// 先检查钱包是否存在
	exists, err := s.client.Exists(s.ctx, key).Result()
	if err != nil {
		return fmt.Errorf("检查钱包存在失败: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("钱包不存在: %s", wallet.ID)
	}

	// 更新余额
	err = s.client.HSet(s.ctx, key, "balance", wallet.Balance).Err()
	if err != nil {
		return fmt.Errorf("更新钱包失败: %w", err)
	}

	return nil
}

// GetAll 返回所有钱包
func (s *RedisStorage) GetAll() []*model.Wallet {
	// 使用 SCAN 遍历所有 wallet:* 的 key
	var cursor uint64
	wallets := []*model.Wallet{}

	for {
		keys, nextCursor, err := s.client.Scan(s.ctx, cursor, "wallet:*", 100).Result()
		if err != nil {
			return wallets
		}

		for _, key := range keys {
			data, err := s.client.HGetAll(s.ctx, key).Result()
			if err != nil {
				continue
			}
			if len(data) == 0 {
				continue
			}

			createdUnix, _ := strconv.ParseInt(data["created"], 10, 64)
			balance, _ := strconv.ParseInt(data["balance"], 10, 64)

			wallets = append(wallets, &model.Wallet{
				ID:      data["id"],
				Balance: balance,
				Created: time.Unix(createdUnix, 0),
			})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return wallets
}

// TransferAtomically 在两个钱包间执行原子转账操作
// 使用 Redis MULTI/EXEC 事务确保原子性
func (s *RedisStorage) TransferAtomically(fromID, toID string, amount int64) error {
	fromKey := walletKey(fromID)
	toKey := walletKey(toID)

	// 使用 WATCH + MULTI/EXEC 实现乐观锁事务
	err := s.client.Watch(s.ctx, func(tx *redis.Tx) error {
		// 检查源钱包是否存在
		fromData, err := tx.HGetAll(s.ctx, fromKey).Result()
		if err != nil {
			return err
		}
		if len(fromData) == 0 {
			return fmt.Errorf("一个或两个钱包不存在")
		}

		// 检查目标钱包是否存在
		toData, err := tx.HGetAll(s.ctx, toKey).Result()
		if err != nil {
			return err
		}
		if len(toData) == 0 {
			return fmt.Errorf("一个或两个钱包不存在")
		}

		// 检查余额是否足够
		fromBalance, _ := strconv.ParseInt(fromData["balance"], 10, 64)
		if fromBalance < amount {
			return fmt.Errorf("余额不足")
		}

		// 执行转账
		_, err = tx.TxPipelined(s.ctx, func(pipe redis.Pipeliner) error {
			pipe.HIncrBy(s.ctx, fromKey, "balance", -amount)
			pipe.HIncrBy(s.ctx, toKey, "balance", amount)
			return nil
		})

		return err
	}, fromKey, toKey)

	if err != nil {
		return err
	}

	return nil
}

// Close 关闭 Redis 连接
func (s *RedisStorage) Close() error {
	return s.client.Close()
}
