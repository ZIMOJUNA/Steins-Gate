package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/Future-Game-Laboratory/Steins-Gate/config"
	"github.com/gofiber/storage/redis/v3"
	"github.com/google/uuid"
)

// 全局 Redis 客户端（初始化一次）
var redisStore *redis.Storage

var ErrNotInitialized = errors.New("redis not initialized")

// Init 初始化 Redis。
func Init() error {
	if redisStore != nil {
		return nil
	}

	redisStore = redis.New(redis.Config{
		Host:     config.Conf.Redis.Host,
		Port:     config.Conf.Redis.Port,
		Password: config.Conf.Redis.Password,
		Database: config.Conf.Redis.DB,
	})
	// 连通性测试：Set + Get 验证
	testKey := "redis_connect_test"
	testVal := "ok"
	err := redisStore.SetWithContext(context.Background(), testKey, []byte(testVal), 10*time.Second)
	if err != nil {
		redisStore = nil
		return fmt.Errorf("Redis 连接失败（Set 测试）：%w", err)
	}

	val, err := redisStore.GetWithContext(context.Background(), testKey)
	if err != nil {
		redisStore = nil
		return fmt.Errorf("Redis 连接失败（Get 测试）：%w", err)
	}
	if string(val) != testVal {
		redisStore = nil
		return errors.New("Redis 连接失败：数据不一致")
	}

	// 清理测试键
	_ = redisStore.DeleteWithContext(context.Background(), testKey)

	log.Println("✅ Redis 全局初始化完成！")
	return nil
}

// GenerateToken 生成唯一Token并存入Redis
// key格式：user_token:{token}
// value：绑定的用户ID/业务ID
// expire：过期时间，如 time.Hour * 24 * 7
func GenerateToken(bizID string, expire time.Duration) (string, error) {
	if redisStore == nil {
		return "", ErrNotInitialized
	}

	// 1. 生成唯一token：UUID + 16字节随机串
	token, err := genUniqueToken()
	if err != nil {
		return "", err
	}

	// 2. Redis key
	key := fmt.Sprintf("user_token:%s", token)

	// 3. 存入Redis
	err = redisStore.SetWithContext(
		context.Background(),
		key,
		[]byte(bizID),
		expire,
	)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetStorage 获取全局 Redis 存储。
func GetStorage() *redis.Storage {
	return redisStore
}

// genUniqueToken 生成安全唯一token
func genUniqueToken() (string, error) {
	// UUID保证唯一
	uid := uuid.NewString()

	// 16字节随机增强安全性
	rnd := make([]byte, 16)
	if _, err := rand.Read(rnd); err != nil {
		return "", err
	}
	rndStr := base64.URLEncoding.EncodeToString(rnd)

	// 拼接最终token
	return uid + rndStr, nil
}

// GetTokenBizID 根据token获取绑定的业务ID
func GetTokenBizID(token string) (string, error) {
	if redisStore == nil {
		return "", ErrNotInitialized
	}

	key := fmt.Sprintf("user_token:%s", token)
	val, err := redisStore.GetWithContext(context.Background(), key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// DeleteToken 删除token（登出用）
func DeleteToken(token string) error {
	if redisStore == nil {
		return ErrNotInitialized
	}

	key := fmt.Sprintf("user_token:%s", token)
	return redisStore.DeleteWithContext(context.Background(), key)
}

func Set(ctx context.Context, key string, value []byte, expire time.Duration) error {
	if redisStore == nil {
		return ErrNotInitialized
	}
	return redisStore.SetWithContext(ctx, key, value, expire)
}

func Get(ctx context.Context, key string) ([]byte, error) {
	if redisStore == nil {
		return nil, ErrNotInitialized
	}
	return redisStore.GetWithContext(ctx, key)
}

func Del(ctx context.Context, key string) error {
	if redisStore == nil {
		return ErrNotInitialized
	}
	return redisStore.DeleteWithContext(ctx, key)
}

func IncrWithExpire(ctx context.Context, key string, expire time.Duration) (int64, error) {
	if redisStore == nil {
		return 0, ErrNotInitialized
	}

	count, err := redisStore.Conn().Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 && expire > 0 {
		if err := redisStore.Conn().Expire(ctx, key, expire).Err(); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func GetInt(ctx context.Context, key string) (int64, error) {
	data, err := Get(ctx, key)
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(string(data), 10, 64)
}
