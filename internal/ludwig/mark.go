/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         MARK
//
// Description:  Mark manipulation routines.

package ludwig

// removeFromMarks removes a mark from a mark list.
func removeFromMarks(markList *[]*MarkObject, mark *MarkObject) {
	if markList == nil || len(*markList) == 0 {
		return
	}

	for i, m := range *markList {
		if m == mark {
			*markList = append((*markList)[:i], (*markList)[i+1:]...)
			return
		}
	}
}

// MarkCreate creates or moves a mark to the specified line and column.
func MarkCreate(inLine *LineHdrObject, column int, mark **MarkObject) bool {
	if *mark == nil {
		*mark = &MarkObject{
			Line: inLine,
			Col:  column,
		}
		inLine.Marks = append([]*MarkObject{*mark}, inLine.Marks...)
		return true
	}

	currentLine := (*mark).Line
	if currentLine == inLine {
		// Mark is already on this line, just update column
		(*mark).Col = column
	} else {
		// Move mark from current line to new line
		removeFromMarks(&currentLine.Marks, *mark)
		inLine.Marks = append([]*MarkObject{*mark}, inLine.Marks...)
		(*mark).Line = inLine
		(*mark).Col = column
	}
	return true
}

// MarkDestroy removes a mark from its line and sets the pointer to nil.
func MarkDestroy(mark **MarkObject) bool {
	removeFromMarks(&(*mark).Line.Marks, *mark)
	*mark = nil
	return true
}

// MarksSqueeze moves all marks between first and last positions to the last position.
func MarksSqueeze(
	firstLine *LineHdrObject,
	firstColumn int,
	lastLine *LineHdrObject,
	lastColumn int,
) bool {
	if firstLine == lastLine {
		// Simple case: all marks are on the same line
		for _, mark := range lastLine.Marks {
			if mark.Col >= firstColumn && mark.Col < lastColumn {
				mark.Col = lastColumn
			}
		}
		return true
	}

	// Multi-line case
	{
		// Move marks in last line that are before the last column
		for _, mark := range lastLine.Marks {
			if mark.Col < lastColumn {
				mark.Col = lastColumn
			}
		}

		// Move marks from intermediate lines (first_line..last_line-1)
		currentLine := firstLine
		for currentLine != lastLine {
			marks := currentLine.Marks
			i := 0
			for i < len(marks) {
				mark := marks[i]
				if mark.Col >= firstColumn {
					// Move this mark to the last line
					mark.Col = lastColumn
					mark.Line = lastLine
					lastLine.Marks = append([]*MarkObject{mark}, lastLine.Marks...)
					// Remove from current list
					marks = append(marks[:i], marks[i+1:]...)
				} else {
					i++
				}
			}
			currentLine.Marks = marks
			currentLine = currentLine.FLink
			firstColumn = 1
		}
	}
	return true
}

// MarksShift moves marks from source range to destination range.
func MarksShift(
	sourceLine *LineHdrObject,
	sourceColumn int,
	width int,
	destLine *LineHdrObject,
	destColumn int,
) bool {
	sourceEnd := sourceColumn + width - 1
	offset := destColumn - sourceColumn

	if sourceLine == destLine {
		// Same line: just adjust column positions
		for _, mark := range sourceLine.Marks {
			if mark.Col >= sourceColumn && mark.Col <= sourceEnd {
				mark.Col = min(mark.Col+offset, MaxStrLenP)
			}
		}
		return true
	}

	// Different lines: move marks to destination line
	marks := sourceLine.Marks
	i := 0
	for i < len(marks) {
		mark := marks[i]
		if mark.Col >= sourceColumn && mark.Col <= sourceEnd {
			// Move mark to destination line with offset
			mark.Line = destLine
			mark.Col = min(mark.Col+offset, MaxStrLenP)
			destLine.Marks = append([]*MarkObject{mark}, destLine.Marks...)
			// Remove from source list (don't increment i)
			marks = append(marks[:i], marks[i+1:]...)
		} else {
			i++
		}
	}
	sourceLine.Marks = marks

	return true
}
