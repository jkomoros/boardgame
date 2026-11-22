package subset

// EngineGaps is everything the TypeScript engine lacks today that the
// translator would otherwise target: the API table's "none (gap)" rows (keyed
// by facade name), the stock move types of package moves that
// server/static/src/engine/rules/stock.ts does not port (keyed
// "moves.<Type>"), and the behaviors it does not port (keyed
// "behaviors.<Type>"). The value says what is missing.
//
// It is the one table of gaps. Remove an entry when the engine gains it;
// TestEngineGapsMatchStockTS keeps the stock and behavior entries equal to
// what rules/*.ts does not mention.
var EngineGaps = map[string]string{
	// The API table's "none (gap)" rows.
	"rangeValue":          "range enums (enum.ImmutableRangeVal.RangeValue)",
	"rangeToValue":        "range enums (enum.RangeEnum.RangeToValue)",
	"valueToRange":        "range enums (enum.RangeEnum.ValueToRange)",
	"neighborsByKey":      "grid graphs (graph.Graph.NeighborsByKey)",
	"componentAtKey":      "enum-keyed stack access (Stack.ComponentAtKey)",
	"swapComponentsByKey": "enum-keyed stack access (Stack.SwapComponentsByKey)",

	// Stock move types rules/stock.ts does not port.
	"moves.ActivateEmptySeat":                      "stock move not ported",
	"moves.ActivateFilledSeat":                     "stock move not ported",
	"moves.AdminPlayer":                            "stock move base not ported",
	"moves.AdvanceToken":                           "stock move not ported",
	"moves.AllPlayersSubmitted":                    "stock move not ported",
	"moves.ApplyCountTimes":                        "stock move not ported",
	"moves.ApplyUntil":                             "stock move not ported",
	"moves.ApplyUntilCount":                        "stock move not ported",
	"moves.CloseAllSeats":                          "stock move not ported",
	"moves.CloseEmptySeat":                         "stock move not ported",
	"moves.CollectAllComponents":                   "stock move not ported",
	"moves.CollectComponentsUntilGameCountReached": "stock move not ported",
	"moves.CollectComponentsUntilPlayerCountLeft":  "stock move not ported",
	"moves.CollectCountComponents":                 "stock move not ported",
	"moves.DealAllComponents":                      "stock move not ported",
	"moves.DealComponentsUntilGameCountLeft":       "stock move not ported",
	"moves.DealComponentsUntilPlayerCountReached":  "stock move not ported",
	"moves.DefaultComponent":                       "stock move not ported",
	"moves.Done":                                   "stock move not ported",
	"moves.HopAlongPath":                           "stock move not ported",
	"moves.Increment":                              "stock move not ported",
	"moves.MoveAllComponents":                      "stock move not ported",
	"moves.MoveComponentsUntilCountLeft":           "stock move not ported",
	"moves.MoveComponentsUntilCountReached":        "stock move not ported",
	"moves.MoveCountComponents":                    "stock move not ported",
	"moves.MoveOnGraph":                            "stock move not ported",
	"moves.RecordCurrentPlayerChoice":              "stock move not ported",
	"moves.ReplenishMarket":                        "stock move not ported",
	"moves.ResetAllPlayerSubmissions":              "stock move not ported",
	"moves.RoundRobin":                             "stock move not ported",
	"moves.RoundRobinNumRounds":                    "stock move not ported",
	"moves.SelectColor":                            "stock move not ported",
	"moves.SelectRole":                             "stock move not ported",
	"moves.SelectTeam":                             "stock move not ported",
	"moves.ShuffleStack":                           "stock move not ported",

	// Behaviors rules/stock.ts does not port.
	"behaviors.ComponentColor":      "behavior not ported",
	"behaviors.FaceUpMarket":        "behavior not ported",
	"behaviors.GameAdministrator":   "behavior not ported",
	"behaviors.LocationBehavior":    "behavior not ported",
	"behaviors.MoveBudget":          "behavior not ported",
	"behaviors.PlayerColor":         "behavior not ported",
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
