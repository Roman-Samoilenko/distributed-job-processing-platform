package models

import (
	"encoding/json"
	"errors"
)

// JobType определяет тип задачи
type JobType string

const (
	JobTypeHttpGet     JobType = "HTTP_GET"
	JobTypeImageResize JobType = "IMAGE_RESIZE"
	JobTypeSleep       JobType = "SLEEP"
)

// AllJobTypes - полный список всех поддерживаемых типов задач
var AllJobTypes = []JobType{
	JobTypeHttpGet,
	JobTypeImageResize,
	JobTypeSleep,
}

// JobStatus определяет текущее состояние
type JobStatus string

const (
	StatusCreated    JobStatus = "CREATED"
	StatusInProgress JobStatus = "IN_PROGRESS"
	StatusCompleted  JobStatus = "COMPLETED"
	StatusFailed     JobStatus = "FAILED"
)

// Job — основная структура задачи внутри воркера
type Job struct {
	ID        int64   `json:"id"`
	Type      JobType `json:"type"`
	Payload   string  `json:"payload"`
	CreatedAt int64   `json:"created_at"` // Unix timestamp
}

// JobResult — результат выполнения задачи
type JobResult struct {
	JobID  int64
	Status JobStatus
	Result string
	Error  string
}

// PayloadHttpGet — структура payload для HTTP задач
type PayloadHttpGet struct {
	URL string `json:"url"`
}

// PayloadImageResize — структура payload для картинок
type PayloadImageResize struct {
	ImageURL string `json:"image_url"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// PayloadSleep — структура payload для sleep
type PayloadSleep struct {
	DurationMs int `json:"duration_ms"`
}

// ParsePayload — вспомогательный метод
func ParsePayload[T any](payloadJSON string) (*T, error) {
	var t T
	if err := json.Unmarshal([]byte(payloadJSON), &t); err != nil {
		return nil, errors.New("invalid payload format: " + err.Error())
	}
	return &t, nil
}

// IsValid проверяет, что тип задачи поддерживается
func (jt JobType) IsValid() bool {
	for _, validType := range AllJobTypes {
		if jt == validType {
			return true
		}
	}
	return false
}
