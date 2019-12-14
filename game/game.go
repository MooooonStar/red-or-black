package game

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	bot "github.com/MooooonStar/bot-api-go-client"
	"github.com/MooooonStar/red-or-black/config"
	"github.com/MooooonStar/red-or-black/models"
	"github.com/MooooonStar/red-or-black/session"
	"github.com/go-redis/redis/v7"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
)

const (
	ButtonJoin       = "enqueue"
	RedisWaitingList = "waitinglist"
	CmdAlreadyPaid   = "ihavepaidalready"
	GroupOneName     = "A组"
	GroupTwoName     = "B组"
)

type Judge struct{}

func (j *Judge) Run(ctx context.Context) {
	go j.Listen(ctx)
	go j.sendMessagsLoop(ctx)
	go PollSnapshots(ctx)
	go PollTransfers(ctx)
}

func (j *Judge) Listen(ctx context.Context) {
	for {
		bc := bot.NewBlazeClient(config.UserID, config.SessionID, config.PrivateKey)
		if err := bc.Loop(ctx, j); err != nil {
			log.Println("error in loop: ", err)
			time.Sleep(time.Second)
		}
	}
}

func (j *Judge) OnMessage(ctx context.Context, msgView bot.MessageView, clientID string) error {
	content, _ := base64.StdEncoding.DecodeString(msgView.Data)
	prefix := fmt.Sprintf("@%d ", config.IdentityNumber)
	raw := string(content)
	if strings.HasPrefix(raw, prefix) {
		raw = raw[len(prefix):]
	}
	log.Info(msgView.MessageId, ", ", string(content))
	if msgView.Category != bot.MessageCategoryPlainText {
		return nil
	}
	user, err := models.FindUser(ctx, msgView.UserId)
	if err != nil {
		return err
	} else if user == nil {
		user = &models.User{
			UserID:       msgView.UserId,
			PaidAsset:    models.BTC,
			PaidAmount:   "0",
			EarnedAmount: "0",
			Status:       models.UserStatusUnpaid,
		}
		if err = models.InsertUser(ctx, user); err != nil {
			return err
		}
	}
	switch user.Status {
	case models.UserStatusUnpaid:
		payment, err := checkPayment(ctx, msgView.UserId)
		if err != nil {
			return err
		} else if !payment.Paid {
			return sendUnpaidMessage(ctx, msgView.ConversationId, msgView.UserId, payment.TraceID)
		}
		if raw != CmdAlreadyPaid {
			return nil
		}
		_, err = models.UpdateUserStatus(ctx, msgView.UserId, models.UserStatusWaiting)
		if err != nil {
			return err
		}
		err = session.Redis(ctx).ZAdd(RedisWaitingList, &redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: msgView.UserId,
		}).Err()
		if err != nil {
			return err
		}
		users, err := session.Redis(ctx).ZRange(RedisWaitingList, 0, config.UsersPerRound).Result()
		if err != nil {
			return err
		}
		if len(users) < config.UsersPerRound {
			return sendWaitingMessage(ctx, msgView.ConversationId, msgView.UserId)
		}
		return startGame(ctx, users)
	case models.UserStatusWaiting:
		return sendWaitingMessage(ctx, msgView.ConversationId, msgView.UserId)
	case models.UserStatusActive:
		items := strings.Split(raw, " ")
		if len(items) != 2 {
			return nil
		}
		player, err := models.FindCurrentPlayer(ctx, msgView.UserId)
		if err != nil || player == nil {
			return err
		}
		game, err := models.FindGame(ctx, player.GameID)
		if err != nil {
			return err
		}
		round, _ := strconv.Atoi(items[0])
		if round != game.Round || (items[1] != "red" && items[1] != "black") {
			return nil
		}
		key := fmt.Sprintf("voted.no%v.round%v", game.ID, game.Round)
		n, err := session.Redis(ctx).SAdd(key, msgView.UserId).Result()
		if err != nil || n == 0 {
			return err
		}
		key1 := fmt.Sprintf("votes.no%v.round%v", game.ID, game.Round)
		field := fmt.Sprintf("%v.%v", player.Side, items[1])
		_, err = session.Redis(ctx).HIncrBy(key1, field, 1).Result()
		if err != nil {
			return err
		}
		count, err := session.Redis(ctx).SCard(key).Result()
		if err != nil || count != config.UsersPerRound {
			return err
		}
		_, err = countVotes(ctx, game.ID, round)
		if err != nil {
			return err
		}
		if err = sendVotesResults(ctx, game.ID); err != nil {
			return err
		}
		if round < config.MaxRound {
			return nextGameStage(ctx, game.ID, round)
		}
		return settleGame(ctx, game.ID)
	}
	return nil
}

func settleGame(ctx context.Context, id int64) error {
	records, err := models.FindGameRecords(ctx, id)
	if err != nil {
		return err
	}
	if len(records) != config.MaxRound {
		return nil
	}
	var a, b int
	for _, record := range records {
		a, b = a+record.OneCredit, b+record.TwoCredit
	}
	var (
		prize string
		side  string
		count int
	)
	amountPerRound, _ := strconv.ParseFloat(config.AmountPerRound, 64)
	if a <= 0 && b <= 0 {
		amount := (1.0 - config.PunishRate) * config.UsersPerRound * amountPerRound
		prize = fmt.Sprintf("%.8f", amount)
		side = "none"
	} else {
		prize = "0"
		if a == b {
			side = models.SideOne + models.SideTwo
			count = config.UsersPerRound
		} else {
			side = models.SideOne
			if a < b {
				side = models.SideTwo
			}
			count = config.UsersPerRound / 2
		}
	}
	prizePool, err := models.FindPrizePool(ctx, id)
	if err != nil {
		return err
	}
	pool, _ := decimal.NewFromString(prizePool)
	total := pool.Add(decimal.NewFromFloat(float64(config.UsersPerRound) * amountPerRound))
	var average decimal.Decimal
	if count > 0 {
		average = total.Div(decimal.NewFromFloat(float64(count)))
	}
	players, err := models.FindGamePlayers(ctx, id)
	if err != nil {
		return err
	}
	err = session.RunInTransaction(ctx, func(tx *gorm.DB) error {
		err := tx.Model(models.Game{}).Where("id = ?", id).Update(map[string]interface{}{
			"status": models.GameStatusDone,
			"prize":  prize,
			"used":   count > 0,
		}).Error
		if err != nil {
			return err
		}
		var users []string
		for _, player := range players {
			users = append(users, player.UserID)
			var content string
			if strings.Contains(side, player.Side) {
				t := models.Transfer{
					TransferID: bot.UniqueConversationId(player.UserID, "PRIZE FROM GAME"+fmt.Sprint(id)),
					AssetID:    models.BTC,
					Amount:     average.String(),
					OpponentID: player.UserID,
					Memo:       "Game prize",
				}
				err = tx.Create(&t).Error
				if err != nil {
					return err
				}
				content = "恭喜您获得游戏胜利👏👏👏"
			} else {
				content = "很遗憾，您输了😪😪😪"
			}
			m := models.Message{
				UserID:         config.UserID,
				ConversationID: player.Conversation,
				RecipientID:    player.UserID,
				MessageID:      uuid.Must(uuid.NewV4()).String(),
				Category:       "PLAIN_TEXT",
				Data:           base64.StdEncoding.EncodeToString([]byte(content)),
			}
			if err = tx.Create(&m).Error; err != nil {
				return err
			}
		}
		err = tx.Where("user_id IN (?)", users).Debug().Delete(models.Payment{}).Error
		if err != nil {
			return err
		}
		err = tx.Where("game_id = ?", id).Debug().Delete(models.Player{}).Error
		if err != nil {
			return err
		}
		err = tx.Model(models.User{}).Where("user_id IN (?)", users).Update("status", models.UserStatusUnpaid).Error
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func sendUnpaidMessage(ctx context.Context, conversation, user, trace string) error {
	url := fmt.Sprintf("mixin://pay?recipient=%v&asset=%v&amount=%v&trace=%v&memo=%v", config.UserID, models.BTC, config.AmountPerRound, trace, "Pay to join.")
	btns, _ := json.Marshal([]map[string]interface{}{
		map[string]interface{}{
			"label":  "您需要支付以加入排队",
			"color":  "#FF0000",
			"action": url,
		},
		map[string]interface{}{
			"label":  "我已支付",
			"color":  "#000000",
			"action": fmt.Sprintf("input:%v", CmdAlreadyPaid),
		},
	})
	m := models.Message{
		UserID:         config.UserID,
		ConversationID: conversation,
		RecipientID:    user,
		MessageID:      uuid.Must(uuid.NewV4()).String(),
		Category:       "APP_BUTTON_GROUP",
		Data:           base64.StdEncoding.EncodeToString(btns),
	}
	return models.InsertMessage(ctx, &m)
}

func sendWaitingMessage(ctx context.Context, conversation, user string) error {
	n, err := session.Redis(ctx).ZCard(RedisWaitingList).Result()
	if err != nil || n == 0 {
		return err
	}
	content := fmt.Sprintf("排队中，当前排队人数 %v，请耐心等待🤝🤝🤝", n)
	m := models.Message{
		UserID:         config.UserID,
		ConversationID: conversation,
		RecipientID:    user,
		MessageID:      uuid.Must(uuid.NewV4()).String(),
		Category:       "PLAIN_TEXT",
		Data:           base64.StdEncoding.EncodeToString([]byte(content)),
	}
	return models.InsertMessage(ctx, &m)
}

func checkPayment(ctx context.Context, user string) (*models.Payment, error) {
	payment, err := models.FindPayment(ctx, user)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		payment = &models.Payment{
			UserID:  user,
			AssetID: models.BTC,
			TraceID: uuid.Must(uuid.NewV4()).String(),
			Amount:  config.AmountPerRound,
		}
		if err := models.InsertPayment(ctx, payment); err != nil {
			return nil, err
		}
	}
	return payment, nil
}

func startGame(ctx context.Context, users []string) (err error) {
	mid := config.UsersPerRound / 2
	n := len(users)
	for i := 0; i < n; i++ {
		j := rand.Intn(n)
		users[i], users[j] = users[j], users[i]
	}

	var id1, id2 string
	var started bool
	err = session.RunInTransaction(ctx, func(tx *gorm.DB) error {
		g := &models.Game{
			Prize:     "0",
			Round:     1,
			Status:    models.GameStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := tx.Create(g).Error
		if err != nil {
			return err
		}
		id1 = bot.UniqueConversationId(config.UserID, fmt.Sprintf("GAME.ONE.%v", g.ID))
		err = sendNotifyMessages(tx, id1, users[:mid], 1)
		if err != nil {
			return err
		}
		id2 = bot.UniqueConversationId(config.UserID, fmt.Sprintf("GAME.TWO.%v", g.ID))
		err = sendNotifyMessages(tx, id2, users[mid:], 1)
		if err != nil {
			return err
		}
		err = tx.Model(models.User{}).Where("user_id IN (?)", users).Update(map[string]interface{}{
			"status": models.UserStatusActive,
		}).Error
		if err != nil {
			return err
		}
		for i, user := range users {
			var side, conversation string
			if i < config.UsersPerRound/2 {
				side, conversation = models.SideOne, id1
			} else {
				side, conversation = models.SideTwo, id2
			}
			player := &models.Player{
				GameID:       g.ID,
				UserID:       user,
				Side:         side,
				Conversation: conversation,
			}
			if err = tx.FirstOrCreate(player).Error; err != nil {
				return err
			}
		}
		started = true
		return nil
	})
	if err != nil || !started {
		return
	}

	var groupA, groupB []bot.Participant
	for i, id := range users {
		if i < config.UsersPerRound/2 {
			groupA = append(groupA, bot.Participant{UserId: id})
		} else {
			groupB = append(groupB, bot.Participant{UserId: id})
		}
	}
	_, err = bot.CreateConversation(ctx, GroupOneName, "GROUP", id1, groupA, config.UserID, config.SessionID, config.PrivateKey)
	if err != nil {
		return
	}
	_, err = bot.CreateConversation(ctx, GroupTwoName, "GROUP", id2, groupB, config.UserID, config.SessionID, config.PrivateKey)
	if err != nil {
		return
	}
	err = session.Redis(ctx).ZRemRangeByRank(RedisWaitingList, 0, config.UsersPerRound).Err()
	return
}

func nextGameStage(ctx context.Context, id int64, round int) error {
	players, err := models.FindGamePlayers(ctx, id)
	if err != nil {
		return err
	}
	var a, b string
	var one, two []string
	for _, player := range players {
		if player.Side == models.SideOne {
			one = append(one, player.UserID)
			if a == "" {
				a = player.Conversation
			}
		} else if player.Side == models.SideTwo {
			two = append(two, player.UserID)
			if b == "" {
				b = player.Conversation
			}
		}
	}
	return session.RunInTransaction(ctx, func(tx *gorm.DB) error {
		err = sendNotifyMessages(tx, a, one, round+1)
		if err != nil {
			return err
		}
		err = sendNotifyMessages(tx, b, two, round+1)
		if err != nil {
			return err
		}
		err = tx.Model(&models.Game{}).Where("id = ?", id).Update(map[string]interface{}{"round": round + 1}).Error
		return err
	})
}

func sendNotifyMessages(tx *gorm.DB, conversation string, users []string, round int) error {
	content := fmt.Sprintf("第 %d 轮游戏开始，请于 60s 内完成投票。", round)
	btns, _ := json.Marshal([]map[string]interface{}{
		map[string]interface{}{
			"label":  " 红 ",
			"color":  "#FF0000",
			"action": fmt.Sprintf("input:@%v %v red", config.IdentityNumber, round),
		},
		map[string]interface{}{
			"label":  " 黑 ",
			"color":  "#000000",
			"action": fmt.Sprintf("input:@%v %v black", config.IdentityNumber, round),
		},
	})
	for _, user := range users {
		m := models.Message{
			UserID:         config.UserID,
			ConversationID: conversation,
			RecipientID:    user,
		}
		m.Category = "PLAIN_TEXT"
		m.MessageID = uuid.Must(uuid.NewV4()).String()
		m.Data = base64.StdEncoding.EncodeToString([]byte(content))
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
		m.ID = 0
		m.Category = "APP_BUTTON_GROUP"
		m.MessageID = uuid.Must(uuid.NewV4()).String()
		m.Data = base64.StdEncoding.EncodeToString(btns)
		if err := tx.Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}

func sendVotesResults(ctx context.Context, id int64) error {
	records, err := models.FindGameRecords(ctx, id)
	if err != nil {
		return err
	}

	template := `投票结果为(红/黑):
	------------------------------------
	A | %4v | %4v | %4v | %4v | %4v |%4v 
	------------------------------------
	B | %4v | %4v | %4v | %4v | %4v |%4v   
	------------------------------------
	`
	var list [12]interface{}
	for i := 0; i < len(list); i++ {
		list[i] = "-/-"
	}
	var a, b int
	for i, record := range records {
		list[i] = fmt.Sprintf("%v/%v", record.OneRed, record.OneBlack)
		list[6+i] = fmt.Sprintf("%v/%v", record.TwoRed, record.TwoBlack)
		a, b = a+record.OneCredit, b+record.TwoCredit
	}
	list[5], list[11] = fmt.Sprint(a), fmt.Sprint(b)
	content := fmt.Sprintf(template, list[:]...)

	players, err := models.FindGamePlayers(ctx, id)
	if err != nil {
		return err
	}
	for _, player := range players {
		m := &models.Message{
			UserID:         config.UserID,
			ConversationID: player.Conversation,
			RecipientID:    player.UserID,
			MessageID:      uuid.Must(uuid.NewV4()).String(),
			Category:       "PLAIN_TEXT",
			Data:           base64.StdEncoding.EncodeToString([]byte(content)),
		}
		if err := models.InsertMessage(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
