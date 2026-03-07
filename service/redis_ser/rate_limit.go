package redis_ser

import (
	"time"

	"gvb-server/global"
)

func IncrWithTTL(key string, ttl time.Duration) (int64, error) {
	count, err := global.Redis.Incr(key).Result()
	if err != nil {
		return 0, err
	}
	if count == 1 {
		if err = global.Redis.Expire(key, ttl).Err(); err != nil {
			return count, err
		}
	}
	return count, nil
}

func GetInt64(key string) int64 {
	value, err := global.Redis.Get(key).Int64()
	if err != nil {
		return 0
	}
	return value
}

func DelKeys(keys ...string) {
	if len(keys) == 0 {
		return
	}
	_ = global.Redis.Del(keys...).Err()
}
