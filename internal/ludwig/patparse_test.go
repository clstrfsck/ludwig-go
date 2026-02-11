// Tests for functions in patparse.go

package ludwig

import (
	"math/big"
	"math/bits"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper functions for testing

// OnesCount returns the number of set bits in a big.Int
func OnesCount(x *big.Int) int {
	count := 0
	for _, word := range x.Bits() {
		count += bits.OnesCount(uint(word))
	}
	return count
}

// createTestTpar creates a TParObject for testing
func createTestTpar(content string) *TParObject {
	tpar := &TParObject{
		Str: NewStrObjectFrom(content),
		Len: len(content),
		Dlm: TpdLit,
	}
	return tpar
}

// Tests for helper functions

func TestSingletonSet(t *testing.T) {
	t.Run("CreateSingletonSet", func(t *testing.T) {
		set := singletonSet('a')
		assert.NotNil(t, set, "Set should not be nil")
		assert.Equal(t, 1, OnesCount(set), "Only one bit should be set in the singleton set")
		assert.Equal(t, uint(1), set.Bit(int('a')), "Bit for 'a' should be set")
		assert.Equal(t, uint(0), set.Bit(int('b')), "Bit for 'b' should not be set")
	})

	t.Run("SingletonSetDifferentChars", func(t *testing.T) {
		setA := singletonSet('A')
		setZ := singletonSet('Z')
		set0 := singletonSet('0')

		assert.Equal(t, uint(1), setA.Bit(int('A')), "Bit for 'A' should be set")
		assert.Equal(t, 1, OnesCount(setA), "Only one bit should be set in the singleton set")
		assert.Equal(t, uint(1), setZ.Bit(int('Z')), "Bit for 'Z' should be set")
		assert.Equal(t, 1, OnesCount(setZ), "Only one bit should be set in the singleton set")
		assert.Equal(t, uint(1), set0.Bit(int('0')), "Bit for '0' should be set")
		assert.Equal(t, 1, OnesCount(set0), "Only one bit should be set in the singleton set")

		assert.Equal(t, uint(0), setA.Bit(int('Z')), "Bit for 'Z' should not be in setA")
	})
}

func TestRangeSet(t *testing.T) {
	t.Run("CreateRangeAtoZ", func(t *testing.T) {
		set := rangeSet('a', 'z')
		assert.NotNil(t, set, "Set should not be nil")

		// Check all lowercase letters are in the set
		assert.Equal(t, 26, OnesCount(set), "There should be 26 bits set for letters a-z")
		for ch := byte('a'); ch <= byte('z'); ch++ {
			assert.Equal(t, uint(1), set.Bit(int(ch)), "Bit for '%c' should be set", ch)
		}

		// Check uppercase letters are not in the set
		assert.Equal(t, uint(0), set.Bit(int('A')), "Bit for 'A' should not be set")
	})

	t.Run("SingleCharRange", func(t *testing.T) {
		set := rangeSet('x', 'x')
		assert.Equal(t, 1, OnesCount(set), "Only one bit should be set for a single char range")
		assert.Equal(t, uint(1), set.Bit(int('x')), "Bit for 'x' should be set")
		assert.Equal(t, uint(0), set.Bit(int('y')), "Bit for 'y' should not be set")
	})
}

func TestSetAdd(t *testing.T) {
	t.Run("AddToEmptySet", func(t *testing.T) {
		set := new(big.Int)
		setAdd(set, 'x')

		assert.Equal(t, 1, OnesCount(set), "Only one bit should be set after adding a character")
		assert.Equal(t, uint(1), set.Bit(int('x')), "Bit for 'x' should be set")
		assert.Equal(t, uint(0), set.Bit(int('y')), "Bit for 'y' should not be set")
	})

	t.Run("AddMultipleChars", func(t *testing.T) {
		set := new(big.Int)
		setAdd(set, 'a')
		setAdd(set, 'b')
		setAdd(set, 'c')

		assert.Equal(t, 3, OnesCount(set), "Three bits should be set after adding three characters")
		assert.Equal(t, uint(1), set.Bit(int('a')), "Bit for 'a' should be set")
		assert.Equal(t, uint(1), set.Bit(int('b')), "Bit for 'b' should be set")
		assert.Equal(t, uint(1), set.Bit(int('c')), "Bit for 'c' should be set")
	})
}

func TestSetAddRange(t *testing.T) {
	t.Run("AddRangeToEmptySet", func(t *testing.T) {
		set := new(big.Int)
		setAddRange(set, 'a', 'e')

		assert.Equal(t, 5, OnesCount(set), "Five bits should be set after adding five characters")
		for ch := byte('a'); ch <= byte('e'); ch++ {
			assert.Equal(t, uint(1), set.Bit(int(ch)), "Bit for '%c' should be set", ch)
		}
		assert.Equal(t, uint(0), set.Bit(int('f')), "Bit for 'f' should not be set")
	})

	t.Run("AddMultipleRanges", func(t *testing.T) {
		set := new(big.Int)
		setAddRange(set, 'a', 'c')
		setAddRange(set, 'x', 'z')

		assert.Equal(t, 3+3, OnesCount(set), "Six bits should be set after adding two ranges")
		assert.Equal(t, uint(1), set.Bit(int('a')), "Bit for 'a' should be set")
		assert.Equal(t, uint(1), set.Bit(int('b')), "Bit for 'b' should be set")
		assert.Equal(t, uint(1), set.Bit(int('c')), "Bit for 'c' should be set")
		assert.Equal(t, uint(0), set.Bit(int('m')), "Bit for 'm' should not be set")
		assert.Equal(t, uint(1), set.Bit(int('x')), "Bit for 'x' should be set")
		assert.Equal(t, uint(1), set.Bit(int('y')), "Bit for 'y' should be set")
		assert.Equal(t, uint(1), set.Bit(int('z')), "Bit for 'z' should be set")
	})
}

func TestSetClear(t *testing.T) {
	t.Run("ClearNonEmptySet", func(t *testing.T) {
		set := rangeSet('a', 'z')
		assert.Equal(t, uint(1), set.Bit(int('a')), "Bit for 'a' should be set initially")

		setClear(set)

		assert.Equal(t, 0, OnesCount(set), "All bits should be cleared after setClear")
		assert.Equal(t, uint(0), set.Bit(int('a')), "Bit for 'a' should be cleared")
		assert.Equal(t, uint(0), set.Bit(int('m')), "Bit for 'm' should be cleared")
		assert.Equal(t, uint(0), set.Bit(int('z')), "Bit for 'z' should be cleared")
	})
}

func TestSetUnion(t *testing.T) {
	t.Run("UnionOfTwoRanges", func(t *testing.T) {
		setA := rangeSet('a', 'c')
		setB := rangeSet('x', 'z')

		result := setUnion(setA, setB)

		assert.Equal(t, 3+3, OnesCount(result), "Six bits should be set in the union of two ranges")
		assert.Equal(t, uint(1), result.Bit(int('a')), "Bit for 'a' should be set")
		assert.Equal(t, uint(1), result.Bit(int('b')), "Bit for 'b' should be set")
		assert.Equal(t, uint(1), result.Bit(int('c')), "Bit for 'c' should be set")
		assert.Equal(t, uint(0), result.Bit(int('m')), "Bit for 'm' should not be set")
		assert.Equal(t, uint(1), result.Bit(int('x')), "Bit for 'x' should be set")
		assert.Equal(t, uint(1), result.Bit(int('y')), "Bit for 'y' should be set")
		assert.Equal(t, uint(1), result.Bit(int('z')), "Bit for 'z' should be set")
	})

	t.Run("UnionWithEmptySet", func(t *testing.T) {
		setA := rangeSet('a', 'c')
		setB := new(big.Int)

		result := setUnion(setA, setB)

		assert.Equal(t, 3, OnesCount(result), "Three bits should be set when union with empty set")
		assert.Equal(t, uint(1), result.Bit(int('a')), "Bit for 'a' should be set")
		assert.Equal(t, uint(1), result.Bit(int('b')), "Bit for 'b' should be set")
		assert.Equal(t, uint(1), result.Bit(int('c')), "Bit for 'c' should be set")
	})

	t.Run("UnionWithOverlap", func(t *testing.T) {
		setA := rangeSet('a', 'e')
		setB := rangeSet('c', 'g')

		result := setUnion(setA, setB)

		assert.Equal(t, 7, OnesCount(result), "Seven bits should be set in the union with overlap")
		for ch := byte('a'); ch <= byte('g'); ch++ {
			assert.Equal(t, uint(1), result.Bit(int(ch)), "Bit for '%c' should be set", ch)
		}
	})
}

func TestSetRemove(t *testing.T) {
	t.Run("RemoveSubset", func(t *testing.T) {
		setA := rangeSet('a', 'z')
		setB := rangeSet('m', 'p')

		result := setRemove(setA, setB)

		assert.Equal(t, 26-4, OnesCount(result), "Four bits should be removed from the original set")
		assert.Equal(t, uint(1), result.Bit(int('a')), "Bit for 'a' should still be set")
		assert.Equal(t, uint(1), result.Bit(int('l')), "Bit for 'l' should still be set")
		assert.Equal(t, uint(0), result.Bit(int('m')), "Bit for 'm' should be removed")
		assert.Equal(t, uint(0), result.Bit(int('n')), "Bit for 'n' should be removed")
		assert.Equal(t, uint(0), result.Bit(int('o')), "Bit for 'o' should be removed")
		assert.Equal(t, uint(0), result.Bit(int('p')), "Bit for 'p' should be removed")
		assert.Equal(t, uint(1), result.Bit(int('q')), "Bit for 'q' should still be set")
		assert.Equal(t, uint(1), result.Bit(int('z')), "Bit for 'z' should still be set")
	})

	t.Run("RemoveDisjointSet", func(t *testing.T) {
		setA := rangeSet('a', 'e')
		setB := rangeSet('x', 'z')

		result := setRemove(setA, setB)

		for ch := byte('a'); ch <= byte('e'); ch++ {
			assert.Equal(t, uint(1), result.Bit(int(ch)), "Bit for '%c' should still be set", ch)
		}
	})

	t.Run("RemoveFromEmptySet", func(t *testing.T) {
		setA := new(big.Int)
		setB := rangeSet('a', 'z')

		result := setRemove(setA, setB)

		assert.Equal(t, uint(0), result.Bit(int('a')), "Result should be empty")
		assert.Equal(t, uint(0), result.Bit(int('m')), "Result should be empty")
	})
}

// Tests for predefined character sets

func TestPredefinedCharacterSets(t *testing.T) {
	t.Run("SpaceSet", func(t *testing.T) {
		assert.Equal(t, uint(1), spaceSet.Bit(int(' ')), "Space should be in spaceSet")
		assert.Equal(t, uint(0), spaceSet.Bit(int('a')), "'a' should not be in spaceSet")
		assert.Equal(t, uint(0), spaceSet.Bit(int('\t')), "Tab should not be in spaceSet")
	})

	t.Run("LowerSet", func(t *testing.T) {
		for ch := byte('a'); ch <= byte('z'); ch++ {
			assert.Equal(t, uint(1), lowerSet.Bit(int(ch)), "'%c' should be in lowerSet", ch)
		}
		assert.Equal(t, uint(0), lowerSet.Bit(int('A')), "'A' should not be in lowerSet")
		assert.Equal(t, uint(0), lowerSet.Bit(int('0')), "'0' should not be in lowerSet")
	})

	t.Run("UpperSet", func(t *testing.T) {
		for ch := byte('A'); ch <= byte('Z'); ch++ {
			assert.Equal(t, uint(1), upperSet.Bit(int(ch)), "'%c' should be in upperSet", ch)
		}
		assert.Equal(t, uint(0), upperSet.Bit(int('a')), "'a' should not be in upperSet")
		assert.Equal(t, uint(0), upperSet.Bit(int('0')), "'0' should not be in upperSet")
	})

	t.Run("AlphaSet", func(t *testing.T) {
		for ch := byte('a'); ch <= byte('z'); ch++ {
			assert.Equal(t, uint(1), alphaSet.Bit(int(ch)), "'%c' should be in alphaSet", ch)
		}
		for ch := byte('A'); ch <= byte('Z'); ch++ {
			assert.Equal(t, uint(1), alphaSet.Bit(int(ch)), "'%c' should be in alphaSet", ch)
		}
		assert.Equal(t, uint(0), alphaSet.Bit(int('0')), "'0' should not be in alphaSet")
		assert.Equal(t, uint(0), alphaSet.Bit(int(' ')), "Space should not be in alphaSet")
	})

	t.Run("NumericSet", func(t *testing.T) {
		for ch := byte('0'); ch <= byte('9'); ch++ {
			assert.Equal(t, uint(1), numericSet.Bit(int(ch)), "'%c' should be in numericSet", ch)
		}
		assert.Equal(t, uint(0), numericSet.Bit(int('a')), "'a' should not be in numericSet")
	})

	t.Run("PrintableSet", func(t *testing.T) {
		// Check space through tilde (32-126)
		for i := 32; i <= 126; i++ {
			assert.Equal(t, uint(1), printableSet.Bit(i), "'%c' (%d) should be in printableSet", byte(i), i)
		}
		assert.Equal(t, uint(0), printableSet.Bit(int(31)), "Control char should not be in printableSet")
		assert.Equal(t, uint(0), printableSet.Bit(int(127)), "DEL should not be in printableSet")
	})

	t.Run("PunctuationSet", func(t *testing.T) {
		punctChars := []byte{33, 34, 39, 40, 41, 44, 46, 58, 59, 63, 96} // !"'(),.:;?`
		for _, ch := range punctChars {
			assert.Equal(t, uint(1), punctuationSet.Bit(int(ch)), "'%c' should be in punctuationSet", ch)
		}
		assert.Equal(t, uint(0), punctuationSet.Bit(int('a')), "'a' should not be in punctuationSet")
		assert.Equal(t, uint(0), punctuationSet.Bit(int('0')), "'0' should not be in punctuationSet")
	})
}

// Tests for static accept sets

func TestStaticAcceptSets(t *testing.T) {
	t.Run("QuotedSet", func(t *testing.T) {
		assert.True(t, quotedSet[TpdLit], "TpdLit should be in quotedSet")
		assert.True(t, quotedSet[TpdExact], "TpdExact should be in quotedSet")
		assert.False(t, quotedSet['a'], "'a' should not be in quotedSet")
	})

	t.Run("DelimitedSet", func(t *testing.T) {
		assert.True(t, delimitedSet[PatternKStar], "* should be in delimitedSet")
		assert.True(t, delimitedSet[PatternPlus], "+ should be in delimitedSet")
		assert.True(t, delimitedSet[PatternLRangeDelim], "[ should be in delimitedSet")
		for ch := byte('0'); ch <= byte('9'); ch++ {
			assert.True(t, delimitedSet[ch], "'%c' should be in delimitedSet", ch)
		}
		assert.False(t, delimitedSet['a'], "'a' should not be in delimitedSet")
	})

	t.Run("CharsetsSet", func(t *testing.T) {
		charsetChars := []byte{'s', 'S', 'a', 'A', 'c', 'C', 'l', 'L', 'u', 'U', 'n', 'N', 'p', 'P'}
		for _, ch := range charsetChars {
			assert.True(t, charsetsSet[ch], "'%c' should be in charsetsSet", ch)
		}
		assert.False(t, charsetsSet['x'], "'x' should not be in charsetsSet")
	})

	t.Run("PositionalsSet", func(t *testing.T) {
		posChars := []byte{'<', '>', '{', '}', '^'}
		for _, ch := range posChars {
			assert.True(t, positionalsSet[ch], "'%c' should be in positionalsSet", ch)
		}
		assert.False(t, positionalsSet['a'], "'a' should not be in positionalsSet")
	})

	t.Run("ChAndPosSet", func(t *testing.T) {
		// Should be union of charsetsSet and positionalsSet
		assert.True(t, chAndPosSet['s'], "'s' should be in chAndPosSet")
		assert.True(t, chAndPosSet['<'], "'<' should be in chAndPosSet")
		assert.False(t, chAndPosSet['x'], "'x' should not be in chAndPosSet")
	})

	t.Run("SyntaxSet", func(t *testing.T) {
		// Test a few key syntax characters
		assert.True(t, syntaxSet[PatternLParen], "'(' should be in syntaxSet")
		assert.True(t, syntaxSet[PatternKStar], "'*' should be in syntaxSet")
		assert.True(t, syntaxSet[PatternPlus], "'+' should be in syntaxSet")
		assert.True(t, syntaxSet['a'], "'a' should be in syntaxSet")
		assert.True(t, syntaxSet['0'], "'0' should be in syntaxSet")
		assert.True(t, syntaxSet['<'], "'<' should be in syntaxSet")
	})
}

// Tests for PatternParser function

func TestPatternParser(t *testing.T) {
	// Helper to create test environment
	setupParser := func() (*NFATableType, *PatternDefType) {
		nfaTable := &NFATableType{}
		patternDef := &PatternDefType{
			Strng: *NewBlankStrObject(MaxStrLen),
		}
		return nfaTable, patternDef
	}

	t.Run("EmptyPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.False(t, result, "Empty pattern should fail")
	})

	t.Run("SimpleLiteralPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("'abc'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Simple literal pattern should succeed")
		assert.Greater(t, statesUsed, PatternNFAStart, "Should have allocated NFA states")
		assert.Greater(t, patternDef.Length, 0, "Pattern definition should have content")
	})

	t.Run("ExactCasePattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar(`"test"`)

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Exact case pattern should succeed")
	})

	t.Run("CharacterClassSpace", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("s")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Space character class should succeed")
		// Verify the accept set contains space
		foundSpace := false
		for i := PatternNFAStart; i < statesUsed; i++ {
			if !nfaTable[i].EpsilonOut && nfaTable[i].AcceptSet.Bit(int(' ')) == 1 {
				foundSpace = true
				break
			}
		}
		assert.True(t, foundSpace, "Accept set should contain space character")
	})

	t.Run("CharacterClassLowercase", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("l")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Lowercase character class should succeed")
	})

	t.Run("CharacterClassUppercase", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("u")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Uppercase character class should succeed")
	})

	t.Run("CharacterClassAlpha", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("a")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Alpha character class should succeed")
	})

	t.Run("CharacterClassNumeric", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("n")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Numeric character class should succeed")
	})

	t.Run("CharacterClassPunctuation", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("p")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Punctuation character class should succeed")
	})

	t.Run("NegatedCharacterClass", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("-n")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Negated character class should succeed")
	})

	t.Run("KleeneStarRepetition", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before the element: *'a' not 'a'*
		pattern := createTestTpar("*'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Kleene star should succeed")
	})

	t.Run("KleenePlusRepetition", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before the element: +'a' not 'a'+
		pattern := createTestTpar("+'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Kleene plus should succeed")
	})

	t.Run("ExactCountRepetition", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before the element: 3'a' not 'a'3
		pattern := createTestTpar("3'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Exact count repetition should succeed")
	})

	t.Run("RangeRepetition", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before the element: [2,5]'a' not 'a'[2,5]
		pattern := createTestTpar("[2,5]'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Range repetition should succeed")
	})

	t.Run("OpenEndedRange", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before the element: [2,]'a' not 'a'[2,]
		pattern := createTestTpar("[2,]'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Open-ended range should succeed")
	})

	t.Run("BeginningOfLinePositional", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("<")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Beginning of line positional should succeed")
	})

	t.Run("EndOfLinePositional", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar(">")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "End of line positional should succeed")
	})

	t.Run("LeftMarginPositional", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("{")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Left margin positional should succeed")
	})

	t.Run("RightMarginPositional", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("}")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Right margin positional should succeed")
	})

	t.Run("DotColumnPositional", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("^")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Dot column positional should succeed")
	})

	t.Run("AlternationPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("'a'|'b'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Alternation pattern should succeed")
	})

	t.Run("GroupedPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("('a''b')")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Grouped pattern should succeed")
	})

	t.Run("ComplexPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Pattern: one or more alphas, then space, then one or more numerics
		// Repetition comes before: +a+s+n not a+s+n+
		pattern := createTestTpar("+a+s+n")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Complex pattern should succeed")
	})

	t.Run("ContextPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Pattern with left context, middle, and right context
		pattern := createTestTpar("'a','b','c'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Context pattern should succeed")
		assert.NotEqual(t, leftEnd, firstStart, "leftEnd should be different from start")
		assert.NotEqual(t, middleEnd, leftEnd, "middleEnd should be different from leftEnd")
	})

	t.Run("InvalidCharacterInPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Using a character that's not in syntaxSet
		pattern := createTestTpar("#")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.False(t, result, "Invalid character should fail")
	})
}

// Edge case tests

func TestPatternParserEdgeCases(t *testing.T) {
	setupParser := func() (*NFATableType, *PatternDefType) {
		nfaTable := &NFATableType{}
		patternDef := &PatternDefType{
			Strng: *NewBlankStrObject(MaxStrLen),
		}
		return nfaTable, patternDef
	}

	t.Run("SpacesInPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("  'a'  ")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Pattern with spaces should succeed")
	})

	t.Run("NestedGroups", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("(('a''b'))")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Nested groups should succeed")
	})

	t.Run("MultipleAlternations", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("'a'|'b'|'c'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Multiple alternations should succeed")
	})

	t.Run("ZeroRepetitionRange", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before: [0,3]'a' not 'a'[0,3]
		pattern := createTestTpar("[0,3]'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Zero repetition range should succeed")
	})

	t.Run("SingleRepetitionRange", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Repetition comes before: [1,1]'a' not 'a'[1,1]
		pattern := createTestTpar("[1,1]'a'")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Single repetition range should succeed")
	})

	t.Run("CombinedCharacterClasses", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		pattern := createTestTpar("alsn")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Combined character classes should succeed")
	})

	t.Run("UppercaseCharacterClasses", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Uppercase versions should work the same as lowercase
		pattern := createTestTpar("ALSN")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Uppercase character classes should succeed")
	})
}

// Integration tests

func TestPatternParserIntegration(t *testing.T) {
	setupParser := func() (*NFATableType, *PatternDefType) {
		nfaTable := &NFATableType{}
		patternDef := &PatternDefType{
			Strng: *NewBlankStrObject(MaxStrLen),
		}
		return nfaTable, patternDef
	}

	t.Run("EmailPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Simple pattern: one or more alphas, @, one or more alphas, ., one or more alphas
		// Repetition comes before: +a'@'+a'.'+a
		pattern := createTestTpar("+a'@'+a'.'+a")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Email-like pattern should succeed")
	})

	t.Run("WordPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Pattern: beginning of line, one or more alphas, end of line
		// Repetition comes before: <+a>
		pattern := createTestTpar("<+a>")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Word pattern with anchors should succeed")
	})

	t.Run("NumberPattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Pattern: optional plus/minus, one or more digits
		// Repetition comes before: ('+'|'-'|)+n
		pattern := createTestTpar("('+'|'-'|)+n")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Number pattern should succeed")
	})

	t.Run("WhitespacePattern", func(t *testing.T) {
		nfaTable, patternDef := setupParser()
		// Pattern: one or more spaces
		// Repetition comes before: +s not s+
		pattern := createTestTpar("+s")

		var firstStart, finalState, leftEnd, middleEnd, statesUsed int

		result := PatternParser(pattern, nfaTable, &firstStart, &finalState, &leftEnd, &middleEnd, patternDef, &statesUsed)

		assert.True(t, result, "Whitespace pattern should succeed")
	})
}
