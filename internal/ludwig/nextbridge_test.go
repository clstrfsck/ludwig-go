// Tests for nextbridge.go

package ludwig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// makeCharSet builds a [256]bool from the given characters.
func makeCharSet(chars ...byte) *[256]bool {
	var set [256]bool
	for _, ch := range chars {
		set[ch] = true
	}
	return &set
}

func TestSearchForward(t *testing.T) {
	t.Run("Match at start of line", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")

		line, col := searchForward(makeCharSet('h'), lines[0], 1)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 1, col)
	})

	t.Run("Match mid-line skips earlier chars", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")

		// First 'l' is at col 3
		line, col := searchForward(makeCharSet('l'), lines[0], 1)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 3, col)
	})

	t.Run("Match on next line when not found on current", func(t *testing.T) {
		_, lines := setupTestFrame(2)
		setLineContent(lines[0], "aaa")
		setLineContent(lines[1], "bbb")

		line, col := searchForward(makeCharSet('b'), lines[0], 1)

		assert.Equal(t, lines[1], line)
		assert.Equal(t, 1, col)
	})

	t.Run("Space matches at EOL position", func(t *testing.T) {
		// "abc" has Used=3; EOL sentinel is at col 4 (Used+1)
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc")

		line, col := searchForward(makeCharSet(' '), lines[0], 1)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 4, col)
	})

	t.Run("Non-space does not match at EOL position", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc")

		line, col := searchForward(makeCharSet('x'), lines[0], 1)

		assert.Nil(t, line)
		assert.Equal(t, 0, col)
	})

	t.Run("Search starting past used still matches EOL space", func(t *testing.T) {
		// col=4 is exactly Used+1; space should match there
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc") // Used=3

		line, col := searchForward(makeCharSet(' '), lines[0], 4)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 4, col)
	})
}

func TestSearchBackward(t *testing.T) {
	t.Run("Match on same line", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")

		// Rightmost 'l' is at col 4
		line, col := searchBackward(makeCharSet('l'), lines[0], 5, false)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 4, col)
	})

	t.Run("Match on previous line when not found on current", func(t *testing.T) {
		_, lines := setupTestFrame(2)
		setLineContent(lines[0], "abc")
		setLineContent(lines[1], "xyz")

		// 'b' is at col 2 in lines[0]
		line, col := searchBackward(makeCharSet('b'), lines[1], 1, false)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 2, col)
	})

	t.Run("Col past EOL with space in set returns immediately", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc") // Used=3

		// col=5 > Used; space is in set → return (line, 5) immediately
		line, col := searchBackward(makeCharSet(' '), lines[0], 5, false)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 5, col)
	})

	t.Run("Col past EOL without space adjusts col to used", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc") // Used=3

		// col=5 > Used; space not in set → col adjusted to 3; 'c' at col 3
		line, col := searchBackward(makeCharSet('c'), lines[0], 5, false)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 3, col)
	})

	t.Run("Not found with bridge false returns nil", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc")

		line, col := searchBackward(makeCharSet('x'), lines[0], 3, false)

		assert.Nil(t, line)
		assert.Equal(t, 0, col)
	})

	t.Run("Not found with bridge true at first line returns current line", func(t *testing.T) {
		_, lines := setupTestFrame(1)
		setLineContent(lines[0], "abc")

		// BLink is nil; bridge=true → returns (line, col) as sentinel
		line, col := searchBackward(makeCharSet('x'), lines[0], 3, true)

		assert.Equal(t, lines[0], line)
		assert.Equal(t, 3, col)
	})
}

func TestNextbridgeCommand(t *testing.T) {
	oldCurrentFrame := CurrentFrame
	defer func() { CurrentFrame = oldCurrentFrame }()

	t.Run("Count zero sets MarkEquals and returns true", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")
		frame.Dot = &MarkObject{Line: lines[0], Col: 3}
		CurrentFrame = frame

		ok := NextbridgeCommand(0, createTestTpar("x"), false)

		assert.True(t, ok)
		assert.NotNil(t, frame.Marks[MarkEquals])
		assert.Equal(t, lines[0], frame.Marks[MarkEquals].Line)
		assert.Equal(t, 3, frame.Marks[MarkEquals].Col)
	})

	t.Run("Count one forward moves dot to matching char", func(t *testing.T) {
		// "hello world": space at col 6
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello world")
		frame.Dot = &MarkObject{Line: lines[0], Col: 1}
		CurrentFrame = frame

		ok := NextbridgeCommand(1, createTestTpar(" "), false)

		assert.True(t, ok)
		assert.Equal(t, 6, frame.Dot.Col)
	})

	t.Run("Count two forward moves dot to second occurrence", func(t *testing.T) {
		// "hello world test": spaces at col 6 and col 12
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello world test")
		frame.Dot = &MarkObject{Line: lines[0], Col: 1}
		CurrentFrame = frame

		ok := NextbridgeCommand(2, createTestTpar(" "), false)

		assert.True(t, ok)
		assert.Equal(t, 12, frame.Dot.Col)
	})

	t.Run("Count one forward not found returns false", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")
		frame.Dot = &MarkObject{Line: lines[0], Col: 1}
		CurrentFrame = frame

		ok := NextbridgeCommand(1, createTestTpar("x"), false)

		assert.False(t, ok)
	})

	t.Run("Count one backward lands one past the found char", func(t *testing.T) {
		// "hello world": space at col 6; searching from col 8 ('o')
		// finds space at col 6, lands at col 7 ('w' = start of word)
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello world")
		frame.Dot = &MarkObject{Line: lines[0], Col: 8}
		CurrentFrame = frame

		ok := NextbridgeCommand(-1, createTestTpar(" "), false)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col)
	})

	t.Run("Count one backward not found returns false", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")
		frame.Dot = &MarkObject{Line: lines[0], Col: 3}
		CurrentFrame = frame

		ok := NextbridgeCommand(-1, createTestTpar("x"), false)

		assert.False(t, ok)
	})

	t.Run("Bridge inverts char set to find first non-member", func(t *testing.T) {
		// bridge=true with "aeiou": searches for non-vowels
		// "hello", dot at col 2 ('e'): 'e' is vowel (skip), 'l' at col 3 is consonant (match)
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello")
		frame.Dot = &MarkObject{Line: lines[0], Col: 2}
		CurrentFrame = frame

		ok := NextbridgeCommand(1, createTestTpar("aeiou"), true)

		assert.True(t, ok)
		assert.Equal(t, 3, frame.Dot.Col)
	})

	t.Run("Range syntax a..z covers all lowercase letters", func(t *testing.T) {
		// "HELLO world": first lowercase is 'w' at col 7
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "HELLO world")
		frame.Dot = &MarkObject{Line: lines[0], Col: 1}
		CurrentFrame = frame

		ok := NextbridgeCommand(1, createTestTpar("a..z"), false)

		assert.True(t, ok)
		assert.Equal(t, 7, frame.Dot.Col)
	})

	t.Run("MarkEquals records dot position before move", func(t *testing.T) {
		frame, lines := setupTestFrame(1)
		setLineContent(lines[0], "hello world")
		frame.Dot = &MarkObject{Line: lines[0], Col: 1}
		CurrentFrame = frame

		ok := NextbridgeCommand(1, createTestTpar(" "), false)

		assert.True(t, ok)
		assert.NotNil(t, frame.Marks[MarkEquals])
		assert.Equal(t, 1, frame.Marks[MarkEquals].Col)
		assert.Equal(t, 6, frame.Dot.Col)
	})
}
