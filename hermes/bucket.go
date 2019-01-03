package hermes

type bucket [4]byte

func (b *bucket) add(fp byte) bool {
	for i, tfp := range b {
		if tfp == nullFp {
			b[i] = fp
			return true
		}
	}
	return false
}

func (b *bucket) delete(fp byte) bool {
	for i, tfp := range b {
		if tfp == fp {
			b[i] = nullFp
			return true
		}
	}
	return false
}

func (b *bucket) getIndex(fp byte) int {
	for i, tfp := range b {
		if tfp == fp {
			return i
		}
	}
	return -1
}
