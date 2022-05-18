package entity

import (
	"context"

	"github.com/cheekybits/genny/generic"
)

//go:generate genny -in=./mongosrc/iter_generic.go -out=./mongosrc/iter_gnerated.go gen "item=entity.Commit,entity.Repository,entity.Job,entity.Organization,entity.PullRequest,entity.DeveloperExp,entity.File,metrics.ChangeMeasures,metrics.FileMeasures,entity.Feedback"
//go:generate genny -in=iter_generic.go -out=iter_gnerated.go gen "item=Commit,Repository,Job,Organization,PullRequest,DeveloperExp,File,ChangeMeasures,FileMeasures,Feedback"

type item generic.Type

type itemIter interface {
	ForEach(context.Context, func(*item) error) error
	Slice(context.Context) ([]*item, error)
}
