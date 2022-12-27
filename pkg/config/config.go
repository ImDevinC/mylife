package config

import (
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

type AppConfig struct {
	LifesheetFile string
	TelegramToken string
	ChatID        int64
	CSVPath       string
	Mongo         MongoConfig
}

type MongoConfig struct {
	Username string
	Password string
	URL      string
	Port     string
	Database string
}

func New() (*AppConfig, error) {
	rawChatID := os.Getenv("TELEGRAM_CHAT_ID")
	chatID, err := strconv.ParseInt(rawChatID, 10, 64)
	if err != nil {
		return nil, err
	}
	mongoCfg := MongoConfig{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
		URL:      os.Getenv("MONGO_URL"),
		Port:     os.Getenv("MONGO_PORT"),
		Database: os.Getenv("MONGO_DB"),
	}
	appConfig := AppConfig{
		LifesheetFile: "lifesheet.json",
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		ChatID:        chatID,
		CSVPath:       "database.csv",
		Mongo:         mongoCfg,
	}

	return &appConfig, nil
}
