/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         NEXTBRIDGE
//
// Description:  The NEXT and BRIDGE commands.

package ludwig

// NextbridgeCommand implements the NEXT and BRIDGE commands
// NEXT searches for characters in the set, BRIDGE searches for characters NOT in the set
func NextbridgeCommand(count int, tpar *TParObject, bridge bool) bool {
	// Form the character set
	chars := [256]bool{}
	i := 1
	for i <= tpar.Len {
		ch1 := tpar.Str.Get(i)
		ch2 := ch1
		i++
		if i+2 <= tpar.Len {
			if (tpar.Str.Get(i) == '.') && (tpar.Str.Get(i+1) == '.') {
				ch2 = tpar.Str.Get(i + 2)
				i += 3
			}
		}
		// Add range ch1..ch2 to set
		for ch := ch1; ch <= ch2; ch++ {
			chars[ch] = true
		}
	}

	if bridge {
		// Bridge inverts the character set
		oldChars := chars
		chars = [256]bool{}
		for i := range oldChars {
			chars[i] = !oldChars[i]
		}
	}

	// Search for a character in the set
	newLine := CurrentFrame.Dot.Line
	var newCol int

	if count > 0 {
		newCol = CurrentFrame.Dot.Col
		if !bridge {
			newCol++
		}
		for {
			for newLine != nil {
				i = newCol
				for i <= newLine.Used {
					if chars[newLine.Str.Get(i)] {
						newCol = i
						goto l1
					}
					i++
				}
				if chars[' '] && (i == newLine.Used+1) { // Match a space at EOL
					newCol = i
					goto l1
				}
				newLine = newLine.FLink
				newCol = 1
			}
			return false
		l1:
			newCol++
			count--
			if count == 0 {
				break
			}
		}
		newCol--
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkEquals],
		) {
			return false
		}
	} else if count < 0 {
		newCol = CurrentFrame.Dot.Col - 1
		if !bridge {
			newCol--
		}
		for {
			for newLine != nil {
				if newLine.Used < newCol {
					if chars[' '] {
						goto l2
					}
					newCol = newLine.Used
				}
				for j := newCol; j >= 1; j-- {
					if chars[newLine.Str.Get(j)] {
						newCol = j
						goto l2
					}
				}
				if newLine.BLink != nil {
					newLine = newLine.BLink
					newCol = newLine.Used + 1
				} else if bridge {
					goto l2 // This is safe since only -1BR is allowed
				} else {
					return false
				}
			}
		l2:
			newCol--
			count++
			if count == 0 {
				break
			}
		}
		newCol += 2
		if !MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkEquals],
		) {
			return false
		}
	} else {
		return MarkCreate(
			CurrentFrame.Dot.Line,
			CurrentFrame.Dot.Col,
			&CurrentFrame.Marks[MarkEquals],
		)
	}

	// Found it, move dot to new point
	return MarkCreate(newLine, newCol, &CurrentFrame.Dot)
}
