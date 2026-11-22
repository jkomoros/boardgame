package subset

// EngineGaps is everything the TypeScript engine lacks today that the
// translator would otherwise target: the API table's facade calls the engine
// does not export (keyed by facade name; none today), the stock move types of
// package moves the engine cannot build from a definition (keyed
// "moves.<Type>"), and the behaviors it does not port (keyed
// "behaviors.<Type>"). The value says what is missing.
//
// It is the one table of gaps. The source of truth for the stock and behavior
// entries is server/static/src/engine/rules/stock-inventory.json:
// TestEngineGapsMatchStockInventory fails unless a stock move type is listed
// here exactly when the inventory does not have it both ported and named in
// the stock registry (registryName), and a behavior exactly when the
// inventory does not have it ported; the same test fails if a facade name
// listed here is exported by engine/go/*.ts or engine/lib/*.ts.
//
// A stock type ported as parts but absent from the registry (RoundRobin,
// RoundRobinNumRounds: roundRobinParts needs the game's ConditionMet,
// PlayerConditionMet and RoundRobinAction) is a gap: gameFromData builds a
// stock move only through stockMove, which returns undefined for it, so a
// game function that selects through one, or calls its Apply or Legal, has
// no move in the loaded game to run in, however faithfully it is translated.
//
// Not detected here (known holes): a configuration stockMove cannot build
// although the type is registered (WaitForEnoughPlayers with
// moves.WithRequireExplicitStart, which DefaultRoundSetup's
// moves.WithManualStart sets; ShuffleDiscardIntoDraw with
// moves.WithDrawDiscardPairField; DrawToPlayer whose stacks Go infers at boot
// from a DrawDiscardPair or a single stack), a behavior ported only in part
// (PlayerColor's OwnsToken, Token, TokenFromDeck and TokenSpaceIndex; a call
// to one is refused anyway, as unknown-call, since the API table does not
// list them), and the phase hooks interfaces.BeforeLeavePhaser and
// BeforeEnterPhaser, which StartPhase does not call. None of the seven
// example games uses one.
var EngineGaps = map[string]string{
	// Stock move types the inventory does not port, or ports but the stock
	// registry cannot build.
	"moves.AdminPlayer":               "stock move not ported (needs behaviors.GameAdministrator)",
	"moves.AdvanceToken":              "stock move not ported (needs behaviors.LocationBehavior)",
	"moves.AllPlayersSubmitted":       "stock move not ported (needs behaviors.PlayerSubmission)",
	"moves.ApplyCountTimes":           "stock move base not ported",
	"moves.ApplyUntil":                "stock move base not ported",
	"moves.ApplyUntilCount":           "stock move base not ported",
	"moves.HopAlongPath":              "stock move not ported (needs behaviors.LocationBehavior and graph shortest paths)",
	"moves.Increment":                 "stock move not ported",
	"moves.MoveOnGraph":               "stock move not ported (needs behaviors.LocationBehavior and graph shortest paths)",
	"moves.RecordCurrentPlayerChoice": "stock move not ported (move choice recording)",
	"moves.ReplenishMarket":           "stock move not ported (needs behaviors.FaceUpMarket)",
	"moves.ResetAllPlayerSubmissions": "stock move not ported (needs behaviors.PlayerSubmission)",
	"moves.RoundRobin":                "stock move ported as parts (roundRobinParts) but not in the stock registry",
	"moves.RoundRobinNumRounds":       "stock move ported as parts (roundRobinParts) but not in the stock registry",
	"moves.SelectColor":               "stock move not ported",
	"moves.SelectRole":                "stock move not ported",
	"moves.SelectTeam":                "stock move not ported (needs behaviors.PlayerTeam)",

	// Behaviors the inventory does not port.
	"behaviors.FaceUpMarket":        "behavior not ported",
	"behaviors.GameAdministrator":   "behavior not ported",
	"behaviors.LocationBehavior":    "behavior not ported",
	"behaviors.MoveBudget":          "behavior not ported",
	"behaviors.PlayerOrderBehavior": "behavior not ported",
	"behaviors.PlayerSubmission":    "behavior not ported",
	"behaviors.PlayerTeam":          "behavior not ported",
}

// IsGap reports whether key (a facade name, "moves.<Type>" or
// "behaviors.<Type>") is something the engine lacks.
func IsGap(key string) bool {
	_, gap := EngineGaps[key]
	return gap
}
