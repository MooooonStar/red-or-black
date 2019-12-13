package session

import (
	"context"

	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
)

type contextKey int

const (
	databaseContextKey contextKey = 0
	redisContextKey    contextKey = 1
)

func WithDatabase(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, databaseContextKey, db)
}

func WithRedis(ctx context.Context, client *redis.Client) context.Context {
	return context.WithValue(ctx, redisContextKey, client)
}

func Redis(ctx context.Context) *redis.Client {
	return ctx.Value(redisContextKey).(*redis.Client)
}

func Database(ctx context.Context) *gorm.DB {
	return ctx.Value(databaseContextKey).(*gorm.DB)
}
