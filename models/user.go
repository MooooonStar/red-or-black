package models

import (
	"time"
)

const (
	BTC = "XXX"

	UserStatusUnpaid  = "UNPAID"
	UserStatusPending = "PENDING"
	UserStatusActive  = "ACTIVE"
)

type User struct {
	UserID       string
	Conversation string
	PaidAsset    string
	PaidAmount   string
	EarnedAmount string
	Status       string
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

func FindUser(ctx context.Context, userID) (*User, error){
	var user User
	err := session.Database(ctx).Where("user_id = ?", userID).First(&user).Error
	if gorm.IsErrRecordNotFound(err) {
		return nil,nil
	}
	return err
}

func InsertUser(ctx context.Context, user *User) error{
	return session.Database(ctx).FirstOrCreate(user).Error
}
