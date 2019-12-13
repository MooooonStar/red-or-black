package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

const (
	SideOne = "ONE"
	SideTwo = "TWO"
)

type Player struct {
	GameID       int    `gorm:"PRIMARY_KEY"`
	UserID       string `gorm:"PRIMARY_KEY"`
	Side         string
	Conversation string
	CreatedAt    time.Time
	DeletedAt    time.Time
}

func FindCurrentPlayer(ctx context.Context, user string) (*Player, error) {
	var player Player
	err := session.Database(ctx).Where("user_id = ?", user).First(&player).Error
	return &player, err
}

func FindGamePlayers(ctx context.Context, id int64) ([]*Player, error) {
	var players []*Player
	err := session.Database(ctx).Where("game_id = ?", id).Find(&players).Error
	return players, err
}

func FindGameSidePlayers(ctx context.Context, id int64, side string) ([]*Player, error) {
	var players []*Player
	err := session.Database(ctx).Where("game_id = ? AND side = ?", id, side).Find(&players).Error
	return players, err
}
