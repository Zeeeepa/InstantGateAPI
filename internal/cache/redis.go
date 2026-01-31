package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/proyaai/instantgate/internal/config"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(cfg *config.RedisConfig) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	ttl := cfg.CacheTTL
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &Cache{
		client: client,
		ttl:    ttl,
	}, nil
}

func (c *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *Cache) SetWithTTL(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return c.client.Del(ctx, keys...).Err()
}

func (c *Cache) Invalidate(ctx context.Context, pattern string) error {
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

func (c *Cache) Exists(ctx context.Context, key string) bool {
	n, err := c.client.Exists(ctx, key).Result()
	return err == nil && n > 0
}

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Flush(ctx context.Context) error {
	return c.client.FlushAll(ctx).Err()
}

func GenerateQueryKey(table string, params interface{}) string {
	h := sha256.New()
	json.NewEncoder(h).Encode(params)
	hash := hex.EncodeToString(h.Sum(nil))[:16]

	return fmt.Sprintf("query:%s:%s", table, hash)
}

func GenerateTableKey(table string) string {
	return fmt.Sprintf("table:%s", table)
}

func GenerateRecordKey(table, id string) string {
	return fmt.Sprintf("record:%s:%s", table, id)
}

func (c *Cache) InvalidateTable(ctx context.Context, table string) error {
	pattern := fmt.Sprintf("*%s*", table)
	return c.Invalidate(ctx, pattern)
}

var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)
