package moves

//FixUp is a simple move type that just wraps moves.Base. Its primary effect
//is to have the default IsFixUp for auto.Config to default to true. When you
//have a custom fix up move, it's best to embed this, because otherwise it's
//easy to forget to pass moves.WithIsFixUp to auto.Config.
//
//+autoreader
type FixUp struct {
	Base
}

//MoveTypeFallbackName returns FixUp Move"
func (f *FixUp) MoveTypeFallbackName() string {
	return "FixUp Move"
}

//MoveTypeFallbackHelpText returns "A move that is applied automatically to
//fix up the state after a player makes a move."
func (f *FixUp) MoveTypeFallbackHelpText() string {
	return "A move that is applied automatically to fix up the state after a player makes a move."
}

//MoveTypeFallbackIsFixUp returns true. This is the primary effect of using
//the FixUp move.
func (f *FixUp) MoveTypeFallbackIsFixUp() bool {
	return true
}
