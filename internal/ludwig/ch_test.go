// Tests for ch.go functions
package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to set a string at a specific 1-based position
func setStringAt(s *StrObject, pos int, str string) {
	for i, ch := range []byte(str) {
		s.Set(pos+i, ch)
	}
}

// TestSgn tests the sign function
func TestSgn(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Positive", 42, 1},
		{"Negative", -42, -1},
		{"Zero", 0, 0},
		{"One", 1, 1},
		{"MinusOne", -1, -1},
		{"LargePositive", 1000000, 1},
		{"LargeNegative", -1000000, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sgn(tt.input)
			assert.Equal(t, tt.expected, result, "sgn(%d)", tt.input)
		})
	}
}

// TestChFillCopy tests the ChFillCopy function
func TestChFillCopy(t *testing.T) {
	t.Run("CopyFullSource", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}

		ChFillCopy(src, 1, 5, dst, 1, 5, ' ')

		assert.Equal(t, "HELLO", dst.Slice(1, 5))
	})

	t.Run("CopyWithFill", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HI")
		dst := &StrObject{}

		ChFillCopy(src, 1, 2, dst, 1, 5, '*')

		assert.Equal(t, "HI***", dst.Slice(1, 5))
	})

	t.Run("FillOnly", func(t *testing.T) {
		src := &StrObject{}
		dst := &StrObject{}

		ChFillCopy(src, 1, 0, dst, 1, 5, '-')

		assert.Equal(t, "-----", dst.Slice(1, 5))
	})

	t.Run("ZeroDestLen", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}
		setStringAt(dst, 1, "XXXXX")

		ChFillCopy(src, 1, 5, dst, 1, 0, ' ')

		// Should not modify dst
		assert.Equal(t, "XXXXX", dst.Slice(1, 5), "Expected unchanged")
	})

	t.Run("SourceLongerThanDest", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}

		ChFillCopy(src, 1, 5, dst, 1, 3, ' ')

		assert.Equal(t, "HEL", dst.Slice(1, 3))
	})

	t.Run("CopyWithOffset", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}

		ChFillCopy(src, 2, 3, dst, 3, 4, '.')

		assert.Equal(t, "ELL.", dst.Slice(3, 4))
	})
}

// TestChCompareStr tests string comparison
func TestChCompareStr(t *testing.T) {
	t.Run("ExactMatch", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "HELLO")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 5, true, &nchIdent)

		assert.Equal(t, 0, result, "Expected equal")
		assert.Equal(t, 5, nchIdent, "Expected 5 identical chars")
	})

	t.Run("FirstLessThanSecond", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "WORLD")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 5, false, &nchIdent)

		assert.Equal(t, -1, result, "Expected less than")
		assert.Equal(t, 0, nchIdent, "Expected 0 identical chars")
	})

	t.Run("FirstGreaterThanSecond", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "WORLD")
		s2 := &StrObject{}
		setStringAt(s2, 1, "HELLO")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 5, false, &nchIdent)

		assert.Equal(t, 1, result, "Expected greater than")
		assert.Equal(t, 0, nchIdent, "Expected 0 identical chars")
	})

	t.Run("PartialMatch", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "HELP")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 4, false, &nchIdent)

		assert.Equal(t, -1, result, "Expected -1 (L < P)")
		assert.Equal(t, 3, nchIdent, "Expected 3 identical chars (HEL)")
	})

	t.Run("DifferentLengths", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "HEL")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 3, false, &nchIdent)

		assert.Equal(t, 1, result, "Expected 1 (longer)")
		assert.Equal(t, 3, nchIdent, "Expected 3 identical chars")
	})

	t.Run("CaseInsensitiveMatch", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "hello")
		var nchIdent int

		// exactcase=true means case-insensitive (confusing name but per code/comment)
		result := ChCompareStr(s1, 1, 5, s2, 1, 5, false, &nchIdent)

		assert.Equal(t, 0, result, "Expected 0 (equal case-insensitive)")
		assert.Equal(t, 5, nchIdent, "Expected 5 identical chars")
	})

	t.Run("CaseInsensitiveDifferent", func(t *testing.T) {
		s1 := &StrObject{}
		setStringAt(s1, 1, "HELLO")
		s2 := &StrObject{}
		setStringAt(s2, 1, "world")
		var nchIdent int

		result := ChCompareStr(s1, 1, 5, s2, 1, 5, true, &nchIdent)

		assert.Equal(t, -1, result, "Expected -1 (H < W)")
	})

	t.Run("EmptyStrings", func(t *testing.T) {
		s1 := &StrObject{}
		s2 := &StrObject{}
		var nchIdent int

		result := ChCompareStr(s1, 1, 0, s2, 1, 0, false, &nchIdent)

		assert.Equal(t, 0, result, "Expected 0 (both empty)")
		assert.Equal(t, 0, nchIdent, "Expected 0 identical chars")
	})
}

// TestChReverseStr tests string reversal
func TestChReverseStr(t *testing.T) {
	t.Run("ReverseOddLength", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}

		ChReverseStr(src, dst, 5)

		assert.Equal(t, "OLLEH", dst.Slice(1, 5))
	})

	t.Run("ReverseEvenLength", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "TEST")
		dst := &StrObject{}

		ChReverseStr(src, dst, 4)

		assert.Equal(t, "TSET", dst.Slice(1, 4))
	})

	t.Run("ReverseSingleChar", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "A")
		dst := &StrObject{}

		ChReverseStr(src, dst, 1)

		assert.Equal(t, "A", dst.Slice(1, 1))
	})

	t.Run("ReverseInPlace", func(t *testing.T) {
		str := &StrObject{}
		setStringAt(str, 1, "ABCDE")

		ChReverseStr(str, str, 5)

		assert.Equal(t, "EDCBA", str.Slice(1, 5))
	})

	t.Run("ReverseZeroLength", func(t *testing.T) {
		src := &StrObject{}
		setStringAt(src, 1, "HELLO")
		dst := &StrObject{}
		setStringAt(dst, 1, "XXXXX")

		ChReverseStr(src, dst, 0)

		// Should not modify dst
		if dst.Slice(1, 5) != "XXXXX" {
			t.Errorf("Expected 'XXXXX' (unchanged), got '%s'", dst.Slice(1, 5))
		}
	})
}

// TestChToUpper tests character uppercasing
func TestChToUpper(t *testing.T) {
	tests := []struct {
		name     string
		input    byte
		expected byte
	}{
		{"LowercaseA", 'a', 'A'},
		{"LowercaseZ", 'z', 'Z'},
		{"UppercaseA", 'A', 'A'},
		{"UppercaseZ", 'Z', 'Z'},
		{"Digit", '5', '5'},
		{"Space", ' ', ' '},
		{"Punctuation", '!', '!'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ChToUpper(tt.input)
			assert.Equal(t, tt.expected, result, "ChToUpper(%c)", tt.input)
		})
	}
}

// TestChApplyN tests applying a function to n characters
func TestChApplyN(t *testing.T) {
	t.Run("ApplyToUpper", func(t *testing.T) {
		str := &StrObject{}
		setStringAt(str, 1, "hello world")

		ChApplyN(str, ChToUpper, 5)

		assert.Equal(t, "HELLO world", str.Slice(1, 11))
	})

	t.Run("ApplyToAll", func(t *testing.T) {
		str := &StrObject{}
		setStringAt(str, 1, "hello")

		ChApplyN(str, ChToUpper, 5)

		assert.Equal(t, "HELLO", str.Slice(1, 5))
	})

	t.Run("ApplyZeroChars", func(t *testing.T) {
		str := &StrObject{}
		setStringAt(str, 1, "hello")

		ChApplyN(str, ChToUpper, 0)

		assert.Equal(t, "hello", str.Slice(1, 5), "Expected unchanged")
	})

	t.Run("ApplyCustomFunction", func(t *testing.T) {
		str := &StrObject{}
		setStringAt(str, 1, "12345")

		// Custom function: add 1 to each character
		addOne := func(ch byte) byte { return ch + 1 }
		ChApplyN(str, addOne, 3)

		assert.Equal(t, "23445", str.Slice(1, 5))
	})
}

// TestChSearchStr tests string searching
func TestChSearchStr(t *testing.T) {
	t.Run("SearchForward", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "WORLD")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO WORLD TEST")
		var foundLoc int

		found := ChSearchStr(target, 1, 5, text, 1, 16, true, false, &foundLoc)

		assert.True(t, found, "Expected to find 'WORLD'")
		assert.Equal(t, 6, foundLoc, "Expected foundLoc=6")
	})

	t.Run("SearchReverse", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "DLROW") // "WORLD" reversed
		text := &StrObject{}
		setStringAt(text, 1, "HELLO WORLD TEST")
		var foundLoc int

		found := ChSearchStr(target, 1, 5, text, 1, 16, true, true, &foundLoc)

		assert.True(t, found, "Expected to find 'WORLD'")
		assert.Equal(t, 6, foundLoc, "Expected foundLoc=6")
	})

	t.Run("SearchNotFound", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "XYZ")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO WORLD")
		var foundLoc int

		found := ChSearchStr(target, 1, 3, text, 1, 11, true, false, &foundLoc)

		assert.False(t, found, "Expected not to find 'XYZ'")
	})

	t.Run("SearchCaseInsensitive", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "WORLD")
		text := &StrObject{}
		setStringAt(text, 1, "hello world test")
		var foundLoc int

		// exactcase=false for case-insensitive search (text is uppercased)
		found := ChSearchStr(target, 1, 5, text, 1, 16, false, false, &foundLoc)

		assert.True(t, found, "Expected to find 'world' (case-insensitive)")
		assert.Equal(t, 6, foundLoc, "Expected foundLoc=6")
	})

	t.Run("SearchAtBeginning", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "HELLO")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO WORLD")
		var foundLoc int

		found := ChSearchStr(target, 1, 5, text, 1, 11, true, false, &foundLoc)

		assert.True(t, found, "Expected to find 'HELLO'")
		assert.Equal(t, 0, foundLoc, "Expected foundLoc=0")
	})

	t.Run("SearchAtEnd", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "TEST")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO TEST")
		var foundLoc int

		found := ChSearchStr(target, 1, 4, text, 1, 10, false, false, &foundLoc)

		assert.True(t, found, "Expected to find 'TEST'")
		assert.Equal(t, 6, foundLoc, "Expected foundLoc=6")
	})

	t.Run("SearchSingleChar", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "W")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO WORLD")
		var foundLoc int

		found := ChSearchStr(target, 1, 1, text, 1, 11, false, false, &foundLoc)

		assert.True(t, found, "Expected to find 'W'")
		assert.Equal(t, 6, foundLoc, "Expected foundLoc=6")
	})

	t.Run("SearchTargetLongerThanText", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "HELLO WORLD")
		text := &StrObject{}
		setStringAt(text, 1, "HI")
		var foundLoc int

		found := ChSearchStr(target, 1, 11, text, 1, 2, false, false, &foundLoc)

		assert.False(t, found, "Expected not to find target longer than text")
	})

	t.Run("SearchMultipleOccurrences", func(t *testing.T) {
		target := &StrObject{}
		setStringAt(target, 1, "L")
		text := &StrObject{}
		setStringAt(text, 1, "HELLO")
		var foundLoc int

		// Should find first occurrence
		found := ChSearchStr(target, 1, 1, text, 1, 5, false, false, &foundLoc)

		assert.True(t, found, "Expected to find 'L'")
		assert.Equal(t, 2, foundLoc, "Expected foundLoc=2 (first L)")
	})
}

// BenchmarkChFillCopy benchmarks the fill copy operation
func BenchmarkChFillCopy(b *testing.B) {
	src := &StrObject{}
	setStringAt(src, 1, "HELLO")
	dst := &StrObject{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ChFillCopy(src, 1, 5, dst, 1, 10, ' ')
	}
}

// BenchmarkChCompareStr benchmarks string comparison
func BenchmarkChCompareStr(b *testing.B) {
	s1 := &StrObject{}
	setStringAt(s1, 1, "HELLO WORLD")
	s2 := &StrObject{}
	setStringAt(s2, 1, "HELLO WORLD")
	var nchIdent int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ChCompareStr(s1, 1, 11, s2, 1, 11, false, &nchIdent)
	}
}

// BenchmarkChSearchStr benchmarks string searching
func BenchmarkChSearchStr(b *testing.B) {
	target := &StrObject{}
	setStringAt(target, 1, "WORLD")
	text := &StrObject{}
	setStringAt(text, 1, "HELLO WORLD TEST")
	var foundLoc int

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ChSearchStr(target, 1, 5, text, 1, 16, false, false, &foundLoc)
	}
}
