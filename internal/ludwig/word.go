/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         WORD
//
// Description:  Word Processing Commands in Ludwig.

package ludwig

// WordFill fills/wraps text within margins
func WordFill(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	var there *MarkObject
	defer func() {
		MarkDestroy(&here)
		MarkDestroy(&there)
	}()

	leaveDotAlone := false
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if rept == LeadParamNone {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			return false
		}
		// Adjust the current line to the margins
		if frame.Dot.Line.BLink != nil {
			if frame.Dot.Line.BLink.Used != 0 {
				// Not the first line of a paragraph, adjust to left margin
				startChar := 1
				for (frame.Dot.Line.Str.Get(startChar) == ' ') &&
					(startChar < frame.Dot.Line.Used) {
					startChar++
				}
				if (startChar < frame.MarginLeft) &&
					(startChar < frame.Dot.Line.Used) {
					MarkCreate(frame.Dot.Line, startChar, &here)
					if !TextInsert(
						true, 1, BlankString, frame.MarginLeft-startChar, here,
					) {
						return false
					}
					MarkDestroy(&here)
				} else {
					// Might have to remove some spaces
					endChar := frame.MarginLeft
					if endChar < frame.Dot.Line.Used {
						for (frame.Dot.Line.Str.Get(endChar) == ' ') &&
							(endChar < frame.Dot.Line.Used) {
							endChar++
						}
						if endChar > 1 {
							MarkCreate(frame.Dot.Line, frame.MarginLeft, &here)
							MarkCreate(frame.Dot.Line, endChar, &there)
							if !TextRemove(here, there) {
								return false
							}
							MarkDestroy(&here)
							MarkDestroy(&there)
						}
					}
				}
			}
		}
		if frame.Dot.Line.Used > frame.MarginRight {
			// Must split this line if possible
			// 1. Scan back for first non-blank
			endChar := frame.MarginRight + 1
			if frame.Dot.Line.Str.Get(endChar) != ' ' {
				for (frame.Dot.Line.Str.Get(endChar) != ' ') &&
					(endChar > frame.MarginLeft) {
					endChar--
				}
				if endChar == frame.MarginLeft {
					return false
				}
			}
			startChar := endChar
			for (frame.Dot.Line.Str.Get(endChar) == ' ') &&
				(endChar > frame.MarginLeft) {
				endChar--
			}
			if endChar == frame.MarginLeft {
				return false
			}
			// 2. Scan forward for first non-blank
			for frame.Dot.Line.Str.Get(startChar) == ' ' {
				startChar++
			}
			// 3. Split the line and ensure new line starts at margin_left
			MarkCreate(frame.Dot.Line, startChar, &here)
			if frame.Dot.Col > endChar {
				MarkCreate(frame.Dot.Line, endChar, &frame.Dot)
			}
			if !TextSplitLine(here, frame.MarginLeft, &there) {
				return false
			}
			MarkDestroy(&here)
			MarkDestroy(&there)
			if rept != LeadParamPIndef {
				count++
			}
		} else {
			// Need to get stuff from the next line
			for {
				// 1. Figure out how many chars we can fit in
				spaceToAdd := frame.MarginRight - frame.Dot.Line.Used - 1
				// 2. See if we can find a word to fit
				if (spaceToAdd > 0) && (frame.Dot.Line.FLink.Used != 0) {
					startChar := 1
					for frame.Dot.Line.FLink.Str.Get(startChar) == ' ' {
						startChar++
					}
					endChar := startChar
					oldEnd := endChar
					for endChar <= frame.Dot.Line.FLink.Used {
						for frame.Dot.Line.FLink.Str.Get(endChar) == ' ' {
							endChar++
						}
						for (frame.Dot.Line.FLink.Str.Get(endChar) != ' ') &&
							(endChar < frame.Dot.Line.FLink.Used) {
							endChar++
						}
						if endChar == frame.Dot.Line.FLink.Used {
							endChar++
						}
						if spaceToAdd < (endChar - startChar) {
							endChar = frame.Dot.Line.FLink.Used + 1
						} else {
							oldEnd = endChar
						}
					}
					if ((oldEnd - startChar) <= spaceToAdd) && (oldEnd != startChar) {
						// It will fit
						var oldHere *MarkObject
						var oldThere *MarkObject
						MarkCreate(frame.Dot.Line.FLink, startChar, &here)
						MarkCreate(frame.Dot.Line.FLink, startChar, &oldHere)
						MarkCreate(frame.Dot.Line.FLink, oldEnd, &there)
						MarkCreate(frame.Dot.Line.FLink, oldEnd, &oldThere)
						frame.Dot.Col = frame.Dot.Line.Used + 2
						// Copy the text
						if !TextMove(true, 1, here, there, frame.Dot, &here, &there) {
							MarkDestroy(&oldHere)
							MarkDestroy(&oldThere)
							return false
						}
						// Copy the marks
						MarksShift(frame.Dot.Line.FLink, oldHere.Col, oldThere.Col-oldHere.Col, here.Line, here.Col)
						// Wipe out the old text
						MarkCreate(frame.Dot.Line.FLink, 1, &oldHere)
						MarkCreate(frame.Dot.Line.FLink, oldEnd, &oldThere)
						if !TextRemove(oldHere, oldThere) {
							MarkDestroy(&oldHere)
							MarkDestroy(&oldThere)
							return false
						}
						MarkDestroy(&oldHere)
						MarkDestroy(&oldThere)
						// If next line is now empty, delete it
						if frame.Dot.Line.FLink.Used == 0 {
							thisLine := frame.Dot.Line.FLink
							MarksSqueeze(frame.Dot.Line.FLink, 1, frame.Dot.Line.FLink.FLink, 1)
							LinesExtract(thisLine, thisLine)
							count--
							if count > 0 {
								continue
							}
							leaveDotAlone = true
						}
					}
					// Make sure first char in next line is at left margin
					if (count > 0) && (frame.Dot.Line.FLink.Used != 0) {
						startChar := 1
						for frame.Dot.Line.FLink.Str.Get(startChar) == ' ' {
							startChar++
						}
						MarkCreate(frame.Dot.Line.FLink, startChar, &there)
						if startChar < frame.MarginLeft {
							// Must insert some chars
							if !TextInsert(
								true,
								1,
								BlankString,
								frame.MarginLeft-startChar,
								there,
							) {
								return false
							}
						} else {
							MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &here)
							if !TextRemove(here, there) {
								return false
							}
						}
					}
				}
				MarkDestroy(&here)
				MarkDestroy(&there)
				break
			}
		}
		count--
		if !leaveDotAlone {
			MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		}
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordCentre centers text between margins
func WordCentre(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	var there *MarkObject

	defer func() {
		MarkDestroy(&here)
		MarkDestroy(&there)
	}()

	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			return false
		}
		if (frame.Dot.Line.Used < frame.MarginLeft) ||
			(frame.Dot.Line.Used > frame.MarginRight) {
			return false
		}
		startChar := 1
		for frame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar < frame.MarginLeft {
			return false
		}
		spaceToAdd := (frame.MarginRight-frame.MarginLeft-
			(frame.Dot.Line.Used-startChar))/2 -
			(startChar - frame.MarginLeft)
		if spaceToAdd > 0 {
			MarkCreate(frame.Dot.Line, frame.MarginLeft, &here)
			if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
				return false
			}
			MarkDestroy(&here)
		} else if spaceToAdd < 0 {
			MarkCreate(frame.Dot.Line, frame.MarginLeft, &here)
			MarkCreate(frame.Dot.Line, frame.MarginLeft-spaceToAdd, &there)
			if !TextRemove(here, there) {
				return false
			}
			MarkDestroy(&here)
			MarkDestroy(&there)
		}
		count--
		MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordJustify space-justifies text between margins
func WordJustify(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	defer MarkDestroy(&here)

	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			return false
		}
		if frame.Dot.Line.FLink.Used != 0 {
			if frame.Dot.Line.Used > frame.MarginRight {
				return false
			}

			// Figure out how many spaces to add
			spaceToAdd := frame.MarginRight - frame.Dot.Line.Used
			// Find number of holes for space distribution
			startChar := frame.MarginLeft
			for (frame.Dot.Line.Str.Get(startChar) == ' ') &&
				(startChar < frame.Dot.Line.Used) {
				startChar++
			}
			endChar := startChar
			holes := 0
			for {
				for (frame.Dot.Line.Str.Get(startChar) != ' ') &&
					(startChar < frame.Dot.Line.Used) {
					startChar++
				}
				for (frame.Dot.Line.Str.Get(startChar) == ' ') &&
					(startChar < frame.Dot.Line.Used) {
					startChar++
				}
				holes++
				if !(startChar < frame.Dot.Line.Used) {
					break
				}
			}
			holes--
			fillRatio := 0.0
			if holes > 0 {
				fillRatio = float64(spaceToAdd) / float64(holes)
			}
			debit := 0.0
			startChar = endChar
			for i := 1; i <= holes; i++ {
				// Find a hole
				for frame.Dot.Line.Str.Get(startChar) != ' ' {
					startChar++
				}
				debit += fillRatio
				spaceToAdd = int(debit + 0.5)
				if spaceToAdd > 0 {
					here = nil
					MarkCreate(frame.Dot.Line, startChar, &here)
					if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
						return false
					}
					MarkDestroy(&here)
					debit -= float64(spaceToAdd)
				}
				for frame.Dot.Line.Str.Get(startChar) == ' ' {
					startChar++
				}
			}
		}
		count--
		MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordSqueeze removes multiple spaces from lines
func WordSqueeze(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	var there *MarkObject

	defer func() {
		MarkDestroy(&here)
		MarkDestroy(&there)
	}()

	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			// on EOP line so abort
			return false
		}
		startChar := 1
		for frame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar += 1
		}
		for {
			for frame.Dot.Line.Str.Get(startChar) != ' ' &&
				startChar < frame.Dot.Line.Used {
				startChar++
			}
			if frame.Dot.Line.Str.Get(startChar) != ' ' {
				break // Nothing more to do
			}
			endChar := startChar
			for frame.Dot.Line.Str.Get(endChar) == ' ' {
				endChar++
			}
			if (endChar - startChar) > 1 {
				here = nil
				MarkCreate(frame.Dot.Line, startChar, &here)
				there = nil
				MarkCreate(frame.Dot.Line, endChar-1, &there)
				if !TextRemove(here, there) {
					return false
				}
				startChar = here.Col
			} else {
				startChar = endChar
			}
		}

		count -= 1
		MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordRight right-aligns text
func WordRight(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	var there *MarkObject
	defer func() {
		MarkDestroy(&here)
		MarkDestroy(&there)
	}()

	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			return false
		}
		if (frame.Dot.Line.Used < frame.MarginLeft) ||
			(frame.Dot.Line.Used > frame.MarginRight) {
			return false
		}
		startChar := 1
		for frame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar < frame.MarginLeft {
			return false
		}
		spaceToAdd := frame.MarginRight - frame.Dot.Line.Used
		if spaceToAdd > 0 {
			MarkCreate(frame.Dot.Line, startChar, &here)
			if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
				return false
			}
			MarkDestroy(&here)
		} else if spaceToAdd < 0 {
			MarkCreate(frame.Dot.Line, startChar, &there)
			MarkCreate(frame.Dot.Line, startChar-spaceToAdd, &here)
			if !TextRemove(there, here) {
				return false
			}
			MarkDestroy(&here)
			MarkDestroy(&there)
		}
		count--
		MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordLeft left-aligns text
func WordLeft(frame *FrameObject, rept LeadParam, count int) bool {
	var here *MarkObject
	var there *MarkObject
	defer func() {
		MarkDestroy(&here)
		MarkDestroy(&there)
	}()

	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount := count
		thisLine := frame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				return false
			}
		}
		if lineCount != 0 {
			return false
		}
	}
	for (count > 0) && (frame.Dot.Line.Used > 0) {
		if frame.Dot.Line.FLink == nil {
			return false
		}
		if (frame.Dot.Line.Used < frame.MarginLeft) ||
			(frame.Dot.Line.Used > frame.MarginRight) {
			return false
		}
		startChar := 1
		for frame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar != frame.MarginLeft {
			if startChar < frame.MarginLeft {
				MarkCreate(frame.Dot.Line, startChar, &here)
				if !TextInsert(
					true, 1, BlankString, frame.MarginLeft-startChar, here,
				) {
					return false
				}
				MarkDestroy(&here)
			} else {
				MarkCreate(frame.Dot.Line, frame.MarginLeft, &here)
				MarkCreate(frame.Dot.Line, startChar, &there)
				if !TextRemove(here, there) {
					return false
				}
				MarkDestroy(&here)
				MarkDestroy(&there)
			}
		}
		count--
		MarkCreate(frame.Dot.Line.FLink, frame.MarginLeft, &frame.Dot)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	}
	return (count <= 0) || (rept == LeadParamPIndef)
}

// WordAdvanceWord advances cursor to start of a word
func WordAdvanceWord(frame *FrameObject, rept LeadParam, count int) bool {
	thisLine := frame.Dot.Line
	pos := frame.Dot.Col

	if rept == LeadParamMarker {
		Screen.Message(MsgSyntaxError)
		return false
	}
	if rept == LeadParamNone || rept == LeadParamPlus || rept == LeadParamPIndef ||
		((rept == LeadParamPInt) && (count != 0)) {
		// Handle PINDEF case
		if rept == LeadParamPIndef {
			// Get to blank line between paragraphs
			for (thisLine.Used != 0) && (thisLine.FLink != nil) {
				thisLine = thisLine.FLink
			}
			pos = 1
			count = 1
		}
	outerLoop:
		for count > 0 {
			// Move forwards - locate next whitespace
			for {
				if pos < thisLine.Used {
					if thisLine.Str.Get(pos) != ' ' {
						pos++
					} else {
						break
					}
				} else {
					break
				}
			}
			// Skip whitespace until non-space
			if pos >= thisLine.Used {
				// Must move to next line
				pos = 1
				// Get next line with something on it
				for {
					if thisLine.FLink == nil {
						if rept == LeadParamPIndef {
							break outerLoop
						}
						return false
					}
					thisLine = thisLine.FLink
					if !(thisLine.Used <= 0) {
						break
					}
				}
			}
			for thisLine.Str.Get(pos) == ' ' {
				pos++
			}
			count--
		}
		MarkCreate(thisLine, pos, &frame.Dot)
	} else if rept == LeadParamNIndef {
		// Find non-blank line in paragraph
		for (thisLine.Used == 0) && (thisLine.BLink != nil) {
			thisLine = thisLine.BLink
		}
		// Find blank line separating this para from previous
		for (thisLine.Used != 0) && (thisLine.BLink != nil) {
			thisLine = thisLine.BLink
		}
		// Find first non-blank
		pos = 1
		for thisLine.Used == 0 {
			if thisLine.FLink == nil {
				return false
			}
			thisLine = thisLine.FLink
		}
		for thisLine.Str.Get(pos) == ' ' {
			pos++
		}
		MarkCreate(thisLine, pos, &frame.Dot)
	} else {
		// Move backwards
		count = -count
		if pos > thisLine.Used {
			pos = thisLine.Used
		}
		for {
			// If at start of line or on eop-line, go back
			if (pos == 0) || (thisLine.FLink == nil) {
				for {
					if thisLine.BLink == nil {
						return false
					}
					thisLine = thisLine.BLink
					pos = thisLine.Used
					if pos > 0 {
						break
					}
				}
			}
			// Skip whitespace
			for (thisLine.Str.Get(pos) == ' ') && (pos > 1) {
				pos--
			}
			if (pos == 1) && (thisLine.Str.Get(1) == ' ') {
				for {
					if thisLine.BLink == nil {
						return false
					}
					thisLine = thisLine.BLink
					pos = thisLine.Used
					if !(pos <= 0) {
						break
					}
				}
			}
			// Find start of word
			for (thisLine.Str.Get(pos) != ' ') && (pos > 1) {
				pos--
			}
			count--
			if count < 0 {
				if thisLine.Str.Get(pos) == ' ' {
					pos++
				}
			} else {
				pos--
			}
			if !(count >= 0) {
				break
			}
		}
		MarkCreate(thisLine, pos, &frame.Dot)
	}
	return true
}

// WordDeleteWord deletes words at cursor
func WordDeleteWord(frame *FrameObject, rept LeadParam, count int) bool {
	var oldPos *MarkObject
	var here *MarkObject
	var theOtherMark *MarkObject
	defer func() {
		MarkDestroy(&oldPos)
		MarkDestroy(&here)
		MarkDestroy(&theOtherMark)
	}()

	if rept == LeadParamMarker {
		Screen.Message(MsgSyntaxError)
		return false
	}
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &oldPos)
	// Get to beginning of word if in middle
	if !WordAdvanceWord(frame, LeadParamPInt, 0) {
		return false
	}
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &here)
	if !WordAdvanceWord(frame, rept, count) {
		// Put dot back and bail out
		MarkCreate(oldPos.Line, oldPos.Col, &frame.Dot)
		return false
	}
	// Wipe out everything from dot to here
	oldDotCol := frame.Dot.Col
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &theOtherMark)
	lineNr := LineToNumber(frame.Dot.Line)
	newLineNr := LineToNumber(here.Line)
	if (lineNr > newLineNr) ||
		((lineNr == newLineNr) && (frame.Dot.Col > here.Col)) {
		// Reverse mark pointers
		anotherMark := here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	result := false
	if frame != FrameOops {
		// Make sure oops_span is okay
		MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo)
		result = TextMove(
			false,
			1,
			theOtherMark,
			here,
			FrameOops.Span.MarkTwo,
			&FrameOops.Marks[MarkEquals],
			&FrameOops.Dot,
		)
	} else {
		result = TextRemove(theOtherMark, here)
	}
	if lineNr != newLineNr {
		result = TextSplitLine(frame.Dot, oldDotCol, &here)
	}
	return result
}
