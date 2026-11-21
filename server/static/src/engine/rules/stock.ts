// Stock pieces a game definition is assembled from: the property sets of
// Go's behaviors, the move installers, and function helpers for the reusable
// moves in Go's `moves` package. Go builds these by struct embedding; here
// each is a function that returns a MoveSpec (or the parts of one to spread
// into your own), so there is no inheritance tree.

import type { PropDef } from '../definition.ts';
import { mayMoveFirstTo, mayMoveFirstToSlot, mayMoveToSlot } from '../legal/catalogStack.ts';
import { playerIndexEquivalent } from '../legal/playerIndex.ts';
import { ADMIN_PLAYER_INDEX, OBSERVER_PLAYER_INDEX, type LegalSpec } from '../legal/types.ts';
import type { Stack } from '../stack.ts';
import { type ProgressionGroup, groupMoveNames, optionalGroup, serialGroup } from './progression.ts';
import type { RulesState } from './state.ts';
import type { LegalResult, MoveInstance, MoveSpec, TurnHooks } from './types.ts';

// ----- behaviors: the properties Go's embedded structs add ----------------

/** behaviors.CurrentPlayerBehavior. */
export const currentPlayerProps: Readonly<Record<string, PropDef>> = { CurrentPlayer: { Type: 'playerIndex' } };
/** behaviors.PhaseBehavior: the game needs an enum named "phase". */
export const phaseProps: Readonly<Record<string, PropDef>> = { Phase: { Type: 'enum', Enum: 'phase' } };
/** behaviors.RoundRobin. */
export const roundRobinProps: Readonly<Record<string, PropDef>> = {
	RRLastPlayer: { Type: 'playerIndex' },
	RRStarterPlayer: { Type: 'playerIndex' },
	RRRoundCount: { Type: 'int' },
	RRHasStarted: { Type: 'bool' },
};
/** behaviors.Seat. Set the player trait `seat`. */
export const seatProps: Readonly<Record<string, PropDef>> = { SeatFilled: { Type: 'bool' }, SeatClosed: { Type: 'bool' } };
/** behaviors.InactivePlayer. Set the player trait `inactive`. */
export const inactivePlayerProps: Readonly<Record<string, PropDef>> = { PlayerInactive: { Type: 'bool' } };
/** behaviors.ScoreBehavior. Set the player trait `gameScore` to scoreGameScore. */
export const scoreProps: Readonly<Record<string, PropDef>> = { Score: { Type: 'int' } };
/** behaviors.PlayerElimination. Set the player trait `elimination`. */
export const eliminationProps: Readonly<Record<string, PropDef>> = { Eliminated: { Type: 'bool' } };

/** behaviors.ScoreBehavior.GameScore. */
export const scoreGameScore = (state: RulesState, player: number): number => state.players[player].Score as number;

// ----- installers ---------------------------------------------------------

type Groupable = MoveSpec | InstalledGroup;

/** A progression group together with the move specs it spans. */
export interface InstalledGroup {
	group: ProgressionGroup;
	moves: MoveSpec[];
}

function isGroup(item: Groupable): item is InstalledGroup {
	return (item as InstalledGroup).group !== undefined;
}

function asGroup(item: Groupable): InstalledGroup {
	if (isGroup(item)) return item;
	return { group: { kind: 'move', name: item.name, allowMultiple: item.allowMultipleInProgression === true }, moves: [item] };
}

/** moves.Serial. */
export function serial(...items: Groupable[]): InstalledGroup {
	const groups = items.map(asGroup);
	return { group: serialGroup(...groups.map(g => g.group)), moves: groups.flatMap(g => g.moves) };
}

/** moves.Optional. */
export function optional(item: Groupable): InstalledGroup {
	const inner = asGroup(item);
	return { group: optionalGroup(inner.group), moves: inner.moves };
}

/** moves.Add: moves legal in every phase, in no particular order. */
export function add(...moves: MoveSpec[]): MoveSpec[] {
	return moves;
}

/** moves.Combine. */
export function combine(...lists: MoveSpec[][]): MoveSpec[] {
	return lists.flat();
}

/** moves.AddForPhase: the moves are legal only in `phase` (a phase enum key). */
export function addForPhase(phase: number, ...moves: MoveSpec[]): MoveSpec[] {
	return moves.map(move => ({ ...move, legalPhases: [...(move.legalPhases ?? []), phase] }));
}

/**
 * moves.AddOrderedForPhase: the moves are legal only in `phase`, and only in
 * this order. The last one must end the phase (a startPhase move or a noOp).
 */
export function addOrderedForPhase(phase: number, ...items: Groupable[]): MoveSpec[] {
	if (items.length === 0) return [];
	const implied = serial(...items);
	if (implied.moves.length === 0) return [];
	const last = implied.moves[implied.moves.length - 1];
	if (!last.endsPhaseProgression) {
		throw new Error(
			'Error earlier in move processing: The end of your phase run for phase ' + phase +
				' did not end with a StartPhase move, which is typical. If this was intentional, end that phase with a moves.NoOp to override this error.',
		);
	}
	const names = groupMoveNames(implied.group);
	if (names.length !== implied.moves.length) throw new Error('progression group and move list disagree');
	return addForPhase(phase, ...implied.moves.map(move => ({ ...move, progression: implied.group })));
}

// ----- turn-taking --------------------------------------------------------

function currentPlayerInvalid(state: RulesState): boolean {
	const current = state.currentPlayerIndex();
	return !state.valid(current) || current < 0;
}

function turnHooks(state: RulesState): TurnHooks {
	const hooks = state.game.spec.player.traits?.turn;
	if (!hooks) throw new Error('The current player interface did not implement PlayerTurnFinisher');
	return hooks;
}

/** FinishTurn.Apply: end the current player's turn and start the next player's. */
export function finishTurnApply(state: RulesState): void {
	const hooks = turnHooks(state);
	hooks.resetForTurnEnd(state, state.currentPlayerIndex());
	const next = state.next(state.currentPlayerIndex());
	state.g.CurrentPlayer = next;
	hooks.resetForTurnStart(state, next);
}

/**
 * moves.FinishTurn: a fix-up that advances to the next player once the
 * current player's turn is done (the player trait `turn`).
 */
export function finishTurn(options: { name?: string; helpText?: string } = {}): MoveSpec {
	return {
		name: options.name ?? 'Finish Turn',
		helpText: options.helpText ?? "Advances to the next player when the current player's turn is done.",
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (currentPlayerInvalid(state)) return 'Current player is not valid';
			const notDone = turnHooks(state).turnDone(state, state.currentPlayerIndex());
			if (notDone !== null) return 'The current player is not done with their turn because ' + notDone;
			return null;
		},
		apply: finishTurnApply,
	};
}

/**
 * moves.ForceFinishTurn: the admin ends the current player's turn whether or
 * not it is done. Never a fix-up, and its legality ignores phases, as in Go.
 */
export function forceFinishTurn(options: { name: string; helpText?: string }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText ?? "Advances to the next player when the current player's turn is done.",
		fixUp: false,
		legal(state, _move, proposer) {
			if (proposer !== ADMIN_PLAYER_INDEX) return 'ForceFinishTurn can only be proposed by AdminPlayerIndex';
			if (currentPlayerInvalid(state)) return 'Current player is not valid';
			return null;
		},
		apply: finishTurnApply,
	};
}

/**
 * moves.AnyPlayer.Legal, for a move with base 'anyPlayer': any seated player
 * may make the move for themself. Call it first in the move's `legal`.
 *
 * As in Go, it SETS the move's TargetPlayerIndex to the proposer when the
 * move still has its default. That is how the applied move comes to name
 * its proposer, and it also means a later check of the same move object by
 * another proposer sees the first one's index.
 */
export function anyPlayerLegal(state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult): LegalResult {
	const err = base();
	if (err !== null) return err;
	let target = move.fields.TargetPlayerIndex as number;
	if (target === OBSERVER_PLAYER_INDEX && proposer >= 0) {
		target = proposer;
		move.fields.TargetPlayerIndex = proposer;
	}
	if (target < 0) return 'target player must be a seated player, not a special index';
	if (target >= state.numPlayers()) return 'target player index is out of bounds';
	if (!playerIndexEquivalent(target, proposer)) return 'you can only make this move for yourself';
	if (state.game.spec.player.traits?.seat && state.players[target].SeatFilled !== true) return 'your seat is not yet filled';
	return null;
}

/** behaviors.PlayerRole. Set the player trait `role`; the game needs an enum named "role". */
export const roleProps: Readonly<Record<string, PropDef>> = { Role: { Type: 'enum', Enum: 'role', Sanitize: 'other:hidden' } };

// ----- phases -------------------------------------------------------------

/** moves.StartPhase: a fix-up that enters `phase`. Ends an ordered phase. */
export function startPhase(options: { phase: number; phaseName: string; name?: string; helpText?: string } & Partial<MoveSpec>): MoveSpec {
	const { phase, phaseName, ...rest } = options;
	return {
		fixUp: true,
		endsPhaseProgression: true,
		apply: state => state.setPhase(phase),
		...rest,
		name: options.name ?? 'Start Phase ' + phaseName,
		helpText: options.helpText ?? 'Enters phase ' + phaseName,
	};
}

/** moves.NoOp: a fix-up that does nothing; may end an ordered phase. */
export function noOp(options: { name?: string; helpText?: string } = {}): MoveSpec {
	return {
		name: options.name ?? 'No Op',
		helpText: options.helpText ?? 'A move that does nothing and is primarily used in specific move progression situations.',
		fixUp: true,
		endsPhaseProgression: true,
		apply: () => undefined,
	};
}

// ----- seating and activity -----------------------------------------------

function suffixed(name: string, suffix: string | undefined): string {
	return suffix ? name + ' - ' + suffix : name;
}

function targetOf(move: MoveInstance): number {
	return move.fields.TargetPlayerIndex as number;
}

function targetOutOfRange(state: RulesState, move: MoveInstance): boolean {
	const target = targetOf(move);
	return target < 0 || target >= state.numPlayers();
}

/**
 * moves.SeatPlayer: applied when the host injects a seat. Fills the seat
 * and, if players can be inactive, leaves the new player inactive until an
 * activate move runs.
 */
export function seatPlayer(options: { nameSuffix?: string } = {}): MoveSpec {
	return {
		name: suffixed('Seat Player', options.nameSuffix),
		helpText: 'Marks the next available seat as seated, which when done will mean the next player is part of the game',
		fixUp: true,
		seatsPlayer: true,
		fields: { TargetPlayerIndex: { type: 'playerIndex' } },
		defaults(state, move) {
			const pending = state.env.pendingSeat;
			if (pending !== null && pending >= 0 && pending < state.numPlayers()) {
				move.fields.TargetPlayerIndex = pending;
				return;
			}
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.players[i].SeatClosed === true || state.players[i].SeatFilled === true) continue;
				move.fields.TargetPlayerIndex = i;
				return;
			}
		},
		legal(state, move, proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (state.env.pendingSeat === null) return 'No player to seat';
			if (proposer !== ADMIN_PLAYER_INDEX) return 'This move may only be proposed by an admin';
			if (targetOutOfRange(state, move)) return 'TargetPlayerIndex is invalid';
			const player = state.players[targetOf(move)];
			if (player.SeatClosed === true) return 'The selected seat was closed';
			if (player.SeatFilled === true) return 'The selected seat was already filled';
			return null;
		},
		apply(state, move) {
			if (state.env.pendingSeat === null) throw new Error('No player to seat');
			if (targetOutOfRange(state, move)) throw new Error('TargetPlayerIndex is invalid');
			const player = state.players[targetOf(move)];
			player.SeatFilled = true;
			player.SeatClosed = true;
			if (state.game.spec.player.traits?.inactive) player.PlayerInactive = true;
		},
	};
}

/** moves.ActivateInactivePlayer: a fix-up that activates one inactive player; repeats until none is left. */
export function activateInactivePlayer(options: { nameSuffix?: string; helpText?: string } = {}): MoveSpec {
	const legalTarget = (state: RulesState, move: MoveInstance): LegalResult => {
		if (targetOutOfRange(state, move)) return 'Invalid TargetPlayerIndex';
		if (!state.playerIsInactive(targetOf(move))) return 'The selected player is not inactive; there must be no inactive players to activate';
		return null;
	};
	return {
		name: suffixed('Activate Inactive Player', options.nameSuffix),
		helpText: options.helpText ?? 'Activates any players who are not currently active.',
		fixUp: true,
		allowMultipleInProgression: true,
		fields: { TargetPlayerIndex: { type: 'playerIndex' } },
		defaults(state, move) {
			for (let i = 0; i < state.numPlayers(); i++) {
				if (!state.playerIsInactive(i)) continue;
				move.fields.TargetPlayerIndex = i;
				return;
			}
		},
		legal: (state, move, _proposer, base) => base() ?? legalTarget(state, move),
		apply(state, move) {
			const err = legalTarget(state, move);
			if (err !== null) throw new Error(err);
			state.players[targetOf(move)].PlayerInactive = false;
		},
	};
}

/** moves.InactivateEmptySeat: a fix-up that marks one empty seat inactive; repeats. Only in a game that seats players. */
export function inactivateEmptySeat(options: { nameSuffix?: string; helpText?: string } = {}): MoveSpec {
	const legalTarget = (state: RulesState, move: MoveInstance): LegalResult => {
		if (targetOutOfRange(state, move)) return 'Invalid TargetPlayerIndex';
		if (state.players[targetOf(move)].SeatFilled === true) {
			return 'The selected player seat is already filled, not empty. There must not be any seats left to apply to';
		}
		if (state.playerIsInactive(targetOf(move))) return 'Player is already inactive. There must not be any any seats left to apply to';
		return null;
	};
	return {
		name: suffixed('Inactivate Empty Seat', options.nameSuffix),
		helpText: options.helpText ?? 'Marks any empty seats as being inactive, so other game logic will skip them',
		fixUp: true,
		allowMultipleInProgression: true,
		fields: { TargetPlayerIndex: { type: 'playerIndex' } },
		defaults(state, move) {
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.players[i].SeatFilled === true || state.playerIsInactive(i)) continue;
				move.fields.TargetPlayerIndex = i;
				return;
			}
		},
		legal(state, move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (!state.record.seating) return "Game will never seat players, so we shouldn't inactivate any, or we'd inactivate all of them";
			return legalTarget(state, move);
		},
		apply(state, move) {
			const err = legalTarget(state, move);
			if (err !== null) throw new Error(err);
			state.players[targetOf(move)].PlayerInactive = true;
		},
	};
}

/**
 * moves.WaitForEnoughPlayers: a fix-up that does nothing but hold a
 * progression until enough active players are seated. In a game that never
 * seats players it is legal at once.
 */
export function waitForEnoughPlayers(options: { nameSuffix?: string; helpText?: string; targetCount?: number } = {}): MoveSpec {
	return {
		name: suffixed('Wait For Enough Players', options.nameSuffix),
		helpText: options.helpText ?? 'Waits until at least target count players are active and seated before applying itself',
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (!state.record.seating) return null;
			const target = options.targetCount ?? state.game.spec.minNumPlayers;
			const seated = state.numSeatedActivePlayers();
			if (seated < target) return 'Only ' + seated + ' are seated, but move requires at least ' + target;
			const ready = state.game.spec.readyToStart ? state.game.spec.readyToStart(state) : null;
			if (ready !== null) return 'not ready to start: ' + ready;
			return null;
		},
		apply: () => undefined,
	};
}

/**
 * moves.DefaultRoundSetup: activate every inactive player, wait for enough
 * of them, then inactivate the seats still empty.
 */
export function defaultRoundSetup(options: { nameSuffix?: string } = {}): InstalledGroup {
	return serial(optional(activateInactivePlayer(options)), waitForEnoughPlayers(options), optional(inactivateEmptySeat(options)));
}

// ----- dealing ------------------------------------------------------------

function previousMoveName(state: RulesState): string {
	const history = state.committed.history;
	return history.length === 0 ? '' : history[history.length - 1].Name;
}

/**
 * moves.DealCountComponents: a fix-up that deals one component from a game
 * stack to one player each time it applies, going round the players (from
 * the current player) `targetCount` times. It keeps its place in the
 * RoundRobin properties of the game state (roundRobinProps).
 */
export function dealCountComponents(options: { name: string; helpText: string; gameStack: string; playerStack: string; targetCount?: number }): MoveSpec {
	const numRounds = options.targetCount ?? 1;
	// RoundRobinNumRounds.ConditionMet: enough rounds, or no players at all.
	const conditionMet = (state: RulesState): boolean => (state.g.RRRoundCount as number) >= numRounds || state.numPlayers() === 0;
	const starter = (state: RulesState): number => state.ensureValid(state.currentPlayerIndex());
	// RoundRobin.nextPlayerIndex, for a round robin whose players never
	// report their own condition as met (PlayerConditionMet is false).
	const nextPlayer = (state: RulesState): { player: number; roundSkip: boolean } => {
		let current = state.g.RRHasStarted === true ? state.ensureValid(state.g.RRLastPlayer as number) : state.previous(state.ensureValid(starter(state)));
		current = state.next(current);
		return { player: current, roundSkip: current === state.ensureValid(state.g.RRStarterPlayer as number) };
	};
	return {
		name: options.name,
		helpText: options.helpText,
		fixUp: true,
		allowMultipleInProgression: true,
		legal(state, move, _proposer, base) {
			let err = base();
			if (err !== null) return err;
			if (state.g.RRHasStarted !== true) {
				if (move.name === previousMoveName(state)) {
					return "can't start this round robin move because the last move was also part of this round robin";
				}
			} else {
				err = base();
				if (err !== null) return err;
				if (conditionMet(state)) return 'the condition was met, so the move is no longer legal';
			}
			const { player } = nextPlayer(state);
			if (player < 0 || player >= state.numPlayers()) return 'runtime error: index out of range [' + player + ']';
			return state.gameStack(options.gameStack).mayMoveCountTo(state.playerStack(player, options.playerStack), 1);
		},
		apply(state) {
			if (state.g.RRHasStarted !== true) {
				const first = starter(state);
				state.g.RRLastPlayer = state.previous(first);
				state.g.RRStarterPlayer = first;
				state.g.RRRoundCount = 0;
				state.g.RRHasStarted = true;
			}
			if (conditionMet(state)) {
				state.g.RRHasStarted = false;
				return;
			}
			const { player } = nextPlayer(state);
			state.gameStack(options.gameStack).moveCountTo(state.playerStack(player, options.playerStack), 1);
			state.g.RRLastPlayer = player;
			if (nextPlayer(state).roundSkip) state.g.RRRoundCount = (state.g.RRRoundCount as number) + 1;
			if (conditionMet(state)) state.g.RRHasStarted = false;
		},
	};
}

// ----- draw and discard ---------------------------------------------------

/**
 * moves.ShuffleDiscardIntoDraw: a fix-up for a game with the drawDiscardPair
 * trait. When the draw stack is empty and the discard stack is not, moves
 * the discards back and shuffles.
 */
export function shuffleDiscardIntoDraw(options: { name?: string; helpText?: string } = {}): MoveSpec {
	const pair = (state: RulesState): { draw: Stack; discard: Stack } => {
		const names = state.game.spec.game.traits?.drawDiscardPair;
		if (!names) throw new Error('gameState does not embed behaviors.DrawDiscardPair');
		return { draw: state.gameStack(names.draw), discard: state.gameStack(names.discard) };
	};
	return {
		name: options.name ?? 'Shuffle Discard Into Draw',
		helpText: options.helpText ?? 'When the draw pile is empty, shuffles the discard pile back in.',
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			const { draw, discard } = pair(state);
			if (!(draw.numComponents() === 0 && discard.numComponents() > 0)) return 'draw stack is not empty or discard stack has no components';
			const may = discard.mayMoveAllTo(draw);
			if (may !== null) return 'discard cannot be moved into draw stack: ' + may;
			return null;
		},
		apply(state) {
			const { draw, discard } = pair(state);
			discard.moveAllTo(draw);
			draw.shuffle();
		},
	};
}

/** What a reusable move contributes to a MoveSpec: spread it in, and call `apply` from your own apply. */
export type MoveParts = Pick<MoveSpec, 'base' | 'planEnabled' | 'contributes' | 'sourceProperty' | 'destinationProperty'> & {
	apply: (state: RulesState, move: MoveInstance) => void;
};

/**
 * moves.DrawToPlayer: the current player draws the top component of a game
 * stack into the first slot of one of their own. Contributes the check that
 * the component may move.
 */
export function drawToPlayer(options: { gameStack: string; playerStack: string }): MoveParts {
	return {
		base: 'currentPlayer',
		planEnabled: true,
		contributes: base => [...base, mayMoveFirstTo('game.' + options.gameStack, 'player.' + options.playerStack)],
		apply(state) {
			const source = state.gameStack(options.gameStack);
			const current = state.currentPlayerIndex();
			if (current < 0 || current >= state.numPlayers()) throw new Error('the game has no valid current player');
			const first = source.firstComponentIndex();
			if (first < 0 || source.componentAt(first) === null) throw new Error('the draw stack was empty');
			source.moveComponentToFirstSlot(first, state.playerStack(current, options.playerStack));
		},
	};
}

interface StackPath {
	kind: 'game' | 'player' | 'moveField';
	prop: string;
	moveField: string;
}

function parseStackPath(spec: string): StackPath {
	if (spec.startsWith('players[move.')) {
		const close = spec.indexOf(']');
		if (close === -1 || spec[close + 1] !== '.') throw new Error('invalid stack property spec ' + JSON.stringify(spec));
		return { kind: 'moveField', prop: spec.slice(close + 2), moveField: spec.slice('players[move.'.length, close) };
	}
	if (spec.startsWith('game.')) return { kind: 'game', prop: spec.slice(5), moveField: '' };
	if (spec.startsWith('player.')) return { kind: 'player', prop: spec.slice(7), moveField: '' };
	if (spec.indexOf('.') >= 0) throw new Error('invalid stack property spec ' + JSON.stringify(spec) + ': unknown path kind');
	return { kind: 'game', prop: spec, moveField: '' };
}

function legalPath(path: StackPath): string {
	if (path.kind === 'player') return 'player.' + path.prop;
	if (path.kind === 'moveField') return 'players[move.' + path.moveField + '].' + path.prop;
	return 'game.' + path.prop;
}

function resolveStack(state: RulesState, move: MoveInstance, path: StackPath): Stack {
	if (path.kind === 'game') return state.gameStack(path.prop);
	const player = path.kind === 'player' ? state.currentPlayerIndex() : (move.fields[path.moveField] as number);
	if (player < 0 || player >= state.numPlayers()) throw new Error('stack property did not name a concrete player');
	return state.playerStack(player, path.prop);
}

/**
 * moves.MoveComponentToSlot: the current player moves a component (the
 * first of the source stack, or the one at `sourceSlotField`) into the slot
 * named by the move's `slotField`. Stacks are named by path: "game.X",
 * "player.X" (the current player's) or "players[move.Field].X". Contributes
 * the check that the component may move to that slot.
 */
export function moveComponentToSlot(options: { source: string; destination: string; slotField: string; sourceSlotField?: string }): MoveParts {
	const source = parseStackPath(options.source);
	const destination = parseStackPath(options.destination);
	return {
		base: 'currentPlayer',
		planEnabled: true,
		sourceProperty: options.source,
		destinationProperty: options.destination,
		contributes(base: LegalSpec[]): LegalSpec[] {
			const specs = base.filter(spec => spec.name !== 'stackConstraints');
			const slot = 'move.' + options.slotField;
			if (options.sourceSlotField) {
				return [...specs, mayMoveToSlot(legalPath(source), legalPath(destination), 'move.' + options.sourceSlotField, slot)];
			}
			return [...specs, mayMoveFirstToSlot(legalPath(source), legalPath(destination), slot)];
		},
		apply(state, move) {
			const from = resolveStack(state, move, source);
			const to = resolveStack(state, move, destination);
			const index = options.sourceSlotField ? (move.fields[options.sourceSlotField] as number) : from.firstComponentIndex();
			if (index < 0 || from.componentAt(index) === null) throw new Error('there was no component to move');
			from.moveComponent(index, to, move.fields[options.slotField] as number);
		},
	};
}

export { OBSERVER_PLAYER_INDEX };
