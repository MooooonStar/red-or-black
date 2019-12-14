package game

import (
	"context"
	"fmt"
	"testing"

	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/session"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"github.com/smallnest/rpcx/log"
	"github.com/stretchr/testify/assert"
	_ "github.com/go-sql-driver/mysql"
)

func TestSettleGame(t *testing.T) {
	ctx := initContext()
	err := settleGame(ctx, 1)
	assert.Nil(t, err)
}

func initContext() context.Context {
	ctx := context.Background()
	path := fmt.Sprintf("%s:%s@%s(%s)/%s?parseTime=True&charset=utf8mb4",
		config.DatabaseUserName,
		config.DatabasePassword,
		"tcp",
		config.DatabaseHost,
		config.DatabaseName,
	)
	db, err := gorm.Open("mysql", path)
	if err != nil {
		log.Fatal(err)
	}
	err = db.DB().PingContext(ctx)
	if err != nil {
		log.Panic(err)
	}
	ctx = session.WithDatabase(ctx, db)

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		DB:       config.RedisDB,
		Password: config.RedisPassword,
	})
	err = redisClient.Ping().Err()
	if err != nil {
		log.Panic(err)
	}
	ctx = session.WithRedis(ctx, redisClient)
	return ctx
}
