/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Name:         OPSYS
//
// Description:  This routine executes a command in a subprocess and
//               transfers the result into the current frame.

package ludwig

// OpsysCommand executes a command in a subprocess and returns the output as lines
func OpsysCommand(command *TParObject, first **LineHdrObject, last **LineHdrObject, actualCnt *int) bool {
	*first = nil
	*last = nil
	*actualCnt = 0

	var mbx FileObject
	mbx.Valid = false
	mbx.FirstLine = nil
	mbx.LastLine = nil
	mbx.LineCount = 0
	mbx.OutputFlag = false
	if command.Len <= FileNameLen {
		mbx.Filename = command.Str.Slice(1, command.Len)
	} else {
		mbx.Filename = ""
	}

	if !FilesysCreateOpen(&mbx, nil, false) {
		return false
	}

	opsysResult := false
	for !mbx.Eof {
		result := NewFilled(' ', MaxStrLen)
		var outlen int
		if FilesysRead(&mbx, result, &outlen) {
			var line *LineHdrObject
			var line2 *LineHdrObject
			if !LinesCreate(1, &line, &line2) {
				goto l98
			}
			if !LineChangeLength(line, outlen) {
				LinesDestroy(&line, &line2)
				goto l98
			}
			ChFillCopy(result, 1, outlen, line.Str, 1, line.Len, ' ')
			line.Used = outlen
			line.BLink = *last
			if *last != nil {
				(*last).FLink = line
			} else {
				*first = line
			}
			*last = line
			*actualCnt += 1
		} else if !mbx.Eof { // Something terrible has happened!
			goto l98
		}
	}
	opsysResult = true
l98:
	FilesysClose(&mbx, 0, false)
	return opsysResult
}
