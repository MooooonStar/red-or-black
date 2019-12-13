package game

import (
	"context"
	"fmt"
	"strconv"

	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/models"
	"github.com/MooooonStar/red-or-black/session"
)

type Vote struct {
	UserID string `json:"user_id"`
	Choice string `json:"choice"`
}

func countVotes(ctx context.Context, id int64, round int, follow bool) error {
	key := fmt.Sprintf("game.no%d", id)
	votes, err := session.Redis(ctx).HGetAll(key).Result()
	if err != nil {
		return err
	}
	a, _ := strconv.ParseInt(votes["one.red"], 10, 64)
	b, _ := strconv.ParseInt(votes["one.black"], 10, 64)
	c, _ := strconv.ParseInt(votes["two.red"], 10, 64)
	d, _ := strconv.ParseInt(votes["two.black"], 10, 64)
	if a+b != config.UsersPerRound/2 || c+d != config.UsersPerRound/2 {
		return nil
	} else if follow {
		return nil
	}
	record := &models.Record{
		GameID:   id,
		Round:    round,
		OneRed:   int(a),
		OneBlack: int(b),
		TwoRed:   int(c),
		TwoBlack: int(d),
	}
	err = models.InsertRecord(ctx, record)
	if err != nil {
		return err
	}
	return nil
}
