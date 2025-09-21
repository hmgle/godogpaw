package engine

const FILTER_MASK Key = (1 << 14) - 1

type BloomFilter struct {
	table [1 << 14]uint8
}

func (b *BloomFilter) Incr(key Key) {
	b.table[key&FILTER_MASK]++
}

func (b *BloomFilter) Decr(key Key) {
	b.table[key&FILTER_MASK]--
}

func (b *BloomFilter) Value(key Key) uint8 {
	return b.table[key&FILTER_MASK]
}

func (b *BloomFilter) Reset() {
	clear(b.table[:])
}
