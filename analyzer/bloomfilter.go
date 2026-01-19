package analyzer

import (
	"hash/fnv"
)

// BloomFilter is a probabilistic data structure for membership testing
type BloomFilter struct {
	bits      []bool
	size      uint
	hashCount uint
}

// NewBloomFilter creates a new Bloom filter
func NewBloomFilter(size uint, hashCount uint) *BloomFilter {
	return &BloomFilter{
		bits:      make([]bool, size),
		size:      size,
		hashCount: hashCount,
	}
}

// Add inserts an item into the Bloom filter
func (bf *BloomFilter) Add(item string) {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		bf.bits[hash%bf.size] = true
	}
}

// Contains checks if an item might be in the set
func (bf *BloomFilter) Contains(item string) bool {
	for i := uint(0); i < bf.hashCount; i++ {
		hash := bf.hash(item, i)
		if !bf.bits[hash%bf.size] {
			return false
		}
	}
	return true
}

// hash generates a hash value for an item with a seed
func (bf *BloomFilter) hash(item string, seed uint) uint {
	h := fnv.New64a()
	h.Write([]byte(item))
	h.Write([]byte{byte(seed)})
	return uint(h.Sum64())
}

// Clear resets the Bloom filter
func (bf *BloomFilter) Clear() {
	for i := range bf.bits {
		bf.bits[i] = false
	}
}
