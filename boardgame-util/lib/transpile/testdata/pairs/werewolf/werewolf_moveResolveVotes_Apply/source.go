func (m *moveResolveVotes) Apply(state boardgame.State) error {
	return finishVoteResolution(state)
}