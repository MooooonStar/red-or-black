package game

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client"
	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/models"
	"github.com/shopspring/decimal"
)

const (
	PollInterval                    = 100 * time.Millisecond
	CheckpointMixinNetworkSnapshots = "checkpoint-mixin-network-snapshots"
)

func PollSnapshots(ctx context.Context) {
	const limit = 500
	checkpoint, err := models.ReadPropertyAsTime(ctx, CheckpointMixinNetworkSnapshots)
	if err != nil {
		log.Println("ReadPropertyAsTime CheckpointMixinNetworkSnapshots", err)
		panic(err)
	}
	if checkpoint.IsZero() {
		checkpoint = time.Now().UTC()
	}
	filter := make(map[string]bool)
	for {
		snapshots, err := requestMixinNetwork(ctx, checkpoint, limit)
		if err != nil {
			log.Println("PollMixinNetwork ERROR", err)
			time.Sleep(PollInterval)
			continue
		}
		for _, s := range snapshots {
			if filter[s.SnapshotID] {
				continue
			}
			ensureProcessSnapshot(ctx, s)
			checkpoint = s.CreatedAt
			filter[s.SnapshotID] = true
		}
		if len(snapshots) < limit {
			time.Sleep(PollInterval)
		}
		err = models.WriteTimeProperty(ctx, CheckpointMixinNetworkSnapshots, checkpoint)
		if err != nil {
			log.Println("WriteTimeProperty CheckpointMixinNetworkSnapshots", err)
		}
	}
}

func requestMixinNetwork(ctx context.Context, checkpoint time.Time, limit int) ([]*models.Snapshot, error) {
	uri := fmt.Sprintf("/network/snapshots?offset=%s&order=ASC&limit=%d", checkpoint.Format(time.RFC3339Nano), limit)
	token, err := bot.SignAuthenticationToken(config.UserID, config.SessionID, config.PrivateKey, "GET", uri, "")
	if err != nil {
		return nil, err
	}
	body, err := bot.Request(ctx, "GET", uri, nil, token)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Data  []*models.Snapshot `json:"data"`
		Error *models.MixinError `json:"error"`
	}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Data, nil
}

func ensureProcessSnapshot(ctx context.Context, s *models.Snapshot) {
	for {
		err := processSnapshot(ctx, s)
		if err == nil {
			break
		}
		log.Println("ensureProcessSnapshot", err)
		time.Sleep(100 * time.Millisecond)
	}
}

func processSnapshot(ctx context.Context, s *models.Snapshot) error {
	if s.UserID == "" || s.TraceID == "" {
		return nil
	}
	s.AssetID = s.Asset.AssetID
	s.Symbol = s.Asset.Symbol
	if err := models.CreateSnapshot(ctx, s); err != nil {
		return err
	}
	amount, _ := decimal.NewFromString(s.Amount)
	if !amount.IsPositive() {
		return nil
	}
	payment, err := models.FindPaymentByTrace(ctx, s.TraceID)
	if err != nil {
		return err
	}
	if payment == nil {
		//return refundSnapshot(ctx, s, "Invalid payment")
		return nil
	}
	am, _ := decimal.NewFromString(payment.Amount)
	if payment.AssetID != s.AssetID || !IsZero(am.Sub(amount)) || payment.UserID != s.OpponentID {
		log.Println(payment.AssetID == s.AssetID, IsZero(am.Sub(amount)), payment.UserID != s.OpponentID)
		//return refundSnapshot(ctx, s, "Invalid payment")
		return nil
	}
	return models.UpdatePaymentPaid(ctx, payment.ID)
}

func IsZero(a decimal.Decimal) bool {
	min := decimal.NewFromFloat(0.00000001)
	return a.Abs().LessThan(min)
}

func refundSnapshot(ctx context.Context, s *models.Snapshot, memo string) error {
	t := models.Transfer{
		TransferID: bot.UniqueConversationId(s.SnapshotID, "REFUND"),
		CreatedAt:  time.Now(),
		AssetID:    s.Asset.AssetID,
		Amount:     s.Amount,
		OpponentID: s.OpponentID,
		Memo:       memo,
	}
	return models.CreateTransfer(ctx, &t)
}
