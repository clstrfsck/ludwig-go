/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         SCREEN
//
// Description:  Map a range of lines onto the screen, or unmap the
//               screen.
//               Maintain that mapping.
//               Also SCREEN supports the HARDCOPY/BATCH mode of editing,
//               by providing methods of outputting lines and error
//               messages under these circumstances as well.

package ludwig

import (
	"fmt"
	"strings"
)

// Constants
const (
	PAUSE_MSG   = "Pausing until RETURN pressed: "
	YNAQM_MSG   = "Reply Y(es),N(o),A(lways),Q(uit),M(ore)"
	YNAQM_CHARS = " YNAQM123456789"
)

// Slide types
type slideType int

const (
	slideDont slideType = iota
	slideLeft
	slideRight
	slideRedraw
)

// Scroll types
type scrollType int

const (
	scrollDont scrollType = iota
	scrollForward
	scrollBack
	scrollRedraw
)

func spc(count int) string {
	return strings.Repeat(" ", count)
}

// ScreenMessage puts a message out to the user
func ScreenMessage(message string) {
	if Hangup {
		return
	}

	if LudwigMode == LudwigScreen {
		for i := 0; i < len(message); {
			ScreenFreeBottomLine()
			VduMoveCurs(1, TerminalInfo.Height)
			j := len(message) - i
			if j > TerminalInfo.Width-1 {
				j = TerminalInfo.Width - 1
			}
			VduBold()
			VduDisplayStr(message[i:i+j], 3)
			VduNormal()
			i += j
		}
	} else {
		fmt.Println(message)
	}
}

// ScreenDrawLine draws a line if it is on the screen
func ScreenDrawLine(line *LineHdrObject) {
	VduMoveCurs(1, line.ScrRowNr)
	offset := ScrFrame.ScrOffset
	var strlen int

	eopLine := false
	if line.FLink != nil {
		strlen = line.Used - offset
	} else {
		strlen = line.Len()
		offset = 0
		eopLine = true
	}

	if strlen <= 0 {
		VduClearEOL()
	} else {
		if strlen > ScrFrame.ScrWidth {
			strlen = ScrFrame.ScrWidth
		}
		if eopLine {
			VduDim()
		}
		VduDisplayStr(line.Str.Slice(offset+1, strlen), 3)
		if eopLine {
			VduNormal()
		}
	}

	if line.ScrRowNr == ScrMsgRow {
		ScrMsgRow++
	}
}

// ScreenRedraw redraws the screen exactly as is
func ScreenRedraw() {
	if ScrFrame != nil {
		VduClearScr()
		ScrMsgRow = TerminalInfo.Height + 1
		ScrNeedsFix = false
		line := ScrTopLine
		for line != ScrBotLine {
			ScreenDrawLine(line)
			line = line.FLink
		}
		ScreenDrawLine(line)
	}
}

// screenSlideLine slides a line according to slide_dist and slide_state
func screenSlideLine(line *LineHdrObject, slideDist int, slideState slideType) {
	if line.FLink == nil {
		return
	}

	offset := ScrFrame.ScrOffset
	width := ScrFrame.ScrWidth

	VduMoveCurs(1, line.ScrRowNr)

	if slideState == slideLeft {
		overlap := line.Used - offset
		if overlap > 0 {
			if overlap > slideDist {
				VduInsertChars(slideDist)
				overlap = slideDist
			}
			VduDisplayStr(line.Str.Slice(offset+1, overlap), 2)
		}
	} else {
		if offset-slideDist < line.Used {
			overlap := line.Used - (offset - slideDist + width)
			if slideDist >= width {
				VduClearEOL()
				slideDist = width
			} else {
				VduDeleteChars(slideDist)
				VduMoveCurs(width+1-slideDist, line.ScrRowNr)
			}
			if overlap > 0 {
				if overlap > slideDist {
					overlap = slideDist
				}
				VduDisplayStr(line.Str.Slice(offset+width+1-slideDist+1, overlap), 2)
			}
		}
	}
}

// ScreenSlide slides the whole screen the specified distance
func ScreenSlide(dist int) {
	if ScrFrame != nil {
		if dist != 0 {
			ScrFrame.ScrOffset += dist
			var s slideType
			if dist < 0 {
				s = slideLeft
				dist = -dist
			} else {
				s = slideRight
			}
			l := ScrTopLine
			for l != nil {
				screenSlideLine(l, dist, s)
				if l == ScrBotLine {
					l = nil
				} else {
					l = l.FLink
				}
			}
		}
	}
}

// ScreenUnload unloads the screen
func ScreenUnload() {
	if ScrFrame != nil {
		if ScrFrame.Dot.Line.ScrRowNr == 0 {
			ScrFrame.ScrDotLine =
				(ScrFrame.MarginTop+ScrFrame.ScrHeight-ScrFrame.MarginBottom+1)/2 +
					(TerminalInfo.Height-ScrFrame.ScrHeight)/2
		} else {
			ScrFrame.ScrDotLine = ScrFrame.Dot.Line.ScrRowNr
		}
		VduClearScr()
		ScrMsgRow = TerminalInfo.Height + 1
		ScrNeedsFix = false
		ScrTopLine.ScrRowNr = 0
		for ScrTopLine != ScrBotLine {
			ScrTopLine = ScrTopLine.FLink
			ScrTopLine.ScrRowNr = 0
		}
		ScrFrame = nil
		ScrBotLine = nil
		ScrTopLine = nil
	}
}

// ScreenScroll scrolls the screen forward or back
func ScreenScroll(count int, expand bool) {
	if ScrFrame == nil {
		return
	}

	botLine := ScrBotLine
	topLine := ScrTopLine

	if count >= 0 {
		// FORWARD DIRECTION
		var botLineNr int
		if expand {
			LineToNumber(botLine, &botLineNr)
			eopLineNr := ScrFrame.LastGroup.FirstLineNr + ScrFrame.LastGroup.LastLine.OffsetNr
			remainingLines := eopLineNr - botLineNr
			if remainingLines < count {
				count = remainingLines
			}
			botLineRow := botLine.ScrRowNr
			freeLines := TerminalInfo.Height - botLineRow
			if freeLines > count {
				freeLines = count
			}
			if count-freeLines <= botLineRow {
				for rowNr := botLineRow + 1; rowNr <= botLineRow+freeLines; rowNr++ {
					botLine = botLine.FLink
					botLine.ScrRowNr = rowNr
					ScreenDrawLine(botLine)
				}
				ScrBotLine = botLine
				count -= freeLines
				if count == 0 {
					return
				}
			}
		}

		if count > botLine.ScrRowNr {
			// Would have to scroll too far, redraw instead
			var frame *FrameObject
			if expand {
				frame = ScrFrame
				botLineNr += count
				LineFromNumber(ScrFrame, botLineNr, &botLine)
			}
			ScreenUnload()
			if expand {
				ScrFrame = frame
				ScrTopLine = botLine
				ScrBotLine = botLine
				botLine.ScrRowNr = TerminalInfo.Height
				ScreenDrawLine(botLine)
				screenExpand(true, false)
			}
			return
		}

		// SCROLL 'COUNT' LINES ONTO THE SCREEN
		for count > 0 {
			count--
			VduScrollUp(1)
			if ScrMsgRow <= TerminalInfo.Height {
				ScrMsgRow--
			}

			if expand {
				botLine.FLink.ScrRowNr = botLine.ScrRowNr
				ScreenDrawLine(botLine.FLink)
				botLine = botLine.FLink
			} else {
				botLine.ScrRowNr--
			}

			topLine.ScrRowNr--
			if topLine.ScrRowNr == 0 {
				topLine.FLink.ScrRowNr = 1
				topLine = topLine.FLink
			}
		}
	} else {
		// BACKWARD DIRECTION
		count = -count
		var topLineNr int
		if expand {
			LineToNumber(topLine, &topLineNr)
			remainingLines := topLineNr - 1
			if remainingLines < count {
				count = remainingLines
			}
			topLineRow := topLine.ScrRowNr
			freeLines := topLineRow - 1
			if freeLines >= count {
				freeLines = count
			}

			if topLineRow+count-freeLines <= TerminalInfo.Height+1 {
				for rowNr := topLineRow - 1; rowNr >= topLineRow-freeLines; rowNr-- {
					topLine = topLine.BLink
					topLine.ScrRowNr = rowNr
					ScreenDrawLine(topLine)
				}
				ScrTopLine = topLine
				count -= freeLines
				if count == 0 {
					return
				}
			}
		}

		if count+topLine.ScrRowNr > TerminalInfo.Height+1 {
			// REDRAW
			var frame *FrameObject
			var tmpTopLine *LineHdrObject
			if expand {
				frame = ScrFrame
				topLineNr -= count
				LineFromNumber(ScrFrame, topLineNr, &tmpTopLine)
			}
			ScreenUnload()
			if expand {
				ScrFrame = frame
				ScrTopLine = tmpTopLine
				ScrBotLine = tmpTopLine
				tmpTopLine.ScrRowNr = 1 + TerminalInfo.Height - ScrFrame.ScrHeight
				ScreenDrawLine(tmpTopLine)
				screenExpand(false, true)
			}
			return
		}

		// SCROLL 'COUNT' LINES ONTO THE SCREEN
		for count > 0 {
			count--
			VduMoveCurs(1, 1)
			VduInsertLines(1)
			if ScrMsgRow <= TerminalInfo.Height {
				ScrMsgRow++
			}

			if expand {
				topLine.BLink.ScrRowNr = topLine.ScrRowNr
				ScreenDrawLine(topLine.BLink)
				topLine = topLine.BLink
			} else {
				topLine.ScrRowNr--
			}

			if botLine.ScrRowNr == TerminalInfo.Height {
				botLine.ScrRowNr = 0
				botLine.BLink.ScrRowNr = TerminalInfo.Height
				botLine = botLine.BLink
			} else {
				botLine.ScrRowNr++
			}
		}
	}

	// NOW RESET THE DAMAGED SCREEN POINTERS AND LINE NUMBERS
	ScrTopLine = topLine
	ScrBotLine = botLine

	rowNr := topLine.ScrRowNr
	for topLine != botLine {
		topLine.ScrRowNr = rowNr
		topLine = topLine.FLink
		rowNr++
	}
}

// screenExpand expands a screen out to at least the frame's specified screen height
func screenExpand(initUpwards bool, initDownwards bool) {
	upwards := initUpwards
	downwards := initDownwards

	height := ScrFrame.ScrHeight
	botLine := ScrBotLine
	topLine := ScrTopLine

	linesOnScr := botLine.ScrRowNr + 1 - topLine.ScrRowNr

	for linesOnScr < height && (upwards || downwards) {
		if downwards {
			downwards = false
			curRow := botLine.ScrRowNr
			if botLine.FLink != nil {
				if curRow < TerminalInfo.Height {
					downwards = true
					linesOnScr++
					botLine = botLine.FLink
					botLine.ScrRowNr = curRow + 1
					ScreenDrawLine(botLine)
				}
			}
		}

		if upwards {
			upwards = false
			curRow := topLine.ScrRowNr
			if curRow > 1 {
				if topLine.BLink != nil {
					upwards = true
					linesOnScr++
					topLine = topLine.BLink
					topLine.ScrRowNr = curRow - 1
					ScreenDrawLine(topLine)
				}
			}
		}
	}

	ScrBotLine = botLine
	ScrTopLine = topLine

	// If just expanding wasn't enough then try scrolling to get the lines.
	if linesOnScr < height {
		if initDownwards {
			if botLine.FLink != nil {
				ScreenScroll(height-linesOnScr, true)
				linesOnScr = ScrBotLine.ScrRowNr + 1 - ScrTopLine.ScrRowNr
			}
		}
		if initUpwards && linesOnScr < height {
			var nrLines int
			if LineToNumber(ScrTopLine, &nrLines) {
				if nrLines >= height-linesOnScr {
					nrLines = height - linesOnScr
				}
				ScreenScroll(-nrLines, true)
			}
		}
	}

	// Redraw the <TOP> and <BOTTOM> markers.
	if ScrBotLine.FLink != nil {
		curRow := ScrBotLine.ScrRowNr
		if curRow < TerminalInfo.Height {
			curRow += 1
			VduMoveCurs(1, curRow)
			VduBold()
			VduDisplayStr("<BOTTOM>", 3)
			VduNormal()
			if curRow == ScrMsgRow {
				ScrMsgRow += 1
			}
		}
	}

	if ScrTopLine.ScrRowNr > 1 {
		VduMoveCurs(1, ScrTopLine.ScrRowNr-1)
		VduBold()
		VduDisplayStr("<TOP>", 3)
		VduNormal()
	}
}

// ScreenLinesExtract extracts lines from the screen
func ScreenLinesExtract(firstLine *LineHdrObject, lastLine *LineHdrObject) {
	if lastLine != ScrBotLine {
		// EXTRACTION NOT AT BOT-OF-SCR ACCOMPLISHED VIA TERMINAL H/W
		VduMoveCurs(1, firstLine.ScrRowNr)
		count := lastLine.ScrRowNr + 1 - firstLine.ScrRowNr
		VduDeleteLines(count)
		if ScrMsgRow <= TerminalInfo.Height {
			ScrMsgRow -= count
		}

		lineLimit := lastLine.FLink
		if firstLine == ScrTopLine {
			ScrTopLine = lineLimit
		}
		count = lineLimit.ScrRowNr - firstLine.ScrRowNr
		for {
			firstLine.ScrRowNr = 0
			firstLine = firstLine.FLink
			if firstLine == lineLimit {
				break
			}
		}
		lineLimit = ScrBotLine.FLink
		for {
			firstLine.ScrRowNr -= count
			firstLine = firstLine.FLink
			if firstLine == lineLimit {
				break
			}
		}
		return
	}

	if firstLine == ScrTopLine {
		ScreenUnload()
	} else {
		lineLimit := firstLine.BLink
		for {
			ScrBotLine.ScrRowNr = 0
			ScrBotLine = ScrBotLine.BLink
			if ScrBotLine == lineLimit {
				break
			}
		}
		VduMoveCurs(1, ScrBotLine.ScrRowNr+1)
		VduClearEOS()
		ScrMsgRow = TerminalInfo.Height + 1
	}
}

// ScreenLinesInject injects lines into the screen
func ScreenLinesInject(firstLine *LineHdrObject, count int, beforeLine *LineHdrObject) {
	// HEURISTIC -- KEEP AS MANY LINES ON THE SCREEN AS POSSIBLE
	freeSpaceBelow := TerminalInfo.Height - ScrBotLine.ScrRowNr
	freeSpaceAbove := ScrTopLine.ScrRowNr - 1

	if freeSpaceAbove > 0 && beforeLine != ScrTopLine &&
		TerminalInfo.Height > beforeLine.ScrRowNr-freeSpaceAbove+count {
		scrollupCount := count - freeSpaceBelow
		if scrollupCount > 0 {
			if scrollupCount > freeSpaceAbove {
				scrollupCount = freeSpaceAbove
			}
			VduScrollUp(scrollupCount)
			if ScrMsgRow <= TerminalInfo.Height {
				ScrMsgRow = ScrMsgRow - scrollupCount
			}

			line := ScrTopLine
			rowNr := line.ScrRowNr - scrollupCount
			for {
				line.ScrRowNr = rowNr
				rowNr++
				line = line.FLink
				if line == firstLine {
					break
				}
			}

			ScrBotLine.ScrRowNr -= scrollupCount
			beforeLine.ScrRowNr = rowNr
		}
	}

	// The screen is now optimally placed for the insertion
	if beforeLine == ScrTopLine && freeSpaceAbove > 0 && count+1 <= TerminalInfo.Height {
		rowNr := ScrTopLine.ScrRowNr - 1
		for rowNr > 1 && count > 0 {
			ScrTopLine = ScrTopLine.BLink
			ScrTopLine.ScrRowNr = rowNr
			ScreenDrawLine(ScrTopLine)
			rowNr--
			count--
		}
		beforeLine = ScrTopLine
	}

	// Finally do the insert if necessary
	if count > 0 {
		if beforeLine == ScrTopLine {
			ScrTopLine = firstLine
		}
		rowNr := beforeLine.ScrRowNr
		VduMoveCurs(1, rowNr)
		VduInsertLines(count)
		if ScrMsgRow <= TerminalInfo.Height {
			ScrMsgRow = ScrMsgRow + count
			if ScrMsgRow > TerminalInfo.Height {
				ScrMsgRow = TerminalInfo.Height + 1
			}
		}

		// Patch up the pointers and scr_row_nr's of lines pushed off screen
		line := ScrBotLine
		for i := line.ScrRowNr + count; i >= TerminalInfo.Height+1; i-- {
			if line.ScrRowNr == 0 {
				break
			}
			line.ScrRowNr = 0
			line = line.BLink
		}

		if line.ScrRowNr != 0 {
			// Lines were pushed but left on the screen
			ScrBotLine = line
			rowNr = line.ScrRowNr + count
			for {
				if line.ScrRowNr == 0 {
					line.ScrRowNr = rowNr
					ScreenDrawLine(line)
				} else {
					line.ScrRowNr = rowNr
				}
				rowNr--
				line = line.BLink
				if line == firstLine {
					break
				}
			}
			line.ScrRowNr = rowNr
			ScreenDrawLine(line)
		} else {
			// No lines were left on the screen
			ScrBotLine = firstLine
			firstLine.ScrRowNr = rowNr
			ScreenDrawLine(firstLine)
			screenExpand(false, true)
		}
	}
}

// ScreenLoad loads a screen centered on the given line
func ScreenLoad(line *LineHdrObject) {
	frame := line.Group.Frame

	switch LudwigMode {
	case LudwigBatch:
		// Do nothing

	case LudwigHardcopy:
		newRow := frame.ScrHeight / 2
		for newRow > 0 && line.BLink != nil {
			line = line.BLink
			newRow--
		}
		dotLine := frame.Dot.Line
		dotCol := frame.Dot.Col
		newRow = 1
		for newRow <= frame.ScrHeight && line != nil {
			if newRow == 1 {
				fmt.Println("WINDOW:")
			}
			buflen := line.Used
			if line.FLink == nil {
				buflen = line.Len()
			}
			if buflen > 0 && line.Str != nil {
				fmt.Println(line.Str.Slice(1, buflen))
			} else {
				fmt.Println("")
			}
			if line == dotLine {
				switch dotCol {
				case 1:
					fmt.Println("<")
				case MaxStrLenP:
					fmt.Print(strings.Repeat(" ", MaxStrLen-1))
					fmt.Println(">")
				default:
					fmt.Print(strings.Repeat(" ", dotCol-2))
					fmt.Println("><")
				}
			}
			newRow++
			line = line.FLink
		}

	case LudwigScreen:
		if ScrFrame != nil {
			ScreenUnload()
		} else {
			VduClearScr()
			ScrMsgRow = TerminalInfo.Height + 1
			ScrNeedsFix = false
		}

		newRow := frame.ScrDotLine
		var lineNr int
		LineToNumber(line, &lineNr)
		eopLineNr := frame.LastGroup.FirstLineNr + frame.LastGroup.NrLines - 1
		if (eopLineNr - lineNr) < (TerminalInfo.Height - newRow) {
			newRow = TerminalInfo.Height - (eopLineNr - lineNr)
		}
		if lineNr < newRow {
			newRow = lineNr
		}

		line.ScrRowNr = newRow

		// Move left or right in 1/2 window chunks until DOT on screen
		dotCol := frame.Dot.Col
		for dotCol <= frame.ScrOffset || dotCol > frame.ScrOffset+frame.ScrWidth {
			halfWidth := frame.ScrWidth / 2
			if halfWidth == 0 {
				halfWidth = 1
			}
			if dotCol <= frame.ScrOffset {
				if frame.ScrOffset > halfWidth {
					frame.ScrOffset -= halfWidth
				} else {
					frame.ScrOffset = 0
				}
			} else if frame.ScrOffset+halfWidth+frame.ScrWidth < MaxStrLenP {
				frame.ScrOffset += halfWidth
			} else {
				frame.ScrOffset = MaxStrLenP - frame.ScrWidth
			}
		}

		// Load the screen
		ScrFrame = frame
		ScrBotLine = line
		ScrTopLine = line
		ScreenDrawLine(line)
		screenExpand(true, true)
	}
}

// ScreenPosition positions the screen and cursor
func ScreenPosition(newLine *LineHdrObject, newCol int) {
	if newLine.Group.Frame != ScrFrame {
		ScreenLoad(newLine)
		return
	}

	offset := ScrFrame.ScrOffset
	width := ScrFrame.ScrWidth
	topMargin := ScrFrame.MarginTop
	botMargin := ScrFrame.MarginBottom

	// Check if position is already on screen between margins
	if newLine.ScrRowNr == 0 ||
		(newLine.ScrRowNr-ScrTopLine.ScrRowNr < topMargin && ScrTopLine.BLink != nil) ||
		(ScrBotLine.ScrRowNr-newLine.ScrRowNr < botMargin && ScrBotLine.FLink != nil) ||
		newCol <= offset || newCol > offset+width {

		height := ScrFrame.ScrHeight
		botLine := ScrBotLine
		topLine := ScrTopLine

		// Compute horizontal adjusting needed
		slideState := slideDont
		slideDist := offset + 1 - newCol
		if slideDist > 0 {
			slideState = slideRedraw
			if offset < width/4 {
				slideState = slideLeft
				slideDist = offset
			}
		} else {
			slideDist = newCol - (offset + width)
			if slideDist > 0 {
				slideState = slideRedraw
				if offset > MaxStrLenP-width/4 {
					slideState = slideRight
					slideDist = MaxStrLenP - (offset + width)
				}
			}
		}

		// Compute vertical adjusting needed
		scrollState := scrollDont
		var scrollDist int
		if slideState != slideRedraw &&
			(newLine.ScrRowNr == 0 ||
				(newLine.ScrRowNr-ScrTopLine.ScrRowNr < topMargin && ScrTopLine.BLink != nil) ||
				(ScrBotLine.ScrRowNr-newLine.ScrRowNr < botMargin && ScrBotLine.FLink != nil)) {

			var botLineNr, newLineNr, topLineNr int
			LineToNumber(botLine, &botLineNr)
			LineToNumber(newLine, &newLineNr)
			LineToNumber(topLine, &topLineNr)

			scrollState = scrollRedraw
			if newLineNr < topLineNr ||
				(newLineNr < topLineNr+topMargin && newLineNr < botLineNr) {
				scrollState = scrollBack
				scrollDist = topLineNr + topMargin - newLineNr
				if scrollDist >= topLineNr {
					scrollDist = topLineNr - 1
				}
			} else {
				scrollState = scrollForward
				scrollDist = newLineNr - (botLineNr - botMargin)
				if scrollDist <= 0 {
					scrollState = scrollDont
				}
			}
			if scrollState != scrollRedraw && scrollDist > height {
				scrollState = scrollRedraw
			}
		}

		// Execute the scroll and slide operations
		if scrollState == scrollRedraw || slideState == slideRedraw {
			ScreenLoad(newLine)
		} else {
			if slideState != slideDont {
				// Adjust the screen offset
				if slideState == slideLeft {
					ScrFrame.ScrOffset -= slideDist
				} else {
					ScrFrame.ScrOffset += slideDist
				}

				line := topLine
				switch scrollState {
				case scrollDont, scrollForward:
					if scrollState == scrollDont {
						scrollDist = 0
					}
					// Predict which lines will be left on screen
					nrRows := TerminalInfo.Height - botLine.ScrRowNr
					var rowNr int
					if nrRows >= scrollDist {
						rowNr = 0
					} else {
						rowNr = scrollDist - nrRows
					}
					// Adjust lines that will be left on screen
					for line != nil {
						if line.ScrRowNr > rowNr {
							screenSlideLine(line, slideDist, slideState)
						}
						if line != botLine {
							line = line.FLink
						} else {
							line = nil
						}
					}

				case scrollBack:
					// Decide which lines will be left on screen
					nrRows := topLine.ScrRowNr - 1
					var rowNr int
					if nrRows < scrollDist {
						rowNr = TerminalInfo.Height
					} else {
						rowNr = TerminalInfo.Height - (scrollDist - nrRows)
					}
					// Adjust lines that will be left on screen
					for line != nil {
						if line.ScrRowNr <= rowNr {
							screenSlideLine(line, slideDist, slideState)
						}
						if line != topLine {
							line = line.BLink
						} else {
							line = nil
						}
					}
				}
			}

			switch scrollState {
			case scrollForward:
				ScreenScroll(scrollDist, true)
			case scrollBack:
				ScreenScroll(-scrollDist, true)
			}
		}
	}
	screenExpand(true, true)
}

// ScreenPause waits until user types RETURN
func ScreenPause() {
	if LudwigMode == LudwigScreen {
		if ScrFrame != nil {
			VduMoveCurs(1, 1)
		} else {
			VduDisplayCrLf()
		}
		var buffer *StrObject
		var outlen int
		VduGetInput(PAUSE_MSG, &buffer, MaxStrLen, &outlen)
		if ScrTopLine != nil {
			if ScrTopLine.ScrRowNr == 1 {
				ScreenDrawLine(ScrTopLine)
			} else {
				VduMoveCurs(1, 1)
				VduClearEOL()
				if ScrTopLine.ScrRowNr == 2 {
					ScrNeedsFix = true
				}
			}
		}
	}
}

// ScreenClearMsgs clears any messages off the screen
func ScreenClearMsgs(pause bool) {
	if ScrMsgRow <= TerminalInfo.Height {
		if pause {
			ScreenPause()
		}
		if ScrFrame == nil {
			VduClearScr()
		} else {
			VduMoveCurs(1, ScrMsgRow)
			VduClearEOS()
		}
		ScrMsgRow = TerminalInfo.Height + 1
	}
}

// changeFrameSize changes frame size based on terminal
func changeFrameSize(frm *FrameObject, band int, halfScreen int) {
	if frm.ScrHeight == InitialScrHeight || frm.ScrHeight > TerminalInfo.Height {
		frm.ScrHeight = TerminalInfo.Height
	}
	if frm.ScrWidth == InitialScrWidth || frm.ScrWidth > TerminalInfo.Width {
		frm.ScrWidth = TerminalInfo.Width
	}
	if frm.MarginTop == InitialMarginTop || frm.MarginTop >= halfScreen {
		frm.MarginTop = band
	}
	if frm.MarginBottom == InitialMarginBottom || frm.MarginBottom >= halfScreen {
		frm.MarginBottom = band
	}
	if frm.MarginLeft > TerminalInfo.Width {
		frm.MarginLeft = 1
	}
	if frm.MarginRight == InitialMarginRight || frm.MarginRight > TerminalInfo.Width {
		frm.MarginRight = TerminalInfo.Width
	}
}

// ScreenResize handles screen resize
func ScreenResize() {
	TtWinChanged = false
	VduGetNewDimensions(&TerminalInfo.Width, &TerminalInfo.Height)
	ScrMsgRow = TerminalInfo.Height + 1
	VduClearScr()

	band := TerminalInfo.Height / 6
	halfScreen := TerminalInfo.Height / 2
	nextSpan := FirstSpan
	for nextSpan != nil {
		nextFrame := nextSpan.Frame
		if nextFrame != nil {
			changeFrameSize(nextFrame, band, halfScreen)
		}
		nextSpan = nextSpan.FLink
	}

	InitialMarginRight = TerminalInfo.Width
	InitialMarginBottom = band
	InitialMarginTop = band
	InitialScrWidth = TerminalInfo.Width
	InitialScrHeight = TerminalInfo.Height

	ScreenLoad(CurrentFrame.Dot.Line)
	ScrNeedsFix = false
	screenExpand(true, true)
	VduMoveCurs(
		CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
		CurrentFrame.Dot.Line.ScrRowNr,
	)
}

// ScreenFixup makes sure the screen is correct
func ScreenFixup() {
	if TtWinChanged {
		ScreenResize()
	} else {
		if ScrFrame != CurrentFrame {
			if ScrMsgRow <= TerminalInfo.Height {
				ScreenClearMsgs(true)
			}
			ScreenLoad(CurrentFrame.Dot.Line)
		} else {
			needsReposition := CurrentFrame.Dot.Line.ScrRowNr == 0 ||
				(CurrentFrame.Dot.Line.ScrRowNr-ScrTopLine.ScrRowNr < CurrentFrame.MarginTop &&
					ScrTopLine.BLink != nil) ||
				(ScrBotLine.ScrRowNr-CurrentFrame.Dot.Line.ScrRowNr < CurrentFrame.MarginBottom &&
					ScrBotLine.FLink != nil) ||
				CurrentFrame.Dot.Col <= CurrentFrame.ScrOffset ||
				CurrentFrame.Dot.Col > CurrentFrame.ScrOffset+CurrentFrame.ScrWidth

			if needsReposition {
				if ScrMsgRow <= TerminalInfo.Height {
					ScreenClearMsgs(true)
				}
				ScreenPosition(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col)
			} else if ScrMsgRow <= TerminalInfo.Height {
				VduMoveCurs(
					CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
					CurrentFrame.Dot.Line.ScrRowNr,
				)
				key := VduGetKey()
				ScreenClearMsgs(false)
				if TtControlC {
					return
				}
				VduTakeBackKey(key)
			}
		}
		ScrNeedsFix = false
		screenExpand(true, true)
		VduMoveCurs(
			CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
			CurrentFrame.Dot.Line.ScrRowNr,
		)
	}
}

// ScreenGetLineP gets a line from the user
func ScreenGetLineP(
	prompt string,
	outbuf **StrObject,
	outlen *int,
	maxTp int,
	thisTp int,
) {
	*outlen = 0
	var tmpLine *LineHdrObject
	maxTp = abs(maxTp)

	if !TtControlC {
		if LudwigMode == LudwigScreen {
			PromptRegion[thisTp].LineNr = 0
			PromptRegion[thisTp].Redraw = nil

			if ScrTopLine == nil {
				goto l1
			}
			ScreenFixup()
			PromptRegion[thisTp].LineNr = thisTp
			if ScrTopLine.ScrRowNr > maxTp {
				goto l1
			}
			if ScrBotLine.ScrRowNr < ScrMsgRow-maxTp {
				PromptRegion[thisTp].LineNr = ScrMsgRow - maxTp + thisTp - 1
				goto l1
			}
			tmpLine = ScrTopLine
			for index := ScrTopLine.ScrRowNr; index <= thisTp-1; index++ {
				tmpLine = tmpLine.FLink
			}
			PromptRegion[thisTp].Redraw = tmpLine
			if ScrFrame.Dot.Line.ScrRowNr > 2 {
				goto l1
			}

			tmpLine = ScrBotLine
			for index := TerminalInfo.Height - ScrBotLine.ScrRowNr; index <= maxTp-thisTp-1; index++ {
				tmpLine = tmpLine.BLink
			}
			if TerminalInfo.Height-ScrBotLine.ScrRowNr > maxTp-thisTp {
				tmpLine = nil
			}
			PromptRegion[thisTp].Redraw = tmpLine
			PromptRegion[thisTp].LineNr = TerminalInfo.Height - maxTp + thisTp

		l1:
			if PromptRegion[thisTp].LineNr != 0 {
				VduMoveCurs(1, PromptRegion[thisTp].LineNr)
			}
			VduGetInput(prompt, outbuf, MaxStrLen, outlen)
			if TtControlC {
				goto l2
			}
			if *outlen == 0 {
				for index := thisTp + 1; index <= maxTp; index++ {
					PromptRegion[index].LineNr = 0
					PromptRegion[index].Redraw = nil
				}
			}
			if thisTp == maxTp || *outlen == 0 {
				for index := 1; index <= maxTp; index++ {
					if PromptRegion[index].Redraw != nil {
						ScreenDrawLine(PromptRegion[index].Redraw)
					} else if PromptRegion[index].LineNr != 0 {
						VduMoveCurs(1, PromptRegion[index].LineNr)
						VduClearEOL()
					}
				}
			}
		} else {
			fmt.Print(prompt)
			// Read from stdin (simplified version)
			*outlen = 0
		}
	}
l2:
	if TtControlC {
		*outlen = 0
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ScreenFreeBottomLine frees the bottom line of the screen
func ScreenFreeBottomLine() {
	// This routine assumes that the editor is in SCREEN mode.
	// This routine frees the bottom line of the screen for use by the caller.
	// The main use of the area is the outputting of messages.
	if ScrFrame == nil {
		// IF SCREEN NOT MAPPED.
		VduDisplayCrLf()
		VduDeleteLines(1)
		return
	}
	ScrNeedsFix = true
	// IF BOTTOM LINE FREE.
	if (ScrMsgRow > TerminalInfo.Height) && (ScrBotLine.ScrRowNr < TerminalInfo.Height) {
		// Nothing
	} else if ScrBotLine.ScrRowNr+2 < ScrMsgRow {
		// IF ROOM BELOW BOT LINE.
		// +2 because of <eos> line.
		VduMoveCurs(1, ScrBotLine.ScrRowNr+2)
		VduDeleteLines(1)
	} else if ScrTopLine.ScrRowNr != 1 {
		// IF TOP LINE FREE.
		ScreenScroll(1, false)
	} else {
		if ScrBotLine.ScrRowNr+1 < ScrMsgRow {
			// IF ROOM FOR MORE MSGS.
			VduMoveCurs(1, ScrBotLine.ScrRowNr+1)
			VduDeleteLines(1)
		} else if (ScrFrame.Dot.Line != ScrTopLine) &&
			!((ScrFrame.Dot.Line != ScrBotLine) &&
				(ScrBotLine.ScrRowNr == TerminalInfo.Height)) {
			// IF DOT NOT ON TOP LINE,
			// AND WE CANT USE THE BOT.
			ScreenScroll(1, false)
		} else if ScrMsgRow <= TerminalInfo.Height/2 {
			// 1/2 SCREEN ALREADY MSGS.
			VduMoveCurs(1, ScrMsgRow)
			VduDeleteLines(1)
			return
		} else {
			// CONTRACT SCREEN 1 LINE.
			ScrBotLine.ScrRowNr = 0
			ScrBotLine = ScrBotLine.BLink
			VduMoveCurs(1, ScrMsgRow-1)
			VduDeleteLines(1)
		}
	}
	ScrMsgRow -= 1
}

// ScreenVerify gets verification from user
func ScreenVerify(prompt string) VerifyResponse {
	const verHeight = 4

	verify := VerifyReplyQuit

	oldHeight := CurrentFrame.ScrHeight
	oldTopM := CurrentFrame.MarginTop
	oldBotM := CurrentFrame.MarginBottom
	if oldHeight > verHeight {
		CurrentFrame.MarginTop = verHeight / 2
		CurrentFrame.ScrHeight = verHeight
		CurrentFrame.MarginBottom = verHeight / 2
	}

	usePrompt := true
	var key int
	var more bool

	for {
		switch LudwigMode {
		case LudwigScreen:
			ScreenFixup()
			VduBold()
			if usePrompt {
				ScreenMessage(prompt)
			} else {
				ScreenMessage(YNAQM_MSG)
			}
			VduNormal()
			VduMoveCurs(
				CurrentFrame.Dot.Col-CurrentFrame.ScrOffset,
				CurrentFrame.Dot.Line.ScrRowNr,
			)
			key = VduGetKey()
			if key >= 'a' && key <= 'z' {
				key = key - 'a' + 'A'
			}
			if key == 13 {
				key = 'N' // RETURN <=> NO
			}
			ScreenClearMsgs(false)

		case LudwigBatch, LudwigHardcopy:
			var response *StrObject
			var respLen int
			if usePrompt {
				ScreenGetLineP(prompt, &response, &respLen, 1, 1)
			} else {
				ScreenGetLineP(YNAQM_MSG, &response, &respLen, 1, 1)
			}
			if respLen == 0 {
				key = 'N'
			} else {
				k := response.Get(1)
				if k >= 'a' && k <= 'z' {
					key = int(k - 'a' + 'A')
				} else {
					key = int(k)
				}
			}
		}

		if TtControlC {
			goto l99
		}

		more = false
		if strings.IndexByte(YNAQM_CHARS, byte(key)) != -1 {
			switch key {
			case ' ', 'Y':
				verify = VerifyReplyYes
			case 'N':
				verify = VerifyReplyNo
			case 'A':
				verify = VerifyReplyAlways
			case 'Q':
				verify = VerifyReplyQuit
			case '1', '2', '3', '4', '5', '6', '7', '8', '9', 'M':
				// MORE CONTEXT PLEASE
				if ScrTopLine != nil {
					CurrentFrame.ScrHeight = ScrBotLine.ScrRowNr + 1 - ScrTopLine.ScrRowNr
				}
				if key == 'M' {
					key = '1'
				}
				if key-'0'+CurrentFrame.ScrHeight < TerminalInfo.Height {
					CurrentFrame.ScrHeight += key - '0'
				} else {
					CurrentFrame.ScrHeight = TerminalInfo.Height
				}
				if ScrTopLine == nil {
					ScreenLoad(CurrentFrame.Dot.Line)
				} else {
					screenExpand(true, true)
				}
				more = true
				usePrompt = true
			}
		} else {
			ScreenBeep()
			more = true
			usePrompt = false
		}

		if !more {
			break
		}
	}

l99:
	CurrentFrame.ScrHeight = oldHeight
	CurrentFrame.MarginTop = oldTopM
	CurrentFrame.MarginBottom = oldBotM

	if verify == VerifyReplyQuit {
		ExitAbort = true
	}
	return verify
}

// ScreenBeep produces a beep
func ScreenBeep() {
	if LudwigMode == LudwigScreen {
		VduBeep()
	}
}

// ScreenHome moves cursor to home position
func ScreenHome(clear bool) {
	if LudwigMode == LudwigScreen {
		VduMoveCurs(1, 1)
		if clear {
			VduClearScr()
		}
		VduFlush()
	} else {
		fmt.Println("")
		fmt.Println("")
	}
}

// ScreenWriteInt writes an integer to screen
func ScreenWriteInt(intVal int, width int) {
	if LudwigMode == LudwigScreen {
		str := fmt.Sprintf("%*d", width, intVal)
		VduDisplayStr(str, 0)
	} else {
		fmt.Printf("%d", intVal)
	}
}

// ScreenWriteCh writes a character with indent
func ScreenWriteCh(indent int, ch byte) {
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+string(ch), 0)
	} else {
		fmt.Print(spc(indent) + string(ch))
	}
}

// ScreenWriteStr writes a string with indent
func ScreenWriteStr(indent int, str string) {
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+str, 3)
		VduFlush()
	} else {
		fmt.Print(spc(indent) + str)
	}
}

// ScreenWriteStrWidth writes a string with indent and width
func ScreenWriteStrWidth(indent int, str string, width int) {
	if LudwigMode == LudwigScreen {
		strLen := min(len(str), width)
		trailingSpaces := width - strLen
		VduDisplayStr(spc(indent)+str[:strLen]+spc(trailingSpaces), 3)
	} else {
		fmt.Print(spc(indent) + str)
	}
}

// ScreenWriteNameStr writes a name string with indent and width
func ScreenWriteNameStr(indent int, str string, width int) {
	strLen := min(len(str), width)
	trailingSpaces := width - strLen
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+str[:strLen]+spc(trailingSpaces), 3)
	} else {
		fmt.Print(spc(indent))
		fmt.Print(str[:strLen])
		fmt.Print(spc(trailingSpaces))
	}
}

// ScreenWriteFileNameStr writes a file name with indent and width
func ScreenWriteFileNameStr(indent int, str string, width int) {
	if LudwigMode == LudwigScreen {
		for range indent {
			VduDisplayCh(' ')
		}
		for i := range width {
			if i < len(str) && str[i] >= 32 && str[i] <= 126 {
				VduDisplayCh(str[i])
			} else {
				VduDisplayCh(' ')
			}
		}
	} else {
		fmt.Print(spc(indent))
		for i := 0; i < width; i++ {
			if i < len(str) {
				fmt.Print(string(str[i]))
			} else {
				fmt.Print(" ")
			}
		}
	}
}

// ScreenWriteln writes a newline
func ScreenWriteln() {
	if LudwigMode == LudwigScreen {
		VduDisplayCrLf()
	} else {
		fmt.Println("")
	}
}

// ScreenWritelnClel writes newline and clears to end of line
func ScreenWritelnClel() {
	if LudwigMode == LudwigScreen {
		VduClearEOL()
		VduDisplayCrLf()
	} else {
		fmt.Println("")
	}
}

// ScreenHelpPrompt displays help prompt
func ScreenHelpPrompt(prompt string) string {
	var reply string

	switch LudwigMode {
	case LudwigScreen, LudwigHardcopy:
		if LudwigMode == LudwigScreen {
			VduBold()
		}
		ScreenWriteStr(0, prompt)
		if LudwigMode == LudwigScreen {
			VduNormal()
		}
		terminated := false
		for !terminated {
			key := VduGetKey()
			if key == 13 {
				terminated = true
			} else if key == 127 {
				if len(reply) > 0 {
					reply = reply[:len(reply)-1]
					VduDisplayCh(8)
					VduDisplayCh(' ')
					VduDisplayCh(8)
				}
			} else if ChIsPrintable(rune(key)) {
				VduDisplayCh(byte(key))
				reply += string(byte(key))
				terminated = (key == ' ') || len(reply) == KeyLen
			}
		}
		ScreenWriteln()

	case LudwigBatch:
		reply = ""
	}

	return reply
}
