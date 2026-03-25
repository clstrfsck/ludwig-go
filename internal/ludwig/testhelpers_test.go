package ludwig

// emptyTestLineInFrame creates a test line within a frame with a group
func emptyTestLineInFrame() (*FrameObject, *LineHdrObject) {
	return contentLineInFrame("")
}

// contentLineInFrame creates a frame containing a single content line plus the
// trailing NULL/sentinel line
func contentLineInFrame(content string) (*FrameObject, *LineHdrObject) {
	frame, lines := contentLinesInFrame([]string{content})
	return frame, lines[0]
}

// contentLinesInFrame creates a frame with a single group whose content
// consists of the provided lines, followed by a trailing NULL/sentinel
// line. The group's NrLines field counts only the real content lines
// (len(lines)); the extra sentinel line (with FLink == nil) is added at
// the end of the linked list to match the editor's internal layout.
func contentLinesInFrame(lines []string) (*FrameObject, []*LineHdrObject) {
	if len(lines) == 0 {
		panic("Must have at least one line")
	}
	frame := &FrameObject{
		SpaceLeft:   MaxSpace,
		SpaceLimit:  MaxSpace,
		ScrWidth:    80,
		MarginLeft:  1,
		MarginRight: MaxStrLen,
	}

	group := &GroupObject{
		Frame:       frame,
		FirstLineNr: 1,
		NrLines:     len(lines),
	}

	var prevLine *LineHdrObject
	var firstLine *LineHdrObject
	var textLines []*LineHdrObject = make([]*LineHdrObject, len(lines))

	for i, content := range lines {
		line := &LineHdrObject{
			Group:    group,
			OffsetNr: i,
			Used:     len(content),
			Str:      NewBlankStrObject(MaxStrLen),
			Marks:    make([]*MarkObject, 0), // Initialize marks slice
		}

		textLines[i] = line

		// Copy content
		line.Str.Assign(content)

		if prevLine != nil {
			prevLine.FLink = line
			line.BLink = prevLine
		} else {
			firstLine = line
		}
		prevLine = line
	}

	// Add NULL line at the end
	nullLine := &LineHdrObject{
		Group:    group,
		OffsetNr: len(lines),
		BLink:    prevLine,
		Marks:    make([]*MarkObject, 0),
	}
	if prevLine != nil {
		prevLine.FLink = nullLine
	}

	group.FirstLine = firstLine
	group.LastLine = nullLine
	frame.FirstGroup = group
	frame.LastGroup = group

	// Create dot mark using MarkCreate
	MarkCreate(firstLine, 1, &frame.Dot)

	return frame, textLines
}
