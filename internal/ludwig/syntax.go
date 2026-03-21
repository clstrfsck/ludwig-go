/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         SYNTAX
//
// Description:  This module handles syntax highlighting using ncurses.

package ludwig

import (
	"strings"

	"ludwig-go/internal/highlight"
	nc "ludwig-go/internal/ncurses"
)

// syntaxColorTable maps highlight Group IDs to ncurses color pair numbers
var syntaxColorLookup map[highlight.Group]int

// buildColorTable maps highlight group names to ncurses color pairs
func buildColorTable(groupColours map[string]int) {
	if possibleDefault, ok := groupColours["default"]; ok {
		defaultPair = possibleDefault
	}

	syntaxColorLookup = make(map[highlight.Group]int)
	for name, pair := range groupColours {
		if g, ok := highlight.Groups[name]; ok {
			syntaxColorLookup[g] = pair
		}
	}

	// Pre-compute hierarchical lookups for all groups
	for name, groupID := range highlight.Groups {
		if syntaxColorLookup[groupID] == 0 {
			syntaxColorLookup[groupID] = findColorForGroup(name)
		}
	}
}

func findColorForGroup(name string) int {
	for name != "" {
		if g, ok := highlight.Groups[name]; ok {
			if id := syntaxColorLookup[g]; id != 0 {
				return id
			}
		}
		if idx := strings.LastIndex(name, "."); idx != -1 {
			name = name[:idx]
		} else {
			break
		}
	}
	return 0
}

var (
	syntaxDefs    []*highlight.Def
	syntaxHeaders []*highlight.Header
	syntaxEnabled bool
	defaultPair   int = 0
)

// SyntaxInit loads all embedded syntax definitions
func SyntaxInit(groupColors map[string]int) {
	if groupColors == nil {
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
	buildColorTable(groupColors)
	syntaxEnabled = len(syntaxDefs) > 0
}

// SyntaxDetectFilename returns a Highlighter for the given filename, or nil if none matches
func SyntaxDetectFilename(filename string) *highlight.Highlighter {
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

// SyntaxDetectHeader returns a Highlighter for the given header, or nil if none matches
func SyntaxDetectHeader(header []byte) *highlight.Highlighter {
	if !syntaxEnabled {
		return nil
	}
	for i, hdr := range syntaxHeaders {
		if hdr.MatchFileHeader(header) {
			return highlight.NewHighlighter(syntaxDefs[i])
		}
	}
	return nil
}

// SetupSyntaxHighlighting sets up syntax highlighting for a file based on the
// extension of the file, or if no match based on the contents of an optionally
// specified first line of a file.
func SetupSyntaxHighlighting(frame *FrameObject) {
	if !syntaxEnabled {
		return
	}

	if frame.InputFile <= 0 {
		clearFrameHighlighting(frame)
		return
	}

	// If a highlighter is already attached, we should just re-apply it unless
	// we are at the start of a file.  This avoids changing or detaching the
	// highlighter when paging changes the first visible line, but also allows
	// for loading new files into the frame.
	inputFile := Files[frame.InputFile]
	if frame.Highlighter != nil {
		lastLine := frame.LastGroup.LastLine
		if LineToNumber(lastLine) < inputFile.LCounter+1 {
			SyntaxHighlightFrame(frame, frame.Highlighter)
			return
		}
	}

	filename := inputFile.Filename
	if filename != "" {
		if h := SyntaxDetectFilename(filename); h != nil {
			frame.Highlighter = h
			SyntaxHighlightFrame(frame, h)
			return
		}
	}
	firstLine := frame.FirstGroup.FirstLine
	if firstLine.Str != nil {
		lineContents := firstLine.Str.Slice(1, firstLine.Used)
		if h := SyntaxDetectHeader([]byte(lineContents)); h != nil {
			frame.Highlighter = h
			SyntaxHighlightFrame(frame, h)
			return
		}
	}

	// No highlighter.  Make sure we remove existing.
	clearFrameHighlighting(frame)
}

// clearFrameHighlighting removes any per-line syntax highlighting state from
// all lines in the given frame. This is used when detaching a highlighter to
// ensure stale highlighting does not remain visible.
func clearFrameHighlighting(frame *FrameObject) {
	if frame == nil {
		return
	}
	for line := frame.FirstGroup.FirstLine; line != nil; line = line.FLink {
		line.HlMatch = nil
		line.HlState = nil
	}
	frame.Highlighter = nil
	frame.DirtyLine = 0
}

// LudwigBuffer implements highlight.LineStates over a frame's line list
// using lazy cursor-based navigation to avoid materialising the full line slice.
type LudwigBuffer struct {
	frame    *FrameObject
	numLines int
	curIdx   int
	curLine  *LineHdrObject
}

// newLudwigBuffer creates a LudwigBuffer for frame. It counts lines in
// O(groups) time rather than walking every line.
func newLudwigBuffer(frame *FrameObject) *LudwigBuffer {
	total := 0
	for g := frame.FirstGroup; g != nil; g = g.FLink {
		total += g.NrLines
	}
	if total > 0 {
		total-- // non-empty frame: subtract 1 to exclude the EOP sentinel (last line with FLink==nil)
	}
	return &LudwigBuffer{frame: frame, numLines: total, curIdx: -1}
}

// lineAt returns the line at 0-based index n. Sequential forward/backward
// access is O(1) via the cursor; random access uses group metadata (O(groups)).
func (b *LudwigBuffer) lineAt(n int) *LineHdrObject {
	if n < 0 || n >= b.numLines {
		return nil
	}
	if b.curIdx == n && b.curLine != nil {
		return b.curLine
	}
	// Common cases: one step forward or backward from the cursor.
	if b.curIdx >= 0 && b.curLine != nil {
		switch n {
		case b.curIdx + 1:
			if line := b.curLine.FLink; line != nil {
				b.curIdx, b.curLine = n, line
				return line
			}
		case b.curIdx - 1:
			if line := b.curLine.BLink; line != nil {
				b.curIdx, b.curLine = n, line
				return line
			}
		}
	}
	// General case: locate the group that owns 1-based line n+1, then walk
	// within it. Groups hold at most MaxGroupLines lines each, so the
	// inner walk is bounded and fast in practice.
	oneBasedN := n + 1
	for g := b.frame.FirstGroup; g != nil; g = g.FLink {
		if oneBasedN >= g.FirstLineNr && oneBasedN < g.FirstLineNr+g.NrLines {
			offset := oneBasedN - g.FirstLineNr
			line := g.FirstLine
			for range offset {
				if line = line.FLink; line == nil {
					return nil
				}
			}
			b.curIdx, b.curLine = n, line
			return line
		}
	}
	return nil
}

func (b *LudwigBuffer) LineBytes(n int) []byte {
	line := b.lineAt(n)
	if line == nil || line.Str == nil || line.Used == 0 {
		return nil
	}
	return []byte(line.Str.Slice(1, line.Used))
}

func (b *LudwigBuffer) LinesNum() int { return b.numLines }

func (b *LudwigBuffer) State(n int) highlight.State {
	if line := b.lineAt(n); line != nil {
		return line.HlState
	}
	return nil
}

func (b *LudwigBuffer) SetState(n int, s highlight.State) {
	if line := b.lineAt(n); line != nil {
		line.HlState = s
	}
}

func (b *LudwigBuffer) SetMatch(n int, m highlight.MatchEntries) {
	if line := b.lineAt(n); line != nil {
		line.HlMatch = m
	}
}

func (b *LudwigBuffer) Lock()   {} // Ludwig is single-threaded
func (b *LudwigBuffer) Unlock() {}

// SyntaxHighlightFrame runs a full highlight pass on a frame
func SyntaxHighlightFrame(frame *FrameObject, h *highlight.Highlighter) {
	buf := newLudwigBuffer(frame)
	h.HighlightStates(buf)
	if buf.LinesNum() > 0 {
		h.HighlightMatches(buf, 0, buf.LinesNum()-1)
	}
}

// SyntaxRehighlightFrom re-highlights incrementally from a changed line
func SyntaxRehighlightFrom(frame *FrameObject, h *highlight.Highlighter, fromLine int) {
	buf := newLudwigBuffer(frame)
	end := h.ReHighlightStates(buf, fromLine)
	h.HighlightMatches(buf, fromLine, end)
}

// SyntaxMarkLineDirty records that a line has been modified and needs re-highlighting.
// The re-highlighting is deferred and applied by SyntaxApplyDirty before the next screen redraw.
func SyntaxMarkLineDirty(frame *FrameObject, line *LineHdrObject) {
	if frame.Highlighter == nil {
		return
	}
	idx := max(line.Group.FirstLineNr+line.OffsetNr, 1)
	if frame.DirtyLine == 0 || idx < frame.DirtyLine {
		frame.DirtyLine = idx
	}
}

// SyntaxApplyDirty applies any pending incremental re-highlighting for the frame
// and redraws all currently visible lines so the updated colors take effect.
// screenExpand only draws lines not yet on screen, so we must do this ourselves.
func SyntaxApplyDirty(frame *FrameObject) {
	h := frame.Highlighter
	if h == nil {
		return
	}
	if frame.DirtyLine > 0 {
		idx := frame.DirtyLine - 1
		frame.DirtyLine = 0
		SyntaxRehighlightFrom(frame, h, idx)
		// Redraw visible lines so updated HlMatch values are shown.
		if ScrFrame == frame && ScrTopLine != nil {
			scrTopLineNum := ScrTopLine.Group.FirstLineNr + ScrTopLine.OffsetNr
			line := ScrTopLine
			for idx+1 > scrTopLineNum && line.FLink != nil {
				line = line.FLink
				scrTopLineNum += 1
			}
			for line != nil {
				ScreenDrawLine(line)
				if line == ScrBotLine {
					break
				}
				line = line.FLink
			}
		}
	}
}

// syntaxDrawLine renders line content with syntax highlighting.
// offset and strlen are already computed by ScreenDrawLine.
func syntaxDrawLine(line *LineHdrObject, offset, strlen int) {
	match := line.HlMatch

	currentPair := defaultPair
	if currentPair != 0 {
		nc.ColorOn(currentPair)
	}
	defer func() {
		if currentPair != 0 {
			nc.ColorOff(currentPair)
		}
	}()
	if len(match) == 0 {
		VduDisplayStr(line.Str.Slice(offset+1, strlen), true)
		return
	}

	// Walk the line segment by segment, switching colors
	col := offset // current rune column (0-based)
	end := offset + strlen

	for _, entry := range match {
		if entry.Position >= end {
			break
		}
		newPair, found := syntaxColorLookup[entry.Group]
		if !found || newPair == 0 {
			newPair = defaultPair
		}

		// Render any gap before this position
		if entry.Position > col {
			segLen := entry.Position - col
			if col+segLen > end {
				segLen = end - col
			}
			VduDisplayStr(line.Str.Slice(col+1, segLen), false)
			col += segLen
		}

		// Switch color
		if currentPair != newPair {
			if currentPair != 0 {
				nc.ColorOff(currentPair)
			}
			currentPair = newPair
			if currentPair != 0 {
				nc.ColorOn(currentPair)
			}
		}
	}

	// Render the remaining tail
	if col < end {
		VduDisplayStr(line.Str.Slice(col+1, end-col), false)
	}
}
