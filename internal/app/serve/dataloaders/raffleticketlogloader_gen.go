// Code generated by github.com/vektah/dataloaden, DO NOT EDIT.

package dataloaders

import (
	"sync"
	"time"

	"nt-folly-xmaxx-comp/internal/app/serve/graphql/gqlmodels"
)

// RaffleTicketLogLoaderConfig captures the config to create a new RaffleTicketLogLoader
type RaffleTicketLogLoaderConfig struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []IDRaffleUserKey) ([][]*gqlmodels.RaffleTicketLog, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewRaffleTicketLogLoader creates a new RaffleTicketLogLoader given a fetch, wait, and maxBatch
func NewRaffleTicketLogLoader(config RaffleTicketLogLoaderConfig) *RaffleTicketLogLoader {
	return &RaffleTicketLogLoader{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// RaffleTicketLogLoader batches and caches requests
type RaffleTicketLogLoader struct {
	// this method provides the data for the loader
	fetch func(keys []IDRaffleUserKey) ([][]*gqlmodels.RaffleTicketLog, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[IDRaffleUserKey][]*gqlmodels.RaffleTicketLog

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *raffleTicketLogLoaderBatch

	// mutex to prevent races
	mu sync.Mutex
}

type raffleTicketLogLoaderBatch struct {
	keys    []IDRaffleUserKey
	data    [][]*gqlmodels.RaffleTicketLog
	error   []error
	closing bool
	done    chan struct{}
}

// Load a RaffleTicketLog by key, batching and caching will be applied automatically
func (l *RaffleTicketLogLoader) Load(key IDRaffleUserKey) ([]*gqlmodels.RaffleTicketLog, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a RaffleTicketLog.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *RaffleTicketLogLoader) LoadThunk(key IDRaffleUserKey) func() ([]*gqlmodels.RaffleTicketLog, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() ([]*gqlmodels.RaffleTicketLog, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &raffleTicketLogLoaderBatch{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() ([]*gqlmodels.RaffleTicketLog, error) {
		<-batch.done

		var data []*gqlmodels.RaffleTicketLog
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *RaffleTicketLogLoader) LoadAll(keys []IDRaffleUserKey) ([][]*gqlmodels.RaffleTicketLog, []error) {
	results := make([]func() ([]*gqlmodels.RaffleTicketLog, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	raffleTicketLogs := make([][]*gqlmodels.RaffleTicketLog, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		raffleTicketLogs[i], errors[i] = thunk()
	}
	return raffleTicketLogs, errors
}

// LoadAllThunk returns a function that when called will block waiting for a RaffleTicketLogs.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *RaffleTicketLogLoader) LoadAllThunk(keys []IDRaffleUserKey) func() ([][]*gqlmodels.RaffleTicketLog, []error) {
	results := make([]func() ([]*gqlmodels.RaffleTicketLog, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([][]*gqlmodels.RaffleTicketLog, []error) {
		raffleTicketLogs := make([][]*gqlmodels.RaffleTicketLog, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			raffleTicketLogs[i], errors[i] = thunk()
		}
		return raffleTicketLogs, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *RaffleTicketLogLoader) Prime(key IDRaffleUserKey, value []*gqlmodels.RaffleTicketLog) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		// make a copy when writing to the cache, its easy to pass a pointer in from a loop var
		// and end up with the whole cache pointing to the same value.
		cpy := make([]*gqlmodels.RaffleTicketLog, len(value))
		copy(cpy, value)
		l.unsafeSet(key, cpy)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *RaffleTicketLogLoader) Clear(key IDRaffleUserKey) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *RaffleTicketLogLoader) unsafeSet(key IDRaffleUserKey, value []*gqlmodels.RaffleTicketLog) {
	if l.cache == nil {
		l.cache = map[IDRaffleUserKey][]*gqlmodels.RaffleTicketLog{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *raffleTicketLogLoaderBatch) keyIndex(l *RaffleTicketLogLoader, key IDRaffleUserKey) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *raffleTicketLogLoaderBatch) startTimer(l *RaffleTicketLogLoader) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *raffleTicketLogLoaderBatch) end(l *RaffleTicketLogLoader) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
