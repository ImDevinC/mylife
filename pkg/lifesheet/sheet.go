package lifesheet

import (
	"encoding/json"
	"fmt"
	"os"
)

type Lifesheet struct {
	Categories map[string]Category
}

type Category struct {
	Description string     `json:"description"`
	Schedule    string     `json:"schedule"`
	Questions   []Question `json:"questions"`
}

type Question struct {
	Key     string            `json:"key"`
	Text    string            `json:"question"`
	Type    string            `json:"type"`
	Buttons map[string]string `json:"buttons"`
	Replies map[string]string `json:"replies"`
}

func LoadFromFile(file string) (*Lifesheet, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file. %v", err)
	}
	var sheet map[string]Category
	err = json.Unmarshal(data, &sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal sheet. %v", err)
	}
	return &Lifesheet{Categories: sheet}, nil
}
