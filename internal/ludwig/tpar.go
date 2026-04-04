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

// tparDuplicateCon duplicates the con chain of a tpar
func tparDuplicateCon(tpar *TParObject, tpO *TParObject) {
	*tpO = *tpar
	tpO.Str = tpar.Str.Clone()
	tpO.Nxt = nil
	var tp2 *TParObject
	for tpar.Con != nil {
		tpar = tpar.Con
		tp := &TParObject{}
		*tp = *tpar
		tp.Str = tpar.Str.Clone()
		if tp2 == nil {
			tpO.Con = tp
		} else {
			tp2.Con = tp
		}
		tp2 = tp
	}
}

// TparDuplicate duplicates a trailing parameter list
func TparDuplicate(fromTp *TParObject) *TParObject {
	if fromTp != nil {
		toTp := &TParObject{}
		tparDuplicateCon(fromTp, toTp)
		fromTp = fromTp.Nxt
		tmpTp := toTp
		for fromTp != nil {
			tmpTp.Nxt = &TParObject{}
			tmpTp = tmpTp.Nxt
			tparDuplicateCon(fromTp, tmpTp)
			fromTp = fromTp.Nxt
		}
		return toTp
	} else {
		return nil
	}
}

// TparToMark converts a tpar string to a mark number
func TparToMark(strng *TParObject) (int, bool) {
	if strng.Len == 0 {
		ScreenMessage(MsgIllegalMarkNumber)
		return -1, false
	}
	mch := strng.Str.Get(1)
	if mch >= '0' && mch <= '9' {
		i := 1
		mark, found := TparToInt(strng, &i)
		if !found {
			return -1, false
		}
		if (i <= strng.Len) || ((mark < 1) || (mark > MaxUserMarkNumber)) {
			ScreenMessage(MsgIllegalMarkNumber)
			return -1, false
		}
		return mark, true
	} else {
		if (strng.Len > 1) || (mch != '=' && mch != '%') {
			ScreenMessage(MsgIllegalMarkNumber)
			return -1, false
		}
		if mch == '=' {
			return MarkEquals, true
		} else {
			return MarkModified, true
		}
	}
}

// TparToInt converts a tpar string to an integer
func TparToInt(strng *TParObject, chpos *int) (int, bool) {
	var ch byte
	if *chpos > strng.Len {
		ch = '\x00'
	} else {
		ch = strng.Str.Get(*chpos)
	}
	if ch < '0' || ch > '9' {
		ScreenMessage(MsgInvalidInteger)
		return 0, false
	}
	number := 0
	for {
		digit := int(ch - '0')
		if number <= (MaxInt-digit)/10 {
			number *= 10
			number += digit
		} else {
			ScreenMessage(MsgInvalidInteger)
			return 0, false
		}
		(*chpos)++
		if *chpos > strng.Len {
			ch = '\x00'
		} else {
			ch = strng.Str.Get(*chpos)
		}
		if !(ch >= '0' && ch <= '9') {
			break
		}
	}
	return number, true
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
			tpar.Str.FillCopy(startMark.Line.Str, startMark.Col, srclen, 1, tpar.Len, ' ')
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
				tmpTp2 := &TParObject{
					Str: NewBlankStrObject(MaxStrLen),
				}
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
			tmpTp2 := &TParObject{
				Str: NewBlankStrObject(MaxStrLen),
			}
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
			tmpTp.Str.FillCopy(endMark.Line.Str, 1, endMark.Line.Used, 1, tmpTp.Len, ' ')
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
func findEnquiry(frame *FrameObject, name string) (string, bool) {
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
		for _, r = range name[i:] {
			if variableType == varTypeOpsys {
				item.WriteRune(r)
			} else {
				item.WriteRune(unicode.ToUpper(r))
			}
		}
		itemStr = item.String()

		switch variableType {
		case varTypeTerminal:
			switch itemStr {
			case "NAME":
				return TerminalInfo.Name, true
			case "HEIGHT":
				return leftPadded(EnquiryNumLen, TerminalInfo.Height), true
			case "WIDTH":
				return leftPadded(EnquiryNumLen, TerminalInfo.Width), true
			case "SPEED":
				return leftPadded(EnquiryNumLen, 0), true
			}

		case varTypeFrame:
			if frame == nil {
				return "", false
			}
			switch itemStr {
			case "NAME":
				return frame.Span.Name, true
			case "INPUTFILE":
				if frame.InputFile == 0 {
					return "", true
				} else {
					return Files[frame.InputFile].Filename, true
				}
			case "OUTPUTFILE":
				if frame.OutputFile == 0 {
					return "", true
				} else {
					return Files[frame.OutputFile].Filename, true
				}
			case "MODIFIED":
				if frame.TextModified {
					return "Y", true
				} else {
					return "N", true
				}
			}

		case varTypeOpsys:
			val, ok := os.LookupEnv(itemStr)
			if len(val) > MaxStrLen {
				val = val[:MaxStrLen]
			}
			return val, ok

		case varTypeLudwig:
			switch itemStr {
			case "VERSION":
				return LudwigVersion, true
			case "OPSYS":
				return SystemName, true
			case "COMMAND_INTRODUCER":
				if !ChIsPrintable(rune(CommandIntroducer)) {
					ScreenMessage(MsgNonprintableIntroducer)
					return "", true
				} else {
					return string(rune(CommandIntroducer)), true
				}
			case "INSERT_MODE":
				if (EditMode == ModeInsert) || ((EditMode == ModeCommand) && (PreviousMode == ModeInsert)) {
					return "Y", true
				} else {
					return "N", true
				}
			case "OVERTYPE_MODE":
				if (EditMode == ModeOvertype) ||
					((EditMode == ModeCommand) && (PreviousMode == ModeOvertype)) {
					return "Y", true
				} else {
					return "N", true
				}
			}

		case varTypeUnknown:
			// Nothing to do
		}
	}
	return "", false
}

// tparEnquire performs environment enquiries
func tparEnquire(frame *FrameObject, tpar *TParObject) bool {
	tpar.Dlm = '\x00'
	name := tpar.Str.Slice(1, tpar.Len)
	if result, found := findEnquiry(frame, name); found {
		tpar.Str.Assign(result)
		tpar.Len = len(result)
		return true
	}
	ScreenMessage(MsgUnknownItem)
	ExitAbort = true
	return false
}

// TparAnalyse analyses and processes trailing parameters
func TparAnalyse(frame *FrameObject, cmd Commands, tran *TParObject, depth int, thisTp int) bool {
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
						if !TparAnalyse(frame, cmd, tran, depth+1, thisTp) {
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
						if !TparAnalyse(frame, cmd, tran, depth+1, thisTp) {
							return false
						}
					}
				}
			}
			switch delim {
			case TpdSpan:
				if !tparSubstitute(tran, cmd, thisTp) {
					return false
				}
			case TpdEnvironment:
				if FileData.OldCmds {
					ScreenMessage(MsgReservedTpd)
					return false
				} else {
					if !tparEnquire(frame, tran) {
						return false
					}
				}
			case TpdPrompt:
				if LudwigMode != LudwigBatch {
					if cmd == CmdVerify {
						var verifyReply VerifyResponse
						if tran.Len == 0 {
							prompt := DfltPrompts[CmdAttrib[cmd].TparInfo[thisTp].PromptName]
							verifyReply = ScreenVerify(&Screen, frame, prompt)
						} else {
							prompt := tran.Str.Slice(1, tran.Len)
							verifyReply = ScreenVerify(&Screen, frame, prompt)
						}
						switch verifyReply {
						case VerifyReplyYes:
							tran.Str = NewStrObjectFrom("Y")
						case VerifyReplyNo:
							tran.Str = NewStrObjectFrom("N")
						case VerifyReplyAlways:
							tran.Str = NewStrObjectFrom("A")
						case VerifyReplyQuit:
							tran.Str = NewStrObjectFrom("Q")
						}
						tran.Len = 1
					} else if tran.Len == 0 {
						prompt := DfltPrompts[CmdAttrib[cmd].TparInfo[thisTp].PromptName]
						tran.Str, tran.Len = ScreenGetLineP(&Screen, frame, prompt, CmdAttrib[cmd].TpCount, thisTp)
					} else {
						if tran.Con != nil {
							ScreenMessage(MsgPromptsAreOneLine)
							return false
						} else {
							prompt := tran.Str.Slice(1, tran.Len)
							tran.Str, tran.Len = ScreenGetLineP(&Screen, frame, prompt, CmdAttrib[cmd].TpCount, thisTp)
						}
					}
					tran.Dlm = '\x00'
				} else {
					ScreenMessage(MsgInteractiveModeOnly)
					return false
				}
			default:
				ended = true
			}
		}
	}
	return !TtControlC
}

// trim trims leading spaces and uppercases a tpar
func trim(request *TParObject) {
	if request.Len > 0 {
		originalLen := request.Len
		// Find first non-blank character
		i := 1
		for {
			if i > request.Len || request.Str.Get(i) != ' ' {
				break
			}
			i += 1
		}
		request.Len -= i - 1
		if request.Len > 0 {
			// Erase leading spaces
			request.Str.Erase(i-1, 1)
			request.Str.ApplyN(ChToUpper, request.Len, 1)
		}
		if request.Len < originalLen {
			request.Str.Fill(' ', request.Len+1, originalLen)
		}
	}
}

// TparGet1 gets and processes the first trailing parameter
func TparGet1(frame *FrameObject, tpar *TParObject, cmd Commands, tran *TParObject) bool {
	if tpar == nil {
		return false
	}
	tparDuplicateCon(tpar, tran)

	if TparAnalyse(frame, cmd, tran, 1, 1) {
		if CmdAttrib[cmd].TparInfo[1].TrimReply {
			trim(tran)
		}
		return true
	}
	return false
}

// TparGet2 gets and processes two trailing parameters
func TparGet2(frame *FrameObject, tpar *TParObject, cmd Commands, trn1 *TParObject, trn2 *TParObject) bool {
	if tpar == nil {
		return false
	}
	if tpar.Nxt == nil {
		return false
	}

	tparDuplicateCon(tpar, trn1)
	tparDuplicateCon(tpar.Nxt, trn2)

	if !TparAnalyse(frame, cmd, trn1, 1, 1) {
		return false
	}
	if trn1.Len != 0 {
		if !TparAnalyse(frame, cmd, trn2, 1, 2) {
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
