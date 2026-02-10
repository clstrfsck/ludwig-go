/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         EQSGETREP
//
// Description:  The EQS, GET and REPLACE commands.

package ludwig

import (
	"math"
)

const (
	thisOne        = "This one?"
	replaceThisOne = "Replace this one?"
)

// eqsgetrepExactcase converts a target to exact case
func eqsgetrepExactcase(target *TParObject) bool {
	if target.Dlm != '"' {
		target.Str.ApplyN(ChToUpper, target.Len, 1)
		return false
	}
	return true
}

func eqsgetrepSamePatternDef(pattern1 *PatternDefType, pattern2 *PatternDefType) bool {
	if pattern1.Length != 0 && pattern2.Length != 0 && pattern1.Length == pattern2.Length {
		for count := 1; count <= pattern1.Length; count++ {
			if pattern1.Strng.Get(count) != pattern2.Strng.Get(count) {
				return false
			}
		}
		return true
	}
	return false
}

func eqsgetrepPatternBuild(tpar TParObject, patternPtr **DFATableObject) bool {
	patternDefinition := PatternDefType{Strng: *NewBlankStrObject(MaxStrLen)}
	var nfaTable NFATableType
	var firstPatternStart int
	var patternFinalState int
	var leftContextEnd int
	var middleContextEnd int
	var statesUsed int

	if PatternParser(
		&tpar,
		&nfaTable,
		&firstPatternStart,
		&patternFinalState,
		&leftContextEnd,
		&middleContextEnd,
		&patternDefinition,
		&statesUsed,
	) {
		var alreadyBuilt bool
		if *patternPtr != nil {
			alreadyBuilt = eqsgetrepSamePatternDef(&patternDefinition, &(*patternPtr).Definition)
		} else {
			alreadyBuilt = false
		}
		if !alreadyBuilt {
			if !PatternDFATableInitialize(patternPtr, patternDefinition) {
				return false
			}
			var dfaStart, dfaEnd int
			if !PatternDFAConvert(
				&nfaTable,
				*patternPtr,
				firstPatternStart,
				&patternFinalState,
				leftContextEnd,
				middleContextEnd,
				&dfaStart,
				&dfaEnd,
			) {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func EqsGetRepEqs(rept LeadParam, tpar TParObject) bool {
	success := false

	if tpar.Dlm == TpdSmart {
		if !eqsgetrepPatternBuild(tpar, &CurrentFrame.EqsPatternPtr) {
			return false
		}
		markFlag := false
		var startCol int
		var endPos int
		found := PatternRecognize(
			CurrentFrame.EqsPatternPtr,
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&markFlag,
			&startCol,
			&endPos,
		)
		switch rept {
		case LeadParamNone, LeadParamPlus:
			success = (CurrentFrame.Dot.Col == startCol) && found
		case LeadParamMinus:
			success = !((CurrentFrame.Dot.Col == startCol) && found)
		case LeadParamPIndef:
			success = (endPos <= CurrentFrame.Dot.Line.Used) && found
		case LeadParamNIndef:
			success = (endPos >= CurrentFrame.Dot.Line.Used) && found
		}
		if success && rept != LeadParamMinus {
			success = MarkCreate(CurrentFrame.Dot.Line, endPos, &CurrentFrame.Marks[MarkEquals-MinMarkNumber])
		}
	} else {
		exactcase := eqsgetrepExactcase(&tpar)
		startCol := CurrentFrame.Dot.Col
		var length int
		if startCol > CurrentFrame.Dot.Line.Used {
			length = 0
			startCol = 1
		} else {
			length = CurrentFrame.Dot.Line.Used + 1 - CurrentFrame.Dot.Col
		}
		if length > tpar.Len {
			length = tpar.Len
		}

		var nchIdent int
		result := ChCompareStr(
			tpar.Str,
			1,
			tpar.Len,
			CurrentFrame.Dot.Line.Str,
			startCol,
			length,
			exactcase,
			&nchIdent,
		)
		switch rept {
		case LeadParamNone, LeadParamPlus:
			success = (result == 0)
		case LeadParamMinus:
			success = (result != 0)
		case LeadParamPIndef:
			success = (result <= 0)
		case LeadParamNIndef:
			success = (result >= 0)
		}
		if success && rept != LeadParamMinus {
			success = MarkCreate(
				CurrentFrame.Dot.Line,
				CurrentFrame.Dot.Col+nchIdent,
				&CurrentFrame.Marks[MarkEquals-MinMarkNumber],
			)
		}
	}
	return success
}

func eqsgetrepDumbGet(count int, tpar TParObject, fromSpan bool) bool {
	result := (count == 0)

	dotLine := CurrentFrame.Dot.Line
	dotCol := CurrentFrame.Dot.Col
	exactcase := eqsgetrepExactcase(&tpar)
	line := dotLine
	newlen := tpar.Len
	var tailSpace bool
	if newlen > 1 && tpar.Str.Get(newlen) == ' ' {
		tailSpace = true
		newlen--
	} else {
		tailSpace = false
	}
	newstr := NewBlankStrObject(MaxStrLen)
	var backwards bool
	var startCol int
	var length int
	if count < 0 {
		count = -count
		ChReverseStr(tpar.Str, newstr, newlen)
		backwards = true
		startCol = 1
		length = CurrentFrame.Dot.Col - 1
		if length > line.Used {
			length = line.Used
		}
	} else {
		newstr = tpar.Str
		backwards = false
		startCol = CurrentFrame.Dot.Col
		if startCol > line.Used {
			length = 0
		} else {
			length = line.Used + 1 - startCol
		}
	}

	for count > 0 && !TtControlC {
		var found bool
		var offset int
		if length == 0 {
			found = false
		} else {
			found = ChSearchStr(
				newstr,
				1,
				newlen,
				line.Str,
				startCol,
				length,
				exactcase,
				backwards,
				&offset,
			)
		}
		if found {
			if tailSpace {
				var tailChar byte
				if startCol+offset+newlen <= line.Used {
					tailChar = line.Str.Get(startCol + offset + newlen)
				} else if startCol+offset+newlen == line.Used+1 {
					if line.Used+1 == MaxStrLenP {
						tailChar = 0
					} else {
						tailChar = ' '
					}
				} else {
					tailChar = 0
				}
				if tailChar != ' ' {
					if backwards {
						startCol = startCol + offset + newlen - 1
					} else {
						startCol++
					}
					goto l2
				}
			}
			startCol += offset
			if !backwards {
				startCol += tpar.Len
			}
			count--
			if count == 0 {
				if !MarkCreate(line, startCol, &CurrentFrame.Dot) {
					goto l99
				}
				if !fromSpan {
					switch ScreenVerify(thisOne) {
					case VerifyReplyAlways, VerifyReplyYes:
						break
					case VerifyReplyQuit, VerifyReplyNo:
						count = 1
						if !MarkCreate(dotLine, dotCol, &CurrentFrame.Dot) {
							goto l99
						}
						if ExitAbort {
							goto l99
						} else {
							goto l1
						}
					}
				}
				if backwards {
					if !MarkCreate(line, startCol+tpar.Len, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
						goto l99
					}
				} else {
					if !MarkCreate(line, startCol-tpar.Len, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
						goto l99
					}
				}
				result = true
				goto l99
			l1:
			}
		l2:
			if backwards {
				length = startCol - 1
				startCol = 1
			} else if startCol > line.Used {
				length = 0
			} else {
				length = line.Used + 1 - startCol
			}
		} else {
			if backwards {
				line = line.BLink
			} else {
				line = line.FLink
			}
			if line == nil {
				goto l99
			}
			startCol = 1
			length = line.Used
		}
	}
l99:
	return result
}

func eqsgetrepPatternGet(count int, tpar TParObject, fromSpan bool, replaceFlag bool) bool {
	result := (count == 0)

	var patternPtr *DFATableObject
	if !replaceFlag {
		if !eqsgetrepPatternBuild(tpar, &CurrentFrame.GetPatternPtr) {
			return result
		}
		patternPtr = CurrentFrame.GetPatternPtr
	} else {
		patternPtr = CurrentFrame.RepPatternPtr
	}

	dotLine := CurrentFrame.Dot.Line
	dotCol := CurrentFrame.Dot.Col
	line := dotLine
	markFlag := false
	backwards := count < 0
	var startCol int
	if backwards {
		startCol = 1
	} else {
		startCol = dotCol
	}
	count = int(math.Abs(float64(count)))
	if startCol > line.Used {
		startCol = line.Used + 1
	}

	for count > 0 && !TtControlC {
		var matchedStartCol int
		var matchedFinishCol int
		if PatternRecognize(
			patternPtr,
			line,
			startCol,
			&markFlag,
			&matchedStartCol,
			&matchedFinishCol,
		) {
			if !((line == dotLine) && (matchedFinishCol >= dotCol) && backwards) {
				count--
				if count == 0 {
					if backwards {
						if !MarkCreate(line, matchedStartCol, &CurrentFrame.Dot) {
							goto l99
						}
					} else {
						if !MarkCreate(line, matchedFinishCol, &CurrentFrame.Dot) {
							goto l99
						}
					}
					if !fromSpan {
						switch ScreenVerify(thisOne) {
						case VerifyReplyAlways, VerifyReplyYes:
							break
						case VerifyReplyQuit, VerifyReplyNo:
							count = 1
							if !MarkCreate(dotLine, dotCol, &CurrentFrame.Dot) {
								goto l99
							}
							if ExitAbort {
								goto l99
							} else {
								goto l1
							}
						}
					}
					if backwards {
						if !MarkCreate(line, matchedFinishCol, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
							goto l99
						}
					} else if !MarkCreate(line, matchedStartCol, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
						goto l99
					}
					result = true
					goto l99
				l1:
				}
				startCol = matchedFinishCol
				if startCol == matchedStartCol {
					markFlag = true
				}
				if startCol > line.Used {
					if backwards {
						line = line.BLink
					} else {
						line = line.FLink
					}
					if line == nil {
						goto l99
					}
					markFlag = false
					startCol = 1
				}
			} else {
				line = line.BLink
				if line == nil {
					goto l99
				}
				markFlag = false
				startCol = 1
			}
		} else {
			if backwards {
				line = line.BLink
			} else {
				line = line.FLink
			}
			if line == nil {
				goto l99
			}
			markFlag = false
			startCol = 1
		}
	}

l99:
	return result
}

func EqsGetRepGet(count int, tpar TParObject, fromSpan bool) bool {
	if tpar.Dlm == TpdSmart {
		return eqsgetrepPatternGet(count, tpar, fromSpan, false)
	}
	return eqsgetrepDumbGet(count, tpar, fromSpan)
}

func EqsGetRepRep(rept LeadParam, count int, tpar TParObject, tpar2 TParObject, fromSpan bool) bool {
	var getcount int
	var length int
	var delta int
	var startCol int
	var oldDot *MarkObject
	var oldEquals *MarkObject
	var okay bool
	result := false

	if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &oldDot) {
		goto l99
	}
	if CurrentFrame.Marks[MarkEquals-MinMarkNumber] != nil {
		if !MarkCreate(
			CurrentFrame.Marks[MarkEquals-MinMarkNumber].Line,
			CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col,
			&oldEquals,
		) {
			goto l99
		}
	}
	if tpar.Dlm == TpdSmart {
		if !eqsgetrepPatternBuild(tpar, &CurrentFrame.RepPatternPtr) {
			goto l99
		}
	}
	getcount = 1
	if rept == LeadParamMinus || rept == LeadParamNIndef || rept == LeadParamNInt {
		getcount = -1
	}
	if rept == LeadParamPIndef || rept == LeadParamNIndef {
		count = MaxInt
	} else if count < 0 {
		count = -count
	}
	for count > 0 {
		for {
			okay = true
			if TtControlC || ExitAbort {
				goto l1
			}
			if tpar.Dlm == TpdSmart {
				if !eqsgetrepPatternGet(getcount, tpar, true, true) {
					goto l1
				}
			} else if !eqsgetrepDumbGet(getcount, tpar, true) {
				goto l1
			}
			if TtControlC || ExitAbort {
				goto l1
			}
			if !fromSpan {
				switch ScreenVerify(replaceThisOne) {
				case VerifyReplyAlways:
					fromSpan = true
				case VerifyReplyYes:
					break
				case VerifyReplyQuit, VerifyReplyNo:
					okay = false
				}
			}
			if okay {
				break
			}
		}
		length = CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col - CurrentFrame.Dot.Col
		if length < 0 {
			CurrentFrame.Dot.Col = CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col
			CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col = CurrentFrame.Dot.Col - length
			length = -length
		}
		if tpar2.Con == nil {
			startCol = CurrentFrame.Dot.Col
			delta = length - tpar2.Len
			if delta > 0 {
				if CurrentFrame.Dot.Col+delta > CurrentFrame.Dot.Line.Used+1 {
					delta = CurrentFrame.Dot.Line.Used + 1 - CurrentFrame.Dot.Col
				}
				if delta > 0 {
					if !CharcmdDelete(CmdDeleteChar, LeadParamPInt, delta, true) {
						goto l99
					}
				}
			} else if delta < 0 {
				if !CharcmdInsert(CmdInsertChar, LeadParamPInt, -delta, true) {
					goto l99
				}
				CurrentFrame.Dot.Col = startCol
			}
			if !TextOvertype(true, 1, tpar2.Str, tpar2.Len, CurrentFrame.Dot) {
				goto l99
			}
			if getcount > 0 {
				if !MarkCreate(
					CurrentFrame.Dot.Line, startCol, &CurrentFrame.Marks[MarkEquals-MinMarkNumber],
				) {
					goto l99
				}
			} else {
				if !MarkCreate(
					CurrentFrame.Dot.Line,
					startCol+tpar2.Len,
					&CurrentFrame.Marks[MarkEquals-MinMarkNumber],
				) {
					goto l99
				}
				CurrentFrame.Dot.Col = startCol
			}
		} else {
			if !CharcmdDelete(CmdDeleteChar, LeadParamPInt, length, true) {
				goto l99
			}
			if !TextInsertTpar(&tpar2, CurrentFrame.Dot, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
				goto l99
			}
		}
		if !MarkCreate(CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &oldDot) {
			goto l99
		}
		if !MarkCreate(
			CurrentFrame.Marks[MarkEquals-MinMarkNumber].Line,
			CurrentFrame.Marks[MarkEquals-MinMarkNumber].Col,
			&oldEquals,
		) {
			goto l99
		}
		CurrentFrame.TextModified = true
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkModified-MinMarkNumber],
		) {
			goto l99
		}
		count--
	}
l1:
	if !MarkCreate(oldDot.Line, oldDot.Col, &CurrentFrame.Dot) {
		goto l99
	}
	if !MarkDestroy(&oldDot) {
		goto l99
	}
	if oldEquals != nil {
		if !MarkCreate(oldEquals.Line, oldEquals.Col, &CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
			goto l99
		}
		if !MarkDestroy(&oldEquals) {
			goto l99
		}
	} else if CurrentFrame.Marks[MarkEquals-MinMarkNumber] != nil {
		if !MarkDestroy(&CurrentFrame.Marks[MarkEquals-MinMarkNumber]) {
			goto l99
		}
	}
	result = (count == 0) || rept == LeadParamPIndef || rept == LeadParamNIndef
l99:
	if oldDot != nil {
		MarkDestroy(&oldDot)
	}
	if oldEquals != nil {
		MarkDestroy(&oldEquals)
	}
	return result
}
