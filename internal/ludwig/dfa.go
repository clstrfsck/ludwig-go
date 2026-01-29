/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         DFA
//
// Description:  Builds the deterministic FSA for the pattern recognizer.

package ludwig

import "math/big"

// acceptSetPartitionType represents a partition of an accept set
type acceptSetPartitionType struct {
	acceptSetPartition big.Int // bitset
	nfaTransitionList  NFAAttributeType
	flink              *acceptSetPartitionType
	blink              *acceptSetPartitionType
}

// closureKill frees the equiv_list in a closure
func closureKill(closure *NFAAttributeType) {
	pointer1 := closure.EquivList
	for pointer1 != nil {
		pointer2 := pointer1.NextElt
		pointer1 = pointer2
	}
	closure.EquivList = nil
}

// transitionKill frees a transition list
func transitionKill(pointer1 **TransitionObject) {
	for *pointer1 != nil {
		pointer2 := (*pointer1).NextTransition
		*pointer1 = pointer2
	}
}

// PatternDFATableKill destroys a DFA table and frees its memory
func PatternDFATableKill(patternPtr **DFATableObject) bool {
	if *patternPtr != nil {
		for count := 0; count <= (*patternPtr).DFAStatesUsed; count++ {
			transitionKill(&(*patternPtr).DFATable[count].Transitions)
			closureKill(&(*patternPtr).DFATable[count].NFAAttributes)
		}
		*patternPtr = nil
	}
	return true
}

// PatternDFATableInitialize initializes or reinitializes a DFA table
func PatternDFATableInitialize(patternPtr **DFATableObject, patternDefinition PatternDefType) bool {
	if *patternPtr != nil {
		for count := 0; count <= (*patternPtr).DFAStatesUsed; count++ {
			transitionKill(&(*patternPtr).DFATable[count].Transitions)
			closureKill(&(*patternPtr).DFATable[count].NFAAttributes)
		}
		(*patternPtr).DFAStatesUsed = 0
	} else {
		*patternPtr = &DFATableObject{}
		(*patternPtr).DFAStatesUsed = 0
		for count := 0; count <= MaxDFAStateRange; count++ {
			(*patternPtr).DFATable[count].Transitions = nil
			(*patternPtr).DFATable[count].NFAAttributes.EquivList = nil
		}
	}
	(*patternPtr).Definition = patternDefinition
	return true
}

// Helper functions for PatternDFAConvert

// epsilonClosures computes the epsilon closure of a state set
func epsilonClosures(
	nfaTable *NFATableType,
	stateSet *NFAAttributeType,
	closure *NFAAttributeType,
) bool {
	const maxStackSize = 50
	stack := make([]int, maxStackSize+1)
	stackTop := 0
	failEquivalent := false

	pushStack := func(state int) bool {
		if stackTop < maxStackSize {
			stackTop++
			stack[stackTop] = state
			closure.EquivSet[state] = true
		} else {
			ScreenMessage(MsgPatPatternTooComplex)
			return false
		}
		return true
	}

	// Clear closure sets
	for i := range closure.EquivSet {
		closure.EquivSet[i] = false
	}
	for i := range closure.GeneratorSet {
		closure.GeneratorSet[i] = false
	}

	stateEltPtr := stateSet.EquivList
	for stateEltPtr != nil {
		if !pushStack(stateEltPtr.StateElt) {
			return false
		}
		stateEltPtr = stateEltPtr.NextElt
	}

	for stackTop != 0 && !failEquivalent {
		auxState := stack[stackTop] // pop off stack
		stackTop--
		nta := &nfaTable[auxState]
		if nta.Fail {
			failEquivalent = true
		}
		if nta.EpsilonOut {
			if nta.FirstOut != PatternNull {
				if !closure.EquivSet[nta.FirstOut] {
					if !pushStack(nta.FirstOut) {
						return false
					}
				}
			}
			if nta.SecondOut != PatternNull {
				if !closure.EquivSet[nta.SecondOut] {
					if !pushStack(nta.SecondOut) {
						return false
					}
				}
			}
		}
	}

	if failEquivalent {
		closure.EquivList = nil
		closure.EquivSet[PatternDFAFail] = true
	}
	return true
}

// epsilonAndMask computes epsilon closure and mask for a state
func epsilonAndMask(
	nfaTable *NFATableType,
	state int,
	closureSet *[MaxNFAStateRange + 1]bool,
	mask *[MaxNFAStateRange + 1]bool,
	maxim bool,
) bool {
	var transitionSet NFAAttributeType

	auxEltPtr := &StateEltObject{}
	auxEltPtr.NextElt = nil
	auxEltPtr.StateElt = state
	transitionSet.EquivList = auxEltPtr
	transitionSet.EquivSet[state] = true
	if !epsilonClosures(nfaTable, &transitionSet, &transitionSet) {
		return false
	}
	*closureSet = transitionSet.EquivSet

	var auxElt int
	if maxim {
		auxElt = MaxNFAStateRange // the elt corr to M-C is always present
		for !(*closureSet)[auxElt] {
			auxElt--
		}
		for i := 0; i <= auxElt; i++ {
			mask[i] = true
		}
	} else {
		auxElt = PatternNFAStart
		for !(*closureSet)[auxElt] {
			auxElt++
		}
		for i := 0; i < auxElt; i++ {
			mask[i] = true
		}
	}
	return true
}

// transitionListMerge makes a copy of 2 lists and concatenates them
func transitionListMerge(list1 *StateEltObject, list2 *StateEltObject) *StateEltObject {
	var aux1 *StateEltObject
	for list1 != nil {
		aux2 := aux1
		aux1 = &StateEltObject{}
		aux1.StateElt = list1.StateElt
		aux1.NextElt = aux2
		list1 = list1.NextElt
	}
	for list2 != nil {
		aux2 := aux1
		aux1 = &StateEltObject{}
		aux1.StateElt = list2.StateElt
		aux1.NextElt = aux2
		list2 = list2.NextElt
	}
	return aux1
}

// transitionListAppend makes a copy of list2 and concatenates it to list1
func transitionListAppend(list1 *StateEltObject, list2 *StateEltObject) *StateEltObject {
	aux1 := list1
	for list2 != nil {
		aux2 := aux1
		aux1 = &StateEltObject{}
		aux1.StateElt = list2.StateElt
		aux1.NextElt = aux2
		list2 = list2.NextElt
	}
	return aux1
}

// Helper functions for bitset operations
func bitsetClear(set *[MaxNFAStateRange + 1]bool) {
	for i := range set {
		set[i] = false
	}
}

func bitsetEquals(set1 *big.Int, set2 *big.Int) bool {
	return set1.Cmp(set2) == 0
}

func bitsetEqualsNFA(set1 [MaxNFAStateRange + 1]bool, set2 [MaxNFAStateRange + 1]bool) bool {
	for i := range set1 {
		if set1[i] != set2[i] {
			return false
		}
	}
	return true
}

func bitsetDifference(set1 [MaxNFAStateRange + 1]bool, set2 [MaxNFAStateRange + 1]bool) [MaxNFAStateRange + 1]bool {
	var result [MaxNFAStateRange + 1]bool
	for i := range set1 {
		result[i] = set1[i] && !set2[i]
	}
	return result
}

func bitsetIsEmpty(set [MaxNFAStateRange + 1]bool) bool {
	for _, v := range set {
		if v {
			return false
		}
	}
	return true
}

func bitsetIntersection(set1 *big.Int, set2 *big.Int) *big.Int {
	s := new(big.Int)
	return s.And(set1, set2)
}

func bitsetIsEmptyAccept(set *big.Int) bool {
	return set.Sign() == 0
}

func bitsetRemove(set1 *big.Int, set2 *big.Int) {
	set1.AndNot(set1, set2)
}

func bitsetRemoveNFA(set1 *[MaxNFAStateRange + 1]bool, set2 [MaxNFAStateRange + 1]bool) {
	for i := range *set1 {
		if set2[i] {
			set1[i] = false
		}
	}
}

func bitsetSetRange(set *big.Int, start int, end int) {
	for i := start; i <= end && i <= MaxSetRange; i++ {
		set.SetBit(set, i, 1)
	}
}

func bitsetSetRangeNFA(set *[MaxNFAStateRange + 1]bool, start int, end int) {
	for i := start; i <= end && i <= MaxNFAStateRange; i++ {
		set[i] = true
	}
}

// PatternDFAConvert converts an NFA to a DFA
func PatternDFAConvert(
	nfaTable *NFATableType,
	dfaTablePointer *DFATableObject,
	nfaStart int,
	nfaEnd *int,
	middleContextStart int,
	rightContextStart int,
	dfaStart *int,
	dfaEnd *int,
) bool {
	var auxElt *StateEltObject
	var state, currentState int
	var closureSet [MaxNFAStateRange + 1]bool
	var transferState NFAAttributeType
	var transitionSet NFAAttributeType
	var auxStatePtr *StateEltObject
	var statesUsed int
	var auxCount, auxCount2 int
	var incomingTranPtr, killTranPtr, auxTranPtr *TransitionObject
	var auxTranPtr2 *TransitionObject
	var found bool
	var auxSet [MaxNFAStateRange + 1]bool
	var mask [MaxNFAStateRange + 1]bool
	var killSet big.Int
	var partitionPtr *acceptSetPartitionType
	var auxPartitionPtr *acceptSetPartitionType
	var currentPartitionPtr *acceptSetPartitionType
	var followerPtr *acceptSetPartitionType
	var insertPartition *acceptSetPartitionType
	var auxEquivPtr *StateEltObject
	var auxClosure NFAAttributeType

	// patternNewDFA creates a new DFA state
	patternNewDFA := func(equivalentSet *NFAAttributeType, stateCount *int) bool {
		if statesUsed < MaxDFAStateRange {
			statesUsed++
			dts := &dfaTablePointer.DFATable[statesUsed]
			dts.NFAAttributes = *equivalentSet
			dts.NFAAttributes.EquivList = nil
			for i := 0; i <= MaxNFAStateRange; i++ {
				if equivalentSet.EquivSet[i] {
					auxElt := &StateEltObject{}
					auxElt.NextElt = dts.NFAAttributes.EquivList
					auxElt.StateElt = i
					dts.NFAAttributes.EquivList = auxElt
				}
			}
			dts.Transitions = nil
			dts.Marked = false
			dts.PatternStart = false
			dts.LeftTransition = false
			dts.RightTransition = false
			dts.LeftContextCheck = false
			dts.FinalAccept = false
		} else {
			ScreenMessage(MsgPatPatternTooComplex)
			return false
		}
		*stateCount = statesUsed
		return true
	}

	// dfaSearch finds the position in DFA_table of a state with given NFA equivalent
	dfaSearch := func(stateHead [MaxNFAStateRange + 1]bool, position *int) bool {
		for i := 0; i <= statesUsed; i++ {
			if bitsetEqualsNFA(stateHead, dfaTablePointer.DFATable[i].NFAAttributes.EquivSet) {
				*position = i
				return true
			}
		}
		return false
	}

	// patternAddDFA adds a DFA transition
	patternAddDFA := func(
		transferState NFAAttributeType,
		acceptSet *big.Int,
		fromState int,
	) bool {
		var position int
		if !dfaSearch(transferState.EquivSet, &position) {
			if !patternNewDFA(&transferState, &position) {
				return false
			}
		}
		dtf := &dfaTablePointer.DFATable[fromState]
		auxTransition := &TransitionObject{}
		auxTransition.NextTransition = dtf.Transitions
		dtf.Transitions = auxTransition
		auxTransition.TransitionAcceptSet.Set(acceptSet)
		auxTransition.AcceptNextState = position
		auxTransition.StartFlag = false
		return true
	}

	// unmarkedStates finds an unmarked state
	unmarkedStates := func(unmarkedState *int) bool {
		for i := *dfaStart; i <= statesUsed; i++ {
			if !dfaTablePointer.DFATable[i].Marked {
				*unmarkedState = i
				return true
			}
		}
		return false
	}

	ExitAbort = true // true in case we blow the dfa table or something

	// Initialize DFA kill state
	dtk := &dfaTablePointer.DFATable[PatternDFAKill]
	dtk.Transitions = nil
	dtk.Marked = true
	bitsetClear(&dtk.NFAAttributes.EquivSet)
	dtk.PatternStart = false
	dtk.LeftTransition = false
	dtk.RightTransition = false
	dtk.LeftContextCheck = false
	dtk.FinalAccept = false

	// Initialize DFA fail state
	dtfail := &dfaTablePointer.DFATable[PatternDFAFail]
	dtfail.Transitions = nil
	dtfail.Marked = true
	dtfail.NFAAttributes.EquivSet[PatternDFAFail] = true
	dtfail.PatternStart = false
	dtfail.LeftTransition = false
	dtfail.RightTransition = false
	dtfail.LeftContextCheck = false
	dtfail.FinalAccept = false

	// Build initial state
	statesUsed = 1
	auxElt = &StateEltObject{}
	auxElt.NextElt = nil
	auxElt.StateElt = nfaStart
	transitionSet.EquivList = auxElt
	transitionSet.EquivSet[nfaStart] = true
	if !epsilonClosures(nfaTable, &transitionSet, &auxClosure) {
		return false
	}
	if !patternNewDFA(&auxClosure, dfaStart) {
		return false
	}

	for unmarkedStates(&currentState) {
		if TtControlC {
			dfaTablePointer.Definition.Length = 0 // invalidate the table
			return false
		}
		bitsetSetRange(&killSet, 0, MaxSetRange)
		partitionPtr = nil
		dtc := &dfaTablePointer.DFATable[currentState]
		dtc.Marked = true
		auxEquivPtr = dtc.NFAAttributes.EquivList

		for auxEquivPtr != nil {
			nta := &nfaTable[auxEquivPtr.StateElt]
			if !nta.EpsilonOut {
				auxPartitionPtr = partitionPtr
				partitionPtr = &acceptSetPartitionType{}
				partitionPtr.acceptSetPartition.Set(&nta.AcceptSet)
				bitsetRemove(&killSet, &nta.AcceptSet)
				partitionPtr.flink = auxPartitionPtr
				partitionPtr.blink = nil
				if partitionPtr.flink != nil {
					partitionPtr.flink.blink = partitionPtr
				}
				partitionPtr.nfaTransitionList.EquivList = &StateEltObject{}
				partitionPtr.nfaTransitionList.EquivList.NextElt = nil
				partitionPtr.nfaTransitionList.EquivList.StateElt = nta.NextState
			}
			auxEquivPtr = auxEquivPtr.NextElt
		}

		// Partition the list
		if partitionPtr != nil && partitionPtr.flink != nil {
			currentPartitionPtr = partitionPtr
			followerPtr = currentPartitionPtr.flink
			for currentPartitionPtr != nil {
				if followerPtr == currentPartitionPtr {
					followerPtr = followerPtr.flink
				}
				auxPartitionPtr = followerPtr
				for auxPartitionPtr != nil {
					if bitsetEquals(
						&currentPartitionPtr.acceptSetPartition,
						&auxPartitionPtr.acceptSetPartition,
					) {
						// merge entries
						auxStatePtr = currentPartitionPtr.nfaTransitionList.EquivList
						for auxStatePtr.NextElt != nil {
							auxStatePtr = auxStatePtr.NextElt
						}
						auxStatePtr.NextElt = auxPartitionPtr.nfaTransitionList.EquivList
						// remove aux entry
						auxPartitionPtr.blink.flink = auxPartitionPtr.flink
						if auxPartitionPtr.flink != nil {
							auxPartitionPtr.flink.blink = auxPartitionPtr.blink
						}
						if followerPtr == auxPartitionPtr {
							followerPtr = auxPartitionPtr.flink
						}
						auxPartitionPtr = auxPartitionPtr.flink
					} else {
						// form partition
						intersectionSet := bitsetIntersection(
							&currentPartitionPtr.acceptSetPartition,
							&auxPartitionPtr.acceptSetPartition,
						)
						if !bitsetIsEmptyAccept(intersectionSet) {
							if bitsetEquals(intersectionSet, &currentPartitionPtr.acceptSetPartition) {
								currentPartitionPtr.nfaTransitionList.EquivList = transitionListAppend(
									currentPartitionPtr.nfaTransitionList.EquivList,
									auxPartitionPtr.nfaTransitionList.EquivList,
								)
								bitsetRemove(&auxPartitionPtr.acceptSetPartition, intersectionSet)
							} else if bitsetEquals(intersectionSet, &auxPartitionPtr.acceptSetPartition) {
								auxPartitionPtr.nfaTransitionList.EquivList = transitionListAppend(
									auxPartitionPtr.nfaTransitionList.EquivList,
									currentPartitionPtr.nfaTransitionList.EquivList,
								)
								bitsetRemove(&currentPartitionPtr.acceptSetPartition, intersectionSet)
							} else {
								// need to do a full partition
								insertPartition = &acceptSetPartitionType{}
								insertPartition.acceptSetPartition.Set(intersectionSet)
								insertPartition.flink = followerPtr
								insertPartition.blink = followerPtr.blink
								insertPartition.flink.blink = insertPartition
								insertPartition.blink.flink = insertPartition
								insertPartition.nfaTransitionList.EquivList = transitionListMerge(
									currentPartitionPtr.nfaTransitionList.EquivList,
									auxPartitionPtr.nfaTransitionList.EquivList,
								)
								bitsetRemove(&currentPartitionPtr.acceptSetPartition, intersectionSet)
								bitsetRemove(&auxPartitionPtr.acceptSetPartition, intersectionSet)
							}
						}
						auxPartitionPtr = auxPartitionPtr.flink
					}
				}
				currentPartitionPtr = currentPartitionPtr.flink
			}
		}

		// Form DFA using partitioned list
		dtc2 := &dfaTablePointer.DFATable[currentState]
		dtc2.Transitions = &TransitionObject{}
		dtc2.Transitions.AcceptNextState = PatternDFAKill
		dtc2.Transitions.StartFlag = false
		dtc2.Transitions.NextTransition = nil
		dtc2.Transitions.TransitionAcceptSet.Set(&killSet)

		for partitionPtr != nil {
			auxPartitionPtr = partitionPtr
			if !epsilonClosures(nfaTable, &auxPartitionPtr.nfaTransitionList, &transferState) {
				return false
			}
			if !patternAddDFA(transferState, &auxPartitionPtr.acceptSetPartition, currentState) {
				return false
			}
			partitionPtr = auxPartitionPtr.flink
			auxPartitionPtr = nil
		}
	}

	// END OF DFA GENERATION
	// Now fix it up so it will drive the recognizer

	// Find all final states
	for auxCount = 0; auxCount <= statesUsed; auxCount++ {
		if dfaTablePointer.DFATable[auxCount].NFAAttributes.EquivSet[*nfaEnd] {
			dfaTablePointer.DFATable[auxCount].FinalAccept = true
		}
	}

	// Start pattern flag creation
	incomingTranPtr = dfaTablePointer.DFATable[PatternDFAStart].Transitions
	for incomingTranPtr != nil {
		if (incomingTranPtr.AcceptNextState != PatternDFAKill) &&
			(incomingTranPtr.AcceptNextState != PatternDFAFail) &&
			!dfaTablePointer.DFATable[incomingTranPtr.AcceptNextState].FinalAccept {
			dtans := &dfaTablePointer.DFATable[incomingTranPtr.AcceptNextState]
			dtans.PatternStart = true
			killTranPtr = dtans.Transitions
			for (killTranPtr != nil) && (killTranPtr.AcceptNextState != PatternDFAKill) {
				killTranPtr = killTranPtr.NextTransition
			}
			if killTranPtr != nil {
				auxTransitionSet := bitsetIntersection(
					&incomingTranPtr.TransitionAcceptSet,
					&killTranPtr.TransitionAcceptSet,
				)
				if !bitsetIsEmptyAccept(auxTransitionSet) {
					auxTranPtr = &TransitionObject{}
					auxTranPtr.TransitionAcceptSet.Set(auxTransitionSet)
					auxTranPtr.AcceptNextState = incomingTranPtr.AcceptNextState
					auxTranPtr.NextTransition = nil
					auxTranPtr.StartFlag = true
					bitsetRemove(&killTranPtr.TransitionAcceptSet, &auxTranPtr.TransitionAcceptSet)
					killTranPtr.NextTransition = auxTranPtr
				}
			}
		}
		incomingTranPtr = incomingTranPtr.NextTransition
	}

	// Find all end of left context states
	if !epsilonAndMask(nfaTable, middleContextStart, &closureSet, &mask, true) {
		return false
	}
	for auxCount = PatternDFAStart; auxCount <= statesUsed; auxCount++ {
		dtac := &dfaTablePointer.DFATable[auxCount]
		if dtac.NFAAttributes.EquivSet[middleContextStart] &&
			bitsetIsEmpty(bitsetDifference(dtac.NFAAttributes.EquivSet, mask)) {
			dtac.LeftTransition = true
		}
	}

	// Create auxSet for middle context range
	for i := 0; i <= MaxNFAStateRange; i++ {
		auxSet[i] = false
	}
	for i := middleContextStart; i <= rightContextStart && i <= MaxNFAStateRange; i++ {
		if closureSet[i] {
			auxSet[i] = true
		}
	}

	for auxCount2 = PatternDFAStart; auxCount2 <= statesUsed; auxCount2++ {
		if dfaTablePointer.DFATable[auxCount2].LeftTransition {
			auxTranPtr2 = dfaTablePointer.DFATable[auxCount2].Transitions
			for auxTranPtr2 != nil {
				state = auxTranPtr2.AcceptNextState
				if state > auxCount2 {
					auxTranPtr = dfaTablePointer.DFATable[state].Transitions
					found = false
					for auxTranPtr != nil && !found {
						if auxTranPtr.AcceptNextState == state {
							found = true
						} else {
							auxTranPtr = auxTranPtr.NextTransition
						}
					}
					if found {
						dfaTablePointer.DFATable[state].LeftContextCheck = true
						for auxCount = middleContextStart; auxCount <= rightContextStart; auxCount++ {
							if nfaTable[auxCount].Indefinite &&
								auxSet[auxCount] &&
								dfaTablePointer.DFATable[state].NFAAttributes.EquivSet[auxCount] {
								dfaTablePointer.DFATable[state].LeftContextCheck = false
							}
						}
					}
				}
				auxTranPtr2 = auxTranPtr2.NextTransition
			}
		}
	}

	// Find all end of middle context states
	if !epsilonAndMask(nfaTable, rightContextStart, &closureSet, &mask, true) {
		return false
	}
	for auxCount = 0; auxCount <= statesUsed; auxCount++ {
		dtac := &dfaTablePointer.DFATable[auxCount]
		if dtac.NFAAttributes.EquivSet[rightContextStart] &&
			bitsetIsEmpty(bitsetDifference(dtac.NFAAttributes.EquivSet, mask)) {
			dtac.RightTransition = true
		}
	}

	*dfaEnd = statesUsed
	dfaTablePointer.DFAStatesUsed = statesUsed

	ExitAbort = false
	return true
}
