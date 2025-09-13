package perch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Loader[T any] func(context.Context, string) (T, error)

// Perch is a bounded, per-key TTL, singleflight, zero-alloc-on-hit LRU cache.
type Perch[T any] struct {
	mu    *sync.Mutex
	cap   int
	table map[string]uint32 // key -> slot index (1..len(slots)); 0 means nil
	slots []entry[T]        // 1-based addressing to keep 0 = null

	// intrusive doubly-linked LRU list using indices
	head uint32 // MRU
	tail uint32 // LRU

	free uint32 // head of freelist (stack of indices)
}

type entry[T any] struct {
	// per-entry state (guards loading/result fields)
	mu  *sync.Mutex
	cv  *sync.Cond
	key string

	// LRU links (indices into Cache.slots)
	prev, next uint32

	// value+state
	val     T
	expires time.Time
	loading bool
	err     error
	inuse   bool // true once inserted in table (occupied)
}

func New[T any](capacity int) *Perch[T] {
	if capacity <= 0 {
		panic("capacity must be > 0")
	}
	c := &Perch[T]{
		mu:    &sync.Mutex{},
		cap:   capacity,
		table: make(map[string]uint32, capacity*2),
		slots: make([]entry[T], capacity+1), // 1-based
	}
	// build freelist [1..capacity]
	for i := 1; i <= capacity; i++ {
		c.slots[i].next = uint32(c.free)
		c.free = uint32(i)

		// prepare per-entry Cond without allocs
		c.slots[i].mu = &sync.Mutex{}
		c.slots[i].cv = &sync.Cond{L: c.slots[i].mu}
	}
	return c
}

// Get returns cached value if present+fresh. Otherwise calls loader once per key,
// caches result with TTL > 0, and returns it. ttl<=0 means "do not cache".
func (c *Perch[T]) Get(ctx context.Context, key string, ttl time.Duration, loader Loader[T]) (T, error) {
	var zero T
	now := time.Now() // cache this for perf

	// if the provided ttl is 0, we should not cache
	if ttl <= 0 {
		// do not do anything - just call the loader and get out
		// the rest of this logic does a lot of things to effectively manage contention
		// but we don't need to do that if we're not caching
		return loader(ctx, key)
	}

	// Fast path: lookup under global lock.
	c.mu.Lock()
	if idx := c.table[key]; idx != 0 {
		e := &c.slots[idx]
		// Take entry lock to check freshness or wait if loading.
		e.mu.Lock()
		c.mu.Unlock()

		// If another goroutine is loading this key, wait.
		for e.loading {
			e.cv.Wait()
		}
		// Fresh?
		if e.inuse && !e.expires.IsZero() && now.Before(e.expires) {
			// copy under e.mu, then release before taking c.mu
			v := e.val
			e.mu.Unlock()

			// bump MRU safely (don’t hold e.mu here)
			c.mu.Lock()
			if c.table[key] == idx { // still same slot?
				c.moveToFront(idx)
			}
			c.mu.Unlock()

			fmt.Println("hit cache")

			return v, nil
		}

		// Stale: we'll (re)load below.
		e.loading = true
		e.err = nil
		e.mu.Unlock()
		// proceed to load with this existing slot index
		return c.loadInto(ctx, idx, key, ttl, loader)
	}
	// Miss: need a slot (either free or evict LRU).
	var idx uint32
	if c.free != 0 {
		idx = c.free
		c.free = c.slots[idx].next
	} else {
		// evict tail
		idx = c.tail
		if idx == 0 {
			// shouldn't happen (cap>0)
			c.mu.Unlock()
			return zero, errors.New("lru: no slot available")
		}
		et := &c.slots[idx]
		// unlink from list
		c.unlink(idx)
		// remove old key from table
		delete(c.table, et.key)
		et.key = ""
		et.inuse = false
	}
	e := &c.slots[idx]
	// prepare this slot for the new key
	e.mu.Lock()
	e.key = key
	e.loading = true
	e.err = nil
	e.inuse = true
	// insert at head (MRU) now; if loader fails we’ll clear in place.
	c.linkFront(idx)
	c.table[key] = idx

	// Unlock
	e.mu.Unlock()
	c.mu.Unlock()

	// Load outside of global lock.
	return c.loadInto(ctx, idx, key, ttl, loader)
}

// loadInto performs the loader call for a specific slot index and signals waiters.
func (c *Perch[T]) loadInto(ctx context.Context, idx uint32, key string, ttl time.Duration, loader Loader[T]) (T, error) {
	e := &c.slots[idx]

	// wrap the loader - so that we can intercept panics if they happen
	wrappedLoader := func(ctx context.Context, key string) (t T, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("loader panicked: %v", r)
			}
		}()
		return loader(ctx, key)
	}

	// do the load without holding global lock
	val, err := wrappedLoader(ctx, key)

	now := time.Now()

	e.mu.Lock()
	if err != nil || ttl <= 0 {
		// On error or no-cache request, clear expiry to mark as uncached.
		e.expires = time.Time{}
		e.val = *new(T) // zero it
		e.err = err
	} else {
		e.val = val
		e.expires = now.Add(ttl)
		e.err = nil
	}
	e.loading = false
	e.mu.Unlock()
	e.cv.Broadcast()

	// If error or ttl<=0, remove from cache map/LRU (but still woke waiters).
	if err != nil || ttl <= 0 {
		c.mu.Lock()
		// Ensure idx still maps to key (it might, unless evicted/raced).
		if c.table[key] == idx {
			delete(c.table, key)
			c.unlink(idx)
			// push to freelist
			e.key = ""
			e.inuse = false
			e.next = c.free
			c.free = idx
		}
		c.mu.Unlock()
		return *new(T), err
	}

	// Success with caching: bump to MRU (cheap if already at head).
	c.mu.Lock()
	if c.table[key] == idx {
		c.moveToFront(idx)
	}
	c.mu.Unlock()
	return val, nil
}

func (c *Perch[T]) Delete(key string) {
	c.mu.Lock()
	idx := c.table[key]
	if idx == 0 {
		c.mu.Unlock()
		return
	}
	delete(c.table, key)
	c.unlink(idx)
	e := &c.slots[idx]
	// free-list push under c.mu
	e.next = c.free
	c.free = idx
	c.mu.Unlock()

	// clean entry state under e.mu (no c.mu here)
	e.mu.Lock()
	e.key = ""
	e.inuse = false
	e.loading = false
	e.err = nil
	e.expires = time.Time{}
	var zero T
	e.val = zero
	e.mu.Unlock()
}

// Peek returns (value, true) only if cached and fresh at the time of call.
func (c *Perch[T]) Peek(key string) (T, bool) {
	now := time.Now()
	c.mu.Lock()
	if idx := c.table[key]; idx != 0 {
		e := &c.slots[idx]
		e.mu.Lock()
		c.mu.Unlock()
		fresh := e.inuse && !e.loading && !e.expires.IsZero() && now.Before(e.expires)
		if !fresh {
			e.mu.Unlock()
			var zero T
			return zero, false
		}
		v := e.val
		e.mu.Unlock()
		// (Optional) we could bump to MRU here; omitted to keep Peek read-only.
		return v, true
	}
	c.mu.Unlock()
	var zero T
	return zero, false
}

// --- intrusive LRU helpers (c.mu held) ---

func (c *Perch[T]) unlink(idx uint32) {
	e := &c.slots[idx]
	prev, next := e.prev, e.next
	if prev != 0 {
		c.slots[prev].next = next
	} else {
		c.head = next
	}
	if next != 0 {
		c.slots[next].prev = prev
	} else {
		c.tail = prev
	}
	e.prev, e.next = 0, 0
}

func (c *Perch[T]) linkFront(idx uint32) {
	e := &c.slots[idx]
	e.prev = 0
	e.next = c.head
	if c.head != 0 {
		c.slots[c.head].prev = idx
	}
	c.head = idx
	if c.tail == 0 {
		c.tail = idx
	}
}

func (c *Perch[T]) moveToFront(idx uint32) {
	if c.head == idx {
		return
	}
	c.unlink(idx)
	c.linkFront(idx)
}
