/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         RECOGNIZE
//
// Description:  The pattern matcher for EQS, GET and REPLACE.

package ludwig

import "math/big"

func emptySet() *big.Int {
	return big.NewInt(0)

}

// patternGetInputElt gets the next input element for pattern matching
func patternGetInputElt(
	line *LineHdrObject,
	ch *byte,
	inputSet *big.Int,
	column *int,
	length int,
	markFlag *bool,
	endOfLine *bool,
) {
	// Clear the input set
	inputSet.SetInt64(0)

	if length == 0 {
		*ch = PatternSpace // not 100% correct but OK
		setAdd(inputSet, PatternBegLine)
		setAdd(inputSet, PatternEndLine)
		*markFlag = true
		*endOfLine = true
	} else {
		markFound := false
		if !*markFlag { // the last time through was not a mark so look for them this time
			if *column == 1 {
				setAdd(inputSet, PatternBegLine)
				markFound = true
			}
			if *column > length {
				setAdd(inputSet, PatternEndLine)
				*endOfLine = true
				markFound = true
			}
			if *column == CurrentFrame.MarginLeft {
				setAdd(inputSet, PatternLeftMargin)
				markFound = true
			}
			if *column == CurrentFrame.MarginRight {
				setAdd(inputSet, PatternRightMargin)
				markFound = true
			}
			if *column == CurrentFrame.Dot.Col {
				setAdd(inputSet, PatternDotColumn)
				markFound = true
			}
			if len(line.Marks) > 0 { // if any marks on this line
				for markNo := MinMarkNumber; markNo <= MaxMarkNumber; markNo++ {
					// run through user accessible ones
					if CurrentFrame.Marks[markNo-MinMarkNumber] != nil {
						if (CurrentFrame.Marks[markNo-MinMarkNumber].Line == line) &&
							(CurrentFrame.Marks[markNo-MinMarkNumber].Col == *column) {
							setAdd(inputSet, byte(markNo+PatternMarksStart))
							markFound = true
						}
					}
				}
			}
			// markFound is true if there is a user mark on this column
			*markFlag = markFound // will not test for a mark next time through
		}

		if !markFound {
			// if there is not a mark or we have already processed it, then get the char
			*markFlag = false // will test for a mark next time through
			*ch = line.Str.Get(*column)
			if *column <= length {
				*column++
			}
		}
	}
}

// setIntersection returns true if the two sets have any common elements
func setIntersection(set1 *big.Int, set2 *big.Int) bool {
	s := new(big.Int)
	return s.And(set1, set2).Sign() != 0
}

// patternNextState determines the next state in the DFA
func patternNextState(
	dfaTablePointer *DFATableObject,
	ch byte,
	inputSet *big.Int,
	markFlag bool,
	state *int,
	started *bool,
) bool {
	found := false
	transitionPointer := dfaTablePointer.DFATable[*state].Transitions
	if markFlag { // look for transitions on positionals only
		auxState := PatternDFAKill
		for transitionPointer != nil && !found {
			if setIntersection(&transitionPointer.TransitionAcceptSet, inputSet) {
				found = true
				if transitionPointer.StartFlag && !*started {
					auxState = PatternDFAKill
				} else {
					auxState = transitionPointer.AcceptNextState
				}
			} else {
				transitionPointer = transitionPointer.NextTransition
			}
		}
		if auxState == PatternDFAKill {
			found = false
		} else {
			*state = auxState
		}
	} else {
		// look for transitions on characters
		for transitionPointer != nil && !found {
			if transitionPointer.TransitionAcceptSet.Bit(int(ch)) != 0 {
				found = true
				if transitionPointer.StartFlag && !*started {
					*state = PatternDFAKill
				} else {
					*state = transitionPointer.AcceptNextState
				}
			} else {
				transitionPointer = transitionPointer.NextTransition
			}
		}
	}
	*started = (*started && dfaTablePointer.DFATable[*state].PatternStart) ||
		(*state == PatternDFAFail) || (*state == PatternDFAKill)
	return found
}

// PatternRecognize performs pattern matching on a line
func PatternRecognize(
	dfaTablePointer *DFATableObject,
	line *LineHdrObject,
	startCol int,
	markFlag *bool,
	startPos *int,
	finishPos *int,
) bool {
	lineCounter := startCol
	*startPos = startCol
	*finishPos = startCol
	state := PatternDFAStart
	found := false
	fail := false
	started := true
	endOfLine := false
	leftFlag := false
	var positionalSet big.Int
	var ch byte

	for {
		for {
			patternGetInputElt(
				line, &ch, &positionalSet, &lineCounter, line.Used, markFlag, &endOfLine,
			)
			if patternNextState(
				dfaTablePointer, ch, &positionalSet, *markFlag, &state, &started,
			) {
				if state == PatternDFAKill {
					state = PatternDFAStart
				} else if state == PatternDFAFail {
					fail = true
					startCol++
					lineCounter = startCol + 1
					state = PatternDFAStart
				}
				if dfaTablePointer.DFATable[state].LeftTransition {
					*startPos = lineCounter
				} else if dfaTablePointer.DFATable[state].LeftContextCheck {
					if leftFlag {
						*startPos = lineCounter - 1
					} else {
						leftFlag = true
					}
				}
				if dfaTablePointer.DFATable[state].RightTransition {
					*finishPos = lineCounter
				}
			}
			if dfaTablePointer.DFATable[state].FinalAccept || endOfLine {
				break
			}
		}

		if !endOfLine {
			for {
				patternGetInputElt(
					line, &ch, &positionalSet, &lineCounter, line.Used, markFlag, &endOfLine,
				)
				if patternNextState(
					dfaTablePointer, ch, &positionalSet, *markFlag, &state, &started,
				) {
					if state == PatternDFAKill {
						found = true
					} else if state == PatternDFAFail {
						fail = true
						startCol++
						lineCounter = startCol + 1
						state = PatternDFAStart
					}
					if dfaTablePointer.DFATable[state].RightTransition {
						*finishPos = lineCounter
					}
				}
				if found || fail || endOfLine {
					break
				}
			}
		}

		if found || endOfLine {
			break
		}
	}

	if !found { // must also be end of line push through the white space at end of line
		flag := dfaTablePointer.DFATable[state].FinalAccept
		if patternNextState(dfaTablePointer, ' ', emptySet(), false, &state, &started) {
			// note end of line positional will already have been processed
			if (state == PatternDFAKill) && flag {
				found = true
			}
			if dfaTablePointer.DFATable[state].RightTransition {
				*finishPos = line.Used + 1
			}
			if dfaTablePointer.DFATable[state].LeftTransition {
				*startPos = line.Used + 1
			}
			if dfaTablePointer.DFATable[state].FinalAccept {
				// does not need to be pushed to the kill state
				// as there is no more input, so there is no
				// possibility of a fail being generated
				found = true
			}
		}
	}
	return found
}
