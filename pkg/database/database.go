package database

import (
	"context"

	"github.com/imdevinc/mylife/pkg/bot"
)

type Database interface {
	SaveAnswer(context.Context, bot.MessageResponse) error
}
