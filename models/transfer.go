package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

type Transfer struct {
	TransferID string `gorm:"PRIMARY_KEY"`
	AssetID    string
	Amount     string
	OpponentID string
	Memo       string
	CreatedAt  time.Time
}

func (Transfer) TableName() string {
	return "transfers"
}

func CreateTransfer(ctx context.Context, t *Transfer) error {
	return session.Database(ctx).Debug().Create(t).Error
}

func ListPendingTransfers(ctx context.Context, limit int) ([]*Transfer, error) {
	transfers := make([]*Transfer, 0)
	err := session.Database(ctx).Order("created_at ASC").Limit(limit).Find(&transfers).Error
	return transfers, err
}

func ExpireTransfer(ctx context.Context, id string) error {
	return session.Database(ctx).Where("transfer_id = ?", id).Delete(&Transfer{}).Error
}
