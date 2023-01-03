package fit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/oauth2"
)

type Aggregation struct {
	AggregateBy     []AggregateBy `json:"aggregateBy"`
	BucketByTime    BucketByTime  `json:"bucketByTime"`
	EndTimeMillis   int           `json:"endTimeMillis"`
	StartTimeMillis int           `json:"startTimeMillis"`
}

type AggregateBy struct {
	DatasourceID string `json:"dataSourceId"`
}

type BucketByTime struct {
	DurationMillis int `json:"durationMillis"`
}

func Get(ctx context.Context, token string) error {
	agg := Aggregation{
		AggregateBy: []AggregateBy{{
			DatasourceID: "derived:com.google.step_count.delta:com.google.android.gms:estimated_steps",
		}},
		BucketByTime:    BucketByTime{DurationMillis: 86400000},
		EndTimeMillis:   1672360200000,
		StartTimeMillis: 1672300800000,
	}
	outgoingPayload, err := json.Marshal(agg)
	if err != nil {
		return fmt.Errorf("failed to marshal body. %w", err)
	}
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	req, err := http.NewRequest(http.MethodPost, "https://www.googleapis.com/fitness/v1/users/me/dataset:aggregate", bytes.NewReader(outgoingPayload))
	if err != nil {
		return fmt.Errorf("failed to create request. %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request. %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body. %w", err)
	}
	payload := map[string]interface{}{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		return fmt.Errorf("failed to unmarshal payload.  %w", err)
	}
	fmt.Printf("%+v\n", payload)
	return nil
}
