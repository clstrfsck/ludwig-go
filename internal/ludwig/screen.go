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
	"bufio"
	"fmt"
	"strings"
)

// Constants
const (
	PAUSE_MSG   = "Pausing until RETURN pressed: "
	YNAQM_MSG   = "Reply Y(es),N(o),A(lways),Q(uit),M(ore)"
	YNAQM_CHARS = " YNAQM123456789"
)

// ScreenState holds all screen display state
type ScreenState struct {
	Frame       *FrameObject
	TopLine     *LineHdrObject
	BotLine     *LineHdrObject
	MsgRow      int
	NeedsFix    bool
	StdinReader *bufio.Reader
}

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

// Message puts a message out to the user
func (ss *ScreenState) Message(message string) {
	if Hangup {
		return
	}

	if LudwigMode == LudwigScreen {
		for i := 0; i < len(message); {
			ss.FreeBottomLine()
			VduMoveCurs(1, TerminalInfo.Height)
			j := min(len(message)-i, TerminalInfo.Width-1)
			VduBold()
			VduDisplayStr(message[i:i+j], true)
			VduNormal()
			i += j
		}
	} else {
		fmt.Println(message)
	}
}

// DrawLine draws a line if it is on the screen
func (ss *ScreenState) DrawLine(line *LineHdrObject) {
	VduMoveCurs(1, line.ScrRowNr)
	offset := ss.Frame.ScrOffset
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
		if strlen > ss.Frame.ScrWidth {
			strlen = ss.Frame.ScrWidth
		}
		if eopLine {
			VduDim()
			VduDisplayStr(line.Str.Slice(offset+1, strlen), true)
			VduNormal()
		} else if line.HlMatch != nil {
			syntaxDrawLine(line, offset, strlen)
			VduClearEOL()
		} else {
			VduDisplayStr(line.Str.Slice(offset+1, strlen), true)
		}
	}

	if line.ScrRowNr == ss.MsgRow {
		ss.MsgRow++
	}
}

// Redraw redraws the screen exactly as is
func (ss *ScreenState) Redraw() {
	if ss.Frame != nil {
		VduClearScr()
		ss.MsgRow = TerminalInfo.Height + 1
		ss.NeedsFix = false
		line := ss.TopLine
		for line != ss.BotLine {
			ss.DrawLine(line)
			line = line.FLink
		}
		ss.DrawLine(line)
	}
}

// slideLine slides a line according to slide_dist and slide_state
func (ss *ScreenState) slideLine(line *LineHdrObject, slideDist int, slideState slideType) {
	if line.FLink == nil {
		return
	}

	offset := ss.Frame.ScrOffset
	width := ss.Frame.ScrWidth

	VduMoveCurs(1, line.ScrRowNr)

	if slideState == slideLeft {
		overlap := line.Used - offset
		if overlap > 0 {
			if overlap > slideDist {
				VduInsertChars(slideDist)
				overlap = slideDist
			}
			VduDisplayStr(line.Str.Slice(offset+1, overlap), false)
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
				VduDisplayStr(line.Str.Slice(offset+width+1-slideDist+1, overlap), false)
			}
		}
	}
}

// Slide slides the whole screen the specified distance
func (ss *ScreenState) Slide(dist int) {
	if ss.Frame != nil {
		if dist != 0 {
			ss.Frame.ScrOffset += dist
			var s slideType
			if dist < 0 {
				s = slideLeft
				dist = -dist
			} else {
				s = slideRight
			}
			l := ss.TopLine
			for l != nil {
				ss.slideLine(l, dist, s)
				if l == ss.BotLine {
					l = nil
				} else {
					l = l.FLink
				}
			}
		}
	}
}

// Unload unloads the screen
func (ss *ScreenState) Unload() {
	if ss.Frame != nil {
		if ss.Frame.Dot.Line.ScrRowNr == 0 {
			ss.Frame.ScrDotLine =
				(ss.Frame.MarginTop+ss.Frame.ScrHeight-ss.Frame.MarginBottom+1)/2 +
					(TerminalInfo.Height-ss.Frame.ScrHeight)/2
		} else {
			ss.Frame.ScrDotLine = ss.Frame.Dot.Line.ScrRowNr
		}
		VduClearScr()
		ss.MsgRow = TerminalInfo.Height + 1
		ss.NeedsFix = false
		ss.TopLine.ScrRowNr = 0
		for ss.TopLine != ss.BotLine {
			ss.TopLine = ss.TopLine.FLink
			ss.TopLine.ScrRowNr = 0
		}
		ss.Frame = nil
		ss.BotLine = nil
		ss.TopLine = nil
	}
}

// Scroll scrolls the screen forward or back
func (ss *ScreenState) Scroll(count int, expand bool) {
	if ss.Frame == nil {
		return
	}

	botLine := ss.BotLine
	topLine := ss.TopLine

	if count >= 0 {
		// FORWARD DIRECTION
		var botLineNr int
		if expand {
			botLineNr = LineToNumber(botLine)
			eopLineNr := ss.Frame.LastGroup.FirstLineNr + ss.Frame.LastGroup.LastLine.OffsetNr
			remainingLines := eopLineNr - botLineNr
			if remainingLines < count {
				count = remainingLines
			}
			botLineRow := botLine.ScrRowNr
			freeLines := min(TerminalInfo.Height-botLineRow, count)
			if count-freeLines <= botLineRow {
				for rowNr := botLineRow + 1; rowNr <= botLineRow+freeLines; rowNr++ {
					botLine = botLine.FLink
					botLine.ScrRowNr = rowNr
					ss.DrawLine(botLine)
				}
				ss.BotLine = botLine
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
				frame = ss.Frame
				botLineNr += count
				botLine = LineFromNumber(ss.Frame, botLineNr)
			}
			ss.Unload()
			if expand {
				ss.Frame = frame
				ss.TopLine = botLine
				ss.BotLine = botLine
				botLine.ScrRowNr = TerminalInfo.Height
				ss.DrawLine(botLine)
				ss.expand(true, false)
			}
			return
		}

		// SCROLL 'COUNT' LINES ONTO THE SCREEN
		for count > 0 {
			count--
			VduScrollUp(1)
			if ss.MsgRow <= TerminalInfo.Height {
				ss.MsgRow--
			}

			if expand {
				botLine.FLink.ScrRowNr = botLine.ScrRowNr
				ss.DrawLine(botLine.FLink)
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
			topLineNr = LineToNumber(topLine)
			remainingLines := topLineNr - 1
			if remainingLines < count {
				count = remainingLines
			}
			topLineRow := topLine.ScrRowNr
			freeLines := min(topLineRow-1, count)

			if topLineRow+count-freeLines <= TerminalInfo.Height+1 {
				for rowNr := topLineRow - 1; rowNr >= topLineRow-freeLines; rowNr-- {
					topLine = topLine.BLink
					topLine.ScrRowNr = rowNr
					ss.DrawLine(topLine)
				}
				ss.TopLine = topLine
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
				frame = ss.Frame
				topLineNr -= count
				tmpTopLine = LineFromNumber(ss.Frame, topLineNr)
			}
			ss.Unload()
			if expand {
				ss.Frame = frame
				ss.TopLine = tmpTopLine
				ss.BotLine = tmpTopLine
				tmpTopLine.ScrRowNr = 1 + TerminalInfo.Height - ss.Frame.ScrHeight
				ss.DrawLine(tmpTopLine)
				ss.expand(false, true)
			}
			return
		}

		// SCROLL 'COUNT' LINES ONTO THE SCREEN
		for count > 0 {
			count--
			VduMoveCurs(1, 1)
			VduInsertLines(1)
			if ss.MsgRow <= TerminalInfo.Height {
				ss.MsgRow++
			}

			if expand {
				topLine.BLink.ScrRowNr = topLine.ScrRowNr
				ss.DrawLine(topLine.BLink)
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
	ss.TopLine = topLine
	ss.BotLine = botLine

	rowNr := topLine.ScrRowNr
	for topLine != botLine {
		topLine.ScrRowNr = rowNr
		topLine = topLine.FLink
		rowNr++
	}
}

// expand expands a screen out to at least the frame's specified screen height
func (ss *ScreenState) expand(initUpwards bool, initDownwards bool) {
	upwards := initUpwards
	downwards := initDownwards

	height := ss.Frame.ScrHeight
	botLine := ss.BotLine
	topLine := ss.TopLine

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
					ss.DrawLine(botLine)
				}
			}
		}

		if upwards {
			upwards = false
			curRow := topLine.ScrRowNr
			if curRow > 1 {
				if topLine.BLink != nil {
					upwards = true
					linesOnScr += 1
					topLine = topLine.BLink
					topLine.ScrRowNr = curRow - 1
					ss.DrawLine(topLine)
				}
			}
		}
	}

	ss.BotLine = botLine
	ss.TopLine = topLine

	// If just expanding wasn't enough then try scrolling to get the lines.
	if linesOnScr < height {
		if initDownwards {
			if botLine.FLink != nil {
				ss.Scroll(height-linesOnScr, true)
				linesOnScr = ss.BotLine.ScrRowNr + 1 - ss.TopLine.ScrRowNr
			}
		}
		if initUpwards && linesOnScr < height {
			nrLines := LineToNumber(ss.TopLine)
			if nrLines >= height-linesOnScr {
				nrLines = height - linesOnScr
			}
			ss.Scroll(-nrLines, true)
		}
	}

	// Redraw the <TOP> and <BOTTOM> markers.
	if ss.BotLine.FLink != nil {
		curRow := ss.BotLine.ScrRowNr
		if curRow < TerminalInfo.Height {
			curRow += 1
			VduMoveCurs(1, curRow)
			VduBold()
			VduDisplayStr("<BOTTOM>", true)
			VduNormal()
			if curRow == ss.MsgRow {
				ss.MsgRow += 1
			}
		}
	}

	if ss.TopLine.ScrRowNr > 1 {
		VduMoveCurs(1, ss.TopLine.ScrRowNr-1)
		VduBold()
		VduDisplayStr("<TOP>", true)
		VduNormal()
	}
}

// LinesExtract extracts lines from the screen
func (ss *ScreenState) LinesExtract(firstLine *LineHdrObject, lastLine *LineHdrObject) {
	if lastLine != ss.BotLine {
		// EXTRACTION NOT AT BOT-OF-SCR ACCOMPLISHED VIA TERMINAL H/W
		VduMoveCurs(1, firstLine.ScrRowNr)
		count := lastLine.ScrRowNr + 1 - firstLine.ScrRowNr
		VduDeleteLines(count)
		if ss.MsgRow <= TerminalInfo.Height {
			ss.MsgRow -= count
		}

		lineLimit := lastLine.FLink
		if firstLine == ss.TopLine {
			ss.TopLine = lineLimit
		}
		count = lineLimit.ScrRowNr - firstLine.ScrRowNr
		for {
			firstLine.ScrRowNr = 0
			firstLine = firstLine.FLink
			if firstLine == lineLimit {
				break
			}
		}
		lineLimit = ss.BotLine.FLink
		for {
			firstLine.ScrRowNr -= count
			firstLine = firstLine.FLink
			if firstLine == lineLimit {
				break
			}
		}
		return
	}

	if firstLine == ss.TopLine {
		ss.Unload()
	} else {
		lineLimit := firstLine.BLink
		for {
			ss.BotLine.ScrRowNr = 0
			ss.BotLine = ss.BotLine.BLink
			if ss.BotLine == lineLimit {
				break
			}
		}
		VduMoveCurs(1, ss.BotLine.ScrRowNr+1)
		VduClearEOS()
		ss.MsgRow = TerminalInfo.Height + 1
	}
}

// LinesInject injects lines into the screen
func (ss *ScreenState) LinesInject(firstLine *LineHdrObject, count int, beforeLine *LineHdrObject) {
	// HEURISTIC -- KEEP AS MANY LINES ON THE SCREEN AS POSSIBLE
	freeSpaceBelow := TerminalInfo.Height - ss.BotLine.ScrRowNr
	freeSpaceAbove := ss.TopLine.ScrRowNr - 1

	if freeSpaceAbove > 0 && beforeLine != ss.TopLine &&
		TerminalInfo.Height > beforeLine.ScrRowNr-freeSpaceAbove+count {
		scrollupCount := count - freeSpaceBelow
		if scrollupCount > 0 {
			if scrollupCount > freeSpaceAbove {
				scrollupCount = freeSpaceAbove
			}
			VduScrollUp(scrollupCount)
			if ss.MsgRow <= TerminalInfo.Height {
				ss.MsgRow = ss.MsgRow - scrollupCount
			}

			line := ss.TopLine
			rowNr := line.ScrRowNr - scrollupCount
			for {
				line.ScrRowNr = rowNr
				rowNr++
				line = line.FLink
				if line == firstLine {
					break
				}
			}

			ss.BotLine.ScrRowNr -= scrollupCount
			beforeLine.ScrRowNr = rowNr
		}
	}

	// The screen is now optimally placed for the insertion
	if beforeLine == ss.TopLine && freeSpaceAbove > 0 && count+1 <= TerminalInfo.Height {
		rowNr := ss.TopLine.ScrRowNr - 1
		for rowNr > 1 && count > 0 {
			ss.TopLine = ss.TopLine.BLink
			ss.TopLine.ScrRowNr = rowNr
			ss.DrawLine(ss.TopLine)
			rowNr--
			count--
		}
		beforeLine = ss.TopLine
	}

	// Finally do the insert if necessary
	if count > 0 {
		if beforeLine == ss.TopLine {
			ss.TopLine = firstLine
		}
		rowNr := beforeLine.ScrRowNr
		VduMoveCurs(1, rowNr)
		VduInsertLines(count)
		if ss.MsgRow <= TerminalInfo.Height {
			ss.MsgRow = ss.MsgRow + count
			if ss.MsgRow > TerminalInfo.Height {
				ss.MsgRow = TerminalInfo.Height + 1
			}
		}

		// Patch up the pointers and scr_row_nr's of lines pushed off screen
		line := ss.BotLine
		for i := line.ScrRowNr + count; i >= TerminalInfo.Height+1; i-- {
			if line.ScrRowNr == 0 {
				break
			}
			line.ScrRowNr = 0
			line = line.BLink
		}

		if line.ScrRowNr != 0 {
			// Lines were pushed but left on the screen
			ss.BotLine = line
			rowNr = line.ScrRowNr + count
			for {
				if line.ScrRowNr == 0 {
					line.ScrRowNr = rowNr
					ss.DrawLine(line)
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
			ss.DrawLine(line)
		} else {
			// No lines were left on the screen
			ss.BotLine = firstLine
			firstLine.ScrRowNr = rowNr
			ss.DrawLine(firstLine)
			ss.expand(false, true)
		}
	}
}

// Load loads a screen centered on the given line
func (ss *ScreenState) Load(line *LineHdrObject) {
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
		if ss.Frame != nil {
			ss.Unload()
		} else {
			VduClearScr()
			ss.MsgRow = TerminalInfo.Height + 1
			ss.NeedsFix = false
		}

		newRow := frame.ScrDotLine
		lineNr := LineToNumber(line)
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
		ss.Frame = frame
		ss.BotLine = line
		ss.TopLine = line
		ss.DrawLine(line)
		ss.expand(true, true)
	}
}

// Position positions the screen and cursor
func (ss *ScreenState) Position(newLine *LineHdrObject, newCol int) {
	if newLine.Group.Frame != ss.Frame {
		ss.Load(newLine)
		return
	}

	offset := ss.Frame.ScrOffset
	width := ss.Frame.ScrWidth
	topMargin := ss.Frame.MarginTop
	botMargin := ss.Frame.MarginBottom

	// Check if position is already on screen between margins
	if newLine.ScrRowNr == 0 ||
		(newLine.ScrRowNr-ss.TopLine.ScrRowNr < topMargin && ss.TopLine.BLink != nil) ||
		(ss.BotLine.ScrRowNr-newLine.ScrRowNr < botMargin && ss.BotLine.FLink != nil) ||
		newCol <= offset || newCol > offset+width {

		height := ss.Frame.ScrHeight
		botLine := ss.BotLine
		topLine := ss.TopLine

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
				(newLine.ScrRowNr-ss.TopLine.ScrRowNr < topMargin && ss.TopLine.BLink != nil) ||
				(ss.BotLine.ScrRowNr-newLine.ScrRowNr < botMargin && ss.BotLine.FLink != nil)) {

			botLineNr := LineToNumber(botLine)
			newLineNr := LineToNumber(newLine)
			topLineNr := LineToNumber(topLine)

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
			ss.Load(newLine)
		} else {
			if slideState != slideDont {
				// Adjust the screen offset
				if slideState == slideLeft {
					ss.Frame.ScrOffset -= slideDist
				} else {
					ss.Frame.ScrOffset += slideDist
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
							ss.slideLine(line, slideDist, slideState)
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
							ss.slideLine(line, slideDist, slideState)
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
				ss.Scroll(scrollDist, true)
			case scrollBack:
				ss.Scroll(-scrollDist, true)
			}
		}
	}
	ss.expand(true, true)
}

// Pause waits until user types RETURN
func (ss *ScreenState) Pause() {
	if LudwigMode == LudwigScreen {
		if ss.Frame != nil {
			VduMoveCurs(1, 1)
		} else {
			VduDisplayCrLf()
		}
		VduGetInput(PAUSE_MSG, MaxStrLen)
		if ss.TopLine != nil {
			if ss.TopLine.ScrRowNr == 1 {
				ss.DrawLine(ss.TopLine)
			} else {
				VduMoveCurs(1, 1)
				VduClearEOL()
				if ss.TopLine.ScrRowNr == 2 {
					ss.NeedsFix = true
				}
			}
		}
	}
}

// ClearMsgs clears any messages off the screen
func (ss *ScreenState) ClearMsgs(pause bool) {
	if ss.MsgRow <= TerminalInfo.Height {
		if pause {
			ss.Pause()
		}
		if ss.Frame == nil {
			VduClearScr()
		} else {
			VduMoveCurs(1, ss.MsgRow)
			VduClearEOS()
		}
		ss.MsgRow = TerminalInfo.Height + 1
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

// Resize handles screen resize
func (ss *ScreenState) Resize(frame *FrameObject) {
	TtWinChanged = false
	TerminalInfo.Width, TerminalInfo.Height = VduGetNewDimensions()
	ss.MsgRow = TerminalInfo.Height + 1
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

	ss.Load(frame.Dot.Line)
	ss.NeedsFix = false
	ss.expand(true, true)
	VduMoveCurs(
		frame.Dot.Col-frame.ScrOffset,
		frame.Dot.Line.ScrRowNr,
	)
}

func advance(line *LineHdrObject, count int) *LineHdrObject {
	for count > 0 && line != nil {
		line = line.FLink
		count--
	}
	return line
}

// Fixup makes sure the screen is correct
func (ss *ScreenState) Fixup(frame *FrameObject) {
	if TtWinChanged {
		ss.Resize(frame)
	} else {
		if ss.Frame != frame {
			if ss.MsgRow <= TerminalInfo.Height {
				ss.ClearMsgs(true)
			}
			ss.Load(frame.Dot.Line)
		} else {
			needsReposition := frame.Dot.Line.ScrRowNr == 0 ||
				(frame.Dot.Line.ScrRowNr-ss.TopLine.ScrRowNr < frame.MarginTop &&
					ss.TopLine.BLink != nil) ||
				(TerminalInfo.Height-frame.Dot.Line.ScrRowNr < frame.MarginBottom &&
					advance(ss.BotLine.FLink, TerminalInfo.Height-ss.BotLine.ScrRowNr) != nil) ||
				frame.Dot.Col <= frame.ScrOffset ||
				frame.Dot.Col > frame.ScrOffset+frame.ScrWidth

			if needsReposition {
				if ss.MsgRow <= TerminalInfo.Height {
					ss.ClearMsgs(true)
				}
				ss.Position(frame.Dot.Line, frame.Dot.Col)
			} else if ss.MsgRow <= TerminalInfo.Height {
				VduMoveCurs(
					frame.Dot.Col-frame.ScrOffset,
					frame.Dot.Line.ScrRowNr,
				)
				key := VduGetKey()
				ss.ClearMsgs(false)
				if TtControlC {
					return
				}
				VduTakeBackKey(key)
			}
		}
		ss.NeedsFix = false
		SyntaxApplyDirty(frame)
		ss.expand(true, true)
		VduMoveCurs(
			frame.Dot.Col-frame.ScrOffset,
			frame.Dot.Line.ScrRowNr,
		)
	}
}

// computePromptPosition sets PromptRegion[thisTp].LineNr and .Redraw to
// determine where to display the prompt on screen. When the screen is not
// yet initialised (ss.TopLine == nil) or the content fits entirely above or
// below the prompt band, only LineNr is set. When the prompt must overwrite
// an existing screen line, Redraw records that line so it can be restored
// after input is complete.
func (ss *ScreenState) computePromptPosition(frame *FrameObject, thisTp, maxTp int) {
	PromptRegion[thisTp].LineNr = 0
	PromptRegion[thisTp].Redraw = nil

	if ss.TopLine == nil {
		return
	}

	ss.Fixup(frame)
	PromptRegion[thisTp].LineNr = thisTp

	if ss.TopLine.ScrRowNr > maxTp {
		return
	}

	if ss.BotLine.ScrRowNr < ss.MsgRow-maxTp {
		PromptRegion[thisTp].LineNr = ss.MsgRow - maxTp + thisTp - 1
		return
	}

	// Find the screen line that will be overwritten by this prompt row.
	tmpLine := ss.TopLine
	for index := ss.TopLine.ScrRowNr; index <= thisTp-1; index++ {
		tmpLine = tmpLine.FLink
	}
	PromptRegion[thisTp].Redraw = tmpLine

	if ss.Frame.Dot.Line.ScrRowNr > 2 {
		return
	}

	// Dot is near the top of the screen; reposition the prompt to the bottom
	// to avoid obscuring it.
	tmpLine = ss.BotLine
	for index := TerminalInfo.Height - ss.BotLine.ScrRowNr; index <= maxTp-thisTp-1; index++ {
		tmpLine = tmpLine.BLink
	}
	if TerminalInfo.Height-ss.BotLine.ScrRowNr > maxTp-thisTp {
		tmpLine = nil
	}
	PromptRegion[thisTp].Redraw = tmpLine
	PromptRegion[thisTp].LineNr = TerminalInfo.Height - maxTp + thisTp
}

// restorePromptLines redraws or clears any screen lines that were displaced
// by prompt regions during multi-line input.
func (ss *ScreenState) restorePromptLines(maxTp int) {
	for index := 1; index <= maxTp; index++ {
		if PromptRegion[index].Redraw != nil {
			ss.DrawLine(PromptRegion[index].Redraw)
		} else if PromptRegion[index].LineNr != 0 {
			VduMoveCurs(1, PromptRegion[index].LineNr)
			VduClearEOL()
		}
	}
}

// GetLineP gets a line from the user
func (ss *ScreenState) GetLineP(
	frame *FrameObject,
	prompt string,
	maxTp int,
	thisTp int,
) (*StrObject, int) {
	maxTp = iabs(maxTp)

	if TtControlC {
		return nil, 0
	}

	if LudwigMode != LudwigScreen {
		fmt.Print(prompt)

		line, err := ss.StdinReader.ReadString('\n')
		if err != nil && len(line) == 0 {
			// no input was read before EOF or another error
			return nil, 0
		}
		line = strings.TrimRight(line, "\r\n")
		if len(line) > MaxStrLen {
			line = line[:MaxStrLen]
		}
		return NewStrObjectFrom(line), len(line)
	}

	ss.computePromptPosition(frame, thisTp, maxTp)
	if PromptRegion[thisTp].LineNr != 0 {
		VduMoveCurs(1, PromptRegion[thisTp].LineNr)
	}
	outbuf, outlen := VduGetInput(prompt, MaxStrLen)

	if TtControlC {
		return nil, 0
	}

	if outlen == 0 {
		for index := thisTp + 1; index <= maxTp; index++ {
			PromptRegion[index].LineNr = 0
			PromptRegion[index].Redraw = nil
		}
	}
	if thisTp == maxTp || outlen == 0 {
		ss.restorePromptLines(maxTp)
	}
	return outbuf, outlen
}

// FreeBottomLine frees the bottom line of the screen
func (ss *ScreenState) FreeBottomLine() {
	// This routine assumes that the editor is in SCREEN mode.
	// This routine frees the bottom line of the screen for use by the caller.
	// The main use of the area is the outputting of messages.
	if ss.Frame == nil {
		// IF SCREEN NOT MAPPED.
		VduDisplayCrLf()
		VduDeleteLines(1)
		return
	}
	ss.NeedsFix = true
	// IF BOTTOM LINE FREE.
	if (ss.MsgRow > TerminalInfo.Height) && (ss.BotLine.ScrRowNr < TerminalInfo.Height) {
		// Nothing
	} else if ss.BotLine.ScrRowNr+2 < ss.MsgRow {
		// IF ROOM BELOW BOT LINE.
		// +2 because of <eos> line.
		VduMoveCurs(1, ss.BotLine.ScrRowNr+2)
		VduDeleteLines(1)
	} else if ss.TopLine.ScrRowNr != 1 {
		// IF TOP LINE FREE.
		ss.Scroll(1, false)
	} else {
		if ss.BotLine.ScrRowNr+1 < ss.MsgRow {
			// IF ROOM FOR MORE MSGS.
			VduMoveCurs(1, ss.BotLine.ScrRowNr+1)
			VduDeleteLines(1)
		} else if (ss.Frame.Dot.Line != ss.TopLine) &&
			!((ss.Frame.Dot.Line != ss.BotLine) &&
				(ss.BotLine.ScrRowNr == TerminalInfo.Height)) {
			// IF DOT NOT ON TOP LINE,
			// AND WE CANT USE THE BOT.
			ss.Scroll(1, false)
		} else if ss.MsgRow <= TerminalInfo.Height/2 {
			// 1/2 SCREEN ALREADY MSGS.
			VduMoveCurs(1, ss.MsgRow)
			VduDeleteLines(1)
			return
		} else {
			// CONTRACT SCREEN 1 LINE.
			ss.BotLine.ScrRowNr = 0
			ss.BotLine = ss.BotLine.BLink
			VduMoveCurs(1, ss.MsgRow-1)
			VduDeleteLines(1)
		}
	}
	ss.MsgRow -= 1
}

// Verify gets verification from user
func (ss *ScreenState) Verify(frame *FrameObject, prompt string) VerifyResponse {
	const verHeight = 4

	verify := VerifyReplyQuit

	oldHeight := frame.ScrHeight
	oldTopM := frame.MarginTop
	oldBotM := frame.MarginBottom
	if oldHeight > verHeight {
		frame.MarginTop = verHeight / 2
		frame.ScrHeight = verHeight
		frame.MarginBottom = verHeight / 2
	}

	usePrompt := true
	var key int
	var more bool

	for {
		switch LudwigMode {
		case LudwigScreen:
			ss.Fixup(frame)
			VduBold()
			if usePrompt {
				ss.Message(prompt)
			} else {
				ss.Message(YNAQM_MSG)
			}
			VduNormal()
			VduMoveCurs(
				frame.Dot.Col-frame.ScrOffset,
				frame.Dot.Line.ScrRowNr,
			)
			key = VduGetKey()
			if key >= 'a' && key <= 'z' {
				key = key - 'a' + 'A'
			}
			if key == 13 {
				key = 'N' // RETURN <=> NO
			}
			ss.ClearMsgs(false)

		case LudwigBatch, LudwigHardcopy:
			var response *StrObject
			var respLen int
			if usePrompt {
				response, respLen = ss.GetLineP(frame, prompt, 1, 1)
			} else {
				response, respLen = ss.GetLineP(frame, YNAQM_MSG, 1, 1)
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
			break
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
				if ss.TopLine != nil {
					frame.ScrHeight = ss.BotLine.ScrRowNr + 1 - ss.TopLine.ScrRowNr
				}
				if key == 'M' {
					key = '1'
				}
				if key-'0'+frame.ScrHeight < TerminalInfo.Height {
					frame.ScrHeight += key - '0'
				} else {
					frame.ScrHeight = TerminalInfo.Height
				}
				if ss.TopLine == nil {
					ss.Load(frame.Dot.Line)
				} else {
					ss.expand(true, true)
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

	frame.ScrHeight = oldHeight
	frame.MarginTop = oldTopM
	frame.MarginBottom = oldBotM

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
		VduDisplayStr(str, false)
	} else {
		fmt.Printf("%d", intVal)
	}
}

// ScreenWriteCh writes a character with indent
func ScreenWriteCh(indent int, ch byte) {
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+string(ch), false)
	} else {
		fmt.Print(spc(indent) + string(ch))
	}
}

// ScreenWriteStr writes a string with indent
func ScreenWriteStr(indent int, str string) {
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+str, true)
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
		VduDisplayStr(spc(indent)+str[:strLen]+spc(trailingSpaces), true)
	} else {
		fmt.Print(spc(indent) + str)
	}
}

// ScreenWriteNameStr writes a name string with indent and width
func ScreenWriteNameStr(indent int, str string, width int) {
	strLen := min(len(str), width)
	trailingSpaces := width - strLen
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent)+str[:strLen]+spc(trailingSpaces), true)
	} else {
		fmt.Print(spc(indent))
		fmt.Print(str[:strLen])
		fmt.Print(spc(trailingSpaces))
	}
}

// ScreenWriteFileNameStr writes a file name with the given indent
func ScreenWriteFileNameStr(indent int, str string) {
	if LudwigMode == LudwigScreen {
		VduDisplayStr(spc(indent), false)
		VduDisplayStr(str, false)
	} else {
		fmt.Print(spc(indent))
		fmt.Print(str)
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
