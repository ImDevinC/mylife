package bot

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	log "github.com/sirupsen/logrus"
)

type BotConfig struct {
	Token   string
	Debug   bool
	Timeout int
	ChatID  int64
}

type Telegram struct {
	bot                *tgbotapi.BotAPI
	cfg                *BotConfig
	WaitingForResponse bool
	LastQuestion       AskedQuestion
	skipRemaining      bool
}

type MessageResponse struct {
	Text        string
	QuestionKey string
	IsCommand   bool
	Acknowledge bool
	Question    string
	Type        string
}

type AskedQuestion struct {
	Question string
	Text     string
	Key      string
	Replies  map[string]string
	Type     string
	Buttons  map[string]string
}

type MessageChannel chan MessageResponse

const defaultTimeout int = 30

func New(config *BotConfig) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot. %v", err)
	}
	bot.Debug = config.Debug
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	return &Telegram{bot: bot, cfg: config}, nil
}

func (t *Telegram) Start() MessageChannel {
	log.Info("starting bot")
	ch := make(chan MessageResponse)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = t.cfg.Timeout
	updates := t.bot.GetUpdatesChan(u)
	go func() {
		for update := range updates {
			var chatID int64
			var messageID int
			var text string
			var location *tgbotapi.Location
			if update.CallbackQuery != nil {
				chatID = update.CallbackQuery.Message.Chat.ID
				messageID = update.CallbackQuery.Message.MessageID
				location = update.CallbackQuery.Message.Location
				text = update.CallbackData()
				callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
				t.bot.Send(callback) // ignore error for now
			} else if update.Message != nil {
				chatID = update.Message.Chat.ID
				messageID = update.Message.MessageID
				location = update.Message.Location
				text = update.Message.Text
			} else {
				return
			}
			t.ProcessMessage(chatID, messageID, text, location, ch)
		}
	}()
	return ch
}

func (t *Telegram) SendQuestion(message AskedQuestion) error {
	log.Debug("sending message")
	t.WaitingForResponse = true
	t.LastQuestion = message
	msg := tgbotapi.NewMessage(t.cfg.ChatID, message.Text)
	if len(message.Buttons) > 0 {
		rows := [][]tgbotapi.InlineKeyboardButton{}
		for k, v := range message.Buttons {
			row := tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(v, k))
			rows = append(rows, row)
		}
		keyb := tgbotapi.NewInlineKeyboardMarkup(rows...)
		msg.ReplyMarkup = keyb
	}
	if message.Type == "boolean" {
		keyb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("Yes", "true")),
			tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("No", "false")),
		)
		msg.ReplyMarkup = keyb
	} else if message.Type == "location" {
		btn := tgbotapi.NewKeyboardButtonLocation("Provide your location")
		keyb := tgbotapi.NewOneTimeReplyKeyboard([]tgbotapi.KeyboardButton{btn})
		msg.ReplyMarkup = keyb
	}
	_, err := t.bot.Send(msg)
	if err != nil {
		return err
	}
	return nil
}

func (t *Telegram) SendMessage(message string) error {
	msg := tgbotapi.NewMessage(t.cfg.ChatID, message)
	msg.ReplyMarkup = map[string]bool{
		"hide_keyboard": true,
	}
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (t *Telegram) ProcessMessage(chatID int64, messageID int, text string, location *tgbotapi.Location, ch chan MessageResponse) {
	if chatID != t.cfg.ChatID {
		msg := tgbotapi.NewMessage(chatID, "This is not the bot you're looking for")
		t.bot.Send(msg)
		return
	}

	if strings.HasPrefix(text, "/") && strings.ToLower(text) != "/skip" && strings.ToLower(text) != "/skip_all" {
		ch <- MessageResponse{
			Text:        strings.TrimPrefix(text, "/"),
			IsCommand:   true,
			Acknowledge: false,
		}
		return
	}

	if t.LastQuestion.Key == "" {
		msg := tgbotapi.NewMessage(t.cfg.ChatID, "I didn't ask a question")
		if _, err := t.bot.Send(msg); err != nil {
			log.WithError(err).Error("failed to send message")
		}
		return
	}

	if val, ok := t.LastQuestion.Replies[text]; ok {
		msg := tgbotapi.NewMessage(t.cfg.ChatID, val)
		msg.ReplyToMessageID = messageID
		if _, err := t.bot.Send(msg); err != nil {
			log.WithError(err).Error("failed to send reply")
		}
	}

	resp := MessageResponse{
		Text:        text,
		QuestionKey: t.LastQuestion.Key,
		Question:    t.LastQuestion.Question,
		Type:        t.LastQuestion.Type,
	}

	if location != nil {
		resp.QuestionKey = "locationLat"
		resp.Text = fmt.Sprintf("%f", location.Latitude)
		ch <- resp
		resp.QuestionKey = "locationLong"
		resp.Text = fmt.Sprintf("%f", location.Longitude)
		resp.Acknowledge = true
		ch <- resp
		t.LastQuestion = AskedQuestion{}
		t.WaitingForResponse = false
		return
	}

	resp.Acknowledge = true

	if strings.ToLower(text) == "/skip_all" {
		t.skipRemaining = true
		t.LastQuestion = AskedQuestion{}
		t.WaitingForResponse = false
	} else if strings.ToLower(text) == "/skip" {
		t.LastQuestion = AskedQuestion{}
		t.WaitingForResponse = false
	} else {
		ch <- resp
	}

	// if strings.ToLower(text) != "/skip" && strings.ToLower(text) != "/skip_all" {
	// 	ch <- resp

	// }
	// if strings.ToLower(text) == "/skip_all" {
	// 	t.skipRemaining = true
	// }
	// t.LastQuestion = AskedQuestion{}
	// t.WaitingForResponse = false
}

func (t *Telegram) NextQuestion() {
	t.WaitingForResponse = false
}

func (t *Telegram) ResetQuestions() {
	t.LastQuestion = AskedQuestion{}
	t.WaitingForResponse = false
	t.skipRemaining = false
}

func (t *Telegram) ShouldSkipRemaining() bool {
	return t.skipRemaining
}
func (t *Telegram) SendImageURL(u string) error {
	image := tgbotapi.NewInputMediaPhoto(tgbotapi.FileURL(u))
	mediaGroup := tgbotapi.NewMediaGroup(t.cfg.ChatID, []interface{}{image})
	_, err := t.bot.Request(mediaGroup)
	if err != nil {
		return fmt.Errorf("failed to send image. %v", err)
	}
	return nil
}
