package game

import (
	"context"
	"log"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	number "github.com/MixinNetwork/go-number"
	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/models"
)

func PollTransfers(ctx context.Context) {
	limit := 10
	for {
		transfers, err := models.ListPendingTransfers(ctx, limit)
		if err != nil {
			log.Println("ListPendingTransfers", err)
			time.Sleep(PollInterval)
			continue
		}
		for _, t := range transfers {
			if err := processTransfer(ctx, t); err != nil {
				log.Printf("Process transfer error: %v, %v", err, t.AssetID)
				continue
			}
			if err := models.ExpireTransfer(ctx, t.TransferID); err != nil {
				log.Printf("Expire transfer error: %v, %v", err, t.TransferID)
				continue
			}
		}
		if len(transfers) < limit {
			time.Sleep(PollInterval)
		}
	}
}

func processTransfer(ctx context.Context, t *models.Transfer) error {
	return bot.CreateTransfer(ctx, &bot.TransferInput{
		AssetId:     t.AssetID,
		RecipientId: t.OpponentID,
		Amount:      number.FromString(t.Amount),
		TraceId:     t.TransferID,
		Memo:        t.Memo,
	}, config.UserID, config.SessionID, config.PrivateKey, config.Pin, config.PinToken)
}
