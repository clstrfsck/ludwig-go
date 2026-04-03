/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         CODE
//
// Description:  Ludwig compiler and interpreter.

package ludwig

type parseState struct {
	status       string
	key          int
	eoln         bool
	pc           int
	codeBase     int
	currentPoint MarkObject
	startPoint   MarkObject
	endPoint     MarkObject
	verifyCount  int
	fromSpan     bool
}

func isInterpCmd(cmd Commands) bool {
	switch cmd {
	case CmdPcJump,
		CmdExitTo,
		CmdFailTo,
		CmdIterate,
		CmdExitSuccess,
		CmdExitFail,
		CmdExitAbort,
		CmdExtended,
		CmdVerify,
		CmdNoop:
		return true
	}
	return false
}

// CodeDiscard releases the specified code and compacts the code array
func CodeDiscard(codeHead **CodeHeader) {
	if *codeHead == nil {
		return
	}

	(*codeHead).Ref--
	if (*codeHead).Ref == 0 {
		start := (*codeHead).Code
		size := (*codeHead).Len

		for source := start; source < start+size; source++ {
			if CompilerCode[source].Code != nil {
				CodeDiscard(&CompilerCode[source].Code)
			}
			CompilerCode[source].Tpar = nil
		}

		for source := start + size; source <= CodeTop; source++ {
			CompilerCode[source-size] = CompilerCode[source]
		}
		CodeTop -= size

		link := (*codeHead).BLink
		for link != CodeList {
			link.Code -= size
			link = link.BLink
		}

		(*codeHead).FLink.BLink = (*codeHead).BLink
		(*codeHead).BLink.FLink = (*codeHead).FLink

		*codeHead = nil
	}
}

func errorMsg(frame *FrameObject, ps *parseState, errText string) *FrameObject {
	ps.status = MsgSyntaxError
	if ps.fromSpan {
		// If possible, backup the current point one character
		if ps.currentPoint.Line != ps.startPoint.Line {
			if ps.currentPoint.Col > 1 {
				ps.currentPoint.Col--
			} else {
				ps.currentPoint.Line = ps.currentPoint.Line.BLink
				ps.currentPoint.Col = ps.currentPoint.Line.Used + 1
				if ps.currentPoint.Col > 1 {
					ps.currentPoint.Col--
				}
				if ps.currentPoint.Line == ps.startPoint.Line {
					if ps.currentPoint.Col < ps.startPoint.Col {
						ps.currentPoint.Col = ps.startPoint.Col
					}
				}
			}
		} else if ps.currentPoint.Col > ps.startPoint.Col {
			ps.currentPoint.Col--
		}

		// Insert the error message into the span
		if LudwigMode == LudwigScreen {
			newFrame, ok := FrameEdit(frame, ps.currentPoint.Line.Group.Frame.Span.Name)
			if !ok {
				return frame
			}
			frame = newFrame
			if frame.Marks[MarkEquals] != nil {
				MarkDestroy(&frame.Marks[MarkEquals])
			}
			eLine, _ := LinesCreate(1)

			str := NewBlankStrObject(MaxStrLen)
			i := ps.currentPoint.Col
			str.Set(i, '!')
			if i < MaxStrLen {
				i++
				str.Set(i, ' ')
			}
			for (i < MaxStrLen) && len(errText) > 0 {
				i++
				str.Set(i, errText[0])
				errText = errText[1:]
			}
			LineChangeLength(eLine, i)
			// "i" can't be zero here, so e_line->str != nullptr
			eLine.Str.Copy(str, 1, i, 1)
			eLine.Used = str.TrimmedLen(' ', i)
			LinesInject(eLine, eLine, ps.currentPoint.Line)
			MarkCreate(eLine, ps.currentPoint.Col, &frame.Dot)
		}
	}
	return frame
}

func nextKey(ps *parseState) bool {
	ps.eoln = false
	if !ps.fromSpan {
		ps.key = VduGetKey()
		if TtControlC {
			return false
		}
	} else {
		if (ps.currentPoint.Line == ps.endPoint.Line) &&
			(ps.currentPoint.Col == ps.endPoint.Col) {
			ps.key = 0 // finished span
		} else {
			if ps.currentPoint.Col <= ps.currentPoint.Line.Used {
				ps.key = int(ps.currentPoint.Line.Str.Get(ps.currentPoint.Col))
				ps.currentPoint.Col++
			} else if ps.currentPoint.Line != ps.endPoint.Line {
				ps.key = ' '
				ps.eoln = true
				ps.currentPoint.Line = ps.currentPoint.Line.FLink
				ps.currentPoint.Col = 1
			} else {
				ps.key = 0 // finished the span
			}
		}
	}
	return true
}

func nextNonBl(ps *parseState) bool {
	for {
		for {
			if !nextKey(ps) {
				return false
			}
			if ps.fromSpan {
				pscp := &ps.currentPoint
				if ps.key == '<' && pscp.Col <= pscp.Line.Used && pscp.Line.Str.Get(pscp.Col) == '>' {
					ps.key = 0
				}
			}
			if ps.key != ' ' {
				break
			}
		}
		if ps.key != '!' {
			return true
		}
		if !ps.fromSpan {
			ps.status = MsgCommentsIllegal
			return false
		}
		ps.currentPoint.Col = ps.currentPoint.Line.Used + 1
	}
}

func generate(
	ps *parseState,
	irep LeadParam,
	icnt int,
	iop Commands,
	itpar *TParObject,
	ilbl int,
	icode *CodeHeader,
) bool {
	ps.pc++
	if ps.codeBase+ps.pc > MaxCode {
		ps.status = MsgCompilerCodeOverflow
		return false
	}
	cc := &CompilerCode[ps.codeBase+ps.pc]
	cc.Rep = irep
	cc.Cnt = icnt
	cc.Op = iop
	cc.Tpar = itpar
	cc.Lbl = ilbl
	cc.Code = icode
	return true
}

func poke(codeBase, location, newLabel int) {
	CompilerCode[codeBase+location].Lbl = newLabel
}

func getCount(frame *FrameObject, ps *parseState, repCount *int) (*FrameObject, bool) {
	const maxRepCount = 65535

	if ps.key >= '0' && ps.key <= '9' {
		*repCount = 0
		for {
			digit := ps.key - '0'
			if *repCount <= (maxRepCount-digit)/10 {
				*repCount = *repCount*10 + digit
			} else {
				return errorMsg(frame, ps, "Count too large"), false
			}
			if !nextKey(ps) {
				return frame, false
			}
			if ps.key < '0' || ps.key > '9' {
				break
			}
		}
	} else {
		*repCount = 1
	}
	return frame, true
}

func scanLeadingParam(frame *FrameObject, ps *parseState, repSym *LeadParam, repCount *int) (*FrameObject, bool) {
	var ok bool
	switch ps.key {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		*repSym = LeadParamPInt
		return getCount(frame, ps, repCount)

	case '+':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamPlus
		*repCount = 1
		if ps.key >= '0' && ps.key <= '9' {
			*repSym = LeadParamPInt
			return getCount(frame, ps, repCount)
		}

	case '-':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamMinus
		*repCount = -1
		if ps.key >= '0' && ps.key <= '9' {
			*repSym = LeadParamNInt
			frame, ok = getCount(frame, ps, repCount)
			if !ok {
				return frame, false
			}
			*repCount = -*repCount
		}

	case '>', '.':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamPIndef
		*repCount = 0

	case '<', ',':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamNIndef
		*repCount = 0

	case '@':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamMarker
		frame, ok = getCount(frame, ps, repCount)
		if !ok {
			return frame, false
		}
		if (*repCount <= 0) || (*repCount > MaxUserMarkNumber) {
			return errorMsg(frame, ps, "Illegal mark number"), false
		}

	case '=':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamMarker
		*repCount = MarkEquals

	case '%':
		if !nextKey(ps) {
			return frame, false
		}
		*repSym = LeadParamMarker
		*repCount = MarkModified

	default:
		*repSym = LeadParamNone
		*repCount = 1
	}
	return frame, true
}

func scanTrailingParam(frame *FrameObject, ps *parseState, command Commands, repSym LeadParam) (*FrameObject, *TParObject, bool) {
	tc := CmdAttrib[command].TpCount
	var result *TParObject

	// Some commands only take trailing parameters when repcount is +ve
	if tc < 0 {
		if repSym == LeadParamMinus {
			tc = 0
		} else {
			tc = -tc
		}
	}

	if tc > 0 {
		if !nextKey(ps) {
			return frame, nil, false
		}
		parDelim := ps.key
		if ps.key < 0 || ps.key > MaxSetRange || !ChIsPunctuation(rune(parDelim)) {
			return errorMsg(frame, ps, "Illegal parameter delimiter"), nil, false
		}

		var tpl *TParObject
		for tci := 1; tci <= tc; tci++ {
			for {
				parLength := 0
				parString := *NewBlankStrObject(MaxStrLen)
				for {
					if !nextKey(ps) {
						return frame, nil, false
					}
					if ps.key == 0 {
						return errorMsg(frame, ps, "Missing trailing delimiter"), nil, false
					}
					parLength++
					parString.Set(parLength, byte(ps.key))
					if ps.eoln || ps.key == parDelim {
						break
					}
				}
				parLength--
				if ps.eoln && !CmdAttrib[command].TparInfo[tci].MlAllowed {
					return errorMsg(frame, ps, "Missing trailing delimiter"), nil, false
				}

				tp := &TParObject{
					Len: parLength,
					Dlm: byte(parDelim),
					Str: &parString,
				}

				if result == nil {
					// 1st time through
					result = tp
					tpl = tp
				} else {
					if tpl != nil {
						tpl.Con = tp
						tpl = tp
					} else {
						result.Nxt = tp
						tpl = tp
					}
				}

				if ps.key == parDelim {
					break
				}
			}
			tpl = nil
		}
	}
	return frame, result, true
}

func scanCommand(frame *FrameObject, ps *parseState, fullScan bool) (*FrameObject, bool) {
	var repCount int
	var repSym LeadParam
	var ok bool

	frame, ok = scanLeadingParam(frame, ps, &repSym, &repCount)
	if !ok {
		return frame, false
	}

	if ps.key >= 0 && ps.key <= MaxSetRange {
		ps.key = int(ChToUpper(byte(ps.key)))
	}

	command := Lookup[ps.key].Command
	for Prefixes.Bit(int(command)) != 0 {
		if !nextKey(ps) {
			return frame, false
		}
		if ps.key < 0 {
			return errorMsg(frame, ps, "Command not valid"), false
		}
		i := LookupExpPtr[command]
		j := LookupExpPtr[command+1]
		for (i < j) && (int(ChToUpper(byte(ps.key))) != int(LookupExp[i].Extn)) {
			i++
		}
		if i < j {
			command = LookupExp[i].Command
		} else {
			return errorMsg(frame, ps, "Command not valid"), false
		}
	}

	var pc1 int
	if ps.key == '(' {
		var pc2, pc3 int
		frame, ok = scanCompoundCommand(frame, ps, repSym, repCount, &pc1, &pc2, &pc3)
		if !ok {
			return frame, false
		}
	} else if command != CmdNoop {
		var tparam *TParObject
		var lookupCode *CodeHeader
		frame, ok = scanSimpleCommand(frame, ps, command, repSym, &repCount, &tparam, &lookupCode, &pc1, fullScan)
		if !ok {
			return frame, false
		}
	} else {
		return errorMsg(frame, ps, "Command not valid"), false
	}

	if fullScan {
		var pc4 int
		return scanExitHandler(frame, ps, pc1, &pc4, fullScan)
	}
	return frame, true
}

func scanExitHandler(frame *FrameObject, ps *parseState, pc1 int, pc4 *int, fullScan bool) (*FrameObject, bool) {
	if !nextNonBl(ps) {
		return frame, false
	}
	if ps.key == '[' {
		if !nextNonBl(ps) {
			return frame, false
		}
		var ok bool
		for (ps.key != ':') && (ps.key != ']') {
			// Construct exit part
			frame, ok = scanCommand(frame, ps, fullScan)
			if !ok {
				return frame, false
			}
		}
		if ps.key == ':' {
			// Jump over fail handler
			if !generate(ps, LeadParamNone, 0, CmdPcJump, nil, 0, nil) {
				return frame, false
			}
			*pc4 = ps.pc
			poke(ps.codeBase, pc1, ps.pc+1) // Set fail label for command
			if !nextNonBl(ps) {
				return frame, false
			}
			for ps.key != ']' {
				// Construct fail part
				frame, ok = scanCommand(frame, ps, fullScan)
				if !ok {
					return frame, false
				}
			}
			poke(ps.codeBase, *pc4, ps.pc+1) // End of fail handler
		} else {
			poke(ps.codeBase, pc1, ps.pc+1) // Set fail label
		}
		if !nextNonBl(ps) {
			return frame, false
		}
	}
	return frame, true
}

func scanCompoundCommand(frame *FrameObject, ps *parseState, repSym LeadParam, repCount int, pc1, pc2, pc3 *int) (*FrameObject, bool) {
	if repSym != LeadParamNone && repSym != LeadParamPlus && repSym != LeadParamPInt &&
		repSym != LeadParamPIndef {
		return errorMsg(frame, ps, "Illegal leading parameter"), false
	}
	if !generate(ps, LeadParamNone, 0, CmdExitTo, nil, 0, nil) {
		return frame, false
	}
	*pc2 = ps.pc
	if !generate(ps, LeadParamNone, 0, CmdFailTo, nil, 0, nil) {
		return frame, false
	}
	*pc1 = ps.pc
	*pc3 = ps.pc + 1
	if repSym != LeadParamPIndef {
		if !generate(ps, LeadParamNone, repCount, CmdIterate, nil, 0, nil) {
			return frame, false
		}
	}
	if !nextNonBl(ps) {
		return frame, false
	}
	for ps.key != ')' {
		var ok bool
		frame, ok = scanCommand(frame, ps, true)
		if !ok {
			return frame, false
		}
	}
	if !generate(ps, LeadParamNone, 0, CmdPcJump, nil, *pc3, nil) {
		return frame, false
	}
	poke(ps.codeBase, *pc2, ps.pc+1) // Fill in exit label
	return frame, true
}

func scanSimpleCommand(
	frame *FrameObject,
	ps *parseState,
	command Commands,
	repSym LeadParam,
	repCount *int,
	tparam **TParObject,
	lookupCode **CodeHeader,
	pc1 *int,
	fullScan bool,
) (*FrameObject, bool) {
	// Check if leading parameter is allowed
	lpAllowed := CmdAttrib[command].LpAllowed
	allowed := false
	// LpAllowed is a bitset stored as uint32
	if (lpAllowed & (1 << uint(repSym))) != 0 {
		allowed = true
	}
	if !allowed {
		return errorMsg(frame, ps, "Illegal leading parameter"), false
	}

	if command == CmdVerify {
		ps.verifyCount++
		if ps.verifyCount > MaxVerify {
			return errorMsg(frame, ps, "Too many verify commands in span"), false
		}
		*repCount = ps.verifyCount
	}

	*lookupCode = Lookup[ps.key].Code
	if Lookup[ps.key].Tpar == nil {
		if CmdAttrib[command].TpCount != 0 {
			if fullScan {
				var found bool
				if frame, *tparam, found = scanTrailingParam(frame, ps, command, repSym); !found {
					return frame, false
				}
			} else {
				*tparam = &TParObject{
					Str: EmptyStrObject(),
					Dlm: TpdPrompt,
				}
				tmpTp := *tparam
				for i := 2; i <= CmdAttrib[command].TpCount; i++ {
					tmpTp.Nxt = &TParObject{
						Str: EmptyStrObject(),
						Dlm: TpdPrompt,
					}
					tmpTp = tmpTp.Nxt
				}
			}
		} else {
			*tparam = nil
		}
	} else {
		*tparam = TparDuplicate(Lookup[ps.key].Tpar)
	}

	if *lookupCode != nil {
		(*lookupCode).Ref++
	}
	if !generate(ps, repSym, *repCount, command, *tparam, 0, *lookupCode) {
		return frame, false
	}
	*pc1 = ps.pc
	return frame, true
}

// CodeCompile compiles a span into executable code
func CodeCompile(frame *FrameObject, span *SpanObject, fromSpan bool) (*FrameObject, bool) {
	var ps parseState
	ps.status = ""
	ps.eoln = false
	ps.fromSpan = fromSpan

	defer func() {
		if ps.status != "" {
			ExitAbort = true
			ScreenMessage(ps.status)
		}
	}()

	if fromSpan {
		ps.startPoint = *span.MarkOne
		ps.endPoint = *span.MarkTwo
		ps.currentPoint = ps.startPoint
	}

	if span.Code != nil {
		CodeDiscard(&span.Code)
	}

	ps.codeBase = CodeTop
	ps.pc = 0
	ps.verifyCount = 0

	if !nextNonBl(&ps) {
		return frame, false
	}
	if ps.key == 0 {
		return errorMsg(frame, &ps, "Span contains no commands"), false
	}

	if fromSpan {
		for ps.key != 0 {
			var ok bool
			frame, ok = scanCommand(frame, &ps, true)
			if !ok {
				return frame, false
			}
		}
	} else {
		var ok bool
		frame, ok = scanCommand(frame, &ps, false)
		if !ok {
			return frame, false
		}
	}

	if !generate(&ps, LeadParamPInt, 1, CmdExitSuccess, nil, 0, nil) {
		return frame, false
	}

	// Fill in code header
	span.Code = &CodeHeader{
		Ref:   1,
		Code:  ps.codeBase + 1,
		Len:   ps.pc,
		FLink: CodeList.FLink,
		BLink: CodeList,
	}
	CodeList.FLink.BLink = span.Code
	CodeList.FLink = span.Code
	CodeTop = ps.codeBase + ps.pc
	return frame, true
}

type labelsType struct {
	exitLabel int
	failLabel int
	count     int
}

// CodeInterpret interprets compiled code
func CodeInterpret(frame *FrameObject, rept LeadParam, count int, codeHead *CodeHeader, fromSpan bool) (*FrameObject, bool) {
	const maxLevel = 100
	labels := make([]labelsType, maxLevel+1)

	request := TParObject{
		Nxt: nil,
		Con: nil,
	}

	codeHead.Ref++
	defer CodeDiscard(&codeHead)

	if rept == LeadParamPIndef {
		count = -1
	}

	const (
		success = iota
		failure
		failForever
	)
	interpStatus := success
	verifyAlways := InitialVerify

	for (count != 0) && (interpStatus == success) {
		count--
		level := 1
		labels[1].exitLabel = 0
		labels[1].failLabel = 0
		labels[1].count = 0
		pc := 1

		for pc != 0 {
			if pc > codeHead.Len {
				ScreenMessage(DbgPcOutOfRange)
				return frame, false
			}

			interpStatus = success
			cc := &CompilerCode[codeHead.Code-1+pc]
			currLbl := cc.Lbl
			currOp := cc.Op
			currRep := cc.Rep
			currCnt := cc.Cnt
			currTpar := cc.Tpar
			currCode := cc.Code
			pc++

			if isInterpCmd(currOp) {
				switch currOp {
				case CmdPcJump:
					pc = currLbl

				case CmdExitTo:
					fromSpan = true
					level++
					labels[level].exitLabel = currLbl
					labels[level].failLabel = 0
					labels[level].count = 0

				case CmdFailTo:
					labels[level].failLabel = currLbl

				case CmdIterate:
					if labels[level].count == currCnt {
						pc = labels[level].exitLabel
						level--
					} else {
						labels[level].count++
					}

				case CmdExitSuccess:
					if currRep == LeadParamPIndef {
						currCnt = level
					}
					if currCnt > 0 {
						if currCnt >= level {
							level = 0
						} else {
							level -= currCnt
						}
					}
					pc = labels[level+1].exitLabel

				case CmdExitFail:
					interpStatus = failure
					if currRep == LeadParamPIndef {
						currCnt = level
					}
					if currCnt > 0 {
						if currCnt >= level {
							level = 0
						} else {
							level -= currCnt
						}
					}
					pc = labels[level+1].failLabel

				case CmdExitAbort:
					ExitAbort = true
					interpStatus = failForever
					pc = 0

				case CmdExtended:
					if currCode == nil {
						ScreenMessage(DbgCodePtrIsNil)
						return frame, false
					}
					frame, _ = CodeInterpret(frame, currRep, currCnt, currCode, true)

				case CmdVerify:
					if !verifyAlways[currCnt] {
						if LudwigMode == LudwigBatch {
							ExitAbort = true
							interpStatus = failForever
							pc = 0
						} else if TparGet1(frame, currTpar, CmdVerify, &request) {
							if request.Len == 0 {
								request = frame.VerifyTpar
								if request.Len == 0 {
									ScreenMessage(MsgNoDefaultStr)
									return frame, false
								}
							} else {
								frame.VerifyTpar = request
							}
							if request.Str.Get(1) == 'Y' {
								// do nothing
							} else if request.Str.Get(1) == 'A' {
								verifyAlways[currCnt] = true
							} else if request.Str.Get(1) == 'Q' {
								ExitAbort = true
								interpStatus = failForever
								pc = 0
							} else {
								interpStatus = failure
								pc = currLbl
							}
						}
					}

				case CmdNoop:
					ScreenMessage(DbgIllegalInstruction)
					return frame, false
				}
			} else {
				// Call execute command
				var ok bool
				frame, ok = Execute(frame, currOp, currRep, currCnt, currTpar, fromSpan)
				if !ok {
					interpStatus = failure
					pc = currLbl
				}
				if ExitAbort {
					interpStatus = failForever
					pc = 0
				}
			}

			if TtControlC {
				interpStatus = failForever
				pc = 0
			}

			if interpStatus == failure {
				for pc == 0 && level >= 1 {
					pc = labels[level].failLabel
					level--
				}
			}
		}
	}

	return frame, (interpStatus == success)
}
