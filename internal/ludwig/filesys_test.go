package ludwig

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestStrObject creates a StrObject of size MaxStrLen with the given
// content at the start and spaces filling the rest, matching how line
// buffers are allocated in the editor.
func newTestStrObject(content string) *StrObject {
	return NewStrObjectFrom(content)
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f, err := os.CreateTemp("", "filesys-test-*")
			require.NoError(t, err)
			defer os.Remove(f.Name())
			defer f.Close()

			fyle := &FileObject{
				Fd:    int(f.Fd()),
				Entab: tc.entab,
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
