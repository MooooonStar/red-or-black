package main

import (
	"context"
	"fmt"
	"log"

	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/game"
	"github.com/MooooonStar/red-or-black/session"
	"github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

func main() {
	ctx := initContext()
	j := &game.Judge{}
	j.Run(ctx)

	err := make(chan error)
	<-err

	// id := "edd18cfa-3c13-38b5-8074-c10cc71731da"
	// token, _ := bot.SignAuthenticationToken(config.UserID, config.SessionID, config.PrivateKey, "GET", "/conversations/"+id, "")
	// conversation, _ := bot.ConversationShow(ctx, id, token)
	// bt, _ := json.Marshal(conversation)
	// log.Println(string(bt))
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
