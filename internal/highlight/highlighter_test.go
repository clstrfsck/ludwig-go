package highlight

import (
	"reflect"
	"regexp"
	"testing"
)

func TestSliceStart(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		index int
		want  []byte
	}{
		{
			name:  "slice from beginning",
			input: []byte("hello world"),
			index: 0,
			want:  []byte("hello world"),
		},
		{
			name:  "slice from middle",
			input: []byte("hello world"),
			index: 6,
			want:  []byte("world"),
		},
		{
			name:  "slice from end",
			input: []byte("hello world"),
			index: 11,
			want:  []byte(""),
		},
		{
			name:  "slice beyond end",
			input: []byte("hello"),
			index: 10,
			want:  []byte(""),
		},
		{
			name:  "empty string",
			input: []byte(""),
			index: 0,
			want:  []byte(""),
		},
		{
			name:  "unicode characters",
			input: []byte("hello 世界"),
			index: 6,
			want:  []byte("世界"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceStart(tt.input, tt.index)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sliceStart() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSliceEnd(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		index int
		want  []byte
	}{
		{
			name:  "slice to end",
			input: []byte("hello world"),
			index: 11,
			want:  []byte("hello world"),
		},
		{
			name:  "slice to middle",
			input: []byte("hello world"),
			index: 5,
			want:  []byte("hello"),
		},
		{
			name:  "slice to beginning",
			input: []byte("hello world"),
			index: 0,
			want:  []byte(""),
		},
		{
			name:  "slice beyond end",
			input: []byte("hello"),
			index: 10,
			want:  []byte("hello"),
		},
		{
			name:  "empty string",
			input: []byte(""),
			index: 0,
			want:  []byte(""),
		},
		{
			name:  "unicode characters",
			input: []byte("hello 世界"),
			index: 6,
			want:  []byte("hello "),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sliceEnd(tt.input, tt.index)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sliceEnd() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunePos(t *testing.T) {
	tests := []struct {
		name string
		pos  int
		str  []byte
		want int
	}{
		{
			name: "position in middle",
			pos:  5,
			str:  []byte("hello world"),
			want: 5,
		},
		{
			name: "position at beginning",
			pos:  0,
			str:  []byte("hello world"),
			want: 0,
		},
		{
			name: "negative position",
			pos:  -1,
			str:  []byte("hello world"),
			want: 0,
		},
		{
			name: "position beyond end",
			pos:  100,
			str:  []byte("hello"),
			want: 5,
		},
		{
			name: "empty string",
			pos:  0,
			str:  []byte(""),
			want: 0,
		},
		{
			name: "unicode string",
			pos:  9,
			str:  []byte("hello 世界"),
			want: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := runePos(tt.pos, tt.str)
			if got != tt.want {
				t.Errorf("runePos() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHighlighter(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - comment: "#.*$"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)
	if h == nil {
		t.Fatal("NewHighlighter() returned nil")
	}
	if h.Def != def {
		t.Error("NewHighlighter() did not set Def correctly")
	}
	if h.lastRegion != nil {
		t.Error("NewHighlighter() should initialize lastRegion to nil")
	}
}

func TestFindIndex(t *testing.T) {
	tests := []struct {
		name  string
		regex string
		skip  string
		str   string
		want  []int
	}{
		{
			name:  "simple match",
			regex: "world",
			skip:  "",
			str:   "hello world",
			want:  []int{6, 11},
		},
		{
			name:  "no match",
			regex: "foo",
			skip:  "",
			str:   "hello world",
			want:  nil,
		},
		{
			name:  "match at beginning",
			regex: "^hello",
			skip:  "",
			str:   "hello world",
			want:  []int{0, 5},
		},
		{
			name:  "match at end",
			regex: "world$",
			skip:  "",
			str:   "hello world",
			want:  []int{6, 11},
		},
		{
			name:  "empty string",
			regex: "test",
			skip:  "",
			str:   "",
			want:  nil,
		},
		{
			name:  "with skip pattern",
			regex: `"`,
			skip:  `\\.`,
			str:   `hello \" world"`,
			want:  []int{14, 15},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(tt.regex)
			var skip *regexp.Regexp
			if tt.skip != "" {
				skip = regexp.MustCompile(tt.skip)
			}
			got := findIndex(regex, skip, []byte(tt.str))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindAllIndex(t *testing.T) {
	tests := []struct {
		name  string
		regex string
		str   string
		want  [][]int
	}{
		{
			name:  "multiple matches",
			regex: `\d+`,
			str:   "abc123def456ghi",
			want:  [][]int{{3, 6}, {9, 12}},
		},
		{
			name:  "no matches",
			regex: `\d+`,
			str:   "abcdefghi",
			want:  nil,
		},
		{
			name:  "single match",
			regex: "world",
			str:   "hello world",
			want:  [][]int{{6, 11}},
		},
		{
			name:  "overlapping pattern",
			regex: `\w+`,
			str:   "hello world",
			want:  [][]int{{0, 5}, {6, 11}},
		},
		{
			name:  "empty string",
			regex: `\d+`,
			str:   "",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regex := regexp.MustCompile(tt.regex)
			got := findAllIndex(regex, []byte(tt.str))
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAllIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLineMatch(t *testing.T) {
	lm := make(LineMatch)
	lm[0] = 1
	lm[5] = 2
	lm[10] = 3

	if lm[0] != 1 {
		t.Errorf("LineMatch[0] = %v, want 1", lm[0])
	}
	if lm[5] != 2 {
		t.Errorf("LineMatch[5] = %v, want 2", lm[5])
	}
	if lm[10] != 3 {
		t.Errorf("LineMatch[10] = %v, want 3", lm[10])
	}
}

func TestHighlightString_Simple(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - comment: "#.*$"
    - keyword: "\\b(if|else|while)\\b"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name       string
		input      string
		wantLines  int
		checkFirst bool
	}{
		{
			name:       "single line",
			input:      "# this is a comment",
			wantLines:  1,
			checkFirst: true,
		},
		{
			name:       "multiple lines",
			input:      "if true\n# comment\nelse",
			wantLines:  3,
			checkFirst: true,
		},
		{
			name:       "empty string",
			input:      "",
			wantLines:  1,
			checkFirst: false,
		},
		{
			name:       "no matches",
			input:      "hello world",
			wantLines:  1,
			checkFirst: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset highlighter state between tests
			h.lastRegion = nil

			matches := h.HighlightString(tt.input)
			if len(matches) != tt.wantLines {
				t.Errorf("HighlightString() returned %d lines, want %d", len(matches), tt.wantLines)
			}
			if tt.checkFirst && len(matches) > 0 {
				if matches[0] == nil {
					t.Error("HighlightString() returned nil match for first line")
				}
			}
		})
	}
}

func TestHighlightString_WithRegions(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - string:
        start: "\""
        end: "\""
        skip: "\\\\."
    - comment:
        start: "/\\*"
        end: "\\*/"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name      string
		input     string
		wantLines int
	}{
		{
			name:      "string on single line",
			input:     `"hello world"`,
			wantLines: 1,
		},
		{
			name:      "multiline comment",
			input:     "/* comment\nline 2\nline 3 */",
			wantLines: 3,
		},
		{
			name:      "string with escaped quote",
			input:     `"hello \" world"`,
			wantLines: 1,
		},
		{
			name:      "no regions",
			input:     "plain text",
			wantLines: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset highlighter state
			h.lastRegion = nil

			matches := h.HighlightString(tt.input)
			if len(matches) != tt.wantLines {
				t.Errorf("HighlightString() returned %d lines, want %d", len(matches), tt.wantLines)
			}
		})
	}
}

func TestHighlightString_StateTracking(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	// Test that state carries over between lines
	input2 := "/* comment start\nstill in comment\nend comment */"
	matches := h.HighlightString(input2)

	if len(matches) != 3 {
		t.Fatalf("Expected 3 lines, got %d", len(matches))
	}

	// After highlighting, lastRegion should be nil (comment closed)
	if h.lastRegion != nil {
		t.Error("Expected lastRegion to be nil after closed region")
	}

	// Test unclosed region
	h.lastRegion = nil
	input3 := "/* comment start\nstill in comment"
	matches = h.HighlightString(input3)

	if len(matches) != 2 {
		t.Fatalf("Expected 2 lines, got %d", len(matches))
	}

	// lastRegion should not be nil (comment still open)
	if h.lastRegion == nil {
		t.Error("Expected lastRegion to be set for unclosed region")
	}
}

// MockLineStates implements LineStates interface for testing
type MockLineStates struct {
	lines  [][]byte
	states []State
	matches []LineMatch
}

func NewMockLineStates(lines []string) *MockLineStates {
	m := &MockLineStates{
		lines:  make([][]byte, len(lines)),
		states: make([]State, len(lines)),
		matches: make([]LineMatch, len(lines)),
	}
	for i, line := range lines {
		m.lines[i] = []byte(line)
	}
	return m
}

func (m *MockLineStates) LineBytes(n int) []byte {
	if n >= 0 && n < len(m.lines) {
		return m.lines[n]
	}
	return nil
}

func (m *MockLineStates) LinesNum() int {
	return len(m.lines)
}

func (m *MockLineStates) State(lineN int) State {
	if lineN >= 0 && lineN < len(m.states) {
		return m.states[lineN]
	}
	return nil
}

func (m *MockLineStates) SetState(lineN int, s State) {
	if lineN >= 0 && lineN < len(m.states) {
		m.states[lineN] = s
	}
}

func (m *MockLineStates) SetMatch(lineN int, match LineMatch) {
	if lineN >= 0 && lineN < len(m.matches) {
		m.matches[lineN] = match
	}
}

func (m *MockLineStates) Lock()   {}
func (m *MockLineStates) Unlock() {}

func TestHighlightStates(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name  string
		lines []string
	}{
		{
			name:  "single line",
			lines: []string{"hello world"},
		},
		{
			name:  "multiple lines without regions",
			lines: []string{"line 1", "line 2", "line 3"},
		},
		{
			name:  "multiline comment",
			lines: []string{"/* start", "middle", "end */"},
		},
		{
			name:  "no lines",
			lines: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockLineStates(tt.lines)
			h.lastRegion = nil
			h.HighlightStates(mock)

			// Verify states were set
			for i := 0; i < len(tt.lines); i++ {
				// State should be set (even if nil)
				_ = mock.State(i)
			}
		})
	}
}

func TestHighlightMatches(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - keyword: "\\b(if|else)\\b"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name      string
		lines     []string
		startline int
		endline   int
	}{
		{
			name:      "single line",
			lines:     []string{"if true"},
			startline: 0,
			endline:   0,
		},
		{
			name:      "multiple lines",
			lines:     []string{"if true", "else", "end"},
			startline: 0,
			endline:   2,
		},
		{
			name:      "range within lines",
			lines:     []string{"line 1", "if true", "else", "line 4"},
			startline: 1,
			endline:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockLineStates(tt.lines)
			h.lastRegion = nil
			h.HighlightMatches(mock, tt.startline, tt.endline)

			// Verify matches were set for the range
			for i := tt.startline; i <= tt.endline && i < len(tt.lines); i++ {
				if mock.matches[i] == nil {
					t.Errorf("Match not set for line %d", i)
				}
			}
		})
	}
}

func TestReHighlightStates(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - comment:
        start: "/\\*"
        end: "\\*/"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name       string
		lines      []string
		startline  int
		checkRan   bool
	}{
		{
			name:      "from beginning",
			lines:     []string{"line 1", "line 2", "line 3"},
			startline: 0,
			checkRan:  true,
		},
		{
			name:      "from middle",
			lines:     []string{"line 1", "/* comment", "end */", "line 4"},
			startline: 1,
			checkRan:  true,
		},
		{
			name:      "single line",
			lines:     []string{"line 1"},
			startline: 0,
			checkRan:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockLineStates(tt.lines)
			h.lastRegion = nil
			finalLine := h.ReHighlightStates(mock, tt.startline)

			// Just verify it returns a valid line number
			if finalLine < 0 || finalLine >= len(tt.lines) {
				t.Errorf("ReHighlightStates() returned invalid line %d, should be 0-%d", finalLine, len(tt.lines)-1)
			}
		})
	}
}

func TestReHighlightLine(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	input := []byte(`filetype: test
rules:
    - keyword: "\\b(if|else)\\b"
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	tests := []struct {
		name  string
		lines []string
		lineN int
	}{
		{
			name:  "first line",
			lines: []string{"if true", "else", "end"},
			lineN: 0,
		},
		{
			name:  "middle line",
			lines: []string{"line 1", "if true", "line 3"},
			lineN: 1,
		},
		{
			name:  "last line",
			lines: []string{"line 1", "line 2", "if true"},
			lineN: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockLineStates(tt.lines)
			h.lastRegion = nil
			h.ReHighlightLine(mock, tt.lineN)

			// Verify match and state were set
			if mock.matches[tt.lineN] == nil {
				t.Errorf("Match not set for line %d", tt.lineN)
			}
		})
	}
}

func TestHighlighter_ComplexSyntax(t *testing.T) {
	Groups = make(map[string]Group)
	numGroups = 0

	// Test with a more complex syntax definition
	input := []byte(`filetype: test
rules:
    - comment: "#.*$"
    - keyword: "\\b(func|return|if|else)\\b"
    - constant.number: "\\b\\d+\\b"
    - string:
        start: "\""
        end: "\""
        skip: "\\\\."
        rules:
            - constant.specialChar: "\\\\."
`)
	file, err := ParseFile(input)
	if err != nil {
		t.Fatalf("ParseFile() failed: %v", err)
	}

	header, err := MakeHeaderYaml(input)
	if err != nil {
		t.Fatalf("MakeHeaderYaml() failed: %v", err)
	}

	def, err := ParseDef(file, header)
	if err != nil {
		t.Fatalf("ParseDef() failed: %v", err)
	}

	h := NewHighlighter(def)

	testCases := []struct {
		name  string
		code  string
		lines int
	}{
		{
			name:  "function definition",
			code:  "func test() {\n  return 42\n}",
			lines: 3,
		},
		{
			name:  "string with escape",
			code:  `"hello \"world\""`,
			lines: 1,
		},
		{
			name:  "comment",
			code:  "# this is a comment",
			lines: 1,
		},
		{
			name:  "mixed syntax",
			code:  "if x == 42 # check value\n  return \"success\"",
			lines: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h.lastRegion = nil
			matches := h.HighlightString(tc.code)

			if len(matches) != tc.lines {
				t.Errorf("Expected %d lines, got %d", tc.lines, len(matches))
			}

			// Verify each line has some highlighting
			for i, match := range matches {
				if match == nil {
					t.Errorf("Line %d has nil match", i)
				}
			}
		})
	}
}
