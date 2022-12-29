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
	if err := populateFields(&msg); err != nil {
		return fmt.Errorf("failed to populate data. %w", err)
	}
	if _, err := d.collection.InsertOne(ctx, msg); err != nil {
		return fmt.Errorf("failed to save answer. %w", err)
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

func populateFields(answer *AnswerResponse) error {
	answer.ID = primitive.NewObjectID()
	ts := time.Now()
	answer.Timestamp = ts.Unix()

	answer.Day = ts.Day()
	answer.Hour = ts.Hour()
	answer.Minute = ts.Minute()
	answer.Year = ts.Year()
	month := int(ts.Month())
	answer.Month = month
	if month <= 3 {
		answer.Quarter = 1
	} else if month <= 6 {
		answer.Quarter = 2
	} else if month <= 9 {
		answer.Quarter = 3
	} else {
		answer.Quarter = 4
	}
	_, week := ts.ISOWeek()
	answer.Week = week
	yearWeekRaw := fmt.Sprintf("%d%02d", ts.Year(), week)
	yearWeek, err := strconv.Atoi(yearWeekRaw)
	if err != nil {
		return fmt.Errorf("failed to parse yearWeek. %w", err)
	}
	answer.YearWeek = yearWeek

	yearMonthRaw := fmt.Sprintf("%d%02d", ts.Year(), ts.Month())
	yearMonth, err := strconv.Atoi(yearMonthRaw)
	if err != nil {
		return fmt.Errorf("failed to parse yearMonth. %w", err)
	}
	answer.YearMonth = yearMonth
	return nil
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
