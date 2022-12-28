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

// SchedulerConfig holds the construction information
// for a new Scheduler
type SchedulerConfig struct {
	Bot   *bot.Telegram
	Sheet *lifesheet.Lifesheet
}

// Scheduler handles the calls to Scheduler
type Scheduler struct {
	Bot *bot.Telegram
}

// Start initiates a new scheduler and schedules questions
// to be asked at a specific time
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
		case "fiveTimesADay":
			sched.Every(1).Day().At("09:00").Do(s.AskQuestions, c.Questions)
			sched.Every(1).Day().At("12:00").Do(s.AskQuestions, c.Questions)
			sched.Every(1).Day().At("15:00").Do(s.AskQuestions, c.Questions)
			sched.Every(1).Day().At("18:00").Do(s.AskQuestions, c.Questions)
			sched.Every(1).Day().At("21:00").Do(s.AskQuestions, c.Questions)
		default:
			return fmt.Errorf("invalid schedule. %s", c.Schedule)
		}
	}
	log.Info("scheduler started")
	sched.StartBlocking()
	return nil
}

// StartQuestions looks at the key being provided to determine
// which set of questions to ask
func StartQuestions(cfg *SchedulerConfig, key string) {
	s := Scheduler{Bot: cfg.Bot}
	key = strings.ToLower(key)
	var questionKey string
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

// AskQuestions goes through each question and sends it through the bot.
// We use multiple waits in this function to make sure the question
// gets answered or we bail in time
func (s *Scheduler) AskQuestions(questions []lifesheet.Question) {
	var wg sync.WaitGroup
	for _, q := range questions {
		ignoreQuestions := false
		// If a question is waiting a response, don't send the next one
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
		msg := q.Text
		s.Bot.SendQuestion(bot.AskedQuestion{Text: msg, Key: q.Key, Replies: q.Replies, Type: q.Type, Buttons: q.Buttons})
		if q.Type == "header" {
			s.Bot.WaitingForResponse = false
			continue
		}
		// Give the user a limited time to answer
		ts := time.Now()
		wg.Add(1)
		go func() {
			for {
				if !s.Bot.WaitingForResponse {
					wg.Done()
					break
				}
				now := time.Now()
				if now.Sub(ts) > (30 * time.Minute) {
					log.Info("timeout")
					s.Bot.SendMessage("Maybe you're busy, no worry. We'll skip the check-in for now")
					ignoreQuestions = true
					wg.Done()
					break
				}
			}
		}()
		wg.Wait()
		// If the user didn't answer the question in time, assume they are busy
		if ignoreQuestions {
			s.Bot.ResetQuestions()
			break
		}
	}
}
