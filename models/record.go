package models

type Record struct {
	ID         int
	GameID int
	Round      int
	Side       string
	Credits    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}


