package game

import (
	"context"
	"time"

	bot "github.com/MooooonStar/bot-api-go-client"
	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/models"
	"github.com/smallnest/rpcx/log"
)

func (j *Judge) sendMessagsLoop(ctx context.Context) {
	for {
		if err := j.send(ctx); err != nil {
			log.Error("SEND MESSAGES ERROR: ", err)
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (j *Judge) send(ctx context.Context) error {
	const limit = 100
	messages, err := models.FindMessages(ctx, 0, limit)
	if err != nil {
		return err
	}

	var msgs []*bot.MessageRequest
	var ids []int64
	filter := make(map[string]bool)
	for _, m := range messages {
		if filter[m.RecipientID] {
			continue
		}
		msgs = append(msgs, &bot.MessageRequest{
			ConversationId:   m.ConversationID,
			RecipientId:      m.RecipientID,
			MessageId:        m.MessageID,
			Category:         m.Category,
			Data:             m.Data,
			RepresentativeId: m.RepresentativeID,
			QuoteMessageId:   m.QuoteMessageID,
		})
		filter[m.RecipientID] = true
		ids = append(ids, m.ID)
	}

	if err := bot.PostMessages(ctx, msgs, config.UserID, config.SessionID, config.PrivateKey); err != nil {
		return err
	}

	return models.DeleteMessages(ctx, ids...)
}
