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
	"strings"
)

const (
	endOfFile = "<End of File>   "
	newValues = "  New Values: "
)

func isNPunct(ch byte) bool {
	return !ChIsPunctuation(rune(ch))
}

// FrameEdit creates or edits a frame with the specified name.
// This is the \ED command. If frame_name doesn't exist, then it is created.
func FrameEdit(frameName string) bool {
	const (
		spnCreated  = 0x0001
		frmCreated  = 0x0002
		mrk1Created = 0x0004
		mrk2Created = 0x0008
		grpCreated  = 0x0010
		dotCreated  = 0x0020
	)

	created := 0
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
	created |= frmCreated | spnCreated

	var gptr *GroupObject
	if LineEOPCreate(fptr, &gptr) {
		created |= grpCreated

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

		if MarkCreate(gptr.FirstLine, 1, &sptr.MarkOne) {
			created |= mrk1Created
			if MarkCreate(gptr.LastLine, 1, &sptr.MarkTwo) {
				created |= mrk2Created
				fptr.Dot = nil
				if MarkCreate(gptr.FirstLine, InitialMarginLeft, &fptr.Dot) {
					created |= dotCreated
				}
			}
		}

		if (created & dotCreated) != 0 {
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

			if LineChangeLength(gptr.LastLine, NameLen+len(endOfFile)) {
				// Copy end-of-file message and frame name
				if gptr.LastLine.Str != nil {
					lineLen := gptr.LastLine.Len()
					gptr.LastLine.Str.FillCopyBytes([]byte(endOfFile), 1, lineLen, ' ')
					eofLen := len(endOfFile) + 1
					gptr.LastLine.Str.FillCopyBytes([]byte(fname), eofLen, lineLen-eofLen, ' ')
					gptr.LastLine.Used = 0 // Special feature of the NULL line!
					CurrentFrame = fptr
					return true
				}
			}
		}
	}

	// Something terrible has happened - cleanup
	if (created & mrk1Created) != 0 {
		MarkDestroy(&sptr.MarkOne)
	}
	if (created & mrk2Created) != 0 {
		MarkDestroy(&sptr.MarkTwo)
	}
	if (created & grpCreated) != 0 {
		LineEOPDestroy(&gptr)
	}

	return false
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
	if !MarkDestroy(&thisFrame.Dot) {
		return false
	}
	for i := MinMarkNumber; i <= MaxMarkNumber; i++ {
		if thisFrame.Marks[i-MinMarkNumber] != nil {
			if !MarkDestroy(&thisFrame.Marks[i-MinMarkNumber]) {
				return false
			}
		}
	}

	ptr2 := thisFrame.LastGroup.LastLine.BLink
	if ptr2 != nil {
		ptr1 := thisFrame.FirstGroup.FirstLine
		if !LinesExtract(ptr1, ptr2) {
			return false
		}
		if !LinesDestroy(&ptr1, &ptr2) {
			return false
		}
	}

	// Step 3b: Destroy the <eop> line
	if !LineEOPDestroy(&thisFrame.FirstGroup) {
		return false
	}

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

// nextchar gets the next non-space character from a tpar
func nextchar(request *TParObject, pos *int) byte {
	for (*pos < request.Len) && (request.Str.Get(*pos) == ' ') {
		*pos++
	}
	var ch byte
	if (*pos > request.Len) || (request.Str.Get(*pos) == ' ') {
		ch = 0
	} else {
		ch = request.Str.Get(*pos)
	}
	if *pos <= request.Len {
		*pos++
	}
	return ch
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
	minSize := usedStorage + 800
	if minSize > CurrentFrame.SpaceLimit {
		minSize = CurrentFrame.SpaceLimit
	}
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

// setOptions sets frame options
func setOptions(request *TParObject, pos *int, setInitial bool) bool {
	ok := false
	ch := nextchar(request, pos)
	if ch == '(' {
		for {
			seton := true
			ch = nextchar(request, pos)
			if ch == '-' {
				seton = false
				ch = nextchar(request, pos)
			}
			if setInitial {
				setOpt(ch, seton, &InitialOptions)
			}
			ok = setOpt(ch, seton, &CurrentFrame.Options)
			ch = nextchar(request, pos)
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
			ch = nextchar(request, pos)
		}
		if setInitial {
			setOpt(ch, seton, &InitialOptions)
		}
		ok = setOpt(ch, seton, &CurrentFrame.Options)
	}
	return ok
}

// setcmdintr sets the command introducer key
func setcmdintr(request *TParObject, pos *int) bool {
	if LudwigMode == LudwigScreen {
		var keyName strings.Builder
		terminate := false
		for *pos <= request.Len && !terminate {
			if request.Str.Get(*pos) == ',' {
				terminate = true
			} else {
				keyName.WriteByte(request.Str.Get(*pos))
				*pos++
			}
		}
		keyNameStr := keyName.String()

		var keyCode int
		if len(keyNameStr) == 1 {
			if ChIsPunctuation(rune(keyNameStr[0])) {
				CommandIntroducer = int(keyNameStr[0])
				VduNewIntroducer(CommandIntroducer)
				return true
			}
			ScreenMessage(MsgInvalidCmdIntroducer)
		} else if UserKeyNameToCode(keyNameStr, &keyCode) {
			if KeyIntroducers[keyCode] {
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

// setMode sets the editing mode
func setMode(request *TParObject, pos *int) bool {
	ch := nextchar(request, pos)
	if ch == 'I' {
		EditMode = ModeInsert
	} else if ch == 'O' {
		EditMode = ModeOvertype
	} else if ch == 'C' {
		EditMode = ModeCommand
	} else {
		ScreenMessage(MsgModeError)
		// FIXME: Original always returns true, but this should probably fail.
		// return false
	}
	return true
}

// setTabs sets tab stops for the current frame
func setTabs(request *TParObject, pos *int, setInitial bool) bool {
	ch := nextchar(request, pos)
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
		var firstLine *LineHdrObject
		var lastLine *LineHdrObject
		if !LinesCreate(1, &firstLine, &lastLine) {
			return false
		}
		if !LineChangeLength(firstLine, MaxStrLen) {
			return false
		}
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
		if !LinesInject(firstLine, lastLine, CurrentFrame.Dot.Line) {
			return false
		}
		if !MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Dot) {
			return false
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(firstLine, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified-MinMarkNumber]) {
			return false
		}

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
		if !LinesDestroy(&firstLine, &firstLine) {
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

	case '(': // multi-columns specified
		var temptab TabArray
		for i := range temptab {
			temptab[i] = false
		}
		temptab[0] = true
		temptab[MaxStrLenP] = true
		for {
			var j int
			if !TparToInt(request, pos, &j) {
				return false
			}
			if j >= 1 && j <= MaxStrLen {
				temptab[j] = true
			} else {
				ScreenMessage(MsgOutOfRangeTabValue)
				return false
			}
			ch = nextchar(request, pos)
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

// getMar gets a margin value from the tpar
func getMar(ch *byte, pos *int, request *TParObject, loBnd int, hiBnd int, margin *int) bool {
	if *ch >= '0' && *ch <= '9' {
		*pos--
		if !TparToInt(request, pos, margin) {
			return false
		}
		if *margin < loBnd || *margin > hiBnd {
			ScreenMessage(MsgMarginOutOfRange)
			return false
		}
		*ch = nextchar(request, pos)
	}
	return true
}

// getMargins gets left/right or top/bottom margins from the tpar
func getMargins(loBnd int, hiBnd int, request *TParObject, pos *int, lower *int, upper *int, lr bool) bool {
	ch := nextchar(request, pos)
	if ch != '(' {
		ScreenMessage(MsgMarginSyntaxError)
		return false
	}
	ch = nextchar(request, pos)
	if ch == '.' {
		if lr {
			*lower = CurrentFrame.Dot.Col
		} else {
			*lower = CurrentFrame.Dot.Line.ScrRowNr
		}
		ch = nextchar(request, pos)
	} else if !getMar(&ch, pos, request, loBnd, hiBnd, lower) {
		return false
	}
	if ch == ',' {
		ch = nextchar(request, pos)
		if ch == '.' {
			if lr {
				*upper = CurrentFrame.Dot.Col
			} else {
				*upper = CurrentFrame.ScrHeight - CurrentFrame.Dot.Line.ScrRowNr
			}
			ch = nextchar(request, pos)
		} else if !getMar(&ch, pos, request, loBnd, hiBnd, upper) {
			return false
		}
	}
	if ch != ')' {
		ScreenMessage(MsgMarginSyntaxError)
		return false
	}
	return true
}

// setLRMargin sets the left and right margins
func setLRMargin(request *TParObject, pos *int, setInitial bool) bool {
	var tl, tr int
	if setInitial {
		tl = InitialMarginLeft
		tr = InitialMarginRight
	} else {
		tl = CurrentFrame.MarginLeft
		tr = CurrentFrame.MarginRight
	}
	if !getMargins(1, MaxStrLen, request, pos, &tl, &tr, true) {
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
func setTBMargin(request *TParObject, pos *int, setInitial bool) bool {
	var tt, tb int
	if setInitial {
		tt = InitialMarginTop
		tb = InitialMarginBottom
	} else {
		tt = CurrentFrame.MarginTop
		tb = CurrentFrame.MarginBottom
	}
	if !getMargins(0, CurrentFrame.ScrHeight, request, pos, &tt, &tb, false) {
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

// setparam parses and sets frame parameters
func setparam(request *TParObject) bool {
	pos := 1
	ch := nextchar(request, &pos)
	for ch != 0 {
		setInitial := false
		if ch == '$' { // setting an initial value for a new frame
			setInitial = true
			ch = nextchar(request, &pos)
		}
		if nextchar(request, &pos) != '=' {
			ScreenMessage(MsgOptionsSyntaxError)
			return false
		}
		ok := false
		var temp int
		switch ch {
		case 'O':
			ok = setOptions(request, &pos, setInitial)
		case 'S':
			if TparToInt(request, &pos, &temp) {
				ok = setmemory(temp, setInitial)
			}
		case 'H':
			if TparToInt(request, &pos, &temp) {
				ok = FrameSetHeight(temp, setInitial)
			}
		case 'W':
			if TparToInt(request, &pos, &temp) {
				ok = setwidth(temp, setInitial)
			}
		case 'C':
			ok = setcmdintr(request, &pos)
		case 'T':
			ok = setTabs(request, &pos, setInitial)
		case 'M':
			ok = setLRMargin(request, &pos, setInitial)
		case 'V':
			ok = setTBMargin(request, &pos, setInitial)
		case 'K':
			ok = setMode(request, &pos)
		default:
			ScreenMessage(MsgInvalidParameterCode)
			return false
		}
		if !ok {
			return false
		}
		ch = nextchar(request, &pos)
		if ch == ',' || ch == 0 {
			ch = nextchar(request, &pos)
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
	ScreenWriteStr(0, " (")
	ScreenWriteInt(m1, 1)
	ScreenWriteCh(0, ',')
	ScreenWriteInt(m2, 1)
	ScreenWriteCh(0, ')')
	count := 6
	if m1 > 99 {
		count += 2
	} else if m1 > 9 {
		count++
	}
	if m2 > 99 {
		count += 2
	} else if m2 > 9 {
		count++
	}
	if count < 14 {
		ScreenWriteStr(0, strings.Repeat(" ", 14-count))
	}
}

// FrameParameter handles the frame parameters command
func FrameParameter(tpar *TParObject) bool {
	result := false
	tpar.Nxt = nil
	tpar.Con = nil
	var request TParObject
	if !TparGet1(tpar, CmdFrameParameters, &request) {
		goto l99
	}
	if request.Len > 0 {
		result = setparam(&request)
		goto l99
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
		var temp int
		if LineToNumber(CurrentFrame.LastGroup.LastLine, &temp) {
			temp--
		} else {
			temp = 0
		}
		ScreenWriteInt(temp, 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Lines read from input file so far    =")
		ScreenWriteInt(int(CurrentFrame.InputCount), 9)
		ScreenWritelnClel()
		ScreenWriteStr(3, "Current Line number in this frame    =")
		if !LineToNumber(CurrentFrame.Dot.Line, &temp) {
			temp = 0
		}
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
			var keyName string
			if UserKeyCodeToName(CommandIntroducer, &keyName) {
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
		ScreenGetLineP(newValues, &request.Str, &request.Len, 1, 1)
		if request.Len > 0 {
			ChApplyN(request.Str, ChToUpper, request.Len)
			if !setparam(&request) {
				ScreenBeep()
			}
		}
		if TtControlC || request.Len == 0 {
			break
		}
	}
	result = true
l99:
	TparCleanObject(&request)
	return result
}
