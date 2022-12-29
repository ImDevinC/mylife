package database_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/imdevinc/mylife/pkg/database"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetValues(t *testing.T) {
	if testing.Short() {
		t.Skip("this connects to a live database, only run on full tests")
	}
	godotenv.Load("../../.env")
	db, err := database.NewMongoDB(context.TODO(), database.MongoDatabaseOptions{
		Username: os.Getenv("MONGO_USERNAME"),
		Password: os.Getenv("MONGO_PASSWORD"),
		URL:      os.Getenv("MONGO_URL"),
		Port:     os.Getenv("MONGO_PORT"),
		Database: os.Getenv("MONGO_DB"),
	})
	if !assert.NoError(t, err, "expected no error") {
		t.FailNow()
	}
	vals, err := db.GetValues(context.TODO(), "mood")
	if !assert.NoError(t, err, "expected no error") {
		t.FailNow()
	}
	if !assert.Greater(t, len(vals.Values), 0, "expected more results") {
		t.Fail()
	}
	url := fmt.Sprintf("https://chart.googleapis.com/chart?cht=lc&chd=t:%s&chs=800x350&chl=%s&chtt=%s&chf=bg,s,e0e0e0&chco=000000,0000FF&chma=30,30,30,30&chds=%d,%d", strings.Join(vals.Values, ","), strings.Join(vals.Times, "%7C"), "mood", vals.Minimum, vals.Maximum)
	t.Log(url)
}
