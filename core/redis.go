package core

import (
	"context"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"gvb-server/global"
	"strconv"
	"time"
)

func ConnectRedis() *redis.Client {
	return ConnectRedisDB(0)
}

func ConnectRedisDB(db int) *redis.Client {
	redisConf := global.Config.Redis
	reb := redis.NewClient(&redis.Options{
		Addr:     redisConf.Ip + ":" + strconv.Itoa(redisConf.Port),
		Password: redisConf.Password,
		DB:       db,
		PoolSize: redisConf.PoolSize,
	})
	fmt.Println(redisConf.Ip, redisConf.Password, redisConf.PoolSize, db)
	_, concel := context.WithTimeout(context.Background(), 5*time.Second)
	defer concel()
	_, err := reb.Ping().Result()
	if err != nil {
		logrus.Errorf("连接redis失败，检查redis配置 %s", err)
		return nil
	}
	return reb
}
