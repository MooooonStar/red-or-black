package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
	"github.com/shopspring/decimal"
)

const (
	MaxRound         = 5
	GameStatusActive = "ACTIVE"
	GameStatusDone   = "DONE"
)

type Game struct {
	ID        int64 `gorm:"PRIMARY_KEY"`
	Round     int
	Prize     string
	Used      bool
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func InsertGame(ctx context.Context, g *Game) error {
	return session.Database(ctx).FirstOrCreate(g).Error
}

func FindGame(ctx context.Context, id int) (*Game, error) {
	var g Game
	err := session.Database(ctx).Where("id = ?", id).First(&g).Error
	return &g, err
}

func FindPrizePool(ctx context.Context, deadline int64) (string, error) {
	var games []*Game
	err := session.Database(ctx).Where("status = ? AND used = ? AND id <= ?", GameStatusDone, false, deadline).Find(&games).Error
	if err != nil {
		return "0", err
	}
	var sum decimal.Decimal
	for _, game := range games {
		amount, _ := decimal.NewFromString(game.Prize)
		sum = sum.Add(amount)
	}
	return sum.String(), nil
}
