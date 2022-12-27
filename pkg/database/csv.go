package database

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/imdevinc/mylife/pkg/bot"
)

type CSVDatabase struct {
	path string
}

func NewCSV(path string) Database {
	return &CSVDatabase{path: path}
}

func (d *CSVDatabase) SaveAnswer(_ context.Context, msg bot.MessageResponse) error {
	f, err := os.OpenFile(d.path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open CSV file. %v", err)
	}
	defer f.Close()
	w := csv.NewWriter(f)
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	err = w.Write([]string{msg.QuestionKey, msg.Text, strconv.FormatBool(msg.Skipped), ts})
	if err != nil {
		return fmt.Errorf("failed to save data. %v", err)
	}
	w.Flush()
	return nil
}
