package hermes

import (
	"sync/atomic"
)

// Pretty basic
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	DelHits    int64 `json:"delete_hits"`
	DelMisses  int64 `json:"delete_misses"`
	Collisions int64 `json:"collisions"`
}

func NewStats() *Stats {
	return new(Stats)
}

func (s *Stats) getStats() *Stats {
	return s
}

func (s *Stats) hit() {
	atomic.AddInt64(&s.Hits, 1)
}

func (s *Stats) miss() {
	atomic.AddInt64(&s.Misses, 1)
}

func (s *Stats) delhit() {
	atomic.AddInt64(&s.DelHits, 1)
}

func (s *Stats) delmiss() {
	atomic.AddInt64(&s.DelMisses, 1)
}

func (s *Stats) collision() {
	atomic.AddInt64(&s.Collisions, 1)
}

type FilterStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
}

func NewFilterStats() *FilterStats {
	return new(FilterStats)
}

func (s *FilterStats) getStats() *FilterStats {
	return s
}

func (s *FilterStats) hit() {
	atomic.AddInt64(&s.Hits, 1)
}

func (s *FilterStats) miss() {
	atomic.AddInt64(&s.Misses, 1)
}
