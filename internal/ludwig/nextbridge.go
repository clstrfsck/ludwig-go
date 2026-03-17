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

// searchForward finds the first character in chars at or after (line, col), scanning forward.
// Returns the matching (line, col), or (nil, 0) if not found.
func searchForward(chars [256]bool, line *LineHdrObject, col int) (*LineHdrObject, int) {
	for line != nil {
		i := col
		for i <= line.Used {
			if chars[line.Str.Get(i)] {
				return line, i
			}
			i += 1
		}
		// Match a space at EOL
		if chars[' '] && i == line.Used+1 {
			return line, i
		}
		line = line.FLink
		col = 1
	}
	return nil, 0
}

// searchBackward finds the first character in chars at or before (line, col), scanning backward.
// Returns the matching (line, col) and true, or (nil, 0, false) if not found.
func searchBackward(chars [256]bool, line *LineHdrObject, col int, bridge bool) (*LineHdrObject, int) {
	for line != nil {
		if line.Used < col {
			if chars[' '] {
				return line, col
			}
			col = line.Used
		}
		for j := col; j >= 1; j -= 1 {
			if chars[line.Str.Get(j)] {
				return line, j
			}
		}
		if line.BLink != nil {
			line = line.BLink
			col = line.Used + 1
		} else if bridge {
			return line, col // This is safe since only -1BR is allowed
		} else {
			return nil, 0
		}
	}
	return nil, 0
}

// NextbridgeCommand implements the NEXT and BRIDGE commands
// NEXT searches for characters in the set, BRIDGE searches for characters NOT in the set
func NextbridgeCommand(count int, tpar *TParObject, bridge bool) bool {
	// Form the character set
	chars := [256]bool{}
	i := 1
	for i <= tpar.Len {
		ch1 := tpar.Str.Get(i)
		ch2 := ch1
		i += 1
		if i+2 <= tpar.Len {
			if (tpar.Str.Get(i) == '.') && (tpar.Str.Get(i+1) == '.') {
				ch2 = tpar.Str.Get(i + 2)
				i += 3
			}
		}
		// Add range ch1..ch2 to set
		for ch := ch1; ch <= ch2; ch += 1 {
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
			newCol += 1
		}
		for {
			newLine, newCol = searchForward(chars, newLine, newCol)
			if newLine == nil {
				return false
			}
			newCol += 1
			count -= 1
			if count == 0 {
				break
			}
		}
		newCol -= 1
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
			newCol -= 1
		}
		for {
			newLine, newCol = searchBackward(chars, newLine, newCol, bridge)
			if newLine == nil {
				return false
			}
			newCol -= 1
			count += 1
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
