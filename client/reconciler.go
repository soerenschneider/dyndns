package client

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/events"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"sync"
	"time"
)

type Reconciler struct {
	env         *common.Envelope
	dispatchers map[string]events.EventDispatch
	mutex       sync.Mutex

	pendingChanges map[string]events.EventDispatch
}

func NewReconciler(dispatchers map[string]events.EventDispatch) (*Reconciler, error) {
	if len(dispatchers) < 1 {
		return nil, errors.New("no dispatchers supplied")
	}

	return &Reconciler{
		dispatchers: dispatchers,
		mutex:       sync.Mutex{},
	}, nil
}

func (r *Reconciler) RegisterUpdate(env *common.Envelope) {
	if env == nil {
		return
	}

	r.mutex.Lock()
	r.env = env

	r.pendingChanges = make(map[string]events.EventDispatch, len(r.dispatchers))
	for i, dispatcher := range r.dispatchers {
		r.pendingChanges[i] = dispatcher
	}
	metrics.ReconcilersActive.WithLabelValues(env.PublicIp.Host).Set(float64(len(r.pendingChanges)))

	r.mutex.Unlock()
	r.dispatch()
}

func (r *Reconciler) dispatch() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.pendingChanges) == 0 {
		return
	}

	metrics.ReconcilerTimestamp.WithLabelValues(r.env.PublicIp.Host).SetToCurrentTime()
	log.Info().Msgf("Reconciling %d dispatchers", len(r.pendingChanges))

	timeStart := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(len(r.pendingChanges))
	for key, _ := range r.pendingChanges {
		dispatcher := r.pendingChanges[key]
		go func(key string) {
			err := dispatcher.Notify(r.env)
			if err == nil {
				r.pendingChanges[key] = nil
				delete(r.pendingChanges, key)
				metrics.UpdatesDispatched.Inc()
				log.Info().Msgf("Reconciliation for dispatcher %s successful", key)
			} else {
				log.Warn().Msgf("Reconciliation for dispatcher %s failed: %v", key, err)
			}
			wg.Done()
		}(key)
	}

	wg.Wait()
	timeSpent := time.Now().Sub(timeStart)

	log.Info().Msgf("Spent %v on reconciliation (%d dispatchers)", timeSpent, len(r.dispatchers))
	metrics.ReconcilersActive.WithLabelValues(r.env.PublicIp.Host).Set(float64(len(r.pendingChanges)))
}

func (r *Reconciler) Run() {
	interval := 1 * time.Minute
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			r.dispatch()
		}
	}
}
