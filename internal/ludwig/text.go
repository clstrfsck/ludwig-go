/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         TEXT
//
// Description:  Text manipulation routines.

package ludwig

// TextReturnCol calculates which column to place dot on after a return/split
func TextReturnCol(curLine *LineHdrObject, curCol int, splitting bool) int {
	newLine := curLine
	var newCol int

	if curCol >= curLine.Group.Frame.MarginLeft {
		newCol = curLine.Group.Frame.MarginLeft
	} else {
		newCol = 1
	}

	if curLine.Group.Frame.Options.Has(OptAutoIndent) && (newLine.FLink != nil) {
		str1 := newLine.Str
		used1 := newLine.Used
		str2 := newLine.Str
		used2 := newLine.Used

		if (newLine.FLink.FLink != nil) && !splitting {
			str2 = newLine.FLink.Str
			used2 = newLine.FLink.Used
		}

		for {
			if newCol <= used1 {
				if str1.Get(newCol) != ' ' {
					break
				}
			}
			if newCol <= used2 {
				if str2.Get(newCol) != ' ' {
					break
				}
			} else {
				if newCol >= used1 {
					break
				}
			}
			newCol++
		}
	}
	return newCol
}

// TextRealizeNull converts a null line into a real line
func TextRealizeNull(oldNull *LineHdrObject) bool {
	var newNull *LineHdrObject
	if LinesCreate(1, &newNull, &newNull) {
		if LinesInject(newNull, newNull, oldNull) {
			if MarksShift(oldNull, 1, MaxStrLenP, newNull, 1) {
				newNull.Group.Frame.TextModified = true
				if MarkCreate(
					newNull.Group.Frame.Dot.Line,
					newNull.Group.Frame.Dot.Col,
					&newNull.Group.Frame.Marks[MarkModified-MinMarkNumber],
				) {
					return true
				}
			}
		}
	}
	return false
}

// TextInsert inserts text at a specified position
func TextInsert(
	updateScreen bool,
	count int,
	buf StrObject,
	bufLen int,
	dst *MarkObject,
) bool {
	insertLen := count * bufLen
	if insertLen > 0 {
		dstLine := dst.Line
		dstCol := dst.Col

		finalLen := dstCol - 1 + insertLen
		tailLen := dstLine.Used + 1 - dstCol
		if tailLen <= 0 {
			tailLen = 0
		} else {
			finalLen += tailLen
		}
		if finalLen > MaxStrLen {
			return false
		}
		if dstLine.FLink == nil {
			if !TextRealizeNull(dstLine) {
				return false
			}
			dstLine = dstLine.BLink
		}

		if finalLen > dstLine.Len {
			if !LineChangeLength(dstLine, finalLen) {
				return false
			}
		}
		if !MarksShift(dstLine, dstCol, MaxStrLenP-dstCol, dstLine, dstCol+insertLen) {
			return false
		}
		if tailLen > 0 {
			// Shift tail to make room
			for i := dstLine.Used; i >= dstCol; i-- {
				dstLine.Str.Set(i+insertLen, dstLine.Str.Get(i))
			}
		}
		newCol := dstCol
		for i := 0; i < count; i++ {
			dstLine.Str.Copy(&buf, 1, bufLen, newCol)
			newCol += bufLen
		}

		// Re-compute length of line, to remove trailing spaces
		if tailLen == 0 {
			dstLine.Used = dstLine.Str.Length(' ', dstLine.Len)
		} else {
			dstLine.Used += insertLen
		}

		// Update screen if necessary
		if updateScreen && dstLine.ScrRowNr != 0 {
			scrCol := dstCol - dstLine.Group.Frame.ScrOffset
			if scrCol <= dstLine.Group.Frame.ScrWidth {
				if scrCol <= 0 {
					scrCol = 1
				}
				VduMoveCurs(scrCol, dstLine.ScrRowNr)
				firstColRedraw := dstCol
				var lastColRedraw int
				if firstColRedraw <= dstLine.Group.Frame.ScrOffset {
					firstColRedraw = dstLine.Group.Frame.ScrOffset + 1
				}
				if (scrCol+insertLen <= dstLine.Group.Frame.ScrWidth) &&
					(scrCol+dstLine.Group.Frame.ScrOffset <= dstLine.Used) {
					VduInsertChars(insertLen)
					lastColRedraw = firstColRedraw + insertLen - 1
				} else {
					lastColRedraw = dstLine.Used
				}
				if lastColRedraw > dstLine.Group.Frame.ScrWidth+dstLine.Group.Frame.ScrOffset {
					lastColRedraw = dstLine.Group.Frame.ScrWidth + dstLine.Group.Frame.ScrOffset
				}
				lenRedraw := lastColRedraw - firstColRedraw + 1
				if lenRedraw > 0 {
					VduDisplayStr(dstLine.Str.Slice(firstColRedraw, lenRedraw), 0)
				}
			}
		}
	}
	return true
}

// TextOvertype overtypes text at a specified position
func TextOvertype(
	updateScreen bool,
	count int,
	buf StrObject,
	bufLen int,
	dst *MarkObject,
) bool {
	overtypeLen := count * bufLen
	if overtypeLen > 0 {
		dstLine := dst.Line
		finalLen := dst.Col + overtypeLen - 1
		if finalLen > MaxStrLen {
			return false
		}
		if dstLine.FLink == nil {
			if !TextRealizeNull(dst.Line) {
				return false
			}
			dstLine = dstLine.BLink
		}

		if finalLen > dstLine.Len {
			if !LineChangeLength(dstLine, finalLen) {
				return false
			}
		}
		newCol := dst.Col
		for i := 0; i < count; i++ {
			dstLine.Str.Copy(&buf, 1, bufLen, newCol)
			newCol += bufLen
		}

		// Re-compute length of line, to remove trailing spaces
		if newCol > dstLine.Used {
			dstLine.Used = dstLine.Str.Length(' ', dstLine.Len)
		}

		// Update screen if necessary
		if updateScreen && dstLine.ScrRowNr != 0 {
			firstColOnScr := dst.Col
			if firstColOnScr <= dstLine.Group.Frame.ScrOffset {
				firstColOnScr = dstLine.Group.Frame.ScrOffset + 1
			}
			lastColOnScr := newCol
			if lastColOnScr > dstLine.Group.Frame.ScrWidth+dstLine.Group.Frame.ScrOffset {
				lastColOnScr = dstLine.Group.Frame.ScrWidth + dstLine.Group.Frame.ScrOffset + 1
			}

			lenOnScr := lastColOnScr - firstColOnScr
			if lenOnScr > 0 {
				VduMoveCurs(firstColOnScr-dstLine.Group.Frame.ScrOffset, dstLine.ScrRowNr)
				VduDisplayStr(dstLine.Str.Slice(firstColOnScr, lenOnScr), 0)
			}
		}

		// Update the destination mark
		dst.Col += overtypeLen
	}
	return true
}

// TextInsertTpar inserts a trailing parameter at a mark position
func TextInsertTpar(tp *TParObject, beforeMark *MarkObject, equalsMark **MarkObject) bool {
	result := false
	var firstLine *LineHdrObject
	var lastLine *LineHdrObject
	discard := false

	// Check for the simple case
	if tp.Con == nil {
		if !TextInsert(true, 1, tp.Str, tp.Len, beforeMark) {
			ScreenMessage(MsgNoRoomOnLine)
			return false
		}
		if !MarkCreate(beforeMark.Line, beforeMark.Col-tp.Len, equalsMark) {
			return false
		}
	} else {
		if beforeMark.Col+tp.Len > MaxStrLen {
			ScreenMessage(MsgNoRoomOnLine)
			return false
		}
		lineCount := 0
		tmpTp := tp.Con
		for tmpTp.Con != nil {
			lineCount++
			tmpTp = tmpTp.Con
		}
		if tmpTp.Len+(beforeMark.Line.Used-beforeMark.Col) > MaxStrLen {
			ScreenMessage(MsgNoRoomOnLine)
			return false
		}
		if !LinesCreate(lineCount, &firstLine, &lastLine) {
			return false
		}
		discard = true
		if beforeMark.Line.FLink == nil && !TextRealizeNull(beforeMark.Line) {
			goto cleanup
		}
		if !TextSplitLine(beforeMark, 1, equalsMark) {
			goto cleanup
		}
		if !TextInsert(true, 1, tp.Str, tp.Len, *equalsMark) {
			goto cleanup
		}
		(*equalsMark).Col -= tp.Len
		tmpTp = tp.Con
		tmpLine := firstLine
		for lc := 0; lc < lineCount; lc++ {
			if !LineChangeLength(tmpLine, tmpTp.Len) {
				goto cleanup
			}
			tmpLine.Str.Copy(&tmpTp.Str, 1, tmpTp.Len, 1)
			if tmpTp.Len == 0 {
				tmpLine.Used = 0
			} else {
				tmpLine.Used = tmpLine.Str.Length(' ', tmpTp.Len)
			}
			tmpTp = tmpTp.Con
			tmpLine = tmpLine.FLink
		}
		if lineCount != 0 {
			if !LinesInject(firstLine, lastLine, beforeMark.Line) {
				goto cleanup
			}
		}
		discard = false
		if !TextInsert(true, 1, tmpTp.Str, tmpTp.Len, beforeMark) {
			return false
		}
	}
	result = true
cleanup:
	if discard {
		LinesDestroy(&firstLine, &lastLine)
	}
	return result
}

// textIntraRemove removes text within a single line
func textIntraRemove(markOne *MarkObject, size int) bool {
	ln := markOne.Line
	colOne := markOne.Col
	colTwo := colOne + size
	if !MarksSqueeze(ln, colOne, ln, colTwo) {
		return false
	}
	if !MarksShift(ln, colTwo, MaxStrLenP+1-colTwo, ln, colOne) {
		return false
	}
	if size == 0 {
		return true
	}

	oldUsed := ln.Used
	if colOne > oldUsed {
		return true
	}
	dstLen := oldUsed + 1 - colOne
	if colTwo <= oldUsed {
		// Shift data down
		for i := 0; i < oldUsed+1-colTwo; i++ {
			ln.Str.Set(colOne+i, ln.Str.Get(colTwo+i))
		}
		// Fill remainder with spaces
		for i := oldUsed + 1 - colTwo; i < dstLen; i++ {
			ln.Str.Set(colOne+i, ' ')
		}
	} else {
		// Fill with spaces
		for i := 0; i < dstLen; i++ {
			ln.Str.Set(colOne+i, ' ')
		}
	}
	ln.Used = ln.Str.Length(' ', oldUsed)

	// Update screen if necessary
	if ln.ScrRowNr == 0 {
		return true
	}
	offsetPWidth := ln.Group.Frame.ScrOffset + ln.Group.Frame.ScrWidth
	if colOne > offsetPWidth {
		return true
	}
	distance := colTwo - colOne
	firstColOnScr := colOne
	if colOne <= ln.Group.Frame.ScrOffset {
		firstColOnScr = ln.Group.Frame.ScrOffset + 1
	}

	// If possible, drag any characters on screen to final place
	if (firstColOnScr+distance <= offsetPWidth) && (firstColOnScr+distance <= oldUsed) {
		VduMoveCurs(firstColOnScr-ln.Group.Frame.ScrOffset, ln.ScrRowNr)
		VduDeleteChars(distance)
		firstColOnScr = offsetPWidth + 1 - distance
		if firstColOnScr > ln.Used {
			return true
		}
	}

	// Fix the remainder of the line's appearance on the screen
	VduMoveCurs(firstColOnScr-ln.Group.Frame.ScrOffset, ln.ScrRowNr)
	var bufLen int
	if ln.Used <= ln.Group.Frame.ScrOffset+ln.Group.Frame.ScrWidth {
		bufLen = ln.Used + 1 - firstColOnScr
	} else {
		bufLen = ln.Group.Frame.ScrOffset + ln.Group.Frame.ScrWidth + 1 - firstColOnScr
	}
	if bufLen <= 0 {
		VduClearEOL()
	} else {
		VduDisplayStr(ln.Str.Slice(firstColOnScr, bufLen), 3)
	}
	return true
}

// textInterRemove removes text spanning multiple lines
func textInterRemove(markOne *MarkObject, markTwo *MarkObject) bool {
	result := false
	var markStart *MarkObject
	var extrOne *LineHdrObject
	var extrTwo *LineHdrObject
	var textLen int
	var strng StrObject
	var strngTail StrObject
	var delta int
	var lineOne *LineHdrObject
	var colOne int

	if (markTwo.Line.FLink == nil) && (markOne.Col != 1) {
		lineOne = markOne.Line
		colOne = markOne.Col
		extrOne = lineOne.FLink
		extrTwo = markTwo.Line
		if !textIntraRemove(markOne, MaxStrLenP-markOne.Col) {
			goto cleanup
		}
		if !MarksSqueeze(lineOne, colOne, markTwo.Line, markTwo.Col) {
			goto cleanup
		}
		if !MarksShift(markTwo.Line, markTwo.Col, MaxStrLenP+1-markTwo.Col, lineOne, colOne) {
			goto cleanup
		}
		if extrOne == extrTwo {
			goto success
		}
		extrTwo = extrTwo.BLink
		goto extract
	}

	// Bring the start of lineOne down to replace the start of lineTwo
	textLen = markOne.Line.Used
	if markOne.Col <= textLen {
		textLen = markOne.Col - 1
	}
	if markOne.Col > 1 {
		ChFillCopy(markOne.Line.Str, 1, textLen, &strng, 1, markOne.Col-1, ' ')
	}
	textLen = markOne.Col - 1
	delta = markOne.Col - markTwo.Col
	if delta < 0 {
		if !MarkCreate(markTwo.Line, markOne.Col, &markStart) {
			goto cleanup
		}
		if !textIntraRemove(markStart, markTwo.Col-markStart.Col) {
			goto cleanup
		}
	} else if delta > 0 {
		strngTail.Copy(&strng, markTwo.Col, delta, 1)
		if !TextInsert(true, 1, strngTail, delta, markTwo) {
			goto cleanup
		}
		textLen -= delta
	}
	if !MarkCreate(markTwo.Line, 1, &markStart) {
		goto cleanup
	}
	if textLen > 0 {
		if !TextOvertype(true, 1, strng, textLen, markStart) {
			goto cleanup
		}
	}
	colOne = markOne.Col
	extrOne = markOne.Line
	extrTwo = markTwo.Line.BLink
	if !MarksSqueeze(extrOne, colOne, markTwo.Line, markTwo.Col) {
		goto cleanup
	}
	if colOne > 1 {
		if !MarksShift(extrOne, 1, colOne-1, markTwo.Line, 1) {
			goto cleanup
		}
	}
extract:
	if !LinesExtract(extrOne, extrTwo) {
		goto cleanup
	}
	if !LinesDestroy(&extrOne, &extrTwo) {
		goto cleanup
	}
success:
	result = true
cleanup:
	if markStart != nil {
		MarkDestroy(&markStart)
	}
	return result
}

// TextRemove removes text between two marks
func TextRemove(markOne *MarkObject, markTwo *MarkObject) bool {
	if markOne.Line == markTwo.Line {
		return textIntraRemove(markOne, markTwo.Col-markOne.Col)
	}
	return textInterRemove(markOne, markTwo)
}

// textIntraMove moves/copies text within a single line
func textIntraMove(
	copy bool,
	count int,
	markOne *MarkObject,
	markTwo *MarkObject,
	dst *MarkObject,
	newStart **MarkObject,
	newEnd **MarkObject,
) bool {
	colOne := markOne.Col
	colTwo := markTwo.Col

	fullLen := colTwo - colOne
	var textStr StrObject
	if fullLen != 0 {
		if fullLen*count > MaxStrLen {
			return false
		}
		textLen := fullLen
		if colOne > markOne.Line.Used {
			textLen = 0
		} else {
			if colTwo > markOne.Line.Used {
				textLen = markOne.Line.Used + 1 - colOne
			}
		}
		ChFillCopy(markOne.Line.Str, colOne, textLen, &textStr, 1, fullLen, ' ')
		textLen = fullLen
		for i := 1; i < count; i++ {
			textStr.Copy(&textStr, 1, textLen, 1+fullLen)
			fullLen += textLen
			if TtControlC {
				return false
			}
		}
	}
	if !copy {
		dstCol := dst.Col
		dstUsed := dst.Line.Used
		if markOne.Line == dst.Line {
			if dstCol > colTwo {
				dstCol = dstCol - (colTwo - colOne)
			} else if dstCol > colOne {
				dstCol = colOne
			}
			if dstUsed > colTwo {
				dstUsed -= colTwo - colOne
			} else {
				dstUsed = colOne - 1
			}
		}
		tailLen := 0
		if dstCol <= dstUsed {
			tailLen = dstUsed + 1 - dstCol
		}
		if dstCol+fullLen+tailLen > MaxStrLenP {
			return false
		}
		if fullLen != 0 {
			if !textIntraRemove(markOne, markTwo.Col-markOne.Col) {
				return false
			}
		}
	}
	if fullLen != 0 {
		if !TextInsert(true, 1, textStr, fullLen, dst) {
			return false
		}
	}
	dstCol := dst.Col
	if !MarkCreate(dst.Line, dstCol-fullLen, newStart) {
		return false
	}
	if !MarkCreate(dst.Line, dstCol, newEnd) {
		return false
	}
	return true
}

// textInterMove moves/copies text spanning multiple lines
func textInterMove(
	copy bool,
	count int,
	markOne *MarkObject,
	markTwo *MarkObject,
	dst *MarkObject,
	newStart **MarkObject,
	newEnd **MarkObject,
) bool {
	var lastLine *LineHdrObject
	var nextSrcLine *LineHdrObject
	var nextDstLine *LineHdrObject
	var firstNicked *LineHdrObject
	var lastNicked *LineHdrObject
	var lineOneNr int
	var lineTwoNr int
	var lineDstNr int
	var linesRequired int
	var lastLineLength int
	var tempLen int
	var textLen int
	var textStr StrObject
	var dstCol int
	var dstUsed int
	var dstLine *LineHdrObject
	var i int

	result := false
	var firstLine *LineHdrObject
	lineOne := markOne.Line
	colOne := markOne.Col
	lineTwo := markTwo.Line
	colTwo := markTwo.Col
	if !LineToNumber(lineOne, &lineOneNr) {
		goto cleanup
	}
	if !LineToNumber(lineTwo, &lineTwoNr) {
		goto cleanup
	}

	// Predict dst.Col and dst.Line.Used just before insertion
	dstCol = dst.Col
	dstUsed = dst.Line.Used
	if !copy && dst.Line.Group.Frame == lineOne.Group.Frame {
		if !LineToNumber(dst.Line, &lineDstNr) {
			goto cleanup
		}
		if (lineOneNr <= lineDstNr) && (lineDstNr <= lineTwoNr) {
			if (lineTwoNr == lineDstNr) && (dstCol >= colTwo) {
				dstCol = colOne + dstCol - colTwo
			} else if (lineOneNr != lineDstNr) || (dstCol >= colOne) {
				dstCol = colOne
			}
			tempLen = 0
			if colTwo <= lineTwo.Used {
				tempLen = lineTwo.Used + 1 - colTwo
			}
			dstUsed = colOne - 1 + tempLen
		}
	}
	if (colTwo <= lineTwo.Used) && (colOne+lineTwo.Used-colTwo > MaxStrLenP) {
		goto cleanup
	}
	if (colOne <= lineOne.Used) && (dstCol+lineOne.Used-colOne > MaxStrLenP) {
		goto cleanup
	}
	if (dstCol <= dstUsed) && (colTwo+dstUsed-dstCol > MaxStrLen) {
		goto cleanup
	}
	if (count > 1) && (colOne <= lineOne.Used) && (colTwo+lineOne.Used-colOne > MaxStrLen) {
		goto cleanup
	}

	// Create extra lines required
	linesRequired = count * (lineTwoNr - lineOneNr)
	if !copy {
		linesRequired -= lineTwoNr - lineOneNr - 1
	}
	if !LinesCreate(linesRequired, &firstLine, &lastLine) {
		goto cleanup
	}

	// Copy end of first line
	textLen = 0
	if colOne <= lineOne.Used {
		textLen = lineOne.Used + 1 - colOne
		textStr.Copy(lineOne.Str, colOne, textLen, 1)
	}

	// Take count copies
	nextDstLine = firstLine
	for i = count - 1; i >= 0; i-- {
		if TtControlC {
			goto cleanup
		}
		nextSrcLine = lineOne.FLink
		if i == 0 {
			if !copy && (lineTwoNr-lineOneNr > 1) {
				firstNicked = lineOne.FLink
				lastNicked = lineTwo.BLink
				if !MarksSqueeze(firstNicked, 1, lastNicked.FLink, 1) {
					goto cleanup
				}
				if !LinesExtract(firstNicked, lastNicked) {
					goto cleanup
				}
				lastNicked.FLink = nextDstLine
				firstNicked.BLink = nextDstLine.BLink
				if firstNicked.BLink != nil {
					firstNicked.BLink.FLink = firstNicked
				}
				nextDstLine.BLink = lastNicked
				if nextDstLine == firstLine {
					firstLine = firstNicked
				}
				nextSrcLine = lineTwo
			}
		}

		// Copy interior lines
		for nextSrcLine != lineTwo {
			if !LineChangeLength(nextDstLine, nextSrcLine.Used) {
				goto cleanup
			}
			nextDstLine.Str.Copy(nextSrcLine.Str, 1, nextSrcLine.Used, 1)
			nextDstLine.Used = nextSrcLine.Used
			nextSrcLine = nextSrcLine.FLink
			nextDstLine = nextDstLine.FLink
		}

		// Copy the last line
		if i != 0 {
			if !LineChangeLength(nextDstLine, colTwo-1+textLen) {
				goto cleanup
			}
			nextDstLine.Str.Copy(&textStr, 1, textLen, colTwo)
		} else {
			if !LineChangeLength(nextDstLine, colTwo-1) {
				goto cleanup
			}
		}
		if colTwo > 1 {
			if colTwo <= nextSrcLine.Used {
				nextDstLine.Str.Copy(nextSrcLine.Str, 1, colTwo-1, 1)
			} else {
				ChFillCopy(nextSrcLine.Str, 1, nextSrcLine.Used, nextDstLine.Str, 1, colTwo-1, ' ')
			}
		}
		if i != 0 {
			nextDstLine.Used = nextDstLine.Str.Length(' ', colTwo-1+textLen)
		} else {
			nextDstLine.Used = colTwo - 1
		}
		nextDstLine = nextDstLine.FLink
	}

	// Remove original if necessary
	if !copy && !textInterRemove(markOne, markTwo) {
		goto cleanup
	}

	// Insert source text into destination
	dstLine = dst.Line
	dstCol = dst.Col

	// Complete the last line with rest of destination line
	lastLineLength = lastLine.Used
	i = dstLine.Used + 1 - dstCol
	if i > 0 {
		if !LineChangeLength(lastLine, lastLine.Used+i) {
			goto cleanup
		}
		lastLine.Str.Copy(dstLine.Str, dstCol, i, lastLineLength+1)
		lastLine.Used = lastLine.Str.Length(' ', lastLineLength+i)
	} else if lastLineLength > 0 {
		lastLine.Used = lastLine.Str.Length(' ', lastLineLength)
	}

	// Special case for NULL line
	if dstLine.FLink == nil {
		if (dstCol != 1) || (lastLineLength != 0) {
			if !TextRealizeNull(dstLine) {
				goto cleanup
			}
			dstLine = dstLine.BLink
		} else {
			if firstLine != lastLine {
				firstNicked = lastLine
				lastLine = lastLine.BLink
				lastLine.FLink = nil
				firstNicked.BLink = nil
				firstNicked.FLink = firstLine
				firstLine.BLink = firstNicked
				firstLine = firstNicked
			}

			if textLen > 0 {
				if !LineChangeLength(firstLine, textLen) {
					goto cleanup
				}
				ChFillCopy(&textStr, 1, textLen, firstLine.Str, 1, firstLine.Len, ' ')
				firstLine.Used = textLen
			}
			if !LinesInject(firstLine, lastLine, dstLine) {
				goto cleanup
			}

			lastLine = dstLine
			dstLine = firstLine
			firstLine = nil
			goto finished
		}
	}

	if !LinesInject(firstLine, lastLine, dstLine.FLink) {
		goto cleanup
	}
	if !MarksShift(dstLine, dstCol, MaxStrLenP+1-dstCol, lastLine, colTwo) {
		goto cleanup
	}
	firstLine = nil
	if textLen > 0 {
		if !LineChangeLength(dstLine, dstCol+textLen-1) {
			goto cleanup
		}
		ChFillCopy(&textStr, 1, textLen, dstLine.Str, dstCol, dstLine.Len+1-dstCol, ' ')
		dstLine.Used = dstCol + textLen - 1
		if dstLine.ScrRowNr != 0 {
			ScreenDrawLine(dstLine)
		}
	} else if dstCol <= dstLine.Used {
		dstLine.Str.Fill(' ', dstCol, dstLine.Used)
		dstLine.Used = dstLine.Str.Length(' ', dstCol)
		if dstLine.ScrRowNr != 0 {
			ScreenDrawLine(dstLine)
		}
	}

finished:
	if !MarkCreate(dstLine, dstCol, newStart) {
		goto cleanup
	}
	if !MarkCreate(lastLine, colTwo, newEnd) {
		goto cleanup
	}
	result = true
cleanup:
	if firstLine != nil {
		LinesDestroy(&firstLine, &lastLine)
	}
	return result
}

// TextMove moves or copies text between two marks to a destination
func TextMove(
	copy bool,
	count int,
	markOne *MarkObject,
	markTwo *MarkObject,
	dst *MarkObject,
	newStart **MarkObject,
	newEnd **MarkObject,
) bool {
	if count > 0 {
		var cmdSuccess bool
		if markOne.Line == markTwo.Line {
			cmdSuccess = textIntraMove(copy, count, markOne, markTwo, dst, newStart, newEnd)
		} else {
			cmdSuccess = textInterMove(copy, count, markOne, markTwo, dst, newStart, newEnd)
		}
		if TtControlC {
			return false
		}
		if !cmdSuccess {
			ScreenMessage(MsgNoRoomOnLine)
			return false
		}
		if !copy {
			markTwo.Line.Group.Frame.TextModified = true
			if !MarkCreate(
				markTwo.Line,
				markTwo.Col,
				&markTwo.Line.Group.Frame.Marks[MarkModified-MinMarkNumber],
			) {
				return false
			}
		}
		(*newEnd).Line.Group.Frame.TextModified = true
		if !MarkCreate(
			(*newEnd).Line,
			(*newEnd).Col,
			&(*newEnd).Line.Group.Frame.Marks[MarkModified-MinMarkNumber],
		) {
			return false
		}
	}
	return true
}

// TextSplitLine splits a line at a mark position
func TextSplitLine(beforeMark *MarkObject, newCol int, equalsMark **MarkObject) bool {
	var saveCol int
	var length int
	var shift int
	var newLine *LineHdrObject
	var cost int
	var equalsCol int
	var equalsLine *LineHdrObject

	result := false
	discard := false
	if beforeMark.Line.FLink == nil {
		ScreenMessage(MsgCantSplitNullLine)
		goto cleanup
	}
	if newCol == 0 {
		newCol = TextReturnCol(beforeMark.Line, beforeMark.Col, true)
	}
	length = beforeMark.Line.Used + 1 - beforeMark.Col
	if length <= 0 {
		length = 0
	} else {
		if (newCol + length) > MaxStrLenP {
			ScreenMessage(MsgNoRoomOnLine)
			goto cleanup
		}
	}

	if !LinesCreate(1, &newLine, &newLine) {
		goto cleanup
	}
	discard = true

	// Heuristic to decide which way to do the split
	shift = newCol - beforeMark.Col
	cost = MaxInt
	if (beforeMark.Col <= beforeMark.Line.Used) && (beforeMark.Line.ScrRowNr != 0) {
		if shift == 0 {
			cost = beforeMark.Col + beforeMark.Col
		} else if shift > 0 {
			cost = beforeMark.Col + beforeMark.Col + 3*shift
		} else {
			cost = beforeMark.Col + beforeMark.Col - 3*shift
		}
	}

	// Do the split
	if 2*length < cost {
		// Move end to next (new) line
		equalsCol = beforeMark.Col
		equalsLine = beforeMark.Line
		if length > 0 {
			if !LineChangeLength(newLine, newCol+length-1) {
				goto cleanup
			}
			newLine.Str.FillN(' ', newCol-1, 1)
			newLine.Str.Copy(beforeMark.Line.Str, beforeMark.Col, length, newCol)
			beforeMark.Line.Str.Fill(' ', beforeMark.Col, beforeMark.Col+length-1)
			beforeMark.Line.Used = beforeMark.Line.Str.Length(' ', beforeMark.Line.Used)
			newLine.Used = newCol + length - 1
			if beforeMark.Line.ScrRowNr != 0 {
				if beforeMark.Line.Used <= beforeMark.Line.Group.Frame.ScrOffset {
					VduMoveCurs(1, beforeMark.Line.ScrRowNr)
					VduClearEOL()
				} else if beforeMark.Line.Used+1 <= beforeMark.Line.Group.Frame.ScrOffset+beforeMark.Line.Group.Frame.ScrWidth {
					VduMoveCurs(beforeMark.Line.Used+1-beforeMark.Line.Group.Frame.ScrOffset, beforeMark.Line.ScrRowNr)
					VduClearEOL()
				}
			}
		}
		if !LinesInject(newLine, newLine, beforeMark.Line.FLink) {
			goto cleanup
		}
		discard = false
		if !MarksShift(beforeMark.Line, beforeMark.Col, MaxStrLenP+1-beforeMark.Col, newLine, newCol) {
			goto cleanup
		}
	} else {
		// Move front up and adjust rest
		equalsCol = beforeMark.Col
		equalsLine = newLine
		if beforeMark.Col <= beforeMark.Line.Used {
			shift = beforeMark.Col - 1
		} else {
			beforeMark.Col = beforeMark.Line.Used
		}
		if shift > 0 {
			if !LineChangeLength(newLine, shift) {
				goto cleanup
			}
			newLine.Str.Copy(beforeMark.Line.Str, 1, shift, 1)
			newLine.Used = newLine.Str.Length(' ', shift)
		}
		if !LinesInject(newLine, newLine, beforeMark.Line) {
			goto cleanup
		}
		discard = false
		if beforeMark.Col > 1 {
			if !MarksShift(beforeMark.Line, 1, beforeMark.Col-1, newLine, 1) {
				goto cleanup
			}
		}
		shift = newCol - beforeMark.Col
		if shift <= 0 {
			if shift < 0 {
				beforeMark.Col += shift
				if !textIntraRemove(beforeMark, -shift) {
					goto cleanup
				}
			}
			if newCol > 1 {
				saveCol = beforeMark.Col
				beforeMark.Col = 1
				if !TextOvertype(true, 1, BlankString, newCol-1, beforeMark) {
					goto cleanup
				}
				beforeMark.Col = saveCol
			}
		} else {
			if !TextInsert(true, 1, BlankString, shift, beforeMark) {
				goto cleanup
			}
			if newCol > 1 {
				saveCol = beforeMark.Col
				beforeMark.Col = 1
				if !TextOvertype(true, 1, BlankString, newCol-1, beforeMark) {
					goto cleanup
				}
				beforeMark.Col = saveCol
			}
		}
	}
	if !MarkCreate(
		beforeMark.Line,
		beforeMark.Col,
		&beforeMark.Line.Group.Frame.Marks[MarkModified-MinMarkNumber],
	) {
		goto cleanup
	}
	beforeMark.Line.Group.Frame.TextModified = true
	if !MarkCreate(equalsLine, equalsCol, equalsMark) {
		goto cleanup
	}
	result = true
cleanup:
	if discard {
		LinesDestroy(&newLine, &newLine)
	}
	return result
}
