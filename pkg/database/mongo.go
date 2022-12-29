package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

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

func (d *MongoDatabase) SaveAnswer(ctx context.Context, msg AnswerResponse) error {
	if msg.Timestamp == 0 {
		msg.Timestamp = time.Now().Unix()
	}
	msg.ID = primitive.NewObjectID()
	_, err := d.collection.InsertOne(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to save answer. %v", err)
	}
	return nil
}

func (d *MongoDatabase) Disconnect(ctx context.Context) {
	d.client.Disconnect(ctx)
}

func (d *MongoDatabase) GetValues(ctx context.Context, key string) (PastValues, error) {
	filter := bson.D{primitive.E{Key: "key", Value: key}}
	var limit int64 = 300
	cursor, err := d.collection.Find(ctx, filter, &options.FindOptions{Limit: &limit})
	if err != nil {
		return PastValues{}, fmt.Errorf("failed to query database. %v", err)
	}
	results := []AnswerResponse{}
	err = cursor.All(ctx, &results)
	if err != nil {
		return PastValues{}, fmt.Errorf("failed to marshal database response. %v", err)
	}
	returnValue := PastValues{}
	for _, result := range results {
		returnValue.Values = append(returnValue.Values, result.Answer)
		t := time.Unix(result.Timestamp, 0).Format("02-01")
		returnValue.Times = append(returnValue.Times, t)
		val, err := strconv.Atoi(result.Answer)
		if err != nil {
			return PastValues{}, fmt.Errorf("failed to parse answer %s. %v", result.Answer, err)
		}
		if val < returnValue.Minimum {
			returnValue.Minimum = val
		}
		if val > returnValue.Maximum {
			returnValue.Maximum = val
		}
	}
	return returnValue, nil
}

// func (d *MongoDatabase) GetAverage(ctx context.Context, key string) error {
// 	query := bson.D{
// 		{"$group", bson.D{
// 			{"_id", "$key"},
// 			{"average", bson.D{{"$avg", "$answer"}}},
// 		}},
// 	}
// 	match := bson.D{
// 		{"$match", bson.D{{"key", key}}},
// 	}
// 	cursor, err := d.collection.Aggregate(ctx, mongo.Pipeline{match, query})
// 	if err != nil {
// 		return fmt.Errorf("failed to get aggregate. %v", err)
// 	}
// 	avg := []bson.M{}
// 	if err = cursor.All(ctx, &avg); err != nil {
// 		return fmt.Errorf("failed to get average cursor data. %v", err)
// 	}
// 	for _, a := range avg {
// 		fmt.Printf("Average price for %v: %v\n", a["_id"], a["average"])
// 	}
// 	return nil
// }
