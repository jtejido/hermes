package hermes

// All policies implemented (or wish to be implemented) for hermes follows this interface.
type Policy interface {
	Set(uint64, []byte)
	Get(uint64) ([]byte, bool)
	Len() int
	Remove(uint64) bool
	Clear()
	RemoveElement()
	SetEvictedFunc(func(key uint64, value []byte))
}

// General interface for k-v struct used in any policies you wish to implement
type Entry interface {
	GetKey() uint64
	GetValue() []byte
}
