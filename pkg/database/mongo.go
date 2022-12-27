package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/imdevinc/mylife/pkg/bot"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDatabase struct {
	client     *mongo.Client
	collection *mongo.Collection
}

type MongoDatabaseOptions struct {
	Username string
	Password string
	URL      string
	Port     string
	Database string
}

func NewMongoDB(ctx context.Context, cfg MongoDatabaseOptions) (Database, error) {
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%s", cfg.Username, cfg.Password, cfg.URL, cfg.Port)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database. %v", err)
	}
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping database. %v", err)
	}
	collection := client.Database(cfg.Database).Collection("answers")
	return &MongoDatabase{client: client, collection: collection}, nil
}

func (d *MongoDatabase) SaveAnswer(ctx context.Context, msg bot.MessageResponse) error {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	doc := bson.D{
		primitive.E{Key: "key", Value: msg.QuestionKey},
		primitive.E{Key: "answer", Value: msg.Text},
		primitive.E{Key: "skipped", Value: msg.Skipped},
		primitive.E{Key: "timestamp", Value: ts}}
	_, err := d.collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to save answer. %v", err)
	}
	return nil
}

func (d *MongoDatabase) Disconnect(ctx context.Context) {
	d.client.Disconnect(ctx)
}
