// Handles a variable-size string object with 1-based indexing.
// Calls to methods will panic if indices are out of range.

package ludwig

import (
	"bytes"
	"fmt"
	"math"
)

const (
	// MinIndex is the minimum index (1-based indexing)
	MinIndex = 1
)

// StrObject represents a variable-size string object with 1-based indexing
type StrObject struct {
	array []byte
}

// Size returns the allocated size (max valid 1-based index)
func (s *StrObject) Size() int {
	return len(s.array)
}

// checkIndex validates that an index with optional offset is within valid range
func (s *StrObject) checkIndex(index, offset int) {
	if offset > math.MaxInt-index {
		panic("index + offset overflow")
	}

	combinedIndex := index + offset

	if combinedIndex < MinIndex || combinedIndex > len(s.array) {
		panic("index out of range")
	}
}

// adjustIndex converts 1-based index to 0-based array index
func (s *StrObject) adjustIndex(index, offset int) int {
	s.checkIndex(index, offset)
	return index + offset - MinIndex
}

// clampLen clamps a length so that (index + length - 1) doesn't exceed the array size.
// The start index must be valid (>= MinIndex and <= Size()+1).
// Returns the clamped length (may be 0 if start is at Size()+1).
func (s *StrObject) clampLen(index, length int) int {
	maxLen := len(s.array) - index + MinIndex
	if length > maxLen {
		return maxLen
	}
	return length
}

// clampEnd clamps an end index to the array size
func (s *StrObject) clampEnd(end int) int {
	if end > len(s.array) {
		return len(s.array)
	}
	return end
}

// NewBlankStrObject creates a new StrObject of the given size filled with the
// given element
func NewBlankStrObject(size int) *StrObject {
	s := &StrObject{array: make([]byte, size)}
	for i := range s.array {
		s.array[i] = ' '
	}
	return s
}

// NewStrObjectFrom creates a new StrObject from a string
func NewStrObjectFrom(str string) *StrObject {
	s := &StrObject{array: make([]byte, len(str))}
	s.Assign(str)
	return s
}

// NewStrObjectCopy creates a new StrObject by copying srcLen characters from
// src starting at srcIndex, and filling the rest of dstLen with spaces if
// dstLen > srcLen
func NewStrObjectCopy(src *StrObject, srcIndex int, srcLen int, dstLen int) *StrObject {
	s := &StrObject{array: make([]byte, dstLen)}
	s.Copy(src, srcIndex, srcLen, 1)
	s.Fill(' ', srcLen+1, dstLen)
	return s
}

// EmptyStrObject returns a StrObject with an empty, zero length string
func EmptyStrObject() *StrObject {
	// We return a fresh object, but could consider a shared object.
	return &StrObject{array: make([]byte, 0)}
}

// Clone creates a copy of the StrObject
func (s *StrObject) Clone() *StrObject {
	newArray := make([]byte, len(s.array))
	copy(newArray, s.array)
	return &StrObject{array: newArray}
}

// Get returns the character at the given 1-based index
func (s *StrObject) Get(index int) byte {
	idx := s.adjustIndex(index, 0)
	return s.array[idx]
}

// Set sets the character at the given 1-based index
func (s *StrObject) Set(index int, value byte) {
	idx := s.adjustIndex(index, 0)
	s.array[idx] = value
}

// Assign sets the content from a string, reallocating if necessary
func (s *StrObject) Assign(str string) {
	thisLen := len(s.array)
	if thisLen < len(str) {
		s.array = make([]byte, len(str))
	}
	s.FillCopyBytes([]byte(str), 1, len(s.array), ' ')
}

// Equals compares n characters starting at srcOffset with another StrObject starting at dstOffset
func (s *StrObject) Equals(other *StrObject, n, srcOffset, dstOffset int) bool {
	if n == 0 {
		return true
	}
	s.checkIndex(srcOffset, n-1)
	other.checkIndex(dstOffset, n-1)
	srcIdx := s.adjustIndex(srcOffset, 0)
	dstIdx := other.adjustIndex(dstOffset, 0)
	return bytes.Equal(s.array[srcIdx:srcIdx+n], other.array[dstIdx:dstIdx+n])
}

// ApplyN applies a function to n characters starting at start
func (s *StrObject) ApplyN(f func(byte) byte, n, start int) {
	if n <= 0 {
		return
	}
	n = s.clampLen(start, n)
	if n <= 0 {
		return
	}
	idx := s.adjustIndex(start, 0)
	for i := idx; i < idx+n; i++ {
		s.array[i] = f(s.array[i])
	}
}

// Copy copies count characters from src starting at srcOffset to this object at dstOffset
func (s *StrObject) Copy(src *StrObject, srcOffset, count, dstOffset int) {
	if count <= 0 {
		return
	}
	count = s.clampLen(dstOffset, count)
	count = src.clampLen(srcOffset, count)
	if count <= 0 {
		return
	}
	srcIdx := src.adjustIndex(srcOffset, 0)
	dstIdx := s.adjustIndex(dstOffset, 0)
	copy(s.array[dstIdx:dstIdx+count], src.array[srcIdx:srcIdx+count])
}

// CopyN copies count bytes from src to this object at dstOffset
func (s *StrObject) CopyN(src []byte, count, dstOffset int) {
	if count <= 0 {
		return
	}
	count = s.clampLen(dstOffset, count)
	if count <= 0 {
		return
	}
	dstIdx := s.adjustIndex(dstOffset, 0)
	copy(s.array[dstIdx:dstIdx+count], src[:count])
}

// Erase removes n characters starting at from
func (s *StrObject) Erase(n, from int) {
	if n <= 0 {
		return
	}
	n = s.clampLen(from, n)
	if n <= 0 {
		return
	}
	dstIdx := s.adjustIndex(from, 0)
	copy(s.array[dstIdx:], s.array[dstIdx+n:])
}

// Fill fills the range [start, end] with value
func (s *StrObject) Fill(value byte, start, end int) {
	end = s.clampEnd(end)
	if start > end {
		return
	}
	startIdx := s.adjustIndex(start, 0)
	endIdx := end // already 0-based upper bound (clamped to len)
	for i := startIdx; i < endIdx; i++ {
		s.array[i] = value
	}
}

// FillN fills n characters starting at start with value
func (s *StrObject) FillN(value byte, n, start int) {
	if n <= 0 {
		return
	}
	n = s.clampLen(start, n)
	if n <= 0 {
		return
	}
	startIdx := s.adjustIndex(start, 0)
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
	dstLen = s.clampLen(dstIndex, dstLen)
	if dstLen <= 0 {
		return
	}
	dstIdx := s.adjustIndex(dstIndex, 0)
	length := min(srcLen, dstLen)

	if length > 0 {
		length = src.clampLen(srcIndex, length)
		if length > 0 {
			srcIdx := src.adjustIndex(srcIndex, 0)
			copy(s.array[dstIdx:dstIdx+length], src.array[srcIdx:srcIdx+length])
		}
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
	dstLen = s.clampLen(dstIndex, dstLen)
	if dstLen <= 0 {
		return
	}
	dstIdx := s.adjustIndex(dstIndex, 0)
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
	n = s.clampLen(at, n)
	if n <= 0 {
		return
	}
	atIdx := s.adjustIndex(at, 0)
	copy(s.array[atIdx+n:], s.array[atIdx:len(s.array)-n])
}

// Len returns the maximum size of the underlying array
func (s *StrObject) Len() int {
	return len(s.array)
}

// Length returns the position of the last character that is not equal to value,
// searching backwards from the 'from' position
func (s *StrObject) Length(value byte, from int) int {
	from = s.clampEnd(from)
	lastIdx := from - MinIndex
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
	s.checkIndex(index, length-1)
	idx := s.adjustIndex(index, 0)
	return string(s.array[idx : idx+length])
}

// String returns the entire array as a string
func (s *StrObject) String() string {
	return string(s.array[:])
}

// TrimmedString returns the string up to the last non-space character
func (s *StrObject) TrimmedString() string {
	length := s.Length(' ', len(s.array))
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
	return bytes.Equal(s.array, other.array)
}

// Bytes returns a copy of the underlying byte array
func (s *StrObject) Bytes() []byte {
	return append([]byte(nil), s.array[:]...)
}

// Format implements fmt.Formatter for pretty printing
func (s *StrObject) Format(f fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		length := s.Length(' ', len(s.array))
		f.Write(s.array[:length])
	case 'q':
		fmt.Fprintf(f, "%q", s.array[:])
	default:
		fmt.Fprintf(f, "%%!%c(StrObject)", verb)
	}
}
