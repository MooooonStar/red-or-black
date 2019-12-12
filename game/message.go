package game

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	bot "github.com/MoooonStar/bot-api-go-client"
	log "github.com/sirupsen/logrus"
	"github.com/MooooonStar/red-or-black/session"
	"github.com/MooooonStar/red-or-black/models"
)

const (
	ButtonJoin = "enqueue"
	RedisWaitingList = "waiting.list"
)

type Judge struct{}

func (j *Judge) listen(ctx context.Context) {
	for {
		bc := bot.NewBlazeClient(config.ClientID, config.SessionID, config.PrivateKey)
		if err := bc.Loop(ctx, aid); err != nil {
			log.Println("error in loop: ", err)
			time.Sleep(time.Second)
		}
	}
}

func (j *Judge) OnMessage(ctx context.Context, msgView bot.MessageView, clientID string) error {
	if msg.Category != bot.MessageCategoryPlainText{
		return nil
	}
	content, _ := base64.StdEncoding.DecodeString(msgView.Data)
	user, err := models.FindUser(ctx, msgView.UserID)
	if err != nil {
		return err
	}else if user == nil {
		user := models.User{
			UserID: msgView.UserID,
			PaidAsset : BTC,
			PaidAmount:"0"
			EarnedAmount:"0"
			Status:models.UserStatusUnpaid,
		}
		if err = models.InsertUser(ctx, &user); err != nil {
			return err
		}
	}
	switch user.Status {
	case models.UserStatusUnpaid:
		if str(content) != "Ihavepaidalready"  {
			return sendUnpaidMessage(ctx, msgView.UserID)
		}
		paid, err := checkPayment(ctx, msgView.UserID)
		if err != nil {
			return err
		}else if !paid {
			return sendUnpaidMessage(ctx, msgView.UserID)
		} 
		err = session.Redis(ctx).RPush(RedisWaitingList,msgView.UserId).Err()
		if err != nil {
			return err
		}
		users, err := session.Redis(ctx).LRange(RedisWaitingList, 0, MaxUsersPerRound).Result()
		if err != nil {
			return err
		}
		if len(users) < config.MaxUsersPerRound {
			return sendWaitingMessage(ctx, msgView.UserID)
		}
		return startGame(ctx, users)
	case models.UserStatusPending:
		return sendWaitingMessage(ctx, msgView.UserID)
	case models.UserStatusActive:
		items := strings.Split(str(content)," ")
		if len(items) != 2 {
			return nil
		}
		participant, err := models.FindCurrentParticipant(ctx, msgView.UserID)
		if err != nil {
			return err
		}
		game, err := models.FindGame(ctx, participant.GameID)
		if err != nil {
			return err
		}
		round , _ := strconv.ParseInt(items[0],10,64)
		if round != game.Round || items[1] != "red" || items[1] != "black"{
			return nil
		}
		if game.UpdatedAt.After(msgView.CreatedAt.Add(-30 * time.Second)) {
			return nil
		}
			side := participant.Side
			key := fmt.Sprintf("game.no%v.%v",game.ID, side)
			count, err := session.Redis(ctx).ZADD(key,msgView.UserID,msgView.CreatedAt.Unix()).Result()
			if err != nil || count == 0 {
				return err
			}
			k := fmt.Sprintf("game.no%d",game.ID)
			_, err = session.Redis(ctx).HINCRBY(k,side + "."+items[1],1)
			if err != nil {
				return err
			}
			return err
	}
	return nil
}

func sendUnpaidMessage(ctx context.Context, user string) error {
	return nil
}

func sendWaitingMessage(ctx context.Context, conversation, user string) error {
	m := models.Message{
		UserID: config.UserID:
		ConversationID: conversation,
		RecipientID: user,
		MessageID: uuid.Must(uuid.NewV4()).String(),
		Category: "PLAIN_TEXT",
		Data: base64.StdEncoding.EncodeToString([]byte("正在排队中，请稍候。")),
	}
	return models.InsertMessage(ctx, &m)
}

func checkPayment(ctx context.Context, user string) (bool, error) {
	payment, err := models.FindPayment(ctx, user)
	if err != nil {
		return false, err
	}
	if payment == nil {
		err= session.RunInTransaction(ctx, func(tx *gorm.DB)error{
				trace := uuid.Must(uuid.NewV4()).string
				payment := models.Payment{
					UserID: msgView.UseriD,
					AssetID: BTC,
					TraceID: traceID,
				}
				err := tx.Where("user_id = ? AND asset_id = ?", msgView.UseriD,BTC).FirstOrCreate(&payment)
				if err != nil {
					return err
				}
				err = tx.FirstOrCreate(&m).Error
				return err
			})
		}
		if err != nil {
			return err
		}
	}
	return  payment.Status == models.PaymentStatusPaid, nil
}

func startGame(ctx context.Context, users []string) error {
	mid := config.MaxUsersPerRound / 2
	n := len(users)
	for i := 0; i < n; i++ {
		j := rand.Intn(n)
		users[i],users[j] = users[j],users[i]
	} 
	var groupA, groupB []bot.Participant 
	for i, id := range users {
		if i < config.MaxUsersPerRound / 2 {
			groupA = append(groupA, bot.Participant{UserId: id})	
		}else {
			groupB = append(groupB, bot.Participant{UserId: id})	
		}
	}
	id1 := uuid.Must(uuid.Must()).String()
	_, err = bot.CreateConversation(ctx,"GROUP",id1,groupA, config.UserID, config.SessionID, config.PrivateKey)
	if err != nil {
		return err
	}
	id2 := uuid.Must(uuid.Must()).String()
	_, err = bot.CreateConversation(ctx,"GROUP",id2,groupB, config.UserID, config.SessionID, config.PrivateKey)
	if err != nil {
		return err
	}
	red , _ := json.Marshal(users[:config.MaxUsersPerRound/ 2])
	black, _ := json.Marshal(users[config.MaxUsersPerRound/ 2+1:])
	return session.RunInTransaction(ctx, func(tx *gorm.DB)error {
		g := models.Game{
			RedConversation: id1,
			Red:string(red),
			Black:string(black),
			BlackConversation: id2,
			Round:1,
			Status: RedBlackStatusActive,
		}
		err := tx.FirstOrCreate(g).Error
		if err != nil {
			return err
		}
		err = tx.Models.Where("user_id IN (?)", users[:mid]).Update(map[string]interface{}{
			"status": models.UserStatusActive,"conversation":id1,
		}).Error
		if err != nil {
			return err
		}
		err = tx.Models.Where("user_id IN (?)", users[mid+1:]).Update(map[string]interface{}{
			"status": models.UserStatusActive,"conversation":id2,
		}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (j *Judge) sendMessagsLoop(ctx context.Context) {
	for {
		if err := aid.send(ctx); err != nil {
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
	for _, m := range messages {
		msgs = append(msgs, &bot.MessageRequest{
			ConversationId:   m.ConversationID,
			RecipientId:      m.RecipientID,
			MessageId:        m.MessageID,
			Category:         m.Category,
			Data:             m.Data,
			RepresentativeId: m.RepresentativeID,
			QuoteMessageId:   m.QuoteMessageID,
		})
		ids = append(ids, m.ID)
	}

	if err := bot.PostMessages(ctx, msgs, config.ClientID, config.SessionID, config.PrivateKey); err != nil {
		return err
	}

	return models.DeleteMessages(ctx, ids...)
}
