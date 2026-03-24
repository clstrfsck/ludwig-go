// Test helpers
package ludwig

import "testing"

// withFrame sets CurrentFrame and TtControlC for the duration of a test and
// restores them on cleanup.
func withFrame(t testing.TB, frame *FrameObject) {
	t.Helper()
	oldFrame := CurrentFrame
	oldCC := TtControlC
	CurrentFrame = frame
	TtControlC = false
	t.Cleanup(func() {
		CurrentFrame = oldFrame
		TtControlC = oldCC
	})
}
