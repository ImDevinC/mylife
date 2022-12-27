package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/imdevinc/mylife/pkg/bot"
	"github.com/imdevinc/mylife/pkg/lifesheet"

	log "github.com/sirupsen/logrus"
)

type SchedulerConfig struct {
	Bot   *bot.Telegram
	Sheet *lifesheet.Lifesheet
}

type Scheduler struct {
	Bot *bot.Telegram
}

func Start(cfg *SchedulerConfig) error {
	s := Scheduler{Bot: cfg.Bot}
	sched := gocron.NewScheduler(time.Local)
	for k, c := range cfg.Sheet.Categories {
		switch c.Schedule {
		case "daily":
			if k == "awake" {
				sched.Every(1).Day().At("08:00:00").Do(s.AskQuestions, c.Questions)
			} else if k == "asleep" {
				sched.Every(1).Day().At("22:00:00").Do(s.AskQuestions, c.Questions)
			}
		case "weekly":
			sched.Every(1).Sunday().At("08:00:00").Do(s.AskQuestions, c.Questions)
		case "eightTimesADay":
			sched.Every(3).Hours().At("08:00:00").Do(s.AskQuestions, c.Questions)
		default:
			return fmt.Errorf("invalid schedule. %s", c.Schedule)
		}
	}
	log.Info("scheduler started")
	sched.StartBlocking()
	return nil
}

func StartQuestions(cfg *SchedulerConfig, key string) {
	s := Scheduler{Bot: cfg.Bot}
	key = strings.ToLower(key)
	var questionKey string
	log.Info(key)
	if strings.HasPrefix(key, "track ") {
		questionKey = strings.TrimPrefix(key, "track ")
	}
	for k, c := range cfg.Sheet.Categories {
		if questionKey != "" {
			for _, q := range c.Questions {
				if strings.ToLower(q.Key) == questionKey {
					s.AskQuestions([]lifesheet.Question{q})
					return
				}
			}
		} else {
			if strings.ToLower(k) != key {
				continue
			}
			if questionKey == "" {
				s.AskQuestions(c.Questions)
				return
			}
		}
	}
}

func (s *Scheduler) AskQuestions(questions []lifesheet.Question) {
	var wg sync.WaitGroup
	for _, q := range questions {
		msg := q.Text
		// if len(q.Buttons) > 0 {
		// 	options := []string{}
		// 	for k, v := range q.Buttons {
		// 		options = append(options, fmt.Sprintf("[%s] %s", k, v))
		// 	}
		// 	msg = fmt.Sprintf("%s\n%s", q.Text, strings.Join(options, "\n"))
		// }
		s.Bot.SendMessage(bot.AskedQuestion{Text: msg, Key: q.Key, Replies: q.Replies, Type: q.Type, Buttons: q.Buttons})
		time.Sleep(1 * time.Second)
		if q.Type == "header" {
			s.Bot.WaitingForResponse = false
		}
		wg.Add(1)
		go func() {
			for {
				if !s.Bot.WaitingForResponse {
					wg.Done()
					break
				}
			}
		}()
		wg.Wait()
	}
}
