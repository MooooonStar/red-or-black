package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
	"github.com/jinzhu/gorm"
)

const (
	UserStatusUnpaid  = "UNPAID"
	UserStatusWaiting = "WAITING"
	UserStatusActive  = "ACTIVE"
)

type User struct {
	UserID       string `gorm:"PRIMARY_KEY"`
	PaidAsset    string
	PaidAmount   string
	EarnedAmount string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func FindUser(ctx context.Context, userID string) (*User, error) {
	var user User
	err := session.Database(ctx).Where("user_id = ?", userID).First(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, nil
	}
	return &user, err
}

func InsertUser(ctx context.Context, user *User) error {
	return session.Database(ctx).FirstOrCreate(user).Error
}

func UpdateUserStatus(ctx context.Context, userID, status string) (int64, error) {
	db := session.Database(ctx).Model(User{}).Where("user_id = ?", userID).Update(map[string]interface{}{
		"status": status,
	})
	return db.RowsAffected, db.Error
}
