package ludwig

import (
	"bufio"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newReadFyle creates a FileObject backed by a temp file containing content.
func newReadFyle(t *testing.T, content string) *FileObject {
	t.Helper()
	f, err := os.CreateTemp("", "filesys-read-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		f.Close()
		os.Remove(f.Name())
	})
	_, err = f.WriteString(content)
	require.NoError(t, err)
	_, err = f.Seek(0, 0)
	require.NoError(t, err)
	return &FileObject{OsFile: f, Reader: bufio.NewReader(f)}
}

// readLine reads one line via FilesysRead into a fresh MaxStrLen buffer.
func readLine(fyle *FileObject) (string, int, bool) {
	buf := NewBlankStrObject(MaxStrLen)
	var outlen int
	oldTabs := FileData.TabWidth
	defer func() {
		FileData.TabWidth = oldTabs
	}()
	FileData.TabWidth = 8
	ok := FilesysRead(fyle, buf, &outlen)
	return buf.Slice(1, outlen), outlen, ok
}

// newTestStrObject creates a StrObject of size MaxStrLen with the given
// content at the start and spaces filling the rest, matching how line
// buffers are allocated in the editor.
func newTestStrObject(content string) *StrObject {
	return NewStrObjectFrom(content)
}

func TestFilesysRead(t *testing.T) {
	t.Run("empty file returns false and sets Eof", func(t *testing.T) {
		fyle := newReadFyle(t, "")
		buf := NewBlankStrObject(MaxStrLen)
		var outlen int
		ok := FilesysRead(fyle, buf, &outlen)
		assert.False(t, ok)
		assert.True(t, fyle.Eof)
		assert.Equal(t, 0, outlen)
	})

	t.Run("single line no newline returns content and sets Eof", func(t *testing.T) {
		fyle := newReadFyle(t, "hello")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, "hello", line)
		assert.Equal(t, 5, outlen)
		assert.True(t, fyle.Eof)
		// next read returns false
		_, _, ok2 := readLine(fyle)
		assert.False(t, ok2)
	})

	t.Run("single line with newline", func(t *testing.T) {
		fyle := newReadFyle(t, "hello\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, "hello", line)
		assert.Equal(t, 5, outlen)
	})

	t.Run("multiple lines", func(t *testing.T) {
		fyle := newReadFyle(t, "foo\nbar\nbaz\n")
		lines := []string{"foo", "bar", "baz"}
		for i, want := range lines {
			line, _, ok := readLine(fyle)
			assert.True(t, ok, "line %d should succeed", i+1)
			assert.Equal(t, want, line, "line %d content", i+1)
		}
		_, _, ok := readLine(fyle)
		assert.False(t, ok, "should be EOF after last line")
		assert.Equal(t, 3, fyle.LCounter)
	})

	t.Run("LCounter increments per line", func(t *testing.T) {
		fyle := newReadFyle(t, "a\nb\n")
		readLine(fyle)
		assert.Equal(t, 1, fyle.LCounter)
		readLine(fyle)
		assert.Equal(t, 2, fyle.LCounter)
	})

	t.Run("tab expansion to 8-column boundary", func(t *testing.T) {
		// \t at position 0 → 8 spaces; \t at position 3 → 5 spaces
		fyle := newReadFyle(t, "\tabc\tX\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		// col 0: tab → 8 spaces; "abc" at cols 8-10; tab at col 11 → 5 spaces to col 16; "X"
		assert.Equal(t, strings.Repeat(" ", 8)+"abc"+strings.Repeat(" ", 5)+"X", line)
		assert.Equal(t, 17, outlen)
	})

	t.Run("carriage return terminates line", func(t *testing.T) {
		fyle := newReadFyle(t, "hi\rworld\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, "hi", line)
		assert.Equal(t, 2, outlen)
	})

	t.Run("control characters except tab and newlines are skipped", func(t *testing.T) {
		// \x01 (SOH), \x07 (BEL) are non-printable control chars, should be dropped
		fyle := newReadFyle(t, "a\x01b\x07c\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, "abc", line)
		assert.Equal(t, 3, outlen)
	})

	t.Run("unicode characters are read correctly", func(t *testing.T) {
		fyle := newReadFyle(t, "héllo wörld\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, "héllo wörld", line)
		assert.Equal(t, len("héllo wörld"), outlen)
	})

	t.Run("line truncated at MaxStrLen", func(t *testing.T) {
		long := strings.Repeat("x", MaxStrLen+10)
		fyle := newReadFyle(t, long+"\n")
		line, outlen, ok := readLine(fyle)
		assert.True(t, ok)
		assert.Equal(t, MaxStrLen, outlen)
		assert.Equal(t, strings.Repeat("x", MaxStrLen), line)
	})
}

func TestFilesysWriteEntab(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		bufsiz   int
		entab    bool
		expected string
	}{
		{
			name:     "no entab passthrough",
			input:    "        hello",
			bufsiz:   13,
			entab:    false,
			expected: "        hello\n",
		},
		{
			name:     "entab replaces 8 leading spaces with tab",
			input:    "        hello",
			bufsiz:   13,
			entab:    true,
			expected: "\thello\n",
		},
		{
			name:     "entab replaces 16 leading spaces with two tabs",
			input:    "                hello",
			bufsiz:   21,
			entab:    true,
			expected: "\t\thello\n",
		},
		{
			name:     "entab with 7 leading spaces does nothing",
			input:    "       hello",
			bufsiz:   12,
			entab:    true,
			expected: "       hello\n",
		},
		{
			name:     "entab with no leading spaces does nothing",
			input:    "hello",
			bufsiz:   5,
			entab:    true,
			expected: "hello\n",
		},
	}

	oldTabs := FileData.TabWidth
	defer func() {
		FileData.TabWidth = oldTabs
	}()
	FileData.TabWidth = 8

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "filesys-test-*")
			require.NoError(t, err)
			defer os.Remove(f.Name())
			defer f.Close()

			fyle := &FileObject{
				OsFile: f,
				Entab:  tc.entab,
			}
			buf := newTestStrObject(tc.input)
			ok := FilesysWrite(fyle, buf, tc.bufsiz)
			assert.True(t, ok, "FilesysWrite should succeed")

			_, err = f.Seek(0, 0)
			require.NoError(t, err)
			data, err := io.ReadAll(f)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, string(data))
		})
	}
}
