package game

import (
	"context"
	"fmt"
	"strconv"

	"github.com/MooooonStar/red-or-black/models"
	"github.com/MooooonStar/red-or-black/session"
)

func countVotes(ctx context.Context, id int64, round int) (*models.Record, error) {
	key := fmt.Sprintf("votes.no%v.round%v", id, round)
	votes, err := session.Redis(ctx).HGetAll(key).Result()
	if err != nil {
		return nil, err
	}
	a, _ := strconv.Atoi(votes["one.red"])
	b, _ := strconv.Atoi(votes["one.black"])
	c, _ := strconv.Atoi(votes["two.red"])
	d, _ := strconv.Atoi(votes["two.black"])
	optionA, optionB := OptionRed, OptionRed
	if a < b {
		optionA = OptionBlack
	}
	if c < d {
		optionB = OptionBlack
	}
	x, y := Credit(optionA, optionB, round)
	record := &models.Record{
		GameID:    id,
		Round:     round,
		OneRed:    a,
		OneBlack:  b,
		OneCredit: x,
		TwoRed:    c,
		TwoBlack:  d,
		TwoCredit: y,
	}
	return record, models.InsertRecord(ctx, record)
}

const (
	OptionRed   = "red"
	OptionBlack = "black"
)

func Credit(a, b string, round int) (x int, y int) {
	if a == OptionRed {
		if b == OptionRed {
			x, y = -3, -3
		} else if b == OptionBlack {
			x, y = 5, -5
		}
	} else if a == OptionBlack {
		if b == OptionRed {
			x, y = -5, 5
		} else if b == OptionBlack {
			x, y = 3, 3
		}
	}
	if round == 2 {
		x, y = 2*x, 2*y
	} else if round == 4 {
		x, y = 3*x, 3*y
	}
	return
}
