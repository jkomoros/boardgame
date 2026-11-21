// RulesState: one version of a game's state as the rules see it. It wraps the
// core's EngineState (the JSON, stacks, ids and generator) and adds what Go
// reaches through the game, its manager and its delegate: the game's
// definition, the move history, the current player and phase, timers, and
// move legality.
//
// A RulesState is what every function in a GameSpec is handed.

import type { GameDefinition, PolicyName } from '../definition.ts';
import type { LegalContext, LegalFramework, LegalMove, LegalState } from '../legal/access.ts';
import { renderLegalVerdict } from '../legal/message.ts';
import { type LegalPlan, type LegalPlanResult, evaluateLegalPlan } from '../legal/plan.ts';
import { playerIndexEnsureValid, playerIndexNext, playerIndexValid } from '../legal/playerIndex.ts';
import { ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from '../legal/types.ts';
import type { LegalRegistry } from '../legal/predicate.ts';
import type { GoRand } from '../rand.ts';
import type { Stack, StackView } from '../stack.ts';
import type { EngineState, SubStateJSON } from '../state.ts';
import { type ComponentSchema, type ValuePropDef, bindLegalMove, bindLegalState } from './binding.ts';
import type { BoundEnum, EnumRegistry } from './enums.ts';
import { matchTape } from './progression.ts';
import { type TimerRecord, ZERO_TIMER_MOVE_TIMESTAMP, findTimerRecord } from './timers.ts';
import { ZERO_TIME } from './time.ts';
import type { GameRecord, GameSpec, InputCodec, InputDisposition, MoveFieldSpec, MoveInstance, MoveRecord, MoveSpec } from './types.ts';

/** One resolved input field of a move (ResolveMoveInputFields). */
export interface InputField {
	name: string;
	disposition: InputDisposition;
	codec: InputCodec | '';
}

/** A move type, with everything derived from its spec at compile time. */
export interface CompiledMove {
	readonly spec: MoveSpec;
	readonly index: number;
	/** Every field, including the base's TargetPlayerIndex. */
	readonly fields: Readonly<Record<string, MoveFieldSpec>>;
	/** Field names, sorted. */
	readonly fieldNames: readonly string[];
	readonly schema: Readonly<Record<string, ValuePropDef>>;
	/** Per field: group to policy, from its sanitize tag (default hidden). */
	readonly fieldPolicies: Readonly<Record<string, Readonly<Record<string, PolicyName>>>>;
	/** Group to policy for the move's name. */
	readonly namePolicy: Readonly<Record<string, PolicyName>>;
	/** True when legality is a declared plan whose ledger viewers receive. */
	readonly declared: boolean;
	/** The base checks, then (when declared) the authored ones and the custom. */
	readonly plan: LegalPlan;
	/** Sorted by name. */
	readonly inputFields: readonly InputField[];
}

/** A game ready to run: its spec, the core definition, and derived tables. */
export interface CompiledGame {
	readonly spec: GameSpec;
	readonly def: GameDefinition;
	readonly enums: EnumRegistry;
	readonly components: ComponentSchema;
	readonly moves: readonly CompiledMove[];
	/** By the move's own name. */
	readonly movesByName: Readonly<Record<string, CompiledMove>>;
	/** Lower-cased name to the move's own name: Go matches move names without regard to case. */
	readonly canonicalMoveNames: Readonly<Record<string, string>>;
	/** delegate.PhaseEnum(): the enum named "phase" (or "Phase"), or null. */
	readonly phaseEnum: BoundEnum | null;
	/** The merged legal message templates (defaults plus the game's own). */
	readonly templates: Readonly<Record<string, string>>;
	readonly registry: LegalRegistry;
	/** Move names that are in no progression (they never appear on the tape). */
	readonly noProgressionMoves: Readonly<Record<string, boolean>>;
	/** BuildMoveInputSchema, as Go encodes it. */
	readonly moveInputSchema: unknown[];
	readonly moveInputSchemaFingerprint: string;
	/** BuildMoveChoiceProjectionSchema; empty when the game declares none. */
	readonly choiceSchema: ChoiceSchemaItem[];
	readonly choiceSchemaFingerprint: string;
}

/** MoveChoiceProjectionSchema. */
export interface ChoiceSchemaItem {
	moveName: string;
	fieldName: string;
	source: 'players' | 'enum-values' | 'stack-slots';
	stackSource?: { scope: 'game' | 'proposing-player'; property: string };
	candidateValues?: string[];
	disclosure: 'actor-exact';
}

/** What a proposal carries beside the state: not persisted in the state. */
export interface Environment {
	/**
	 * The seat a seat injection asked for (Go: the PlayerToSeat rendezvous),
	 * or null when none is pending. A negative value means "the next open
	 * seat". Cleared when the seating move commits.
	 */
	pendingSeat: number | null;
}

/** Marks the LegalState and LegalMove objects built here, so a plan's custom check can find its way back. */
interface Tagged {
	rules?: RulesState;
	instance?: MoveInstance;
}

const TIMER_ID_CHARS = 'ABCDEF0123456789';
const TIMER_ID_LENGTH = 16;

/** A Timer property of the state (timer.go). */
export interface TimerHandle {
	readonly state: RulesState;
	readonly record: TimerRecord;
	readonly node: { ID: string; IsTimer: true };
	/** Timer.Active. */
	active(): boolean;
	readonly id: string;
	/**
	 * Timer.Start: when `durationMs` has passed (counted from the moment the
	 * state is saved), `move` is proposed as the admin. Draws the timer's id
	 * from the state's generator: 16 draws.
	 */
	start(durationMs: number, move: MoveInstance & { version?: number; initiator?: number }): void;
	/** Timer.Cancel: whether the timer was active. */
	cancel(): boolean;
}

/** Creates a `TimerHandle`. */
export function newTimerHandle(state: RulesState, record: TimerRecord, node: { ID: string; IsTimer: true }): TimerHandle {
	const self: TimerHandle = {
		state,
		record,
		node,
		/** Timer.Active. */
		active(): boolean {
			return this.record.Status === 'active';
		},
		get id(): string {
			return this.record.ID;
		},
		/**
		 * Timer.Start: when `durationMs` has passed (counted from the moment the
		 * state is saved), `move` is proposed as the admin. Draws the timer's id
		 * from the state's generator: 16 draws.
		 */
		start(durationMs: number, move: MoveInstance & { version?: number; initiator?: number }): void {
			const phase = this.state.currentPhaseKey();
			this.record.Generation++;
			let id = '';
			const rand = this.state.rand();
			while (id.length < TIMER_ID_LENGTH) id += TIMER_ID_CHARS[rand.intn(TIMER_ID_CHARS.length)];
			this.record.ID = id;
			this.record.Deadline = ZERO_TIME;
			this.record.Status = 'active';
			this.record.Move = {
				Name: move.name,
				Version: move.version ?? 0,
				Initiator: move.initiator ?? 0,
				Timestamp: ZERO_TIMER_MOVE_TIMESTAMP,
				Phase: phase,
				Proposer: ADMIN_PLAYER_INDEX,
				Blob: { ...move.fields },
			};
			this.node.ID = id;
			this.state.pendingDeadlines.push({ record: this.record, durationMs });
		},
		/** Timer.Cancel: whether the timer was active. */
		cancel(): boolean {
			this.node.ID = '';
			this.record.ID = '';
			if (!this.active()) return false;
			this.record.Deadline = ZERO_TIME;
			this.record.Status = 'canceled';
			delete this.record.Move;
			const pending = this.state.pendingDeadlines;
			for (let i = pending.length - 1; i >= 0; i--) if (pending[i].record === this.record) pending.splice(i, 1);
			return true;
		},
	};
	return self;
}

export interface RulesState {
	readonly game: CompiledGame;
	readonly engine: EngineState;
	readonly record: GameRecord;
	/** The move records of versions 1 through the committed version. */
	readonly history: readonly MoveRecord[];
	readonly env: Environment;
	/**
	 * The committed state this one is a working copy of (itself, for a
	 * committed state). Go's state.Game().MoveByName binds a new move to the
	 * game's current state, not to the copy a move is being applied to.
	 */
	readonly committed: RulesState;
	/** Timers started on this copy whose deadline is set when it is saved. */
	readonly pendingDeadlines: { record: TimerRecord; durationMs: number }[];
	legalState: LegalState | null;
	readonly framework: LegalFramework;
	// ----- the state itself ----------------------------------------------
	readonly version: number;
	/** The game sub-state's JSON; scalars are read and written here. */
	readonly g: SubStateJSON;
	/** Each player sub-state's JSON. */
	readonly players: SubStateJSON[];
	numPlayers(): number;
	gameStack(prop: string): Stack;
	playerStack(player: number, prop: string): Stack;
	/** Any game stack property, including a merged one, read-only. */
	gameView(prop: string): StackView;
	/** Any player stack property, including a merged one, read-only. */
	playerView(player: number, prop: string): StackView;
	/** A component's static values ({} for a component that has none). */
	component(deck: string, deckIndex: number): Readonly<Record<string, unknown>>;
	/** A component's dynamic values. */
	dynamic(deck: string, deckIndex: number): SubStateJSON;
	/** state.Rand(): draw only while applying a move or setting up. */
	rand(): GoRand;
	/** The variant the game was created with (never null: {} when empty). */
	variant(): Readonly<Record<string, string>>;
	// ----- what Go asks the delegate -------------------------------------
	/** The evaluator's view of this state. */
	readonly legal: LegalState;
	/** delegate.CurrentPlayerIndex(state). */
	currentPlayerIndex(): number;
	/** behaviors.PlayerIsInactive. */
	playerIsInactive(player: number): boolean;
	/** behaviors.PlayerIsEliminated. */
	playerIsEliminated(player: number): boolean;
	/** delegate.PlayerMayBeActive. */
	playerMayBeActive(player: number): boolean;
	/** delegate.CustomPlayerOrder. */
	customPlayerOrder(): readonly number[] | null;
	/** delegate.NumSeatedActivePlayers. */
	numSeatedActivePlayers(): number;
	/** PlayerIndex.Valid. */
	valid(player: number): boolean;
	/** PlayerIndex.Next. */
	next(player: number): number;
	/** PlayerIndex.EnsureValid. */
	ensureValid(player: number): number;
	/** PlayerIndex.Previous. */
	previous(player: number): number;
	/** PlayerIndex.WithinBounds. */
	withinBounds(player: number): boolean;
	// ----- phases --------------------------------------------------------
	/** delegate.CurrentPhase: null when the game state has no Phase enum. */
	currentPhase(): { key: number; name: string; ancestors: readonly number[] | null } | null;
	/** The current phase's key, or 0 with no phase (as Go records it). */
	currentPhaseKey(): number;
	/** behaviors.PhaseBehavior.SetCurrentPhase. */
	setPhase(key: number): void;
	/**
	 * Default.legalMoveInProgression: would this move, appended to the moves
	 * applied since the current phase was entered, still match its phase's
	 * progression? Null when it would (or the move is in no progression).
	 */
	moveInProgression(moveName: string): string | null;
	// ----- moves ---------------------------------------------------------
	/**
	 * A new move of the named type with its defaults for this state (Go:
	 * MoveByNameForState). Null for an unknown name.
	 */
	newMove(name: string): (MoveInstance & { version: number; initiator: number }) | null;
	/** state.Game().MoveByName(name): bound to the game's committed state. */
	moveByName(name: string): MoveInstance & { version: number; initiator: number };
	legalContext(compiled: CompiledMove, move: MoveInstance, proposer: number): LegalContext;
	/**
	 * move.Legal(state, proposer): null when legal, otherwise the error text.
	 * A check that throws is an error, as a Go panic there would be.
	 */
	legalError(move: MoveInstance, proposer: number): string | null;
	/** GameManager.LegalEvaluateLedger: every predicate's verdict. Null for a move whose legality is not declared. */
	legalLedger(move: MoveInstance, proposer: number): LegalPlanResult | null;
	/** delegate.ProposeFixUpMove: the first fix-up move legal for the admin. */
	proposeFixUpMove(): (MoveInstance & { version: number; initiator: number }) | null;
	// ----- timers --------------------------------------------------------
	/** A Timer property of the game state. */
	timer(prop: string): TimerHandle;
	/** A Timer property of a player state. */
	playerTimer(player: number, prop: string): TimerHandle;
}

/** Creates a `RulesState`. */
export function newRulesState(game: CompiledGame, engine: EngineState, record: GameRecord, history: readonly MoveRecord[], env: Environment, committed: RulesState | null = null): RulesState {
	const self: RulesState = {
		game,
		engine,
		record,
		history,
		env,
	/**
	 * The committed state this one is a working copy of (itself, for a
	 * committed state). Go's state.Game().MoveByName binds a new move to the
	 * game's current state, not to the copy a move is being applied to.
	 */
		get committed(): RulesState {
			return committed ?? self;
		},
		pendingDeadlines: [],
		legalState: null,
		framework: {
			currentPhase: () => self.currentPhase(),
			moveInProgression: moveName => self.moveInProgression(moveName),
		},
		// ----- the state itself ----------------------------------------------
		get version(): number {
			return this.engine.version;
		},
		/** The game sub-state's JSON; scalars are read and written here. */
		get g(): SubStateJSON {
			return this.engine.game;
		},
		/** Each player sub-state's JSON. */
		get players(): SubStateJSON[] {
			return this.engine.players;
		},
		numPlayers(): number {
			return this.engine.numPlayers();
		},
		gameStack(prop: string): Stack {
			return this.engine.gameStack(prop);
		},
		playerStack(player: number, prop: string): Stack {
			return this.engine.playerStack(player, prop);
		},
		/** Any game stack property, including a merged one, read-only. */
		gameView(prop: string): StackView {
			return this.engine.view({ Group: 'game', Prop: prop });
		},
		/** Any player stack property, including a merged one, read-only. */
		playerView(player: number, prop: string): StackView {
			return this.engine.view({ Group: 'player', Player: player, Prop: prop });
		},
		/** A component's static values ({} for a component that has none). */
		component(deck: string, deckIndex: number): Readonly<Record<string, unknown>> {
			return this.game.spec.decks[deck].components[deckIndex] ?? {};
		},
		/** A component's dynamic values. */
		dynamic(deck: string, deckIndex: number): SubStateJSON {
			const values = this.engine.dynamicValues(deck, deckIndex);
			if (!values) throw new Error('deck ' + deck + ' has no dynamic component values');
			return values;
		},
		/** state.Rand(): draw only while applying a move or setting up. */
		rand(): GoRand {
			return this.engine.rand();
		},
		/** The variant the game was created with (never null: {} when empty). */
		variant(): Readonly<Record<string, string>> {
			return this.record.variant ?? {};
		},
		// ----- what Go asks the delegate -------------------------------------
		/** The evaluator's view of this state. */
		get legal(): LegalState {
			if (this.legalState === null) {
				const bound = bindLegalState(
					this.engine,
					{ enums: this.game.enums, components: this.game.components },
					{
						currentPlayerIndex: () => this.currentPlayerIndex(),
						playerMayBeActive: index => this.playerMayBeActive(index),
						customPlayerOrder: () => this.customPlayerOrder(),
						playerIsInactive: index => this.playerIsInactive(index),
					},
				);
				(bound as LegalState & Tagged).rules = this;
				this.legalState = bound;
			}
			return this.legalState;
		},
		/** delegate.CurrentPlayerIndex(state). */
		currentPlayerIndex(): number {
			const custom = this.game.spec.currentPlayerIndex;
			if (custom) return custom(this);
			const prop = this.game.def.Game.Props.CurrentPlayer;
			if (!prop || prop.Type !== 'playerIndex') return OBSERVER_PLAYER_INDEX;
			return this.ensureValid(this.g.CurrentPlayer as number);
		},
		/** behaviors.PlayerIsInactive. */
		playerIsInactive(player: number): boolean {
			const traits = this.game.spec.player.traits;
			return traits !== undefined && traits.inactive === true && this.players[player].PlayerInactive === true;
		},
		/** behaviors.PlayerIsEliminated. */
		playerIsEliminated(player: number): boolean {
			const traits = this.game.spec.player.traits;
			return traits !== undefined && traits.elimination === true && this.players[player].Eliminated === true;
		},
		/** delegate.PlayerMayBeActive. */
		playerMayBeActive(player: number): boolean {
			return !this.playerIsInactive(player);
		},
		/** delegate.CustomPlayerOrder. */
		customPlayerOrder(): readonly number[] | null {
			const traits = this.game.spec.game.traits;
			return traits && traits.playerOrder ? traits.playerOrder(this) : null;
		},
		/** delegate.NumSeatedActivePlayers. */
		numSeatedActivePlayers(): number {
			const seat = this.game.spec.player.traits?.seat === true;
			let count = 0;
			for (let i = 0; i < this.numPlayers(); i++) {
				if (this.playerIsInactive(i)) continue;
				if (seat && this.players[i].SeatFilled !== true) continue;
				count++;
			}
			return count;
		},
		/** PlayerIndex.Valid. */
		valid(player: number): boolean {
			return playerIndexValid(player, this.legal);
		},
		/** PlayerIndex.Next. */
		next(player: number): number {
			return playerIndexNext(player, this.legal);
		},
		/** PlayerIndex.EnsureValid. */
		ensureValid(player: number): number {
			return playerIndexEnsureValid(player, this.legal);
		},
		/** PlayerIndex.Previous. */
		previous(player: number): number {
			if (player === ADMIN_PLAYER_INDEX || player === OBSERVER_PLAYER_INDEX || player === ANY_PLAYER_INDEX) return player;
			const count = this.numPlayers();
			const order = this.customPlayerOrder();
			if (order !== null) {
				const pos = order.indexOf(player);
				if (pos === -1) return player;
				const n = order.length;
				for (let i = 1; i <= n; i++) {
					const candidate = order[(pos - i + n) % n];
					if (candidate < 0 || candidate >= count) return player;
					if (this.playerMayBeActive(candidate)) return candidate;
				}
				return player;
			}
			const original = player;
			player--;
			if (player < 0) player = count - 1;
			let tried = 0;
			while (!this.playerMayBeActive(player)) {
				player--;
				if (player < 0) player = count - 1;
				tried++;
				if (tried >= count) return original;
			}
			return player;
		},
		/** PlayerIndex.WithinBounds. */
		withinBounds(player: number): boolean {
			if (player === ADMIN_PLAYER_INDEX || player === OBSERVER_PLAYER_INDEX || player === ANY_PLAYER_INDEX) return true;
			return player >= 0 && player < this.numPlayers();
		},
		// ----- phases --------------------------------------------------------
		/** delegate.CurrentPhase: null when the game state has no Phase enum. */
		currentPhase(): { key: number; name: string; ancestors: readonly number[] | null } | null {
			const prop = this.game.def.Game.Props.Phase;
			if (!prop || prop.Type !== 'enum') return null;
			const e = this.game.enums.get(prop.Enum);
			if (e === null) return null;
			const key = e.keyOf(this.g.Phase);
			if (key === null) return null;
			return { key, name: e.stringOf(key), ancestors: e.ancestors ? e.ancestors(key) : null };
		},
		/** The current phase's key, or 0 with no phase (as Go records it). */
		currentPhaseKey(): number {
			const phase = this.currentPhase();
			return phase === null ? 0 : phase.key;
		},
		/** behaviors.PhaseBehavior.SetCurrentPhase. */
		setPhase(key: number): void {
			const prop = this.game.def.Game.Props.Phase;
			const e = prop ? this.game.enums.get(prop.Enum) : null;
			if (e === null) throw new Error('The gameState does not implement CurrentPhaseSetter');
			if (e.stringOf(key) === '' && e.keyOf('') !== key) throw new Error('That value is not legal for this enum');
			this.g.Phase = e.stringOf(key);
		},
		/**
		 * Default.legalMoveInProgression: would this move, appended to the moves
		 * applied since the current phase was entered, still match its phase's
		 * progression? Null when it would (or the move is in no progression).
		 */
		moveInProgression(moveName: string): string | null {
			const move = this.game.movesByName[moveName];
			if (!move) return 'moves: inProgression: no move type registered with name ' + moveName;
			const group = move.spec.progression;
			if (!group) return null;
			const phase = this.currentPhaseKey();
			const kept: string[] = [];
			const history = this.committed.history;
			for (let i = history.length - 1; i >= 0; i--) {
				const record = history[i];
				if (this.game.noProgressionMoves[record.Name]) continue;
				if (record.Phase !== phase) break;
				kept.push(record.Name);
			}
			kept.reverse();
			kept.push(moveName);
			return matchTape(group, kept);
		},
		// ----- moves ---------------------------------------------------------
		/**
		 * A new move of the named type with its defaults for this state (Go:
		 * MoveByNameForState). Null for an unknown name.
		 */
		newMove(name: string): (MoveInstance & { version: number; initiator: number }) | null {
			const compiled = findMove(this.game, name);
			if (!compiled) return null;
			name = compiled.spec.name;
			const fields: Record<string, unknown> = {};
			for (const fieldName of compiled.fieldNames) {
				const field = compiled.fields[fieldName];
				if (field.type === 'bool') fields[fieldName] = false;
				else if (field.type === 'string') fields[fieldName] = '';
				else if (field.type === 'enum') {
					const e = this.game.enums.get(field.enum);
					fields[fieldName] = e === null ? '' : e.stringOf(e.def.DefaultValue);
				} else fields[fieldName] = 0;
			}
			const move = { name, fields, version: this.version + 1, initiator: 0 };
			if (compiled.spec.base === 'currentPlayer') fields.TargetPlayerIndex = this.ensureValid(this.currentPlayerIndex());
			else if (compiled.spec.base === 'anyPlayer') fields.TargetPlayerIndex = OBSERVER_PLAYER_INDEX;
			if (compiled.spec.defaults) compiled.spec.defaults(this, move);
			return move;
		},
		/** state.Game().MoveByName(name): bound to the game's committed state. */
		moveByName(name: string): MoveInstance & { version: number; initiator: number } {
			const move = this.committed.newMove(name);
			if (move === null) throw new Error('no move named ' + name);
			return move;
		},
		legalContext(compiled: CompiledMove, move: MoveInstance, proposer: number): LegalContext {
			const bound: LegalMove & Tagged = bindLegalMove(move.fields, compiled.schema, this.game.enums);
			bound.instance = move;
			return { state: this.legal, move: bound, proposer, framework: this.framework };
		},
		/**
		 * move.Legal(state, proposer): null when legal, otherwise the error text.
		 * A check that throws is an error, as a Go panic there would be.
		 */
		legalError(move: MoveInstance, proposer: number): string | null {
			const compiled = this.game.movesByName[move.name];
			if (!compiled) return 'That move is not configured for this game.';
			const base = (): string | null => {
				const result = evaluateLegalPlan(compiled.plan, this.legalContext(compiled, move, proposer));
				const text = renderLegalVerdict(result.verdict, this.game.templates);
				return text === '' ? null : text;
			};
			return compiled.spec.legal ? compiled.spec.legal(this, move, proposer, base) : base();
		},
		/** GameManager.LegalEvaluateLedger: every predicate's verdict. Null for a move whose legality is not declared. */
		legalLedger(move: MoveInstance, proposer: number): LegalPlanResult | null {
			const compiled = this.game.movesByName[move.name];
			if (!compiled || !compiled.declared) return null;
			return evaluateLegalPlan(compiled.plan, this.legalContext(compiled, move, proposer), true);
		},
		/** delegate.ProposeFixUpMove: the first fix-up move legal for the admin. */
		proposeFixUpMove(): (MoveInstance & { version: number; initiator: number }) | null {
			for (const compiled of this.game.moves) {
				if (!compiled.spec.fixUp) continue;
				const move = this.newMove(compiled.spec.name);
				if (move === null) continue;
				let err: string | null;
				try {
					err = this.legalError(move, ADMIN_PLAYER_INDEX);
				} catch (thrown) {
					err = thrown instanceof Error ? thrown.message : String(thrown);
				}
				if (err === null) return move;
			}
			return null;
		},
		// ----- timers --------------------------------------------------------
		/** A Timer property of the game state. */
		timer(prop: string): TimerHandle {
			const record = findTimerRecord(this.engine, { Group: 0, PropName: prop, PlayerIndex: 0, DeckName: '', DeckIndex: 0 });
			if (record === null) throw new Error('timer property no longer exists');
			return newTimerHandle(this, record, this.g[prop] as { ID: string; IsTimer: true });
		},
		/** A Timer property of a player state. */
		playerTimer(player: number, prop: string): TimerHandle {
			const record = findTimerRecord(this.engine, { Group: 1, PropName: prop, PlayerIndex: player, DeckName: '', DeckIndex: 0 });
			if (record === null) throw new Error('timer property no longer exists');
			return newTimerHandle(this, record, this.players[player][prop] as { ID: string; IsTimer: true });
		},
	};
	return self;
}

/** GameManager.moveTypeByName: the move type of that name, compared without regard to case. */
export function findMove(game: CompiledGame, name: string): CompiledMove | undefined {
	const canonical = game.canonicalMoveNames[name.toLowerCase()];
	return canonical === undefined ? undefined : game.movesByName[canonical];
}

/** The RulesState a LegalContext was built from (for a plan's custom check). */
export function rulesOf(ctx: LegalContext): { state: RulesState; move: MoveInstance } {
	const state = (ctx.state as (LegalState & Tagged) | null)?.rules;
	const move = (ctx.move as (LegalMove & Tagged) | null)?.instance;
	if (!state || !move) throw new Error('custom: move does not implement CustomLegaler');
	return { state, move };
}
