package hamt

import "container/list"

type root struct {
	maxTableDepth uint
	slotCount     uint
	mask          uint64
	slots         []node
}

func (r *root) delete(key uint64) (err error) {

	hc := key
	ndx := hc & r.mask
	if r.slots[ndx] == nil {
		err = notFound
	}
	if err == nil {
		node := r.slots[ndx]
		if node.IsLeaf() {
			myLeaf := node.(*leaf)
			myKey := myLeaf.key
			searchKey := key
			if searchKey == myKey {
				r.slots[ndx] = nil
			} else {
				err = notFound
			}
		} else {
			if 1 > r.maxTableDepth {
				err = notFound
			} else {
				tDeeper := node.(*table)
				hc >>= fanoutlog2
				err = tDeeper.delete(hc, 1, key)
			}
		}
	}
	return
}

func (r root) get(key uint64) (value *list.Element, err error) {

	hc := key
	ndx := hc & r.mask
	p := &r.slots
	if (*p)[ndx] != nil {
		node := (*p)[ndx]
		if node.IsLeaf() {
			myLeaf := node.(*leaf)
			myKey := myLeaf.key
			searchKey := key
			if searchKey == myKey {
				value = myLeaf.value
			} else {
				value = nil
			}
		} else {
			if 1 <= r.maxTableDepth {
				tDeeper := node.(*table)
				hc >>= fanoutlog2
				value, err = tDeeper.get(hc, 1, key)
			}
		}
	}
	return
}

func (r *root) set(l *leaf) (err error) {

	newHC := l.key
	slotNbr := uint(newHC & r.mask)

	p := &r.slots
	if (*p)[slotNbr] == nil {
		(*p)[slotNbr] = l
	} else {
		node := (*p)[slotNbr]
		if node.IsLeaf() {
			oldLeaf := node.(*leaf)
			curKey := oldLeaf.key
			newKey := l.key
			if curKey == newKey {
				oldLeaf.value = l.value
			} else {
				var tableDeeper *table
				tableDeeper, err = NewTable(1, r, oldLeaf)
				if err == nil {
					if 1 > r.maxTableDepth {
						err = maxTableDepthExceeded
					} else {
						newHC >>= fanoutlog2
						err = tableDeeper.set(newHC, 1, l)
						if err == nil {
							(*p)[slotNbr] = tableDeeper
						}
					}
				}
			}
		} else {
			if 1 > r.maxTableDepth {
				err = maxTableDepthExceeded
			} else {
				tDeeper := node.(*table)
				newHC >>= fanoutlog2
				err = tDeeper.set(newHC, 1, l)
			}
		}
	}
	return
}
