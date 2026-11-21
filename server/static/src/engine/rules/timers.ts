// Timers as state. Ports the durable half of timer.go: a Timer property's
// lifecycle lives in the state's storage record (the top-level "Timers"
// list), and nothing lives outside it. The engine has no clock and no
// scheduler. A timer is a deadline in the state; when the host sees that
// deadline pass it proposes a tick, which the engine validates against the
// timer's record and the host's "now".

import type { GameDefinition, SubStateDef } from '../definition.ts';
import type { EngineState, StateJSON, SubStateJSON } from '../state.ts';
import { ZERO_TIME, formatInstant, parseInstant } from './time.ts';

/** Go's zero time: a timer's intended move has never been applied. */
export const ZERO_TIMER_MOVE_TIMESTAMP = ZERO_TIME;

/** StatePropertyRef: Group 0 is the game state, 1 a player state, 3 dynamic component values. */
export interface TimerRef {
	Group: number;
	PropName: string;
	PlayerIndex: number;
	DeckName: string;
	DeckIndex: number;
}

export type TimerStatus = 'inactive' | 'active' | 'canceled' | 'fired';

/** The move a timer will propose as the admin (a MoveStorageRecord). */
export interface TimerMove {
	Name: string;
	/** The version that started the timer, not the one the move will get. */
	Version: number;
	Initiator: number;
	Timestamp: string;
	Phase: number;
	Proposer: number;
	Blob: Record<string, unknown>;
}

/** persistedTimerRecord. */
export interface TimerRecord {
	Ref: TimerRef;
	/** '' when not active. Drawn from the state's generator. */
	ID: string;
	Generation: number;
	/** RFC 3339. Go's zero time until the timer first starts. */
	Deadline: string;
	Status: TimerStatus;
	/** Present while active. */
	Move?: TimerMove;
}

const GROUP_GAME = 0;
const GROUP_PLAYER = 1;
const GROUP_DYNAMIC = 3;

function timerProps(def: SubStateDef): string[] {
	const names: string[] = [];
	for (const name of Object.keys(def.Props)) if (def.Props[name].Type === 'timer') names.push(name);
	return names.sort();
}

function sameRef(a: TimerRef, b: TimerRef): boolean {
	return a.Group === b.Group && a.PropName === b.PropName && a.PlayerIndex === b.PlayerIndex && a.DeckName === b.DeckName && a.DeckIndex === b.DeckIndex;
}

/** timerLocations: every Timer property of a state, in Go's order. */
export function timerRefs(def: GameDefinition, data: StateJSON): TimerRef[] {
	const refs: TimerRef[] = [];
	for (const name of timerProps(def.Game)) refs.push({ Group: GROUP_GAME, PropName: name, PlayerIndex: 0, DeckName: '', DeckIndex: 0 });
	const playerProps = timerProps(def.Player);
	for (let i = 0; i < data.Players.length; i++) {
		for (const name of playerProps) refs.push({ Group: GROUP_PLAYER, PropName: name, PlayerIndex: i, DeckName: '', DeckIndex: 0 });
	}
	for (const deckName of Object.keys(data.Components).sort()) {
		const sub = def.DynamicComponentValues[deckName];
		if (!sub) continue;
		const props = timerProps(sub);
		for (let i = 0; i < data.Components[deckName].length; i++) {
			for (const name of props) refs.push({ Group: GROUP_DYNAMIC, PropName: name, PlayerIndex: 0, DeckName: deckName, DeckIndex: i });
		}
	}
	return refs;
}

function timerNode(data: StateJSON, ref: TimerRef): SubStateJSON {
	if (ref.Group === GROUP_GAME) return data.Game;
	if (ref.Group === GROUP_PLAYER) return data.Players[ref.PlayerIndex];
	return data.Components[ref.DeckName][ref.DeckIndex];
}

/** The state's timer records, as its JSON holds them. */
export function timerRecords(state: EngineState): TimerRecord[] {
	return (state.data.Timers ?? []) as TimerRecord[];
}

/**
 * Makes the state's "Timers" list hold one record per Timer property, in
 * Go's order, keeping the records already there (timerStorageRecords after
 * restoreTimerStorageRecords). A missing record is an inactive timer.
 */
export function syncTimerRecords(state: EngineState): void {
	const existing = timerRecords(state);
	const records: TimerRecord[] = [];
	for (const ref of timerRefs(state.def, state.data)) {
		let record: TimerRecord | undefined;
		for (const candidate of existing) if (sameRef(candidate.Ref, ref)) record = candidate;
		if (!record) record = { Ref: ref, ID: '', Generation: 0, Deadline: ZERO_TIME, Status: 'inactive' };
		records.push(record);
		const node = timerNode(state.data, ref);
		node[ref.PropName] = { ID: record.ID, IsTimer: true };
	}
	state.data.Timers = records;
}

export function findTimerRecord(state: EngineState, ref: TimerRef): TimerRecord | null {
	for (const record of timerRecords(state)) if (sameRef(record.Ref, ref)) return record;
	return null;
}

/** timerGuard: which start of which timer a tick claims to complete. */
export interface TimerGuard {
	Ref: TimerRef;
	ID: string;
	Generation: number;
}

/** validateTimerGuard: null, or Go's error text. */
export function timerGuardError(state: EngineState, guard: TimerGuard): string | null {
	const record = findTimerRecord(state, guard.Ref);
	if (record === null) return 'timer property no longer exists';
	if (record.Status !== 'active' || record.ID !== guard.ID || record.Generation !== guard.Generation) {
		return 'timer completion was stale or already handled';
	}
	return null;
}

/** consumeTimer: marks the guarded timer fired on the state being built. */
export function consumeTimer(state: EngineState, guard: TimerGuard): string | null {
	const err = timerGuardError(state, guard);
	if (err !== null) return err;
	const record = findTimerRecord(state, guard.Ref) as TimerRecord;
	record.Status = 'fired';
	delete record.Move;
	return null;
}

/** finalizeTimerDeadlines: a started timer's deadline is the save time plus its duration. */
export function finalizeTimerDeadlines(pending: { record: TimerRecord; durationMs: number }[], nowMs: number): void {
	for (const item of pending) {
		if (item.record.Status !== 'active') continue;
		item.record.Deadline = formatInstant(nowMs + item.durationMs);
	}
	pending.length = 0;
}

/** One active timer, as a host schedules it. */
export interface ActiveTimer {
	guard: TimerGuard;
	/** Milliseconds since the epoch. */
	deadlineMs: number;
	move: TimerMove;
}

/**
 * The active timers of a stored state, earliest deadline first: what a host
 * must wake up for (Go: TimerWakeupsFromStateStorage, with the move intent).
 */
export function activeTimers(stateJSON: StateJSON): ActiveTimer[] {
	const result: ActiveTimer[] = [];
	for (const record of (stateJSON.Timers ?? []) as TimerRecord[]) {
		if (record.Status !== 'active' || record.ID === '' || !record.Move) continue;
		const deadlineMs = parseInstant(record.Deadline);
		if (deadlineMs === null) continue;
		result.push({ guard: { Ref: record.Ref, ID: record.ID, Generation: record.Generation }, deadlineMs, move: record.Move });
	}
	// Stable, so timers with equal deadlines keep the state's order.
	return result
		.map((timer, index) => ({ timer, index }))
		.sort((a, b) => a.timer.deadlineMs - b.timer.deadlineMs || a.index - b.index)
		.map(item => item.timer);
}

/**
 * The "ActiveTimers" block of a viewer's game object: timer id to the
 * milliseconds left at `nowMs` (never negative), for each active timer.
 */
export function activeTimersView(stateJSON: StateJSON, nowMs: number): Record<string, { TimeLeft: number }> {
	const result: Record<string, { TimeLeft: number }> = {};
	for (const timer of activeTimers(stateJSON)) {
		const left = timer.deadlineMs - nowMs;
		result[timer.guard.ID] = { TimeLeft: left < 0 ? 0 : left };
	}
	return result;
}
