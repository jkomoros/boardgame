/*

Package listing is a simple package of constants for listing of games, primarily
in a separate package to avoid circular dependencies.

*/
package listing

//Type is an enum of the type of list that should be returned.
type Type int

const (
	//All denotes All games
	All Type = iota
	//ParticipatingActive denotes games this player is in that are not finished
	ParticipatingActive
	//ParticipatingFinished denotes games this player is in that are finished
	ParticipatingFinished
	//VisibleJoinableActive denotes games that this player is not in that are visible, possible to join
	//(open + slots) and Active
	VisibleJoinableActive
	//VisibleActive denotes games that this player is not in that are visible and active but not
	//joinable. (Popcorn games)
	VisibleActive
)
