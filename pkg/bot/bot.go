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
}

type MessageResponse struct {
	Text        string
	QuestionKey string
	Skipped     bool
	IsCommand   bool
}

type AskedQuestion struct {
	Text    string
	Key     string
	Replies map[string]string
	Type    string
	Buttons map[string]string
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

func (t *Telegram) SendMessage(message AskedQuestion) error {
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
	}
	if _, err := t.bot.Send(msg); err != nil {
		return err
	}
	return nil
}

func (t *Telegram) Get() *tgbotapi.BotAPI {
	return t.bot
}

func (t *Telegram) ProcessMessage(chatID int64, messageID int, text string, location *tgbotapi.Location, ch chan MessageResponse) {
	if chatID != t.cfg.ChatID {
		msg := tgbotapi.NewMessage(chatID, "This is not the bot you're looking for")
		t.bot.Send(msg)
		return
	}

	if strings.HasPrefix(text, "/") && strings.ToLower(text) != "/skip" {
		ch <- MessageResponse{
			Text:      strings.TrimPrefix(text, "/"),
			IsCommand: true,
		}
		return
	}

	if t.LastQuestion.Key == "" {
		msg := tgbotapi.NewMessage(t.cfg.ChatID, "I didn't ask a question")
		t.bot.Send(msg)
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
		Skipped:     strings.ToLower(text) == "/skip",
	}

	if location != nil {
		resp.QuestionKey = "locationLat"
		resp.Text = fmt.Sprintf("%f", location.Latitude)
		ch <- resp
		resp.QuestionKey = "locationLong"
		resp.Text = fmt.Sprintf("%f", location.Longitude)
		ch <- resp
		return
	}

	ch <- resp
	t.LastQuestion = AskedQuestion{}
	t.WaitingForResponse = false
}
