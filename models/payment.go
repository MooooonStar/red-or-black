package models

import (
	"time"
	"github.com/jinzhu/gorm"
)


const (
	PaymentAmount = "0.00001"
	PaymentStatusPending = "PENDING"
	PaymentStatusPaid = "PAID"
)

type Payment struct {
	ID        int
	UserID    string
	AssetID   string
	Amount    string
	TraceID   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

func InsertPayment(ctx context.Context, traceID string) (*Payment, error) {
	payment := Payment{
		UserID: userID,
		AssetID: BTC,
		TraceID: traceID,
		Status: PaymentStatus,
		Amount: PaymentAmount,
	}
	err := session.Database(ctx).Where("user_id = ?", userID).FirstOrCreate(&payment).Error
	if gorm.IsErrRecordNotFound(err) {
		return nil, nil
	}
	return *payment,err
}

func FindPayment(ctx context.Context, userID string) (*Payment, error) {
	var payment Payment
	err := session.Database(ctx).Where("user_id = ?", userID).First(&payment).Error
	if gorm.IsErrRecordNotFound(err) {
		return nil, nil
	}
	return *payment,err
}
