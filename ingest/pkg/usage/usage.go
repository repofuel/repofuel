package usage

import (
	"context"
	"net/http"

	"github.com/repofuel/repofuel/accounts/pkg/permission"
	"github.com/repofuel/repofuel/ingest/internal/entity"
)

type Tracker struct {
	VisitsDB entity.VisitDataSource
}

func NewTracker(visitsDB entity.VisitDataSource) *Tracker {
	return &Tracker{VisitsDB: visitsDB}
}

func (t *Tracker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		go t.LogVisit(r.Context(), r.Header)
		next.ServeHTTP(w, r)
	})
}

func (t *Tracker) LogVisit(ctx context.Context, h http.Header) error {
	viewer := permission.ViewerCtx(ctx)
	if viewer == nil || viewer.UserInfo == nil {
		return nil
	}

	return t.VisitsDB.Insert(context.Background(), &entity.Visit{
		UserID:  viewer.UserID,
		Referer: h.Get("referer"),
	})
}
