package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
	"github.com/jinzhu/gorm"
)

type Payment struct {
	ID        int64 `gorm:"PRIMARY_KEY"`
	UserID    string
	AssetID   string
	Amount    string
	TraceID   string
	Paid      bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (pay Payment) TableName() string {
	return "payments"
}

func InsertPayment(ctx context.Context, payment *Payment) error {
	return session.Database(ctx).Create(payment).Error
}

func UpdatePaymentPaid(ctx context.Context, id int64) error {
	return session.Database(ctx).Model(&Payment{}).Where("id = ?", id).Update(map[string]interface{}{"paid": true}).Error
}

func FindPayment(ctx context.Context, userID string) (*Payment, error) {
	var payment Payment
	err := session.Database(ctx).Where("user_id = ?", userID).First(&payment).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &payment, err
}

func FindPaymentByTrace(ctx context.Context, trace string) (*Payment, error) {
	var payment Payment
	err := session.Database(ctx).Where("trace_id = ?", trace).First(&payment).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &payment, err
}
