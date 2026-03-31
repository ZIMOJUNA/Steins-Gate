package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"server/config"
	"time"

	"github.com/gofiber/storage/redis/v3"
	"github.com/google/uuid"
)

// 全局 Redis 客户端（初始化一次）
var redisStore *redis.Storage

// InitRedis 初始化Redis
func InitRedis() {
	redisStore = redis.New(redis.Config{
		Host:     config.Conf.Redis.Host,
		Port:     config.Conf.Redis.Port,
		Password: config.Conf.Redis.Password,
		Database: config.Conf.Redis.DB,
	})
}

// GenerateToken 生成唯一Token并存入Redis
// key格式：user_token:{token}
// value：绑定的用户ID/业务ID
// expire：过期时间，如 time.Hour * 24 * 7
func GenerateToken(bizID string, expire time.Duration) (string, error) {
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
	key := fmt.Sprintf("user_token:%s", token)
	val, err := redisStore.GetWithContext(context.Background(), key)
	if err != nil {
		return "", err
	}
	return string(val), nil
}

// DeleteToken 删除token（登出用）
func DeleteToken(token string) error {
	key := fmt.Sprintf("user_token:%s", token)
	return redisStore.DeleteWithContext(context.Background(), key)
}
