package repofuel

import (
	"context"
	"fmt"
	"net/http"

	"github.com/repofuel/repofuel/pkg/metrics"
)

type MLService service

type PredictionStatus uint8

const (
	PredictUnknownState PredictionStatus = iota + 1
	PredictLastModel
	PredictOk
	PredictLowTrainingData
	PredictClassUnbalanced
	PredictLowModelQuality
	PredictFailDataPreparing
	PredictFailTraining
	PredictFailPredicting
)

type PredictionResult struct {
	Predictions []Prediction       `json:"predictions,omitempty"`
	Quantiles   *metrics.Quantiles `json:"quantiles,omitempty"`
	Confidence  float32            `json:"confidence"`
	Status      PredictionStatus   `json:"status"`
}

type Prediction struct {
	CommitID   string  `json:"commit_id"`
	Score      float32 `json:"score"`
	Experience float32 `json:"experience"`
	History    float32 `json:"history"`
	Size       float32 `json:"size"`
	Diffusion  float32 `json:"diffusion"`
}

func (s *MLService) PredictByJob(ctx context.Context, repoID string, currentJob string, oldestJob string) (*PredictionResult, *http.Response, error) {
	var u string
	if oldestJob == "" {
		u = fmt.Sprintf("repositories/%s/jobs/%s/prediction", repoID, currentJob)
	} else {
		u = fmt.Sprintf("repositories/%s/jobs/%s/prediction?oldest_job=%s", repoID, currentJob, oldestJob)
	}

	req, err := (*service)(s).NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}

	var prediction PredictionResult
	resp, err := s.client.Do(req, &prediction)
	return &prediction, resp, err
}
