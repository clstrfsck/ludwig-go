/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         FRAME
//
// Description:  Creation/destruction, manipulation of Frames.

package ludwig

import (
	"fmt"
	"strings"
)

const (
	endOfFile = "<End of File>   "
	newValues = "  New Values: "
)

type tparParser struct {
	request *TParObject
	pos     int
}

// newTparParser returns a newly initialised parser
func newTparParser(request *TParObject) *tparParser {
	return &tparParser{request: request, pos: 1}
}

// nextChar gets the next non-space character from a tpar
func (p *tparParser) nextChar() byte {
	for (p.pos < p.request.Len) && (p.request.Str.Get(p.pos) == ' ') {
		p.pos += 1
	}
	var ch byte
	if (p.pos > p.request.Len) || (p.request.Str.Get(p.pos) == ' ') {
		ch = 0
	} else {
		ch = p.request.Str.Get(p.pos)
	}
	if p.pos <= p.request.Len {
		p.pos += 1
	}
	return ch
}

// toInt returns the next int, true, or 0, false
func (p *tparParser) toInt() (int, bool) {
	return TparToInt(p.request, &p.pos)
}

// setOptions sets options on CurrentFrame and InitialOptions if requested
func (p *tparParser) setOptions(setInitial bool) bool {
	ok := false
	ch := p.nextChar()
	if ch == '(' {
		for {
			seton := true
			ch = p.nextChar()
			if ch == '-' {
				seton = false
				ch = p.nextChar()
			}
			if setInitial {
				setOpt(ch, seton, &InitialOptions)
			}
			ok = setOpt(ch, seton, &CurrentFrame.Options)
			ch = p.nextChar()
			if ch != ',' && ch != ')' {
				ScreenMessage(MsgSyntaxErrorInOptions)
				return false
			}
			if !ok || ch == ')' {
				break
			}
		}
	} else {
		// single option
		seton := true
		if ch == '-' {
			seton = false
			ch = p.nextChar()
		}
		if setInitial {
			setOpt(ch, seton, &InitialOptions)
		}
		ok = setOpt(ch, seton, &CurrentFrame.Options)
	}
	return ok
}

func (p *tparParser) setMode() bool {
	ch := p.nextChar()
	switch ch {
	case 'I':
		EditMode = ModeInsert
	case 'O':
		EditMode = ModeOvertype
	case 'C':
		EditMode = ModeCommand
	default:
		ScreenMessage(MsgModeError)
		// FIXME: Original always returns true, but this should probably fail.
		// return false
	}
	return true
}

func (p *tparParser) setCmdIntr() bool {
	if LudwigMode == LudwigScreen {
		var keyName strings.Builder
		terminate := false
		for p.pos <= p.request.Len && !terminate {
			if p.request.Str.Get(p.pos) == ',' {
				terminate = true
			} else {
				keyName.WriteByte(p.request.Str.Get(p.pos))
				p.pos += 1
			}
		}
		keyNameStr := keyName.String()

		if len(keyNameStr) == 1 {
			if ChIsPunctuation(rune(keyNameStr[0])) {
				CommandIntroducer = int(keyNameStr[0])
				VduNewIntroducer(CommandIntroducer)
				return true
			}
			ScreenMessage(MsgInvalidCmdIntroducer)
		} else if keyCode, found := UserKeyNameToCode(keyNameStr); found {
			if _, found := KeyIntroducers[keyCode]; found {
				ScreenMessage(MsgInvalidCmdIntroducer)
			} else {
				CommandIntroducer = keyCode
				VduNewIntroducer(CommandIntroducer)
				return true
			}
		} else {
			ScreenMessage(MsgUnrecognizedKeyName)
		}
	} else {
		ScreenMessage(MsgScreenModeOnly)
	}
	return false
}

// setTabs sets tab stops for the current frame
func (p *tparParser) setTabs(setInitial bool) bool {
	ch := p.nextChar()
	switch ch {
	case 'D': // default tabs
		if setInitial {
			InitialTabStops = DefaultTabStops
		}
		CurrentFrame.TabStops = DefaultTabStops

	case 'T': // template match
		if CurrentFrame.Dot.Line.Used > 0 {
			ts := CurrentFrame.Dot.Line.Str.Get(1) != ' '
			if setInitial {
				InitialTabStops[1] = ts
			}
			CurrentFrame.TabStops[1] = ts
		}
		for i := 2; i <= CurrentFrame.Dot.Line.Used; i++ {
			chi := CurrentFrame.Dot.Line.Str.Get(i)
			chim1 := CurrentFrame.Dot.Line.Str.Get(i - 1)
			if setInitial {
				InitialTabStops[i] = (chi != ' ') && (chim1 == ' ')
			}
			CurrentFrame.TabStops[i] = (chi != ' ') && (chim1 == ' ')
		}
		for i := CurrentFrame.Dot.Line.Used; i <= MaxStrLen; i++ {
			if setInitial {
				InitialTabStops[i] = false
			}
			CurrentFrame.TabStops[i] = false
		}

	case 'I': // insert tabs
		firstLine, lastLine := LinesCreate(1)
		LineChangeLength(firstLine, MaxStrLen)
		if setInitial {
			for i := 1; i <= MaxStrLen; i++ {
				if InitialTabStops[i] {
					firstLine.Str.Set(i, 'T')
				}
			}
			firstLine.Str.Set(InitialMarginLeft, 'L')
			firstLine.Str.Set(InitialMarginRight, 'R')
		} else {
			for i := 1; i < MaxStrLen; i++ {
				if CurrentFrame.TabStops[i] {
					firstLine.Str.Set(i, 'T')
				}
			}
			firstLine.Str.Set(CurrentFrame.MarginLeft, 'L')
			firstLine.Str.Set(CurrentFrame.MarginRight, 'R')
		}
		// Calculate used length
		firstLine.Used = firstLine.Str.Length(' ', MaxStrLen)
		LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line)
		MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot)
		CurrentFrame.TextModified = true
		MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified])

	case 'R': // Template Ruler
		i := 1
		legal := true
		const (
			lmNone = iota
			lmLeft
			lmRight
		)
		lastMargin := lmNone
		for i <= CurrentFrame.Dot.Line.Used && legal {
			chi := ChToUpper(CurrentFrame.Dot.Line.Str.Get(i))
			legal = (chi == 'T') || (chi == 'L') || (chi == 'R') || (chi == ' ')
			switch chi {
			case 'L':
				legal = legal && (lastMargin == lmNone)
				lastMargin = lmLeft
			case 'R':
				legal = legal && (lastMargin == lmLeft)
				lastMargin = lmRight
			}
			i++
		}
		legal = legal && (lastMargin == lmRight)
		if !legal {
			ScreenMessage(MsgInvalidRuler)
			return false
		}

		i = 1
		for i <= CurrentFrame.Dot.Line.Used {
			chi := ChToUpper(CurrentFrame.Dot.Line.Str.Get(i))
			if setInitial {
				InitialTabStops[i] = (chi != ' ')
			}
			CurrentFrame.TabStops[i] = (chi != ' ')
			switch chi {
			case 'L':
				if setInitial {
					InitialMarginLeft = i
				}
				CurrentFrame.MarginLeft = i
			case 'R':
				if setInitial {
					InitialMarginRight = i
				}
				CurrentFrame.MarginRight = i
			}
			i++
		}
		for j := CurrentFrame.Dot.Line.Used + 1; j <= MaxStrLen; j++ {
			if setInitial {
				InitialTabStops[j] = false
			}
			CurrentFrame.TabStops[j] = false
		}

		firstLine := CurrentFrame.Dot.Line
		dotCol := CurrentFrame.Dot.Col
		if !MarksSqueeze(firstLine, 1, firstLine.FLink, 1) {
			return false
		}
		if !LinesExtract(firstLine, firstLine) {
			return false
		}
		CurrentFrame.Dot.Col = dotCol

	case 'S': // Set tab
		if CurrentFrame.Dot.Col == MaxStrLenP {
			ScreenMessage(MsgOutOfRangeTabValue)
			return false
		}
		if setInitial {
			InitialTabStops[CurrentFrame.Dot.Col] = true
		}
		CurrentFrame.TabStops[CurrentFrame.Dot.Col] = true

	case 'C': // Clear tab
		if CurrentFrame.Dot.Col == MaxStrLenP {
			ScreenMessage(MsgOutOfRangeTabValue)
			return false
		}
		if setInitial {
			InitialTabStops[CurrentFrame.Dot.Col] = false
		}
		CurrentFrame.TabStops[CurrentFrame.Dot.Col] = false

	case 'W': // Regular width tabs
		if w, found := p.toInt(); found && w > 1 {
			var temptab TabArray
			temptab[0] = true
			temptab[MaxStrLenP] = true
			for i := 1; i <= MaxStrLen; i++ {
				if i%w == 1 {
					temptab[i] = true
				}
			}
			if setInitial {
				InitialTabStops = temptab
			}
			CurrentFrame.TabStops = temptab
		} else {
			return false
		}

	case '(': // multi-columns specified
		var temptab TabArray
		temptab[0] = true
		temptab[MaxStrLenP] = true
		for {
			n, found := p.toInt()
			if !found {
				ScreenMessage(MsgBadFormatInTabTable)
				return false
			}
			if n >= 1 && n <= MaxStrLen {
				temptab[n] = true
			} else {
				ScreenMessage(MsgOutOfRangeTabValue)
				return false
			}
			ch = p.nextChar()
			if ch != ',' && ch != ')' {
				ScreenMessage(MsgBadFormatInTabTable)
				return false
			}
			if ch == ')' {
				break
			}
		}
		if setInitial {
			InitialTabStops = temptab
		}
		CurrentFrame.TabStops = temptab

	default:
		ScreenMessage(MsgInvalidTOption)
		return false
	}
	return true
}

// setLRMargin sets the left and right margins
func (p *tparParser) setLRMargin(setInitial bool) bool {
	var tl, tr int
	if setInitial {
		tl = InitialMarginLeft
		tr = InitialMarginRight
	} else {
		tl = CurrentFrame.MarginLeft
		tr = CurrentFrame.MarginRight
	}
	if !p.getMargins(1, MaxStrLen, &tl, &tr, true) {
		return false
	}
	if tl < tr {
		if setInitial {
			InitialMarginLeft = tl
			InitialMarginRight = tr
		}
		CurrentFrame.MarginLeft = tl
		CurrentFrame.MarginRight = tr
	} else {
		ScreenMessage(MsgLeftMarginGeRight)
		return false
	}
	return true
}

// setTBMargin sets the top and bottom margins
func (p *tparParser) setTBMargin(setInitial bool) bool {
	var tt, tb int
	if setInitial {
		tt = InitialMarginTop
		tb = InitialMarginBottom
	} else {
		tt = CurrentFrame.MarginTop
		tb = CurrentFrame.MarginBottom
	}
	if !p.getMargins(0, CurrentFrame.ScrHeight, &tt, &tb, false) {
		return false
	}
	if tt+tb >= CurrentFrame.ScrHeight {
		ScreenMessage(MsgMarginOutOfRange)
		return false
	}
	if setInitial {
		InitialMarginTop = tt
		InitialMarginBottom = tb
	}
	CurrentFrame.MarginTop = tt
	CurrentFrame.MarginBottom = tb
	return true
}

// getMargins gets left/right or top/bottom margins from the tpar
func (p *tparParser) getMargins(loBnd int, hiBnd int, lower *int, upper *int, lr bool) bool {
	ch := p.nextChar()
	if ch != '(' {
		ScreenMessage(MsgMarginSyntaxError)
		return false
	}
	ch = p.nextChar()
	if ch == '.' {
		if lr {
			*lower = CurrentFrame.Dot.Col
		} else {
			*lower = CurrentFrame.Dot.Line.ScrRowNr
		}
		ch = p.nextChar()
	} else if !p.getMar(&ch, loBnd, hiBnd, lower) {
		return false
	}
	if ch == ',' {
		ch = p.nextChar()
		if ch == '.' {
			if lr {
				*upper = CurrentFrame.Dot.Col
			} else {
				*upper = CurrentFrame.ScrHeight - CurrentFrame.Dot.Line.ScrRowNr
			}
			ch = p.nextChar()
		} else if !p.getMar(&ch, loBnd, hiBnd, upper) {
			return false
		}
	}
	if ch != ')' {
		ScreenMessage(MsgMarginSyntaxError)
		return false
	}
	return true
}

// getMar gets a margin value from the tpar
func (p *tparParser) getMar(ch *byte, loBnd int, hiBnd int, margin *int) bool {
	if *ch >= '0' && *ch <= '9' {
		p.pos -= 1
		m, found := p.toInt()
		if !found {
			return false
		}
		if m < loBnd || m > hiBnd {
			ScreenMessage(MsgMarginOutOfRange)
			return false
		}
		*ch = p.nextChar()
		*margin = m
	}
	return true
}

// FrameEdit creates or edits a frame with the specified name.
// This is the \ED command. If frame_name doesn't exist, then it is created.
func FrameEdit(frameName string) bool {
	fname := frameName
	if fname == "" {
		fname = DefaultFrameName
	}

	var ptr *SpanObject
	var oldp *SpanObject

	if SpanFind(fname, &ptr, &oldp) {
		if ptr.Frame != nil {
			if ptr.Frame != CurrentFrame {
				ptr.Frame.ReturnFrame = CurrentFrame
				CurrentFrame = ptr.Frame
			}
			return true
		}
		ScreenMessage(MsgSpanOfThatNameExists)
		return false
	}

	// No Span/Frame of that name exists, create one.
	fptr := &FrameObject{}
	sptr := &SpanObject{}

	gptr := LineEOPCreate(fptr)

	// Set up span object
	sptr.BLink = oldp
	sptr.FLink = ptr
	if oldp == nil {
		FirstSpan = sptr
	} else {
		oldp.FLink = sptr
	}
	if ptr != nil {
		ptr.BLink = sptr
	}
	sptr.Name = fname
	sptr.Frame = fptr
	sptr.MarkOne = nil
	sptr.MarkTwo = nil
	sptr.Code = nil

	MarkCreate(gptr.FirstLine, 1, &sptr.MarkOne)
	MarkCreate(gptr.LastLine, 1, &sptr.MarkTwo)
	fptr.Dot = nil
	MarkCreate(gptr.FirstLine, InitialMarginLeft, &fptr.Dot)

	// Initialize frame object
	fptr.FirstGroup = gptr
	fptr.LastGroup = gptr
	fptr.Marks = InitialMarks
	fptr.ScrHeight = InitialScrHeight
	fptr.ScrWidth = InitialScrWidth
	fptr.ScrOffset = InitialScrOffset
	fptr.ScrDotLine = 1
	fptr.Span = sptr
	fptr.ReturnFrame = CurrentFrame
	fptr.InputCount = 0
	fptr.SpaceLimit = FileData.Space
	fptr.SpaceLeft = FileData.Space
	fptr.TextModified = false
	fptr.MarginLeft = InitialMarginLeft
	fptr.MarginRight = InitialMarginRight
	fptr.MarginTop = InitialMarginTop
	fptr.MarginBottom = InitialMarginBottom
	fptr.TabStops = InitialTabStops
	fptr.Options = InitialOptions
	fptr.InputFile = 0
	fptr.OutputFile = 0
	fptr.GetTpar = TParObject{}
	fptr.GetPatternPtr = nil
	fptr.EqsTpar = TParObject{}
	fptr.EqsPatternPtr = nil
	fptr.Rep1Tpar = TParObject{}
	fptr.RepPatternPtr = nil
	fptr.Rep2Tpar = TParObject{}
	fptr.VerifyTpar = TParObject{}

	LineChangeLength(gptr.LastLine, NameLen+len(endOfFile))
	// Copy end-of-file message and frame name
	lineLen := gptr.LastLine.Len()
	gptr.LastLine.Str.FillCopyBytes([]byte(endOfFile), 1, lineLen, ' ')
	eofLen := len(endOfFile) + 1
	gptr.LastLine.Str.FillCopyBytes([]byte(fname), eofLen, lineLen-eofLen, ' ')
	gptr.LastLine.Used = 0 // Special feature of the NULL line!
	CurrentFrame = fptr
	return true
}

// FrameKill destroys the specified frame.
// You can't kill frame C or OOPS or the current frame.
func FrameKill(frameName string) bool {
	var oldp *SpanObject
	var sptr *SpanObject

	if !SpanFind(frameName, &sptr, &oldp) {
		ScreenMessage(MsgNoSuchFrame)
		return false
	}
	if sptr.Frame == nil {
		ScreenMessage(MsgNoSuchFrame)
		return false
	}

	thisFrame := sptr.Frame
	if thisFrame == CurrentFrame || thisFrame == ScrFrame ||
		thisFrame.Options.Has(OptSpecialFrame) {
		ScreenMessage(MsgCantKillFrame)
		return false
	}

	if thisFrame.InputFile != 0 || thisFrame.OutputFile != 0 {
		ScreenMessage(MsgFrameHasFilesAttached)
		return false
	}

	// We are now free to destroy this frame
	// Step 1: remove all ERs back to this frame and all spans into this frame
	oldp = FirstSpan
	for oldp != nil {
		sptr = oldp.FLink
		if oldp.Frame != nil {
			if oldp.Frame.ReturnFrame == thisFrame {
				oldp.Frame.ReturnFrame = nil
			}
		} else if oldp.MarkOne.Line.Group.Frame == thisFrame {
			if !SpanDestroy(&oldp) {
				return false
			}
		}
		oldp = sptr
	}

	// Step 2: Destroy the Span
	thisFrame.Span.Frame = nil
	spanPtr := thisFrame.Span
	if !SpanDestroy(&spanPtr) {
		return false
	}

	// Step 3a: Destroy all internal lines
	MarkDestroy(&thisFrame.Dot)
	for i := 0; i <= MaxMarkNumber; i++ {
		MarkDestroy(&thisFrame.Marks[i])
	}

	ptr2 := thisFrame.LastGroup.LastLine.BLink
	if ptr2 != nil {
		ptr1 := thisFrame.FirstGroup.FirstLine
		if !LinesExtract(ptr1, ptr2) {
			return false
		}
	}

	// Step 3b: Destroy the <eop> line
	thisFrame.FirstGroup = nil

	// Step 4: Dispose of the frame header and any pattern tables attached
	if !PatternDFATableKill(&thisFrame.EqsPatternPtr) {
		return false
	}
	if !PatternDFATableKill(&thisFrame.GetPatternPtr) {
		return false
	}
	if !PatternDFATableKill(&thisFrame.RepPatternPtr) {
		return false
	}

	return true
}

// setmemory sets the memory allocation for the current frame
func setmemory(sz int, setInitial bool) bool {
	if sz >= MaxSpace {
		sz = MaxSpace
	}
	if setInitial {
		FileData.Space = sz
	}

	usedStorage := CurrentFrame.SpaceLimit - CurrentFrame.SpaceLeft
	minSize := min(usedStorage+800, CurrentFrame.SpaceLimit)
	if sz < minSize {
		sz = minSize
	}
	CurrentFrame.SpaceLimit = sz
	CurrentFrame.SpaceLeft = sz - usedStorage
	return true
}

// FrameSetHeight sets the screen height for the current frame
func FrameSetHeight(sh int, setInitial bool) bool {
	if sh >= 1 && sh <= TerminalInfo.Height {
		if setInitial {
			InitialScrHeight = sh
		}
		CurrentFrame.ScrHeight = sh
		band := sh / 6
		if setInitial {
			InitialMarginTop = band
		}
		CurrentFrame.MarginTop = band
		if setInitial {
			InitialMarginBottom = band
		}
		CurrentFrame.MarginBottom = band
		return true
	}
	ScreenMessage(MsgInvalidScreenHeight)
	return false
}

// setwidth sets the screen width for the current frame
func setwidth(wid int, setInitial bool) bool {
	if wid >= 10 && wid <= TerminalInfo.Width {
		if setInitial {
			InitialScrWidth = wid
		}
		CurrentFrame.ScrWidth = wid
		return true
	}
	ScreenMessage(MsgScreenWidthInvalid)
	return false
}

// showOptions displays the current frame options
func showOptions() {
	ScreenUnload()
	ScreenHome(true)
	ScreenWriteStr(0, "    Ludwig Option         Code    State")
	ScreenWriteln()
	ScreenWriteStr(0, "    --------------------  ----    -----")
	ScreenWriteln()
	ScreenWriteln()
	ScreenWriteStr(4, "Show current options  S")
	ScreenWriteln()
	ScreenWriteStr(4, "Auto-indenting        I       ")
	if CurrentFrame.Options.Has(OptAutoIndent) {
		ScreenWriteStr(0, "On")
	} else {
		ScreenWriteStr(0, "Off")
	}
	ScreenWriteln()
	ScreenWriteStr(4, "New Line              N       ")
	if CurrentFrame.Options.Has(OptNewLine) {
		ScreenWriteStr(0, "On")
	} else {
		ScreenWriteStr(0, "Off")
	}
	ScreenWriteln()
	ScreenWriteStr(4, "Wrap at Right Margin  W       ")
	if CurrentFrame.Options.Has(OptAutoWrap) {
		ScreenWriteStr(0, "On")
	} else {
		ScreenWriteStr(0, "Off")
	}
	ScreenWriteln()
	ScreenWriteln()
	ScreenPause()
	ScreenHome(true) // wipe out the display
}

// setOpt sets a single option
func setOpt(ch byte, seton bool, options *FrameOptions) bool {
	switch ch {
	case 'S':
		showOptions()
	case 'I':
		if seton {
			options.Set(OptAutoIndent)
		} else {
			options.Clear(OptAutoIndent)
		}
	case 'W':
		if seton {
			options.Set(OptAutoWrap)
		} else {
			options.Clear(OptAutoWrap)
		}
	case 'N':
		if seton {
			options.Set(OptNewLine)
		} else {
			options.Clear(OptNewLine)
		}
	default:
		ScreenMessage(MsgUnknownOption)
		return false
	}
	return true
}

// setparam parses and sets frame parameters
func setparam(request *TParObject) bool {
	p := newTparParser(request)
	ch := p.nextChar()
	for ch != 0 {
		setInitial := false
		if ch == '$' { // setting an initial value for a new frame
			setInitial = true
			ch = p.nextChar()
		}
		if p.nextChar() != '=' {
			ScreenMessage(MsgOptionsSyntaxError)
			return false
		}
		ok := false
		switch ch {
		case 'O':
			ok = p.setOptions(setInitial)
		case 'S':
			if mem, found := p.toInt(); found {
				ok = setmemory(mem, setInitial)
			}
		case 'H':
			if height, found := p.toInt(); found {
				ok = FrameSetHeight(height, setInitial)
			}
		case 'W':
			if width, found := p.toInt(); found {
				ok = setwidth(width, setInitial)
			}
		case 'C':
			ok = p.setCmdIntr()
		case 'T':
			ok = p.setTabs(setInitial)
		case 'M':
			ok = p.setLRMargin(setInitial)
		case 'V':
			ok = p.setTBMargin(setInitial)
		case 'K':
			ok = p.setMode()
		default:
			ScreenMessage(MsgInvalidParameterCode)
			return false
		}
		if !ok {
			return false
		}
		ch = p.nextChar()
		if ch == ',' || ch == 0 {
			ch = p.nextChar()
		} else {
			ScreenMessage(MsgSyntaxErrorInParamCmd)
			return false
		}
	}
	return true
}

// displayOption displays a single option character
func displayOption(ch byte, first *bool) {
	if *first {
		ScreenWriteCh(0, '(')
	} else {
		ScreenWriteCh(0, ',')
	}
	ScreenWriteCh(0, ch)
	*first = false
}

// printOptions prints the frame options
func printOptions(options FrameOptions) {
	count := 1
	first := true
	ScreenWriteCh(0, ' ')
	if options.Has(OptAutoIndent) {
		displayOption('I', &first)
		count += 2
	}
	if options.Has(OptAutoWrap) {
		displayOption('W', &first)
		count += 2
	}
	if options.Has(OptNewLine) {
		displayOption('N', &first)
		count += 2
	}
	if first {
		s := "  None    "
		ScreenWriteStr(0, s)
		count += len(s)
	} else {
		ScreenWriteCh(0, ')')
		count++
	}
	// Pad to 14 characters
	if count < 14 {
		ScreenWriteStr(0, strings.Repeat(" ", 14-count))
	}
}

// printMargins prints margin values
func printMargins(m1 int, m2 int) {
	s := fmt.Sprintf(" (%d,%d)", m1, m2)
	ScreenWriteStr(0, fmt.Sprintf("%-14s", s))
}

// FrameParameter handles the frame parameters command
func FrameParameter(tpar *TParObject) bool {
	request := TParObject{}
	if !TparGet1(tpar, CmdFrameParameters, &request) {
		return false
	}
	if request.Len > 0 {
		return setparam(&request)
	}

	// Display parameters and stats
	ScreenUnload()
	ScreenHome(true)
	for {
		ScreenHome(false) // Don't clear the screen here!
		ScreenWriteStr(0, " Ludwig ")
		for i := 0; i < 8 && i < len(LudwigVersion); i++ {
			ScreenWriteCh(0, LudwigVersion[i])
		}
		ScreenWriteStr(5, "Parameters      Frame: ")
		ScreenWriteNameStr(0, CurrentFrame.Span.Name, NameLen)
		ScreenWritelnClel()
		ScreenWriteStr(0, " ===============     ==========      =====")
		ScreenWritelnClel()
		ScreenWritelnClel()
		ScreenWriteStr(3, "Unused  memory available in frame    =")
		ScreenWriteInt(CurrentFrame.SpaceLeft, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "The number of lines in this frame    =")
		temp := LineToNumber(CurrentFrame.LastGroup.LastLine) - 1
		ScreenWriteInt(temp, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Lines read from input file so far    =")
		ScreenWriteInt(CurrentFrame.InputCount, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Current Line number in this frame    =")
		temp = LineToNumber(CurrentFrame.Dot.Line)
		ScreenWriteInt(temp, 9)
		ScreenWritelnClel()
		ScreenWritelnClel()
		ScreenWriteStr(9, "Parameters")
		ScreenWriteStr(41, "Defaults")
		ScreenWritelnClel()
		ScreenWriteStr(9, "----------")
		ScreenWriteStr(41, "--------")
		ScreenWritelnClel()
		ScreenWriteStr(3, "Keyboard Mode                      K =")
		switch EditMode {
		case ModeOvertype:
			ScreenWriteStr(1, "Overtype Mode")
		case ModeInsert:
			ScreenWriteStr(1, "Insert Mode")
		case ModeCommand:
			ScreenWriteStr(1, "Command Mode")
		}
		ScreenWritelnClel()
		if LudwigMode == LudwigScreen {
			ScreenWriteStr(3, "Command introducer                 C = ")
			if keyName, found := UserKeyCodeToName(CommandIntroducer); found {
				ScreenWriteStr(0, keyName)
			} else {
				ScreenWriteCh(0, byte(CommandIntroducer))
			}
			ScreenWritelnClel()
		}
		ScreenWriteStr(3, "Maximum memory available in frame  S =")
		ScreenWriteInt(CurrentFrame.SpaceLimit, 9)
		ScreenWriteStr(5, "  --  ")
		ScreenWriteInt(FileData.Space, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Screen height  (lines displayed)   H =")
		ScreenWriteInt(CurrentFrame.ScrHeight, 9)
		ScreenWriteStr(5, "  --  ")
		ScreenWriteInt(InitialScrHeight, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Screen width   (characters)        W =")
		ScreenWriteInt(CurrentFrame.ScrWidth, 9)
		ScreenWriteStr(5, "  --  ")
		ScreenWriteInt(InitialScrWidth, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Editing options                    O =")
		printOptions(CurrentFrame.Options)
		ScreenWriteStr(0, "  --  ")
		printOptions(InitialOptions)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Horizontal margins                 M =")
		printMargins(CurrentFrame.MarginLeft, CurrentFrame.MarginRight)
		ScreenWriteStr(0, "  --  ")
		printMargins(InitialMarginLeft, InitialMarginRight)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Vertical margins                   V =")
		printMargins(CurrentFrame.MarginTop, CurrentFrame.MarginBottom)
		ScreenWriteStr(0, "  --  ")
		printMargins(InitialMarginTop, InitialMarginBottom)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Tab settings                       T =")
		ScreenWritelnClel()
		for i := 1; i <= CurrentFrame.ScrWidth; i++ {
			if i == CurrentFrame.MarginLeft {
				ScreenWriteCh(0, 'L')
			} else if i == CurrentFrame.MarginRight {
				ScreenWriteCh(0, 'R')
			} else if CurrentFrame.TabStops[i] {
				ScreenWriteCh(0, 'T')
			} else {
				ScreenWriteCh(0, ' ')
			}
		}
		ScreenWriteln()
		ScreenWritelnClel()
		request.Str, request.Len = ScreenGetLineP(newValues, 1, 1)
		if request.Len > 0 {
			request.Str.ApplyN(ChToUpper, request.Len, 1)
			if !setparam(&request) {
				ScreenBeep()
			}
		}
		if TtControlC || request.Len == 0 {
			break
		}
	}
	return true
}
