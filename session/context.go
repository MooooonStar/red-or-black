package session

import (
	"context"

	"github.com/jinzhu/gorm"
)

func RunInTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	tx := Database(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}
