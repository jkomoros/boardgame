// What one viewer is sent for one version: the game object with the
// sanitized state, the sanitized move record, the legality of each move they
// could propose, and the choices projected for them. Ports
// Game.JSONForPlayer, Game.MoveJSONForPlayer (move_sanitization.go), the
// server's move forms (server/api/main.go generateFormsWithLegality) and
// projected move choices (server/api/projected_move_choices.go).
//
// Everything here is a function of ONE version: its stored state, the move
// history up to it, and the game record. Go's JSONForPlayer mixes in the
// head of the game (Finished, Winners, Version, ActiveTimers); here those
// are the values as of the version viewed.

import { GROUP_ALL, GROUP_OTHER, GROUP_SELF } from '../definition.ts';
import type { PolicyName } from '../definition.ts';
import { type LegalPreconditionEntry, buildPreconditionLedger } from '../legal/evaluable.ts';
import { ADMIN_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from '../legal/types.ts';
import {
	computedGroupMembership,
	playerGroupKeys,
	resolveSanitizationPolicy,
	sanitizationTransformation,
	stringMembership,
	viewFor,
	visibleFacets,
} from '../sanitize.ts';
import { loadEngineState } from '../state.ts';
import type { EngineState } from '../state.ts';
import type { StateJSON, VisibilityJSON } from '../state.ts';
import { checkGameFinished, moveFromRecord } from './engine.ts';
import { type CompiledGame, type CompiledMove, newRulesState } from './state.ts';
import type { RulesState } from './state.ts';
import { activeTimersView, syncTimerRecords } from './timers.ts';
import type { GameRecord, MoveInstance, MoveRecord } from './types.ts';

/** behaviors.CSSColorForKey, by the color enum's key. */
const CSS_COLOR_FOR_KEY = ['#D32F2F', '#1976D2', '#388E3C', '#FBC02D', '#424242', '#FAFAFA', '#E64A19', '#7B1FA2', '#C2185B', '#795548', '#0097A7', '#757575'];
/** behaviors.defaultPlayerPalette: red, blue, green, yellow, orange, purple, cyan, pink, black, brown, gray, white. */
const DEFAULT_PLAYER_PALETTE = [0, 1, 2, 3, 6, 7, 10, 8, 4, 9, 11, 5].map(key => CSS_COLOR_FOR_KEY[key]);

/** One version of one game, as the host holds it. */
export interface VersionSnapshot {
	record: GameRecord;
	/** The move records of versions 1 through this one. */
	history: readonly MoveRecord[];
	/** This version's stored state. */
	state: StateJSON;
	/** The state of the version before (absent for version 0). Move visibility groups are judged against it. */
	previousState?: StateJSON | undefined;
	/**
	 * True when the engine was waiting for outside input at this version (the
	 * last version an input produced). Projected choices exist only then.
	 */
	settled: boolean;
	/** The current time, in milliseconds since the Unix epoch (for timers' time left). */
	now: number;
}

/** Game.MoveJSONForPlayer. */
export interface ViewerMove {
	AnimationKey: string;
	Version: number;
	Properties?: Record<string, unknown>;
}

export interface MoveLegality {
	name: string;
	legalForPlayer: boolean;
	legalForPlayerError?: string;
	legalForAnyone: boolean;
	defaults: Record<string, unknown>;
	preconditions?: LegalPreconditionEntry[];
}

export interface ProjectedMoveChoices {
	StateVersion: number;
	MoveChoiceProjectionSchemaFingerprint: string;
	ProjectionSchemaVersion: number;
	Status: string;
	Sets?: { MoveName: string; FieldName: string; Source: string; Candidates: { Value: unknown; Available: boolean }[] }[];
}

export interface ViewerGame {
	Name: string;
	ID: string;
	NumPlayers: number;
	Agents: string[];
	Variant: Record<string, string> | null;
	Version: number;
	Finished: boolean;
	Winners: number[];
	CurrentPlayerIndex: number;
	ActiveTimers: Record<string, { TimeLeft: number }>;
	CurrentState: StateJSON;
}

export interface ViewerOutput {
	viewer: number;
	game: ViewerGame;
	move: ViewerMove | null;
	legality: MoveLegality[];
	projectedMoveChoices?: ProjectedMoveChoices;
}

function loadVersion(game: CompiledGame, record: GameRecord, history: readonly MoveRecord[], state: StateJSON, version: number): RulesState {
	const engine = loadEngineState(game.def, { gameId: record.id, secretSalt: record.secretSalt }, version, state);
	syncTimerRecords(engine);
	return newRulesState(game, engine, record, history, { pendingSeat: null });
}

/** Whether the game is finished as of a version, and who won. Version 0 is never finished. */
export function versionStatus(game: CompiledGame, at: VersionSnapshot): { finished: boolean; winners: number[] } {
	if (at.history.length === 0) return { finished: false, winners: [] };
	const result = checkGameFinished(loadVersion(game, at.record, at.history, at.state, at.history.length));
	return { finished: result.finished, winners: result.finished ? result.winners.slice() : [] };
}

// ----- computed properties ------------------------------------------------

function cssColorForPlayer(state: RulesState, player: number): string {
	const prop = state.game.def.Player.Props.Color;
	if (prop && prop.Type === 'enum') {
		const e = state.game.enums.get(prop.Enum);
		const key = e === null ? null : e.keyOf(state.players[player].Color);
		if (key !== null && key >= 0 && key < CSS_COLOR_FOR_KEY.length) return CSS_COLOR_FOR_KEY[key];
	}
	return DEFAULT_PLAYER_PALETTE[(player < 0 ? 0 : player) % DEFAULT_PLAYER_PALETTE.length];
}

/**
 * state.computed(): the framework's computed properties and the game's
 * own, computed from the state it is given. Views pass the SANITIZED state,
 * so a value that depends on hidden information differs by viewer.
 */
export function computedProperties(state: RulesState): { Global: Record<string, unknown>; Players: Record<string, unknown>[] } {
	const spec = state.game.spec;
	const global: Record<string, unknown> = {};
	const order = state.customPlayerOrder();
	if (order !== null) global.PlayerOrder = order.slice();
	if (spec.readyToStart) {
		const err = spec.readyToStart(state);
		if (err !== null) global.ReadyToStartError = err;
	}
	const hasRole = spec.player.traits?.role === true;
	if (hasRole) {
		const roles = state.game.enums.get('role');
		if (roles !== null) global.AvailableRoles = roles.keys().map(key => ({ Key: key, Name: roles.stringOf(key) }));
	}
	const players: Record<string, unknown>[] = [];
	const gameScore = spec.player.traits?.gameScore;
	// PropertyFacetAvailable: on a sanitized state, whether the viewer may trust the property's value.
	const visibility = state.engine.sanitized ? state.engine.visibility : null;
	const valueVisible = (player: number, prop: string): boolean =>
		!state.engine.sanitized || (visibility !== null && (visibility.Players[player]?.[prop] ?? []).indexOf('values') >= 0);
	for (let i = 0; i < state.numPlayers(); i++) {
		const values: Record<string, unknown> = { Color: cssColorForPlayer(state, i), MayBeActive: state.playerMayBeActive(i) };
		if (gameScore) values.GameScore = gameScore(state, i);
		if (hasRole && valueVisible(i, 'Role')) values.RoleValue = state.players[i].Role;
		players.push(values);
	}
	for (const computed of spec.computed ?? []) {
		if (computed.scope === 'global') global[computed.name] = computed.compute(state);
		else for (let i = 0; i < players.length; i++) players[i][computed.name] = computed.compute(state, i);
	}
	return { Global: global, Players: players };
}

// ----- moves --------------------------------------------------------------

/** moveSanitizationGroupsFor: the groups a viewer is in, relative to the move's proposer. */
function moveGroups(game: CompiledGame, compiled: CompiledMove, state: RulesState, proposer: number, viewer: number): Record<string, boolean> {
	const def = game.def;
	const players = state.players;
	const proposerState = proposer >= 0 && proposer < players.length ? players[proposer] : null;
	const viewerState = viewer >= 0 && viewer < players.length ? players[viewer] : null;
	const proposerKeys = playerGroupKeys(def, proposerState);
	const viewerKeys = playerGroupKeys(def, viewerState);
	const groups = stringMembership(def, proposerKeys);
	const groupEnum = def.GroupEnum ? def.Enums[def.GroupEnum] : undefined;
	const isEnumGroup = (name: string): boolean => {
		if (!groupEnum) return false;
		for (const key of Object.keys(groupEnum.Values)) if (groupEnum.Values[key] === name) return true;
		return false;
	};
	const names: string[] = [];
	const add = (policies: Readonly<Record<string, PolicyName>>): void => {
		for (const name of Object.keys(policies)) {
			if (name === GROUP_ALL || isEnumGroup(name) || names.indexOf(name) >= 0) continue;
			names.push(name);
		}
	};
	for (const fieldName of compiled.fieldNames) add(compiled.fieldPolicies[fieldName]);
	add(compiled.namePolicy);
	for (const name of names) {
		if (proposerState === null && name !== GROUP_SELF && name !== GROUP_OTHER) {
			groups[name] = false;
			continue;
		}
		groups[name] = computedGroupMembership(def, name, proposer, viewer, proposerKeys, viewerKeys);
	}
	return groups;
}

/** Game.MoveNameVisibleToPlayer. */
function moveNameVisible(game: CompiledGame, compiled: CompiledMove, state: RulesState, proposer: number, viewer: number): boolean {
	if (viewer === ADMIN_PLAYER_INDEX) return true;
	return resolveSanitizationPolicy(compiled.namePolicy, moveGroups(game, compiled, state, proposer, viewer), 'hidden') === 'visible';
}

/** MoveInfo.MoveNamePublic: the name is visible to every viewer. */
export function moveNamePublic(compiled: CompiledMove): boolean {
	return resolveSanitizationPolicy(compiled.namePolicy, { [GROUP_ALL]: true }, 'hidden') === 'visible';
}

/** MoveInfo.AnimationKeys. */
export function animationKeys(compiled: CompiledMove): string[] {
	const keys = [compiled.spec.name];
	const hidden = compiled.spec.hiddenAnimationKey;
	if (hidden && hidden !== compiled.spec.name) keys.push(hidden);
	return keys;
}

/**
 * Game.MoveJSONForPlayer: the move record as one viewer may see it, or null
 * when they may see nothing of it. `before` is the state the move was
 * applied to.
 */
export function sanitizedMove(game: CompiledGame, before: RulesState, record: MoveRecord, viewer: number): ViewerMove | null {
	const compiled = game.movesByName[record.Name];
	if (!compiled) throw new Error("Couldn't find a move with name: " + record.Name);
	const groups = moveGroups(game, compiled, before, record.Proposer, viewer);
	let animationKey = record.Name;
	if (viewer !== ADMIN_PLAYER_INDEX && resolveSanitizationPolicy(compiled.namePolicy, groups, 'hidden') !== 'visible') {
		animationKey = compiled.spec.hiddenAnimationKey ?? '';
		if (animationKey === '') return null;
	}
	const properties: Record<string, unknown> = {};
	let any = false;
	for (const fieldName of compiled.fieldNames) {
		// Move fields are scalars: only a fully visible one carries its value.
		const policy = viewer === ADMIN_PLAYER_INDEX ? 'visible' : resolveSanitizationPolicy(compiled.fieldPolicies[fieldName], groups, 'hidden');
		if (policy !== 'visible') continue;
		properties[fieldName] = record.Blob[fieldName];
		any = true;
	}
	const result: ViewerMove = { AnimationKey: animationKey, Version: record.Version };
	if (any) result.Properties = properties;
	return result;
}

function safeLegalError(state: RulesState, move: MoveInstance, proposer: number): string | null {
	try {
		return state.legalError(move, proposer);
	} catch (thrown) {
		return thrown instanceof Error ? thrown.message : String(thrown);
	}
}

/**
 * The server's move forms for one viewer: one entry per move they are
 * offered, in registration order, each evaluated with the engine's default
 * arguments for this state. Fix-up moves are never listed.
 */
export function moveLegality(game: CompiledGame, state: RulesState, viewer: number): MoveLegality[] {
	const visibility: VisibilityJSON | null =
		viewer === ADMIN_PLAYER_INDEX ? null : visibleFacets(state.engine, sanitizationTransformation(state.engine, viewer));
	const view = { viewer, visibility, currentPlayerIndex: state.currentPlayerIndex() };
	const result: MoveLegality[] = [];
	for (const compiled of game.moves) {
		if (compiled.spec.fixUp) continue;
		const move = state.newMove(compiled.spec.name);
		if (move === null) continue;
		if (!moveNamePublic(compiled)) {
			// privateMoveAvailable: only for a viewer who may see the name and for whom the move is legal now.
			if (viewer !== ADMIN_PLAYER_INDEX) {
				if (viewer < 0) continue;
				if (!moveNameVisible(game, compiled, state, viewer, viewer)) continue;
				if (safeLegalError(state, move, viewer) !== null) continue;
			}
		}
		const form: MoveLegality = { name: compiled.spec.name, legalForPlayer: false, legalForAnyone: false, defaults: { ...move.fields } };
		const ledger = state.legalLedger(move, viewer);
		if (ledger !== null && ledger.entries !== null) form.preconditions = buildPreconditionLedger(view, ledger.entries);
		if (viewer !== OBSERVER_PLAYER_INDEX) {
			const err = safeLegalError(state, move, viewer);
			if (err === null) form.legalForPlayer = true;
			else if (err !== '') form.legalForPlayerError = err;
		}
		form.legalForAnyone = safeLegalError(state, move, ADMIN_PLAYER_INDEX) === null;
		result.push(form);
	}
	return result;
}

/**
 * projectMoveChoicesSnapshot: for a seated, valid player at a settled
 * version, each declared choice set with which candidates are legal now.
 * Null when the game declares no projections or the viewer is not eligible.
 */
export function projectedMoveChoices(game: CompiledGame, state: RulesState, viewer: number): ProjectedMoveChoices | null {
	if (viewer < 0 || !state.valid(viewer)) return null;
	if (game.choiceSchema.length === 0) return null;
	const snapshot: ProjectedMoveChoices = {
		StateVersion: state.version,
		MoveChoiceProjectionSchemaFingerprint: game.choiceSchemaFingerprint,
		ProjectionSchemaVersion: 1,
		Status: 'ready',
	};
	const sets: NonNullable<ProjectedMoveChoices['Sets']> = [];
	for (const schema of game.choiceSchema) {
		const compiled = game.movesByName[schema.moveName];
		if (!moveNameVisible(game, compiled, state, viewer, viewer)) continue;
		const values: unknown[] = [];
		if (schema.source === 'players') {
			for (let i = 0; i < state.numPlayers(); i++) if (state.valid(i)) values.push(i);
		} else if (schema.source === 'enum-values') {
			values.push(...(schema.candidateValues ?? []));
		} else if (schema.stackSource) {
			const stack = schema.stackSource.scope === 'game' ? state.gameView(schema.stackSource.property) : state.playerView(viewer, schema.stackSource.property);
			for (let slot = 0; slot < stack.len(); slot++) if (stack.componentAt(slot) !== null) values.push(slot);
		}
		const candidates: { Value: unknown; Available: boolean }[] = [];
		let available = 0;
		for (const value of values) {
			const move = state.newMove(schema.moveName);
			if (move === null) continue;
			move.fields[schema.fieldName] = value;
			const legal = safeLegalError(state, move, viewer) === null;
			if (legal) available++;
			candidates.push({ Value: value, Available: legal });
		}
		if (available === 0) continue;
		sets.push({ MoveName: schema.moveName, FieldName: schema.fieldName, Source: schema.source, Candidates: candidates });
	}
	if (sets.length > 0) snapshot.Sets = sets;
	return snapshot;
}

/**
 * Everything one viewer is sent for one version. `viewer` is a player
 * index, OBSERVER_PLAYER_INDEX, or ADMIN_PLAYER_INDEX; the admin's output is
 * unsanitized and must never be released to a player.
 */
export function viewerOutput(game: CompiledGame, at: VersionSnapshot, viewer: number): ViewerOutput {
	const version = at.history.length;
	const state = loadVersion(game, at.record, at.history, at.state, version);
	const status = checkGameFinished(state);
	const finished = version > 0 && status.finished;

	let currentPlayerIndex = OBSERVER_PLAYER_INDEX;
	const currentState = viewFor(state.engine, viewer, {
		computed: sanitized => {
			const rules = newRulesState(game, sanitized, at.record, at.history, { pendingSeat: null });
			currentPlayerIndex = rules.currentPlayerIndex();
			return computedProperties(rules);
		},
	});

	const output: ViewerOutput = {
		viewer,
		game: {
			Name: game.spec.name,
			ID: at.record.id,
			NumPlayers: at.record.numPlayers,
			Agents: at.record.agents.slice(),
			Variant: at.record.variant === null ? null : { ...at.record.variant },
			Version: version,
			Finished: finished,
			Winners: finished ? status.winners.slice() : [],
			CurrentPlayerIndex: currentPlayerIndex,
			ActiveTimers: activeTimersView(at.state, at.now),
			CurrentState: currentState,
		},
		move: null,
		legality: moveLegality(game, state, viewer),
	};

	if (version > 0) {
		const record = at.history[version - 1];
		const before = at.previousState ? loadVersion(game, at.record, at.history.slice(0, version - 1), at.previousState, version - 1) : state;
		output.move = sanitizedMove(game, before, record, viewer);
	}
	if (at.settled && viewer >= 0) {
		const choices = projectedMoveChoices(game, state, viewer);
		if (choices !== null) output.projectedMoveChoices = choices;
	}
	return output;
}

export { moveFromRecord };
