package prediction

import (
	"context"
	"sync"
	"time"

	"github.com/ben/ikite-go/internal/store"
)

const cacheTTL = 20 * time.Minute

var (
	cacheMu     sync.RWMutex
	cached      *Result
	cachedKey   string
	cachedAt    time.Time
	computeMu   sync.Mutex
)

// ComputeCached returns a prediction, reusing a recent result for the same hour.
func ComputeCached(st *store.Store, now time.Time, loc *time.Location) (*Result, error) {
	now = now.In(loc)
	key := now.Format("2006-01-02-15")

	cacheMu.RLock()
	if cached != nil && cachedKey == key && time.Since(cachedAt) < cacheTTL {
		res := cloneResult(cached)
		cacheMu.RUnlock()
		attachCurrent(st, res)
		return res, nil
	}
	cacheMu.RUnlock()

	computeMu.Lock()
	defer computeMu.Unlock()

	cacheMu.RLock()
	if cached != nil && cachedKey == key && time.Since(cachedAt) < cacheTTL {
		res := cloneResult(cached)
		cacheMu.RUnlock()
		attachCurrent(st, res)
		return res, nil
	}
	cacheMu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	res, err := compute(ctx, st, now, loc)
	if err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cached = cloneResult(res)
	cachedKey = key
	cachedAt = time.Now()
	cacheMu.Unlock()

	return res, nil
}

func cloneResult(r *Result) *Result {
	if r == nil {
		return nil
	}
	cp := *r
	return &cp
}

func attachCurrent(st *store.Store, res *Result) {
	wind, gust, _, temp, humidity, pressure, err := st.KyLatestReading()
	if err == nil && wind > 0 {
		res.Current = formatCurrent(wind, gust, temp, humidity, pressure)
	}
}
