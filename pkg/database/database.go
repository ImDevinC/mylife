package database

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnswerResponse struct {
	ID        primitive.ObjectID `bson:"_id"`
	Key       string             `bson:"key"`
	Answer    string             `bson:"answer"`
	Skipped   bool               `bson:"skipped"`
	Timestamp int64              `bson:"timestamp"`
	Type      string             `bson:"type"`
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
