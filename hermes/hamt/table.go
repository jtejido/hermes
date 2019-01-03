package hamt

import (
	"container/list"
	"errors"
	"fmt"
	"math/bits"
)

type table struct {
	mask   uint64
	bitmap uint64
	slots  []node
	root   *root
}

func NewTable(depth uint, r *root, firstLeaf *leaf) (t *table, err error) {

	if r == nil {
		err = nilRoot
		return
	}

	tbl := new(table)
	tbl.root = r
	wFlag := uint64(1 << fanoutlog2)
	tbl.mask = wFlag - 1
	shiftCount := fanoutlog2 + (depth-1)*fanoutlog2
	hc := firstLeaf.key >> shiftCount
	ndx := hc & tbl.mask
	flag := uint64(1 << ndx)
	tbl.slots = []node{firstLeaf}
	tbl.bitmap = flag
	if err == nil {
		t = tbl
	}

	return
}

func (t *table) removeFromSlices(offset uint) (err error) {
	curSize := uint(len(t.slots))
	if curSize == 0 {
		err = deleteFromEmptyTable
	} else if offset >= curSize {
		err = errors.New(fmt.Sprintf(
			"InternalError: delete offset %d but table size %d\n",
			offset, curSize))
	} else if curSize == 1 {
		t.slots = t.slots[0:0]
	} else if offset == 0 {
		t.slots = t.slots[1:]
	} else if offset == curSize-1 {
		t.slots = t.slots[0:offset]
	} else {
		shorterSlots := make([]node, curSize-1)
		copy(shorterSlots[0:offset], t.slots[0:offset])
		copy(shorterSlots[offset:], t.slots[offset+1:])
		t.slots = shorterSlots
	}
	return
}

func (t *table) delete(hc uint64, depth uint, key uint64) (err error) {

	if len(t.slots) == 0 {
		err = notFound
	} else {
		ndx := hc & t.mask
		flag := uint64(1 << ndx)
		mask := flag - 1
		if t.bitmap&flag == 0 {
			err = notFound
		} else {
			var slotNbr uint
			if mask != 0 {
				slotNbr = uint(bits.OnesCount64(t.bitmap & mask))
			}
			node := t.slots[slotNbr]
			if node.IsLeaf() {
				myLeaf := node.(*leaf)
				myKey := myLeaf.key
				searchKey := key
				if searchKey == myKey {
					err = t.removeFromSlices(slotNbr)
					t.bitmap &= ^flag
				} else {
					err = notFound
				}
			} else {
				depth++
				if depth > t.root.maxTableDepth {
					err = notFound
				} else {
					tDeeper := node.(*table)
					hc >>= fanoutlog2
					err = tDeeper.delete(hc, depth, key)
				}
			}
		}
	}
	return
}

func (t table) get(hc uint64, depth uint, key uint64) (value *list.Element, err error) {

	ndx := hc & t.mask
	flag := uint64(1 << ndx)

	if t.bitmap&flag != 0 {
		var slotNbr uint
		mask := flag - 1
		if mask != 0 {
			slotNbr = uint(bits.OnesCount64(t.bitmap & mask))
		}
		node := t.slots[slotNbr]
		if node.IsLeaf() {
			myLeaf := node.(*leaf)
			myKey := myLeaf.key
			searchKey := key
			if searchKey == myKey {
				value = myLeaf.value
			}
		} else {
			depth++
			if depth <= t.root.maxTableDepth {
				tDeeper := node.(*table)
				hc >>= fanoutlog2
				value, err = tDeeper.get(hc, depth, key)
			}
		}
	}
	return
}

func (t *table) set(hc uint64, depth uint, l *leaf) (err error) {

	var slotNbr uint
	ndx := hc & t.mask
	flag := uint64(1 << ndx)
	mask := flag - 1
	if mask != 0 {
		slotNbr = uint(bits.OnesCount64(t.bitmap & mask))
	}
	sliceSize := uint(len(t.slots))
	if sliceSize == 0 {
		t.slots = []node{l}
		t.bitmap |= flag
	} else {
		if t.bitmap&flag != 0 {
			entry := t.slots[slotNbr]

			if entry.IsLeaf() {
				curLeaf := entry.(*leaf)
				curKey := curLeaf.key
				newKey := l.key
				if curKey == newKey {
					curLeaf.value = l.value
				} else {
					var (
						tableDeeper *table
					)
					depth++
					if depth > t.root.maxTableDepth {
						err = maxTableDepthExceeded
					} else {
						oldLeaf := entry.(*leaf)
						tableDeeper, err = NewTable(depth, t.root, oldLeaf)
						if err == nil {
							hc >>= fanoutlog2
							err = tableDeeper.set(hc, depth, l)
							if err == nil {
								t.slots[slotNbr] = tableDeeper
							}
						}
					}
				}
			} else {
				depth++
				if depth > t.root.maxTableDepth {
					err = maxTableDepthExceeded
				} else {
					tDeeper := entry.(*table)
					hc >>= fanoutlog2
					err = tDeeper.set(hc, depth, l)
				}
			}
		} else if slotNbr == 0 {
			leftSlots := make([]node, sliceSize+1)
			leftSlots[0] = l
			copy(leftSlots[1:], t.slots[:])
			t.slots = leftSlots
			t.bitmap |= flag
		} else if slotNbr == sliceSize {
			t.slots = append(t.slots, l)
			t.bitmap |= flag
		} else {
			leftSlots := make([]node, sliceSize+1)
			copy(leftSlots[:slotNbr], t.slots[:slotNbr])
			leftSlots[slotNbr] = l
			copy(leftSlots[slotNbr+1:], t.slots[slotNbr:])
			t.slots = leftSlots
			t.bitmap |= flag
		}
	}
	return
}

func (t table) IsLeaf() bool { return false }
