// The engine step. createGame builds a game's version 0 and runs its set-up
// chain; propose takes the stored state, the move history and ONE outside
// input, and returns every version that input produced. Ports Game.setUp,
// Game.applyMove and the fix-up recursion in game.go, the creator-input
// binder in move_input_bind.go, and the timer completion path in timer.go.
//
// Nothing here keeps anything between calls: all that persists is what the
// caller stores (see ../README.md, "What the host does").

import { ADMIN_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from '../legal/types.ts';
import { EngineState } from '../state.ts';
import type { StateJSON } from '../state.ts';
import { GameSpecError, emptyStateJSON, newVariant } from './compile.ts';
import { type CompiledGame, type CompiledMove, type Environment, RulesState, findMove } from './state.ts';
import {
	type TimerGuard,
	type TimerRecord,
	activeTimers,
	consumeTimer,
	finalizeTimerDeadlines,
	syncTimerRecords,
	timerGuardError,
	timerRecords,
} from './timers.ts';
import type { GameRecord, MoveInstance, MoveRecord } from './types.ts';

/** game.go maxRecurseCount. */
export const MAX_FIX_UP_RECURSION = 256;

/** ErrTooManyFixUps. The versions applied before the cap stay applied. */
export const ERR_TOO_MANY_FIX_UPS = 'we recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal';

/** One state version an input produced. */
export interface VersionOutput {
	version: number;
	/** The storage record of the state: what the host persists. */
	state: StateJSON;
	/** The move that produced this version; null for version 0. */
	move: MoveRecord | null;
	/** Whether the game was finished as of this version, and who won. */
	finished: boolean;
	winners: number[];
}

export interface StepResult {
	/** Every version produced, in order. Empty when the input was refused. */
	versions: VersionOutput[];
	/**
	 * Why the input was refused, in Go's words. Set together with versions
	 * only for the fix-up cap, where Go has already committed them.
	 */
	error?: string;
}

/** What the host supplies with every call. */
export interface StepOptions {
	/** The current time, in milliseconds since the Unix epoch. */
	now: number;
}

export interface CreateOptions extends StepOptions {
	id: string;
	/** Fresh randomness, kept secret: ids and every draw derive from it. */
	secretSalt: string;
	/** 0 or absent: the game's default. */
	numPlayers?: number;
	variant?: Readonly<Record<string, string>> | null;
	agents?: string[];
	/** Whether players will be seated through seat inputs. */
	seating?: boolean;
}

export interface CreateResult extends StepResult {
	/** Null when creation was refused. */
	record: GameRecord | null;
}

/** What the host holds for a game between calls. */
export interface GameSnapshot {
	record: GameRecord;
	/** The move records of versions 1 through the current one. */
	history: readonly MoveRecord[];
	/** The current state's storage record. */
	state: StateJSON;
}

/** One outside input. */
export type Input =
	/** A player (or the admin) proposes a move, giving only the fields a proposer supplies. */
	| { kind: 'propose'; name: string; input?: Readonly<Record<string, unknown>>; proposer: number; expectedVersion?: number }
	/** A recorded move is proposed again with all of its stored fields (a replay). */
	| { kind: 'record'; name: string; fields: Readonly<Record<string, unknown>>; proposer: number; expectedVersion?: number }
	/** A player takes a seat: `player`, or the next open seat when negative. */
	| { kind: 'seat'; player: number }
	/**
	 * A timer's deadline has passed. `id` names the timer; absent, the
	 * earliest active one is meant. With `force`, the deadline is not checked
	 * against `now` (Go's ForceNextTimer, for tests).
	 */
	| { kind: 'timer'; id?: string; force?: boolean };

interface Run {
	game: CompiledGame;
	record: GameRecord;
	history: MoveRecord[];
	env: Environment;
	now: number;
	versions: VersionOutput[];
	finished: boolean;
}

function winnersOf(result: { finished: boolean; winners: number[] }): number[] {
	return result.finished ? result.winners.slice() : [];
}

function thrownText(thrown: unknown): string {
	return thrown instanceof Error ? thrown.message : String(thrown);
}

/**
 * base.GameDelegate.CheckGameFinished, or the game's own. Highest (or
 * lowest) score among players who are neither inactive nor eliminated wins.
 */
export function checkGameFinished(state: RulesState): { finished: boolean; winners: number[] } {
	const end = state.game.spec.end;
	if (!end) return { finished: false, winners: [] };
	if (end.check) return end.check(state);
	if (!end.conditionMet || !end.conditionMet(state)) return { finished: false, winners: [] };
	const lowScoreWins = end.lowScoreWins === true;
	const gameScore = state.game.spec.player.traits?.gameScore;
	const score = (player: number): number => (end.playerScore ? end.playerScore(state, player) : gameScore ? gameScore(state, player) : 0);
	const mayWin = (player: number): boolean => !state.playerIsInactive(player) && !state.playerIsEliminated(player);
	let extreme = lowScoreWins ? 2147483647 : 0;
	for (let i = 0; i < state.numPlayers(); i++) {
		if (!mayWin(i)) continue;
		const value = score(i);
		if (lowScoreWins ? value < extreme : value > extreme) extreme = value;
	}
	const winners: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) if (mayWin(i) && score(i) === extreme) winners.push(i);
	return { finished: true, winners };
}

/** state.validateBeforeSave. */
function saveError(state: RulesState): string | null {
	const conservation = state.engine.conservationError();
	if (conservation !== null) return 'component conservation failed: ' + conservation;
	const def = state.game.def;
	const check = (props: typeof def.Game.Props, node: Record<string, unknown>): string | null => {
		for (const name of Object.keys(props).sort()) {
			if (props[name].Type !== 'playerIndex') continue;
			const value = node[name] as number;
			if (!state.withinBounds(value)) return name + ' was an invalid PlayerIndex, with value ' + value;
		}
		return null;
	};
	let err = check(def.Game.Props, state.g);
	if (err !== null) return err;
	for (const player of state.players) {
		err = check(def.Player.Props, player);
		if (err !== null) return err;
	}
	return null;
}

function moveRecord(move: MoveInstance, version: number, initiator: number, phase: number, proposer: number): MoveRecord {
	return { Name: move.name, Version: version, Initiator: initiator, Phase: phase, Proposer: proposer, Blob: { ...move.fields } };
}

/**
 * validateMoveChoiceInputDomain: a move whose field is a projected choice
 * must name a value from that choice's domain.
 */
function choiceDomainError(state: RulesState, compiled: CompiledMove, move: MoveInstance, proposer: number): string | null {
	const choice = compiled.spec.choices;
	if (!choice) return null;
	if (choice.stack) {
		const value = move.fields[choice.field] as number;
		const scope = choice.stack.scope;
		if (proposer === ADMIN_PLAYER_INDEX && scope === 'proposing-player') {
			return 'resolve stack-slot choice field ' + JSON.stringify(choice.field) + ': admin has no proposing-player stack';
		}
		if (scope === 'proposing-player' && (proposer < 0 || proposer >= state.numPlayers())) {
			return 'resolve stack-slot choice field ' + JSON.stringify(choice.field) + ': stack-slot actor ' + proposer + ' is invalid';
		}
		const view = scope === 'game' ? state.gameView(choice.stack.property) : state.playerView(proposer, choice.stack.property);
		if (value < 0 || value >= view.len() || view.componentAt(value) === null) {
			return value + ' is not an occupied slot in ' + scope + ' stack ' + JSON.stringify(choice.stack.property);
		}
		return null;
	}
	const exclude = choice.exclude ?? [];
	const value = move.fields[choice.field];
	for (const excluded of exclude) if (value === excluded) return value + ' is not an allowed value for ' + choice.field;
	return null;
}

/**
 * Game.applyMove. Applies one move to `current` and then, recursively, the
 * fix-up moves that follow it. Returns Go's error text, or null.
 */
function applyMove(run: Run, current: RulesState, move: MoveInstance, proposer: number, isFixUp: boolean, recurseCount: number, initiator: number, guard: TimerGuard | null): string | null {
	const versionToSet = current.version + 1;
	if (run.finished) return 'Game was already finished';
	const compiled = run.game.movesByName[move.name];
	if (!compiled) return 'That move is not configured for this game.';
	if (initiator === -1) initiator = versionToSet;

	if (guard !== null) {
		const err = timerGuardError(current.engine, guard);
		if (err !== null) return err;
	}
	if (!current.valid(proposer)) return 'The proposer was not valid.';
	if (proposer === OBSERVER_PLAYER_INDEX) return 'The proposer was the ObserverPlayerIndex, but observers may never make moves.';

	const domain = choiceDomainError(current, compiled, move, proposer);
	if (domain !== null) return domain;

	let legal: string | null;
	try {
		legal = current.legalError(move, proposer);
	} catch (thrown) {
		legal = thrownText(thrown);
	}
	if (legal !== null) return legal;

	const phase = current.currentPhaseKey();
	const next = new RulesState(run.game, current.engine.copy(versionToSet), run.record, run.history, run.env, current);
	if (guard !== null) {
		const err = consumeTimer(next.engine, guard);
		if (err !== null) return err;
	}
	try {
		compiled.spec.apply(next, move);
	} catch (thrown) {
		return "The move's apply function returned an error:" + thrownText(thrown);
	}
	const invalid = saveError(next);
	if (invalid !== null) return 'The modified state had an invalidity, so the move was not applied. ' + invalid;

	const result = checkGameFinished(next);
	finalizeTimerDeadlines(next.pendingDeadlines, run.now);

	// The save boundary. From here the version exists.
	const record = moveRecord(move, versionToSet, initiator, phase, proposer);
	run.history.push(record);
	if (result.finished) run.finished = true;
	run.versions.push({ version: versionToSet, state: next.engine.toJSON(), move: record, finished: result.finished, winners: winnersOf(result) });
	if (compiled.spec.seatsPlayer) run.env.pendingSeat = null;

	if (recurseCount > MAX_FIX_UP_RECURSION) return ERR_TOO_MANY_FIX_UPS;
	if (run.finished) return null;

	// Go reads the committed state back from storage, so nothing built while
	// applying the move (the component index in particular) carries over.
	const committed = new RulesState(run.game, next.engine.copy(versionToSet), run.record, run.history, run.env);
	const fixUp = committed.proposeFixUpMove();
	if (fixUp !== null) {
		const err = applyMove(run, committed, fixUp, ADMIN_PLAYER_INDEX, true, recurseCount + 1, initiator, null);
		if (err !== null) {
			if (err === ERR_TOO_MANY_FIX_UPS) return err;
			return 'Applying the fix up move failed: ' + recurseCount + ': ' + err;
		}
	}
	return null;
}

/**
 * Game.MoveWithInput: a move of the named type with its defaults for the
 * state and the proposer's input bound onto it. Only fields a proposer
 * supplies ('required' and 'server-defaulted') are accepted.
 */
export function moveWithInput(state: RulesState, name: string, input: Readonly<Record<string, unknown>>): { move: MoveInstance | null; error?: string } {
	const compiled = findMove(state.game, name);
	const move = state.newMove(name);
	if (!compiled || move === null) return { move: null, error: 'unknown move ' + JSON.stringify(name) };
	if (compiled.spec.fixUp) return { move: null, error: 'move ' + JSON.stringify(name) + ' is engine-owned and is not a creator move' };
	// Go ranges over the input map here, so with several unacceptable fields
	// the one it names is arbitrary. The first in sorted order is named.
	for (const fieldName of Object.keys(input).sort()) {
		const field = compiled.inputFields.find(candidate => candidate.name === fieldName);
		if (!field || (field.disposition !== 'required' && field.disposition !== 'server-defaulted')) {
			return { move: null, error: 'move ' + JSON.stringify(compiled.spec.name) + ' does not accept creator field ' + JSON.stringify(fieldName) };
		}
	}
	for (const field of compiled.inputFields) {
		if (!Object.prototype.hasOwnProperty.call(input, field.name)) {
			if (field.disposition === 'required') return { move: null, error: 'move ' + JSON.stringify(name) + ' requires field ' + JSON.stringify(field.name) };
			continue;
		}
		const value = input[field.name];
		if (value === null || value === undefined) return { move: null, error: 'field ' + JSON.stringify(field.name) + ' must not be null' };
		const wrong = (goType: string): { move: null; error: string } => ({
			move: null,
			error: 'move ' + JSON.stringify(name) + ' field ' + JSON.stringify(field.name) + ': json: cannot unmarshal ' + jsonKind(value) + ' into Go value of type ' + goType,
		});
		switch (field.codec) {
			case 'integer':
				if (typeof value !== 'number' || value !== Math.floor(value)) return wrong('int');
				if (compiled.fields[field.name].type === 'enum') {
					const e = state.game.enums.get(compiled.fields[field.name].enum);
					if (e === null || e.keys().indexOf(value) < 0) {
						return { move: null, error: 'move ' + JSON.stringify(name) + ' field ' + JSON.stringify(field.name) + ': that value is not legal for this enum' };
					}
					move.fields[field.name] = e.stringOf(value);
				} else {
					move.fields[field.name] = value;
				}
				break;
			case 'player-index':
				if (typeof value !== 'number' || value !== Math.floor(value)) return wrong('boardgame.PlayerIndex');
				move.fields[field.name] = value;
				break;
			case 'boolean':
				if (typeof value !== 'boolean') return wrong('bool');
				move.fields[field.name] = value;
				break;
			case 'string':
				if (typeof value !== 'string') return wrong('string');
				move.fields[field.name] = value;
				break;
			case 'enum': {
				if (typeof value !== 'string') return wrong('string');
				const e = state.game.enums.get(compiled.fields[field.name].enum);
				const key = e === null ? null : e.valueFromString(value);
				if (e === null || key === null) {
					return { move: null, error: 'move ' + JSON.stringify(name) + ' field ' + JSON.stringify(field.name) + ': that value is not legal for this enum' };
				}
				move.fields[field.name] = e.stringOf(key);
				break;
			}
			default:
				return { move: null, error: 'move ' + JSON.stringify(name) + ' field ' + JSON.stringify(field.name) + ': unsupported input codec ' + JSON.stringify(field.codec) };
		}
	}
	return { move };
}

/** How encoding/json names a value it could not put into a Go type. */
function jsonKind(value: unknown): string {
	if (typeof value === 'string') return 'string';
	if (typeof value === 'number') return value === Math.floor(value) ? 'number' : 'number ' + String(value);
	if (typeof value === 'boolean') return 'bool';
	if (Array.isArray(value)) return 'array';
	return 'object';
}

/** MoveStorageRecord.inflateForState: the move's defaults, overwritten by every stored field. */
export function moveFromRecord(state: RulesState, name: string, fields: Readonly<Record<string, unknown>>): MoveInstance | null {
	const move = state.newMove(name);
	if (move === null) return null;
	const compiled = state.game.movesByName[move.name];
	for (const fieldName of compiled.fieldNames) {
		if (Object.prototype.hasOwnProperty.call(fields, fieldName)) move.fields[fieldName] = fields[fieldName];
	}
	return move;
}

/** Loads a stored state as the rules see it. */
export function loadState(game: CompiledGame, snapshot: GameSnapshot, env: Environment = { pendingSeat: null }): RulesState {
	const identity = { gameId: snapshot.record.id, secretSalt: snapshot.record.secretSalt };
	const engine = EngineState.load(game.def, identity, snapshot.history.length, snapshot.state);
	syncTimerRecords(engine);
	return new RulesState(game, engine, snapshot.record, snapshot.history, env);
}

/**
 * GameManager.NewGame and Game.setUp: builds version 0 and applies the
 * fix-up moves legal at once. versions[0] is version 0 itself.
 */
export function createGame(game: CompiledGame, options: CreateOptions): CreateResult {
	const spec = game.spec;
	let numPlayers = options.numPlayers ?? 0;
	if (numPlayers === 0) numPlayers = spec.defaultNumPlayers;
	const refuse = (error: string): CreateResult => ({ record: null, versions: [], error });
	if (numPlayers < 1) return refuse('The number of players, ' + numPlayers + ' is not legal. There must be one or more players.');
	if (numPlayers < spec.minNumPlayers || numPlayers > spec.maxNumPlayers) return refuse('The number of players, ' + numPlayers + ' was not legal.');
	let variant: Record<string, string> | null;
	try {
		variant = newVariant(spec, options.variant);
	} catch (thrown) {
		if (!(thrown instanceof GameSpecError)) throw thrown;
		return refuse('That variation is not legal for this game: ' + thrown.message);
	}
	let agents = options.agents;
	if (agents && agents.length !== numPlayers) return refuse('If agentNames is not nil, it must have length equivalent to numPlayers.');
	if (!agents) {
		agents = [];
		for (let i = 0; i < numPlayers; i++) agents.push('');
	}

	const record: GameRecord = { name: spec.name, id: options.id, secretSalt: options.secretSalt, numPlayers, variant, agents, seating: options.seating === true };
	const engine = EngineState.load(game.def, { gameId: record.id, secretSalt: record.secretSalt }, 0, emptyStateJSON(game, numPlayers), { skipConservation: true });
	syncTimerRecords(engine);
	const run: Run = { game, record, history: [], env: { pendingSeat: null }, now: options.now, versions: [], finished: false };
	const state = new RulesState(game, engine, record, run.history, run.env);

	try {
		if (spec.setUp.begin) spec.setUp.begin(state, variant ?? {});
	} catch (thrown) {
		return refuse('BeginSetUp errored: ' + thrownText(thrown));
	}
	for (const deckName of Object.keys(game.def.Decks).sort()) {
		for (let i = 0; i < game.def.Decks[deckName].Size; i++) {
			const where = 'Distributing components failed for deck ' + deckName + ':' + i + ':';
			let stack;
			try {
				stack = spec.setUp.distribute(state, deckName, i);
			} catch (thrown) {
				return refuse(where + thrownText(thrown));
			}
			if (!stack) return refuse(where + ' the delegate returned no stack.');
			if (stack.slotsRemaining() < 1) return refuse(where + ' the stack the delegate returned had no more slots.');
			const attachment = engine.stackAttachmentError(stack);
			if (attachment !== null) return refuse(where + ' returned stack is not attached: ' + attachment);
			if (stack.deckName !== deckName) return refuse(where + ' returned stack belongs to a different deck');
			stack.insertStarterComponent(i);
		}
	}
	try {
		if (spec.setUp.finish) spec.setUp.finish(state);
	} catch (thrown) {
		return refuse('FinishSetUp errored: ' + thrownText(thrown));
	}
	const conservation = engine.conservationError();
	if (conservation !== null) return refuse('Initial state violated component conservation: ' + conservation);
	// A timer started during set-up has nothing to save its deadline against.
	finalizeTimerDeadlines(state.pendingDeadlines, options.now);

	run.versions.push({ version: 0, state: engine.toJSON(), move: null, finished: false, winners: [] });
	const fixUp = state.proposeFixUpMove();
	if (fixUp !== null) {
		const err = applyMove(run, state, fixUp, ADMIN_PLAYER_INDEX, true, 0, -1, null);
		if (err !== null) {
			if (err === ERR_TOO_MANY_FIX_UPS) return { record, versions: run.versions, error: err };
			return { record, versions: run.versions, error: 'Applying the first fix up move failed: ' + err };
		}
	}
	return { record, versions: run.versions };
}

/** Whether the game is finished at the snapshot's state, and who won. */
export function gameStatus(game: CompiledGame, snapshot: GameSnapshot): { finished: boolean; winners: number[] } {
	if (snapshot.history.length === 0) return { finished: false, winners: [] };
	const result = checkGameFinished(loadState(game, snapshot));
	return { finished: result.finished, winners: winnersOf(result) };
}

/**
 * The engine step: applies ONE outside input to the current state and
 * returns every version it produced, fix-ups included, each with its move
 * record. A refused input returns no versions and Go's error text; the
 * caller's state is untouched either way (nothing is mutated in place).
 */
export function propose(game: CompiledGame, snapshot: GameSnapshot, input: Input, options: StepOptions): StepResult {
	const env: Environment = { pendingSeat: null };
	const history = snapshot.history.slice();
	const current = new RulesState(game, loadState(game, snapshot).engine, snapshot.record, history, env);
	const run: Run = {
		game,
		record: snapshot.record,
		history,
		env,
		now: options.now,
		versions: [],
		finished: history.length > 0 && checkGameFinished(current).finished,
	};
	const done = (error: string | null): StepResult => (error === null ? { versions: run.versions } : { versions: run.versions, error });

	if (input.kind === 'propose' || input.kind === 'record') {
		let move: MoveInstance | null;
		if (input.kind === 'propose') {
			const bound = moveWithInput(current, input.name, input.input ?? {});
			if (bound.move === null) return done(bound.error ?? 'unknown move');
			move = bound.move;
		} else {
			move = moveFromRecord(current, input.name, input.fields);
			if (move === null) return done(run.finished ? 'Game was already finished' : 'That move is not configured for this game.');
		}
		if (input.expectedVersion !== undefined && input.expectedVersion !== current.version) {
			return done('move proposal used stale game version ' + input.expectedVersion + '; current version is ' + current.version);
		}
		return done(applyMove(run, current, move, input.proposer, false, 0, -1, null));
	}

	if (input.kind === 'seat') {
		// Game.triggerFixUp with a player waiting to be seated: one pass of
		// fix-up proposal, with the seat request visible to the seating move.
		if (run.finished) return done(null);
		env.pendingSeat = input.player;
		const move = current.proposeFixUpMove();
		if (move === null) return done('No fix-up move would seat the player');
		const err = applyMove(run, current, move, ADMIN_PLAYER_INDEX, false, 0, -1, null);
		if (err !== null) return done(err);
		if (env.pendingSeat !== null) return done('The game did not seat the player');
		return done(null);
	}

	// A timer tick: timerManager.fire, validated against the state alone.
	if (run.finished) return done('timer game was already finished');
	const candidates = activeTimers(snapshot.state);
	const timer = input.id === undefined ? candidates[0] : candidates.find(candidate => candidate.guard.ID === input.id);
	if (!timer) return done('timer completion was stale or already handled');
	if (!input.force && timer.deadlineMs > options.now) return done('timer deadline has not passed');
	const move = moveFromRecord(current, timer.move.Name, timer.move.Blob);
	if (move === null) return done("Couldn't find a move with name: " + timer.move.Name);
	return done(applyMove(run, current, move, ADMIN_PLAYER_INDEX, false, 0, -1, timer.guard));
}

/** The timer records of a state the engine loaded; for tests and tools. */
export function stateTimerRecords(state: RulesState): TimerRecord[] {
	return timerRecords(state.engine);
}
