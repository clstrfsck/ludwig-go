/**********************************************************************}
{                                                                      }
{            L      U   U   DDDD   W      W  IIIII   GGGG              }
{            L      U   U   D   D   W    W     I    G                  }
{            L      U   U   D   D   W ww W     I    G   GG             }
{            L      U   U   D   D    W  W      I    G    G             }
{            LLLLL   UUU    DDDD     W  W    IIIII   GGGG              }
{                                                                      }
{**********************************************************************/

// Type declarations used throughout the Ludwig code base.

package ludwig

// SpecialFrames contains global frames used in various parts of the code
type SpecialFrames struct {
	cmd  *FrameObject
	heap *FrameObject
	oops *FrameObject
}

func NewSpecialFrames(cmdFrame *FrameObject, heapFrame *FrameObject, oopsFrame *FrameObject) *SpecialFrames {
	return &SpecialFrames{
		cmd:  cmdFrame,
		heap: heapFrame,
		oops: oopsFrame,
	}
}

func (sf *SpecialFrames) Cmd() *FrameObject {
	return sf.cmd
}

func (sf *SpecialFrames) Heap() *FrameObject {
	return sf.heap
}

func (sf *SpecialFrames) Oops() *FrameObject {
	return sf.oops
}
