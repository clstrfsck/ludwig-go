/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         NEWWORD
//
// Description:  Word Processing Commands for the new Ludwig command set.

package ludwig

// currentWord positions the mark at the start of the current word
func currentWord(dot *MarkObject) bool {
	if dot.Line.Used+2 < dot.Col {
		// check that we aren't past the last word in the para
		if dot.Line.FLink == nil { // no more lines => end of para
			return false
		}
		if dot.Line.FLink.Used == 0 { // next line blank => end of para
			return false
		}
		// In the middle of a paragraph so go to end of line
		dot.Col = dot.Line.Used
	} else if dot.Line.Used < dot.Col {
		dot.Col = dot.Line.Used
	}
	// But were we in the blank line before a paragraph?
	if dot.Col == 0 {
		return false
	}
	for (dot.Col > 1) && WordElements[0].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		dot.Col--
	}
	if WordElements[0].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		// we must have been somewhere on the line before the first word
		if dot.Line.BLink == nil { // oops top of the frame reached
			return false
		}
		if dot.Line.BLink.Used == 0 { // inside a paragraph break
			return false
		}
		if !MarkCreate(dot.Line.BLink, dot.Line.BLink.Used, &dot) {
			return false
		}
	}
	// ASSERT: we now have dot sitting on part of a word
	element := 0
	for WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) == 0 {
		element++
	}
	// Now find the start of this word
	for (dot.Col > 1) && WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		dot.Col--
	}
	if WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) == 0 {
		dot.Col++
	}
	return true
}

// nextWord positions the mark at the start of the next word
func nextWord(dot *MarkObject) bool {
	if dot.Col > dot.Line.Used {
		// check that we aren't on a blank line
		if dot.Line.Used == 0 {
			return false
		}
		// All clear so fake it that we were at the end of the last word!
		dot.Col = dot.Line.Used
	}
	element := 0
	for WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) == 0 {
		element++
	}
	for (dot.Col < dot.Line.Used) && WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		dot.Col++
	}
	if WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		if dot.Line.FLink == nil { // no more lines
			return false
		}
		if dot.Line.FLink.Used == 0 { // end of paragraph
			return false
		}
		if !MarkCreate(dot.Line.FLink, 1, &dot) {
			return false
		}
	}
	for WordElements[0].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		dot.Col++
	}
	return true
}

// previousWord positions the mark at the start of the previous word
func previousWord(dot *MarkObject) bool {
	element := 0
	for WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) == 0 {
		element++
	}
	for (dot.Col > 1) && WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		dot.Col--
	}
	if WordElements[element].Bit(int(dot.Line.Str.Get(dot.Col))) != 0 {
		if dot.Line.BLink == nil { // no more lines
			return false
		}
		if dot.Line.BLink.Used == 0 { // top of paragraph
			return false
		}
		if !MarkCreate(dot.Line.BLink, dot.Line.BLink.Used, &dot) {
			return false
		}
	}
	if !currentWord(dot) {
		return false
	}
	return true
}

// NewwordAdvanceWord advances cursor by word count
func NewwordAdvanceWord(rept LeadParam, count int) bool {
	result := false
	var newDot *MarkObject
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &newDot) {
		return false
	}
	if rept == LeadParamMarker {
		if !MarkCreate(
			CurrentFrame.Marks[count-MinMarkNumber].Line,
			CurrentFrame.Marks[count-MinMarkNumber].Col,
			&newDot,
		) {
			goto l98
		}
		rept = LeadParamNInt
		count = 0
	}
	// If we are doing a 0AW we need to go to the current word, -nAW does this
	if (rept == LeadParamPInt) && (count == 0) {
		rept = LeadParamNInt
	}
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		for count > 0 {
			count--
			if !nextWord(newDot) {
				goto l98
			}
		}
		if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
			goto l98
		}

	case LeadParamMinus, LeadParamNInt:
		count = -count
		if !currentWord(newDot) {
			goto l98
		}
		for count > 0 {
			count--
			if !previousWord(newDot) {
				goto l98
			}
		}
		if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
			goto l98
		}

	case LeadParamPIndef:
		if newDot.Line.Used == 0 { // Fail if we are on a blank line
			goto l98
		}
		if newDot.Col > newDot.Line.Used+2 {
			// check that we aren't past the last word in the para
			if newDot.Line.FLink == nil { // no more lines => end of para
				goto l98
			}
			if newDot.Line.FLink.Used == 0 { // next line blank => end of para
				goto l98
			}
			// In the middle of a paragraph so go it end of line
			newDot.Col = newDot.Line.Used
		}
		for nextWord(newDot) {
			if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
				goto l98
			}
		}
		// now on last word of paragraph
		//*** next statement should be more sophisticated
		//    what about the right margin??
		if newDot.Line.Used+2 > MaxStrLenP {
			if !MarkCreate(newDot.Line, MaxStrLenP, &CurrentFrame.Dot) {
				goto l98
			}
		} else if !MarkCreate(newDot.Line, newDot.Line.Used+2, &CurrentFrame.Dot) {
			goto l98
		}

	case LeadParamNIndef:
		if !currentWord(newDot) {
			goto l98
		}
		if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
			goto l98
		}
		for previousWord(newDot) {
			if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
				goto l98
			}
		}

	default:
		// marker Handled above
		goto l98
	}
	result = true
l98:
	MarkDestroy(&newDot)
	return result
}

// NewwordDeleteWord deletes words (same words as advance word advances over)
func NewwordDeleteWord(rept LeadParam, count int) bool {
	result := false
	var oldPos *MarkObject
	var here *MarkObject
	var theOtherMark *MarkObject
	var lineNr int
	var newLineNr int
	var oldDotCol int

	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &oldPos) {
		goto l99
	}
	// First Step: Get to the beginning of the word if we are in the middle of it
	if !NewwordAdvanceWord(LeadParamPInt, 0) {
		goto l99
	}
	// ASSERTION: We are on the beginning of a word
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &here) {
		goto l99
	}
	if !NewwordAdvanceWord(rept, count) {
		// Put Dot back and bail out
		if !MarkCreate(oldPos.Line, oldPos.Col, &CurrentFrame.Dot) {
			goto l99
		}
	}
	// OK. We now wipe out everything from Dot to here
	oldDotCol = CurrentFrame.Dot.Col
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &theOtherMark) {
		goto l99
	}
	if !LineToNumber(theOtherMark.Line, &lineNr) {
		goto l99
	}
	if !LineToNumber(here.Line, &newLineNr) {
		goto l99
	}
	if (lineNr > newLineNr) || ((lineNr == newLineNr) && (theOtherMark.Col > here.Col)) {
		// Reverse mark pointers to get The_Other_Mark first.
		anotherMark := here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	if CurrentFrame != FrameOops {
		// Make sure oops_span is okay.
		if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo) {
			goto l99
		}
		result = TextMove(
			false,                  // Don't copy, transfer
			1,                      // One instance of
			theOtherMark,           // starting pos.
			here,                   // ending pos.
			FrameOops.Span.MarkTwo, // destination.
			&FrameOops.Marks[MarkEquals-MinMarkNumber], // leave at start.
			&FrameOops.Dot, // leave at end.
		)
	} else {
		result = TextRemove(
			theOtherMark, // starting pos.
			here,         // ending pos.
		)
	}
	if lineNr != newLineNr {
		result = TextSplitLine(CurrentFrame.Dot, oldDotCol, &here)
	}
l99:
	if oldPos != nil {
		MarkDestroy(&oldPos)
	}
	if here != nil {
		MarkDestroy(&here)
	}
	if theOtherMark != nil {
		MarkDestroy(&theOtherMark)
	}
	return result
}

// currentParagraph positions the mark at the start of the current paragraph
func currentParagraph(dot *MarkObject) bool {
	newLine := dot.Line
	var pos int
	if dot.Col < dot.Line.Used {
		pos = dot.Col
		for (pos > 1) && WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
			pos--
		}
		if WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
			if newLine.BLink == nil {
				return false
			}
			newLine = newLine.BLink
		}
	}
	for (newLine.BLink != nil) && (newLine.Used == 0) {
		newLine = newLine.BLink
	}
	if newLine.Used == 0 {
		return false
	}
	for (newLine.BLink != nil) && (newLine.Used != 0) {
		newLine = newLine.BLink
	}
	if newLine.Used == 0 {
		newLine = newLine.FLink // Oops too far!
	}
	pos = 1
	for WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
		pos++
	}
	return MarkCreate(newLine, pos, &dot)
}

// nextParagraph positions the mark at the start of the next paragraph
func nextParagraph(dot *MarkObject) bool {
	newLine := dot.Line
	var pos int
	if dot.Col < dot.Line.Used {
		pos = dot.Col
		for (pos > 1) && WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
			pos--
		}
		if WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
			if newLine.BLink == nil {
				dot.Col = 1
				for WordElements[0].Bit(int(newLine.Str.Get(dot.Col))) != 0 {
					dot.Col++
				}
				return true
			}
			newLine = newLine.BLink
		}
	}
	for (newLine.FLink != nil) && (newLine.Used != 0) {
		newLine = newLine.FLink
	}
	if newLine.Used != 0 {
		return false
	}
	for (newLine.FLink != nil) && (newLine.Used == 0) {
		newLine = newLine.FLink
	}
	if newLine.Used == 0 {
		return false
	}
	pos = 1
	for WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
		pos++
	}
	return MarkCreate(newLine, pos, &dot)
}

// NewwordAdvanceParagraph advances cursor by paragraph count
func NewwordAdvanceParagraph(rept LeadParam, count int) bool {
	result := false
	var newDot *MarkObject

	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &newDot) {
		return false
	}
	if rept == LeadParamMarker {
		if !MarkCreate(
			CurrentFrame.Marks[count-MinMarkNumber].Line,
			CurrentFrame.Marks[count-MinMarkNumber].Col,
			&newDot,
		) {
			goto l98
		}
		rept = LeadParamNInt
		count = 0
	}
	// If we are doing a 0AP we need to go to the current para, -nAP does this
	if (rept == LeadParamPInt) && (count == 0) {
		rept = LeadParamNInt
	}
	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		for count > 0 {
			count--
			if !nextParagraph(newDot) {
				goto l98
			}
		}
		if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
			goto l98
		}

	case LeadParamMinus, LeadParamNInt:
		count = -count
		if !currentParagraph(newDot) {
			goto l98
		}
		for count > 0 {
			count--
			if newDot.Line.BLink == nil {
				goto l98
			}
			if !MarkCreate(newDot.Line.BLink, 1, &newDot) {
				goto l98
			}
			if !currentParagraph(newDot) {
				goto l98
			}
		}
		if !MarkCreate(newDot.Line, newDot.Col, &CurrentFrame.Dot) {
			goto l98
		}

	case LeadParamPIndef:
		if !MarkCreate(
			CurrentFrame.LastGroup.LastLine,
			CurrentFrame.MarginLeft,
			&CurrentFrame.Dot,
		) {
			goto l98
		}

	case LeadParamNIndef:
		newLine := newDot.Line
		for (newLine.BLink != nil) && (newLine.Used == 0) {
			newLine = newLine.BLink
		}
		if newLine.Used == 0 {
			goto l98
		}
		// OK we know that there is a paragraph behind us, so goto
		// the top of the file and go down to the first paragraph
		newLine = CurrentFrame.FirstGroup.FirstLine
		for newLine.Used == 0 {
			newLine = newLine.FLink
		}
		pos := 1
		for WordElements[0].Bit(int(newLine.Str.Get(pos))) != 0 {
			pos++
		}
		if !MarkCreate(newLine, pos, &CurrentFrame.Dot) {
			goto l98
		}

	default:
		// Others handled elsewhere (marker) or ignored.
		break
	}
	result = true
l98:
	MarkDestroy(&newDot)
	return result
}

// NewwordDeleteParagraph deletes paragraphs
func NewwordDeleteParagraph(rept LeadParam, count int) bool {
	result := false
	var oldPos *MarkObject
	var here *MarkObject
	var theOtherMark *MarkObject
	var lineNr int
	var newLineNr int

	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &oldPos) {
		goto l99
	}
	// Get to the beginning of the paragraph
	if !NewwordAdvanceParagraph(LeadParamPInt, 0) {
		goto l99
	}
	if !MarkCreate(CurrentFrame.Dot.Line, 1, &here) {
		goto l99
	}
	if !NewwordAdvanceParagraph(rept, count) {
		// Something wrong so put dot back and abort
		MarkCreate(oldPos.Line, oldPos.Col, &CurrentFrame.Dot)
		goto l99
	}

	// Now delete all the lines between marks dot and here
	if !MarkCreate(CurrentFrame.Dot.Line, 1, &theOtherMark) {
		goto l99
	}
	if !LineToNumber(theOtherMark.Line, &lineNr) {
		goto l99
	}
	if !LineToNumber(here.Line, &newLineNr) {
		goto l99
	}
	if lineNr > newLineNr {
		// reverse marks to get the_other_mark first.
		anotherMark := here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	if CurrentFrame != FrameOops {
		// Make sure oops_span is okay.
		if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo) {
			goto l99
		}
		result = TextMove(
			false,                  // Don't copy, transfer
			1,                      // One instance of
			theOtherMark,           // starting pos.
			here,                   // ending pos.
			FrameOops.Span.MarkTwo, // destination.
			&FrameOops.Marks[MarkEquals-MinMarkNumber], // leave at start.
			&FrameOops.Dot, // leave at end.
		)
	} else {
		result = TextRemove(
			theOtherMark, // starting pos.
			here,         // ending pos.
		)
	}

l99:
	if oldPos != nil {
		MarkDestroy(&oldPos)
	}
	if here != nil {
		MarkDestroy(&here)
	}
	if theOtherMark != nil {
		MarkDestroy(&theOtherMark)
	}
	return result
}
