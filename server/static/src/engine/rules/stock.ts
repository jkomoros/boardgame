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

/**
 * What a reusable move contributes to a MoveSpec: spread it in, and call
 * `apply` from your own apply. `legal` is the stock Legal (its checks are
 * all in the plan, so it is `base()`), for a game move that overrides Legal
 * and calls the embedded one.
 */
export type MoveParts = Pick<MoveSpec, 'base' | 'planEnabled' | 'contributes' | 'sourceProperty' | 'destinationProperty'> & {
	apply: (state: RulesState, move: MoveInstance) => void;
	legal: (state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult;
};

/** The stock Legal of a move whose own checks are only the base ones (moves.Default.Legal). */
export function baseLegal(_state: RulesState, _move: MoveInstance, _proposer: number, base: () => LegalResult): LegalResult {
	return base();
}

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
		legal: baseLegal,
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
		legal: baseLegal,
		apply(state, move) {
			const from = resolveStack(state, move, source);
			const to = resolveStack(state, move, destination);
			const index = options.sourceSlotField ? (move.fields[options.sourceSlotField] as number) : from.firstComponentIndex();
			if (index < 0 || from.componentAt(index) === null) throw new Error('there was no component to move');
			from.moveComponent(index, to, move.fields[options.slotField] as number);
		},
	};
}

// ===== workstream 9: every stock move as parts, and the rest of moves/ ====

/**
 * A stock move's callable parts: the spec to spread, and its Apply, Legal
 * and DefaultsForState as functions a game move that overrides one of them
 * calls where its Go calls the embedded stock method. `legal` is always
 * present (`base()` for a stock move whose Legal is moves.Default's).
 * Structurally the same as stockRegistry.ts's StockParts.
 */
export interface StockMoveParts {
	spec: MoveSpec;
	apply: (state: RulesState, move: MoveInstance) => void;
	legal: (state: RulesState, move: MoveInstance, proposer: number, base: () => LegalResult) => LegalResult;
	defaults?: (state: RulesState, move: MoveInstance) => void;
}

/** The parts of a stock move built by one of the helpers here. */
export function stockParts(spec: MoveSpec): StockMoveParts {
	const parts: StockMoveParts = { spec, apply: spec.apply, legal: spec.legal ?? baseLegal };
	if (spec.defaults) parts.defaults = spec.defaults;
	return parts;
}

// ----- behaviors not ported before ---------------------------------------

/**
 * behaviors.PlayerColor: an enum property Color (enum "color") on the player
 * state. Set the player trait `color`; viewers then get the framework's
 * computed ColorValue and AvailableColors.
 */
export const playerColorProps: Readonly<Record<string, PropDef>> = { Color: { Type: 'enum', Enum: 'color' } };

/** behaviors.ComponentColor: the Color value (enum "color") of a deck's components, for DeckSpec.props. */
export const componentColorProps: Readonly<Record<string, { Type: 'enum'; Enum: string }>> = { Color: { Type: 'enum', Enum: 'color' } };

// ----- the base moves -------------------------------------------------------

interface NamedOptions {
	name: string;
	helpText?: string;
}

/** moves.Default as a whole move: legal by the base checks alone. The game supplies apply. */
export function defaultMove(options: NamedOptions & { apply: MoveSpec['apply'] }): MoveSpec {
	return { name: options.name, helpText: options.helpText ?? 'A default move that does nothing on its own', fixUp: false, legal: baseLegal, apply: options.apply };
}

/** moves.FixUp: a fix-up legal by the base checks alone. */
export function fixUp(options: NamedOptions & { apply: MoveSpec['apply'] }): MoveSpec {
	return {
		name: options.name,
		helpText: options.helpText ?? 'A move that is applied automatically to fix up the state after a player makes a move.',
		fixUp: true,
		legal: baseLegal,
		apply: options.apply,
	};
}

/** moves.FixUpMulti: a fix-up that may apply several times in a row in a progression. */
export function fixUpMulti(options: NamedOptions & { apply: MoveSpec['apply'] }): MoveSpec {
	return { ...fixUp(options), allowMultipleInProgression: true };
}

/** moves.CurrentPlayer: only the current player may propose it (TargetPlayerIndex, contributed check). */
export function currentPlayerMove(options: NamedOptions & { apply: MoveSpec['apply'] }): MoveSpec {
	return { name: options.name, helpText: options.helpText ?? 'A move by the current player.', fixUp: false, base: 'currentPlayer', legal: baseLegal, apply: options.apply };
}

/** moves.Done: a player's move that does nothing, legal by the base checks alone. */
export function done(options: { name?: string; helpText?: string } = {}): MoveSpec {
	return {
		name: options.name ?? 'Done',
		helpText: options.helpText ?? 'The player has signaled that they are done applying moves in this group and are ready to move on.',
		fixUp: false,
		legal: baseLegal,
		apply: () => undefined,
	};
}

// ----- stack paths, for the moves configured with one ---------------------

/**
 * The stack a configured path names ("X" or "game.X", "player.X" for the
 * current player's, "players[move.F].X"), or null where Go's
 * resolveConfiguredStack returns nil.
 */
function configuredStack(state: RulesState, move: MoveInstance, spec: string): Stack | null {
	const path = parseStackPath(spec);
	if (path.kind === 'game') return state.gameStack(path.prop);
	const player = path.kind === 'player' ? state.currentPlayerIndex() : (move.fields[path.moveField] as number);
	if (player < 0 || player >= state.numPlayers()) return null;
	return state.playerStack(player, path.prop);
}

/** The name Go's stackName gives a configured stack in a fallback move name: the property, unqualified. */
function stackLabel(spec: string): string {
	return parseStackPath(spec).prop;
}

function errorText(thrown: unknown): string {
	return thrown instanceof Error ? thrown.message : String(thrown);
}

// ----- DefaultComponent ---------------------------------------------------

/**
 * moves.DefaultComponent: a fix-up that finds, in `sourceStack`, the first
 * component whose Legal (interfaces.LegalComponent; here the deck's
 * `componentLegal`) passes, and records its slot in ComponentIndex. The game
 * supplies apply (Go: "You must provide your own Apply()"); checkers crowns
 * a token with it.
 *
 * `legalType` is the configured LegalType; with `legalTypeEnum` too, a
 * component's Legal receives it as that enum's value, otherwise null (Go:
 * nil), as legalTypeValImpl does.
 */
export function defaultComponent(options: {
	sourceStack: string;
	name?: string;
	helpText?: string;
	legalType?: number;
	legalTypeEnum?: string;
	apply?: MoveSpec['apply'];
}): MoveSpec {
	const legalType = options.legalType ?? 0;
	const legalTypeValue = legalType !== 0 && options.legalTypeEnum !== undefined ? legalType : null;
	const componentLegal = (state: RulesState, deck: string, index: number): LegalResult | undefined => {
		const fn = state.game.spec.decks[deck]?.componentLegal;
		return fn ? fn(state, index, legalTypeValue) : undefined;
	};
	const label = stackLabel(options.sourceStack);
	return {
		name: options.name ?? 'Default Component For ' + label + ' LegalType ' + legalType,
		helpText: options.helpText ?? 'Searches ' + label + ' for a component that returns nil for Legal() with LegalType ' + legalType,
		fixUp: true,
		allowMultipleInProgression: true,
		fields: { ComponentIndex: { type: 'int' } },
		defaults(state, move) {
			const stack = configuredStack(state, move, options.sourceStack);
			if (stack === null) return;
			for (let i = 0; i < stack.len(); i++) {
				const index = stack.componentAt(i);
				if (index === null || index < 0) continue;
				const deck = stack.deckName;
				if (state.game.spec.decks[deck]?.components[index] === null) continue;
				const verdict = componentLegal(state, deck, index);
				if (verdict === undefined || verdict !== null) continue;
				move.fields.ComponentIndex = i;
				return;
			}
		},
		legal(state, move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			const stack = configuredStack(state, move, options.sourceStack);
			if (stack === null) return 'SourceStack returned nil stack';
			const slot = move.fields.ComponentIndex as number;
			const index = slot >= 0 && slot < stack.len() ? stack.componentAt(slot) : null;
			if (index === null || index < 0) return "ComponentIndex didn't specify a valid component";
			const deck = stack.deckName;
			if (state.game.spec.decks[deck]?.components[index] === null) return "Specified component's values were nil";
			const verdict = componentLegal(state, deck, index);
			if (verdict === undefined) return "Specified Component didn't satisfy LegalComponent";
			return verdict;
		},
		apply:
			options.apply ??
			(() => {
				throw new Error('moves.DefaultComponent: the game must supply Apply');
			}),
	};
}

// ----- seats ----------------------------------------------------------------

function firstPlayer(state: RulesState, test: (player: number) => boolean): number {
	for (let i = 0; i < state.numPlayers(); i++) if (test(i)) return i;
	return -1;
}

function seatFilled(state: RulesState, player: number): boolean {
	return state.players[player].SeatFilled === true;
}

function seatClosed(state: RulesState, player: number): boolean {
	return state.players[player].SeatClosed === true;
}

function requireSeat(state: RulesState): string | null {
	return state.game.spec.player.traits?.seat ? null : "Player state didn't implement interfaces.Seater";
}

/** moves.CloseEmptySeat: a fix-up that closes one seat that is neither filled nor closed; repeats. */
export function closeEmptySeat(options: { nameSuffix?: string; helpText?: string } = {}): MoveSpec {
	const legalTarget = (state: RulesState, move: MoveInstance): LegalResult => {
		if (targetOutOfRange(state, move)) return 'Invalid TargetPlayerIndex';
		const seat = requireSeat(state);
		if (seat !== null) return seat;
		if (seatFilled(state, targetOf(move))) return 'The selected player seat is already filled, not empty. There must not be any seats left to apply to';
		if (seatClosed(state, targetOf(move))) return 'The selected player seat is already closed. There must not be any seats left to apply to';
		return null;
	};
	return {
		name: suffixed('Close Empty Seat', options.nameSuffix),
		helpText: options.helpText ?? 'Marks any empty seats as being not open for more people to be seated.',
		fixUp: true,
		allowMultipleInProgression: true,
		fields: { TargetPlayerIndex: { type: 'playerIndex' } },
		defaults(state, move) {
			if (requireSeat(state) !== null) return;
			const target = firstPlayer(state, i => !seatFilled(state, i) && !seatClosed(state, i));
			if (target >= 0) move.fields.TargetPlayerIndex = target;
		},
		legal: (state, move, _proposer, base) => base() ?? legalTarget(state, move),
		apply(state, move) {
			const err = legalTarget(state, move);
			if (err !== null) throw new Error(err);
			state.players[targetOf(move)].SeatClosed = true;
		},
	};
}

/** Go's activationTarget (moves/inactive_player.go), for the two seat-filtered activation moves. */
function activateSeat(options: { nameSuffix?: string; helpText?: string }, filled: boolean): MoveSpec {
	const wrongSeat = filled
		? 'The selected player seat is empty; only filled seats should be activated by this move'
		: 'The selected player seat is filled; only empty seats should be activated by this move';
	const alreadyActive = filled
		? 'Player is already active. There must not be any filled inactive seats left to apply to'
		: 'Player is already active. There must not be any empty inactive seats left to apply to';
	const matches = (state: RulesState, player: number): boolean => state.game.spec.player.traits?.seat === true && seatFilled(state, player) === filled;
	const legalTarget = (state: RulesState, move: MoveInstance): LegalResult => {
		if (targetOutOfRange(state, move)) return 'Invalid TargetPlayerIndex';
		const seat = requireSeat(state);
		if (seat !== null) return seat;
		if (!matches(state, targetOf(move))) return wrongSeat;
		if (!state.playerIsInactive(targetOf(move))) return alreadyActive;
		return null;
	};
	return {
		name: suffixed(filled ? 'Activate Filled Seat' : 'Activate Empty Seat', options.nameSuffix),
		helpText:
			options.helpText ??
			(filled
				? 'Activates any player who has been seated but not yet included in play, without reopening empty seats'
				: 'Re-activates any empty seats that were previously marked inactive'),
		fixUp: true,
		allowMultipleInProgression: true,
		fields: { TargetPlayerIndex: { type: 'playerIndex' } },
		defaults(state, move) {
			const target = firstPlayer(state, i => matches(state, i) && state.playerIsInactive(i));
			if (target >= 0) move.fields.TargetPlayerIndex = target;
		},
		legal: (state, move, _proposer, base) => base() ?? legalTarget(state, move),
		apply(state, move) {
			const err = legalTarget(state, move);
			if (err !== null) throw new Error(err);
			if (!state.game.spec.player.traits?.inactive) throw new Error("Player state didn't implement interfaces.PlayerInactiver");
			state.players[targetOf(move)].PlayerInactive = false;
		},
	};
}

/** moves.ActivateEmptySeat: a fix-up that re-activates one empty, inactive seat; repeats. */
export function activateEmptySeat(options: { nameSuffix?: string; helpText?: string } = {}): MoveSpec {
	return activateSeat(options, false);
}

/** moves.ActivateFilledSeat: a fix-up that activates one seated, inactive player; repeats. */
export function activateFilledSeat(options: { nameSuffix?: string; helpText?: string } = {}): MoveSpec {
	return activateSeat(options, true);
}

/**
 * moves.CloseAllSeats: a player's move (a gathering start move) that closes
 * every open seat once at least `targetCount` (default the game's minimum)
 * players are seated and active. `requireAdmin` is moves.WithRequireAdmin;
 * the game-administrator behavior is not ported, so with it set only the
 * admin index (-2) or a game without that behavior passes, as in Go.
 */
export function closeAllSeats(options: { name?: string; helpText?: string; targetCount?: number; requireAdmin?: boolean } = {}): MoveSpec {
	return {
		name: options.name ?? 'Close All Seats',
		helpText: options.helpText ?? 'Closes all unfilled seats so no more players can join and the game can begin',
		fixUp: false,
		gatheringStart: true,
		legal(state, _move, proposer, base) {
			const err = base();
			if (err !== null) return err;
			const seated = state.game.spec.player.traits?.seat === true;
			if (!seated || firstPlayer(state, i => !seatFilled(state, i) && !seatClosed(state, i)) < 0) return 'No unfilled, unclosed seats to close';
			const target = options.targetCount ?? state.game.spec.minNumPlayers;
			const count = state.numSeatedActivePlayers();
			if (count < target) return 'Only ' + count + ' players are seated, but at least ' + target + ' are required';
			if (options.requireAdmin === true && proposer !== ADMIN_PLAYER_INDEX && proposer < 0) return 'only the game administrator can make this move';
			const ready = state.game.spec.readyToStart ? state.game.spec.readyToStart(state) : null;
			if (ready !== null) return 'not ready to start: ' + ready;
			return null;
		},
		apply(state) {
			if (state.game.spec.player.traits?.seat !== true) return;
			for (let i = 0; i < state.numPlayers(); i++) {
				if (!seatFilled(state, i) && !seatClosed(state, i)) state.players[i].SeatClosed = true;
			}
		},
	};
}

// ----- ShuffleStack -------------------------------------------------------

/** moves.ShuffleStack: a fix-up that shuffles the configured source stack. */
export function shuffleStack(options: { source: string; name?: string; helpText?: string }): MoveSpec {
	return {
		name: options.name ?? 'Shuffle ' + stackLabel(options.source),
		helpText: options.helpText ?? 'Shuffles ' + stackLabel(options.source),
		fixUp: true,
		legal: baseLegal,
		apply(state, move) {
			const stack = configuredStack(state, move, options.source);
			if (stack === null) throw new Error('PrimaryStack returned a nil stack');
			stack.shuffle();
		},
	};
}

// ----- round robins -------------------------------------------------------

/**
 * What a moves.RoundRobin's embedding move supplies (interfaces
 * ConditionMetter, RoundRobinActioner and the unexported
 * playerConditionMet and roundRobinStarterPlayer).
 */
export interface RoundRobinHooks {
	/** ConditionMet(state) == nil: the round robin is over. */
	conditionMet(state: RulesState): boolean;
	/** PlayerConditionMet(player): skip this player. Default: never (RoundRobin's). */
	playerConditionMet?(state: RulesState, player: number): boolean;
	/** RoundRobinStarterPlayer. Default: the current player, made valid. */
	starter?(state: RulesState): number;
	/** RoundRobinAction(player). Throw to fail. */
	action(state: RulesState, move: MoveInstance, player: number): void;
}

/** RoundRobin.ConditionMet's default: every player's condition is met. */
export function roundRobinAllPlayersMet(state: RulesState, playerConditionMet: (state: RulesState, player: number) => boolean): boolean {
	for (let i = 0; i < state.numPlayers(); i++) if (!playerConditionMet(state, i)) return false;
	return true;
}

/** RoundRobin.nextPlayerIndex. */
function roundRobinNext(state: RulesState, hooks: RoundRobinHooks): { player: number; roundSkip: boolean } {
	const condition = hooks.playerConditionMet ?? (() => false);
	const starter = hooks.starter ?? ((s: RulesState) => s.ensureValid(s.currentPlayerIndex()));
	let current = state.g.RRHasStarted === true ? state.ensureValid(state.g.RRLastPlayer as number) : state.previous(state.ensureValid(starter(state)));
	let counter = 0;
	let roundSkip = false;
	const n = state.numPlayers();
	while (counter <= n) {
		current = state.next(current);
		if (current === state.ensureValid(state.g.RRStarterPlayer as number)) roundSkip = true;
		if (current < 0 || current >= n) throw new Error('runtime error: index out of range [' + current + ']');
		if (!condition(state, current)) break;
		counter++;
	}
	if (counter > n) return { player: current, roundSkip: true };
	return { player: current, roundSkip };
}

/**
 * moves.RoundRobin's Legal and Apply over the RoundRobin game properties
 * (roundRobinProps). `afterLegal` is what an embedding move's Legal adds
 * once RoundRobin.Legal passed (DealCountComponents: may the next
 * component move).
 */
export function roundRobinParts(
	hooks: RoundRobinHooks,
	afterLegal?: (state: RulesState, move: MoveInstance, next: number) => LegalResult,
): Pick<MoveSpec, 'legal' | 'apply'> & { legal: NonNullable<MoveSpec['legal']> } {
	const starter = hooks.starter ?? ((s: RulesState) => s.ensureValid(s.currentPlayerIndex()));
	return {
		legal(state, move, _proposer, base) {
			let err = base();
			if (err !== null) return err;
			if (state.g.RRHasStarted !== true) {
				if (move.name === previousMoveName(state)) return "can't start this round robin move because the last move was also part of this round robin";
			} else {
				err = base();
				if (err !== null) return err;
				if (hooks.conditionMet(state)) return 'the condition was met, so the move is no longer legal';
			}
			if (!afterLegal) return null;
			return afterLegal(state, move, roundRobinNext(state, hooks).player);
		},
		apply(state, move) {
			if (state.g.RRHasStarted !== true) {
				const first = starter(state);
				state.g.RRLastPlayer = state.previous(first);
				state.g.RRStarterPlayer = first;
				state.g.RRRoundCount = 0;
				state.g.RRHasStarted = true;
			}
			if (hooks.conditionMet(state)) {
				state.g.RRHasStarted = false;
				return;
			}
			const { player } = roundRobinNext(state, hooks);
			try {
				hooks.action(state, move, player);
			} catch (thrown) {
				throw new Error('RoundRobinAction returned error: ' + errorText(thrown));
			}
			state.g.RRLastPlayer = player;
			if (roundRobinNext(state, hooks).roundSkip) state.g.RRRoundCount = (state.g.RRRoundCount as number) + 1;
			if (hooks.conditionMet(state)) state.g.RRHasStarted = false;
		},
	};
}

/** The eight moves built on DealCountComponents (moves/deal_components.go, collect_components.go). */
export type DealKind =
	| 'DealCountComponents'
	| 'DealComponentsUntilPlayerCountReached'
	| 'DealComponentsUntilGameCountLeft'
	| 'DealAllComponents'
	| 'CollectCountComponents'
	| 'CollectComponentsUntilPlayerCountLeft'
	| 'CollectComponentsUntilGameCountReached'
	| 'CollectAllComponents';

/**
 * Any of the deal and collect moves: a round-robin fix-up that moves one
 * component per application between a game stack and each player's stack.
 * Deal moves go game to player, collect moves player to game. When it stops:
 *
 * | kind | a player is skipped when | the round robin ends when |
 * |---|---|---|
 * | DealCountComponents, CollectCountComponents | never | `targetCount` rounds are done |
 * | DealComponentsUntilPlayerCountReached | the player has `targetCount` | every player is skipped |
 * | CollectComponentsUntilPlayerCountLeft, CollectAllComponents (0) | the player has at most `targetCount` | every player is skipped |
 * | DealComponentsUntilGameCountLeft, DealAllComponents (0) | never | the game stack has at most `targetCount` |
 * | CollectComponentsUntilGameCountReached | never | the game stack has at least `targetCount` |
 *
 * dealCountComponents above is the same move, kept as the hand ports use it.
 */
export function dealComponents(options: { kind: DealKind; gameStack: string; playerStack: string; targetCount?: number; name?: string; helpText?: string }): MoveSpec {
	const kind = options.kind;
	const collect = kind.startsWith('Collect');
	const target = kind === 'DealAllComponents' || kind === 'CollectAllComponents' ? 0 : (options.targetCount ?? 1);
	const playerCount = (state: RulesState, player: number): number => state.playerStack(player, options.playerStack).numComponents();
	const gameCount = (state: RulesState): number => state.gameStack(options.gameStack).numComponents();
	let playerConditionMet: (state: RulesState, player: number) => boolean = () => false;
	if (kind === 'DealComponentsUntilPlayerCountReached') playerConditionMet = (state, player) => playerCount(state, player) >= target;
	if (kind === 'CollectComponentsUntilPlayerCountLeft' || kind === 'CollectAllComponents') playerConditionMet = (state, player) => playerCount(state, player) <= target;
	const allMet = (state: RulesState): boolean => roundRobinAllPlayersMet(state, playerConditionMet);
	let conditionMet: (state: RulesState) => boolean;
	switch (kind) {
		case 'DealCountComponents':
		case 'CollectCountComponents':
			// RoundRobinNumRounds.ConditionMet, NumRounds being TargetCount.
			conditionMet = state => (state.g.RRRoundCount as number) >= target || allMet(state);
			break;
		case 'DealComponentsUntilGameCountLeft':
		case 'DealAllComponents':
			conditionMet = state => gameCount(state) <= target || allMet(state);
			break;
		case 'CollectComponentsUntilGameCountReached':
			conditionMet = state => gameCount(state) >= target || allMet(state);
			break;
		default:
			conditionMet = allMet;
	}
	const ends = (state: RulesState, player: number): { source: Stack; destination: Stack } => {
		const game = state.gameStack(options.gameStack);
		const mine = state.playerStack(player, options.playerStack);
		return collect ? { source: mine, destination: game } : { source: game, destination: mine };
	};
	const rr = roundRobinParts(
		{
			conditionMet,
			playerConditionMet,
			action(state, _move, player) {
				const { source, destination } = ends(state, player);
				source.moveCountTo(destination, 1);
			},
		},
		(state, _move, next) => {
			const { source, destination } = ends(state, next);
			return source.mayMoveCountTo(destination, 1);
		},
	);
	return {
		name: options.name ?? dealFallbackName(kind, options.gameStack, options.playerStack, target),
		helpText: options.helpText ?? dealFallbackHelpText(kind, options.gameStack, options.playerStack, target),
		fixUp: true,
		allowMultipleInProgression: true,
		...rr,
	};
}

/** The FallbackName Go derives for each deal and collect move. */
function dealFallbackName(kind: DealKind, game: string, player: string, count: number): string {
	switch (kind) {
		case 'DealCountComponents':
			return 'Deal Components From ' + game + ' In GameState To ' + player + ' In PlayerState To Each Player ' + count + ' Times';
		case 'DealComponentsUntilPlayerCountReached':
			return 'Deal Components From ' + game + ' In GameState To ' + player + ' In Each PlayerState Until Each Player Has ' + count;
		case 'DealComponentsUntilGameCountLeft':
			return 'Deal Components From ' + game + ' In GameState To ' + player + ' In Each PlayerState Until Game Stack Has ' + count + ' Total';
		case 'DealAllComponents':
			// Go passes its two stack names in the wrong order here.
			return 'Deal All Components From ' + player + ' In GameState To ' + game + ' In Each PlayerState';
		case 'CollectCountComponents':
			return 'Collect Components From ' + player + ' in each PlayerState To ' + game + ' in GameState ' + count + ' Times Per Player';
		case 'CollectComponentsUntilPlayerCountLeft':
			return 'Collect Components From ' + player + ' in each PlayerState To ' + game + ' In GameState Until Each Player Is Down To ' + count;
		case 'CollectComponentsUntilGameCountReached':
			return 'Collect Components From ' + player + ' in each PlayerState To ' + game + ' In GameState Until The Game Has ' + count + ' Total';
		case 'CollectAllComponents':
			return 'Collect All Components From ' + player + ' In Each PlayerState To ' + game + ' In GameState';
	}
}

/** The FallbackHelpText Go derives for each deal and collect move. */
function dealFallbackHelpText(kind: DealKind, game: string, player: string, count: number): string {
	switch (kind) {
		case 'DealCountComponents':
			return 'Deals ' + count + ' components from ' + game + ' in GameState to ' + player + ' in each PlayerState';
		case 'DealComponentsUntilPlayerCountReached':
			return 'Deals components from ' + game + ' in GameState to ' + player + ' in each PlayerState until each player has ' + count;
		case 'DealComponentsUntilGameCountLeft':
			return 'Deals components from ' + game + ' in GameState to ' + player + ' in each PlayerState until the game stack has ' + count + ' left';
		case 'DealAllComponents':
			// The same swap as its name.
			return 'Deals all components from ' + player + ' in GameState to ' + game + ' in each PlayerState';
		case 'CollectCountComponents':
			return 'Collects ' + count + ' components from ' + player + ' in each PlayerState to ' + game + ' in GameState';
		case 'CollectComponentsUntilPlayerCountLeft':
			return 'Collects components from ' + player + ' in each PlayerState to ' + game + ' in GameState until each player has ' + count + ' left';
		case 'CollectComponentsUntilGameCountReached':
			return 'Collects components from ' + player + ' in each PlayerState to ' + game + ' in GameState until the game has ' + count + ' total';
		case 'CollectAllComponents':
			return 'Collects all components from ' + player + ' in each PlayerState to ' + game + ' in GameState';
	}
}

// ----- moving components between two stacks ------------------------------

/** The four moves built on MoveCountComponents (moves/move_components.go). */
export type MoveComponentsKind = 'MoveCountComponents' | 'MoveComponentsUntilCountReached' | 'MoveComponentsUntilCountLeft' | 'MoveAllComponents';

/** countMovesApplied: how many of the latest records, in the current phase, are this move. */
function countMovesApplied(state: RulesState, name: string): number {
	const records = state.committed.history;
	const phase = state.currentPhaseKey();
	let count = 0;
	for (let i = records.length - 1; i >= 0; i--) {
		if (records[i].Phase !== phase || records[i].Name !== name) break;
		count++;
	}
	return count;
}

/**
 * Any of the move-components fix-ups: one component from `source` to
 * `destination` per application, until
 *
 * | kind | count | done when |
 * |---|---|---|
 * | MoveCountComponents | moves of this name applied in a row in this phase | count == `targetCount` |
 * | MoveComponentsUntilCountReached | the destination's components | count >= `targetCount` |
 * | MoveComponentsUntilCountLeft, MoveAllComponents (0) | the source's components | count <= `targetCount` |
 *
 * Their Legal skips the generic stack-constraints check
 * (legalWithoutStackConstraints), so the spec names no source or
 * destination property and none is contributed.
 */
export function moveComponents(options: { kind: MoveComponentsKind; source: string; destination: string; targetCount?: number; name?: string; helpText?: string }): MoveSpec {
	const kind = options.kind;
	const target = kind === 'MoveAllComponents' ? 0 : (options.targetCount ?? 1);
	const stacks = (state: RulesState, move: MoveInstance): { source: Stack | null; destination: Stack | null } => ({
		source: configuredStack(state, move, options.source),
		destination: configuredStack(state, move, options.destination),
	});
	const count = (state: RulesState, move: MoveInstance): number => {
		const { source, destination } = stacks(state, move);
		if (kind === 'MoveCountComponents') return countMovesApplied(state, move.name);
		if (kind === 'MoveComponentsUntilCountReached') return destination === null ? 0 : destination.numComponents();
		return source === null ? 0 : source.numComponents();
	};
	const conditionMet = (state: RulesState, move: MoveInstance): boolean => {
		const c = count(state, move);
		if (kind === 'MoveCountComponents') return c === target;
		if (c < 0 || target < 0) return false;
		return kind === 'MoveComponentsUntilCountReached' ? c >= target : c <= target;
	};
	const source = stackLabel(options.source);
	const destination = stackLabel(options.destination);
	let name: string;
	let helpText: string;
	switch (kind) {
		case 'MoveCountComponents':
			name = 'Move ' + target + ' Components From ' + source + ' To ' + destination;
			helpText = 'Moves ' + target + ' components from ' + source + ' to ' + destination;
			break;
		case 'MoveComponentsUntilCountReached':
			name = 'Move Components From ' + source + ' Until ' + destination + ' Has ' + target;
			helpText = 'Moves components from ' + source + ' to ' + destination + ' until ' + destination + ' has ' + target;
			break;
		case 'MoveComponentsUntilCountLeft':
			name = 'Move Components To ' + destination + ' Until ' + source + ' Has ' + target;
			helpText = 'Moves components from ' + source + ' to ' + destination + ' until ' + source + ' has ' + target + ' left';
			break;
		case 'MoveAllComponents':
			name = 'Move All Components From ' + source + ' To ' + destination;
			helpText = 'Moves all components from ' + source + ' to ' + destination;
			break;
	}
	return {
		name: options.name ?? name,
		helpText: options.helpText ?? helpText,
		fixUp: true,
		allowMultipleInProgression: true,
		legal(state, move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (conditionMet(state, move)) return 'the condition was met, so the move is no longer legal';
			const ends = stacks(state, move);
			if (ends.source === null) return 'Source was nil';
			if (ends.destination === null) return 'Destination was nil';
			const c = count(state, move);
			if (c < 0) return 'Count returned a negative value';
			if (target < 0) return 'TargetCount returned a negative value';
			return ends.source.mayMoveCountTo(ends.destination, Math.abs(target - c));
		},
		apply(state, move) {
			const ends = stacks(state, move);
			if (ends.source === null) throw new Error('Source was nil');
			if (ends.destination === null) throw new Error('Destination was nil');
			ends.source.moveCountTo(ends.destination, 1);
		},
	};
}

export { OBSERVER_PLAYER_INDEX };
