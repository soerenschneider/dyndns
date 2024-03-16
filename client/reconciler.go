package client

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/internal/common"
	"github.com/soerenschneider/dyndns/internal/metrics"
	"go.uber.org/multierr"
)

type Reconciler struct {
	env         *common.UpdateRecordRequest
	dispatchers map[string]EventDispatch
	mutex       sync.Mutex

	stopAfterFirstSuccess bool
	pendingChanges        map[string]EventDispatch
}

func NewReconciler(dispatchers map[string]EventDispatch, stopAfterFirstSuccess bool) (*Reconciler, error) {
	if len(dispatchers) < 1 {
		return nil, errors.New("no dispatchers supplied")
	}

	return &Reconciler{
		dispatchers:           dispatchers,
		mutex:                 sync.Mutex{},
		stopAfterFirstSuccess: stopAfterFirstSuccess,
	}, nil
}

func (r *Reconciler) RegisterUpdate(env *common.UpdateRecordRequest) error {
	if env == nil {
		return errors.New("nil env supplied")
	}

	r.mutex.Lock()
	r.env = env

	r.pendingChanges = make(map[string]EventDispatch, len(r.dispatchers))
	for i, dispatcher := range r.dispatchers {
		r.pendingChanges[i] = dispatcher
	}
	metrics.ReconcilersActive.WithLabelValues(env.PublicIp.Host).Set(float64(len(r.pendingChanges)))

	r.mutex.Unlock()
	return r.dispatch()
}

func (r *Reconciler) dispatch() error {
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
	errLock := &sync.Mutex{}
	var errs error
	var successFullDispatches atomic.Int32
	for key, dispatcher := range r.pendingChanges {
		var disp = dispatcher
		go func(key string) {
			err := disp.Notify(r.env)
			if err == nil {
				successFullDispatches.Add(1)
				r.pendingChanges[key] = nil
				delete(r.pendingChanges, key)
				metrics.UpdatesDispatched.Inc()
				log.Info().Msgf("Reconciliation for dispatcher %s successful", key)
			} else {
				errLock.Lock()
				metrics.UpdateDispatchErrors.WithLabelValues(key).Inc()
				errs = multierr.Append(errs, fmt.Errorf("reconciliation for dispatcher %s failed: %w", key, err))
				errLock.Unlock()
			}
			wg.Done()
		}(key)
	}

	wg.Wait()
	timeSpent := time.Since(timeStart)

	if r.stopAfterFirstSuccess && successFullDispatches.Load() > 0 && len(r.pendingChanges) > 0 {
		log.Info().Msgf("Stopping reconciliation for %d pending changes due to %d successful dispatches", len(r.pendingChanges), successFullDispatches.Load())
		r.pendingChanges = nil
	}

	log.Info().Msgf("Spent %v on reconciliation (%d dispatchers)", timeSpent, len(r.dispatchers))
	metrics.ReconcilersActive.WithLabelValues(r.env.PublicIp.Host).Set(float64(len(r.pendingChanges)))
	return errs
}

func (r *Reconciler) Run() {
	interval := 1 * time.Minute
	ticker := time.NewTicker(interval)

	for range ticker.C {
		if err := r.dispatch(); err != nil {
			log.Error().Err(err).Msg("running reconciler produced errors")
		}
	}
}
