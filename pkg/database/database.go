package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnswerResponse struct {
	ID        primitive.ObjectID `bson:"_id"`
	Key       string
	Answer    string
	Skipped   bool
	Timestamp int64
	Type      string
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
