/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         LINE
//
// Description:  Line manipulation commands.

package ludwig

// LineEOPCreate creates a group containing only the EOP line
func LineEOPCreate(inframe *FrameObject, group **GroupObject) bool {
	newLine := &LineHdrObject{}
	newGroup := &GroupObject{}

	newLine.FLink = nil
	newLine.BLink = nil
	newLine.Group = newGroup
	newLine.OffsetNr = 0
	newLine.Marks = nil
	newLine.Str = nil
	newLine.Len = 0
	newLine.Used = 0
	newLine.ScrRowNr = 0

	newGroup.FLink = nil
	newGroup.BLink = nil
	newGroup.Frame = inframe
	newGroup.FirstLine = newLine
	newGroup.LastLine = newLine
	newGroup.FirstLineNr = 1
	newGroup.NrLines = 1

	*group = newGroup
	return true
}

// LineEOPDestroy destroys a group containing only the EOP line
func LineEOPDestroy(group **GroupObject) bool {
	eopLine := (*group).FirstLine

	if eopLine.Str != nil {
		eopLine.Str = nil
		eopLine.Marks = nil
	}
	*group = nil
	return true
}

// LinesCreate creates a linked list of lines
func LinesCreate(lineCount int, firstLine **LineHdrObject, lastLine **LineHdrObject) bool {
	var topLine *LineHdrObject
	var prevLine *LineHdrObject
	var thisLine *LineHdrObject

	for lineNr := 1; lineNr <= lineCount; lineNr++ {
		thisLine = &LineHdrObject{}

		if topLine == nil {
			topLine = thisLine
		}

		thisLine.FLink = nil
		thisLine.BLink = prevLine
		thisLine.Group = nil
		thisLine.OffsetNr = 0
		thisLine.Marks = nil
		thisLine.Str = nil
		thisLine.Len = 0
		thisLine.Used = 0
		thisLine.ScrRowNr = 0

		if prevLine != nil {
			prevLine.FLink = thisLine
		}
		prevLine = thisLine
	}

	*firstLine = topLine
	*lastLine = thisLine
	return true
}

// LinesDestroy destroys a linked list of lines
func LinesDestroy(firstLine **LineHdrObject, lastLine **LineHdrObject) bool {
	thisLine := *firstLine

	for thisLine != nil {
		if thisLine.Str != nil {
			thisLine.Str = nil
			thisLine.Marks = nil
		}

		nextLine := thisLine.FLink
		thisLine = nextLine
	}

	*firstLine = nil
	*lastLine = nil
	return true
}

// GroupsDestroy destroys a linked list of groups
func GroupsDestroy(firstGroup **GroupObject, lastGroup **GroupObject) bool {
	*firstGroup = nil
	*lastGroup = nil
	return true
}

// LinesInject injects a linked list of lines into the data structure
func LinesInject(firstLine *LineHdrObject, lastLine *LineHdrObject, beforeLine *LineHdrObject) bool {
	// Scan the lines to be inserted, counting lines and checking space used
	var nrNewLines int
	var space int

	thisLine := firstLine
	for thisLine != nil {
		space += thisLine.Len
		nrNewLines++
		thisLine = thisLine.FLink
	}

	/* Define some useful pointers
	 *                       +-------------+   +-------------+
	 *                       | top_group   |   | top_line    |
	 *                       +-------------+   +-------------+
	 *                           |     ^           |     ^
	 * insert groups here ===>   |     |           |     |   <=== insert lines here
	 *                           v     |           v     |
	 *     +-------------+   +-------------+   +-------------+
	 *     | this_frame  |<--| end_group   |<--| before_line |
	 *     +-------------+   +-------------+   +-------------+
	 */
	topLine := beforeLine.BLink
	endGroup := beforeLine.Group
	topGroup := endGroup.BLink
	thisFrame := endGroup.Frame

	// Determine number of free lines available in end_group and top_group
	nrFreeLinesEnd := MaxGroupLines - endGroup.NrLines
	var nrFreeLinesTop int
	if topGroup != nil {
		nrFreeLinesTop = MaxGroupLines - topGroup.NrLines
	} else {
		nrFreeLinesTop = 0
	}
	nrFreeLines := nrFreeLinesEnd + nrFreeLinesTop
	lineNr := endGroup.FirstLineNr

	// If insufficient free lines are available, insert some new groups
	var adjustGroup *GroupObject
	if nrNewLines > nrFreeLines {
		// Create a chain of new groups
		nrNewGroups := (nrNewLines-nrFreeLines-1)/MaxGroupLines + 1
		var firstGroup *GroupObject
		var lastGroup *GroupObject

		for groupNr := 1; groupNr <= nrNewGroups; groupNr++ {
			thisGroup := &GroupObject{}

			if firstGroup == nil {
				firstGroup = thisGroup
			}

			thisGroup.FLink = nil
			thisGroup.BLink = lastGroup
			thisGroup.Frame = thisFrame
			thisGroup.FirstLine = nil
			thisGroup.LastLine = nil
			thisGroup.FirstLineNr = lineNr
			thisGroup.NrLines = 0

			if lastGroup != nil {
				lastGroup.FLink = thisGroup
			}
			lastGroup = thisGroup
		}

		// Link new groups into data structure between top_group and end_group
		lastGroup.FLink = endGroup
		endGroup.BLink = lastGroup
		if topGroup != nil {
			topGroup.FLink = firstGroup
			adjustGroup = topGroup
		} else {
			thisFrame.FirstGroup = firstGroup
			adjustGroup = firstGroup
		}
		firstGroup.BLink = topGroup
	} else if nrNewLines > nrFreeLinesEnd {
		adjustGroup = topGroup
	} else {
		adjustGroup = endGroup
	}

	// Insert lines into data structure between top_line and before_line
	lastLine.FLink = beforeLine
	beforeLine.BLink = lastLine
	if beforeLine.OffsetNr == 0 {
		endGroup.FirstLine = firstLine
	}
	if topLine != nil {
		topLine.FLink = firstLine
	}
	firstLine.BLink = topLine

	// Now put the data structure back together
	nrLinesToAdjust := nrNewLines
	var adjustLine *LineHdrObject

	if nrNewLines > nrFreeLinesEnd {
		adjustLine = endGroup.FirstLine
		nrLinesToAdjust = nrLinesToAdjust + beforeLine.OffsetNr
		endGroup.NrLines = 0
	} else {
		adjustLine = firstLine
		endGroup.NrLines = beforeLine.OffsetNr
	}
	endGroupLastLine := endGroup.LastLine

	for nrLinesToAdjust > 0 {
		nrLinesToAdjustHere := MaxGroupLines - adjustGroup.NrLines
		if nrLinesToAdjustHere > nrLinesToAdjust {
			nrLinesToAdjustHere = nrLinesToAdjust
		}

		if adjustGroup.NrLines == 0 {
			adjustGroup.FirstLine = adjustLine
			adjustGroup.FirstLineNr = lineNr
		}

		for offset := adjustGroup.NrLines; offset < adjustGroup.NrLines+nrLinesToAdjustHere; offset++ {
			adjustLine.Group = adjustGroup
			adjustLine.OffsetNr = offset
			adjustLine = adjustLine.FLink
		}

		adjustGroup.LastLine = adjustLine.BLink
		adjustGroup.NrLines += nrLinesToAdjustHere
		lineNr = adjustGroup.FirstLineNr + adjustGroup.NrLines
		nrLinesToAdjust -= nrLinesToAdjustHere
		adjustGroup = adjustGroup.FLink
	}

	nextGroupFirstLine := endGroupLastLine.FLink
	offset := endGroup.NrLines

	for {
		adjustLine.OffsetNr = offset
		offset++
		adjustLine = adjustLine.FLink
		if adjustLine == nextGroupFirstLine {
			break
		}
	}

	endGroup.LastLine = endGroupLastLine
	if adjustGroup == endGroup {
		endGroup.FirstLineNr = lineNr
		endGroup.FirstLine = beforeLine
	}
	endGroup.NrLines = offset

	adjustGroup = endGroup.FLink
	for adjustGroup != nil {
		adjustGroup.FirstLineNr = adjustGroup.FirstLineNr + nrNewLines
		adjustGroup = adjustGroup.FLink
	}

	thisFrame.SpaceLeft -= space

	// Update the screen
	if beforeLine.ScrRowNr != 0 && beforeLine != ScrTopLine {
		ScreenLinesInject(firstLine, nrNewLines, beforeLine)
	}

	return true
}

// LinesExtract extracts lines from the data structure
func LinesExtract(firstLine *LineHdrObject, lastLine *LineHdrObject) bool {
	// Define some useful pointers
	topLine := firstLine.BLink
	endLine := lastLine.FLink

	firstGroup := firstLine.Group
	lastGroup := lastLine.Group
	var topGroup *GroupObject
	if topLine != nil {
		topGroup = topLine.Group
	}
	endGroup := endLine.Group
	thisFrame := endGroup.Frame

	firstLineOffsetNr := firstLine.OffsetNr
	firstLineNr := firstGroup.FirstLineNr + firstLineOffsetNr
	nrLinesToRemove := lastGroup.FirstLineNr + lastLine.OffsetNr - firstLineNr + 1

	if thisFrame == ScrFrame {
		var firstScrLine *LineHdrObject
		var lastScrLine *LineHdrObject
		needsExtract := false

		if firstLine.ScrRowNr != 0 {
			firstScrLine = firstLine
		} else {
			if firstLineNr < ScrTopLine.Group.FirstLineNr+ScrTopLine.OffsetNr {
				firstScrLine = ScrTopLine
			} else {
				goto done1
			}
		}

		if lastLine.ScrRowNr != 0 {
			lastScrLine = lastLine
		} else {
			if lastLine.Group.FirstLineNr+lastLine.OffsetNr > ScrBotLine.Group.FirstLineNr+ScrBotLine.OffsetNr {
				lastScrLine = ScrBotLine
			} else {
				goto done1
			}
		}
		needsExtract = true
	done1:
		if needsExtract {
			ScreenLinesExtract(firstScrLine, lastScrLine)
		}
	}

	// Unlink the lines
	if topLine != nil {
		topLine.FLink = endLine
	}
	firstLine.BLink = nil
	lastLine.FLink = nil
	endLine.BLink = topLine

	// Determine the space being released by removing these lines
	var space int
	thisLine := firstLine
	for lineNr := 1; lineNr <= nrLinesToRemove; lineNr++ {
		space += thisLine.Len
		thisLine = thisLine.FLink
	}
	thisFrame.SpaceLeft += space

	// Adjust top_group and end_group
	if topGroup != endGroup {
		if topGroup != nil {
			topGroup.LastLine = topLine
		}
		endGroup.FirstLine = endLine
		endGroup.FirstLineNr = firstLineNr
	}

	// Adjust groups below end_group
	thisGroup := endGroup.FLink
	for thisGroup != nil {
		thisGroup.FirstLineNr -= nrLinesToRemove
		thisGroup = thisGroup.FLink
	}

	// Adjust first_group..last_group for removed lines
	if firstGroup == topGroup {
		nrLinesToRemove -= firstGroup.NrLines - firstLineOffsetNr
		firstGroup.NrLines = firstLineOffsetNr
		if firstGroup != lastGroup {
			firstGroup = firstGroup.FLink
		}
	}

	thisGroup = firstGroup
	for nrLinesToRemove > 0 {
		nrLinesToRemove = nrLinesToRemove - thisGroup.NrLines
		thisGroup.NrLines = 0
		thisGroup = thisGroup.FLink
	}

	// Adjust end_group for remaining lines
	if nrLinesToRemove < 0 {
		var offset int
		if topGroup == endGroup {
			offset = firstLineOffsetNr
			endGroup.NrLines = offset - nrLinesToRemove
		} else {
			offset = 0
			endGroup.NrLines = -nrLinesToRemove
		}

		thisLine = endLine
		for ; offset < endGroup.NrLines; offset++ {
			thisLine.OffsetNr = offset
			thisLine = thisLine.FLink
		}
	}

	// Dispose of empty groups
	if firstGroup.NrLines == 0 {
		lastGroup = firstGroup
		endGroup = lastGroup.FLink
		for endGroup.NrLines == 0 {
			lastGroup = endGroup
			endGroup = endGroup.FLink
		}

		topGroup = firstGroup.BLink
		if topGroup != nil {
			topGroup.FLink = endGroup
		} else {
			thisFrame.FirstGroup = endGroup
		}
		firstGroup.BLink = nil
		lastGroup.FLink = nil
		endGroup.BLink = topGroup

		if !GroupsDestroy(&firstGroup, &lastGroup) {
			return false
		}
	}

	return true
}

// LineChangeLength changes the length of the allocated text of a line
func LineChangeLength(line *LineHdrObject, newLength int) bool {
	var newStr *StrObject

	if newLength > 0 {
		// Quantize the length to get some slack
		if newLength < MaxStrLen-10 {
			newLength = (newLength/10 + 1) * 10
		} else {
			newLength = MaxStrLen
		}

		// Create a new str_object and copy the text from the old one
		newStr = &StrObject{}
		ChFillCopy(line.Str, 1, line.Len, newStr, 1, newLength, ' ')
	} else {
		newStr = nil
	}

	// Update the amount of free space available in the frame
	if line.Group != nil {
		line.Group.Frame.SpaceLeft += line.Len - newLength
	}

	// Change line to refer to the new str_object
	line.Str = newStr
	line.Len = newLength

	return true
}

// LineToNumber determines the line number of a given line
func LineToNumber(line *LineHdrObject, number *int) bool {
	*number = line.Group.FirstLineNr + line.OffsetNr
	return true
}

// LineFromNumber finds the line with a given line number in a given frame
func LineFromNumber(frame *FrameObject, number int, line **LineHdrObject) bool {
	thisGroup := frame.LastGroup

	if number >= thisGroup.FirstLineNr+thisGroup.NrLines {
		*line = nil
	} else {
		for thisGroup.FirstLineNr > number {
			thisGroup = thisGroup.BLink
		}

		thisLine := thisGroup.FirstLine
		for lineNr := 1; lineNr <= number-thisGroup.FirstLineNr; lineNr++ {
			thisLine = thisLine.FLink
		}
		*line = thisLine
	}

	return true
}
