package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnswerResponse struct {
	ID        primitive.ObjectID `bson:"_id"`
	Key       string             `bson:"key"`
	Answer    string             `bson:"answer"`
	Timestamp int64              `bson:"timestamp"`
	Type      string             `bson:"type"`
	Day       int                `bson:"day"`
	Hour      int                `bson:"hour"`
	Minute    int                `bson:"minute"`
	Year      int                `bson:"year"`
	Month     int                `bson:"month"`
	Quarter   int                `bson:"quarter"`
	YearWeek  int                `bson:"yearWeek"`
	YearMonth int                `bson:"yearMonth"`
	Week      int                `bson:"week"`
	Question  string             `bson:"question"`
	Source    string             `bson:"source"`
}

type PastValues struct {
	Values  []string
	Times   []string
	Minimum int
	Maximum int
}

type Database interface {
	SaveAnswer(context.Context, AnswerResponse) error
	GetValues(ctx context.Context, key string) (PastValues, error)
	// GetAverage(ctx context.Context, key string) error
}
