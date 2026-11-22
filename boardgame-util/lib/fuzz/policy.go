package fuzz

import (
	"context"
	"fmt"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/bots"
)

// policy picks among a step's legal actions. Every policy falls back to
// random-legal when it has nothing to say.
type policy interface {
	// start is called once the game exists, before the first step.
	start(d *driver) error
	// choose returns an action to take, or ok false to defer to
	// random-legal.
	choose(d *driver, step *step) (chosen action, ok bool, err error)
}

func (d *driver) policySetUp() error {
	name := d.cfg.Policy
	switch {
	case name == PolicyRandomLegal:
		d.policy = nil
	case name == PolicyAgent:
		agents := d.manager.Agents()
		if len(agents) == 0 {
			return fmt.Errorf("game %s configures no agent", d.cfg.Game)
		}
		d.policy = &agentPolicy{agent: agents[0]}
	case strings.HasPrefix(name, PolicyBotPrefix):
		botName := strings.TrimPrefix(name, PolicyBotPrefix)
		entry, ok := botPolicies[botName]
		if !ok {
			return fmt.Errorf("unknown bot %q; known: %v", botName, Bots())
		}
		if entry.game != d.cfg.Game {
			return fmt.Errorf("bot %q plays %s, not %s", botName, entry.game, d.cfg.Game)
		}
		d.policy = &botPolicy{policy: entry.policy()}
	default:
		return fmt.Errorf("unknown policy %q (want %s, %s<name> or %s)", name, PolicyRandomLegal, PolicyBotPrefix, PolicyAgent)
	}
	return nil
}

func (d *driver) policyStart() error {
	if d.policy == nil {
		return nil
	}
	return d.policy.start(d)
}

// choose asks the policy, then picks uniformly among the legal actions.
func (d *driver) choose(step *step) (action, error) {
	if d.policy != nil {
		chosen, ok, err := d.policy.choose(d, step)
		if err != nil {
			return action{}, err
		}
		if ok {
			return chosen, nil
		}
	}
	return step.actions[d.random.Intn(len(step.actions))], nil
}

// currentActor is the delegate's current player when it is one seated
// player, else ok false.
func (d *driver) currentActor() (boardgame.PlayerIndex, bool) {
	state := d.game.CurrentState()
	current := d.manager.Delegate().CurrentPlayerIndex(state)
	for _, actor := range d.actors(state) {
		if actor == current {
			return current, true
		}
	}
	return 0, false
}

// botPolicy lets a registered bots.Policy decide for the current actor,
// with the driver's generator as its randomness.
type botPolicy struct {
	policy bots.Policy
}

func (b *botPolicy) start(d *driver) error { return nil }

func (b *botPolicy) choose(d *driver, step *step) (action, bool, error) {
	actor, ok := d.currentActor()
	if !ok {
		return action{}, false, nil
	}
	observation, err := bots.Observe(d.game, actor)
	if err != nil {
		return action{}, false, nil
	}
	spec, err := b.policy.Decide(context.Background(), observation, d.random)
	if err != nil {
		return action{}, false, nil
	}
	move, err := d.game.MoveWithInput(spec.Name, spec.Input)
	if err != nil {
		return action{}, false, nil
	}
	if legalErr, panicked := safeLegal(move, d.game.CurrentState(), actor); panicked || legalErr != nil {
		return action{}, false, nil
	}
	return action{kind: "propose", move: move.Info().Name(), input: spec.Input, actor: actor}, true, nil
}

// agentPolicy asks the game's boardgame.Agent for the current actor's move.
// Not reproducible: memory's agent draws from the global math/rand.
type agentPolicy struct {
	agent  boardgame.Agent
	states map[boardgame.PlayerIndex][]byte
}

func (a *agentPolicy) start(d *driver) error {
	a.states = map[boardgame.PlayerIndex][]byte{}
	for i := 0; i < d.game.NumPlayers(); i++ {
		a.states[boardgame.PlayerIndex(i)] = a.agent.SetUpForGame(d.game, boardgame.PlayerIndex(i))
	}
	return nil
}

func (a *agentPolicy) choose(d *driver, step *step) (action, bool, error) {
	actor, ok := d.currentActor()
	if !ok {
		return action{}, false, nil
	}
	move, newState := a.agent.ProposeMove(d.game, actor, a.states[actor])
	if newState != nil {
		a.states[actor] = newState
	}
	if move == nil {
		return action{}, false, nil
	}
	if legalErr, panicked := safeLegal(move, d.game.CurrentState(), actor); panicked || legalErr != nil {
		return action{}, false, nil
	}
	return action{kind: "propose", move: move.Info().Name(), actor: actor, bound: move}, true, nil
}
