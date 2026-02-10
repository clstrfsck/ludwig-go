// Test file for functions in text.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for text tests

// setupTestLineInFrame creates a test line within a frame with a group
func setupTestLineInFrame() (*FrameObject, *LineHdrObject) {
	frame := &FrameObject{
		SpaceLeft:    MaxSpace,
		SpaceLimit:   MaxSpace,
		ScrWidth:     80,
		ScrOffset:    0,
		MarginLeft:   1,
		MarginRight:  MaxStrLen,
		TextModified: false,
		Options:      0,
	}

	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: 1,
		NrLines:     1,
	}

	line1 := &LineHdrObject{
		Group:    group,
		OffsetNr: 0,
		Used:     0,
		ScrRowNr: 0,
		Str:      NewBlankStrObject(MaxStrLen),
	}

	// Add NULL line at the end
	nullLine := &LineHdrObject{
		Group:    group,
		OffsetNr: 1,
		FLink:    nil,
		BLink:    line1,
		Used:     0,
		Str:      nil,
	}
	line1.FLink = nullLine

	group.FirstLine = line1
	group.LastLine = nullLine
	frame.FirstGroup = group
	frame.LastGroup = group

	// Create marks array for the frame
	for i := range frame.Marks {
		frame.Marks[i] = nil
	}

	// Create a dot mark (required for many operations)
	frame.Dot = &MarkObject{
		Line: line1,
		Col:  1,
	}

	return frame, line1
}

// setupTestLineWithContent creates a test line with specified content
func setupTestLineWithContent(content string) (*FrameObject, *LineHdrObject) {
	frame, line := setupTestLineInFrame()

	// Copy content into the line
	for i, ch := range content {
		if i < MaxStrLen {
			line.Str.Set(i+1, byte(ch))
		}
	}
	line.Used = len(content)

	return frame, line
}

// setupLinkedLines creates a chain of lines
func setupLinkedLines(count int) (*FrameObject, []*LineHdrObject) {
	frame := &FrameObject{
		SpaceLeft:    MaxSpace,
		SpaceLimit:   MaxSpace,
		ScrWidth:     80,
		ScrOffset:    0,
		MarginLeft:   1,
		MarginRight:  MaxStrLen,
		TextModified: false,
		Options:      0,
	}

	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: 1,
		NrLines:     count,
	}

	lines := make([]*LineHdrObject, count)
	for i := 0; i < count; i++ {
		lines[i] = &LineHdrObject{
			Group:    group,
			OffsetNr: i,
			Used:     0,
			ScrRowNr: 0,
			Str:      NewBlankStrObject(MaxStrLen),
		}

		if i > 0 {
			lines[i].BLink = lines[i-1]
			lines[i-1].FLink = lines[i]
		}
	}

	// Add NULL line at the end
	nullLine := &LineHdrObject{
		Group:    group,
		OffsetNr: count,
		FLink:    nil,
		BLink:    lines[count-1],
		Str:      nil,
	}
	lines[count-1].FLink = nullLine

	group.FirstLine = lines[0]
	group.LastLine = nullLine
	frame.FirstGroup = group
	frame.LastGroup = group

	// Create marks array for the frame
	for i := range frame.Marks {
		frame.Marks[i] = nil
	}

	return frame, lines
}

// TestTextReturnCol tests the TextReturnCol function
func TestTextReturnCol(t *testing.T) {
	t.Run("ColumnBeforeMarginLeft", func(t *testing.T) {
		frame, line := setupTestLineInFrame()
		frame.MarginLeft = 10

		newCol := TextReturnCol(line, 5, false)
		assert.Equal(t, 1, newCol, "Expected column 1 when curCol < MarginLeft")
	})

	t.Run("ColumnAtOrAfterMarginLeft", func(t *testing.T) {
		frame, line := setupTestLineInFrame()
		frame.MarginLeft = 10

		newCol := TextReturnCol(line, 15, false)
		assert.Equal(t, 10, newCol, "Expected column to be MarginLeft")
	})

	t.Run("AutoIndentDisabled", func(t *testing.T) {
		frame, line := setupTestLineInFrame()
		frame.MarginLeft = 5
		frame.Options = 0 // No auto-indent

		// Create next line
		nextLine := &LineHdrObject{
			Group:    line.Group,
			OffsetNr: 1,
			Used:     10,
			Str:      NewBlankStrObject(MaxStrLen),
			BLink:    line,
		}
		line.FLink = nextLine

		// Set some content with leading spaces
		nextLine.Str.Set(5, 'A')

		newCol := TextReturnCol(line, 10, false)
		assert.Equal(t, 5, newCol, "Expected MarginLeft without auto-indent")
	})

	t.Run("AutoIndentEnabled", func(t *testing.T) {
		frame, line := setupTestLineInFrame()
		frame.MarginLeft = 1
		frame.Options.Set(OptAutoIndent)

		// Put content on the current line with leading spaces
		for i := 1; i <= 4; i++ {
			line.Str.Set(i, ' ')
		}
		line.Str.Set(5, 'A')
		line.Used = 5

		// The function should find the first non-space character
		newCol := TextReturnCol(line, 10, false)
		assert.Equal(t, 5, newCol, "Expected auto-indent to column 5")
	})
}

// TestTextRealizeNull tests the TextRealizeNull function
func TestTextRealizeNull(t *testing.T) {
	t.Run("RealizeNullLine", func(t *testing.T) {
		frame, lines := setupLinkedLines(2)

		// Create a null line
		nullLine := &LineHdrObject{
			Group:    lines[0].Group,
			OffsetNr: 2,
			FLink:    nil,
			BLink:    lines[1],
			Str:      nil,
		}
		lines[1].FLink = nullLine

		// Create a dot mark (required for TextRealizeNull)
		dot := &MarkObject{
			Line: nullLine,
			Col:  1,
		}
		frame.Dot = dot

		result := TextRealizeNull(nullLine)
		assert.True(t, result, "TextRealizeNull should succeed")
		assert.True(t, frame.TextModified, "Frame should be marked as modified")
	})
}

// TestTextInsert tests the TextInsert function
func TestTextInsert(t *testing.T) {
	t.Run("InsertIntoEmptyLine", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		// Create a simple string to insert
		insertStr := NewBlankStrObject(MaxStrLen)
		insertStr.Set(1, 'H')
		insertStr.Set(2, 'i')

		result := TextInsert(false, 1, insertStr, 2, mark)
		assert.True(t, result, "TextInsert should succeed")
		assert.Equal(t, byte('H'), line.Str.Get(1), "First character should be 'H'")
		assert.Equal(t, byte('i'), line.Str.Get(2), "Second character should be 'i'")
		assert.Equal(t, 2, line.Used, "Line.Used should be 2")
	})

	t.Run("InsertInMiddleOfLine", func(t *testing.T) {
		_, line := setupTestLineWithContent("Hello")

		mark := &MarkObject{
			Line: line,
			Col:  4, // Insert before 'l'
		}

		insertStr := NewBlankStrObject(MaxStrLen)
		insertStr.Set(1, 'X')
		insertStr.Set(2, 'Y')

		result := TextInsert(false, 1, insertStr, 2, mark)
		assert.True(t, result, "TextInsert should succeed")
		assert.Equal(t, byte('H'), line.Str.Get(1))
		assert.Equal(t, byte('e'), line.Str.Get(2))
		assert.Equal(t, byte('l'), line.Str.Get(3))
		assert.Equal(t, byte('X'), line.Str.Get(4))
		assert.Equal(t, byte('Y'), line.Str.Get(5))
		assert.Equal(t, byte('l'), line.Str.Get(6))
		assert.Equal(t, byte('o'), line.Str.Get(7))
	})

	t.Run("InsertMultipleCopies", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		insertStr := NewBlankStrObject(MaxStrLen)
		insertStr.Set(1, 'A')

		result := TextInsert(false, 3, insertStr, 1, mark)
		assert.True(t, result, "TextInsert should succeed")
		assert.Equal(t, byte('A'), line.Str.Get(1))
		assert.Equal(t, byte('A'), line.Str.Get(2))
		assert.Equal(t, byte('A'), line.Str.Get(3))
		assert.Equal(t, 3, line.Used)
	})

	t.Run("InsertExceedsMaxLength", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		insertStr := NewBlankStrObject(MaxStrLen)

		// Try to insert too much data
		result := TextInsert(false, 1, insertStr, MaxStrLen+10, mark)
		assert.False(t, result, "TextInsert should fail when exceeding MaxStrLen")
	})

	t.Run("InsertAtNullLine", func(t *testing.T) {
		frame, lines := setupLinkedLines(2)

		// Get the null line at the end
		nullLine := lines[1].FLink

		mark := &MarkObject{
			Line: nullLine,
			Col:  1,
		}

		// Create dot for frame (needed for TextRealizeNull)
		frame.Dot = &MarkObject{
			Line: nullLine,
			Col:  1,
		}

		insertStr := NewBlankStrObject(MaxStrLen)
		insertStr.Set(1, 'X')

		result := TextInsert(false, 1, insertStr, 1, mark)
		// This will realize the null line first
		assert.True(t, result, "TextInsert at null line should succeed")
	})
}

// TestTextOvertype tests the TextOvertype function
func TestTextOvertype(t *testing.T) {
	t.Run("OvertypeEmptyLine", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		overtypeStr := NewBlankStrObject(MaxStrLen)
		overtypeStr.Set(1, 'A')
		overtypeStr.Set(2, 'B')

		result := TextOvertype(false, 1, overtypeStr, 2, mark)
		assert.True(t, result, "TextOvertype should succeed")
		assert.Equal(t, byte('A'), line.Str.Get(1))
		assert.Equal(t, byte('B'), line.Str.Get(2))
		assert.Equal(t, 2, line.Used)
		assert.Equal(t, 3, mark.Col, "Mark should advance to column 3")
	})

	t.Run("OvertypeExistingContent", func(t *testing.T) {
		_, line := setupTestLineWithContent("Hello")

		mark := &MarkObject{
			Line: line,
			Col:  2,
		}

		overtypeStr := NewBlankStrObject(MaxStrLen)
		overtypeStr.Set(1, 'X')
		overtypeStr.Set(2, 'Y')

		result := TextOvertype(false, 1, overtypeStr, 2, mark)
		assert.True(t, result, "TextOvertype should succeed")
		assert.Equal(t, byte('H'), line.Str.Get(1))
		assert.Equal(t, byte('X'), line.Str.Get(2))
		assert.Equal(t, byte('Y'), line.Str.Get(3))
		assert.Equal(t, byte('l'), line.Str.Get(4))
		assert.Equal(t, byte('o'), line.Str.Get(5))
		assert.Equal(t, 4, mark.Col, "Mark should advance")
	})

	t.Run("OvertypeMultipleCopies", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		overtypeStr := NewBlankStrObject(MaxStrLen)
		overtypeStr.Set(1, 'Z')

		result := TextOvertype(false, 5, overtypeStr, 1, mark)
		assert.True(t, result, "TextOvertype should succeed")
		for i := 1; i <= 5; i++ {
			assert.Equal(t, byte('Z'), line.Str.Get(i), "Character %d should be 'Z'", i)
		}
		assert.Equal(t, 6, mark.Col, "Mark should advance to column 6")
	})

	t.Run("OvertypeExceedsMaxLength", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  MaxStrLen - 5,
		}

		overtypeStr := NewBlankStrObject(MaxStrLen)

		result := TextOvertype(false, 1, overtypeStr, 20, mark)
		assert.False(t, result, "TextOvertype should fail when exceeding MaxStrLen")
	})
}

// TestTextRemove tests the TextRemove function
func TestTextRemove(t *testing.T) {
	t.Run("RemoveWithinSingleLine", func(t *testing.T) {
		_, line := setupTestLineWithContent("Hello World")

		markOne := &MarkObject{
			Line: line,
			Col:  7, // After "Hello "
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  12, // After "Hello World"
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove should succeed")

		// Should have "Hello " left, but trailing spaces are trimmed
		assert.Equal(t, byte('H'), line.Str.Get(1))
		assert.Equal(t, byte('e'), line.Str.Get(2))
		assert.Equal(t, byte('l'), line.Str.Get(3))
		assert.Equal(t, byte('l'), line.Str.Get(4))
		assert.Equal(t, byte('o'), line.Str.Get(5))
		// The Used field gets trimmed to remove trailing spaces
		assert.Equal(t, 5, line.Used)
	})

	t.Run("RemoveFromStartOfLine", func(t *testing.T) {
		_, line := setupTestLineWithContent("Test")

		markOne := &MarkObject{
			Line: line,
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  3,
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove should succeed")

		// Should have "st" left
		assert.Equal(t, byte('s'), line.Str.Get(1))
		assert.Equal(t, byte('t'), line.Str.Get(2))
		assert.Equal(t, 2, line.Used)
	})

	t.Run("RemoveEntireLine", func(t *testing.T) {
		_, line := setupTestLineWithContent("Test")

		markOne := &MarkObject{
			Line: line,
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  5,
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove should succeed")
		assert.Equal(t, 0, line.Used, "Line should be empty")
	})

	t.Run("RemoveAcrossMultipleLines", func(t *testing.T) {
		_, lines := setupLinkedLines(3)

		// Set content on each line
		for i, content := range []string{"First line", "Second line", "Third line"} {
			for j, ch := range content {
				lines[i].Str.Set(j+1, byte(ch))
			}
			lines[i].Used = len(content)
		}

		// Remove from middle of first line to middle of third line
		markOne := &MarkObject{
			Line: lines[0],
			Col:  7, // After "First "
		}
		markTwo := &MarkObject{
			Line: lines[2],
			Col:  7, // After "Third "
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove across lines should succeed")

		// After inter-line remove, the result is in markTwo.Line (lines[2])
		// It should have "First " + "line" (remainder of third line)
		// = "First line"
		resultLine := markTwo.Line
		assert.Equal(t, byte('F'), resultLine.Str.Get(1))
		assert.Equal(t, byte('i'), resultLine.Str.Get(2))
		assert.Equal(t, byte('r'), resultLine.Str.Get(3))
		assert.Equal(t, byte('s'), resultLine.Str.Get(4))
		assert.Equal(t, byte('t'), resultLine.Str.Get(5))
		assert.Equal(t, byte(' '), resultLine.Str.Get(6))
		assert.Equal(t, byte('l'), resultLine.Str.Get(7))
		assert.Equal(t, byte('i'), resultLine.Str.Get(8))
		assert.Equal(t, byte('n'), resultLine.Str.Get(9))
		assert.Equal(t, byte('e'), resultLine.Str.Get(10))
	})

	t.Run("RemoveFromStartOfFirstLineToEndOfSecondLine", func(t *testing.T) {
		_, lines := setupLinkedLines(3)

		// Set content on lines
		line1Content := "First"
		line2Content := "Second"
		for i, ch := range line1Content {
			lines[0].Str.Set(i+1, byte(ch))
		}
		lines[0].Used = len(line1Content)
		for i, ch := range line2Content {
			lines[1].Str.Set(i+1, byte(ch))
		}
		lines[1].Used = len(line2Content)

		// Remove from start of first line to end of second line
		markOne := &MarkObject{
			Line: lines[0],
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: lines[1],
			Col:  7, // After "Second"
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove from start should succeed")

		// Result should be in markTwo.Line (lines[1]) and should be empty
		// because we removed everything from start of line 0 to end of line 1
		resultLine := markTwo.Line
		assert.Equal(t, 0, resultLine.Used, "Result line should be empty")
	})

	t.Run("RemoveEntireMiddleLine", func(t *testing.T) {
		_, lines := setupLinkedLines(3)

		// Set content
		for i, content := range []string{"AAA", "BBB", "CCC"} {
			for j, ch := range content {
				lines[i].Str.Set(j+1, byte(ch))
			}
			lines[i].Used = len(content)
		}

		// Remove from end of first line to start of third line
		markOne := &MarkObject{
			Line: lines[0],
			Col:  4, // After "AAA"
		}
		markTwo := &MarkObject{
			Line: lines[2],
			Col:  1, // Before "CCC"
		}

		result := TextRemove(markOne, markTwo)
		assert.True(t, result, "TextRemove of middle line should succeed")

		// The result is in markTwo.Line (lines[2])
		// Should have "AAA" (from before markOne.Col) + "CCC" (from markTwo.Col onwards)
		resultLine := markTwo.Line
		assert.Equal(t, byte('A'), resultLine.Str.Get(1))
		assert.Equal(t, byte('A'), resultLine.Str.Get(2))
		assert.Equal(t, byte('A'), resultLine.Str.Get(3))
		assert.Equal(t, byte('C'), resultLine.Str.Get(4))
		assert.Equal(t, byte('C'), resultLine.Str.Get(5))
		assert.Equal(t, byte('C'), resultLine.Str.Get(6))
	})
}

// TestTextSplitLine tests the TextSplitLine function
func TestTextSplitLine(t *testing.T) {
	t.Run("SplitEmptyLine", func(t *testing.T) {
		frame, lines := setupLinkedLines(2)
		line := lines[0]

		// Create dot for frame (needed for split)
		frame.Dot = &MarkObject{
			Line: line,
			Col:  1,
		}

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		var equalsMark *MarkObject
		result := TextSplitLine(mark, 1, &equalsMark)
		assert.True(t, result, "TextSplitLine should succeed")
		assert.NotNil(t, equalsMark, "Equals mark should be created")
		assert.True(t, frame.TextModified, "Frame should be marked as modified")
	})

	t.Run("SplitLineWithContent", func(t *testing.T) {
		frame, lines := setupLinkedLines(2)
		line := lines[0]

		// Set content
		content := "Hello World"
		for i, ch := range content {
			line.Str.Set(i+1, byte(ch))
		}
		line.Used = len(content)

		// Create dot for frame
		frame.Dot = &MarkObject{
			Line: line,
			Col:  1,
		}

		mark := &MarkObject{
			Line: line,
			Col:  7, // Split after "Hello "
		}

		var equalsMark *MarkObject
		result := TextSplitLine(mark, 1, &equalsMark)
		assert.True(t, result, "TextSplitLine should succeed")
		assert.NotNil(t, equalsMark, "Equals mark should be created")
	})

	t.Run("CannotSplitNullLine", func(t *testing.T) {
		_, lines := setupLinkedLines(2)
		nullLine := lines[1].FLink

		mark := &MarkObject{
			Line: nullLine,
			Col:  1,
		}

		var equalsMark *MarkObject
		result := TextSplitLine(mark, 1, &equalsMark)
		assert.False(t, result, "Should not be able to split null line")
	})
}

// TestTextMove tests the TextMove function
func TestTextMove(t *testing.T) {
	t.Run("CopySingleLine", func(t *testing.T) {
		frame, line := setupTestLineWithContent("Hello")

		markOne := &MarkObject{
			Line: line,
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  6,
		}
		dst := &MarkObject{
			Line: line,
			Col:  10,
		}

		// Create dot for frame
		frame.Dot = &MarkObject{
			Line: line,
			Col:  1,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(true, 1, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove (copy) should succeed")
		assert.NotNil(t, newStart, "newStart should be set")
		assert.NotNil(t, newEnd, "newEnd should be set")
	})

	t.Run("MoveWithinSameLine", func(t *testing.T) {
		frame, line := setupTestLineWithContent("Hello World")

		markOne := &MarkObject{
			Line: line,
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  6,
		}
		dst := &MarkObject{
			Line: line,
			Col:  12,
		}

		// Create dot for frame
		frame.Dot = &MarkObject{
			Line: line,
			Col:  1,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(false, 1, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove should succeed")
		assert.NotNil(t, newStart, "newStart should be set")
		assert.NotNil(t, newEnd, "newEnd should be set")
		assert.True(t, frame.TextModified, "Frame should be marked as modified")
	})

	t.Run("CopyZeroCount", func(t *testing.T) {
		_, line := setupTestLineWithContent("Test")

		markOne := &MarkObject{
			Line: line,
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: line,
			Col:  5,
		}
		dst := &MarkObject{
			Line: line,
			Col:  6,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(true, 0, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove with count=0 should succeed")
	})

	t.Run("CopyAcrossMultipleLines", func(t *testing.T) {
		_, lines := setupLinkedLines(4)

		// Set content on first two lines
		line1Content := "First"
		line2Content := "Second"
		for i, ch := range line1Content {
			lines[0].Str.Set(i+1, byte(ch))
		}
		lines[0].Used = len(line1Content)
		for i, ch := range line2Content {
			lines[1].Str.Set(i+1, byte(ch))
		}
		lines[1].Used = len(line2Content)

		// Copy from first line to second line, paste at third line
		markOne := &MarkObject{
			Line: lines[0],
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: lines[1],
			Col:  7, // After "Second"
		}
		dst := &MarkObject{
			Line: lines[2],
			Col:  1,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(true, 1, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove (copy) across lines should succeed")
		assert.NotNil(t, newStart, "newStart should be set")
		assert.NotNil(t, newEnd, "newEnd should be set")

		// Original lines should be unchanged (copy, not move)
		assert.Equal(t, byte('F'), lines[0].Str.Get(1))
		assert.Equal(t, byte('i'), lines[0].Str.Get(2))
		assert.Equal(t, 5, lines[0].Used)
		assert.Equal(t, byte('S'), lines[1].Str.Get(1))
		assert.Equal(t, 6, lines[1].Used)
	})

	t.Run("MoveAcrossMultipleLines", func(t *testing.T) {
		frame, lines := setupLinkedLines(4)

		// Set content on first two lines
		line1Content := "AAA"
		line2Content := "BBB"
		for i, ch := range line1Content {
			lines[0].Str.Set(i+1, byte(ch))
		}
		lines[0].Used = len(line1Content)
		for i, ch := range line2Content {
			lines[1].Str.Set(i+1, byte(ch))
		}
		lines[1].Used = len(line2Content)

		// Move from first line to second line, paste at third line
		markOne := &MarkObject{
			Line: lines[0],
			Col:  2, // After "A"
		}
		markTwo := &MarkObject{
			Line: lines[1],
			Col:  3, // After "BB"
		}
		dst := &MarkObject{
			Line: lines[2],
			Col:  1,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(false, 1, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove across lines should succeed")
		assert.NotNil(t, newStart, "newStart should be set")
		assert.NotNil(t, newEnd, "newEnd should be set")
		assert.True(t, frame.TextModified, "Frame should be marked as modified")
	})

	t.Run("MoveMultipleLinesWithCount", func(t *testing.T) {
		_, lines := setupLinkedLines(5)

		// Set content
		line1Content := "First"
		line2Content := "Second"
		for i, ch := range line1Content {
			lines[0].Str.Set(i+1, byte(ch))
		}
		lines[0].Used = len(line1Content)
		for i, ch := range line2Content {
			lines[1].Str.Set(i+1, byte(ch))
		}
		lines[1].Used = len(line2Content)

		// Copy multiple times
		markOne := &MarkObject{
			Line: lines[0],
			Col:  1,
		}
		markTwo := &MarkObject{
			Line: lines[1],
			Col:  4, // After "Sec"
		}
		dst := &MarkObject{
			Line: lines[3],
			Col:  1,
		}

		var newStart, newEnd *MarkObject
		result := TextMove(true, 2, markOne, markTwo, dst, &newStart, &newEnd)
		assert.True(t, result, "TextMove with count=2 should succeed")
		assert.NotNil(t, newStart, "newStart should be set")
		assert.NotNil(t, newEnd, "newEnd should be set")
	})
}

// TestTextInsertTpar tests the TextInsertTpar function
func TestTextInsertTpar(t *testing.T) {
	t.Run("InsertSimpleTpar", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		// Create a simple TParObject
		tpar := &TParObject{
			Len: 5,
			Dlm: 0,
			Str: NewBlankStrObject(5),
			Nxt: nil,
			Con: nil,
		}
		tpar.Str.Set(1, 'H')
		tpar.Str.Set(2, 'e')
		tpar.Str.Set(3, 'l')
		tpar.Str.Set(4, 'l')
		tpar.Str.Set(5, 'o')

		var equalsMark *MarkObject
		result := TextInsertTpar(tpar, mark, &equalsMark)
		assert.True(t, result, "TextInsertTpar should succeed")
		assert.NotNil(t, equalsMark, "Equals mark should be created")

		// Check content was inserted
		assert.Equal(t, byte('H'), line.Str.Get(1))
		assert.Equal(t, byte('e'), line.Str.Get(2))
		assert.Equal(t, byte('l'), line.Str.Get(3))
		assert.Equal(t, byte('l'), line.Str.Get(4))
		assert.Equal(t, byte('o'), line.Str.Get(5))
	})

	t.Run("InsertEmptyTpar", func(t *testing.T) {
		_, line := setupTestLineInFrame()

		mark := &MarkObject{
			Line: line,
			Col:  1,
		}

		tpar := &TParObject{
			Len: 0,
			Dlm: 0,
			Str: NewBlankStrObject(0),
			Nxt: nil,
			Con: nil,
		}

		var equalsMark *MarkObject
		result := TextInsertTpar(tpar, mark, &equalsMark)
		assert.True(t, result, "TextInsertTpar with empty tpar should succeed")
		assert.NotNil(t, equalsMark, "Equals mark should be created")
	})

	t.Run("InsertMultiLineTpar", func(t *testing.T) {
		frame, line := setupTestLineInFrame()

		// Set some initial content on the line
		line.Str.Assign("Hello World")
		line.Used = line.Str.Length(' ', MaxStrLen)

		var mark *MarkObject
		MarkCreate(line, 7, &mark) // Mark after space

		// Create a multi-line TParObject chain
		// First tpar (the head)
		tpar1 := &TParObject{
			Len: 6,
			Dlm: 0,
			Str: NewBlankStrObject(6),
			Nxt: nil,
			Con: nil,
		}
		tpar1.Str.Assign("Line1")
		tpar1.Len = tpar1.Str.Length(' ', MaxStrLen)

		// Second tpar - middle line
		tpar2 := &TParObject{
			Len: 5,
			Dlm: 0,
			Str: NewBlankStrObject(5),
			Nxt: nil,
			Con: nil,
		}
		tpar2.Str.Assign("Line2")
		tpar2.Len = tpar2.Str.Length(' ', MaxStrLen)

		// Third tpar - last line (the tail)
		tpar3 := &TParObject{
			Len: 5,
			Dlm: 0,
			Str: NewBlankStrObject(5),
			Nxt: nil,
			Con: nil,
		}
		tpar3.Str.Assign("Line3")
		tpar3.Len = tpar3.Str.Length(' ', MaxStrLen)

		// Link them: tpar1.Con -> tpar2, tpar2.Con -> tpar3
		tpar1.Con = tpar2
		tpar2.Con = tpar3

		var equalsMark *MarkObject
		result := TextInsertTpar(tpar1, mark, &equalsMark)
		assert.True(t, result, "TextInsertTpar with multi-line tpar should succeed")
		assert.NotNil(t, equalsMark, "Equals mark should be created")
		assert.True(t, frame.TextModified, "Frame should be marked as modified")

		// Verify multi-line insertion succeeded
		firstLine := equalsMark.Line
		assert.NotNil(t, firstLine, "First line should exist")
		assert.NotNil(t, firstLine.FLink, "Should have created new lines")
		assert.Greater(t, firstLine.Used, 0, "First line should have content")

		// Collect all line contents for verification
		currentLine := firstLine
		assert.Equal(t, "Hello Line1", currentLine.Str.Slice(1, firstLine.Used), "First line content should be 'Hello Line1'")
		assert.NotNil(t, currentLine.FLink, "First line should have a next line")
		currentLine = currentLine.FLink
		assert.Equal(t, "Line2", currentLine.Str.Slice(1, currentLine.Used), "Second line content should be 'Line2'")
		assert.NotNil(t, currentLine.FLink, "Second line should have a next line")
		currentLine = currentLine.FLink
		assert.Equal(t, "Line3World", currentLine.Str.Slice(1, currentLine.Used), "Third line content should be 'Line3World'")
	})
}
