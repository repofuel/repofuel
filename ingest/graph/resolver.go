package graph

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"time"

	"github.com/repofuel/repofuel/ingest/graph/model"
	"github.com/repofuel/repofuel/ingest/internal/entity"
	"github.com/repofuel/repofuel/ingest/pkg/identifier"
	"github.com/repofuel/repofuel/ingest/pkg/manage"
	"github.com/repofuel/repofuel/pkg/repofuel"
)

//go:generate go run github.com/99designs/gqlgen

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

var fieldsFromAccount = map[string]bool{
	"username":  true,
	"firstName": true,
	"lastName":  true,
	"avatarUrl": true,
	"role":      true,
}

const (
	NodeTypeCommit     = "Commit"
	NodeTypeRepository = "Repository"
)

type Resolver struct {
	RepofuelClient *repofuel.Client
	FeedbackDB     entity.FeedbackDataSource
	CommitDB       entity.CommitDataSource
	RepositoryDB   entity.RepositoryDataSource
	PullRequestDB  entity.PullRequestDataSource
	JobDB          entity.JobDataSource
	OrganizationDB entity.OrganizationDataSource
	VisitDB        entity.VisitDataSource
	MonitorDB      entity.MonitorDataSource
	Manager        *manage.Manager
	Observables    *manage.ProgressObservableRegistry
}

func NodeIdToRepoId(id string) (identifier.RepositoryID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return identifier.RepositoryID{}, err
	}

	i := bytes.IndexByte(b, ':')
	nodeType, idBytes := string(b[:i]), b[i+1:]
	if nodeType != NodeTypeRepository {
		return identifier.RepositoryID{}, errors.New("unexpected node ID type")
	}

	return identifier.RepositoryIDFromBytes(idBytes), nil
}

func NodeIdToCommitId(id string) (*identifier.CommitID, error) {
	b, err := base64.StdEncoding.DecodeString(id)
	if err != nil {
		return nil, err
	}

	i := bytes.IndexByte(b, ':')
	nodeType, idBytes := string(b[:i]), b[i+1:]
	if nodeType != NodeTypeRepository {
		return nil, errors.New("unexpected node ID type")
	}

	return identifier.CommitIDFromBytes(idBytes), nil
}

func (r *Resolver) RemoveObserversOnCancel(ctx context.Context, obs chan *manage.ProgressObservable, ids []string) {
	go func() {
		<-ctx.Done()
		for _, id := range ids {
			r.Observables.Get(id).RemoveObserver(obs)
		}
	}()
}

var beginningOfTime = time.Unix(0, 0)

func TimeFromPeriod(p *model.Period) time.Time {
	if p == nil {
		return beginningOfTime
	}

	switch *p {
	case model.PeriodDay:
		return time.Now().AddDate(0, 0, -1)
	case model.PeriodWeek:
		return time.Now().AddDate(0, 0, -7)
	case model.PeriodMonth:
		return time.Now().AddDate(0, -1, 0)
	case model.PeriodYear:
		return time.Now().AddDate(-1, 0, 0)
	case model.PeriodAllTime:
		return beginningOfTime
	default:
		return time.Now()
	}
}
