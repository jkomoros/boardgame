package corpus

import (
	"encoding/json"
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/boardgame-util/lib/conformance"
	"github.com/jkomoros/boardgame/boardgame-util/lib/scenario"
)

// script grows a scenario.Spec one step at a time. A scenario is an explicit
// list of moves, but which move is interesting next (the card that busts, the
// matching pair) depends on hidden state, so a script runs what it has so far
// on the real engine, looks at the authoritative result, and appends the next
// step. The finished Spec is ordinary and replays without any of this.
type script struct {
	newDelegate conformance.NewDelegate
	spec        scenario.Spec
}

func newScript(game, name, seed string, players int, variant boardgame.Variant) *script {
	return &script{newDelegate: games[game], spec: scenario.Spec{Name: name, Seed: seed, Players: players, Variant: variant}}
}

// peek runs the script so far. It decodes the final unsanitized state into
// state and the chest into chest (either may be nil), and reports whether
// the game has finished.
func (s *script) peek(state, chest interface{}) (finished bool, err error) {
	result, err := scenario.Run(s.newDelegate(), s.spec)
	if err != nil {
		return false, err
	}
	if state != nil {
		blob, err := result.History.State(result.History.Game().Version)
		if err != nil {
			return false, err
		}
		if err := json.Unmarshal(blob, state); err != nil {
			return false, err
		}
	}
	if chest != nil {
		if err := json.Unmarshal(result.Replay.Chest, chest); err != nil {
			return false, err
		}
	}
	return result.History.Game().Finished, nil
}

func (s *script) move(label, name string, player boardgame.PlayerIndex, input map[string]interface{}) {
	s.spec.Steps = append(s.spec.Steps, scenario.Step{Label: label, Move: name, Player: player, Input: input})
}

func (s *script) timer(label string) {
	s.spec.Steps = append(s.spec.Steps, scenario.Step{Label: label, Timer: true})
}

func (s *script) seat(player boardgame.PlayerIndex) {
	s.spec.Steps = append(s.spec.Steps, scenario.Step{Label: fmt.Sprintf("Seat player %d", player), Seat: &player})
}

type currentPlayerState struct {
	Game struct{ CurrentPlayer boardgame.PlayerIndex }
}

// memoryFullGame plays a ten-card game to the end.
func memoryFullGame() (scenario.Spec, error) {
	s := newScript("memory", "Memory: a full small game", "memory-full-game-v1", 2, boardgame.Variant{"numcards": "small"})
	var setup struct {
		Game struct {
			CurrentPlayer boardgame.PlayerIndex
			HiddenCards   struct{ Indexes []int }
		}
	}
	var chest struct {
		Decks map[string][]struct{ Values struct{ Type string } }
	}
	if _, err := s.peek(&setup, &chest); err != nil {
		return s.spec, err
	}
	// Cards return to their own slot when hidden, so the setup layout holds
	// for the whole game.
	slotsByType := map[string][]int{}
	var types []string
	for slot, index := range setup.Game.HiddenCards.Indexes {
		if index < 0 || index >= len(chest.Decks["cards"]) {
			return s.spec, fmt.Errorf("slot %d holds no card at setup", slot)
		}
		cardType := chest.Decks["cards"][index].Values.Type
		if len(slotsByType[cardType]) == 0 {
			types = append(types, cardType)
		}
		slotsByType[cardType] = append(slotsByType[cardType], slot)
	}
	if len(types) < 2 {
		return s.spec, fmt.Errorf("setup has fewer than two card types")
	}
	reveal := func(label string, slot int) error {
		var now currentPlayerState
		if _, err := s.peek(&now, nil); err != nil {
			return err
		}
		s.move(label, "Reveal Card", now.Game.CurrentPlayer, map[string]interface{}{"CardIndex": slot})
		return nil
	}
	mismatch := [2]int{slotsByType[types[0]][0], slotsByType[types[1]][0]}

	// A non-match that the timer hides.
	for _, slot := range mismatch {
		if err := reveal("Reveal a card that will not match", slot); err != nil {
			return s.spec, err
		}
	}
	s.timer("The hide-cards timer fires")

	// A non-match that the player hides, cancelling the timer.
	var hider currentPlayerState
	if _, err := s.peek(&hider, nil); err != nil {
		return s.spec, err
	}
	for _, slot := range mismatch {
		if err := reveal("Reveal the same non-matching card", slot); err != nil {
			return s.spec, err
		}
	}
	s.move("Hide the cards before the timer fires", "Hide Cards", hider.Game.CurrentPlayer, nil)

	// Every pair, captured in turn.
	for _, cardType := range types {
		slots := slotsByType[cardType]
		if len(slots) != 2 {
			return s.spec, fmt.Errorf("card type %q has %d cards, want 2", cardType, len(slots))
		}
		for _, slot := range slots {
			if err := reveal("Reveal half of a pair", slot); err != nil {
				return s.spec, err
			}
		}
	}
	finished, err := s.peek(nil, nil)
	if err != nil {
		return s.spec, err
	}
	if !finished {
		return s.spec, fmt.Errorf("the game did not finish")
	}
	return s.spec, nil
}

// blackjackBustAndFinish plays two rounds with three players. In round one
// player 0 hits until bust, player 1 stands, player 2 hits once and stands;
// in round two everyone stands.
func blackjackBustAndFinish() (scenario.Spec, error) {
	s := newScript("blackjack", "Blackjack: a bust, two rounds, game over", "blackjack-bust-v1", 3, boardgame.Variant{"maxrounds": "2"})
	type turn struct {
		round  int
		player boardgame.PlayerIndex
	}
	hits := map[turn]int{}
	for guard := 0; guard < 64; guard++ {
		var now struct {
			Game struct {
				CurrentPlayer   boardgame.PlayerIndex
				RoundsCompleted int
			}
		}
		finished, err := s.peek(&now, nil)
		if err != nil {
			return s.spec, err
		}
		if finished {
			if hits[turn{0, 0}] == 0 {
				return s.spec, fmt.Errorf("player 0 never hit")
			}
			return s.spec, nil
		}
		player := now.Game.CurrentPlayer
		key := turn{now.Game.RoundsCompleted, player}
		hit := false
		if key.round == 0 {
			// Player 0's turn only ends by busting; player 2 hits once.
			hit = player == 0 || (player == 2 && hits[key] == 0)
		}
		if hit {
			hits[key]++
			s.move(fmt.Sprintf("Player %d hits", player), "Current Player Hit", player, nil)
		} else {
			s.move(fmt.Sprintf("Player %d stands", player), "Current Player Stand", player, nil)
		}
	}
	return s.spec, fmt.Errorf("the game did not finish")
}

// blackjackReshuffle plays seven players for four rounds, everyone standing
// at once. Each round deals fourteen cards, so the draw stack runs out during
// the fourth deal and the discard stack is shuffled back into it mid-deal.
func blackjackReshuffle() (scenario.Spec, error) {
	const players, rounds = 7, 4
	s := newScript("blackjack", "Blackjack: the draw stack runs out", "blackjack-reshuffle-v1", players, boardgame.Variant{"maxrounds": "4"})
	// The turn order is fixed, so the moves are listed without running the
	// game before each one: the player whose stand ends a round keeps the
	// turn and opens the next round.
	first := boardgame.PlayerIndex(0)
	for round := 0; round < rounds; round++ {
		player := first
		for turn := 0; turn < players; turn++ {
			s.move(fmt.Sprintf("Player %d stands", player), "Current Player Stand", player, nil)
			first = player
			player = (player + 1) % players
		}
	}
	var end struct {
		Game struct {
			DrawStack struct{ ShuffleCount int }
		}
	}
	finished, err := s.peek(&end, nil)
	if err != nil {
		return s.spec, err
	}
	if !finished {
		return s.spec, fmt.Errorf("the game did not finish")
	}
	// Set-up shuffles once; a second shuffle is the reshuffle.
	if end.Game.DrawStack.ShuffleCount < 2 {
		return s.spec, fmt.Errorf("the draw stack was never reshuffled")
	}
	return s.spec, nil
}

// debugAnimationsShuffles walks the animation test bed's stack operations.
func debugAnimationsShuffles() (scenario.Spec, error) {
	s := newScript("debuganimations", "Debug animations: shuffles and hidden stacks", "debuganimations-shuffles-v1", 2, nil)
	steps := []struct {
		name   string
		player boardgame.PlayerIndex
		input  map[string]interface{}
	}{
		{"Move Fan Card", 0, nil},
		{"Move Fan Card", 1, nil},
		{"Shuffle", 0, nil},
		{"Visible Shuffle", 1, nil},
		{"Shuffle Hidden", 1, nil},
		{"Flip Card Between Hidden and Revealed", 0, nil},
		{"Flip Card Between Hidden and Revealed", 1, nil},
		{"Move Between Hidden", 0, nil},
		{"Move Token", 0, nil},
		{"Move Token Sanitized", 1, nil},
		{"Move Card Between Short Stacks", 0, map[string]interface{}{"FromFirst": true}},
		{"Move Card Between Draw And Discard Stacks", 0, map[string]interface{}{"FromDraw": true}},
		{"Start Move All Components To Hidden", 0, nil},
		{"Start Move All Components To Visible", 1, nil},
		{"Shuffle", 1, nil},
	}
	for _, step := range steps {
		s.move("", step.name, step.player, step.input)
	}
	if _, err := s.peek(nil, nil); err != nil {
		return s.spec, err
	}
	return s.spec, nil
}

// pigShort plays four turns: roll up to twice, then bank if still rolling.
func pigShort() (scenario.Spec, error) {
	s := newScript("pig", "Pig: a few turns", "pig-short-v1", 2, nil)
	current := func() (boardgame.PlayerIndex, error) {
		var now currentPlayerState
		_, err := s.peek(&now, nil)
		return now.Game.CurrentPlayer, err
	}
	banked := false
	for turn := 0; turn < 4 || !banked; turn++ {
		if turn > 16 {
			return s.spec, fmt.Errorf("no turn survived two rolls")
		}
		player, err := current()
		if err != nil {
			return s.spec, err
		}
		stillRolling := true
		for roll := 0; roll < 2 && stillRolling; roll++ {
			s.move(fmt.Sprintf("Player %d rolls", player), "Roll Dice", player, nil)
			next, err := current()
			if err != nil {
				return s.spec, err
			}
			stillRolling = next == player
		}
		if stillRolling {
			s.move(fmt.Sprintf("Player %d banks", player), "Done Turn", player, nil)
			banked = true
		}
	}
	return s.spec, nil
}

// ticTacToeSeatedWin seats both players, then player 0 takes the top row.
func ticTacToeSeatedWin() (scenario.Spec, error) {
	s := newScript("tictactoe", "Tic-tac-toe: seated players, a win", "tictactoe-seated-win-v1", 2, nil)
	s.seat(0)
	s.seat(1)
	for _, slot := range []int{0, 3, 1, 4, 2} {
		var now currentPlayerState
		if _, err := s.peek(&now, nil); err != nil {
			return s.spec, err
		}
		s.move(fmt.Sprintf("Player %d takes slot %d", now.Game.CurrentPlayer, slot), "Place Token", now.Game.CurrentPlayer, map[string]interface{}{"Slot": slot})
	}
	finished, err := s.peek(nil, nil)
	if err != nil {
		return s.spec, err
	}
	if !finished {
		return s.spec, fmt.Errorf("the game did not finish")
	}
	return s.spec, nil
}
