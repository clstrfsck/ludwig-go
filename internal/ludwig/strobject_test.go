// Tests for StrObject functionality

package ludwig

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewFilled tests the NewFilled constructor
func TestNewBlankStrObject(t *testing.T) {
	t.Run("Blank object contains blanks", func(t *testing.T) {
		s := NewBlankStrObject(MaxStrLen)
		assert.NotNil(t, s, "NewBlankStrObject returned nil")
		for i := 0; i < MaxStrLen; i++ {
			assert.Equal(t, byte(' '), s.array[i], "NewBlankStrObject(): array[%d] mismatch", i)
		}
	})
}

// TestClone tests the Clone method
func TestClone(t *testing.T) {
	original := NewBlankStrObject(MaxStrLen)
	original.Set(10, 'B')
	original.Set(20, 'C')

	clone := original.Clone()

	// Verify clone has same content
	assert.True(t, clone.Equal(original), "Clone does not equal original")

	// Verify it's a different object
	clone.Set(10, 'X')
	assert.NotEqual(t, byte('X'), original.Get(10), "Modifying clone affected original")
}

// TestGetSet tests Get and Set methods
func TestGetSet(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)

	// Test valid indices (1-based)
	testCases := []struct {
		index int
		value byte
	}{
		{1, 'A'},
		{10, 'B'},
		{MaxStrLen, 'Z'},
		{100, 'X'},
	}

	for _, tc := range testCases {
		s.Set(tc.index, tc.value)
		got := s.Get(tc.index)
		assert.Equal(t, tc.value, got, "Get(%d) mismatch", tc.index)
	}
}

// TestGetSetPanics tests that Get/Set panic on invalid indices
func TestGetSetPanics(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)

	invalidIndices := []int{0, -1, MaxStrLen + 1, -100}

	for _, idx := range invalidIndices {
		t.Run(fmt.Sprintf("Get(%d)", idx), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Get(%d) should panic but didn't", idx)
				}
			}()
			s.Get(idx)
		})

		t.Run(fmt.Sprintf("Set(%d)", idx), func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("Set(%d) should panic but didn't", idx)
				}
			}()
			s.Set(idx, 'X')
		})
	}
}

// TestAssign tests the Assign method
func TestAssign(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected func(s *StrObject) bool
	}{
		{
			name:  "short string",
			input: "Hello",
			expected: func(s *StrObject) bool {
				return s.Slice(1, 5) == "Hello" && s.Get(6) == ' '
			},
		},
		{
			name:  "empty string",
			input: "",
			expected: func(s *StrObject) bool {
				return s.Get(1) == ' '
			},
		},
		{
			name:  "string with spaces",
			input: "Hello World",
			expected: func(s *StrObject) bool {
				return s.Slice(1, 11) == "Hello World"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBlankStrObject(MaxStrLen)
			s.Assign(tt.input)
			assert.True(t, tt.expected(s), "Assign(%q) failed verification", tt.input)
		})
	}
}

// TestEquals tests the Equals method
func TestEquals(t *testing.T) {
	s1 := NewBlankStrObject(MaxStrLen)
	s1.Assign("ABCDEFGH")

	s2 := NewBlankStrObject(MaxStrLen)
	s2.Assign("ABCDXXGH")

	tests := []struct {
		name      string
		n         int
		srcOffset int
		dstOffset int
		want      bool
	}{
		{"zero length always equal", 0, 1, 1, true},
		{"equal at start", 4, 1, 1, true},
		{"not equal in middle", 6, 1, 1, false},
		{"equal suffix", 2, 7, 7, true},
		{"equal with offset", 2, 1, 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s1.Equals(s2, tt.n, tt.srcOffset, tt.dstOffset)
			assert.Equal(t, tt.want, got, "Equals(n=%d, src=%d, dst=%d) mismatch", tt.n, tt.srcOffset, tt.dstOffset)
		})
	}
}

// TestApplyN tests the ApplyN method
func TestApplyN(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Assign("hello")

	// Convert to uppercase
	toUpper := func(b byte) byte {
		if b >= 'a' && b <= 'z' {
			return b - 32
		}
		return b
	}

	s.ApplyN(toUpper, 5, 1)

	assert.Equal(t, "HELLO", s.Slice(1, 5), "ApplyN(toUpper) failed")

	// Test zero length (should be no-op)
	s.ApplyN(toUpper, 0, 1)
}

// TestCopy tests the Copy method
func TestCopy(t *testing.T) {
	src := NewBlankStrObject(MaxStrLen)
	src.Assign("Source Text")

	dst := NewBlankStrObject(MaxStrLen)
	dst.Fill('X', 1, MaxStrLen)

	// Copy 6 characters from position 1 to position 5
	dst.Copy(src, 1, 6, 5)

	assert.Equal(t, "Source", dst.Slice(5, 6), "Copy() failed")

	// Verify surrounding characters unchanged
	assert.Equal(t, byte('X'), dst.Get(4), "Copy affected preceding character")

	// Test zero count (should be no-op)
	original := dst.Clone()
	dst.Copy(src, 1, 0, 1)
	assert.True(t, dst.Equal(original), "Copy with count=0 modified object")
}

// TestCopyN tests the CopyN method
func TestCopyN(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	src := []byte("Hello World")

	s.CopyN(src, 5, 10)

	assert.Equal(t, "Hello", s.Slice(10, 5), "CopyN() failed")

	// Test zero count
	s.CopyN(src, 0, 1)
}

// TestErase tests the Erase method
func TestErase(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Assign("ABCDEFGHIJ")

	// Erase 3 characters starting at position 4
	// This removes DEF, shifting GHIJ left
	s.Erase(3, 4)

	// Should have "ABCGHIJ" followed by spaces
	expected := "ABCGHIJ"
	got := s.TrimmedString()
	assert.True(t, strings.HasPrefix(got, expected), "Erase() = %q, want prefix %q", got, expected)

	// Test zero count (no-op)
	before := s.Clone()
	s.Erase(0, 1)
	assert.True(t, s.Equal(before), "Erase with n=0 modified object")
}

// TestFill tests the Fill method
func TestFill(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)

	s.Fill('X', 10, 20)

	// Check filled range
	for i := 10; i <= 20; i++ {
		assert.Equal(t, byte('X'), s.Get(i), "Fill: position %d mismatch", i)
	}

	// Check outside range
	assert.NotEqual(t, byte('X'), s.Get(9), "Fill affected position before range")
	assert.NotEqual(t, byte('X'), s.Get(21), "Fill affected position after range")
}

// TestFillN tests the FillN method
func TestFillN(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)

	s.FillN('Y', 10, 15)

	// Check filled positions
	for i := range 10 {
		assert.Equal(t, byte('Y'), s.Get(15+i), "FillN: position %d mismatch", 15+i)
	}

	// Check before range
	assert.NotEqual(t, byte('Y'), s.Get(14), "FillN affected position before range")

	// Test zero count
	s.FillN('Z', 0, 1)
}

// TestFillCopy tests the FillCopy method
func TestFillCopy(t *testing.T) {
	src := NewBlankStrObject(MaxStrLen)
	src.Assign("Source")

	dst := NewBlankStrObject(MaxStrLen)

	tests := []struct {
		name     string
		srcIndex int
		srcLen   int
		dstIndex int
		dstLen   int
		fillVal  byte
		verify   func(t *testing.T, s *StrObject)
	}{
		{
			name:     "exact copy",
			srcIndex: 1,
			srcLen:   6,
			dstIndex: 10,
			dstLen:   6,
			fillVal:  '-',
			verify: func(t *testing.T, s *StrObject) {
				assert.Equal(t, "Source", s.Slice(10, 6), "exact copy failed")
			},
		},
		{
			name:     "copy with padding",
			srcIndex: 1,
			srcLen:   4,
			dstIndex: 20,
			dstLen:   8,
			fillVal:  '-',
			verify: func(t *testing.T, s *StrObject) {
				expected := "Sour----"
				assert.Equal(t, expected, s.Slice(20, 8), "copy with padding failed")
			},
		},
		{
			name:     "copy truncated",
			srcIndex: 1,
			srcLen:   6,
			dstIndex: 30,
			dstLen:   3,
			fillVal:  '-',
			verify: func(t *testing.T, s *StrObject) {
				assert.Equal(t, "Sou", s.Slice(30, 3), "copy truncated failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := NewBlankStrObject(MaxStrLen)
			dst.FillCopy(src, tt.srcIndex, tt.srcLen, tt.dstIndex, tt.dstLen, tt.fillVal)
			tt.verify(t, dst)
		})
	}

	// Test zero dstLen
	dst.FillCopy(src, 1, 5, 1, 0, '-')
}

// TestFillCopyBytes tests the FillCopyBytes method
func TestFillCopyBytes(t *testing.T) {
	src := []byte("Hello")

	tests := []struct {
		name     string
		dstIndex int
		dstLen   int
		fillVal  byte
		expected string
	}{
		{"exact fit", 10, 5, '-', "Hello"},
		{"with padding", 20, 8, '-', "Hello---"},
		{"truncated", 30, 3, '-', "Hel"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBlankStrObject(MaxStrLen)
			s.FillCopyBytes(src, tt.dstIndex, tt.dstLen, tt.fillVal)
			got := s.Slice(tt.dstIndex, tt.dstLen)
			assert.Equal(t, tt.expected, got, "FillCopyBytes() failed")
		})
	}
}

// TestInsert tests the Insert method
func TestInsert(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Assign("ABCDEFGH")

	// Insert 3 spaces at position 4
	// This shifts content from position 4 rightward by 3
	s.Insert(3, 4)

	// After insert, original position 4 ('D') should be at position 7
	// Position 4,5,6 should be undefined (from shifted content)
	if s.Get(7) == 'D' {
		// Correct behavior - D moved from 4 to 7
	} else {
		assert.Fail(t, "Insert did not shift correctly", "D should be at position 7, got %q", s.TrimmedString())
	}

	// Test zero count
	before := s.Clone()
	s.Insert(0, 1)
	assert.True(t, s.Equal(before), "Insert with n=0 modified object")
}

// TestLength tests the Length method
func TestLength(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		value    byte
		from     int
		expected int
	}{
		{"string with trailing spaces", "Hello", ' ', MaxStrLen, 5},
		{"all spaces", "", ' ', MaxStrLen, 0},
		{"find last non-space after ABC", "ABC", ' ', MaxStrLen, 3},
		{"search from middle", "Hello World", ' ', 50, 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBlankStrObject(MaxStrLen)
			if tt.content != "" {
				s.Assign(tt.content)
			}
			got := s.Length(tt.value, tt.from)
			assert.Equal(t, tt.expected, got, "Length(%q, %d) mismatch", tt.value, tt.from)
		})
	}
}

// TestSlice tests the Slice method
func TestSlice(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Assign("Hello World")

	tests := []struct {
		name     string
		index    int
		length   int
		expected string
	}{
		{"beginning", 1, 5, "Hello"},
		{"middle", 7, 5, "World"},
		{"single char", 1, 1, "H"},
		{"zero length", 1, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := s.Slice(tt.index, tt.length)
			assert.Equal(t, tt.expected, got, "Slice(%d, %d) failed", tt.index, tt.length)
		})
	}
}

// TestString tests the String method
func TestString(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Fill('X', 1, MaxStrLen)
	str := s.String()
	assert.Len(t, str, MaxStrLen, "String() length mismatch")
	assert.Equal(t, byte('X'), str[0], "String()[0] mismatch")
}

// TestTrimmedString tests the TrimmedString method
func TestTrimmedString(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{"short string", "Hello", "Hello"},
		{"empty", "", ""},
		{"with trailing spaces", "Test  ", "Test"},
		{"no trailing spaces due to length", "NoTrail", "NoTrail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBlankStrObject(MaxStrLen)
			if tt.content != "" {
				s.Assign(tt.content)
			}
			got := s.TrimmedString()
			assert.Equal(t, tt.expected, got, "TrimmedString() failed")
		})
	}
}

// TestCompare tests the Compare method
func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		content1 string
		content2 string
		expected int
	}{
		{"equal", "AAAA", "AAAA", 0},
		{"less than", "AAAA", "BBBB", -1},
		{"greater than", "BBBB", "AAAA", 1},
		{"prefix less", "AAAB", "AAAC", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s1 := NewBlankStrObject(MaxStrLen)
			s1.Assign(tt.content1)

			s2 := NewBlankStrObject(MaxStrLen)
			s2.Assign(tt.content2)

			got := s1.Compare(s2)
			// Normalize to -1, 0, 1
			if got < 0 {
				got = -1
			} else if got > 0 {
				got = 1
			}

			assert.Equal(t, tt.expected, got, "Compare() failed")
		})
	}
}

// TestEqual tests the Equal method
func TestEqual(t *testing.T) {
	s1 := NewBlankStrObject(MaxStrLen)
	s2 := NewBlankStrObject(MaxStrLen)
	s3 := NewBlankStrObject(MaxStrLen)
	s3.Fill('X', 1, MaxStrLen)

	assert.True(t, s1.Equal(s2), "Equal objects not detected as equal")
	assert.False(t, s1.Equal(s3), "Different objects detected as equal")

	// Test with same content through Assign
	s4 := NewBlankStrObject(MaxStrLen)
	s4.Assign("Test")
	s5 := NewBlankStrObject(MaxStrLen)
	s5.Assign("Test")

	assert.True(t, s4.Equal(s5), "Objects with same assigned content not equal")
}

// TestBytes tests the Bytes method
func TestBytes(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	b := s.Bytes()

	assert.Len(t, b, MaxStrLen, "Bytes() length mismatch")

	// Verify it's a copy, not the original
	b[0] = 'Y'
	assert.NotEqual(t, byte('Y'), s.array[0], "Bytes() returned reference to internal array")
}

// TestFormat tests the Format method
func TestFormat(t *testing.T) {
	s := NewBlankStrObject(MaxStrLen)
	s.Assign("Hello")

	tests := []struct {
		format   string
		expected string
	}{
		{"%s", "Hello"},
		{"%v", "Hello"},
		{"%q", fmt.Sprintf("%q", s.array[:])},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got := fmt.Sprintf(tt.format, s)
			assert.Equal(t, tt.expected, got, "Format(%s) failed", tt.format)
		})
	}
}

// TestCheckIndexOverflow tests overflow protection
func TestCheckIndexOverflow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("checkIndex should panic on overflow")
		}
	}()

	s := NewBlankStrObject(MaxStrLen)
	// This should cause an overflow panic
	s.Get(math.MaxInt)
}
