package hamt

import (
	"container/list"
	"errors"
)

var (
	deleteFromEmptyTable  = errors.New("Internal Error: delete from empty table")
	maxTableDepthExceeded = errors.New("max Table depth exceeded")
	nilRoot               = errors.New("nil root parameter")
	nilValue              = errors.New("nil value parameter")
	notFound              = errors.New("entry not found")
)

const (
	fanoutlog2 = 6
)

// The HAMT implemented here strictly enforces storing *list.Element for the policies.
// It uses up to 2^w for table size and 2^t entries for the root where w=6 and t=6
// When implementing it in a custom policy, see lrfu for its usage.
type HAMT struct {
	root *root
}

func New() *HAMT {
	flag := uint64(1)
	flag <<= fanoutlog2
	count := uint(1 << fanoutlog2)

	r := &root{
		maxTableDepth: (64 - fanoutlog2) / fanoutlog2,
		slotCount:     count,
		mask:          flag - 1,
		slots:         make([]node, count),
	}

	return &HAMT{
		root: r,
	}
}

func (h *HAMT) Delete(k uint64) error {
	return h.root.delete(k)
}

func (h HAMT) Get(k uint64) (*list.Element, error) {
	return h.root.get(k)
}

func (h *HAMT) Set(k uint64, v *list.Element) (err error) {
	leaf, err := NewLeaf(k, v)
	if err == nil {
		err = h.root.set(leaf)
	}
	return
}
