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

var interpCmds = map[Commands]bool{
	CmdPcJump:      true,
	CmdExitTo:      true,
	CmdFailTo:      true,
	CmdIterate:     true,
	CmdExitSuccess: true,
	CmdExitFail:    true,
	CmdExitAbort:   true,
	CmdExtended:    true,
	CmdVerify:      true,
	CmdNoop:        true,
}

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
			if CompilerCode[source].Tpar != nil {
				TparCleanObject(CompilerCode[source].Tpar)
			}
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

func errorMsg(ps *parseState, errText string) {
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
			if !FrameEdit(ps.currentPoint.Line.Group.Frame.Span.Name) {
				return
			}
			if CurrentFrame.Marks[MarkEquals-MinMarkNumber] != nil {
				MarkDestroy(&CurrentFrame.Marks[MarkEquals-MinMarkNumber])
			}
			var eLine *LineHdrObject
			if !LinesCreate(1, &eLine, &eLine) {
				return
			}

			var str StrObject
			str.Fill(' ', 1, MaxStrLen)
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
			if !LineChangeLength(eLine, i) {
				return
			}
			// "i" can't be zero here, so e_line->str != nullptr
			eLine.Str.Copy(&str, 1, i, 1)
			eLine.Used = str.Length(' ', i)
			if !LinesInject(eLine, eLine, ps.currentPoint.Line) {
				return
			}
			MarkCreate(eLine, ps.currentPoint.Col, &CurrentFrame.Dot)
		}
	}
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
l1:
	for {
		if !nextKey(ps) {
			return false
		}
		if ps.fromSpan {
			if (ps.key == '<') && (ps.currentPoint.Col <= ps.currentPoint.Line.Used) {
				if ps.currentPoint.Line.Str.Get(ps.currentPoint.Col) == '>' {
					ps.key = 0
				}
			}
		}
		if ps.key != ' ' {
			break
		}
	}
	if ps.key == '!' { // Comment - throw away rest of line
		if ps.fromSpan {
			ps.currentPoint.Col = ps.currentPoint.Line.Used + 1
			goto l1
		} else {
			ps.status = MsgCommentsIllegal
			return false
		}
	}
	return true
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

func getCount(ps *parseState, repCount *int) bool {
	const maxRepCount = 65535

	if ps.key >= '0' && ps.key <= '9' {
		*repCount = 0
		for {
			digit := ps.key - '0'
			if *repCount <= (maxRepCount-digit)/10 {
				*repCount = *repCount*10 + digit
			} else {
				errorMsg(ps, "Count too large")
				return false
			}
			if !nextKey(ps) {
				return false
			}
			if ps.key < '0' || ps.key > '9' {
				break
			}
		}
	} else {
		*repCount = 1
	}
	return true
}

func scanLeadingParam(ps *parseState, repSym *LeadParam, repCount *int) bool {
	switch ps.key {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		*repSym = LeadParamPInt
		if !getCount(ps, repCount) {
			return false
		}

	case '+':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamPlus
		*repCount = 1
		if ps.key >= '0' && ps.key <= '9' {
			*repSym = LeadParamPInt
			if !getCount(ps, repCount) {
				return false
			}
		}

	case '-':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamMinus
		*repCount = -1
		if ps.key >= '0' && ps.key <= '9' {
			*repSym = LeadParamNInt
			if !getCount(ps, repCount) {
				return false
			}
			*repCount = -*repCount
		}

	case '>', '.':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamPIndef
		*repCount = 0

	case '<', ',':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamNIndef
		*repCount = 0

	case '@':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamMarker
		if !getCount(ps, repCount) {
			return false
		}
		if (*repCount <= 0) || (*repCount > MaxUserMarkNumber) {
			errorMsg(ps, "Illegal mark number")
			return false
		}

	case '=':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamMarker
		*repCount = MarkEquals

	case '%':
		if !nextKey(ps) {
			return false
		}
		*repSym = LeadParamMarker
		*repCount = MarkModified

	default:
		*repSym = LeadParamNone
		*repCount = 1
	}
	return true
}

func scanTrailingParam(ps *parseState, command Commands, repSym LeadParam, tparam **TParObject) bool {
	tc := CmdAttrib[command].TpCount
	*tparam = nil

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
			return false
		}
		parDelim := ps.key
		if ps.key < 0 || ps.key > MaxSetRange || !ChIsPunctuation(rune(parDelim)) {
			errorMsg(ps, "Illegal parameter delimiter")
			return false
		}

		var tpl *TParObject
		for tci := 1; tci <= tc; tci++ {
			for {
				parLength := 0
				var parString StrObject
				for {
					if !nextKey(ps) {
						return false
					}
					if ps.key == 0 {
						errorMsg(ps, "Missing trailing delimiter")
						return false
					}
					parLength++
					parString.Set(parLength, byte(ps.key))
					if ps.eoln || ps.key == parDelim {
						break
					}
				}
				parLength--
				if ps.eoln && !CmdAttrib[command].TparInfo[tci].MlAllowed {
					errorMsg(ps, "Missing trailing delimiter")
					return false
				}

				tp := &TParObject{
					Len: parLength,
					Dlm: byte(parDelim),
					Str: parString,
					Nxt: nil,
					Con: nil,
				}

				if *tparam == nil {
					// 1st time through
					*tparam = tp
					tpl = tp
				} else {
					if tpl != nil {
						tpl.Con = tp
						tpl = tp
					} else {
						(*tparam).Nxt = tp
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
	return true
}

func scanCommand(ps *parseState, fullScan bool) bool {
	var repCount int
	var repSym LeadParam
	if !scanLeadingParam(ps, &repSym, &repCount) {
		return false
	}

	if ps.key >= 0 && ps.key <= MaxSetRange {
		ps.key = int(ChToUpper(byte(ps.key)))
	}

	command := Lookup[ps.key].Command
	for Prefixes.Bit(int(command)) != 0 {
		if !nextKey(ps) {
			return false
		}
		if ps.key < 0 {
			errorMsg(ps, "Command not valid")
			return false
		}
		i := LookupExpPtr[command]
		j := LookupExpPtr[command+1]
		for (i < j) && (int(ChToUpper(byte(ps.key))) != int(LookupExp[i].Extn)) {
			i++
		}
		if i < j {
			command = LookupExp[i].Command
		} else {
			errorMsg(ps, "Command not valid")
			return false
		}
	}

	var pc1 int
	if ps.key == '(' {
		var pc2, pc3 int
		if !scanCompoundCommand(ps, repSym, repCount, &pc1, &pc2, &pc3) {
			return false
		}
	} else if command != CmdNoop {
		var tparam *TParObject
		var lookupCode *CodeHeader
		if !scanSimpleCommand(ps, command, repSym, &repCount, &tparam, &lookupCode, &pc1, fullScan) {
			return false
		}
	} else {
		errorMsg(ps, "Command not valid")
		return false
	}

	if fullScan {
		var pc4 int
		if !scanExitHandler(ps, pc1, &pc4, fullScan) {
			return false
		}
	}
	return true
}

func scanExitHandler(ps *parseState, pc1 int, pc4 *int, fullScan bool) bool {
	if !nextNonBl(ps) {
		return false
	}
	if ps.key == '[' {
		if !nextNonBl(ps) {
			return false
		}
		for (ps.key != ':') && (ps.key != ']') {
			// Construct exit part
			if !scanCommand(ps, fullScan) {
				return false
			}
		}
		if ps.key == ':' {
			// Jump over fail handler
			if !generate(ps, LeadParamNone, 0, CmdPcJump, nil, 0, nil) {
				return false
			}
			*pc4 = ps.pc
			poke(ps.codeBase, pc1, ps.pc+1) // Set fail label for command
			if !nextNonBl(ps) {
				return false
			}
			for ps.key != ']' {
				// Construct fail part
				if !scanCommand(ps, fullScan) {
					return false
				}
			}
			poke(ps.codeBase, *pc4, ps.pc+1) // End of fail handler
		} else {
			poke(ps.codeBase, pc1, ps.pc+1) // Set fail label
		}
		if !nextNonBl(ps) {
			return false
		}
	}
	return true
}

func scanCompoundCommand(ps *parseState, repSym LeadParam, repCount int, pc1, pc2, pc3 *int) bool {
	if repSym != LeadParamNone && repSym != LeadParamPlus && repSym != LeadParamPInt &&
		repSym != LeadParamPIndef {
		errorMsg(ps, "Illegal leading parameter")
		return false
	}
	if !generate(ps, LeadParamNone, 0, CmdExitTo, nil, 0, nil) {
		return false
	}
	*pc2 = ps.pc
	if !generate(ps, LeadParamNone, 0, CmdFailTo, nil, 0, nil) {
		return false
	}
	*pc1 = ps.pc
	*pc3 = ps.pc + 1
	if repSym != LeadParamPIndef {
		if !generate(ps, LeadParamNone, repCount, CmdIterate, nil, 0, nil) {
			return false
		}
	}
	if !nextNonBl(ps) {
		return false
	}
	for ps.key != ')' {
		if !scanCommand(ps, true) {
			return false
		}
	}
	if !generate(ps, LeadParamNone, 0, CmdPcJump, nil, *pc3, nil) {
		return false
	}
	poke(ps.codeBase, *pc2, ps.pc+1) // Fill in exit label
	return true
}

func scanSimpleCommand(
	ps *parseState,
	command Commands,
	repSym LeadParam,
	repCount *int,
	tparam **TParObject,
	lookupCode **CodeHeader,
	pc1 *int,
	fullScan bool,
) bool {
	// Check if leading parameter is allowed
	lpAllowed := CmdAttrib[command].LpAllowed
	allowed := false
	// LpAllowed is a bitset stored as uint32
	if (lpAllowed & (1 << uint(repSym))) != 0 {
		allowed = true
	}
	if !allowed {
		errorMsg(ps, "Illegal leading parameter")
		return false
	}

	if command == CmdVerify {
		ps.verifyCount++
		if ps.verifyCount > MaxVerify {
			errorMsg(ps, "Too many verify commands in span")
			return false
		}
		*repCount = ps.verifyCount
	}

	*lookupCode = Lookup[ps.key].Code
	if Lookup[ps.key].Tpar == nil {
		if CmdAttrib[command].TpCount != 0 {
			if fullScan {
				if !scanTrailingParam(ps, command, repSym, tparam) {
					return false
				}
			} else {
				*tparam = &TParObject{
					Len: 0,
					Dlm: TpdPrompt,
					Nxt: nil,
					Con: nil,
				}
				tmpTp := *tparam
				for i := 2; i <= CmdAttrib[command].TpCount; i++ {
					tmpTp.Nxt = &TParObject{
						Len: 0,
						Dlm: TpdPrompt,
						Nxt: nil,
						Con: nil,
					}
					tmpTp = tmpTp.Nxt
				}
			}
		} else {
			*tparam = nil
		}
	} else {
		TparDuplicate(Lookup[ps.key].Tpar, tparam)
	}

	if *lookupCode != nil {
		(*lookupCode).Ref++
	}
	if !generate(ps, repSym, *repCount, command, *tparam, 0, *lookupCode) {
		return false
	}
	*pc1 = ps.pc
	return true
}

// CodeCompile compiles a span into executable code
func CodeCompile(span *SpanObject, fromSpan bool) bool {
	result := false
	var ps parseState
	ps.status = ""
	ps.eoln = false
	ps.fromSpan = fromSpan

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
		goto l99
	}
	if ps.key == 0 {
		errorMsg(&ps, "Span contains no commands")
		goto l99
	}

	if fromSpan {
		for ps.key != 0 {
			if !scanCommand(&ps, true) {
				goto l99
			}
		}
	} else if !scanCommand(&ps, false) {
		goto l99
	}

	if !generate(&ps, LeadParamPInt, 1, CmdExitSuccess, nil, 0, nil) {
		goto l99
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
	result = true

l99:
	if ps.status != "" {
		ExitAbort = true
		ScreenMessage(ps.status)
	}
	return result
}

type labelsType struct {
	exitLabel int
	failLabel int
	count     int
}

// CodeInterpret interprets compiled code
func CodeInterpret(rept LeadParam, count int, codeHead *CodeHeader, fromSpan bool) bool {
	const maxLevel = 100
	labels := make([]labelsType, maxLevel+1)

	result := false
	request := TParObject{
		Nxt: nil,
		Con: nil,
	}

	codeHead.Ref++

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
				goto l99
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

			if interpCmds[currOp] {
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
						goto l99
					}
					CodeInterpret(currRep, currCnt, currCode, true)

				case CmdVerify:
					if !verifyAlways[currCnt] {
						if LudwigMode == LudwigBatch {
							ExitAbort = true
							interpStatus = failForever
							pc = 0
						} else if TparGet1(currTpar, CmdVerify, &request) {
							if request.Len == 0 {
								request = CurrentFrame.VerifyTpar
								if request.Len == 0 {
									ScreenMessage(MsgNoDefaultStr)
									goto l99
								}
							} else {
								CurrentFrame.VerifyTpar = request
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
					goto l99
				}
			} else {
				// Call execute command
				if !Execute(currOp, currRep, currCnt, currTpar, fromSpan) {
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

	result = (interpStatus == success)
l99:
	TparCleanObject(&request)
	CodeDiscard(&codeHead)
	return result
}
