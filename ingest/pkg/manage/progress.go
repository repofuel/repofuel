package manage

import (
	"context"
	"sync"
	"time"

	"github.com/repofuel/repofuel/ingest/pkg/status"
	"github.com/rs/zerolog/log"
)

type Progress struct {
	Status  status.Stage `json:"status"`
	Total   *int         `json:"total"`
	Current int          `json:"current"`
}

type ProgressObservableRegistry struct {
	observables map[string]*ProgressObservable
	checker     StatesChecker
	mu          sync.Mutex
}

type StatesChecker interface {
	NodeStatus(ctx context.Context, id string) (status.Stage, error)
}

func newProgressObservableRegistry(checker StatesChecker) *ProgressObservableRegistry {
	return &ProgressObservableRegistry{
		observables: make(map[string]*ProgressObservable),
		checker:     checker,
	}
}

func (reg *ProgressObservableRegistry) RemoveEmpty(id string) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	obs, ok := reg.observables[id]
	if !ok {
		return
	}

	if len(obs.observers) > 0 {
		return
	}

	delete(reg.observables, id)
}

func (reg *ProgressObservableRegistry) Get(id string) *ProgressObservable {
	return reg.observables[id]
}

func (reg *ProgressObservableRegistry) GetOrCreate(id string) *ProgressObservable {
	obs, ok := reg.observables[id]
	if ok {
		return obs
	}

	reg.mu.Lock()
	defer reg.mu.Unlock()

	obs, ok = reg.observables[id]
	if ok {
		return obs
	}

	obs = newProgressObservable(id, status.Queued, 2*time.Second)
	reg.observables[id] = obs

	return obs
}

func (reg *ProgressObservableRegistry) GetOrCreateStateful(ctx context.Context, id string) (*ProgressObservable, error) {
	reg.mu.Lock()
	defer reg.mu.Unlock()

	obs, ok := reg.observables[id]
	if !ok {
		initStatus, err := reg.checker.NodeStatus(ctx, id)
		if err != nil {
			return nil, err
		}
		obs = newProgressObservable(id, initStatus, time.Second)
	}
	reg.observables[id] = obs

	return obs, nil
}

type ProgressObservable struct {
	id         string
	progress   Progress
	waiting    time.Duration
	notifiedAt time.Time
	observers  []chan<- *ProgressObservable
	mu         sync.Mutex
}

func (po *ProgressObservable) Target() string {
	return po.id
}

func newProgressObservable(id string, intStatus status.Stage, waiting time.Duration) *ProgressObservable {
	return &ProgressObservable{
		id:      id,
		waiting: waiting,
		progress: Progress{
			Status: intStatus,
		},
	}
}

func (po *ProgressObservable) IncreaseProgress(n int) {
	po.progress.Current += n
	po.notify()
}

func (po *ProgressObservable) SetStageTotal(n int) {
	po.progress.Total = &n
}

func (po *ProgressObservable) SetNewStage(s status.Stage, notify bool) {
	po.progress = Progress{
		Status:  s,
		Total:   nil,
		Current: 0,
	}

	if notify {
		go po.notifyNow()
	}
}

func (po *ProgressObservable) AddObserver(obs chan<- *ProgressObservable) {
	po.mu.Lock()
	defer po.mu.Unlock()

	po.observers = append(po.observers, obs)
}

func (po *ProgressObservable) RemoveObserver(obsToDelete chan<- *ProgressObservable) {
	if po == nil {
		return
	}

	po.mu.Lock()
	defer po.mu.Unlock()

	a := po.observers
	for i, obs := range a {
		if obs == obsToDelete {
			a[i] = a[len(a)-1]
			a[len(a)-1] = nil
			po.observers = a[:len(a)-1]
			return
		}
	}
}

func (po *ProgressObservable) Progress() *Progress {
	return &po.progress
}

func (po *ProgressObservable) Status() status.Stage {
	return po.progress.Status
}

func (po *ProgressObservable) notify() {
	if len(po.observers) == 0 || time.Now().Sub(po.notifiedAt) < po.waiting {
		return
	}

	go po.notifyNow()
}

func (po *ProgressObservable) notifyNow() {
	// Without Lock we will have a liter function but notifications could
	// be missed if an observer is being removed in the same time.
	//po.mu.Lock()
	//defer po.mu.Unlock()

	po.notifiedAt = time.Now()
	for _, obs := range po.observers {
		select {
		case obs <- po:
		default:
			log.Warn().Str("node_id", po.id).Msg("dropped a progress notification")
		}
	}
}
