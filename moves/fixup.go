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

//IsFixUp will return the value passed with WithFixUp, falling back on
//returning true. The returning of true is the primary result of embedding
//this move type.
func (f *FixUp) IsFixUp() bool {
	config := f.Info().Type().CustomConfiguration()
	return overrideIsFixUp(config, true)
}

//FixUpMulti is a simple move type that just wraps move.FixUp. Its primary
//effect is to have AllowMultipleInProgression() return true, which means that
//the logic for ordered move progressions within a phase will allow multiple
//in a row, until its Legal returns an error.
//
//+autoreader
type FixUpMulti struct {
	FixUp
}

//AllowMultipleInProgression returns true because the move is applied until
//ConditionMet returns nil.
func (f *FixUpMulti) AllowMultipleInProgression() bool {
	return true
}
