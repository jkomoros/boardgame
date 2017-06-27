/*

listing is a simple package of constants for listing of games, primarily in a
separate package to avoid circular dependencies.

*/
package listing

type Type int

const (
	//All games
	All Type = iota
	//Games this player is in that are not finished
	ParticipatingActive
	//Games this player is in that are finished
	ParticipatingFinished
	//Games that this player is not in that are visible, possible to join
	//(open + slots) and Active
	VisibleJoinableActive
	//Games that this player is not in that are visible and active but not
	//joinable. (Popcorn games)
	VisibleActive
)
