package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

type Record struct {
	ID        int64 `gorm:"PRIMARY_KEY"`
	GameID    int64
	Round     int
	OneRed    int
	OneBlack  int
	TwoRed    int
	TwoBlack  int
	CreatedAt time.Time
}

func InsertRecord(ctx context.Context, r *Record) error {
	return session.Database(ctx).Where("game_id = ? ADN round = ?", r.GameID, r.Round).FirstOrCreate(r).Error
}

func FindGameRecords(ctx context.Context, gameID int64) ([]*Record, error) {
	var records []*Record
	err := session.Database(ctx).Where("game_id = ?", gameID).Order("round DESC").Find(&records).Error
	return records, err
}
