package hamt

import "container/list"

type leaf struct {
	key   uint64
	value *list.Element
}

func NewLeaf(key uint64, value *list.Element) (l *leaf, err error) {
	if value == nil {
		err = nilValue
	} else {
		l = &leaf{
			key:   key,
			value: value,
		}
	}
	return
}

func (l leaf) IsLeaf() bool { return true }
