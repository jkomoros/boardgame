package moves

import (
	"fmt"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

/*
This file is the moves half of the widened composition seam (design spec
docs/superpowers/specs/2026-11-22-legality-seam-widening-design.md): every
framework move type's legality IS its contribution.

A framework type whose Legal() adds checks after its super-call keeps those
checks verbatim in a private per-type method (legalFinishTurn,
legalApplyUntil, ...), and its Legal() takes one shape:

	func (f *FinishTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
		if err := f.Default.Legal(state, proposer); err != nil {
			return err
		}
		if legalPlanOwnsChain(state, f.Name(), PreconditionFinishTurn) {
			return nil
		}
		return f.legalFinishTurn(state, proposer)
	}

Its ContributedPreconditions is the super-call's contribution followed by
FrameworkPrecondition(PreconditionFinishTurn), an opaque residue atom whose
registry constructor (below) calls that same private method on the concrete
move. For a move that never opted in nothing changes: the guard is false and
the frozen code runs. For an opted-in move the plan (evaluated once, inside
Default.Legal) already ran every level's check in order, so each level
returns after its super-call.

Residue atoms are LegalCustom contributed by a base type: server-only, never
memoized, and rendered with the frozen error text. They can be moved
(suppress, then re-author with FrameworkPrecondition) but never removed; core
enforces that generically (legal_residue.go). moves/seam_source_test.go
enforces the Legal() shape and that every guard's atom is registered here.
*/

// The names of the framework base types' own contributed checks, one per
// type whose Legal() adds checks beyond the move it embeds. Each is the
// lowerCamel name of the type that owns it. Pass one to
// WithoutLegalPrecondition together with FrameworkPrecondition (same name) in
// WithLegalPreconditions to evaluate that check at a different position.
const (
	PreconditionFinishTurn                PreconditionName = "finishTurn"
	PreconditionForceFinishTurn           PreconditionName = "forceFinishTurn"
	PreconditionApplyUntil                PreconditionName = "applyUntil"
	PreconditionRoundRobin                PreconditionName = "roundRobin"
	PreconditionDealCountComponents       PreconditionName = "dealCountComponents"
	PreconditionMoveCountComponents       PreconditionName = "moveCountComponents"
	PreconditionAnyPlayer                 PreconditionName = "anyPlayer"
	PreconditionAdminPlayer               PreconditionName = "adminPlayer"
	PreconditionSelectColor               PreconditionName = "selectColor"
	PreconditionSelectRole                PreconditionName = "selectRole"
	PreconditionSelectTeam                PreconditionName = "selectTeam"
	PreconditionAllPlayersSubmitted       PreconditionName = "allPlayersSubmitted"
	PreconditionResetAllPlayerSubmissions PreconditionName = "resetAllPlayerSubmissions"
	PreconditionSeatPlayer                PreconditionName = "seatPlayer"
	PreconditionCloseEmptySeat            PreconditionName = "closeEmptySeat"
	PreconditionInactivateEmptySeat       PreconditionName = "inactivateEmptySeat"
	PreconditionActivateEmptySeat         PreconditionName = "activateEmptySeat"
	PreconditionActivateFilledSeat        PreconditionName = "activateFilledSeat"
	PreconditionActivateInactivePlayer    PreconditionName = "activateInactivePlayer"
	PreconditionCloseAllSeats             PreconditionName = "closeAllSeats"
	PreconditionWaitForEnoughPlayers      PreconditionName = "waitForEnoughPlayers"
	PreconditionAdvanceToken              PreconditionName = "advanceToken"
	PreconditionHopAlongPath              PreconditionName = "hopAlongPath"
	PreconditionMoveOnGraph               PreconditionName = "moveOnGraph"
	PreconditionDefaultComponent          PreconditionName = "defaultComponent"
	PreconditionShuffleDiscardIntoDraw    PreconditionName = "shuffleDiscardIntoDraw"
	PreconditionReplenishMarket           PreconditionName = "replenishMarket"
)

// FrameworkPrecondition is the spec for one framework base type's own check
// (a Precondition* constant from this file). Base types contribute it
// automatically; a game authors it only to MOVE that check: suppress the name
// with WithoutLegalPrecondition and list FrameworkPrecondition(name) in
// WithLegalPreconditions where it should run instead. Authoring it anywhere
// else is a boot error.
func FrameworkPrecondition(name PreconditionName) legal.Spec {
	return legal.Spec{Name: string(name)}
}

// legalPlanOwnsChain is the guard every framework Legal() level runs right
// after its super-call. It reports true when that level must not run its own
// check imperatively: during the boot probe (which also records that this
// level, owning atom, is on the move's call path) or when the move has an
// assembled plan, which already evaluated atom in order.
func legalPlanOwnsChain(state boardgame.ImmutableState, moveName string, atom PreconditionName) bool {
	if state == nil {
		return false
	}
	manager := state.Manager()
	if manager == nil {
		return false
	}
	if manager.LegalProbeReachResidue(string(atom)) {
		return true
	}
	return manager.LegalPlanAssembled(moveName)
}

// frameworkCheck builds the registry constructor for one residue atom. Owner
// is the private interface the owning type's check method satisfies; it
// promotes through embedding, so the concrete move (however deeply it embeds
// the owning type) satisfies it.
func frameworkCheck[Owner any](name PreconditionName, check func(Owner, boardgame.ImmutableState, boardgame.PlayerIndex) error) *legal.PredicateConstructor {
	return boardgame.NewLegalResidueConstructor(string(name), func(ctx legal.Context) error {
		owner, ok := any(ctx.Move).(Owner)
		if !ok {
			return legal.UnknownVerdict(fmt.Sprintf("moves: %v: the move does not embed the framework type that owns this check", name)).Error()
		}
		return check(owner, ctx.State, ctx.ProposerPlayerIndex)
	})
}

type (
	finishTurnChecker interface {
		legalFinishTurn(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	forceFinishTurnChecker interface {
		legalForceFinishTurn(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	applyUntilChecker interface {
		legalApplyUntil(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	roundRobinChecker interface {
		legalRoundRobin(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	dealCountComponentsChecker interface {
		legalDealCountComponents(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	moveCountComponentsChecker interface {
		legalMoveCountComponents(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	anyPlayerChecker interface {
		legalAnyPlayer(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	adminPlayerChecker interface {
		legalAdminPlayer(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	selectColorChecker interface {
		legalSelectColor(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	selectRoleChecker interface {
		legalSelectRole(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	selectTeamChecker interface {
		legalSelectTeam(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	allPlayersSubmittedChecker interface {
		legalAllPlayersSubmitted(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	resetAllPlayerSubmissionsChecker interface {
		legalResetAllPlayerSubmissions(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	seatPlayerChecker interface {
		legalSeatPlayer(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	closeEmptySeatChecker interface {
		legalCloseEmptySeat(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	inactivateEmptySeatChecker interface {
		legalInactivateEmptySeat(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	activateEmptySeatChecker interface {
		legalActivateEmptySeat(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	activateFilledSeatChecker interface {
		legalActivateFilledSeat(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	activateInactivePlayerChecker interface {
		legalActivateInactivePlayer(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	closeAllSeatsChecker interface {
		legalCloseAllSeats(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	waitForEnoughPlayersChecker interface {
		legalWaitForEnoughPlayers(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	advanceTokenChecker interface {
		legalAdvanceToken(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	hopAlongPathChecker interface {
		legalHopAlongPath(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	moveOnGraphChecker interface {
		legalMoveOnGraph(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	defaultComponentChecker interface {
		legalDefaultComponent(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	shuffleDiscardIntoDrawChecker interface {
		legalShuffleDiscardIntoDraw(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
	replenishMarketChecker interface {
		legalReplenishMarket(boardgame.ImmutableState, boardgame.PlayerIndex) error
	}
)

// frameworkCheckConstructors is every framework base type's residue atom.
func frameworkCheckConstructors() []*legal.PredicateConstructor {
	return []*legal.PredicateConstructor{
		frameworkCheck(PreconditionFinishTurn, finishTurnChecker.legalFinishTurn),
		frameworkCheck(PreconditionForceFinishTurn, forceFinishTurnChecker.legalForceFinishTurn),
		frameworkCheck(PreconditionApplyUntil, applyUntilChecker.legalApplyUntil),
		frameworkCheck(PreconditionRoundRobin, roundRobinChecker.legalRoundRobin),
		frameworkCheck(PreconditionDealCountComponents, dealCountComponentsChecker.legalDealCountComponents),
		frameworkCheck(PreconditionMoveCountComponents, moveCountComponentsChecker.legalMoveCountComponents),
		frameworkCheck(PreconditionAnyPlayer, anyPlayerChecker.legalAnyPlayer),
		frameworkCheck(PreconditionAdminPlayer, adminPlayerChecker.legalAdminPlayer),
		frameworkCheck(PreconditionSelectColor, selectColorChecker.legalSelectColor),
		frameworkCheck(PreconditionSelectRole, selectRoleChecker.legalSelectRole),
		frameworkCheck(PreconditionSelectTeam, selectTeamChecker.legalSelectTeam),
		frameworkCheck(PreconditionAllPlayersSubmitted, allPlayersSubmittedChecker.legalAllPlayersSubmitted),
		frameworkCheck(PreconditionResetAllPlayerSubmissions, resetAllPlayerSubmissionsChecker.legalResetAllPlayerSubmissions),
		frameworkCheck(PreconditionSeatPlayer, seatPlayerChecker.legalSeatPlayer),
		frameworkCheck(PreconditionCloseEmptySeat, closeEmptySeatChecker.legalCloseEmptySeat),
		frameworkCheck(PreconditionInactivateEmptySeat, inactivateEmptySeatChecker.legalInactivateEmptySeat),
		frameworkCheck(PreconditionActivateEmptySeat, activateEmptySeatChecker.legalActivateEmptySeat),
		frameworkCheck(PreconditionActivateFilledSeat, activateFilledSeatChecker.legalActivateFilledSeat),
		frameworkCheck(PreconditionActivateInactivePlayer, activateInactivePlayerChecker.legalActivateInactivePlayer),
		frameworkCheck(PreconditionCloseAllSeats, closeAllSeatsChecker.legalCloseAllSeats),
		frameworkCheck(PreconditionWaitForEnoughPlayers, waitForEnoughPlayersChecker.legalWaitForEnoughPlayers),
		frameworkCheck(PreconditionAdvanceToken, advanceTokenChecker.legalAdvanceToken),
		frameworkCheck(PreconditionHopAlongPath, hopAlongPathChecker.legalHopAlongPath),
		frameworkCheck(PreconditionMoveOnGraph, moveOnGraphChecker.legalMoveOnGraph),
		frameworkCheck(PreconditionDefaultComponent, defaultComponentChecker.legalDefaultComponent),
		frameworkCheck(PreconditionShuffleDiscardIntoDraw, shuffleDiscardIntoDrawChecker.legalShuffleDiscardIntoDraw),
		frameworkCheck(PreconditionReplenishMarket, replenishMarketChecker.legalReplenishMarket),
	}
}
