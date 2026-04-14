package main

import (
	"hash/crc32"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
)

// Algorithm selects a backend. req is passed for L7-aware routing.
type Algorithm interface {
	NextBackend(r *Registry, req *http.Request) *Backend
}

// sessionKey builds the L7 composite key used for sticky routing and hashing.
func sessionKey(req *http.Request) string {
	vu := req.Header.Get("X-Client-VU")
	session := req.Header.Get("X-Session-ID")
	path := req.URL.Path
	role := req.Header.Get("X-Role")
	// Prefer explicit session ID, fall back to VU
	id := session
	if id == "" {
		id = vu
	}
	if id == "" {
		id = req.RemoteAddr
	}
	return path + ":" + role + ":" + id
}

// ---- Round Robin ----

type RoundRobin struct {
	counter uint64
}

func (rr *RoundRobin) NextBackend(r *Registry, req *http.Request) *Backend {
	pool := r.MatchSubset(req.URL.Path, req.Header.Get("X-Role"))
	if pool == nil {
		pool = r.GetHealthyBackends()
	}
	if len(pool) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&rr.counter, 1) % uint64(len(pool))
	return pool[idx]
}

// ---- Least Requests (with session stickiness) ----

type LeastRequests struct{}

func (lr *LeastRequests) NextBackend(r *Registry, req *http.Request) *Backend {
	key := sessionKey(req)

	// Check sticky map first
	if b := r.StickyGet(key); b != nil {
		return b
	}

	pool := r.MatchSubset(req.URL.Path, req.Header.Get("X-Role"))
	if pool == nil {
		pool = r.GetHealthyBackends()
	}
	if len(pool) == 0 {
		return nil
	}

	best := pool[0]
	minReq := atomic.LoadInt32(&best.Stats.ActiveRequests)
	for i := 1; i < len(pool); i++ {
		b := pool[i]
		reqs := atomic.LoadInt32(&b.Stats.ActiveRequests)
		if reqs < minReq {
			minReq = reqs
			best = b
		}
	}

	// Record sticky mapping (must be replicated by Raft/Controller team across LBs)
	r.StickySet(key, best.ID)
	return best
}

// ---- Weighted Round Robin (with session stickiness) ----

type WeightedRoundRobin struct {
	mu sync.Mutex
}

func (wrr *WeightedRoundRobin) NextBackend(r *Registry, req *http.Request) *Backend {
	key := sessionKey(req)

	// Check sticky map first
	if b := r.StickyGet(key); b != nil {
		return b
	}

	pool := r.MatchSubset(req.URL.Path, req.Header.Get("X-Role"))
	if pool == nil {
		pool = r.GetHealthyBackends()
	}
	if len(pool) == 0 {
		return nil
	}

	totalWeight := 0.0
	for _, b := range pool {
		totalWeight += b.Stats.Weight
	}

	wrr.mu.Lock()
	point := rand.Float64() * totalWeight
	wrr.mu.Unlock()

	cur := 0.0
	var best *Backend
	for _, b := range pool {
		cur += b.Stats.Weight
		if point <= cur {
			best = b
			break
		}
	}
	if best == nil {
		best = pool[0]
	}

	// Record sticky mapping (must be replicated by Raft/Controller team across LBs)
	r.StickySet(key, best.ID)
	return best
}

// ---- Maglev Consistent Hashing ----

type Maglev struct {
	mu          sync.RWMutex
	epoch       uint64
	lookupTable []*Backend
	M           int
}

func NewMaglev() *Maglev {
	return &Maglev{M: 251}
}

func mHash(s string) uint32 {
	return crc32.ChecksumIEEE([]byte(s))
}

func (m *Maglev) buildTable(r *Registry) {
	healthy := r.GetHealthyBackends()
	table := make([]*Backend, m.M)
	if len(healthy) == 0 {
		m.lookupTable = table
		return
	}

	next := make([]int, len(healthy))
	perms := make([][]int, len(healthy))
	for i, b := range healthy {
		offset := int(mHash(b.ID+"_1") % uint32(m.M))
		skip := int((mHash(b.ID+"_2")%uint32(m.M-1)) + 1)
		perm := make([]int, m.M)
		for j := 0; j < m.M; j++ {
			perm[j] = (offset + j*skip) % m.M
		}
		perms[i] = perm
	}

	n := 0
	for n < m.M {
		for i, b := range healthy {
			c := perms[i][next[i]]
			for table[c] != nil {
				next[i]++
				c = perms[i][next[i]]
			}
			table[c] = b
			next[i]++
			n++
			if n >= m.M {
				break
			}
		}
	}
	m.lookupTable = table
}

func (m *Maglev) NextBackend(r *Registry, req *http.Request) *Backend {
	m.mu.RLock()
	curEpoch := m.epoch
	tbl := m.lookupTable
	m.mu.RUnlock()

	regEpoch := r.GetEpoch()
	if tbl == nil || curEpoch != regEpoch {
		m.mu.Lock()
		if m.epoch != regEpoch {
			m.buildTable(r)
			m.epoch = regEpoch
		}
		tbl = m.lookupTable
		m.mu.Unlock()
	}

	if len(tbl) == 0 {
		return nil
	}

	// L7-aware composite key: path + role + session/VU — gives natural affinity per chat/session
	key := sessionKey(req)
	idx := mHash(key) % uint32(m.M)
	return tbl[idx]
}
