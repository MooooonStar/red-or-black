package models

import (
	"time"
)

const (
	MaxRound             = 5
	GameStatusActive = "ACTIVE"
	GameStatusDone   = "DONE"
)

type Game struct {
	GameID int
	Round      int
	Status     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func InsertGame(ctx context.Context, g *Game) error{
	return session.Database(ctx).FirstOrCreate(g).Error
}

func FindGame(ctx context.Context, id int) (*Game,error){
	var g Game
	err := session.Database(ctx).Where("game_id = ?", id).First(&g).Error
	return &g, err
}



