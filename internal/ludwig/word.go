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
func WordFill(rept LeadParam, count int) bool {
	var startChar int
	var endChar int
	var oldEnd int
	var lineCount int
	var spaceToAdd int
	var thisLine *LineHdrObject
	var here *MarkObject
	var there *MarkObject
	var oldHere *MarkObject
	var oldThere *MarkObject
	var leaveDotAlone bool

	result := false
	leaveDotAlone = false
	here = nil
	there = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if rept == LeadParamNone {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			goto cleanup
		}
		// Adjust the current line to the margins
		if CurrentFrame.Dot.Line.BLink != nil {
			if CurrentFrame.Dot.Line.BLink.Used != 0 {
				// Not the first line of a paragraph, adjust to left margin
				startChar = 1
				for (CurrentFrame.Dot.Line.Str.Get(startChar) == ' ') &&
					(startChar < CurrentFrame.Dot.Line.Used) {
					startChar++
				}
				if (startChar < CurrentFrame.MarginLeft) &&
					(startChar < CurrentFrame.Dot.Line.Used) {
					if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
						goto cleanup
					}
					if !TextInsert(
						true, 1, BlankString, CurrentFrame.MarginLeft-startChar, here,
					) {
						goto cleanup
					}
					MarkDestroy(&here)
				} else {
					// Might have to remove some spaces
					startChar = CurrentFrame.MarginLeft
					endChar = CurrentFrame.MarginLeft
					if endChar < CurrentFrame.Dot.Line.Used {
						for (CurrentFrame.Dot.Line.Str.Get(endChar) == ' ') &&
							(endChar < CurrentFrame.Dot.Line.Used) {
							endChar++
						}
						if endChar > 1 {
							if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
								goto cleanup
							}
							if !MarkCreate(CurrentFrame.Dot.Line, endChar, &there) {
								goto cleanup
							}
							if !TextRemove(here, there) {
								goto cleanup
							}
							MarkDestroy(&here)
							MarkDestroy(&there)
						}
					}
				}
			}
		}
		if CurrentFrame.Dot.Line.Used > CurrentFrame.MarginRight {
			// Must split this line if possible
			// 1. Scan back for first non-blank
			endChar = CurrentFrame.MarginRight + 1
			if CurrentFrame.Dot.Line.Str.Get(endChar) != ' ' {
				for (CurrentFrame.Dot.Line.Str.Get(endChar) != ' ') &&
					(endChar > CurrentFrame.MarginLeft) {
					endChar--
				}
				if endChar == CurrentFrame.MarginLeft {
					goto cleanup
				}
			}
			startChar = endChar
			for (CurrentFrame.Dot.Line.Str.Get(endChar) == ' ') &&
				(endChar > CurrentFrame.MarginLeft) {
				endChar--
			}
			if endChar == CurrentFrame.MarginLeft {
				goto cleanup
			}
			// 2. Scan forward for first non-blank
			for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
				startChar++
			}
			// 3. Split the line and ensure new line starts at margin_left
			if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
				goto cleanup
			}
			if CurrentFrame.Dot.Col > endChar {
				if !MarkCreate(CurrentFrame.Dot.Line, endChar, &CurrentFrame.Dot) {
					goto cleanup
				}
			}
			if !TextSplitLine(here, CurrentFrame.MarginLeft, &there) {
				goto cleanup
			}
			if !MarkDestroy(&here) {
				goto cleanup
			}
			if !MarkDestroy(&there) {
				goto cleanup
			}
			if rept != LeadParamPIndef {
				count++
			}
		} else {
			// Need to get stuff from the next line
		getMore:
			// 1. Figure out how many chars we can fit in
			spaceToAdd = CurrentFrame.MarginRight - CurrentFrame.Dot.Line.Used - 1
			// 2. See if we can find a word to fit
			startChar = 1
			if (spaceToAdd > 0) && (CurrentFrame.Dot.Line.FLink.Used != 0) {
				for CurrentFrame.Dot.Line.FLink.Str.Get(startChar) == ' ' {
					startChar++
				}
				endChar = startChar
				oldEnd = endChar
				for endChar <= CurrentFrame.Dot.Line.FLink.Used {
					for CurrentFrame.Dot.Line.FLink.Str.Get(endChar) == ' ' {
						endChar++
					}
					for (CurrentFrame.Dot.Line.FLink.Str.Get(endChar) != ' ') &&
						(endChar < CurrentFrame.Dot.Line.FLink.Used) {
						endChar++
					}
					if endChar == CurrentFrame.Dot.Line.FLink.Used {
						endChar++
					}
					if spaceToAdd < (endChar - startChar) {
						endChar = CurrentFrame.Dot.Line.FLink.Used + 1
					} else {
						oldEnd = endChar
					}
				}
				if ((oldEnd - startChar) <= spaceToAdd) && (oldEnd != startChar) {
					// It will fit
					oldHere = nil
					oldThere = nil
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, startChar, &here) {
						goto cleanup
					}
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, startChar, &oldHere) {
						goto cleanup
					}
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, oldEnd, &there) {
						goto cleanup
					}
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, oldEnd, &oldThere) {
						goto cleanup
					}
					CurrentFrame.Dot.Col = CurrentFrame.Dot.Line.Used + 2
					// Copy the text
					if !TextMove(true, 1, here, there, CurrentFrame.Dot, &here, &there) {
						goto cleanup
					}
					// Copy the marks
					if !MarksShift(
						CurrentFrame.Dot.Line.FLink,
						oldHere.Col,
						oldThere.Col-oldHere.Col,
						here.Line,
						here.Col,
					) {
						goto cleanup
					}
					// Wipe out the old text
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, 1, &oldHere) {
						goto cleanup
					}
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, oldEnd, &oldThere) {
						goto cleanup
					}
					if !TextRemove(oldHere, oldThere) {
						goto cleanup
					}
					MarkDestroy(&oldHere)
					MarkDestroy(&oldThere)
					// If next line is now empty, delete it
					if CurrentFrame.Dot.Line.FLink.Used == 0 {
						thisLine = CurrentFrame.Dot.Line.FLink
						if !MarksSqueeze(
							CurrentFrame.Dot.Line.FLink,
							1,
							CurrentFrame.Dot.Line.FLink.FLink,
							1,
						) {
							goto cleanup
						}
						if !LinesExtract(thisLine, thisLine) {
							goto cleanup
						}
						if !LinesDestroy(&thisLine, &thisLine) {
							goto cleanup
						}
						count--
						if count > 0 {
							goto getMore
						}
						leaveDotAlone = true
					}
				}
				// Make sure first char in next line is at left margin
				if (count > 0) && (CurrentFrame.Dot.Line.FLink.Used != 0) {
					startChar = 1
					for CurrentFrame.Dot.Line.FLink.Str.Get(startChar) == ' ' {
						startChar++
					}
					if !MarkCreate(CurrentFrame.Dot.Line.FLink, startChar, &there) {
						goto cleanup
					}
					if startChar < CurrentFrame.MarginLeft {
						// Must insert some chars
						if !TextInsert(
							true,
							1,
							BlankString,
							CurrentFrame.MarginLeft-startChar,
							there,
						) {
							goto cleanup
						}
					} else {
						if !MarkCreate(
							CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &here,
						) {
							goto cleanup
						}
						if !TextRemove(here, there) {
							goto cleanup
						}
					}
				}
				if here != nil {
					if !MarkDestroy(&here) {
						goto cleanup
					}
				}
				if there != nil {
					if !MarkDestroy(&there) {
						goto cleanup
					}
				}
			}
		}
		count--
		if !leaveDotAlone {
			if !MarkCreate(
				CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot,
			) {
				goto cleanup
			}
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified],
		) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	if there != nil {
		MarkDestroy(&there)
	}
	return result
}

// WordCentre centers text between margins
func WordCentre(rept LeadParam, count int) bool {
	var startChar int
	var lineCount int
	var spaceToAdd int
	var thisLine *LineHdrObject
	var here *MarkObject
	var there *MarkObject

	result := false
	here = nil
	there = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			goto cleanup
		}
		if (CurrentFrame.Dot.Line.Used < CurrentFrame.MarginLeft) ||
			(CurrentFrame.Dot.Line.Used > CurrentFrame.MarginRight) {
			goto cleanup
		}
		startChar = 1
		for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar < CurrentFrame.MarginLeft {
			goto cleanup
		}
		spaceToAdd = (CurrentFrame.MarginRight-CurrentFrame.MarginLeft-
			(CurrentFrame.Dot.Line.Used-startChar))/2 -
			(startChar - CurrentFrame.MarginLeft)
		if spaceToAdd > 0 {
			if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
				goto cleanup
			}
			if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
				goto cleanup
			}
			if !MarkDestroy(&here) {
				goto cleanup
			}
		} else if spaceToAdd < 0 {
			if !MarkCreate(CurrentFrame.Dot.Line, startChar, &there) {
				goto cleanup
			}
			if !MarkCreate(CurrentFrame.Dot.Line, startChar-spaceToAdd, &here) {
				goto cleanup
			}
			if !TextRemove(there, here) {
				goto cleanup
			}
			if !MarkDestroy(&here) {
				goto cleanup
			}
			if !MarkDestroy(&there) {
				goto cleanup
			}
		}
		count--
		if !MarkCreate(
			CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot,
		) {
			goto cleanup
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified],
		) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	if there != nil {
		MarkDestroy(&there)
	}
	return result
}

// WordJustify space-justifies text between margins
func WordJustify(rept LeadParam, count int) bool {
	var startChar int
	var endChar int
	var holes int
	var i int
	var lineCount int
	var spaceToAdd int
	var thisLine *LineHdrObject
	var here *MarkObject
	var fillRatio float64
	var debit float64

	result := false
	here = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			goto cleanup
		}
		if CurrentFrame.Dot.Line.FLink.Used == 0 {
			goto nextLine
		}
		if CurrentFrame.Dot.Line.Used > CurrentFrame.MarginRight {
			goto cleanup
		}

		// Figure out how many spaces to add
		spaceToAdd = CurrentFrame.MarginRight - CurrentFrame.Dot.Line.Used
		// Find number of holes for space distribution
		startChar = CurrentFrame.MarginLeft
		for (CurrentFrame.Dot.Line.Str.Get(startChar) == ' ') &&
			(startChar < CurrentFrame.Dot.Line.Used) {
			startChar++
		}
		endChar = startChar
		holes = 0
		for {
			for (CurrentFrame.Dot.Line.Str.Get(startChar) != ' ') &&
				(startChar < CurrentFrame.Dot.Line.Used) {
				startChar++
			}
			for (CurrentFrame.Dot.Line.Str.Get(startChar) == ' ') &&
				(startChar < CurrentFrame.Dot.Line.Used) {
				startChar++
			}
			holes++
			if !(startChar < CurrentFrame.Dot.Line.Used) {
				break
			}
		}
		holes--
		if holes > 0 {
			fillRatio = float64(spaceToAdd) / float64(holes)
		}
		debit = 0.0
		startChar = endChar
		for i = 1; i <= holes; i++ {
			// Find a hole
			for CurrentFrame.Dot.Line.Str.Get(startChar) != ' ' {
				startChar++
			}
			debit += fillRatio
			spaceToAdd = int(debit + 0.5)
			if spaceToAdd > 0 {
				here = nil
				if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
					goto cleanup
				}
				if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
					goto cleanup
				}
				if !MarkDestroy(&here) {
					goto cleanup
				}
				debit -= float64(spaceToAdd)
			}
			for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
				startChar++
			}
		}
	nextLine:
		count--
		if !MarkCreate(
			CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot,
		) {
			goto cleanup
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified],
		) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	return result
}

// WordSqueeze removes multiple spaces from lines
func WordSqueeze(rept LeadParam, count int) bool {
	var startChar int
	var endChar int
	var lineCount int
	var thisLine *LineHdrObject
	var here *MarkObject
	var there *MarkObject

	result := false
	here = nil
	there = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			// on EOP line so abort
			goto cleanup
		}
		startChar = 1
		for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar += 1
		}
		// with line^ do
		for {
			for CurrentFrame.Dot.Line.Str.Get(startChar) != ' ' &&
				startChar < CurrentFrame.Dot.Line.Used {
				startChar++
			}
			if CurrentFrame.Dot.Line.Str.Get(startChar) != ' ' {
				break // Nothing more to do
			}
			endChar = startChar
			for CurrentFrame.Dot.Line.Str.Get(endChar) == ' ' {
				endChar++
			}
			if (endChar - startChar) > 1 {
				here = nil
				if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
					goto cleanup
				}
				there = nil
				if !MarkCreate(CurrentFrame.Dot.Line, endChar-1, &there) {
					goto cleanup
				}
				if !TextRemove(here, there) {
					goto cleanup
				}
				startChar = here.Col
			} else {
				startChar = endChar
			}
		}

		count -= 1
		if !MarkCreate(CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot) {
			goto cleanup
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified]) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	if there != nil {
		MarkDestroy(&there)
	}
	return result
}

// WordRight right-aligns text
func WordRight(rept LeadParam, count int) bool {
	var startChar int
	var lineCount int
	var spaceToAdd int
	var thisLine *LineHdrObject
	var here *MarkObject
	var there *MarkObject

	result := false
	here = nil
	there = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			goto cleanup
		}
		if (CurrentFrame.Dot.Line.Used < CurrentFrame.MarginLeft) ||
			(CurrentFrame.Dot.Line.Used > CurrentFrame.MarginRight) {
			goto cleanup
		}
		startChar = 1
		for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar < CurrentFrame.MarginLeft {
			goto cleanup
		}
		spaceToAdd = CurrentFrame.MarginRight - CurrentFrame.Dot.Line.Used
		if spaceToAdd > 0 {
			if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
				goto cleanup
			}
			if !TextInsert(true, 1, BlankString, spaceToAdd, here) {
				goto cleanup
			}
			if !MarkDestroy(&here) {
				goto cleanup
			}
		} else if spaceToAdd < 0 {
			if !MarkCreate(CurrentFrame.Dot.Line, startChar, &there) {
				goto cleanup
			}
			if !MarkCreate(CurrentFrame.Dot.Line, startChar-spaceToAdd, &here) {
				goto cleanup
			}
			if !TextRemove(there, here) {
				goto cleanup
			}
			if !MarkDestroy(&here) {
				goto cleanup
			}
			if !MarkDestroy(&there) {
				goto cleanup
			}
		}
		count--
		if !MarkCreate(
			CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot,
		) {
			goto cleanup
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified],
		) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	if there != nil {
		MarkDestroy(&there)
	}
	return result
}

// WordLeft left-aligns text
func WordLeft(rept LeadParam, count int) bool {
	var startChar int
	var lineCount int
	var thisLine *LineHdrObject
	var here *MarkObject
	var there *MarkObject

	result := false
	here = nil
	there = nil
	if rept == LeadParamPIndef {
		count = MaxInt
	}
	if (rept == LeadParamNone) || (rept == LeadParamPlus) {
		count = 1
		rept = LeadParamPInt
	}
	if rept == LeadParamPInt {
		lineCount = count
		thisLine = CurrentFrame.Dot.Line
		for (lineCount > 0) && (thisLine.Used > 0) {
			thisLine = thisLine.FLink
			lineCount--
			if thisLine == nil {
				goto cleanup
			}
		}
		if lineCount != 0 {
			goto cleanup
		}
	}
	for (count > 0) && (CurrentFrame.Dot.Line.Used > 0) {
		if CurrentFrame.Dot.Line.FLink == nil {
			goto cleanup
		}
		if (CurrentFrame.Dot.Line.Used < CurrentFrame.MarginLeft) ||
			(CurrentFrame.Dot.Line.Used > CurrentFrame.MarginRight) {
			goto cleanup
		}
		startChar = 1
		for CurrentFrame.Dot.Line.Str.Get(startChar) == ' ' {
			startChar++
		}
		if startChar != CurrentFrame.MarginLeft {
			if startChar < CurrentFrame.MarginLeft {
				if !MarkCreate(CurrentFrame.Dot.Line, startChar, &here) {
					goto cleanup
				}
				if !TextInsert(
					true, 1, BlankString, CurrentFrame.MarginLeft-startChar, here,
				) {
					goto cleanup
				}
				if !MarkDestroy(&here) {
					goto cleanup
				}
			} else {
				if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.MarginLeft, &here) {
					goto cleanup
				}
				if !MarkCreate(CurrentFrame.Dot.Line, startChar, &there) {
					goto cleanup
				}
				if !TextRemove(here, there) {
					goto cleanup
				}
				if !MarkDestroy(&here) {
					goto cleanup
				}
				if !MarkDestroy(&there) {
					goto cleanup
				}
			}
		}
		count--
		if !MarkCreate(
			CurrentFrame.Dot.Line.FLink, CurrentFrame.MarginLeft, &CurrentFrame.Dot,
		) {
			goto cleanup
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified],
		) {
			goto cleanup
		}
	}
	result = (count <= 0) || (rept == LeadParamPIndef)
cleanup:
	if here != nil {
		MarkDestroy(&here)
	}
	if there != nil {
		MarkDestroy(&there)
	}
	return result
}

// WordAdvanceWord advances cursor to start of a word
func WordAdvanceWord(rept LeadParam, count int) bool {
	var thisLine *LineHdrObject
	var pos int

	result := false
	if rept == LeadParamMarker {
		ScreenMessage(MsgSyntaxError)
		goto cleanup
	}
	pos = CurrentFrame.Dot.Col
	thisLine = CurrentFrame.Dot.Line
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
							goto success
						}
						goto cleanup
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
	success:
		if !MarkCreate(thisLine, pos, &CurrentFrame.Dot) {
			goto cleanup
		}
		result = true
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
				goto cleanup
			}
			thisLine = thisLine.FLink
		}
		for thisLine.Str.Get(pos) == ' ' {
			pos++
		}
		if !MarkCreate(thisLine, pos, &CurrentFrame.Dot) {
			goto cleanup
		}
		result = true
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
						goto cleanup
					}
					thisLine = thisLine.BLink
					pos = thisLine.Used
					if !(pos <= 0) {
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
						goto cleanup
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
		if !MarkCreate(thisLine, pos, &CurrentFrame.Dot) {
			goto cleanup
		}
		result = true
	}
cleanup:
	return result
}

// WordDeleteWord deletes words at cursor
func WordDeleteWord(rept LeadParam, count int) bool {
	var oldPos *MarkObject
	var here *MarkObject
	var anotherMark *MarkObject
	var theOtherMark *MarkObject
	var oldDotCol int
	var lineNr int
	var newLineNr int

	result := false
	oldPos = nil
	here = nil
	theOtherMark = nil
	if rept == LeadParamMarker {
		ScreenMessage(MsgSyntaxError)
		goto cleanup
	}
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &oldPos) {
		goto cleanup
	}
	// Get to beginning of word if in middle
	if !WordAdvanceWord(LeadParamPInt, 0) {
		goto cleanup
	}
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &here) {
		goto cleanup
	}
	if !WordAdvanceWord(rept, count) {
		// Put dot back and bail out
		if !MarkCreate(oldPos.Line, oldPos.Col, &CurrentFrame.Dot) {
			goto cleanup
		}
		goto cleanup
	}
	// Wipe out everything from dot to here
	oldDotCol = CurrentFrame.Dot.Col
	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &theOtherMark) {
		goto cleanup
	}
	if !LineToNumber(CurrentFrame.Dot.Line, &lineNr) {
		goto cleanup
	}
	if !LineToNumber(here.Line, &newLineNr) {
		goto cleanup
	}
	if (lineNr > newLineNr) ||
		((lineNr == newLineNr) && (CurrentFrame.Dot.Col > here.Col)) {
		// Reverse mark pointers
		anotherMark = here
		here = theOtherMark
		theOtherMark = anotherMark
	}
	if CurrentFrame != FrameOops {
		// Make sure oops_span is okay
		if !MarkCreate(FrameOops.LastGroup.LastLine, 1, &FrameOops.Span.MarkTwo) {
			goto cleanup
		}
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
		result = TextSplitLine(CurrentFrame.Dot, oldDotCol, &here)
	}
cleanup:
	if oldPos != nil {
		MarkDestroy(&oldPos)
	}
	if here != nil {
		MarkDestroy(&here)
	}
	return result
}
