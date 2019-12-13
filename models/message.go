package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
)

type Message struct {
	ID               int64 `gorm:"PRIMARY_KEY"`
	UserID           string
	ConversationID   string
	RecipientID      string
	MessageID        string
	Category         string
	Data             string
	RepresentativeID string
	QuoteMessageID   string
	CreatedAt        time.Time
}

func (m Message) TableName() string {
	return "messages"
}

func InsertMessage(ctx context.Context, m *Message) error {
	return session.Database(ctx).Where("message_id = ?", m.MessageID).FirstOrCreate(m).Error
}

func FindMessages(ctx context.Context, offset, limit int) ([]*Message, error) {
	var messages []*Message
	err := session.Database(ctx).Where("id > ?", offset).Limit(limit).Find(&messages).Error
	return messages, err
}

func DeleteMessages(ctx context.Context, ids ...int64) error {
	return session.Database(ctx).Model(&Message{}).Where("id IN (?)", ids).Delete(&Message{}).Error
}
