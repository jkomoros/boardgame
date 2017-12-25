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

//MoveTypeFallbackIsFixUp returns true. This is the primary effect of using
//the FixUp move.
func (f *FixUp) MoveTypeFallbackIsFixUp() bool {
	return true
}
