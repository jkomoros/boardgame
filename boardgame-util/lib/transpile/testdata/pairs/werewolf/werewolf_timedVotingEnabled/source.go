func timedVotingEnabled(state boardgame.ImmutableState) bool {
	return state.Game().Variant()[variantKeyTimedVoting] == timedVotingOn
}