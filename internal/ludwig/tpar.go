/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         TPAR
//
// Description:  Tpar (Trailing PARameter) maintenance.

package ludwig

import (
	"fmt"
	"os"
	"strings"
	"unicode"
)

const (
	EnquiryNumLen = 20
	SystemName    = "Go/Linux"
)

type varType int

const (
	varTypeUnknown varType = iota
	varTypeTerminal
	varTypeFrame
	varTypeOpsys
	varTypeLudwig
)

// discardTp recursively discards trailing parameter objects
func discardTp(tp *TParObject) {
	if tp.Nxt != nil {
		discardTp(tp.Nxt)
	} else if tp.Con != nil {
		discardTp(tp.Con)
	}
	// In Go, we don't explicitly delete, garbage collector handles it
	tp.Nxt = nil
	tp.Con = nil
}

// TparCleanObject cleans up a trailing parameter object
func TparCleanObject(tpO *TParObject) {
	if tpO.Con != nil {
		discardTp(tpO.Con)
	}
	if tpO.Nxt != nil {
		discardTp(tpO.Nxt)
	}
	tpO.Con = nil
	tpO.Nxt = nil
}

// tparDuplicateCon duplicates the con chain of a tpar
func tparDuplicateCon(tpar *TParObject, tpO *TParObject) {
	*tpO = *tpar
	tpO.Str = *tpar.Str.Clone()
	tpO.Nxt = nil
	var tp2 *TParObject
	for tpar.Con != nil {
		tpar = tpar.Con
		tp := &TParObject{}
		*tp = *tpar
		tp.Str = *tpar.Str.Clone()
		if tp2 == nil {
			tpO.Con = tp
		} else {
			tp2.Con = tp
		}
		tp2 = tp
	}
}

// TparDuplicate duplicates a trailing parameter list
func TparDuplicate(fromTp *TParObject, toTp **TParObject) {
	if fromTp != nil {
		*toTp = &TParObject{}
		tparDuplicateCon(fromTp, *toTp)
		fromTp = fromTp.Nxt
		tmpTp := *toTp
		for fromTp != nil {
			tmpTp.Nxt = &TParObject{}
			tmpTp = tmpTp.Nxt
			tparDuplicateCon(fromTp, tmpTp)
			fromTp = fromTp.Nxt
		}
	} else {
		*toTp = nil
	}
}

// TparToMark converts a tpar string to a mark number
func TparToMark(strng *TParObject, mark *int) bool {
	if strng.Len == 0 {
		ScreenMessage(MsgIllegalMarkNumber)
		return false
	}
	mch := strng.Str.Get(1)
	if mch >= '0' && mch <= '9' {
		i := 1
		if !TparToInt(strng, &i, mark) {
			return false
		}
		if (i <= strng.Len) || ((*mark < 1) || (*mark > MaxUserMarkNumber)) {
			ScreenMessage(MsgIllegalMarkNumber)
			return false
		}
	} else {
		if (strng.Len > 1) || (mch != '=' && mch != '%') {
			ScreenMessage(MsgIllegalMarkNumber)
			return false
		}
		if mch == '=' {
			*mark = MarkEquals
		} else {
			*mark = MarkModified
		}
	}
	return true
}

// TparToInt converts a tpar string to an integer
func TparToInt(strng *TParObject, chpos *int, intVal *int) bool {
	var ch byte
	if *chpos > strng.Len {
		ch = '\x00'
	} else {
		ch = strng.Str.Get(*chpos)
	}
	if ch < '0' || ch > '9' {
		ScreenMessage(MsgInvalidInteger)
		return false
	}
	number := 0
	for {
		digit := int(ch - '0')
		if number <= (MaxInt-digit)/10 {
			number *= 10
			number += digit
		} else {
			ScreenMessage(MsgInvalidInteger)
			return false
		}
		*chpos++
		if *chpos > strng.Len {
			ch = '\x00'
		} else {
			ch = strng.Str.Get(*chpos)
		}
		if !(ch >= '0' && ch <= '9') {
			break
		}
	}
	*intVal = number
	return true
}

// tparSubstitute substitutes span content into tpar
func tparSubstitute(tpar *TParObject, cmd Commands, thisTp int) bool {
	if tpar.Con != nil {
		ScreenMessage(MsgSpanNamesAreOneLine)
		return false
	}
	// Get the span name and uppercase it
	name := strings.ToUpper(string(tpar.Str.Slice(1, tpar.Len)))
	var span *SpanObject
	var dummy *SpanObject
	if SpanFind(name, &span, &dummy) {
		tpar.Dlm = '\x00'
		startMark := *span.MarkOne
		endMark := *span.MarkTwo
		if startMark.Line == endMark.Line {
			tpar.Len = endMark.Col - startMark.Col
			var srclen int
			if startMark.Col > startMark.Line.Used {
				srclen = 0
			} else if endMark.Col > endMark.Line.Used {
				srclen = endMark.Line.Used - startMark.Col + 1
			} else {
				srclen = tpar.Len
			}
			ChFillCopy(startMark.Line.Str, startMark.Col, srclen, &tpar.Str, 1, tpar.Len, ' ')
		} else if !CmdAttrib[cmd].TparInfo[thisTp].MlAllowed {
			ScreenMessage(MsgSpanMustBeOneLine)
			return false
		} else {
			// Copy entire span into a tpar
			if startMark.Col > startMark.Line.Used {
				tpar.Len = 0
			} else {
				tpar.Len = startMark.Line.Used - startMark.Col + 1
			}
			tpar.Str.Copy(startMark.Line.Str, startMark.Col, tpar.Len, 1)
			// Anything between the start and end marks?
			var tmpTp *TParObject
			startMark.Line = startMark.Line.FLink
			for startMark.Line != endMark.Line {
				tmpTp2 := &TParObject{Str: *NewFilled(' ', MaxStrLen)}
				if tmpTp == nil {
					tpar.Con = tmpTp2
				} else {
					tmpTp.Con = tmpTp2
				}
				tmpTp = tmpTp2
				tmpTp.Dlm = '\x00'
				tmpTp.Nxt = nil
				tmpTp.Con = nil
				tmpTp.Len = startMark.Line.Used
				tmpTp.Str.Copy(startMark.Line.Str, 1, tmpTp.Len, 1)
				startMark.Line = startMark.Line.FLink
			}
			// Create new tpar for last line
			tmpTp2 := &TParObject{Str: *NewFilled(' ', MaxStrLen)}
			if tmpTp == nil {
				tpar.Con = tmpTp2
			} else {
				tmpTp.Con = tmpTp2
			}
			tmpTp = tmpTp2
			tmpTp.Dlm = '\x00'
			tmpTp.Nxt = nil
			tmpTp.Con = nil
			tmpTp.Len = endMark.Col - 1
			ChFillCopy(endMark.Line.Str, 1, endMark.Line.Used, &tmpTp.Str, 1, tpar.Len, ' ')
		}
	} else {
		ScreenMessage(MsgNoSuchSpan)
		return false
	}
	return true
}

// leftPadded returns a string padded on the left to the specified width
func leftPadded(width int, value int) string {
	return fmt.Sprintf("%*d", width, value)
}

// findEnquiry handles environment and system enquiries
func findEnquiry(name string, result *StrObject, reslen *int) bool {
	enquiryResult := false
	variableType := varTypeUnknown
	var item strings.Builder
	length := len(name)

	i := 0
	var r rune
	for i, r = range name {
		if r == '-' {
			break
		}
		item.WriteRune(unicode.ToUpper(r))
	}
	if i < length && r == '-' {
		i++
		itemStr := item.String()
		switch itemStr {
		case "TERMINAL":
			variableType = varTypeTerminal
		case "FRAME":
			variableType = varTypeFrame
		case "ENV":
			variableType = varTypeOpsys
		case "LUDWIG":
			variableType = varTypeLudwig
		}
		item.Reset()
		for i, r = range name[i:] {
			if variableType == varTypeOpsys {
				item.WriteRune(r)
			} else {
				item.WriteRune(unicode.ToUpper(r))
			}
		}
		itemStr = item.String()

		switch variableType {
		case varTypeTerminal:
			enquiryResult = true
			switch itemStr {
			case "NAME":
				*reslen = len(TerminalInfo.Name)
				result.Assign(TerminalInfo.Name)
			case "HEIGHT":
				s := leftPadded(EnquiryNumLen, TerminalInfo.Height)
				*reslen = len(s)
				result.Assign(s)
			case "WIDTH":
				s := leftPadded(EnquiryNumLen, TerminalInfo.Width)
				*reslen = len(s)
				result.Assign(s)
			case "SPEED":
				s := leftPadded(EnquiryNumLen, 0)
				*reslen = len(s)
				result.Assign(s)
			default:
				enquiryResult = false
			}

		case varTypeFrame:
			enquiryResult = true
			switch itemStr {
			case "NAME":
				*reslen = len(CurrentFrame.Span.Name)
				result.Assign(CurrentFrame.Span.Name)
			case "INPUTFILE":
				if CurrentFrame.InputFile == 0 {
					*reslen = 0
				} else {
					*reslen = len(Files[CurrentFrame.InputFile].Filename)
					result.Assign(Files[CurrentFrame.InputFile].Filename)
				}
			case "OUTPUTFILE":
				if CurrentFrame.OutputFile == 0 {
					*reslen = 0
				} else {
					*reslen = len(Files[CurrentFrame.OutputFile].Filename)
					result.Assign(Files[CurrentFrame.OutputFile].Filename)
				}
			case "MODIFIED":
				*reslen = 1
				result.Fill(' ', 1, MaxStrLen)
				if CurrentFrame.TextModified {
					result.Set(1, 'Y')
				} else {
					result.Set(1, 'N')
				}
			default:
				enquiryResult = false
			}

		case varTypeOpsys:
			env, found := os.LookupEnv(itemStr)
			if found {
				enquiryResult = true
				*reslen = min(MaxStrLen, len(env))
				result.Assign(env)
			}

		case varTypeLudwig:
			enquiryResult = true
			switch itemStr {
			case "VERSION":
				*reslen = len(LudwigVersion)
				result.Assign(LudwigVersion)
			case "OPSYS":
				result.Assign(SystemName)
				*reslen = len(SystemName)
			case "COMMAND_INTRODUCER":
				if !ChIsPrintable(rune(CommandIntroducer)) {
					*reslen = 0
					ScreenMessage(MsgNonprintableIntroducer)
				} else {
					*reslen = 1
					result.Set(1, byte(CommandIntroducer))
				}
			case "INSERT_MODE":
				*reslen = 1
				if (EditMode == ModeInsert) || ((EditMode == ModeCommand) && (PreviousMode == ModeInsert)) {
					result.Set(1, 'Y')
				} else {
					result.Set(1, 'N')
				}
			case "OVERTYPE_MODE":
				*reslen = 1
				if (EditMode == ModeOvertype) ||
					((EditMode == ModeCommand) && (PreviousMode == ModeOvertype)) {
					result.Set(1, 'Y')
				} else {
					result.Set(1, 'N')
				}
			default:
				enquiryResult = false
			}

		case varTypeUnknown:
			// Nothing to do
		}
	}
	return enquiryResult
}

// tparEnquire performs environment enquiries
func tparEnquire(tpar *TParObject) bool {
	tpar.Dlm = '\x00'
	name := tpar.Str.Slice(1, tpar.Len)
	if findEnquiry(name, &tpar.Str, &tpar.Len) {
		return true
	}
	ScreenMessage(MsgUnknownItem)
	ExitAbort = true
	return false
}

// TparAnalyse analyses and processes trailing parameters
func TparAnalyse(cmd Commands, tran *TParObject, depth int, thisTp int) bool {
	if depth > MaxTparRecursion {
		ScreenMessage(MsgTparTooDeep)
		return false
	}
	if tran.Dlm != TpdSmart && tran.Dlm != TpdExact && tran.Dlm != TpdLit {
		ended := false
		for !ended && !TtControlC {
			delim := tran.Dlm // Save copy of delimiter
			if tran.Con == nil {
				if tran.Len > 1 {
					ts1 := tran.Str.Get(1)
					if (ts1 == tran.Str.Get(tran.Len)) &&
						(ts1 == TpdSpan || ts1 == TpdPrompt || ts1 == TpdEnvironment ||
							ts1 == TpdSmart || ts1 == TpdExact || ts1 == TpdLit) {
						// Nested delimiters
						tran.Dlm = ts1
						tran.Len -= 2
						// Erase first char
						tran.Str.Erase(1, 1)
						if !TparAnalyse(cmd, tran, depth+1, thisTp) {
							return false
						}
					}
				}
			} else {
				tmpTp := tran.Con
				for tmpTp.Con != nil {
					tmpTp = tmpTp.Con
				}
				if (tran.Len != 0) && (tmpTp.Len != 0) {
					ts1 := tran.Str.Get(1)
					if (ts1 == tmpTp.Str.Get(tmpTp.Len)) &&
						(ts1 == TpdSpan || ts1 == TpdPrompt || ts1 == TpdEnvironment ||
							ts1 == TpdSmart || ts1 == TpdExact || ts1 == TpdLit) {
						// Nested delimiters
						tran.Dlm = ts1
						tran.Len--
						tran.Str.Erase(1, 1)
						tmpTp.Len--
						if !TparAnalyse(cmd, tran, depth+1, thisTp) {
							return false
						}
					}
				}
			}
			if delim == TpdSpan {
				if !tparSubstitute(tran, cmd, thisTp) {
					return false
				}
			} else if delim == TpdEnvironment {
				if FileData.OldCmds {
					ScreenMessage(MsgReservedTpd)
					return false
				} else {
					if !tparEnquire(tran) {
						return false
					}
				}
			} else if delim == TpdPrompt {
				if LudwigMode != LudwigBatch {
					if cmd == CmdVerify {
						var verifyReply VerifyResponse
						if tran.Len == 0 {
							prompt := DfltPrompts[CmdAttrib[cmd].TparInfo[thisTp].PromptName]
							verifyReply = ScreenVerify(prompt)
						} else {
							prompt := tran.Str.Slice(1, tran.Len)
							verifyReply = ScreenVerify(prompt)
						}
						switch verifyReply {
						case VerifyReplyYes:
							tran.Str.Set(1, 'Y')
						case VerifyReplyNo:
							tran.Str.Set(1, 'N')
						case VerifyReplyAlways:
							tran.Str.Set(1, 'A')
						case VerifyReplyQuit:
							tran.Str.Set(1, 'Q')
						}
						tran.Len = 1
					} else if tran.Len == 0 {
						prompt := DfltPrompts[CmdAttrib[cmd].TparInfo[thisTp].PromptName]
						ScreenGetLineP(prompt, &tran.Str, &tran.Len, CmdAttrib[cmd].TpCount, thisTp)
					} else {
						if tran.Con != nil {
							ScreenMessage(MsgPromptsAreOneLine)
							return false
						} else {
							prompt := tran.Str.Slice(1, tran.Len)
							ScreenGetLineP(prompt, &tran.Str, &tran.Len, CmdAttrib[cmd].TpCount, thisTp)
						}
					}
					tran.Dlm = '\x00'
				} else {
					ScreenMessage(MsgInteractiveModeOnly)
					return false
				}
			} else {
				ended = true
			}
		}
	}
	return !TtControlC
}

// trim trims leading spaces and uppercases a tpar
func trim(request *TParObject) {
	if request.Len > 0 {
		// Find first non-blank character
		i := 0
		for {
			i++
			if !(request.Str.Get(i) == ' ' && i != request.Len) {
				break
			}
		}
		request.Len -= i - 1
		if request.Len > 0 {
			// Erase leading spaces
			request.Str.Erase(i-1, 1)
			request.Str.ApplyN(ChToUpper, request.Len, 1)
		}
		if request.Len >= 0 && request.Len < MaxStrLen {
			request.Str.Fill(' ', request.Len+1, MaxStrLen)
		}
	}
}

// TparGet1 gets and processes the first trailing parameter
func TparGet1(tpar *TParObject, cmd Commands, tran *TParObject) bool {
	if tpar == nil {
		return false
	}
	tparDuplicateCon(tpar, tran)

	if TparAnalyse(cmd, tran, 1, 1) {
		if CmdAttrib[cmd].TparInfo[1].TrimReply {
			trim(tran)
		}
		return true
	}
	return false
}

// TparGet2 gets and processes two trailing parameters
func TparGet2(tpar *TParObject, cmd Commands, trn1 *TParObject, trn2 *TParObject) bool {
	if tpar == nil {
		return false
	}
	if tpar.Nxt == nil {
		return false
	}

	tparDuplicateCon(tpar, trn1)
	tparDuplicateCon(tpar.Nxt, trn2)

	if !TparAnalyse(cmd, trn1, 1, 1) {
		return false
	}
	if trn1.Len != 0 {
		if !TparAnalyse(cmd, trn2, 1, 2) {
			return false
		}
	}
	if CmdAttrib[cmd].TparInfo[1].TrimReply {
		trim(trn1)
	}
	if CmdAttrib[cmd].TparInfo[2].TrimReply {
		trim(trn2)
	}
	return true
}
