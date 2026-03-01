package ludwig

import (
	"sort"

	"ludwig-go/internal/highlight"
)

// syntaxColorTable maps highlight Group IDs to ncurses color pair numbers
var syntaxColorTable [256]int

// buildColorTable maps micro highlight group names to ncurses color pairs
func buildColorTable() {
	groupColors := map[string]int{
		"statement":  ColorPairKeyword,
		"keyword":    ColorPairKeyword,
		"type":       ColorPairType,
		"string":     ColorPairString,
		"stringx":    ColorPairString,
		"comment":    ColorPairComment,
		"preproc":    ColorPairPreproc,
		"constant":   ColorPairConstant,
		"special":    ColorPairSpecial,
		"underlined": ColorPairSpecial,
		"todo":       ColorPairSpecial,
		"error":      ColorPairError,
	}
	for name, pair := range groupColors {
		if g, ok := highlight.Groups[name]; ok {
			syntaxColorTable[g] = pair
		}
	}
}

var (
	syntaxDefs    []*highlight.Def
	syntaxHeaders []*highlight.Header
	syntaxEnabled bool
)

// SyntaxInit loads all embedded syntax definitions
func SyntaxInit() {
	if !colorsEnabled {
		return
	}
	entries, err := highlight.SyntaxFiles.ReadDir("syntax")
	if err != nil {
		return
	}
	var files []*highlight.File
	for _, entry := range entries {
		data, err := highlight.SyntaxFiles.ReadFile("syntax/" + entry.Name())
		if err != nil {
			continue
		}
		hdr, err := highlight.MakeHeaderYaml(data)
		if err != nil {
			continue
		}
		f, err := highlight.ParseFile(data)
		if err != nil {
			continue
		}
		def, err := highlight.ParseDef(f, hdr)
		if err != nil {
			continue
		}
		syntaxDefs = append(syntaxDefs, def)
		syntaxHeaders = append(syntaxHeaders, hdr)
		files = append(files, f)
	}
	// Resolve cross-file includes
	for i, def := range syntaxDefs {
		if highlight.HasIncludes(def) {
			highlight.ResolveIncludes(syntaxDefs[i], files)
		}
	}
	buildColorTable()
	syntaxEnabled = len(syntaxDefs) > 0
}

// SyntaxDetect returns a Highlighter for the given filename, or nil if none matches
func SyntaxDetect(filename string) *highlight.Highlighter {
	if !syntaxEnabled {
		return nil
	}
	for i, hdr := range syntaxHeaders {
		if hdr.MatchFileName(filename) {
			return highlight.NewHighlighter(syntaxDefs[i])
		}
	}
	return nil
}

// LudwigBuffer implements highlight.LineStates over a frame's line list
type LudwigBuffer struct {
	lines []*LineHdrObject
}

func newLudwigBuffer(frame *FrameObject) *LudwigBuffer {
	var lines []*LineHdrObject
	line := frame.FirstGroup.FirstLine
	for line != nil && line.FLink != nil { // FLink==nil means EOP
		lines = append(lines, line)
		line = line.FLink
	}
	return &LudwigBuffer{lines: lines}
}

func (b *LudwigBuffer) LineBytes(n int) []byte {
	if n >= len(b.lines) {
		return nil
	}
	line := b.lines[n]
	if line.Str == nil || line.Used == 0 {
		return nil
	}
	return []byte(line.Str.Slice(1, line.Used))
}

func (b *LudwigBuffer) LinesNum() int { return len(b.lines) }

func (b *LudwigBuffer) State(n int) highlight.State {
	if n >= len(b.lines) {
		return nil
	}
	if s, ok := b.lines[n].HlState.(highlight.State); ok {
		return s
	}
	return nil
}

func (b *LudwigBuffer) SetState(n int, s highlight.State) {
	if n < len(b.lines) {
		b.lines[n].HlState = s
	}
}

func (b *LudwigBuffer) SetMatch(n int, m highlight.LineMatch) {
	if n < len(b.lines) {
		b.lines[n].HlMatch = m
	}
}

func (b *LudwigBuffer) Lock()   {} // Ludwig is single-threaded
func (b *LudwigBuffer) Unlock() {}

// SyntaxHighlightFrame runs a full highlight pass on a frame
func SyntaxHighlightFrame(frame *FrameObject, h *highlight.Highlighter) {
	buf := newLudwigBuffer(frame)
	h.HighlightStates(buf)
	h.HighlightMatches(buf, 0, buf.LinesNum())
}

// SyntaxRehighlightFrom re-highlights incrementally from a changed line
func SyntaxRehighlightFrom(frame *FrameObject, h *highlight.Highlighter, fromLine int) {
	buf := newLudwigBuffer(frame)
	end := h.ReHighlightStates(buf, fromLine)
	h.HighlightMatches(buf, fromLine, end+1)
}

// per-frame highlighter storage
var frameHighlighters = map[*FrameObject]*highlight.Highlighter{}

// SyntaxAttach attaches a highlighter to a frame
func SyntaxAttach(frame *FrameObject, h *highlight.Highlighter) {
	frameHighlighters[frame] = h
}

// SyntaxGet returns the highlighter for a frame (may be nil)
func SyntaxGet(frame *FrameObject) *highlight.Highlighter {
	return frameHighlighters[frame]
}

// syntaxDrawLine renders line content with syntax highlighting.
// offset and strlen are already computed by ScreenDrawLine.
func syntaxDrawLine(line *LineHdrObject, offset, strlen int) {
	match, ok := line.HlMatch.(highlight.LineMatch)
	if !ok || len(match) == 0 {
		VduDisplayStr(line.Str.Slice(offset+1, strlen), 3)
		return
	}

	// Collect and sort the match positions
	positions := make([]int, 0, len(match))
	for pos := range match {
		positions = append(positions, pos)
	}
	sort.Ints(positions)

	// Walk the line segment by segment, switching colors
	col := offset // current rune column (0-based)
	end := offset + strlen
	currentPair := 0

	for _, pos := range positions {
		if pos >= end {
			break
		}
		group := match[pos]
		newPair := syntaxColorTable[group]

		// Render any gap before this position
		if pos > col {
			segLen := pos - col
			if col+segLen > end {
				segLen = end - col
			}
			VduDisplayStr(line.Str.Slice(col+1, segLen), 0)
			col += segLen
		}

		// Switch color
		if currentPair != 0 {
			VduColorOff(currentPair)
		}
		currentPair = newPair
		if currentPair != 0 {
			VduColorOn(currentPair)
		}
	}

	// Render the remaining tail
	if col < end {
		VduDisplayStr(line.Str.Slice(col+1, end-col), 0)
	}
	if currentPair != 0 {
		VduColorOff(currentPair)
	}
}
