package aliaslegal

import rules "github.com/jkomoros/boardgame/legal"

var _ = rules.StackEmpty("game.DrawStack")
var _ = rules.PlayerBoolAt(rules.PlayerFromMove("Target"), "Ready", true)
var _ = rules.PlayerSeatIsFilled(rules.Proposer())
var _ = rules.InPhase()
var _ = rules.ProposerIsCurrentPlayer()
var _ = rules.AllActivePlayers(rules.Any(rules.PlayerBool("Ready"), rules.PropAtLeast("player.Score", 1)))

var computed = "game.Computed"
var _ = rules.StackEmpty(computed)

func StackEmpty(string) {}

func ignored() { StackEmpty("game.NotLegal") }
