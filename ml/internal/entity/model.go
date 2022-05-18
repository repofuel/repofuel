package entity

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/repofuel/repofuel/ml/pkg/qa"
	"github.com/repofuel/repofuel/pkg/metrics"
	"github.com/repofuel/repofuel/pkg/repofuel"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrModelNotExist = errors.New("model not exist")
)

type ModelID = primitive.ObjectID

type ModelDataSource interface {
	Insert(context.Context, *Model) error
	FindRepoModels(ctx context.Context, repoId string, opts ...*options.FindOptions) (ModelIter, error)
	FindLatestModel(ctx context.Context, repoId string) (*Model, error)
	LogUsage(context.Context, ModelID) error
}

type Model struct {
	ID         ModelID                   `json:"id"           bson:"_id,omitempty"`
	RepoID     string                    `json:"repo_id"      bson:"repo_id"` //todo: change the time to RepoID
	Version    int                       `json:"version"      bson:"version"`
	Status     repofuel.PredictionStatus `json:"status"       bson:"status"`
	Path       string                    `json:"-"            bson:"path"`
	Report     *qa.ModelReport           `json:"report"       bson:"report,omitempty"`
	DataPoints *qa.DataPoints            `json:"data"         bson:"data,omitempty"`
	Expired    bool                      `json:"expired"      bson:"expired"`
	Medians    *qa.ModelMedians          `json:"medians"      json:"medians"`
	Quantiles  *metrics.Quantiles        `json:"quantiles,omitempty"`
	CreatedAt  time.Time                 `json:"created_at"   bson:"created_at"`
	LastUse    time.Time                 `json:"last_use"     bson:"last_use"`
}

func NewModel(repoId string, version int, path string, s repofuel.PredictionStatus, r *qa.ModelReport, data *qa.DataPoints, medians *qa.ModelMedians, quantiles *metrics.Quantiles) *Model {
	now := time.Now()

	return &Model{
		RepoID:     repoId,
		Version:    version,
		Status:     s,
		Path:       path,
		Report:     r,
		DataPoints: data,
		Expired:    false,
		Medians:    medians,
		Quantiles:  quantiles,
		CreatedAt:  now,
		LastUse:    now,
	}
}

func (m *Model) Confidence() float32 {
	if m.Report == nil {
		return 0
	}

	return m.Report.WeightedAvg.F1Score
}

func (m *Model) FileExists() bool {
	if _, err := os.Stat(m.Path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		log.Println("cannot access model file: ", err)
	}
	return true
}

func (m *Model) IsValid() bool {
	return !m.Expired && m.FileExists()
}

func (m *Model) NextVersion() int {
	if m == nil {
		return 1
	}
	return m.Version + 1
}

type ModelIter interface {
	ForEach(context.Context, func(*Model) error) error
}
