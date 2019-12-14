package models

import (
	"context"
	"time"

	"github.com/MooooonStar/red-or-black/session"
	"github.com/jinzhu/gorm"
)

type Property struct {
	Key       string `gorm:"PRIMARY_KEY"`
	Value     string
	UpdatedAt time.Time
}

func (Property) TableName() string {
	return "properties"
}

func ReadProperty(ctx context.Context, key string) (string, error) {
	var p Property
	err := session.Database(ctx).Where("`key` = ?", key).First(&p).Error
	if gorm.IsRecordNotFoundError(err) {
		return "", nil
	}
	return p.Value, err
}

func WriteProperty(ctx context.Context, key, value string) error {
	return session.Database(ctx).Model(&Property{}).Where("`key` = ?", key).Update(&Property{Value: value}).Error
}

func ReadPropertyAsTime(ctx context.Context, key string) (time.Time, error) {
	var offset time.Time
	timestamp, err := ReadProperty(ctx, key)
	if err != nil {
		return offset, err
	}
	if timestamp != "" {
		return time.Parse(time.RFC3339Nano, timestamp)
	}
	return offset, nil
}

func WriteTimeProperty(ctx context.Context, key string, value time.Time) error {
	return WriteProperty(ctx, key, value.UTC().Format(time.RFC3339Nano))
}
