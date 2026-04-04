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

func eqsgetrepPatternBuild(frame *FrameObject, tpar TParObject, patternPtr **DFATableObject) bool {
	patternDefinition := PatternDefType{Strng: *NewBlankStrObject(MaxStrLen)}
	var nfaTable NFATableType
	var firstPatternStart int
	var patternFinalState int
	var leftContextEnd int
	var middleContextEnd int
	var statesUsed int

	if PatternParser(
		frame,
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

func EqsGetRepEqs(frame *FrameObject, rept LeadParam, tpar TParObject) bool {
	success := false

	if tpar.Dlm == TpdSmart {
		if !eqsgetrepPatternBuild(frame, tpar, &frame.EqsPatternPtr) {
			return false
		}
		markFlag := false
		var startCol int
		var endPos int
		found := PatternRecognize(
			frame,
			frame.EqsPatternPtr,
			frame.Dot.Line,
			frame.Dot.Col,
			&markFlag,
			&startCol,
			&endPos,
		)
		switch rept {
		case LeadParamNone, LeadParamPlus:
			success = (frame.Dot.Col == startCol) && found
		case LeadParamMinus:
			success = !((frame.Dot.Col == startCol) && found)
		case LeadParamPIndef:
			success = (endPos <= frame.Dot.Line.Used) && found
		case LeadParamNIndef:
			success = (endPos >= frame.Dot.Line.Used) && found
		}
		if success && rept != LeadParamMinus {
			MarkCreate(frame.Dot.Line, endPos, &frame.Marks[MarkEquals])
		}
	} else {
		exactcase := eqsgetrepExactcase(&tpar)
		startCol := frame.Dot.Col
		var length int
		if startCol > frame.Dot.Line.Used {
			length = 0
			startCol = 1
		} else {
			length = frame.Dot.Line.Used + 1 - frame.Dot.Col
		}
		if length > tpar.Len {
			length = tpar.Len
		}

		var nchIdent int
		result := ChCompareStr(
			tpar.Str,
			1,
			tpar.Len,
			frame.Dot.Line.Str,
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
			MarkCreate(
				frame.Dot.Line,
				frame.Dot.Col+nchIdent,
				&frame.Marks[MarkEquals],
			)
		}
	}
	return success
}

func eqsgetrepDumbGet(frame *FrameObject, count int, tpar TParObject, fromSpan bool) bool {
	result := (count == 0)

	dotLine := frame.Dot.Line
	dotCol := frame.Dot.Col
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
		length = min(frame.Dot.Col-1, line.Used)
	} else {
		newstr = tpar.Str
		backwards = false
		startCol = frame.Dot.Col
		if startCol > line.Used {
			length = 0
		} else {
			length = line.Used + 1 - startCol
		}
	}

outerLoop:
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
			processMatch := true
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
					processMatch = false
				}
			}
			if processMatch {
				startCol += offset
				if !backwards {
					startCol += tpar.Len
				}
				count--
				if count == 0 {
					MarkCreate(line, startCol, &frame.Dot)
					verified := true
					if !fromSpan {
						switch ScreenVerify(&Screen, frame, thisOne) {
						case VerifyReplyAlways, VerifyReplyYes:
							// accepted
						case VerifyReplyQuit, VerifyReplyNo:
							count = 1
							MarkCreate(dotLine, dotCol, &frame.Dot)
							verified = false
							if ExitAbort {
								break outerLoop
							}
						}
					}
					if verified {
						if backwards {
							MarkCreate(line, startCol+tpar.Len, &frame.Marks[MarkEquals])
						} else {
							MarkCreate(line, startCol-tpar.Len, &frame.Marks[MarkEquals])
						}
						result = true
						break outerLoop
					}
				}
			}
			// Update startCol/length for next search
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
				break
			}
			startCol = 1
			length = line.Used
		}
	}
	return result
}

func eqsgetrepPatternGet(
	frame *FrameObject,
	count int,
	tpar TParObject,
	fromSpan bool,
	replaceFlag bool,
) bool {
	result := (count == 0)

	var patternPtr *DFATableObject
	if !replaceFlag {
		if !eqsgetrepPatternBuild(frame, tpar, &frame.GetPatternPtr) {
			return result
		}
		patternPtr = frame.GetPatternPtr
	} else {
		patternPtr = frame.RepPatternPtr
	}

	dotLine := frame.Dot.Line
	dotCol := frame.Dot.Col
	line := dotLine
	markFlag := false
	backwards := count < 0
	var startCol int
	if backwards {
		startCol = 1
	} else {
		startCol = dotCol
	}
	count = iabs(count)
	if startCol > line.Used {
		startCol = line.Used + 1
	}

outerLoop:
	for count > 0 && !TtControlC {
		var matchedStartCol int
		var matchedFinishCol int
		if PatternRecognize(
			frame,
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
						MarkCreate(line, matchedStartCol, &frame.Dot)
					} else {
						MarkCreate(line, matchedFinishCol, &frame.Dot)
					}
					verified := true
					if !fromSpan {
						switch ScreenVerify(&Screen, frame, thisOne) {
						case VerifyReplyAlways, VerifyReplyYes:
							// accepted
						case VerifyReplyQuit, VerifyReplyNo:
							count = 1
							MarkCreate(dotLine, dotCol, &frame.Dot)
							verified = false
							if ExitAbort {
								break outerLoop
							}
						}
					}
					if verified {
						if backwards {
							MarkCreate(line, matchedFinishCol, &frame.Marks[MarkEquals])
						} else {
							MarkCreate(line, matchedStartCol, &frame.Marks[MarkEquals])
						}
						result = true
						break outerLoop
					}
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
						break
					}
					markFlag = false
					startCol = 1
				}
			} else {
				line = line.BLink
				if line == nil {
					break
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
				break
			}
			markFlag = false
			startCol = 1
		}
	}
	return result
}

func EqsGetRepGet(frame *FrameObject, count int, tpar TParObject, fromSpan bool) bool {
	if tpar.Dlm == TpdSmart {
		return eqsgetrepPatternGet(frame, count, tpar, fromSpan, false)
	}
	return eqsgetrepDumbGet(frame, count, tpar, fromSpan)
}

func EqsGetRepRep(
	frame *FrameObject,
	rept LeadParam,
	count int,
	tpar TParObject,
	tpar2 TParObject,
	fromSpan bool,
) bool {
	result := false

	var oldDot *MarkObject
	var oldEquals *MarkObject

	MarkCreate(frame.Dot.Line, frame.Dot.Col, &oldDot)
	if frame.Marks[MarkEquals] != nil {
		MarkCreate(
			frame.Marks[MarkEquals].Line,
			frame.Marks[MarkEquals].Col,
			&oldEquals,
		)
	}
	defer func() {
		MarkDestroy(&oldDot)
		MarkDestroy(&oldEquals)
	}()

	if tpar.Dlm == TpdSmart {
		if !eqsgetrepPatternBuild(frame, tpar, &frame.RepPatternPtr) {
			return result
		}
	}
	getcount := 1
	if rept == LeadParamMinus || rept == LeadParamNIndef || rept == LeadParamNInt {
		getcount = -1
	}
	if rept == LeadParamPIndef || rept == LeadParamNIndef {
		count = MaxInt
	} else if count < 0 {
		count = -count
	}

outerLoop:
	for count > 0 {
		var okay bool
		for {
			okay = true
			if TtControlC || ExitAbort {
				break outerLoop
			}
			if tpar.Dlm == TpdSmart {
				if !eqsgetrepPatternGet(frame, getcount, tpar, true, true) {
					break outerLoop
				}
			} else if !eqsgetrepDumbGet(frame, getcount, tpar, true) {
				break outerLoop
			}
			if TtControlC || ExitAbort {
				break outerLoop
			}
			if !fromSpan {
				switch ScreenVerify(&Screen, frame, replaceThisOne) {
				case VerifyReplyAlways:
					fromSpan = true
				case VerifyReplyYes:
					// accepted
				case VerifyReplyQuit, VerifyReplyNo:
					okay = false
				}
			}
			if okay {
				break
			}
		}

		length := frame.Marks[MarkEquals].Col - frame.Dot.Col
		if length < 0 {
			frame.Dot.Col = frame.Marks[MarkEquals].Col
			frame.Marks[MarkEquals].Col = frame.Dot.Col - length
			length = -length
		}
		if tpar2.Con == nil {
			startCol := frame.Dot.Col
			delta := length - tpar2.Len
			if delta > 0 {
				if frame.Dot.Col+delta > frame.Dot.Line.Used+1 {
					delta = frame.Dot.Line.Used + 1 - frame.Dot.Col
				}
				if delta > 0 {
					if !CharcmdDelete(frame, LeadParamPInt, delta, true) {
						return result
					}
				}
			} else if delta < 0 {
				if !CharcmdInsert(frame, LeadParamPInt, -delta, true) {
					return result
				}
				frame.Dot.Col = startCol
			}
			if !TextOvertype(true, 1, tpar2.Str, tpar2.Len, frame.Dot) {
				return result
			}
			if getcount > 0 {
				MarkCreate(frame.Dot.Line, startCol, &frame.Marks[MarkEquals])
			} else {
				MarkCreate(frame.Dot.Line, startCol+tpar2.Len, &frame.Marks[MarkEquals])
				frame.Dot.Col = startCol
			}
		} else {
			if !CharcmdDelete(frame, LeadParamPInt, length, true) {
				return result
			}
			if !TextInsertTpar(&tpar2, frame.Dot, &frame.Marks[MarkEquals]) {
				return result
			}
		}
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &oldDot)
		MarkCreate(
			frame.Marks[MarkEquals].Line,
			frame.Marks[MarkEquals].Col,
			&oldEquals,
		)
		frame.TextModified = true
		MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
		count--
	}

	// Restore dot and equals to their last-saved positions (typically the final replacement)
	MarkCreate(oldDot.Line, oldDot.Col, &frame.Dot)
	if oldEquals != nil {
		MarkCreate(oldEquals.Line, oldEquals.Col, &frame.Marks[MarkEquals])
	} else {
		MarkDestroy(&frame.Marks[MarkEquals])
	}
	result = (count == 0) || rept == LeadParamPIndef || rept == LeadParamNIndef
	return result
}
