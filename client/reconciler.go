package client

import (
	"errors"
	"fmt"
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

func (r *Reconciler) RegisterUpdate(env *common.Envelope) []error {
	if env == nil {
		return []error{errors.New("nil env supplied")}
	}

	r.mutex.Lock()
	r.env = env

	r.pendingChanges = make(map[string]events.EventDispatch, len(r.dispatchers))
	for i, dispatcher := range r.dispatchers {
		r.pendingChanges[i] = dispatcher
	}
	metrics.ReconcilersActive.WithLabelValues(env.PublicIp.Host).Set(float64(len(r.pendingChanges)))

	r.mutex.Unlock()
	return r.dispatch()
}

func (r *Reconciler) dispatch() []error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.pendingChanges) == 0 {
		return nil
	}

	metrics.ReconcilerTimestamp.WithLabelValues(r.env.PublicIp.Host).SetToCurrentTime()
	log.Info().Msgf("Reconciling %d dispatchers", len(r.pendingChanges))

	timeStart := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(len(r.pendingChanges))
	var errs []error
	for key, dispatcher := range r.pendingChanges {
		var disp = dispatcher
		go func(key string) {
			err := disp.Notify(r.env)
			if err == nil {
				r.pendingChanges[key] = nil
				delete(r.pendingChanges, key)
				metrics.UpdatesDispatched.Inc()
				log.Info().Msgf("Reconciliation for dispatcher %s successful", key)
			} else {
				err = fmt.Errorf("reconciliation for dispatcher %s failed: %w", key, err)
				errs = append(errs, err)
				log.Warn().Msg(err.Error())
			}
			wg.Done()
		}(key)
	}

	wg.Wait()
	timeSpent := time.Since(timeStart)

	log.Info().Msgf("Spent %v on reconciliation (%d dispatchers)", timeSpent, len(r.dispatchers))
	metrics.ReconcilersActive.WithLabelValues(r.env.PublicIp.Host).Set(float64(len(r.pendingChanges)))
	return errs
}

func (r *Reconciler) Run() {
	interval := 1 * time.Minute
	ticker := time.NewTicker(interval)

	for range ticker.C {
		r.dispatch()
	}
}
