package fit_test

import (
	"context"
	"os"
	"testing"

	"github.com/imdevinc/mylife/pkg/fit"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	godotenv.Load("../../.env")
	err := fit.Get(context.TODO(), os.Getenv("GCLOUD_OAUTH_TOKEN"))
	if !assert.NoError(t, err, "expected no error") {
		t.FailNow()
	}
}
