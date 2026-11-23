func (m *moveResolveVotesOnTimeout) Apply(state boardgame.State) error {
	return finishVoteResolution(state)
}