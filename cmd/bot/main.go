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

const chartAPI string = "https://chart.googleapis.com/chart?cht=lc&chd=t:%s&chs=800x350&chl=%s&chtt=%s&chf=bg,s,e0e0e0&chco=000000,0000FF&chma=30,30,30,30&chds=%d,%d"

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
			if msg.IsCommand && strings.HasPrefix(msg.Text, "graph ") {
				key := strings.TrimPrefix(msg.Text, "graph ")
				vals, err := db.GetValues(context.TODO(), key)
				if err != nil {
					log.WithError(err).Error("failed to save get values from database")
					telegram.SendMessage(fmt.Sprintf("failed to get graph info from database. %s", err))
					continue
				}
				url := fmt.Sprintf(chartAPI, strings.Join(vals.Values, ","), strings.Join(vals.Times, "%7C"), msg.QuestionKey, vals.Minimum, vals.Maximum)
				if err := telegram.SendImageURL(url); err != nil {
					log.WithError(err).Error("failed to send graph")
					telegram.SendMessage(fmt.Sprintf("failed to send graph. %s", err))
				}
				continue
			}
			if msg.IsCommand {
				message := strings.ToLower(msg.Text)
				go scheduler.ProcessCommand(&scheduleCfg, message)
				continue
			}
			if err := db.SaveAnswer(context.TODO(), database.AnswerResponse{
				Key:     msg.QuestionKey,
				Answer:  msg.Text,
				Skipped: msg.Skipped,
			}); err != nil {
				log.WithError(err).Error("failed to save results")
				telegram.SendMessage(fmt.Sprintf("failed to save answer to database. %s", err))
			}
			if !msg.Skipped {
				telegram.SendMessage("👍")
			}

		}
	}()
	if err := scheduler.Start(&scheduleCfg); err != nil {
		log.Fatal(err)
	}
}
