package hermes

import (
	"math/rand"
)

type CuckooFilter struct {
	buckets  []bucket
	count    uint64
	capacity uint64
	stats    *FilterStats
}

func NewCuckoo(capacity uint) *CuckooFilter {

	capacity = uint(nextPowerOfTwo(int(capacity)) / bucketSize)
	if capacity == 0 {
		capacity = 1
	}
	buckets := make([]bucket, capacity)
	for i := range buckets {
		buckets[i] = [bucketSize]byte{}
	}

	return &CuckooFilter{
		buckets:  buckets,
		count:    0,
		stats:    NewFilterStats(),
		capacity: uint64(capacity),
	}
}

func (cf *CuckooFilter) reset() {
	buckets := make([]bucket, cf.capacity)
	for i := range buckets {
		buckets[i] = [bucketSize]byte{}
	}
	cf.buckets = buckets
	cf.stats = NewFilterStats()
	cf.count = 0
}

func (cf *CuckooFilter) contains(data []byte) bool {
	i1, i2, fp := getFilterComponents(data, uint(len(cf.buckets)))
	b1, b2 := cf.buckets[i1], cf.buckets[i2]
	if b1.getIndex(fp) > -1 {
		cf.stats.hit()
		return true
	}

	if b2.getIndex(fp) > -1 {
		cf.stats.hit()
		return true
	}

	cf.stats.miss()

	return false
}

func (cf *CuckooFilter) getStats() *FilterStats {
	return cf.stats
}

func (cf *CuckooFilter) add(data []byte) bool {
	i1, i2, fp := getFilterComponents(data, uint(len(cf.buckets)))
	if cf.buckets[i1].add(fp) || cf.buckets[i1].add(fp) {
		cf.count++
		return true
	}

	ri := randi(i1, i2)

	// try it until a certain step to check as many buckets as possible
	for k := 0; k < maxCuckooCount; k++ {
		j := rand.Intn(bucketSize)
		oldfp := fp
		fp = cf.buckets[ri][j]
		cf.buckets[ri][j] = oldfp

		rndIdx := getAltIndex(fp, ri, uint(len(cf.buckets)))
		if cf.buckets[rndIdx].add(fp) {
			cf.count++
			return true
		}
	}

	return false
}

func (cf *CuckooFilter) addUnique(data []byte) bool {
	if cf.contains(data) {
		return false
	}
	return cf.add(data)
}

func (cf *CuckooFilter) delete(data []byte) bool {
	i1, i2, fp := getFilterComponents(data, uint(len(cf.buckets)))
	if cf.buckets[i1].delete(fp) || cf.buckets[i2].delete(fp) {
		cf.count--
		return true
	}
	return false
}

func (cf *CuckooFilter) len() uint64 {
	return cf.count
}
