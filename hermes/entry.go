package hermes

// internal type used for lrfu
type entry struct {
	key           uint64
	value         []byte
	lastReference float64
	lastCRF       float64
}

func (e entry) GetKey() uint64 {
	return e.key
}

func (e entry) GetValue() []byte {
	return e.value
}
