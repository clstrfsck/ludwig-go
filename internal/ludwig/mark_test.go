// Tests for functions in mark.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for test setup

// createTestLine creates a simple test line with text
func createTestLine() *LineHdrObject {
	line := &LineHdrObject{
		Used:  MaxStrLen,
		Marks: make([]*MarkObject, 0),
	}
	line.FLink = line
	line.BLink = line
	return line
}

// createLinkedLines creates a sequence of linked lines
func createLinkedLines(count int) []*LineHdrObject {
	if count < 1 {
		return nil
	}

	lines := make([]*LineHdrObject, count)
	for i := 0; i < count; i++ {
		lines[i] = &LineHdrObject{
			Used:  MaxStrLen,
			Marks: make([]*MarkObject, 0),
		}
	}

	// Link them together
	for i := 0; i < count; i++ {
		if i == 0 {
			lines[i].BLink = lines[count-1]
		} else {
			lines[i].BLink = lines[i-1]
		}

		if i == count-1 {
			lines[i].FLink = lines[0]
		} else {
			lines[i].FLink = lines[i+1]
		}
	}

	return lines
}

// Tests for removeFromMarks

func TestRemoveFromMarks(t *testing.T) {
	t.Run("RemoveFromNilList", func(t *testing.T) {
		mark := &MarkObject{}
		removeFromMarks(nil, mark)
		// Should not panic
	})

	t.Run("RemoveFromEmptyList", func(t *testing.T) {
		marks := []*MarkObject{}
		mark := &MarkObject{}
		removeFromMarks(&marks, mark)
		assert.Empty(t, marks, "Expected empty list")
	})

	t.Run("RemoveFirstMark", func(t *testing.T) {
		mark1 := &MarkObject{}
		mark2 := &MarkObject{}
		mark3 := &MarkObject{}
		marks := []*MarkObject{mark1, mark2, mark3}

		removeFromMarks(&marks, mark1)

		assert.Len(t, marks, 2, "Expected 2 marks")
		assert.Equal(t, mark2, marks[0], "Expected mark2 at index 0")
		assert.Equal(t, mark3, marks[1], "Expected mark3 at index 1")
	})

	t.Run("RemoveMiddleMark", func(t *testing.T) {
		mark1 := &MarkObject{}
		mark2 := &MarkObject{}
		mark3 := &MarkObject{}
		marks := []*MarkObject{mark1, mark2, mark3}

		removeFromMarks(&marks, mark2)

		assert.Len(t, marks, 2, "Expected 2 marks")
		assert.Equal(t, mark1, marks[0], "Expected mark1 at index 0")
		assert.Equal(t, mark3, marks[1], "Expected mark3 at index 1")
	})

	t.Run("RemoveLastMark", func(t *testing.T) {
		mark1 := &MarkObject{}
		mark2 := &MarkObject{}
		mark3 := &MarkObject{}
		marks := []*MarkObject{mark1, mark2, mark3}

		removeFromMarks(&marks, mark3)

		assert.Len(t, marks, 2, "Expected 2 marks")
		assert.Equal(t, mark1, marks[0], "Expected mark1 at index 0")
		assert.Equal(t, mark2, marks[1], "Expected mark2 at index 1")
	})

	t.Run("RemoveNonExistentMark", func(t *testing.T) {
		mark1 := &MarkObject{}
		mark2 := &MarkObject{}
		markOther := &MarkObject{}
		marks := []*MarkObject{mark1, mark2}

		removeFromMarks(&marks, markOther)

		assert.Len(t, marks, 2, "Expected 2 marks")
		assert.Equal(t, mark1, marks[0], "Expected mark1 at index 0")
		assert.Equal(t, mark2, marks[1], "Expected mark2 at index 1")
	})
}

// Tests for MarkCreate

func TestMarkCreate(t *testing.T) {
	t.Run("CreateNewMark", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		result := MarkCreate(line, 10, &mark)

		assert.True(t, result, "MarkCreate should return true")
		assert.NotNil(t, mark, "Mark should be created")
		assert.Equal(t, line, mark.Line, "Mark should be on the specified line")
		assert.Equal(t, 10, mark.Col, "Mark column should be 10")
		assert.Len(t, line.Marks, 1, "Line should have 1 mark")
		assert.Equal(t, mark, line.Marks[0], "Line's first mark should be the created mark")
	})

	t.Run("MoveMarkOnSameLine", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 10, &mark)
		result := MarkCreate(line, 20, &mark)

		assert.True(t, result, "MarkCreate should return true")
		assert.Equal(t, line, mark.Line, "Mark should still be on the same line")
		assert.Equal(t, 20, mark.Col, "Mark column should be 20")
		assert.Len(t, line.Marks, 1, "Line should still have 1 mark")
	})

	t.Run("MoveMarkToDifferentLine", func(t *testing.T) {
		line1 := createTestLine()
		line2 := createTestLine()
		var mark *MarkObject

		MarkCreate(line1, 10, &mark)
		result := MarkCreate(line2, 15, &mark)

		assert.True(t, result, "MarkCreate should return true")
		assert.Equal(t, line2, mark.Line, "Mark should be on line2")
		assert.Equal(t, 15, mark.Col, "Mark column should be 15")
		assert.Empty(t, line1.Marks, "Line1 should have 0 marks")
		assert.Len(t, line2.Marks, 1, "Line2 should have 1 mark")
	})

	t.Run("MultipleMarksOnSameLine", func(t *testing.T) {
		line := createTestLine()
		var mark1, mark2, mark3 *MarkObject

		MarkCreate(line, 10, &mark1)
		MarkCreate(line, 20, &mark2)
		MarkCreate(line, 30, &mark3)

		assert.Len(t, line.Marks, 3, "Line should have 3 marks")
		// Marks are prepended, so order should be reversed
		assert.Equal(t, mark3, line.Marks[0], "First mark should be mark3")
		assert.Equal(t, mark2, line.Marks[1], "Second mark should be mark2")
		assert.Equal(t, mark1, line.Marks[2], "Third mark should be mark1")
	})
}

// Tests for MarkDestroy

func TestMarkDestroy(t *testing.T) {
	t.Run("DestroyMark", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 10, &mark)
		result := MarkDestroy(&mark)

		assert.True(t, result, "MarkDestroy should return true")
		assert.Nil(t, mark, "Mark should be nil after destruction")
		assert.Empty(t, line.Marks, "Line should have 0 marks after destruction")
	})

	t.Run("DestroyOneOfMultipleMarks", func(t *testing.T) {
		line := createTestLine()
		var mark1, mark2, mark3 *MarkObject

		MarkCreate(line, 10, &mark1)
		MarkCreate(line, 20, &mark2)
		MarkCreate(line, 30, &mark3)

		MarkDestroy(&mark2)

		assert.Nil(t, mark2, "mark2 should be nil after destruction")
		assert.Len(t, line.Marks, 2, "Line should have 2 marks")
		// mark3 was last created (prepended), mark1 was first
		assert.Equal(t, mark3, line.Marks[0], "First mark should be mark3")
		assert.Equal(t, mark1, line.Marks[1], "Second mark should be mark1")
	})
}

// Tests for MarksSqueeze

func TestMarksSqueeze(t *testing.T) {
	t.Run("SqueezeOnSameLine", func(t *testing.T) {
		line := createTestLine()
		var mark1, mark2, mark3, mark4 *MarkObject

		MarkCreate(line, 5, &mark1)  // Before range
		MarkCreate(line, 10, &mark2) // In range
		MarkCreate(line, 15, &mark3) // In range
		MarkCreate(line, 25, &mark4) // After range

		result := MarksSqueeze(line, 10, line, 20)

		assert.True(t, result, "MarksSqueeze should return true")
		assert.Equal(t, 5, mark1.Col, "mark1 should stay at column 5")
		assert.Equal(t, 20, mark2.Col, "mark2 should move to column 20")
		assert.Equal(t, 20, mark3.Col, "mark3 should move to column 20")
		assert.Equal(t, 25, mark4.Col, "mark4 should stay at column 25")
		assert.Len(t, line.Marks, 4, "Line should still have 4 marks")
	})

	t.Run("SqueezeAcrossLines", func(t *testing.T) {
		lines := createLinkedLines(3)
		var mark1, mark2, mark3, mark4 *MarkObject

		MarkCreate(lines[0], 5, &mark1)  // Before firstColumn on first line
		MarkCreate(lines[0], 15, &mark2) // After firstColumn on first line
		MarkCreate(lines[1], 10, &mark3) // On middle line
		MarkCreate(lines[2], 5, &mark4)  // Before lastColumn on last line

		result := MarksSqueeze(lines[0], 10, lines[2], 20)

		assert.True(t, result, "MarksSqueeze should return true")
		assert.Equal(t, lines[0], mark1.Line)
		assert.Equal(t, 5, mark1.Col, "mark1 should stay at line[0] col 5")
		assert.Equal(t, lines[2], mark2.Line)
		assert.Equal(t, 20, mark2.Col, "mark2 should move to line[2] col 20")
		assert.Equal(t, lines[2], mark3.Line)
		assert.Equal(t, 20, mark3.Col, "mark3 should move to line[2] col 20")
		assert.Equal(t, lines[2], mark4.Line)
		assert.Equal(t, 20, mark4.Col, "mark4 should move to line[2] col 20")
	})

	t.Run("SqueezeEmptyRange", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 15, &mark)
		result := MarksSqueeze(line, 10, line, 10)

		assert.True(t, result, "MarksSqueeze should return true")
		assert.Equal(t, 15, mark.Col, "mark should stay at column 15")
	})

	t.Run("SqueezeWithNoMarksInRange", func(t *testing.T) {
		line := createTestLine()
		var mark1, mark2 *MarkObject

		MarkCreate(line, 5, &mark1)
		MarkCreate(line, 25, &mark2)
		result := MarksSqueeze(line, 10, line, 20)

		assert.True(t, result, "MarksSqueeze should return true")
		assert.Equal(t, 5, mark1.Col, "mark1 should stay at column 5")
		assert.Equal(t, 25, mark2.Col, "mark2 should stay at column 25")
	})
}

// Tests for MarksShift

func TestMarksShift(t *testing.T) {
	t.Run("ShiftOnSameLine", func(t *testing.T) {
		line := createTestLine()
		var mark1, mark2, mark3 *MarkObject

		MarkCreate(line, 10, &mark1) // In range
		MarkCreate(line, 15, &mark2) // In range
		MarkCreate(line, 25, &mark3) // Out of range

		// Shift columns 10-19 (width 10) to position 30
		result := MarksShift(line, 10, 10, line, 30)

		assert.True(t, result, "MarksShift should return true")
		assert.Equal(t, 30, mark1.Col, "mark1 should move to column 30")
		assert.Equal(t, 35, mark2.Col, "mark2 should move to column 35")
		assert.Equal(t, 25, mark3.Col, "mark3 should stay at column 25")
	})

	t.Run("ShiftToDifferentLine", func(t *testing.T) {
		line1 := createTestLine()
		line2 := createTestLine()
		var mark1, mark2, mark3 *MarkObject

		MarkCreate(line1, 10, &mark1) // In range
		MarkCreate(line1, 15, &mark2) // In range
		MarkCreate(line1, 25, &mark3) // Out of range

		// Shift columns 10-19 from line1 to line2 starting at position 5
		result := MarksShift(line1, 10, 10, line2, 5)

		assert.True(t, result, "MarksShift should return true")
		assert.Equal(t, line2, mark1.Line)
		assert.Equal(t, 5, mark1.Col, "mark1 should move to line2 col 5")
		assert.Equal(t, line2, mark2.Line)
		assert.Equal(t, 10, mark2.Col, "mark2 should move to line2 col 10")
		assert.Equal(t, line1, mark3.Line)
		assert.Equal(t, 25, mark3.Col, "mark3 should stay at line1 col 25")
		assert.Len(t, line1.Marks, 1, "line1 should have 1 mark")
		assert.Len(t, line2.Marks, 2, "line2 should have 2 marks")
	})

	t.Run("ShiftWithNegativeOffset", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 20, &mark)

		// Shift columns 20-29 to position 10 (negative offset of -10)
		result := MarksShift(line, 20, 10, line, 10)

		assert.True(t, result, "MarksShift should return true")
		assert.Equal(t, 10, mark.Col, "mark should move to column 10")
	})

	t.Run("ShiftBeyondMaxStrLen", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 10, &mark)

		// Try to shift way beyond MaxStrLenP
		result := MarksShift(line, 10, 10, line, MaxStrLenP+100)

		assert.True(t, result, "MarksShift should return true")
		assert.Equal(t, MaxStrLenP, mark.Col, "mark should be clamped to MaxStrLenP")
	})

	t.Run("ShiftWithNoMarksInRange", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 5, &mark)

		// Shift columns 10-19 to position 30
		result := MarksShift(line, 10, 10, line, 30)

		assert.True(t, result, "MarksShift should return true")
		assert.Equal(t, 5, mark.Col, "mark should stay at column 5")
	})

	t.Run("ShiftZeroWidth", func(t *testing.T) {
		line := createTestLine()
		var mark *MarkObject

		MarkCreate(line, 10, &mark)

		result := MarksShift(line, 10, 0, line, 30)

		assert.True(t, result, "MarksShift should return true")
		// With width 0, sourceEnd = 10 + 0 - 1 = 9, so col 10 is not in range
		assert.Equal(t, 10, mark.Col, "mark should stay at column 10")
	})
}

// Integration tests

func TestMarkIntegration(t *testing.T) {
	t.Run("CreateMoveAndDestroy", func(t *testing.T) {
		line1 := createTestLine()
		line2 := createTestLine()
		var mark *MarkObject

		// Create on line1
		MarkCreate(line1, 10, &mark)
		assert.Len(t, line1.Marks, 1, "line1 should have 1 mark")

		// Move to line2
		MarkCreate(line2, 20, &mark)
		assert.Empty(t, line1.Marks, "line1 should have 0 marks")
		assert.Len(t, line2.Marks, 1, "line2 should have 1 mark")

		// Destroy
		MarkDestroy(&mark)
		assert.Empty(t, line2.Marks, "line2 should have 0 marks after destruction")
		assert.Nil(t, mark, "mark should be nil")
	})

	t.Run("ComplexMultiLineScenario", func(t *testing.T) {
		lines := createLinkedLines(4)
		var m1, m2, m3, m4, m5 *MarkObject

		// Set up marks on different lines
		MarkCreate(lines[0], 10, &m1)
		MarkCreate(lines[0], 20, &m2)
		MarkCreate(lines[1], 15, &m3)
		MarkCreate(lines[2], 10, &m4)
		MarkCreate(lines[3], 5, &m5)

		// Squeeze from line[0] col 15 to line[2] col 20
		MarksSqueeze(lines[0], 15, lines[2], 20)

		// Check results
		assert.Equal(t, lines[0], m1.Line)
		assert.Equal(t, 10, m1.Col, "m1 should stay at line[0] col 10")
		assert.Equal(t, lines[2], m2.Line)
		assert.Equal(t, 20, m2.Col, "m2 should move to line[2] col 20")
		assert.Equal(t, lines[2], m3.Line)
		assert.Equal(t, 20, m3.Col, "m3 should move to line[2] col 20")
		assert.Equal(t, lines[2], m4.Line)
		assert.Equal(t, 20, m4.Col, "m4 should move to line[2] col 20")
		assert.Equal(t, lines[3], m5.Line)
		assert.Equal(t, 5, m5.Col, "m5 should stay at line[3] col 5")
	})
}

// Benchmarks

func BenchmarkMarkCreate(b *testing.B) {
	line := createTestLine()
	mark := new(*MarkObject)

	for i := 0; b.Loop(); i++ {
		MarkCreate(line, 10+i%100, mark)
	}
}

func BenchmarkMarkDestroy(b *testing.B) {
	// Create a reasonable number of marks to avoid O(n²) issues
	n := min(b.N, 10000)
	marks := make([]*MarkObject, n)
	lines := make([]*LineHdrObject, n)

	// Pre-create all marks with separate lines
	for i := range n {
		lines[i] = createTestLine()
		MarkCreate(lines[i], 10, &marks[i])
	}

	b.ResetTimer()
	for i := range n {
		MarkDestroy(&marks[i])
	}
}

func BenchmarkMarksSqueeze(b *testing.B) {
	// Setup once
	lines := createLinkedLines(5)
	var marks [10]*MarkObject
	for j := range 10 {
		MarkCreate(lines[j%5], 10+j*5, &marks[j])
	}

	for b.Loop() {
		// Just benchmark the squeeze operation
		// Note: This modifies state, so timing after first iteration reflects modified state
		MarksSqueeze(lines[0], 10, lines[4], 50)
	}
}

func BenchmarkMarksShift(b *testing.B) {
	// Setup once
	line1 := createTestLine()
	line2 := createTestLine()
	var marks [10]*MarkObject
	for j := range 10 {
		MarkCreate(line1, 10+j*5, &marks[j])
	}

	for b.Loop() {
		// Just benchmark the shift operation
		// Note: This modifies state, so timing after first iteration reflects modified state
		MarksShift(line1, 10, 50, line2, 20)
	}
}
