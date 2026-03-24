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
func SwapLine(frame *FrameObject, rept LeadParam, count int) bool {
	// SW is implemented as a ST of the dot line to before the other line.

	thisLine := frame.Dot.Line
	dotCol := frame.Dot.Col
	nextLine := thisLine.FLink

	if nextLine == nil {
		return false
	}

	var topMark *MarkObject
	var endMark *MarkObject
	var destMark *MarkObject

	defer func() {
		MarkDestroy(&topMark)
		MarkDestroy(&endMark)
		MarkDestroy(&destMark)
	}()

	var destLine *LineHdrObject

	switch rept {
	case LeadParamNone, LeadParamPlus, LeadParamPInt:
		destLine = nextLine
		for i := 1; i <= count; i++ {
			destLine = destLine.FLink
			if destLine == nil {
				return false
			}
		}
	case LeadParamMinus, LeadParamNInt:
		destLine = thisLine
		for i := -1; i >= count; i-- {
			destLine = destLine.BLink
			if destLine == nil {
				return false
			}
		}
	case LeadParamPIndef:
		destLine = frame.LastGroup.LastLine
	case LeadParamNIndef:
		destLine = frame.FirstGroup.FirstLine
	case LeadParamMarker:
		destLine = frame.Marks[count].Line
	}

	MarkCreate(thisLine, 1, &topMark)
	MarkCreate(nextLine, 1, &endMark)
	MarkCreate(destLine, 1, &destMark)
	if !TextMove(false, 1, topMark, endMark, destMark, &frame.Dot, &topMark) {
		return false
	}
	frame.TextModified = true
	frame.Dot.Col = dotCol
	MarkCreate(frame.Dot.Line, frame.Dot.Col, &frame.Marks[MarkModified])
	return true
}
