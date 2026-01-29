/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         SWAP
// Description:  Swap Line command.

package ludwig

// SwapLine swaps the current line with another line
func SwapLine(rept LeadParam, count int) bool {
	// SW is implemented as a ST of the dot line to before the other line.
	result := false
	var topMark *MarkObject
	var endMark *MarkObject
	var destMark *MarkObject

	thisLine := CurrentFrame.Dot.Line
	dotCol := CurrentFrame.Dot.Col
	nextLine := thisLine.FLink
	var destLine *LineHdrObject

	if nextLine == nil {
		goto l99
	}

	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		destLine = nextLine
		for i := 1; i <= count; i++ {
			destLine = destLine.FLink
			if destLine == nil {
				goto l99
			}
		}
	case LeadParamMinus, LeadParamNInt:
		destLine = thisLine
		for i := -1; i >= count; i-- {
			destLine = destLine.BLink
			if destLine == nil {
				goto l99
			}
		}
	case LeadParamPIndef:
		destLine = CurrentFrame.LastGroup.LastLine
	case LeadParamNIndef:
		destLine = CurrentFrame.FirstGroup.FirstLine
	case LeadParamMarker:
		destLine = CurrentFrame.Marks[count-MinMarkNumber].Line
	}

	if !MarkCreate(thisLine, 1, &topMark) {
		goto l99
	}
	if !MarkCreate(nextLine, 1, &endMark) {
		goto l99
	}
	if !MarkCreate(destLine, 1, &destMark) {
		goto l99
	}
	if !TextMove(false, 1, topMark, endMark, destMark, &CurrentFrame.Dot, &topMark) {
		goto l99
	}
	CurrentFrame.TextModified = true
	CurrentFrame.Dot.Col = dotCol
	if !MarkCreate(
		CurrentFrame.Dot.Line, CurrentFrame.Dot.Col, &CurrentFrame.Marks[MarkModified-MinMarkNumber],
	) {
		goto l99
	}
	result = true
l99:
	if topMark != nil {
		MarkDestroy(&topMark)
	}
	if endMark != nil {
		MarkDestroy(&endMark)
	}
	if destMark != nil {
		MarkDestroy(&destMark)
	}
	return result
}
