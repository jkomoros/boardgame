package moves

import (
	"testing"
)

// TestCountAnyExactlyOne pins CountAny's semantics: it accepts precisely one
// match — not zero, not more. An old doc comment claimed equivalence to
// CountBetween(0,1) (i.e. zero-or-one); the implementation has always
// required exactly one, and progression matching relies on that. If this
// test starts failing, someone changed live progression semantics — that is
// a behavior change for every game using CountAny and must be a deliberate,
// documented decision, not a drive-by.
func TestCountAnyExactlyOne(t *testing.T) {
	counter := CountAny()

	// length should be irrelevant to CountAny; assert across several.
	for _, length := range []int{0, 1, 2, 10} {
		if err := counter(0, length); err == nil {
			t.Errorf("CountAny()(0, %d): got nil, want error (zero matches must not satisfy CountAny)", length)
		}
		if err := counter(1, length); err != nil {
			t.Errorf("CountAny()(1, %d): got %v, want nil (exactly one match satisfies CountAny)", length, err)
		}
		if err := counter(2, length); err == nil {
			t.Errorf("CountAny()(2, %d): got nil, want error (more than one match must not satisfy CountAny)", length)
		}
	}
}

// TestCountAnyMatchesCountExactlyOne pins the doc comment's claimed
// equivalence: CountAny() and CountExactly(1) agree everywhere we probe.
func TestCountAnyMatchesCountExactlyOne(t *testing.T) {
	anyC := CountAny()
	exactly := CountExactly(1)

	for count := 0; count <= 3; count++ {
		for _, length := range []int{0, 1, 5} {
			anyErr := anyC(count, length)
			exactlyErr := exactly(count, length)
			if (anyErr == nil) != (exactlyErr == nil) {
				t.Errorf("CountAny()(%d,%d) nil-ness %v disagrees with CountExactly(1) nil-ness %v",
					count, length, anyErr == nil, exactlyErr == nil)
			}
		}
	}
}
