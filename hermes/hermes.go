package hermes

import (
	"fmt"
	"github.com/jtejido/hermes/config"
	pb "github.com/jtejido/hermes/hermespb"
	"reflect"
	"sync"
)

var (
	mu                 sync.RWMutex
	cache              *Cache
	initPeerServerOnce sync.Once
	initPeerServer     func()
)

type Getter interface {
	Get(ctx Context, key string) ([]byte, error)
}

// Function to be used as data loader
type GetterFunc func(ctx Context, key string) ([]byte, error)

func (f GetterFunc) Get(ctx Context, key string) ([]byte, error) {
	return f(ctx, key)
}

type Cache struct {
	shards    shards
	mask      uint64
	filter    *CuckooFilter
	peersOnce sync.Once
	peers     PeerPicker
	sync.RWMutex
}

func NewCache(config *config.Config) *Cache {
	return newCache(config, nil)
}

// Returns a New Hermes Cache instance
func newCache(config *config.Config, peers PeerPicker) *Cache {
	initPeerServerOnce.Do(callInitPeerServer)
	size := mBToBytes(nextPowerOfTwo(config.Cache.Size))
	c := new(Cache)
	c.shards = make(shards, config.Cache.ShardCount)
	c.mask = uint64(config.Cache.ShardCount - 1)
	c.peers = peers

	if config.Filter.Enabled {
		c.filter = NewCuckoo(config.Filter.FilterItemCount)
	}

	for i := 0; i < config.Cache.ShardCount; i++ {

		c.shards[i] = &Shard{
			id:      i,
			size:    0,
			maxSize: int64(size / config.Cache.ShardCount),
			policy:  NewLRFU(size/config.Cache.ShardCount, config.Cache.Lambda),
			stats:   NewStats(),
		}

		func(i int) {
			c.shards[i].policy.OnEvicted = func(key uint64, value []byte) {
				// delete key from filter if enabled

				if c.filter != nil {
					c.filter.delete([]byte(getKeyFromEntry(value)))
				}

				if c.peers != nil {
					c.peers.DecrementLoad()
				}

				c.shards[i].size -= int64(hashSizeInBytes) + int64(len(getValueFromEntry(value)))
			}
		}(i)
	}

	cache = c
	return c
}

func (c *Cache) initPeers() {
	if c.peers == nil {
		c.peers = getPeers()
	}
}

// Sets function to be used as data loader
func (c *Cache) Peers(peers PeerPicker) *Cache {
	c.peers = peers
	return c
}

// Returns the data from a given key
func (c *Cache) Get(ctx Context, key string) ([]byte, error) {

	c.peersOnce.Do(c.initPeers)

	if c.shards == nil {
		return nil, errorf(shardsNotInitializedError, nil)
	}

	shard, err := c.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	if err != nil {
		return nil, err
	}

	item, err_i := shard.get(key)

	if err_i != nil {
		if c.peers != nil {
			if peer, ok := c.peers.PickPeer(key); ok {

				value, err_p := c.getFromPeer(ctx, peer, key)

				if err_p != nil {
					return nil, err_p
				}

				return value, nil
			}
		}

		return nil, err_i
	}

	return item, nil
}

func (c *Cache) getFromPeer(ctx Context, peer ProtoGetter, key string) ([]byte, error) {

	req := &pb.GetRequest{
		Key: key,
	}

	res := &pb.GetResponse{}
	err := peer.Get(ctx, req, res)
	if err != nil {
		return nil, err
	}

	return res.Value, nil
}

// Sets the data with a given key
func (c *Cache) Set(ctx Context, key string, data []byte) error {

	c.peersOnce.Do(c.initPeers)

	if c.shards == nil {
		return errorf(shardsNotInitializedError, nil)
	}

	if c.peers != nil {
		if peer, ok := c.peers.PickPeer(key); ok {

			err_p := c.setToPeer(ctx, peer, key, data)

			if err_p != nil {
				return err_p
			}

			return nil
		}
	}

	shard, err := c.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	if err != nil {
		return err
	}

	if c.filter != nil {
		// filter is enabled. So we'll test first, add in filter if not there, then don't cache, assuming it's a one-hit-wonder
		if !c.filter.contains([]byte(key)) {
			c.filter.add([]byte(key))
			return errorf(filterFirstInstanceError, key)
		}
	}

	err_s := shard.set(key, data)

	if err_s != nil {
		return err_s
	}

	if c.peers != nil {
		c.peers.IncrementLoad()
	}

	return nil

}

func (c *Cache) setToPeer(ctx Context, peer ProtoGetter, key string, data []byte) error {

	req := &pb.SetRequest{
		Key:   key,
		Value: data,
	}

	res := &pb.SetResponse{}
	err := peer.Set(ctx, req, res)

	if err != nil {
		return err
	}

	if !reflect.DeepEqual(pb.SetResponse{}, *res) {
		return fmt.Errorf(res.Error.Message)
	}

	return nil
}

// Deletes a data given a key
func (c *Cache) Delete(ctx Context, key string) error {

	c.peersOnce.Do(c.initPeers)

	if c.shards == nil {
		return errorf(shardsNotInitializedError, nil)
	}

	if c.peers != nil {
		if peer, ok := c.peers.PickPeer(key); ok {

			err_p := c.deleteFromPeer(ctx, peer, key)

			if err_p != nil {
				return err_p
			}

			return nil
		}
	}

	shard, err := c.getShard(key)

	if err != nil {
		return err
	}

	shard.Lock()
	defer shard.Unlock()

	if c.filter != nil {
		c.filter.delete([]byte(key))
	}

	err_d := shard.delete(key)

	if err_d != nil {
		return err_d
	}

	if c.peers != nil {
		c.peers.DecrementLoad()
	}

	return nil
}

func (c *Cache) deleteFromPeer(ctx Context, peer ProtoGetter, key string) error {

	req := &pb.DeleteRequest{
		Key: key,
	}

	res := &pb.DeleteResponse{}
	err := peer.Delete(ctx, req, res)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(pb.DeleteResponse{}, *res) {
		return fmt.Errorf(res.Error.Message)
	}

	return nil
}

// Returns cache stats
func (c *Cache) GetStats() *Stats {
	s := Stats{}
	for _, shard := range c.shards {
		stat := shard.getStats()
		s.Hits += stat.Hits
		s.Misses += stat.Misses
		s.DelMisses += stat.DelMisses
		s.DelHits += stat.DelHits
		s.Collisions += stat.Collisions
	}
	return &s
}

// Returns cache stats
func (c *Cache) GetFilterStats() *FilterStats {
	if c.filter != nil {
		return c.filter.getStats()
	}

	return nil
}

// Resets filter
func (c *Cache) ResetFilter() {
	if c.filter != nil {
		c.filter.reset()
	}
}

// Returns the number of items
func (c *Cache) Len() int {
	var len int
	for _, shard := range c.shards {
		len += shard.policy.Len()
	}

	return len
}

// Returns the size in MB
func (c *Cache) Size() (size int64) {
	for _, shard := range c.shards {
		size += shard.size
	}
	return int64(bytesToMB(int(size)))
}

// Returns the maximum given size in MB
func (c *Cache) MaxSize() (maxSize int64) {
	for _, shard := range c.shards {
		maxSize += shard.maxSize
	}
	return int64(bytesToMB(int(maxSize)))
}

// Clears the cache's and filter's (if available) items
func (c *Cache) Clear() {
	for _, shard := range c.shards {
		shard.clear()
	}

	if c.filter != nil {
		c.filter.reset()
	}
}

// Checks the filter for the key's presence.
// Only appends the filter's stats.
func (c *Cache) Contains(key string) bool {
	if c.filter != nil {
		return c.filter.contains([]byte(key))
	}

	return false
}

// Counts items in the filter
func (c *Cache) FilterCount() uint64 {
	if c.filter != nil {
		return c.filter.len()
	}

	return 0
}

func (c *Cache) getShard(key string) (s *Shard, err error) {

	if c.shards == nil {
		return nil, errorf(shardsNotInitializedError, nil)
	}

	k := getIndexKey(key)

	c.Lock()
	defer c.Unlock()

	if (k&c.mask >= 0) && (k&c.mask <= uint64(len(c.shards))) {

		return c.shards[k&c.mask], nil
	}

	return nil, errorf(shardNotFoundForKeyError, key)

}

func callInitPeerServer() {
	if initPeerServer != nil {
		initPeerServer()
	}
}
