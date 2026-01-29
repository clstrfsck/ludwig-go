// Handles a fixed-size string object with 1-based indexing.
// Calls to methods will panic if indices are out of range.

package ludwig

import (
	"bytes"
	"fmt"
	"math"
)

const (
	// MaxStrLen from const.go is used
	// MinIndex is the minimum index (1-based indexing)
	MinIndex = 1
	// MaxIndex is the maximum index
	MaxIndex = MaxStrLen
)

// StrObject represents a fixed-size string object with 1-based indexing
type StrObject struct {
	array [MaxStrLen]byte
}

// checkIndex validates that an index with optional offset is within valid range
func checkIndex(index, offset int) {
	if offset > math.MaxInt-index {
		panic("index + offset overflow")
	}

	combinedIndex := index + offset

	if combinedIndex < MinIndex || combinedIndex > MaxIndex {
		panic("index out of range")
	}
}

// adjustIndex converts 1-based index to 0-based array index
func adjustIndex(index, offset int) int {
	checkIndex(index, offset)
	return index + offset - MinIndex
}

// NewFilled creates a new StrObject filled with the given element
func NewFilled(elem byte) *StrObject {
	s := &StrObject{}
	for i := range s.array {
		s.array[i] = elem
	}
	return s
}

// NewWithPattern creates a new StrObject with a repeating pattern
func NewWithPattern(values []byte) *StrObject {
	if len(values) == 0 {
		return NewFilled(' ')
	}
	s := &StrObject{}
	for i := range s.array {
		s.array[i] = values[i%len(values)]
	}
	return s
}

// Clone creates a copy of the StrObject
func (s *StrObject) Clone() *StrObject {
	return &StrObject{array: s.array}
}

// Get returns the character at the given 1-based index
func (s *StrObject) Get(index int) byte {
	idx := adjustIndex(index, 0)
	return s.array[idx]
}

// Set sets the character at the given 1-based index
func (s *StrObject) Set(index int, value byte) {
	idx := adjustIndex(index, 0)
	s.array[idx] = value
}

// Assign sets the content from a string, padding with spaces if needed
func (s *StrObject) Assign(str string) {
	s.FillCopyBytes([]byte(str), 1, MaxStrLen, ' ')
}

// Equals compares n characters starting at srcOffset with another StrObject starting at dstOffset
func (s *StrObject) Equals(other *StrObject, n, srcOffset, dstOffset int) bool {
	if n == 0 {
		return true
	}
	checkIndex(srcOffset, n-1)
	checkIndex(dstOffset, n-1)
	srcIdx := adjustIndex(srcOffset, 0)
	dstIdx := adjustIndex(dstOffset, 0)
	return bytes.Equal(s.array[srcIdx:srcIdx+n], other.array[dstIdx:dstIdx+n])
}

// ApplyN applies a function to n characters starting at start
func (s *StrObject) ApplyN(f func(byte) byte, n, start int) {
	if n <= 0 {
		return
	}
	checkIndex(start, n-1)
	idx := adjustIndex(start, 0)
	for i := idx; i < idx+n; i++ {
		s.array[i] = f(s.array[i])
	}
}

// Copy copies count characters from src starting at srcOffset to this object at dstOffset
func (s *StrObject) Copy(src *StrObject, srcOffset, count, dstOffset int) {
	if count <= 0 {
		return
	}
	checkIndex(dstOffset, count-1)
	checkIndex(srcOffset, count-1)
	srcIdx := adjustIndex(srcOffset, 0)
	dstIdx := adjustIndex(dstOffset, 0)
	copy(s.array[dstIdx:dstIdx+count], src.array[srcIdx:srcIdx+count])
}

// CopyN copies count bytes from src to this object at dstOffset
func (s *StrObject) CopyN(src []byte, count, dstOffset int) {
	if count <= 0 {
		return
	}
	checkIndex(dstOffset, count-1)
	dstIdx := adjustIndex(dstOffset, 0)
	copy(s.array[dstIdx:dstIdx+count], src[:count])
}

// Erase removes n characters starting at from
func (s *StrObject) Erase(n, from int) {
	if n <= 0 {
		return
	}
	checkIndex(from, n-1)
	dstIdx := adjustIndex(from, 0)
	copy(s.array[dstIdx:], s.array[dstIdx+n:])
}

// Fill fills the range [start, end] with value
func (s *StrObject) Fill(value byte, start, end int) {
	startIdx := adjustIndex(start, 0)
	endIdx := adjustIndex(end, 0) + 1 // end is inclusive
	for i := startIdx; i < endIdx; i++ {
		s.array[i] = value
	}
}

// FillN fills n characters starting at start with value
func (s *StrObject) FillN(value byte, n, start int) {
	if n <= 0 {
		return
	}
	checkIndex(start, n-1)
	startIdx := adjustIndex(start, 0)
	for i := startIdx; i < startIdx+n; i++ {
		s.array[i] = value
	}
}

// FillCopy copies srcLen characters from src at srcIndex to dstLen positions at dstIndex,
// filling remaining positions with value if dstLen > srcLen
func (s *StrObject) FillCopy(src *StrObject, srcIndex, srcLen, dstIndex, dstLen int, value byte) {
	if dstLen <= 0 {
		return
	}
	checkIndex(dstIndex, dstLen-1)
	dstIdx := adjustIndex(dstIndex, 0)
	length := min(srcLen, dstLen)

	if length > 0 {
		srcIdx := adjustIndex(srcIndex, 0)
		copy(s.array[dstIdx:dstIdx+length], src.array[srcIdx:srcIdx+length])
	}
	for i := dstIdx + length; i < dstIdx+dstLen; i++ {
		s.array[i] = value
	}
}

// FillCopyBytes copies bytes from src to dstLen positions at dstIndex,
// filling remaining positions with value if dstLen > len(src)
func (s *StrObject) FillCopyBytes(src []byte, dstIndex, dstLen int, value byte) {
	if dstLen <= 0 {
		return
	}
	checkIndex(dstIndex, dstLen-1)
	dstIdx := adjustIndex(dstIndex, 0)
	length := min(len(src), dstLen)

	if length > 0 {
		copy(s.array[dstIdx:dstIdx+length], src[:length])
	}
	for i := dstIdx + length; i < dstIdx+dstLen; i++ {
		s.array[i] = value
	}
}

// Insert makes space for n characters at position at by shifting characters right
func (s *StrObject) Insert(n, at int) {
	if n <= 0 {
		return
	}
	checkIndex(at, n-1)
	atIdx := adjustIndex(at, 0)
	copy(s.array[atIdx+n:], s.array[atIdx:MaxStrLen-n])
}

// Length returns the position of the last character that is not equal to value,
// searching backwards from the 'from' position
func (s *StrObject) Length(value byte, from int) int {
	lastIdx := adjustIndex(from, 0)
	for i := lastIdx; i >= 0; i-- {
		if s.array[i] != value {
			return i + 1
		}
	}
	return 0
}

// Slice returns a view of length characters starting at index (1-based)
func (s *StrObject) Slice(index, length int) string {
	if length == 0 {
		return ""
	}
	checkIndex(index, length-1)
	idx := adjustIndex(index, 0)
	return string(s.array[idx : idx+length])
}

// String returns the entire array as a string
func (s *StrObject) String() string {
	return string(s.array[:])
}

// TrimmedString returns the string up to the last non-space character
func (s *StrObject) TrimmedString() string {
	length := s.Length(' ', MaxIndex)
	if length == 0 {
		return ""
	}
	return string(s.array[:length])
}

// Compare compares this StrObject with another
func (s *StrObject) Compare(other *StrObject) int {
	return bytes.Compare(s.array[:], other.array[:])
}

// Equal returns true if this StrObject is equal to another
func (s *StrObject) Equal(other *StrObject) bool {
	return s.array == other.array
}

// Bytes returns a copy of the underlying byte array
func (s *StrObject) Bytes() []byte {
	return append([]byte(nil), s.array[:]...)
}

// Format implements fmt.Formatter for pretty printing
func (s *StrObject) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		length := s.Length(' ', MaxIndex)
		f.Write(s.array[:length])
	case 'q':
		fmt.Fprintf(f, "%q", s.array[:])
	default:
		fmt.Fprintf(f, "%%!%c(StrObject)", verb)
	}
}
