/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         PATPARSE
//
// Description:  This is the parser for the pattern matcher for EQS, Get
//               and Replace.

package ludwig

import (
	"math/big"
)

// localException and otherException are used for non-local control flow
type localException struct{}
type otherException struct{}

// Predefined accept sets for pattern parsing
var (
	quotedSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		s[TpdLit] = true
		s[TpdExact] = true
		return s
	}()

	delimitedSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		s[PatternKStar] = true
		s['0'] = true
		s['1'] = true
		s['2'] = true
		s['3'] = true
		s['4'] = true
		s['5'] = true
		s['6'] = true
		s['7'] = true
		s['8'] = true
		s['9'] = true
		s[PatternLRangeDelim] = true
		s[PatternPlus] = true
		return s
	}()

	charsetsSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		s['s'] = true
		s['S'] = true
		s['a'] = true
		s['A'] = true
		s['c'] = true
		s['C'] = true
		s['l'] = true
		s['L'] = true
		s['u'] = true
		s['U'] = true
		s['n'] = true
		s['N'] = true
		s['p'] = true
		s['P'] = true
		return s
	}()

	positionalsSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		s['<'] = true
		s['>'] = true
		s['{'] = true
		s['}'] = true
		s['^'] = true
		return s
	}()

	chAndPosSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		for i := range s {
			s[i] = charsetsSet[i] || positionalsSet[i]
		}
		return s
	}()

	syntaxSet = func() [MaxSetRange + 1]bool {
		var s [MaxSetRange + 1]bool
		chars := []byte{
			TpdSpan, TpdPrompt, TpdExact, TpdLit,
			PatternLParen, PatternLRangeDelim, PatternKStar, PatternPlus,
			PatternNegate, PatternMark, PatternEquals, PatternModified,
			'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
			PatternDefineSetU, PatternDefineSetL,
			'a', 'b', 'c', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
			'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
			'A', 'B', 'C', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
			'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
			'{', '}', '<', '>', '^',
		}
		for _, ch := range chars {
			s[ch] = true
		}
		return s
	}()

	spaceSet       big.Int
	printableSet   big.Int
	alphaSet       big.Int
	lowerSet       big.Int
	upperSet       big.Int
	numericSet     big.Int
	punctuationSet big.Int
)

func init() {
	for i := 0; i <= MaxSetRange; i++ {
		r := rune(i)
		if ChIsSpace(r) {
			spaceSet.SetBit(&spaceSet, i, 1)
		}
		if ChIsPrintable(r) {
			printableSet.SetBit(&printableSet, i, 1)
		}
		if ChIsLetter(r) {
			alphaSet.SetBit(&alphaSet, i, 1)
		}
		if ChIsLower(r) {
			lowerSet.SetBit(&lowerSet, i, 1)
		}
		if ChIsUpper(r) {
			upperSet.SetBit(&upperSet, i, 1)
		}
		if ChIsNumeric(r) {
			numericSet.SetBit(&numericSet, i, 1)
		}
		if ChIsPunctuation(r) {
			punctuationSet.SetBit(&punctuationSet, i, 1)
		}
	}
}

// Helper function to check if a value is in a set
func setContains(set [MaxSetRange + 1]bool, val byte) bool {
	return set[val]
}

// Helper function to set union
func setUnion(a, b *big.Int) *big.Int {
	s := new(big.Int)
	return s.Or(a, b)
}

// Helper function to set difference (remove b from a)
func setRemove(a, b *big.Int) *big.Int {
	s := new(big.Int)
	return s.AndNot(a, b)
}

// Helper function to create a set from a single character
func singletonSet(ch byte) *big.Int {
	s := new(big.Int)
	return s.SetBit(s, int(ch), 1)
}

// Helper function to create a set from a range of characters
func rangeSet(start, end byte) *big.Int {
	s := new(big.Int)
	for i := int(start); i <= int(end); i++ {
		s.SetBit(s, i, 1)
	}
	return s
}

// Helper function to add an element to a set
func setAdd(set *big.Int, ch byte) {
	set.SetBit(set, int(ch), 1)
}

// Helper function to add a range to a set
func setAddRange(set *big.Int, start, end byte) {
	for i := int(start); i <= int(end); i++ {
		set.SetBit(set, i, 1)
	}
}

// Helper function to clear a set
func setClear(set *big.Int) {
	set.SetInt64(0)
}

// PatternParser parses a pattern and builds an NFA table
func PatternParser(
	pattern *TParObject,
	nfaTable *NFATableType,
	firstPatternStart *int,
	patternFinalState *int,
	leftContextEnd *int,
	middleContextEnd *int,
	patternDefinition *PatternDefType,
	statesUsed *int,
) bool {
	var firstPatternEnd int
	var parseCount int

	// patternNewNFA allocates a new NFA state
	patternNewNFA := func() int {
		if *statesUsed < MaxNFAStateRange {
			newNfa := *statesUsed
			nfaTable[*statesUsed].Fail = false
			nfaTable[*statesUsed].Indefinite = false
			*statesUsed++
			return newNfa
		}
		ScreenMessage(MsgPatPatternTooComplex)
		panic(localException{})
	}

	// patternDuplicateNFA duplicates an NFA path
	patternDuplicateNFA := func(copyThisStart, copyThisFinish, currentState int, duplicateFinish *int) bool {
		offset := (currentState - copyThisStart) + 1
		if (*statesUsed + offset) > MaxNFAStateRange {
			ScreenMessage(MsgPatPatternTooComplex)
			panic(localException{})
		}
		duplicateStart := *statesUsed

		for aux := copyThisStart; aux <= copyThisFinish; aux++ {
			auxState := patternNewNFA()
			nfaTable[auxState].Fail = nfaTable[aux].Fail
			if !nfaTable[auxState].Fail {
				nfaTable[auxState].EpsilonOut = nfaTable[aux].EpsilonOut
				if nfaTable[auxState].EpsilonOut {
					if nfaTable[aux].FirstOut == PatternNull {
						nfaTable[auxState].FirstOut = nfaTable[aux].FirstOut
					} else {
						nfaTable[auxState].FirstOut = nfaTable[aux].FirstOut + offset
					}
					if nfaTable[aux].SecondOut == PatternNull {
						nfaTable[auxState].SecondOut = nfaTable[aux].SecondOut
					} else {
						nfaTable[auxState].SecondOut = nfaTable[aux].SecondOut + offset
					}
				} else {
					if nfaTable[aux].NextState == PatternNull {
						nfaTable[auxState].NextState = nfaTable[aux].NextState
					} else {
						nfaTable[auxState].NextState = nfaTable[aux].NextState + offset
					}
					nfaTable[auxState].AcceptSet.Set(&nfaTable[aux].AcceptSet)
				}
			}
			nfaTable[currentState].EpsilonOut = true
			nfaTable[currentState].FirstOut = duplicateStart
		}
		*duplicateFinish = *statesUsed - 1
		return true
	}

	// patternGetch reads next character from pattern
	patternGetch := func(parseCount *int, ch *byte, inString *TParObject) bool {
		result := true
		if *parseCount < inString.Len {
			*parseCount++
			*ch = inString.Str.Get(*parseCount)
			patternDefinition.Length++
			if patternDefinition.Length > MaxStrLen {
				ScreenMessage(MsgPatPatternTooComplex)
				panic(localException{})
			}
			patternDefinition.Strng.Set(patternDefinition.Length, *ch)
		} else {
			*ch = 0 // null
			result = false
		}
		return result
	}

	// patternGetnumb reads a number from pattern
	patternGetnumb := func(parseCount *int, number *int, ch *byte, inString *TParObject) bool {
		auxBool := patternGetch(parseCount, ch, inString)
		result := auxBool && (*ch >= '0' && *ch <= '9')

		*number = 0
		for auxBool && (*ch >= '0' && *ch <= '9') {
			*number = (*number * 10) + int(*ch-'0')
			auxBool = patternGetch(parseCount, ch, inString)
		}
		if auxBool {
			*parseCount--
		}
		return result
	}

	// Forward declarations for mutually recursive functions
	var patternCompound func(first int, finish *int, parseCount *int, inString *TParObject, patCh *byte, depth int)
	var patternPattern func(first int, finish *int, parseCount *int, inString *TParObject, patCh *byte, depth int)

	// patternPattern is the main pattern parsing function
	patternPattern = func(first int, finish *int, parseCount *int, inString *TParObject, patCh *byte, depth int) {
		var leadingParam ParameterType
		var aux, auxCount, temporary int
		var delimiter, auxCh1, auxCh2, auxPatCh byte
		var derefSpan TParObject
		var tparSort Commands
		var currentState, auxState, beginState int
		var endOfInput, negate, noDereference bool
		var rangePatch, rangeStart, rangeEnd int
		var auxi int
		var rangeIndefinite bool

		// patternRangeDelimgen processes range delimiters
		patternRangeDelimgen := func(rangePatch, rangeStart, rangeEnd *int, rangeIndefinite *bool, leadingParam *ParameterType) {
			*rangeIndefinite = false
			*leadingParam = PatternRange
			switch *patCh {
			case PatternKStar:
				*rangeStart = 0
				*rangeEnd = 0
				*rangeIndefinite = true
			case PatternPlus:
				*rangeStart = 1
				*rangeEnd = 0
				*rangeIndefinite = true
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				*parseCount--
				patternGetnumb(parseCount, rangeStart, patCh, inString)
				*rangeEnd = *rangeStart
			case PatternLRangeDelim:
				if !patternGetch(parseCount, patCh, inString) {
					ScreenMessage(MsgPatNoMatchingDelim)
					panic(localException{})
				}
				if *patCh >= '0' && *patCh <= '9' {
					*parseCount--
					_ = patternGetnumb(parseCount, rangeStart, patCh, inString)
					_ = patternGetch(parseCount, patCh, inString)
				} else {
					*rangeStart = 0
				}
				if *patCh == PatternComma {
					if !patternGetch(parseCount, patCh, inString) {
						ScreenMessage(MsgPatNoMatchingDelim)
						panic(localException{})
					}
				} else {
					ScreenMessage(MsgPatErrorInRange)
					panic(localException{})
				}
				if *patCh >= '0' && *patCh <= '9' {
					*parseCount--
					_ = patternGetnumb(parseCount, rangeEnd, patCh, inString)
					_ = patternGetch(parseCount, patCh, inString)
				} else {
					*rangeIndefinite = true
					*rangeEnd = 0
				}
				if *patCh != PatternRRangeDelim {
					ScreenMessage(MsgPatNoMatchingDelim)
					panic(localException{})
				}
			}
			if *rangeStart == 0 {
				*rangePatch = currentState
				nfaTable[*rangePatch].EpsilonOut = true
				nfaTable[*rangePatch].FirstOut = patternNewNFA()
				nfaTable[*rangePatch].SecondOut = PatternNull
				currentState = nfaTable[*rangePatch].FirstOut
			} else {
				*rangePatch = PatternNull
			}
		}

		// patternRangeBuild builds a range pattern
		patternRangeBuild := func(rangeStart, rangeEnd, rangePatch int, indefinite bool) {
			endState := currentState
			indefinitePatch := beginState
			divertPtr := rangePatch

			nfaTable[currentState].EpsilonOut = true
			nfaTable[currentState].FirstOut = PatternNull
			nfaTable[currentState].SecondOut = PatternNull

			for aux := 2; aux <= rangeStart; aux++ {
				indefinitePatch = currentState
				_ = patternDuplicateNFA(beginState, endState, currentState, &currentState)
			}

			if rangeStart > 0 {
				rangeStart--
			}

			for aux := rangeStart + 2; aux <= rangeEnd; aux++ {
				nfaTable[currentState].SecondOut = divertPtr
				divertPtr = currentState
				_ = patternDuplicateNFA(beginState, endState, currentState, &currentState)
			}

			nfaTable[currentState].SecondOut = PatternNull

			for divertPtr != PatternNull {
				auxPtr := nfaTable[divertPtr].SecondOut
				nfaTable[divertPtr].SecondOut = currentState
				divertPtr = auxPtr
			}

			if indefinite {
				nfaTable[currentState].EpsilonOut = true
				nfaTable[currentState].FirstOut = indefinitePatch
				nfaTable[currentState].SecondOut = patternNewNFA()
				currentState = nfaTable[currentState].SecondOut
				nfaTable[indefinitePatch].Indefinite = true
			}
		}

		derefSpan.Nxt = nil
		derefSpan.Con = nil

		if depth > PatternMaxDepth {
			ScreenMessage(MsgPatPatternTooComplex)
			panic(localException{})
		}

		endOfInput = false
		currentState = first
		endOfInput = false

		for !endOfInput && (*patCh == PatternSpace) {
			endOfInput = !patternGetch(parseCount, patCh, inString)
		}

		for (*patCh != PatternComma && *patCh != PatternRParen && *patCh != PatternBar) && !endOfInput {
			if !setContains(syntaxSet, *patCh) {
				ScreenMessage(MsgPatIllegalSymbol)
				panic(localException{})
			}

			switch *patCh {
			case TpdSpan, TpdPrompt:
				delimiter = *patCh
				var derefTpar TParObject
				aux = 0
				if !patternGetch(parseCount, patCh, inString) {
					ScreenMessage(MsgPatNoMatchingDelim)
					panic(localException{})
				}
				for *patCh != delimiter {
					aux++
					derefTpar.Str.Set(aux, *patCh)
					if !patternGetch(parseCount, patCh, inString) {
						ScreenMessage(MsgPatNoMatchingDelim)
						panic(localException{})
					}
				}
				patternDefinition.Length = patternDefinition.Length - (aux + 2)
				tparSort = CmdPatternDummyPattern
				derefTpar.Len = aux
				derefTpar.Dlm = delimiter
				if !TparGet1(&derefTpar, tparSort, &derefSpan) {
					panic(otherException{})
				}

				if setContains(quotedSet, derefSpan.Dlm) {
					// Insert quote at beginning: shift string right and add delimiters
					for i := derefSpan.Len; i >= 1; i-- {
						derefSpan.Str.Set(i+1, derefSpan.Str.Get(i))
					}
					derefSpan.Len += 2
					derefSpan.Str.Set(derefSpan.Len, derefSpan.Dlm)
					derefSpan.Str.Set(1, derefSpan.Dlm)
				}
				auxCount = 0
				if patternGetch(&auxCount, &auxPatCh, &derefSpan) {
					patternCompound(currentState, &currentState, &auxCount, &derefSpan, &auxPatCh, depth+1)
					if (auxCount != derefSpan.Len) || (derefSpan.Str.Get(auxCount) == PatternComma) {
						ScreenMessage(MsgPatErrorInSpan)
						panic(localException{})
					}
				}

			case TpdExact, TpdLit, PatternLParen, PatternLRangeDelim, PatternKStar, PatternPlus,
				PatternNegate, PatternMark, PatternEquals, PatternModified,
				'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
				PatternDefineSetU, PatternDefineSetL,
				'a', 'b', 'c', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
				'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
				'A', 'B', 'C', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
				'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
				'{', '}', '<', '>', '^':

				leadingParam = NullParam
				if setContains(delimitedSet, *patCh) {
					patternRangeDelimgen(&rangePatch, &rangeStart, &rangeEnd, &rangeIndefinite, &leadingParam)
					if !patternGetch(parseCount, patCh, inString) {
						ScreenMessage(MsgPatPrematurePatternEnd)
						panic(localException{})
					}
					beginState = currentState
				}

				switch *patCh {
				case TpdExact, TpdLit:
					auxCh1 = *patCh
					auxCount = *parseCount
					if !patternGetch(parseCount, patCh, inString) {
						ScreenMessage(MsgPatNoMatchingDelim)
						panic(localException{})
					}
					if *patCh == TpdSpan || *patCh == TpdPrompt {
						noDereference = false
						delimiter = *patCh
						var derefTpar TParObject
						aux = 0
						for {
							if !patternGetch(parseCount, patCh, inString) {
								ScreenMessage(MsgPatNoMatchingDelim)
								panic(localException{})
							}
							aux++
							derefTpar.Str.Set(aux, *patCh)
							if *patCh == auxCh1 {
								break
							}
						}
						if aux >= 2 {
							if derefTpar.Str.Get(aux-1) == delimiter {
								patternDefinition.Length -= (aux + 2)
								tparSort = CmdPatternDummyText
								derefTpar.Len = aux - 2
								derefTpar.Dlm = delimiter
								if !TparGet1(&derefTpar, tparSort, &derefSpan) {
									panic(otherException{})
								}
								// Insert quote at beginning: shift string right and add delimiters
								for i := derefSpan.Len; i >= 1; i-- {
									derefSpan.Str.Set(i+1, derefSpan.Str.Get(i))
								}
								derefSpan.Len += 2
								derefSpan.Str.Set(derefSpan.Len, auxCh1)
								derefSpan.Str.Set(1, auxCh1)
								auxCount = 0
								if patternGetch(&auxCount, &auxPatCh, &derefSpan) {
									patternPattern(currentState, &currentState, &auxCount, &derefSpan, &auxPatCh, depth+1)
								}
							} else {
								*patCh = delimiter
								patternDefinition.Length += (aux + 2)
								*parseCount = auxCount + 1
								noDereference = true
							}
						} else {
							*patCh = delimiter
							patternDefinition.Length += (aux + 2)
							*parseCount = auxCount + 1
							noDereference = true
						}
					} else {
						noDereference = true
					}

					if noDereference {
						if auxCh1 == TpdExact {
							for *patCh != TpdExact {
								nfaTable[currentState].EpsilonOut = false
								sset := singletonSet(*patCh)
								nfaTable[currentState].AcceptSet.Set(sset)
								nfaTable[currentState].NextState = patternNewNFA()
								currentState = nfaTable[currentState].NextState
								if !patternGetch(parseCount, patCh, inString) {
									ScreenMessage(MsgPatNoMatchingDelim)
									panic(localException{})
								}
							}
						} else {
							for *patCh != TpdLit {
								nfaTable[currentState].EpsilonOut = false
								if *patCh >= 'a' && *patCh <= 'z' {
									uset := setUnion(
										singletonSet(*patCh),
										singletonSet(ChToUpper(*patCh)),
									)
									nfaTable[currentState].AcceptSet.Set(uset)
								} else if *patCh >= 'A' && *patCh <= 'Z' {
									uset := setUnion(
										singletonSet(*patCh),
										singletonSet(ChToLower(*patCh)),
									)
									nfaTable[currentState].AcceptSet.Set(uset)
								} else {
									sset := singletonSet(*patCh)
									nfaTable[currentState].AcceptSet.Set(sset)
								}
								nfaTable[currentState].NextState = patternNewNFA()
								currentState = nfaTable[currentState].NextState
								if !patternGetch(parseCount, patCh, inString) {
									ScreenMessage(MsgPatNoMatchingDelim)
									panic(localException{})
								}
							}
						}
					}

				case PatternLParen:
					_ = patternGetch(parseCount, patCh, inString)
					patternCompound(currentState, &auxState, parseCount, inString, patCh, depth+1)
					currentState = auxState
					if *patCh != PatternRParen {
						ScreenMessage(MsgPatNoMatchingDelim)
						panic(localException{})
					}

				case PatternMark:
					if patternGetnumb(parseCount, &auxi, patCh, inString) {
						if (auxi == 0) || (auxi > MaxUserMarkNumber) {
							ScreenMessage(MsgPatIllegalMarkNumber)
							panic(localException{})
						}
						nfaTable[currentState].EpsilonOut = false
						sset := singletonSet(byte(auxi + PatternMarksStart))
						nfaTable[currentState].AcceptSet.Set(sset)
						nfaTable[currentState].NextState = patternNewNFA()
						currentState = nfaTable[currentState].NextState
					} else {
						ScreenMessage(MsgPatIllegalMarkNumber)
						panic(localException{})
					}

				case PatternEquals, PatternModified:
					nfaTable[currentState].EpsilonOut = false
					if *patCh == PatternEquals {
						sset := singletonSet(PatternMarksEquals)
						nfaTable[currentState].AcceptSet.Set(sset)
					} else {
						sset := singletonSet(PatternMarksModified)
						nfaTable[currentState].AcceptSet.Set(sset)
					}
					nfaTable[currentState].NextState = patternNewNFA()
					currentState = nfaTable[currentState].NextState

				default:
					negate = false
					if *patCh == PatternNegate {
						negate = true
						if !patternGetch(parseCount, patCh, inString) {
							ScreenMessage(MsgPatPrematurePatternEnd)
							panic(localException{})
						}
					}

					if (*patCh == PatternDefineSetU) || (*patCh == PatternDefineSetL) {
						auxSet := new(big.Int)
						setClear(auxSet)
						if !patternGetch(parseCount, patCh, inString) {
							ScreenMessage(MsgPatPrematurePatternEnd)
							panic(localException{})
						}
						delimiter = *patCh
						temporary = patternDefinition.Length
						patternDefinition.Strng.Set(temporary, 0)

						if delimiter == TpdSpan || delimiter == TpdPrompt {
							var derefTpar TParObject
							derefTpar.Dlm = delimiter
							if !patternGetch(parseCount, patCh, inString) {
								ScreenMessage(MsgPatPrematurePatternEnd)
								panic(localException{})
							}
							aux = 0
							for *patCh != delimiter {
								aux++
								derefTpar.Str.Set(aux, *patCh)
								if !patternGetch(parseCount, patCh, inString) {
									ScreenMessage(MsgPatNoMatchingDelim)
									panic(localException{})
								}
							}
							patternDefinition.Length -= (aux + 1)
							derefTpar.Len = aux
							if !TparGet1(&derefTpar, tparSort, &derefSpan) {
								panic(otherException{})
							}
						} else {
							aux = 0
							if !patternGetch(parseCount, patCh, inString) {
								ScreenMessage(MsgPatPrematurePatternEnd)
								panic(localException{})
							}
							patternDefinition.Strng.Set(patternDefinition.Length, 0)
							for *patCh != delimiter {
								aux++
								derefSpan.Str.Set(aux, *patCh)
								if !patternGetch(parseCount, patCh, inString) {
									ScreenMessage(MsgPatNoMatchingDelim)
									panic(localException{})
								}
							}
							derefSpan.Len = aux
						}

						setClear(auxSet)
						aux = 1
						for aux <= derefSpan.Len {
							auxCh1 = derefSpan.Str.Get(aux)
							auxCh2 = auxCh1
							aux++
							if aux+2 <= derefSpan.Len {
								if (derefSpan.Str.Get(aux) == '.') && (derefSpan.Str.Get(aux+1) == '.') {
									auxCh2 = derefSpan.Str.Get(aux + 2)
									aux += 3
								}
							}
							setAddRange(auxSet, auxCh1, auxCh2)
						}

						for auxCount = 1; auxCount <= derefSpan.Len; auxCount++ {
							patternDefinition.Strng.Set(temporary+auxCount, derefSpan.Str.Get(auxCount))
						}
						patternDefinition.Length = temporary + derefSpan.Len + 1
						patternDefinition.Strng.Set(patternDefinition.Length, 0)
						if negate {
							fullSet := rangeSet(PatternAlphaStart, MaxSetRange)
							auxSet = setRemove(fullSet, auxSet)
						}
						nfaTable[currentState].EpsilonOut = false
						nfaTable[currentState].AcceptSet.Set(auxSet)
						nfaTable[currentState].NextState = patternNewNFA()
						currentState = nfaTable[currentState].NextState

					} else if setContains(chAndPosSet, *patCh) {
						nfaTable[currentState].EpsilonOut = false
						if setContains(positionalsSet, *patCh) {
							if negate {
								ScreenMessage(MsgPatIllegalParameter)
								panic(localException{})
							}
							setClear(&nfaTable[currentState].AcceptSet)
							switch *patCh {
							case '<':
								setAdd(&nfaTable[currentState].AcceptSet, PatternBegLine)
							case '>':
								setAdd(&nfaTable[currentState].AcceptSet, PatternEndLine)
							case '{':
								setAdd(&nfaTable[currentState].AcceptSet, PatternLeftMargin)
							case '}':
								setAdd(&nfaTable[currentState].AcceptSet, PatternRightMargin)
							case '^':
								setAdd(&nfaTable[currentState].AcceptSet, PatternDotColumn)
							}
						} else {
							upperCh := ChToUpper(*patCh)
							switch upperCh {
							case 'S':
								nfaTable[currentState].AcceptSet.Set(&spaceSet)
							case 'C':
								nfaTable[currentState].AcceptSet.Set(&printableSet)
							case 'A':
								nfaTable[currentState].AcceptSet.Set(&alphaSet)
							case 'L':
								nfaTable[currentState].AcceptSet.Set(&lowerSet)
							case 'U':
								nfaTable[currentState].AcceptSet.Set(&upperSet)
							case 'N':
								nfaTable[currentState].AcceptSet.Set(&numericSet)
							case 'P':
								nfaTable[currentState].AcceptSet.Set(&punctuationSet)
							}
							if negate {
								fullSet := rangeSet(PatternAlphaStart, MaxSetRange)
								rset := setRemove(fullSet, &nfaTable[currentState].AcceptSet)
								nfaTable[currentState].AcceptSet.Set(rset)
							}
						}
						nfaTable[currentState].NextState = patternNewNFA()
						currentState = nfaTable[currentState].NextState
					} else {
						ScreenMessage(MsgPatSetNotDefined)
						panic(localException{})
					}
				}

				if leadingParam == PatternRange {
					nfaTable[currentState].EpsilonOut = true
					nfaTable[currentState].FirstOut = PatternNull
					patternRangeBuild(rangeStart, rangeEnd, rangePatch, rangeIndefinite)
				}
			}

			for {
				endOfInput = !patternGetch(parseCount, patCh, inString)
				if endOfInput || (*patCh != PatternSpace) {
					break
				}
			}
		}
		*finish = currentState
		TparCleanObject(&derefSpan)
	}

	// patternCompound handles compound patterns (with alternatives)
	patternCompound = func(first int, finish *int, parseCount *int, inString *TParObject, patCh *byte, depth int) {
		var compoundFinish int
		var currentEStart int

		if depth > PatternMaxDepth {
			ScreenMessage(MsgPatPatternTooComplex)
			panic(localException{})
		}

		currentEStart = first
		nfaTable[currentEStart].EpsilonOut = true
		nfaTable[currentEStart].SecondOut = PatternNull
		nfaTable[currentEStart].FirstOut = patternNewNFA()
		patternPattern(nfaTable[currentEStart].FirstOut, finish, parseCount, inString, patCh, depth+1)
		compoundFinish = *finish

		if *patCh == PatternBar {
			if patternGetch(parseCount, patCh, inString) {
				nfaTable[currentEStart].SecondOut = patternNewNFA()
				patternCompound(nfaTable[currentEStart].SecondOut, finish, parseCount, inString, patCh, depth+1)
				nfaTable[compoundFinish].EpsilonOut = true
				nfaTable[compoundFinish].FirstOut = PatternNull
				nfaTable[compoundFinish].SecondOut = *finish
			} else {
				nfaTable[currentEStart].SecondOut = *finish
			}
		}
	}

	result := true
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case localException, otherException:
				result = false
			default:
				panic(r)
			}
		}
	}()

	ExitAbort = true
	patternDefinition.Length = 0

	nfaTable[PatternNull].EpsilonOut = true
	nfaTable[PatternNull].FirstOut = PatternNull
	nfaTable[PatternNull].SecondOut = PatternNull

	*statesUsed = PatternNFAStart
	*firstPatternStart = patternNewNFA()
	parseCount = 0
	var patCh byte
	if patternGetch(&parseCount, &patCh, pattern) {
		patternCompound(*firstPatternStart, &firstPatternEnd, &parseCount, pattern, &patCh, 1)
		if patCh == PatternComma {
			if patternGetch(&parseCount, &patCh, pattern) {
				*leftContextEnd = firstPatternEnd
				patternCompound(*leftContextEnd, middleContextEnd, &parseCount, pattern, &patCh, 1)
			} else {
				*middleContextEnd = firstPatternEnd
				*patternFinalState = firstPatternEnd
			}
		} else {
			*leftContextEnd = *firstPatternStart
			*patternFinalState = firstPatternEnd
			*middleContextEnd = firstPatternEnd
		}
		if patCh == PatternComma {
			if patternGetch(&parseCount, &patCh, pattern) {
				patternCompound(*middleContextEnd, patternFinalState, &parseCount, pattern, &patCh, 1)
			} else {
				*patternFinalState = *middleContextEnd
			}
		} else {
			*patternFinalState = *middleContextEnd
		}
	} else {
		ScreenMessage(MsgPatNullPattern)
		panic(localException{})
	}

	nfaTable[*patternFinalState].EpsilonOut = true
	nfaTable[*patternFinalState].FirstOut = PatternNull
	nfaTable[*patternFinalState].SecondOut = PatternNull

	result = true
	ExitAbort = false

	return result
}
