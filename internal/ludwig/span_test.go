// Tests for functions in span.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for span tests

// createTestSpanStructure creates a basic frame and span structure for testing
func createTestSpanStructure() (*FrameObject, *GroupObject, *LineHdrObject) {
	frame := &FrameObject{
		SpaceLeft:  MaxSpace,
		SpaceLimit: MaxSpace,
	}

	// Create a group with some lines
	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: 1,
		NrLines:     5,
	}

	// Create lines
	var firstLine, lastLine *LineHdrObject
	LinesCreate(5, &firstLine, &lastLine)

	// Link lines to group
	line := firstLine
	for i := 0; i < 5; i++ {
		line.Group = group
		line.OffsetNr = i
		LineChangeLength(line, 50)
		line.Used = 20
		if i < 5 {
			line = line.FLink
		}
	}

	group.FirstLine = firstLine
	group.LastLine = lastLine
	frame.FirstGroup = group
	frame.LastGroup = group

	return frame, group, firstLine
}

// saveAndClearSpans saves the current span state and clears it
func saveAndClearSpans(t *testing.T) *SpanObject {
	t.Helper()
	origFirstSpan := FirstSpan
	FirstSpan = nil

	t.Cleanup(func() {
		FirstSpan = origFirstSpan
	})

	return origFirstSpan
}

// createTestSpan creates a span with the given name and marks
func createTestSpan(name string, markOne, markTwo *MarkObject) *SpanObject {
	span := &SpanObject{
		Name:    name,
		MarkOne: markOne,
		MarkTwo: markTwo,
		Frame:   nil,
		Code:    nil,
	}
	return span
}

// Tests for SpanFind

func TestSpanFind(t *testing.T) {
	saveAndClearSpans(t)

	t.Run("EmptySpanList", func(t *testing.T) {
		var ptr, oldp *SpanObject

		result := SpanFind("test", &ptr, &oldp)

		assert.False(t, result, "SpanFind should return false for empty span list")
		assert.Nil(t, ptr, "ptr should be nil")
		assert.Nil(t, oldp, "oldp should be nil")
	})

	t.Run("FindSingleSpan", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span := createTestSpan("test", mark1, mark2)
		FirstSpan = span

		var ptr, oldp *SpanObject
		result := SpanFind("test", &ptr, &oldp)

		assert.True(t, result, "SpanFind should return true")
		assert.Equal(t, span, ptr, "ptr should point to the span")
		assert.Nil(t, oldp, "oldp should be nil for first span")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("SpanNotFound", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span := createTestSpan("test", mark1, mark2)
		FirstSpan = span

		var ptr, oldp *SpanObject
		result := SpanFind("notfound", &ptr, &oldp)

		assert.False(t, result, "SpanFind should return false for non-existent span")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("FindInMiddleOfList", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create spans in alphabetical order
		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)
		span3 := createTestSpan("gamma", mark1, mark2)

		// Link them
		span1.FLink = span2
		span2.BLink = span1
		span2.FLink = span3
		span3.BLink = span2

		FirstSpan = span1

		var ptr, oldp *SpanObject
		result := SpanFind("beta", &ptr, &oldp)

		assert.True(t, result, "SpanFind should return true")
		assert.Equal(t, span2, ptr, "ptr should point to span2")
		assert.Equal(t, span1, oldp, "oldp should point to span1")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("FindLastInList", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)

		span1.FLink = span2
		span2.BLink = span1

		FirstSpan = span1

		var ptr, oldp *SpanObject
		result := SpanFind("beta", &ptr, &oldp)

		assert.True(t, result, "SpanFind should return true")
		assert.Equal(t, span2, ptr, "ptr should point to span2")
		assert.Equal(t, span1, oldp, "oldp should point to span1")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("SearchPastEnd", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)

		span1.FLink = span2
		span2.BLink = span1

		FirstSpan = span1

		var ptr, oldp *SpanObject
		result := SpanFind("zulu", &ptr, &oldp)

		assert.False(t, result, "SpanFind should return false for name after all spans")
		assert.Nil(t, ptr, "ptr should be nil")
		assert.Equal(t, span2, oldp, "oldp should point to last span in list")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("FindFrameSpan", func(t *testing.T) {
		frame, group, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		lastLine := group.LastLine
		MarkCreate(lastLine, 1, &mark2)

		// Create a frame span
		span := createTestSpan("framespan", mark1, mark2)
		span.Frame = frame

		FirstSpan = span

		var ptr, oldp *SpanObject
		result := SpanFind("framespan", &ptr, &oldp)

		assert.True(t, result, "SpanFind should return true for frame span")
		assert.Equal(t, span, ptr, "ptr should point to the span")
		// When finding a frame span, it should reset the frame's marks
		// The marks are recreated in place, so the pointers stay the same
		// but the marks should now point to FirstLine and LastLine of the frame
		assert.Equal(t, frame.FirstGroup.FirstLine, span.MarkOne.Line, "MarkOne should point to frame's first line")
		assert.Equal(t, frame.LastGroup.LastLine, span.MarkTwo.Line, "MarkTwo should point to frame's last line")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})
}

// Tests for SpanCreate

func TestSpanCreate(t *testing.T) {
	// Note: We can't easily mock ScreenMessage as it's a regular function,
	// not a variable. The tests will call it, but that's okay for unit tests.
	// In real usage, error messages would be displayed to the user.

	saveAndClearSpans(t)

	t.Run("CreateNewSpan", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		result := SpanCreate("newspan", mark1, mark2)

		assert.True(t, result, "SpanCreate should return true")
		assert.NotNil(t, FirstSpan, "FirstSpan should not be nil")
		assert.Equal(t, "newspan", FirstSpan.Name, "Span name should be 'newspan'")
		assert.Nil(t, FirstSpan.Frame, "Span Frame should be nil")
		assert.NotNil(t, FirstSpan.MarkOne, "MarkOne should be created")
		assert.NotNil(t, FirstSpan.MarkTwo, "MarkTwo should be created")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("CreateSpanWithCorrectOrder", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 5, &mark1)
		MarkCreate(firstLine, 15, &mark2)

		result := SpanCreate("ordered", mark1, mark2)

		assert.True(t, result, "SpanCreate should return true")
		assert.Equal(t, 5, FirstSpan.MarkOne.Col, "MarkOne col should be 5")
		assert.Equal(t, 15, FirstSpan.MarkTwo.Col, "MarkTwo col should be 15")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("CreateSpanWithReverseOrder", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 15, &mark1)
		MarkCreate(firstLine, 5, &mark2)

		result := SpanCreate("reversed", mark1, mark2)

		assert.True(t, result, "SpanCreate should return true")
		// Marks should be swapped to maintain correct order
		assert.Equal(t, 5, FirstSpan.MarkOne.Col, "MarkOne col should be 5 (swapped)")
		assert.Equal(t, 15, FirstSpan.MarkTwo.Col, "MarkTwo col should be 15 (swapped)")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("CreateSpanAcrossLines", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 10, &mark1)
		secondLine := firstLine.FLink
		MarkCreate(secondLine, 5, &mark2)

		result := SpanCreate("multiline", mark1, mark2)

		assert.True(t, result, "SpanCreate should return true")
		// Should keep marks in original order since first line < second line
		assert.Equal(t, firstLine, FirstSpan.MarkOne.Line, "MarkOne should be on first line")
		assert.Equal(t, secondLine, FirstSpan.MarkTwo.Line, "MarkTwo should be on second line")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("RedefineExistingSpan", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2, mark3, mark4 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create initial span
		SpanCreate("redefine", mark1, mark2)

		assert.NotNil(t, FirstSpan, "FirstSpan should not be nil")

		originalSpan := FirstSpan

		// Redefine with new marks
		MarkCreate(firstLine, 5, &mark3)
		MarkCreate(firstLine, 15, &mark4)

		result := SpanCreate("redefine", mark3, mark4)

		assert.True(t, result, "SpanCreate should return true for redefinition")
		assert.Equal(t, originalSpan, FirstSpan, "Should reuse the same span object")
		assert.Equal(t, 5, FirstSpan.MarkOne.Col, "MarkOne col should be updated to 5")
		assert.Equal(t, 15, FirstSpan.MarkTwo.Col, "MarkTwo col should be updated to 15")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("CannotRedefineFrame", func(t *testing.T) {
		frame, group, firstLine := createTestSpanStructure()

		var mark1, mark2, mark3, mark4 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		lastLine := group.LastLine
		MarkCreate(lastLine, 1, &mark2)

		// Create a frame span
		span := createTestSpan("framespan", mark1, mark2)
		span.Frame = frame
		FirstSpan = span

		// Try to redefine it
		MarkCreate(firstLine, 5, &mark3)
		MarkCreate(firstLine, 15, &mark4)

		result := SpanCreate("framespan", mark3, mark4)

		assert.False(t, result, "SpanCreate should return false when trying to redefine a frame")
		// Note: ScreenMessage will be called, but we can't easily verify it in tests

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("InsertSpanInAlphabeticalOrder", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create spans in non-alphabetical order
		SpanCreate("charlie", mark1, mark2)
		SpanCreate("alpha", mark1, mark2)
		SpanCreate("bravo", mark1, mark2)

		// Verify alphabetical order
		assert.Equal(t, "alpha", FirstSpan.Name, "First span should be 'alpha'")
		assert.Equal(t, "bravo", FirstSpan.FLink.Name, "Second span should be 'bravo'")
		assert.Equal(t, "charlie", FirstSpan.FLink.FLink.Name, "Third span should be 'charlie'")

		// Verify back links
		assert.Nil(t, FirstSpan.BLink, "First span BLink should be nil")
		assert.Equal(t, FirstSpan, FirstSpan.FLink.BLink, "Second span BLink should point to first")
		assert.Equal(t, FirstSpan.FLink, FirstSpan.FLink.FLink.BLink, "Third span BLink should point to second")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("CreateSpanSameLineEqualMarks", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 10, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		result := SpanCreate("equal", mark1, mark2)

		assert.True(t, result, "SpanCreate should return true even with equal marks")
		// When marks are equal, order doesn't matter but should still succeed
		assert.Equal(t, 10, FirstSpan.MarkOne.Col, "MarkOne col should be 10")
		assert.Equal(t, 10, FirstSpan.MarkTwo.Col, "MarkTwo col should be 10")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})
}

// Tests for SpanDestroy

func TestSpanDestroy(t *testing.T) {
	// Note: ScreenMessage will be called for frame destruction attempts,
	// but we can't easily mock it. That's okay for unit tests.

	saveAndClearSpans(t)

	t.Run("DestroySingleSpan", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span := createTestSpan("test", mark1, mark2)
		FirstSpan = span

		result := SpanDestroy(&span)

		assert.True(t, result, "SpanDestroy should return true")
		assert.Nil(t, span, "span pointer should be nil after destroy")
		assert.Nil(t, FirstSpan, "FirstSpan should be nil after destroying only span")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("CannotDestroyFrame", func(t *testing.T) {
		frame, group, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		lastLine := group.LastLine
		MarkCreate(lastLine, 1, &mark2)

		span := createTestSpan("framespan", mark1, mark2)
		span.Frame = frame
		FirstSpan = span

		result := SpanDestroy(&span)

		assert.False(t, result, "SpanDestroy should return false for frame span")
		assert.NotNil(t, span, "span pointer should not be nil when destruction fails")
		// Note: ScreenMessage will be called, but we can't easily verify it

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("DestroyFirstSpanInList", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)
		span3 := createTestSpan("gamma", mark1, mark2)

		span1.FLink = span2
		span2.BLink = span1
		span2.FLink = span3
		span3.BLink = span2

		FirstSpan = span1

		result := SpanDestroy(&span1)

		assert.True(t, result, "SpanDestroy should return true")
		assert.Nil(t, span1, "span1 pointer should be nil")
		assert.Equal(t, span2, FirstSpan, "FirstSpan should now be span2")
		assert.Nil(t, span2.BLink, "span2 BLink should be nil")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("DestroyMiddleSpanInList", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)
		span3 := createTestSpan("gamma", mark1, mark2)

		span1.FLink = span2
		span2.BLink = span1
		span2.FLink = span3
		span3.BLink = span2

		FirstSpan = span1

		result := SpanDestroy(&span2)

		assert.True(t, result, "SpanDestroy should return true")
		assert.Nil(t, span2, "span2 pointer should be nil")
		assert.Equal(t, span3, span1.FLink, "span1 should link to span3")
		assert.Equal(t, span1, span3.BLink, "span3 should link back to span1")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("DestroyLastSpanInList", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span1 := createTestSpan("alpha", mark1, mark2)
		span2 := createTestSpan("beta", mark1, mark2)

		span1.FLink = span2
		span2.BLink = span1

		FirstSpan = span1

		result := SpanDestroy(&span2)

		assert.True(t, result, "SpanDestroy should return true")
		assert.Nil(t, span2, "span2 pointer should be nil")
		assert.Nil(t, span1.FLink, "span1 FLink should be nil")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("DestroySpanWithCode", func(t *testing.T) {
		// This tests that CodeDiscard is called when span has code
		// We can't fully test CodeDiscard without more infrastructure,
		// but we can verify the span is destroyed successfully
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		span := createTestSpan("withcode", mark1, mark2)
		// Set Code to non-nil to trigger CodeDiscard path
		span.Code = &CodeHeader{}
		FirstSpan = span

		result := SpanDestroy(&span)

		assert.True(t, result, "SpanDestroy should return true even with code")
		assert.Nil(t, span, "span pointer should be nil")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})
}

// Integration tests

func TestSpanCreateFindDestroyIntegration(t *testing.T) {
	saveAndClearSpans(t)

	t.Run("CreateFindAndDestroy", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create a span
		result := SpanCreate("test", mark1, mark2)
		assert.True(t, result, "SpanCreate failed")

		// Find it
		var ptr, oldp *SpanObject
		result = SpanFind("test", &ptr, &oldp)
		assert.True(t, result, "SpanFind failed")
		assert.Equal(t, "test", ptr.Name, "Found wrong span")

		// Destroy it
		result = SpanDestroy(&ptr)
		assert.True(t, result, "SpanDestroy failed")

		// Verify it's gone
		result = SpanFind("test", &ptr, &oldp)
		assert.False(t, result, "Span should not be found after destruction")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})

	t.Run("MultipleSpanOperations", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create multiple spans
		SpanCreate("span1", mark1, mark2)
		SpanCreate("span2", mark1, mark2)
		SpanCreate("span3", mark1, mark2)

		// Find and verify middle span
		var ptr, oldp *SpanObject
		assert.True(t, SpanFind("span2", &ptr, &oldp), "SpanFind span2 failed")

		// Destroy middle span
		assert.True(t, SpanDestroy(&ptr), "SpanDestroy span2 failed")

		// Verify span1 and span3 still exist
		assert.True(t, SpanFind("span1", &ptr, &oldp), "span1 should still exist")
		assert.True(t, SpanFind("span3", &ptr, &oldp), "span3 should still exist")

		// Verify span2 is gone
		assert.False(t, SpanFind("span2", &ptr, &oldp), "span2 should not exist")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("RedefineAndDestroy", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2, mark3, mark4 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create span
		SpanCreate("redef", mark1, mark2)

		// Redefine it
		MarkCreate(firstLine, 5, &mark3)
		MarkCreate(firstLine, 15, &mark4)
		SpanCreate("redef", mark3, mark4)

		// Verify redefinition
		var ptr, oldp *SpanObject
		assert.True(t, SpanFind("redef", &ptr, &oldp), "SpanFind failed")
		assert.Equal(t, 5, ptr.MarkOne.Col, "Mark not updated, expected col 5")

		// Destroy it
		assert.True(t, SpanDestroy(&ptr), "SpanDestroy failed")

		// Verify destruction
		assert.False(t, SpanFind("redef", &ptr, &oldp), "Span should be destroyed")

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
	})
}

// Tests for span list ordering

func TestSpanListOrdering(t *testing.T) {
	saveAndClearSpans(t)

	t.Run("MaintainAlphabeticalOrder", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create spans in random order
		names := []string{"zebra", "alpha", "mike", "charlie", "tango"}
		for _, name := range names {
			SpanCreate(name, mark1, mark2)
		}

		// Verify alphabetical order
		span := FirstSpan
		var prevName string
		for span != nil {
			if prevName != "" {
				assert.GreaterOrEqual(t, span.Name, prevName, "Spans not in alphabetical order")
			}
			prevName = span.Name
			span = span.FLink
		}

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})

	t.Run("VerifyBidirectionalLinks", func(t *testing.T) {
		frame, _, firstLine := createTestSpanStructure()

		var mark1, mark2 *MarkObject
		MarkCreate(firstLine, 1, &mark1)
		MarkCreate(firstLine, 10, &mark2)

		// Create several spans
		SpanCreate("a", mark1, mark2)
		SpanCreate("b", mark1, mark2)
		SpanCreate("c", mark1, mark2)
		SpanCreate("d", mark1, mark2)

		// Verify forward links
		span := FirstSpan
		var prev *SpanObject
		for span != nil {
			assert.Equal(t, prev, span.BLink, "BLink mismatch for span %s", span.Name)
			prev = span
			span = span.FLink
		}

		// Clean up
		frame.FirstGroup = nil
		frame.LastGroup = nil
		FirstSpan = nil
	})
}
