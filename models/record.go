package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

const (
	RecordStatusPending = "PENDING"
	RecordStatusDone    = "DONE"
)

type Record struct {
	GameID    int64 `gorm:"PRIMARY_KEY"`
	Round     int   `gorm:"PRIMARY_KEY"`
	OneRed    int
	OneBlack  int
	OneCredit int
	TwoRed    int
	TwoBlack  int
	TwoCredit int
	CreatedAt time.Time
	UpdatedAt time.Time
}

func InsertRecord(ctx context.Context, r *Record) error {
	return session.Database(ctx).Where("game_id = ? AND round = ?", r.GameID, r.Round).FirstOrCreate(r).Error
}

func FindGameRecord(ctx context.Context, game, round int64) (*Record, error) {
	var record Record
	err := session.Database(ctx).Where("game_id = ? AND round = ?", game, round).First(&record).Error
	return &record, err
}

func FindGameRecords(ctx context.Context, gameID int64) ([]*Record, error) {
	var records []*Record
	err := session.Database(ctx).Where("game_id = ?", gameID).Order("round ASC").Find(&records).Error
	return records, err
}
