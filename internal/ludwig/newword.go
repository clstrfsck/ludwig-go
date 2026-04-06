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
	for (dot.Col > 1) && ChIsWordElement(0, rune(dot.Line.Str.Get(dot.Col))) {
		dot.Col--
	}
	if ChIsWordElement(0, rune(dot.Line.Str.Get(dot.Col))) {
		// we must have been somewhere on the line before the first word
		if dot.Line.BLink == nil { // oops top of the frame reached
			return false
		}
		if dot.Line.BLink.Used == 0 { // inside a paragraph break
			return false
		}
		MarkCreate(dot.Line.BLink, dot.Line.BLink.Used, &dot)
	}
	// ASSERT: we now have dot sitting on part of a word
	element := 0
	for !ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		element++
	}
	// Now find the start of this word
	for (dot.Col > 1) && ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		dot.Col--
	}
	if !ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
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
	for !ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		element++
	}
	for (dot.Col < dot.Line.Used) && ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		dot.Col++
	}
	if ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		if dot.Line.FLink == nil { // no more lines
			return false
		}
		if dot.Line.FLink.Used == 0 { // end of paragraph
			return false
		}
		MarkCreate(dot.Line.FLink, 1, &dot)
	}
	for ChIsWordElement(0, rune(dot.Line.Str.Get(dot.Col))) {
		dot.Col++
	}
	return true
}

// previousWord positions the mark at the start of the previous word
func previousWord(dot *MarkObject) bool {
	element := 0
	for !ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		element++
	}
	for (dot.Col > 1) && ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		dot.Col--
	}
	if ChIsWordElement(element, rune(dot.Line.Str.Get(dot.Col))) {
		if dot.Line.BLink == nil { // no more lines
			return false
		}
		if dot.Line.BLink.Used == 0 { // top of paragraph
			return false
		}
		MarkCreate(dot.Line.BLink, dot.Line.BLink.Used, &dot)
	}
	if !currentWord(dot) {
		return false
	}
	return true
}

// NewwordAdvanceWord advances cursor by word count
func NewwordAdvanceWord(frame *FrameObject, rept LeadParam, count int) bool {
	var newDot *MarkObject
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &newDot)
	defer MarkDestroy(&newDot)

	if rept == LeadParamMarker {
		MarkCreate(frame.Marks[count].Line, frame.Marks[count].Col, &newDot)
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
				return false
			}
		}
		MarkCreate(newDot.Line, newDot.Col, &frame.Dot)

	case LeadParamMinus, LeadParamNInt:
		count = -count
		if !currentWord(newDot) {
			return false
		}
		for count > 0 {
			count--
			if !previousWord(newDot) {
				return false
			}
		}
		MarkCreate(newDot.Line, newDot.Col, &frame.Dot)

	case LeadParamPIndef:
		if newDot.Line.Used == 0 { // Fail if we are on a blank line
			return false
		}
		if newDot.Col > newDot.Line.Used+2 {
			// check that we aren't past the last word in the para
			if newDot.Line.FLink == nil { // no more lines => end of para
				return false
			}
			if newDot.Line.FLink.Used == 0 { // next line blank => end of para
				return false
			}
			// In the middle of a paragraph so go it end of line
			newDot.Col = newDot.Line.Used
		}
		for nextWord(newDot) {
			MarkCreate(newDot.Line, newDot.Col, &frame.Dot)
		}
		// now on last word of paragraph
		//*** next statement should be more sophisticated
		//    what about the right margin??
		if newDot.Line.Used+2 > MaxStrLenP {
			MarkCreate(newDot.Line, MaxStrLenP, &frame.Dot)
		} else {
			MarkCreate(newDot.Line, newDot.Line.Used+2, &frame.Dot)
		}

	case LeadParamNIndef:
		if !currentWord(newDot) {
			return false
		}
		MarkCreate(newDot.Line, newDot.Col, &frame.Dot)
		for previousWord(newDot) {
			MarkCreate(newDot.Line, newDot.Col, &frame.Dot)
		}

	default:
		// marker Handled above
		return false
	}
	return true
}

// NewwordDeleteWord deletes words (same words as advance word advances over)
// Note that frameOops must not be nil.
func NewwordDeleteWord(frame *FrameObject, frameOops *FrameObject, rept LeadParam, count int) bool {
	result := false
	var oldPos *MarkObject
	var here *MarkObject
	var theOtherMark *MarkObject
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &oldPos)

	defer func() {
		MarkDestroy(&oldPos)
		MarkDestroy(&here)
		MarkDestroy(&theOtherMark)
	}()

	// First Step: Get to the beginning of the word if we are in the middle of it
	if !NewwordAdvanceWord(frame, LeadParamPInt, 0) {
		return false
	}
	// ASSERTION: We are on the beginning of a word
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &here)
	if !NewwordAdvanceWord(frame, rept, count) {
		// Put Dot back and bail out
		MarkCreate(oldPos.Line, oldPos.Col, &frame.Dot)
		return false
	}
	// OK. We now wipe out everything from Dot to here
	oldDotCol := frame.Dot.Col
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &theOtherMark)
	lineNr := LineToNumber(theOtherMark.Line)
	newLineNr := LineToNumber(here.Line)
	if (lineNr > newLineNr) || ((lineNr == newLineNr) && (theOtherMark.Col > here.Col)) {
		// Reverse mark pointers to get The_Other_Mark first.
		anotherMark := here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	if frame != frameOops {
		// Make sure oops_span is okay.
		MarkCreate(frameOops.LastGroup.LastLine, 1, &frameOops.Span.MarkTwo)
		result = TextMove(
			false,                        // Don't copy, transfer
			1,                            // One instance of
			theOtherMark,                 // starting pos.
			here,                         // ending pos.
			frameOops.Span.MarkTwo,       // destination.
			&frameOops.Marks[MarkEquals], // leave at start.
			&frameOops.Dot,               // leave at end.
		)
	} else {
		result = TextRemove(
			theOtherMark, // starting pos.
			here,         // ending pos.
		)
	}
	if lineNr != newLineNr {
		result = TextSplitLine(frame.Dot, oldDotCol, &here)
	}
	return result
}

// currentParagraph positions the mark at the start of the current paragraph
func currentParagraph(dot *MarkObject) bool {
	newLine := dot.Line
	var pos int
	if dot.Col < dot.Line.Used {
		pos = dot.Col
		for (pos > 1) && ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
			pos--
		}
		if ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
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
	for ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
		pos++
	}
	MarkCreate(newLine, pos, &dot)
	return true
}

// nextParagraph positions the mark at the start of the next paragraph
func nextParagraph(dot *MarkObject) bool {
	newLine := dot.Line
	var pos int
	if dot.Col < dot.Line.Used {
		pos = dot.Col
		for (pos > 1) && ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
			pos--
		}
		if ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
			if newLine.BLink == nil {
				dot.Col = 1
				for ChIsWordElement(0, rune(newLine.Str.Get(dot.Col))) {
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
	for ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
		pos++
	}
	MarkCreate(newLine, pos, &dot)
	return true
}

// NewwordAdvanceParagraph advances cursor by paragraph count
func NewwordAdvanceParagraph(frame *FrameObject, rept LeadParam, count int) bool {
	var newDot *MarkObject
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &newDot)
	defer MarkDestroy(&newDot)

	if rept == LeadParamMarker {
		MarkCreate(frame.Marks[count].Line, frame.Marks[count].Col, &newDot)
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
				return false
			}
		}
		MarkCreate(newDot.Line, newDot.Col, &frame.Dot)

	case LeadParamMinus, LeadParamNInt:
		count = -count
		if !currentParagraph(newDot) {
			return false
		}
		for count > 0 {
			count--
			if newDot.Line.BLink == nil {
				return false
			}
			MarkCreate(newDot.Line.BLink, 1, &newDot)
			if !currentParagraph(newDot) {
				return false
			}
		}
		MarkCreate(newDot.Line, newDot.Col, &frame.Dot)

	case LeadParamPIndef:
		MarkCreate(frame.LastGroup.LastLine, frame.MarginLeft, &frame.Dot)

	case LeadParamNIndef:
		newLine := newDot.Line
		for (newLine.BLink != nil) && (newLine.Used == 0) {
			newLine = newLine.BLink
		}
		if newLine.Used == 0 {
			return false
		}
		// OK we know that there is a paragraph behind us, so goto
		// the top of the file and go down to the first paragraph
		newLine = frame.FirstGroup.FirstLine
		for newLine.Used == 0 {
			newLine = newLine.FLink
		}
		pos := 1
		for ChIsWordElement(0, rune(newLine.Str.Get(pos))) {
			pos++
		}
		MarkCreate(newLine, pos, &frame.Dot)

	default:
		// Others handled elsewhere (marker) or ignored.
		break
	}
	return true
}

// NewwordDeleteParagraph deletes paragraphs.
// Note that frameOops must not be nil.
func NewwordDeleteParagraph(frame *FrameObject, frameOops *FrameObject, rept LeadParam, count int) bool {
	var oldPos *MarkObject
	var here *MarkObject
	var theOtherMark *MarkObject
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &oldPos)

	defer func() {
		MarkDestroy(&oldPos)
		MarkDestroy(&here)
		MarkDestroy(&theOtherMark)
	}()

	// Get to the beginning of the paragraph
	if !NewwordAdvanceParagraph(frame, LeadParamPInt, 0) {
		return false
	}
	MarkCreate(frame.Dot.Line, 1, &here)
	if !NewwordAdvanceParagraph(frame, rept, count) {
		// Something wrong so put dot back and abort
		MarkCreate(oldPos.Line, oldPos.Col, &frame.Dot)
		return false
	}

	// Now delete all the lines between marks dot and here
	MarkCreate(frame.Dot.Line, 1, &theOtherMark)
	lineNr := LineToNumber(theOtherMark.Line)
	newLineNr := LineToNumber(here.Line)
	if lineNr > newLineNr {
		// reverse marks to get the_other_mark first.
		anotherMark := here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	if frame != frameOops {
		// Make sure oops_span is okay.
		MarkCreate(frameOops.LastGroup.LastLine, 1, &frameOops.Span.MarkTwo)
		return TextMove(
			false,                        // Don't copy, transfer
			1,                            // One instance of
			theOtherMark,                 // starting pos.
			here,                         // ending pos.
			frameOops.Span.MarkTwo,       // destination.
			&frameOops.Marks[MarkEquals], // leave at start.
			&frameOops.Dot,               // leave at end.
		)
	} else {
		return TextRemove(
			theOtherMark, // starting pos.
			here,         // ending pos.
		)
	}
}
