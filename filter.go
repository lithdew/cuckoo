package cuckoo

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"reflect"
	"unsafe"
)

const (
	// BucketSize is the fingerprint size of each bucket a single Cuckoo filter instance holds.
	BucketSize = 4

	// NumBuckets is the total number of buckets a single Cuckoo filter instance holds.
	NumBuckets = 524288

	// MaxInsertionAttempts is the max number of times a fingerprint of an ID will attempt to be inserted
	// into a Cuckoo filter instance.
	MaxInsertionAttempts = 500
)

// Hash is a 32-byte hash input to Filter.
type Hash [32]byte

// Filter represents a Cuckoo filter whose inputs are pre-hashed 32 byte arrays.
type Filter struct {
	Buckets [NumBuckets]Bucket
	Count   uint

	unsafe bool
}

// UnsafeUnmarshalBinary is roughly 1.7x faster than UnmarshalBinary, but in return does not provide an accurate
// cardinality estimate of the items in the filter. It returns a error should decoding fail.
func UnsafeUnmarshalBinary(buf []byte) (*Filter, error) {
	if len(buf) != NumBuckets*BucketSize {
		return nil, fmt.Errorf("must be %d bytes, but got %d bytes", NumBuckets*BucketSize, len(buf))
	}

	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data
	buckets := *(*[NumBuckets]Bucket)(unsafe.Pointer(ptr))

	return &Filter{Buckets: buckets, unsafe: true}, nil
}

// UnmarshalBinary decodes the bytes from buf into a Filter instance, and returns an error should it fail.
func UnmarshalBinary(buf []byte) (*Filter, error) {
	if len(buf) != NumBuckets*BucketSize {
		return nil, fmt.Errorf("must be %d bytes, but got %d bytes", NumBuckets*BucketSize, len(buf))
	}

	count := CountNonzeroBytes(buf)

	ptr := (*reflect.SliceHeader)(unsafe.Pointer(&buf)).Data
	buckets := *(*[NumBuckets]Bucket)(unsafe.Pointer(ptr))

	return &Filter{Buckets: buckets, Count: count}, nil
}

// MarshalBinary marshals a Cuckoo filter into its byte representation.
func (f *Filter) MarshalBinary() []byte {
	if f.unsafe {
		panic("attempted to re-marshal an unsafely unmarshaled cuckoo filter")
	}

	var buf []byte

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	sh.Data = uintptr(unsafe.Pointer(&f.Buckets))
	sh.Len = NumBuckets * BucketSize
	sh.Cap = NumBuckets * BucketSize

	out := make([]byte, NumBuckets*BucketSize)
	copy(out, buf)

	return out
}

// NewFilter instantiates a new Cuckoo filter.
func NewFilter() *Filter {
	return &Filter{}
}

// Reset resets this Cuckoo filter to its zero value.
func (f *Filter) Reset() {
	f.Buckets = [NumBuckets]Bucket{}
	f.Count = 0
}

// Insert inserts a 32-byte pre-hashed array of bytes into a Cuckoo filter. It returns true if the id is
// inserted successfully into the filter, and false otherwise. Each hashing attempt for this Cuckoo filter
// is done using a 32-bit Jenkins hash. This function assumes that id is pre-hashed with a collision-resistant
// hash function, such as SHA-256 or BLAKE2b or Highway Hash.
func (f *Filter) Insert(id Hash) bool {
	val, a, b := process(id)

	// Assert that the ID has not been inserted into the filter before.
	if f.Buckets[a].IndexOf(val) > -1 || f.Buckets[b].IndexOf(val) > -1 {
		return false
	}

	// Attempt to insert into bucket A.
	if f.Buckets[a].Insert(val) {
		f.Count++
		return true
	}

	// Attempt to insert into bucket B .
	if f.Buckets[b].Insert(val) {
		f.Count++
		return true
	}

	i := a
	if rand.Intn(2) == 0 {
		i = b
	}

	for attempt := 0; attempt < MaxInsertionAttempts; attempt++ {
		j := rand.Intn(BucketSize)

		val, f.Buckets[i][j] = f.Buckets[i][j], val

		i = (i ^ jenkins(uint(val))) % NumBuckets

		if f.Buckets[i].Insert(val) {
			f.Count++
			return true
		}
	}

	return false
}

// Delete attempts to remove the fingerprint of id from the Cuckoo filter.
func (f *Filter) Delete(id Hash) bool {
	val, a, b := process(id)

	if f.Buckets[a].Delete(val) {
		f.Count--
		return true
	}

	if f.Buckets[b].Delete(val) {
		f.Count--
		return true
	}

	return false
}

// Lookup returns true if the Cuckoo filter contains id.
func (f *Filter) Lookup(id Hash) bool {
	val, a, b := process(id)
	return f.Buckets[a].IndexOf(val) > -1 || f.Buckets[b].IndexOf(val) > -1
}

func process(id Hash) (byte, uint, uint) {
	val := byte(binary.BigEndian.Uint64(id[0:8])%255 + 1)
	a := uint(binary.BigEndian.Uint64(id[8:16])) % NumBuckets
	b := (a ^ jenkins(uint(val))) % NumBuckets

	return val, a, b
}

// Bucket is a single Cuckoo filter bucket containing bytes representative of 7-bit fingerprints of hashed inputs.
type Bucket [BucketSize]byte

// Insert attempts to insert a fingerprint of an input into the bucket. It returns true if the operation is
// successful, and false otherwise.
func (b *Bucket) Insert(val byte) bool {
	for i, stored := range b {
		if stored == 0 {
			b[i] = val
			return true
		}
	}

	return false
}

// Delete attempts to remove a fingerprint of an input from the bucket. It returns true if the operation is
// successful, and false otherwise.
func (b *Bucket) Delete(val byte) bool {
	for i, stored := range b {
		if stored == val {
			b[i] = 0
			return true
		}
	}

	return false
}

// IndexOf returns the index in which a provided byte is equal to val within the bucket.
func (b *Bucket) IndexOf(val byte) int {
	for i, stored := range b {
		if stored == val {
			return i
		}
	}

	return -1
}

func jenkins(a uint) uint {
	a = (a + 0x7ed55d16) + (a << 12)
	a = (a ^ 0xc761c23c) ^ (a >> 19)
	a = (a + 0x165667b1) + (a << 5)
	a = (a + 0xd3a2646c) ^ (a << 9)
	a = (a + 0xfd7046c5) + (a << 3)
	a = (a ^ 0xb55a4f09) ^ (a >> 16)

	return a
}
