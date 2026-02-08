/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         CH
//
// Description:  Character array handling routines.

package ludwig

import (
	"unicode"
)

// sgn returns the sign of a value: -1, 0, or 1
func sgn(val int) int {
	if val < 0 {
		return -1
	} else if val > 0 {
		return 1
	}
	return 0
}

// ChFillCopy is a wrapper for array copy that handles copies with fill character.
// It copies srclen bytes from src starting at srcofs to dst starting at dstofs,
// filling the remaining dstlen bytes with the fill character if srclen < dstlen.
func ChFillCopy(
	src *StrObject,
	srcofs int,
	srclen int,
	dst *StrObject,
	dstofs int,
	dstlen int,
	fill byte,
) {
	if dstlen > 0 {
		dst.FillCopy(src, srcofs, srclen, dstofs, dstlen, fill)
	}
}

// ChCompareStr compares two string regions and returns their ordering.
// Returns -1 if target < text, 0 if equal, 1 if target > text.
// The nchIdent output parameter receives the number of identical characters.
// If exactcase is true, the comparison is case-sensitive.
func ChCompareStr(
	target *StrObject,
	st1 int,
	len1 int,
	text *StrObject,
	st2 int,
	len2 int,
	exactcase bool,
	nchIdent *int,
) int {
	i := 0
	if exactcase {
		// Exact case comparison
		for i < len1 && i < len2 {
			if target.Get(st1+i) != text.Get(st2+i) {
				break
			}
			i++
		}
	} else {
		// Case-insensitive comparison (uppercase text characters)
		for i < len1 && i < len2 {
			targetCh := target.Get(st1 + i)
			textCh := ChToUpper(text.Get(st2 + i))
			if targetCh != textCh {
				break
			}
			i++
		}
	}
	*nchIdent = i

	var diff int
	if i < len1 && i < len2 {
		ch1 := target.Get(st1 + i)
		ch2 := text.Get(st2 + i)
		if !exactcase {
			ch2 = ChToUpper(ch2)
		}
		diff = sgn(int(ch1) - int(ch2))
	} else {
		diff = sgn(len1 - len2)
	}
	return diff
}

// ChReverseStr reverses len bytes from src and stores them in dst.
// If src and dst are the same, it reverses in place.
func ChReverseStr(src *StrObject, dst *StrObject, len int) {
	half := (len + 1) / 2
	if half > 0 {
		for i := range half {
			ch := src.Get(i + 1)
			dst.Set(i+1, src.Get(len-i))
			dst.Set(len-i, ch)
		}
	}
}

func ChIsPrintable(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsPrint(ch)
}

func ChIsSpace(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsSpace(ch)
}

func ChIsLetter(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsLetter(ch)
}

func ChIsLower(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsLower(ch)
}

func ChIsUpper(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsUpper(ch)
}

func ChIsNumeric(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	return unicode.IsNumber(ch)
}

func ChIsPunctuation(ch rune) bool {
	if ch < 0 || ch > MaxSetRange {
		return false
	}
	// Broad definition of punctuation.
	return ChIsPrintable(ch) && !ChIsLetter(ch) && !ChIsNumeric(ch) && !ChIsSpace(ch)
}

func ChIsWordElement(set int, ch rune) bool {
	switch set {
	case 0:
		return ChIsSpace(ch)
	case 1:
		return ChIsPrintable(ch) && !ChIsSpace(ch)
	default:
		return false
	}
}

func ChKeyToUpper(key int) int {
	if key >= 0 && key <= MaxSetRange {
		return int(ChToUpper(byte(key)))
	}
	return key
}

// ChToUpper converts a character to uppercase.
func ChToUpper(ch byte) byte {
	return byte(unicode.ToUpper(rune(ch)))
}

// ChToLower converts a character to lowercase.
func ChToLower(ch byte) byte {
	return byte(unicode.ToLower(rune(ch)))
}

// ChApplyN applies a function to n characters in a string object
func ChApplyN(str *StrObject, fn func(byte) byte, n int) {
	if n > 0 {
		str.ApplyN(fn, n, 1)
	}
}

// ChSearchStr searches for a target string within a text string.
// Returns true if found, with foundLoc set to the location.
// If backwards is true, searches from the end of the text.
// If exactcase is true, the search is case-sensitive.
func ChSearchStr(
	target *StrObject,
	st1 int,
	len1 int,
	text *StrObject,
	st2 int,
	len2 int,
	exactcase bool,
	backwards bool,
	foundLoc *int,
) bool {
	// Create a copy of the text segment to search
	var s StrObject
	s.Copy(text, st2, len2, 1)

	// Reverse if searching backwards
	if backwards {
		ChReverseStr(&s, &s, len2)
		*foundLoc = len2
	} else {
		*foundLoc = 0
	}

	// Convert to uppercase if case-insensitive search
	if !exactcase {
		s.ApplyN(ChToUpper, len2, 1)
	}

	// Search for the target
	for i := 1; i <= len2-len1+1; i++ {
		if s.Equals(target, len1, i, st1) {
			if backwards {
				*foundLoc = len2 - (i + len1) + 1
			} else {
				*foundLoc = i - 1
			}
			return true
		}
	}
	return false
}
