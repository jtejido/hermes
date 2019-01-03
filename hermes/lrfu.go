// Lee et al's "LRFU (Least Recently/Frequently Used) Replacement Policy: A Spectrum of Block Replacement Policies"
// This is modified (might as well call it mod-lrfu to avoid confusion) to remove the heap out of the picture.
// The operation here is the same as in the literature, but reversing the selector from the original 0...1 (lfu to lru) to (lru to lfu)
//

package hermes

import (
	"container/list"
	"github.com/jtejido/hermes/hermes/hamt"
	"math"
)

// This is not thread-safe, which means it will depend on the parent implementation to do the locking mechanism.
type LRFU struct {
	maxEntries int
	lambda     float64
	OnEvicted  func(key uint64, value []byte)
	ll         *list.List
	cache      *hamt.HAMT
	count      float64
	smallest   *list.Element
}

func NewLRFU(maxEntries int, lambda float64) *LRFU {

	return &LRFU{
		maxEntries: maxEntries,
		lambda:     lambda,
		ll:         list.New(),
		cache:      hamt.New(), // We use HAMT for a nearly O(1) yet memory efficient hashmap.
		count:      0.0,
	}
}

func (lru *LRFU) SetEvictedFunc(f func(key uint64, value []byte)) {
	lru.OnEvicted = f
}

func (lru *LRFU) Set(key uint64, value []byte) {

	lru.count += 1

	el, err := lru.cache.Get(key)

	if err == nil && el != nil {
		lru.ll.MoveToFront(el)
		el.Value.(*entry).lastCRF = lru.getWeight(0) + lru.getCRF(el.Value.(*entry))
		el.Value.(*entry).lastReference = lru.count
		el.Value.(*entry).value = value
		lru.restore(el)

		return
	}

	e := &entry{
		key:           key,
		value:         value,
		lastReference: lru.count,
		lastCRF:       lru.getWeight(0),
	}

	ele := lru.ll.PushFront(e)

	lru.cache.Set(key, ele)

	lru.restore(ele)

	if lru.maxEntries != 0 && lru.ll.Len() > lru.maxEntries {
		lru.RemoveElement()
	}

}

func (lru *LRFU) Get(key uint64) (value []byte, ok bool) {
	lru.count += 1

	if lru.cache == nil {
		return
	}

	el, err := lru.cache.Get(key)

	if err == nil && el != nil {

		lru.ll.MoveToFront(el)

		el.Value.(*entry).lastCRF = lru.getWeight(0) + lru.getCRF(el.Value.(*entry))

		el.Value.(*entry).lastReference = lru.count

		lru.restore(el)

		return el.Value.(*entry).value, true
	}

	return
}

func (lru *LRFU) restore(ele *list.Element) {

	if lru.smallest == nil {
		lru.smallest = ele
		return
	}

	fe := lru.ll.Front()

	en := ele.Value.(*entry)
	smallest := lru.smallest.Value.(*entry)
	if fe.Value.(*entry).key != en.key {
		if lru.getCRF(en) > lru.getCRF(smallest) {
			*en, *smallest = *smallest, *en
			lru.cache.Set(en.key, ele)
			lru.cache.Set(smallest.key, lru.smallest)
			lru.restore(lru.smallest)

		} else {
			lru.smallest = ele
		}

	}
	return
}

func (lru *LRFU) getCRF(en *entry) float64 {
	return lru.getWeight(lru.count-en.lastReference) * en.lastCRF
}

func (lru *LRFU) RemoveElement() {
	if lru.cache == nil {
		return
	}
	ele := lru.ll.Back()

	if ele != nil {
		lru.removeElement(ele)
	}
}

func (lru *LRFU) getWeight(v float64) float64 {
	return math.Pow((1 / 2), lru.lambda*v)
}

func (lru *LRFU) removeElement(e *list.Element) {

	if lru.smallest.Value.(*entry).key == e.Value.(*entry).key {
		lru.smallest = nil
	}

	lru.ll.Remove(e)
	kv := e.Value.(*entry)

	lru.cache.Delete(kv.key)

	if lru.OnEvicted != nil {
		lru.OnEvicted(kv.key, kv.value)
	}
}

func (lru *LRFU) Len() int {
	if lru.cache == nil {
		return 0
	}
	return lru.ll.Len()
}

func (lru *LRFU) Remove(key uint64) (ok bool) {
	if lru.cache == nil {
		return
	}

	el, err := lru.cache.Get(key)

	if err == nil && el != nil {

		lru.removeElement(el)

		return true
	}

	return false

}

func (lru *LRFU) Clear() {
	// we'll just reset it
	lru.ll = list.New()
	lru.cache = hamt.New()
	lru.count = 0.0
}
