package hermes

import (
	"sync"
	"time"
)

type shards []*Shard

type Shard struct {
	id      int
	size    int64
	maxSize int64
	policy  *LRFU
	stats   *Stats
	sync.RWMutex
}

func (s *Shard) getStats() *Stats {
	return s.stats.getStats()
}

func (s *Shard) get(strKey string) ([]byte, error) {

	if s.policy == nil {
		return nil, errorf(policyNotInitializedError, nil)
	}

	k := getIndexKey(strKey)

	item, ok_l := s.policy.Get(k)

	if !ok_l {
		s.stats.miss()
		return nil, errorf(keyNotFoundInShardError, strKey, s.id)
	}

	if entryKey := getKeyFromEntry(item); strKey != entryKey {
		s.stats.collision()
		return nil, errorf(collisionDetectedError, strKey, s.id)
	}

	s.stats.hit()

	return getValueFromEntry(item), nil
}

func (s *Shard) set(strKey string, data []byte) error {

	currentTimestamp := uint64(time.Now().Unix())

	if s.policy == nil {
		return errorf(policyNotInitializedError, nil)
	}

	k := getIndexKey(strKey)

	v := wrapEntry(currentTimestamp, k, strKey, data)

	s.size += int64(hashSizeInBytes) + int64(len(data))

	if s.size > s.maxSize {
		s.policy.RemoveElement()
	}

	s.policy.Set(k, v)

	return nil
}

func (s *Shard) delete(strKey string) error {

	if s.policy == nil {
		return errorf(policyNotInitializedError, nil)
	}

	k := getIndexKey(strKey)

	if !s.policy.Remove(k) {
		s.stats.delmiss()
		return errorf(keyNotFoundInShardError, strKey, s.id)
	}

	s.stats.delhit()

	return nil
}

func (s *Shard) clear() {
	s.policy.Clear()
	s.size = 0
	s.stats = NewStats()
}
