package ludwig

// emptyTestLineInFrame creates a test line within a frame with a group
func emptyTestLineInFrame() *FrameObject {
	return contentLinesInFrame([]string{""})
}

// contentLineInFrame creates a frame with a line for character command testing
func contentLineInFrame(content string) *FrameObject {
	return contentLinesInFrame([]string{content})
}

// contentLinesInFrame creates a frame with multiple lines
func contentLinesInFrame(lines []string) *FrameObject {
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

	for i, content := range lines {
		line := &LineHdrObject{
			Group:    group,
			OffsetNr: i,
			Used:     len(content),
			Str:      NewBlankStrObject(MaxStrLen),
			Marks:    make([]*MarkObject, 0), // Initialize marks slice
		}

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

	return frame
}
