package models

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

type Snapshot struct {
	SnapshotID string `json:"snapshot_id" gorm:"PRIMARY_KEY"`
	Amount     string `json:"amount"`
	Asset      struct {
		AssetID string `json:"asset_id"`
		Symbol  string `json:"symbol"`
	} `gorm:"-"         json:"asset"`
	TraceID    string    `json:"trace_id"`
	UserID     string    `json:"user_id"`
	OpponentID string    `json:"opponent_id"`
	Data       string    `json:"data"`
	AssetID    string    `json:"asset_id"`
	Symbol     string    `json:"symbol"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Snapshot) TableName() string {
	return "snapshots"
}

func CreateSnapshot(ctx context.Context, s *Snapshot) error {
	return session.Database(ctx).FirstOrCreate(s).Error
}

type MixinError struct {
	Status      int    `json:"status"`
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (err MixinError) Error() string {
	bt, _ := json.Marshal(err)
	return string(bt)
}
