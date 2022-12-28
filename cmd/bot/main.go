package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/imdevinc/mylife/pkg/bot"
	"github.com/imdevinc/mylife/pkg/config"
	"github.com/imdevinc/mylife/pkg/database"
	"github.com/imdevinc/mylife/pkg/lifesheet"
	"github.com/imdevinc/mylife/pkg/scheduler"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	sheet, err := lifesheet.LoadFromFile(cfg.LifesheetFile)
	if err != nil {
		log.Fatal(err)
	}
	telegram, err := bot.New(&bot.BotConfig{Token: cfg.TelegramToken, ChatID: cfg.ChatID})
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.NewMongoDB(context.TODO(), database.MongoDatabaseOptions{
		Username: cfg.Mongo.Username,
		Password: cfg.Mongo.Password,
		URL:      cfg.Mongo.URL,
		Port:     cfg.Mongo.Port,
		Database: cfg.Mongo.Database,
	})
	if err != nil {
		log.Fatal(err)
	}
	scheduleCfg := scheduler.SchedulerConfig{
		Bot:   telegram,
		Sheet: sheet,
	}
	go func() {
		msgChan := telegram.Start()
		for msg := range msgChan {
			log.WithField("response", msg.Text).Debug("got response")
			if msg.IsCommand {
				log.WithField("response", msg.Text).Debug("got command")
				message := strings.ToLower(msg.Text)
				go scheduler.StartQuestions(&scheduleCfg, message)
				continue
			}
			if err := db.SaveAnswer(context.TODO(), msg); err != nil {
				log.WithError(err).Error("failed to save results")
				telegram.SendMessage(fmt.Sprintf("failed to save answer to database. %s", err))
			}
			if !msg.Skipped {
				telegram.SendMessage("👍")
			}

		}
	}()
	// time.Sleep(2 * time.Second) // TODO: Fix this
	if err := scheduler.Start(&scheduleCfg); err != nil {
		log.Fatal(err)
	}
}
