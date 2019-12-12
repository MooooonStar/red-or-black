package models

import (
	"time"
)

const (
	SideTiger           = "TIGER"
	SideLion            = "LION"
)

type Participant struct {
	GameID int
	UserID string 
	Side string 
	Conversation string
	CreatedAt time.Time
	DeletedAt time.Time
}

func FindCurrentParticipant(ctx context.Context, user string)(*Participant, error){
	var p Participant
	err := session.Database(ctx).Where("user_id = ?", user).First(&p).Error
	return &p, err
}