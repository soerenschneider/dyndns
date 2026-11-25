package client

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/soerenschneider/dyndns/v2/internal/metrics"
	"github.com/soerenschneider/dyndns/v2/pkg/update"
	"go.uber.org/multierr"
)

type RecordUpdater interface {
	UpdateRecord(ctx context.Context, msg update.UpdateRecordRequest) error
}

type Reconciler struct {
	env      *update.UpdateRecordRequest
	updaters map[string]RecordUpdater
	mutex    sync.Mutex

	stopAfterFirstSuccess bool
	pendingChanges        map[string]RecordUpdater
}

func NewReconciler(updaters map[string]RecordUpdater, stopAfterFirstSuccess bool) (*Reconciler, error) {
	if len(updaters) < 1 {
		return nil, errors.New("no dispatchers supplied")
	}

	return &Reconciler{
		updaters:              updaters,
		mutex:                 sync.Mutex{},
		stopAfterFirstSuccess: stopAfterFirstSuccess,
	}, nil
}

func (r *Reconciler) RegisterUpdate(ctx context.Context, env *update.UpdateRecordRequest) error {
	if env == nil {
		return errors.New("nil env supplied")
	}

	r.mutex.Lock()
	r.env = env

	r.pendingChanges = make(map[string]RecordUpdater, len(r.updaters))
	for i, dispatcher := range r.updaters {
		r.pendingChanges[i] = dispatcher
	}
	metrics.ReconcilersActive.WithLabelValues(env.PublicIp.Host).Set(float64(len(r.pendingChanges)))

	r.mutex.Unlock()
	return r.dispatch(ctx)
}

func (r *Reconciler) dispatch(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if len(r.pendingChanges) == 0 {
		return nil
	}

	metrics.ReconcilerTimestamp.WithLabelValues(r.env.PublicIp.Host).SetToCurrentTime()
	log.Info().Str("component", "reconciler").Int("num_dispatchers", len(r.pendingChanges)).Msg("Reconciling dispatchers")

	timeStart := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(len(r.pendingChanges))
	// The goroutines only record their results; the map is changed after all of
	// them are done, as concurrent writes to a map are not allowed, not even to
	// distinct keys, and the loop below is still iterating it.
	resultLock := &sync.Mutex{}
	var errs error
	var succeeded []string
	for key, dispatcher := range r.pendingChanges {
		var disp = dispatcher
		go func(key string) {
			err := disp.UpdateRecord(ctx, *r.env)
			resultLock.Lock()
			if err == nil {
				succeeded = append(succeeded, key)
				metrics.UpdatesDispatched.WithLabelValues(r.env.PublicIp.Host).Inc()
				log.Info().Str("component", "reconciler").Str("dispatcher", key).Msg("Reconciliation successful")
			} else {
				metrics.UpdateDispatchErrors.WithLabelValues(key).Inc()
				errs = multierr.Append(errs, fmt.Errorf("reconciliation for dispatcher %s failed: %w", key, err))
			}
			resultLock.Unlock()
			wg.Done()
		}(key)
	}

	wg.Wait()
	timeSpent := time.Since(timeStart)

	for _, key := range succeeded {
		delete(r.pendingChanges, key)
	}

	if r.stopAfterFirstSuccess && len(succeeded) > 0 && len(r.pendingChanges) > 0 {
		log.Info().Str("component", "reconciler").Int("pending", len(r.pendingChanges)).Int("successful_dispatches", len(succeeded)).Msg("Stopping reconciliation due to successful dispatches")
		r.pendingChanges = nil
	}

	log.Info().Str("component", "reconciler").Float64("seconds", timeSpent.Seconds()).Int("num_dispatchers", len(r.updaters)).Msgf("Spent %v on reconciliation", timeSpent)
	metrics.ReconcilersActive.WithLabelValues(r.env.PublicIp.Host).Set(float64(len(r.pendingChanges)))
	return errs
}

func (r *Reconciler) Run(ctx context.Context) {
	interval := 1 * time.Minute
	ticker := time.NewTicker(interval)

	for range ticker.C {
		if err := r.dispatch(ctx); err != nil {
			log.Error().Err(err).Msg("running reconciler produced errors")
		}
	}
}
