package ml

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"path"
	"strconv"

	"github.com/repofuel/repofuel/ml/internal/entity"
	"github.com/repofuel/repofuel/ml/pkg/qa"
	"github.com/repofuel/repofuel/pkg/metrics"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"golang.org/x/oauth2"
)

const (
	modelsPath       = "./models"
	ignoredDaysCount = "90"
	pythonExec       = "/usr/local/bin/python3"
)

var (
	ErrInvalidModelVersion = errors.New("invalid model version format, it should be integer")
)

type ModelResult struct {
	IsBuilt     bool                      `json:"is_built"`
	IsPredicted bool                      `json:"is_predicted"`
	Status      repofuel.PredictionResult `json:"status"`
	Error       *ModelError               `json:"error,omitempty,omitempty"`
	DataPoints  *qa.DataPoints            `json:"data_points,omitempty"`
	Report      *qa.ModelReport           `json:"model_report,omitempty"`
	Medians     *qa.ModelMedians          `json:"medians,omitempty"`
	Quantiles   *metrics.Quantiles        `json:"quantiles,omitempty"`
	Predictions json.RawMessage           `json:"prediction,omitempty"`
}

type ModelError struct {
	Stage   Stage  `json:"stage"`
	Message string `json:"message"`
}

// PredictionResult should match the type repofuel.PredictionResult.
// The only different is the type of the filed `Predictions` is `json.RawMessage`
// instead of `[]repofuel.Prediction` to avoid unnecessary encoding
// and decoding of the result.
type PredictionResult struct {
	Predictions json.RawMessage           `json:"predictions,omitempty"`
	Quantiles   *metrics.Quantiles        `json:"quantiles,omitempty"`
	Confidence  float32                   `json:"confidence"`
	Status      repofuel.PredictionStatus `json:"status"`
}

func (err *ModelError) Error() string {
	return fmt.Sprintf("AI service failed in %s stage", err.Stage)
}

func (err *ModelError) String() string {
	return fmt.Sprintf("AI service failed in %s stage: %s", err.Stage, err.Message)
}

type ModelServer struct {
	ingestURL string
	auth      oauth2.TokenSource
	modelsDB  entity.ModelDataSource
}

func NewModelServer(auth oauth2.TokenSource, m entity.ModelDataSource, ingestURL string) *ModelServer {
	return &ModelServer{
		ingestURL: ingestURL,
		auth:      auth,
		modelsDB:  m,
	}
}

func (s *ModelServer) Predict(ctx context.Context, repoID, oldestJob, currentJob string) (*PredictionResult, error) {
	lastModel, err := s.modelsDB.FindLatestModel(ctx, repoID)
	if err != nil {
		if err != entity.ErrModelNotExist {
			return nil, err
		}
	}

	if lastModel == nil || !lastModel.IsValid() {
		// todo: should we build a new model?
		// fixme: if we rebuild the model it will override the old risk scores
		return s.PredictUsingNewModel(ctx, repoID, lastModel.NextVersion(), currentJob)
	}

	return s.PredictUsingLastModel(ctx, lastModel, oldestJob, currentJob)
}

func (s *ModelServer) PredictUsingNewModel(ctx context.Context, repoID string, version int, lastJob string) (*PredictionResult, error) {
	modelFile := path.Join(modelsPath, repoID, strconv.Itoa(version))

	res, err := s.cmdBuild(ctx, repoID, modelFile, lastJob)
	if err != nil {
		return nil, err
	}

	if !res.IsBuilt {
		//todo: should clean the model file
		return &PredictionResult{
			Status: predictionStatus(res),
		}, nil
	}

	status := predictionStatus(res)
	m := entity.NewModel(repoID, version, modelFile, status, res.Report, res.DataPoints, res.Medians, res.Quantiles)
	err = s.modelsDB.Insert(ctx, m)
	if err != nil {
		return nil, err
	}

	return &PredictionResult{
		Predictions: res.Predictions,
		Quantiles:   res.Quantiles,
		Confidence:  m.Confidence(),
		Status:      m.Status,
	}, nil
}

func (s *ModelServer) PredictUsingLastModel(ctx context.Context, lastModel *entity.Model, startJob, lastJob string) (*PredictionResult, error) {
	medians := lastModel.Medians.All
	mediansScore := lastModel.Report.MedianScore
	res, err := s.cmdPredict(ctx, lastModel.RepoID, lastModel.Path, medians, mediansScore, startJob, lastJob)
	if err != nil {
		return nil, err
	}

	if !res.IsPredicted {
		return &PredictionResult{
			Confidence: lastModel.Confidence(),
			Status:     lastModel.Status,
		}, nil
	}

	err = s.modelsDB.LogUsage(ctx, lastModel.ID)
	if err != nil {
		return nil, err
	}

	return &PredictionResult{
		Predictions: res.Predictions,
		Quantiles:   lastModel.Quantiles,
		Confidence:  lastModel.Confidence(),
		Status:      lastModel.Status,
	}, nil
}

func (s *ModelServer) cmdBuild(ctx context.Context, repoID, modelPath, lastJob string) (*ModelResult, error) {
	token, err := s.auth.Token()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, pythonExec, "./python/main.py",
		"--ingest-url", s.ingestURL,
		"--auth", "Bearer "+token.AccessToken,
		"--repo-id", repoID,
		"--last-job-id", lastJob,
		"build",
		"--model", modelPath,
	)

	return s.runModelCMD(cmd)
}

func (s *ModelServer) cmdPredict(ctx context.Context, repoID, modelPath string, modelMedians metrics.ChangeMeasures, mediansScore float32, startJob, lastJob string) (*ModelResult, error) {
	bytesModelMedians, err := json.Marshal(modelMedians)
	if err != nil {
		return nil, err
	}

	token, err := s.auth.Token()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, pythonExec, "./python/main.py",
		"--ingest-url", s.ingestURL,
		"--auth", "Bearer "+token.AccessToken,
		"--repo-id", repoID,
		"--start-job-id", startJob,
		"--last-job-id", lastJob,
		"predict",
		"--medians", string(bytesModelMedians),
		"--medians-score", fmt.Sprintf("%f", mediansScore),
		"--model", modelPath)

	return s.runModelCMD(cmd)
}

func (s *ModelServer) runModelCMD(cmd *exec.Cmd) (*ModelResult, error) {
	// todo: we can use a pool of buffers
	var outErr bytes.Buffer
	var out bytes.Buffer
	cmd.Stderr = &outErr
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("model CMD error - %w: %s", err, outErr.Bytes())
	}

	var result ModelResult
	err = json.Unmarshal(out.Bytes(), &result)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		log.Println(result.Error.String())
	}

	return &result, nil
}

func predictionStatus(res *ModelResult) repofuel.PredictionStatus {
	if !res.IsBuilt && res.IsPredicted {
		return repofuel.PredictLastModel
	}

	if res.DataPoints == nil {
		return repofuel.PredictFailDataPreparing
	}

	all := res.DataPoints.Train.NumBuggy + res.DataPoints.Train.NumClean
	if all < 50 {
		return repofuel.PredictLowTrainingData
	}

	balance := float32(res.DataPoints.Train.NumBuggy) / float32(all)
	if balance < .10 || balance > .90 || res.DataPoints.Train.NumBuggy < 10 || res.DataPoints.Train.NumClean < 10 {
		return repofuel.PredictClassUnbalanced
	}

	if res.Report == nil {
		return repofuel.PredictFailTraining
	}

	if res.Report.WeightedAvg.F1Score < .5 {
		return repofuel.PredictLowModelQuality
	}

	if res.Error != nil {
		switch res.Error.Stage {
		case StageDataPreparation, StageDataSplitting:
			return repofuel.PredictFailDataPreparing
		case StageModelBuilding:
			return repofuel.PredictFailTraining
		case StagePredicting:
			return repofuel.PredictFailPredicting
		}
	}

	if res.IsBuilt && res.IsPredicted {
		return repofuel.PredictOk
	}

	return repofuel.PredictUnknownState
}
