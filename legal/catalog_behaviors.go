package legal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
)

// PlayerSelector identifies which PlayerIndex source a player predicate reads.
// It is a symbolic selector, not a boardgame.PlayerIndex or player state. Its
// representation is deliberately private so callers choose one of the
// supported, boot-validatable selectors below.
type PlayerSelector struct {
	source playerSelectorSource
	field  string
}

type playerSelectorSource uint8

const (
	playerSelectorInvalid playerSelectorSource = iota
	playerSelectorCurrent
	playerSelectorProposer
	playerSelectorMoveField
)

// CurrentPlayer selects state.CurrentPlayerIndex().
func CurrentPlayer() PlayerSelector {
	return PlayerSelector{source: playerSelectorCurrent}
}

// Proposer selects the concrete player proposing the move. It remains distinct
// from CurrentPlayer during simultaneous-play phases.
func Proposer() PlayerSelector {
	return PlayerSelector{source: playerSelectorProposer}
}

// PlayerFromMove selects the player named by a boardgame.PlayerIndex field on
// the move. Boot validation rejects missing and incorrectly typed fields.
func PlayerFromMove(field string) PlayerSelector {
	return PlayerSelector{source: playerSelectorMoveField, field: field}
}

func (p PlayerSelector) path(prop string) string {
	switch p.source {
	case playerSelectorCurrent:
		return "player." + prop
	case playerSelectorProposer:
		return "proposer." + prop
	case playerSelectorMoveField:
		return "players[move." + p.field + "]." + prop
	default:
		return "." + prop
	}
}

// PlayerBoolAt requires a bool property on player to equal want. Prefer a
// semantic behavior helper such as PlayerHasNotSubmitted when one exists.
func PlayerBoolAt(player PlayerSelector, prop string, want bool) Spec {
	return Spec{Name: "playerBoolAt", Args: []string{player.path(prop), strconv.FormatBool(want)}}
}

func behaviorPlayerBool(player PlayerSelector, prop string, want bool, template string) Spec {
	spec := PlayerBoolAt(player, prop, want).WithTemplateKey(template)
	if player.source == playerSelectorProposer {
		return spec.WithAdminBypass()
	}
	return spec
}

const (
	TemplatePlayerAlreadySubmitted = "legal.player_already_submitted"
	TemplatePlayerNotSubmitted     = "legal.player_not_submitted"
	TemplatePlayerInactive         = "legal.player_inactive"
	TemplatePlayerActive           = "legal.player_active"
	TemplateSeatNotFilled          = "legal.seat_not_filled"
	TemplateSeatNotClosed          = "legal.seat_not_closed"
	TemplatePlayerNotAdmin         = "legal.player_not_admin"
)

// PlayerHasSubmitted requires behaviors.PlayerSubmission's canonical flag.
func PlayerHasSubmitted(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "PlayerSubmitted", true, TemplatePlayerNotSubmitted)
}

// PlayerHasNotSubmitted requires behaviors.PlayerSubmission's canonical flag
// to be false.
func PlayerHasNotSubmitted(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "PlayerSubmitted", false, TemplatePlayerAlreadySubmitted)
}

// PlayerIsActive requires behaviors.InactivePlayer's canonical flag to be false.
func PlayerIsActive(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "PlayerInactive", false, TemplatePlayerInactive)
}

// PlayerIsInactive requires behaviors.InactivePlayer's canonical flag.
func PlayerIsInactive(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "PlayerInactive", true, TemplatePlayerActive)
}

// PlayerSeatIsFilled requires behaviors.Seat's canonical filled flag.
func PlayerSeatIsFilled(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "SeatFilled", true, TemplateSeatNotFilled)
}

// PlayerSeatIsClosed requires behaviors.Seat's canonical closed flag.
func PlayerSeatIsClosed(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "SeatClosed", true, TemplateSeatNotClosed)
}

// PlayerIsAdmin requires behaviors.GameAdministrator's canonical admin flag.
func PlayerIsAdmin(player PlayerSelector) Spec {
	return behaviorPlayerBool(player, "IsAdmin", true, TemplatePlayerNotAdmin)
}

func playerBoolAtConstructor() *PredicateConstructor {
	return &PredicateConstructor{
		Name: "playerBoolAt",
		Constructor: func(spec Spec, chest *boardgame.ComponentChest, resolve func(Spec) (*Predicate, error)) (*Predicate, error) {
			if len(spec.Args) != 2 {
				return nil, fmt.Errorf("legal: playerBoolAt requires 2 args (player path, want), got %d", len(spec.Args))
			}
			path := spec.Args[0]
			if !strings.HasPrefix(path, "player.") && !strings.HasPrefix(path, "proposer.") && !strings.HasPrefix(path, "players[move.") {
				return nil, fmt.Errorf("legal: playerBoolAt path %q must select a player via CurrentPlayer, Proposer, or PlayerFromMove", path)
			}
			want, err := playerBoolWant(spec.Args[1])
			if err != nil {
				return nil, fmt.Errorf("legal: playerBoolAt: arg 2 (want) %s", err)
			}
			template := spec.Message
			if template == "" {
				template = TemplatePlayerBool
			}

			return &Predicate{
				Name:              "playerBoolAt",
				ClientEvaluable:   true,
				UsesProposer:      strings.HasPrefix(path, "proposer."),
				Args:              spec.Args,
				Reads:             []Read{{Path: PropPath(path), Facet: boardgame.LegalFacetValues}},
				RequiredReadTypes: map[PropPath]boardgame.PropertyType{PropPath(path): boardgame.TypeBool},
				Cost:              boardgame.LegalCostTrivial,
				EmittedTemplates:  []string{template},
				EmittedBindings:   map[string][]string{template: {"prop", "want"}},
				Evaluate: func(ctx Context) Verdict {
					value, err := resolveBoolPath(path, ctx)
					if err != nil {
						return UnknownVerdict(err.Error())
					}
					if value == want {
						return PassVerdict()
					}
					return FailT(template, map[string]BindingValue{
						"prop": String(path),
						"want": String(strconv.FormatBool(want)),
					})
				},
			}, nil
		},
	}
}
