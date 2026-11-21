// Werewolf: a hidden-role game. Villagers vote by day to eliminate a
// suspect; werewolves vote by night to eliminate a villager. A move-for-move
// port of examples/werewolf. It is the example of state hidden per player (a
// role only its owner sees), of group-based visibility (a night vote is
// visible to players of the same role), of simultaneous play (anyone may act;
// there is no current player), and of a deadline timer.

import { ADMIN_PLAYER_INDEX, ANY_PLAYER_INDEX, OBSERVER_PLAYER_INDEX } from '../legal/types.ts';
import { checkGameFinished } from '../rules/engine.ts';
import type { RulesState } from '../rules/state.ts';
import type { GameSpec, LegalResult, MoveSpec } from '../rules/types.ts';
import {
	activateInactivePlayer, add, addForPhase, addOrderedForPhase, anyPlayerLegal, combine, eliminationProps, inactivePlayerProps, inactivateEmptySeat,
	optional, phaseProps, roleProps, seatPlayer, seatProps, startPhase, waitForEnoughPlayers,
} from '../rules/stock.ts';

const PHASE_GATHERING = 0;
const PHASE_DAY = 1;
const PHASE_NIGHT = 2;
const VILLAGER = 'Villager';
const WEREWOLF = 'Werewolf';
const VOTE_DURATION_MS = 45000;
const DAY_TIMEOUT = 'Resolve Day Votes on Timeout';
const NIGHT_TIMEOUT = 'Resolve Night Votes on Timeout';

const timedVoting = (state: RulesState): boolean => state.variant().timedvoting === 'on';
const isWerewolf = (state: RulesState, player: number): boolean => state.players[player].Role === WEREWOLF;
const eliminated = (state: RulesState, player: number): boolean => state.players[player].Eliminated === true;

/** The vote a player has cast in this phase, or a negative index for none. */
function voteForPhase(state: RulesState, player: number, phase: number): number {
	return (phase === PHASE_NIGHT ? state.players[player].NightVote : state.players[player].DayVote) as number;
}

function startVoteTimer(state: RulesState): void {
	if (!timedVoting(state)) return;
	const name = state.currentPhaseKey() === PHASE_NIGHT ? NIGHT_TIMEOUT : DAY_TIMEOUT;
	// Bound to the state being built, as Go's MoveByNameForState is here.
	const completion = state.newMove(name);
	if (completion === null) throw new Error('timed vote resolver was not configured');
	state.timer('VoteTimer').start(VOTE_DURATION_MS, completion);
}

/** Each werewolf learns who the other werewolves are. */
function populateFellowWolves(state: RulesState): void {
	const wolves: number[] = [];
	for (let i = 0; i < state.numPlayers(); i++) {
		state.players[i].FellowWolves = null;
		if (state.playerIsInactive(i) || !isWerewolf(state, i)) continue;
		wolves.push(i);
	}
	for (const wolf of wolves) {
		const fellows = wolves.filter(other => other !== wolf);
		if (fellows.length > 0) state.players[wolf].FellowWolves = fellows;
	}
}

/** Tally the phase's votes, eliminate the choice, clear the votes, move to the other phase. */
function resolveVotes(state: RulesState): void {
	const phase = state.currentPhaseKey();
	const counts: number[] = state.players.map(() => 0);
	for (let i = 0; i < state.numPlayers(); i++) {
		if (state.playerIsInactive(i) || eliminated(state, i)) continue;
		if (phase === PHASE_NIGHT && !isWerewolf(state, i)) continue;
		const vote = voteForPhase(state, i, phase);
		if (vote >= 0) counts[vote] = (counts[vote] ?? 0) + 1;
	}
	let maxVotes = 0;
	let chosen = -1;
	let tied = false;
	for (let target = 0; target < state.numPlayers(); target++) {
		const count = counts[target];
		if (count === 0) continue;
		if (count > maxVotes) {
			maxVotes = count;
			chosen = target;
			tied = false;
		} else if (count === maxVotes) {
			tied = true;
		}
	}
	// A tied day vote eliminates nobody; a tied night vote keeps the first.
	if (phase === PHASE_DAY && tied) chosen = -1;
	if (chosen >= 0 && chosen < state.numPlayers()) state.players[chosen].Eliminated = true;
	for (const player of state.players) {
		player.DayVote = -1;
		player.NightVote = -1;
	}
	if (phase === PHASE_DAY) {
		state.setPhase(PHASE_NIGHT);
	} else {
		state.g.RoundNumber = (state.g.RoundNumber as number) + 1;
		state.setPhase(PHASE_DAY);
	}
}

function finishVoteResolution(state: RulesState): void {
	state.timer('VoteTimer').cancel();
	resolveVotes(state);
	if (checkGameFinished(state).finished) return;
	startVoteTimer(state);
}

const beginGame: MoveSpec = startPhase({
	phase: PHASE_DAY,
	phaseName: 'Day',
	name: 'Begin Game',
	helpText: 'Starts the day phase and assigns roles among seated players.',
	apply(state) {
		state.setPhase(PHASE_DAY);
		const active: number[] = [];
		for (let i = 0; i < state.numPlayers(); i++) if (!state.playerIsInactive(i)) active.push(i);
		const numWerewolves = active.length >= 6 ? 2 : 1;
		// One draw: a permutation whose first entries name the werewolves.
		const indices = state.rand().perm(active.length);
		active.forEach((player, i) => {
			state.players[player].Role = indices.slice(0, numWerewolves).indexOf(i) >= 0 ? WEREWOLF : VILLAGER;
		});
		populateFellowWolves(state);
		startVoteTimer(state);
	},
});

/** One spec for both votes: the phase it is installed in decides which vote it casts. */
function castVote(options: Pick<MoveSpec, 'name' | 'helpText' | 'nameSanitization' | 'hiddenAnimationKey'>): MoveSpec {
	return {
		...options,
		fixUp: false,
		base: 'anyPlayer',
		// The voter and players of the voter's role see who was voted for.
		fields: { VoteTarget: { type: 'playerIndex', sanitize: 'self:visible,same-role:visible' } },
		// Offer the voter each player they may currently vote for.
		choices: { field: 'VoteTarget' },
		legal(state, move, proposer, base): LegalResult {
			const err = anyPlayerLegal(state, move, proposer, base);
			if (err !== null) return err;
			const voter = move.fields.TargetPlayerIndex as number;
			if (eliminated(state, voter)) return 'eliminated players cannot vote';
			const phase = state.currentPhaseKey();
			if (voteForPhase(state, voter, phase) >= 0) return 'you have already voted this phase';
			if (phase === PHASE_NIGHT) {
				if (!isWerewolf(state, voter)) return 'only werewolves may act during the night';
			} else if (phase !== PHASE_DAY) {
				return 'voting is not allowed during this phase';
			}
			const target = move.fields.VoteTarget as number;
			if (target < 0 || target >= state.numPlayers()) return 'invalid vote target';
			if (eliminated(state, target)) return 'cannot vote for an eliminated player';
			if (state.playerIsInactive(target)) return 'cannot vote for an inactive player';
			if (target === voter) return 'you cannot vote for yourself';
			return null;
		},
		apply(state, move) {
			const voter = state.players[move.fields.TargetPlayerIndex as number];
			if (state.currentPhaseKey() === PHASE_NIGHT) voter.NightVote = move.fields.VoteTarget;
			else voter.DayVote = move.fields.VoteTarget;
		},
	};
}

/** A fix-up: once everyone eligible has voted, resolve the vote. */
function resolveVotesMove(options: Pick<MoveSpec, 'name' | 'helpText'>): MoveSpec {
	return {
		...options,
		fixUp: true,
		legal(state, _move, _proposer, base) {
			const err = base();
			if (err !== null) return err;
			const phase = state.currentPhaseKey();
			if (phase !== PHASE_DAY && phase !== PHASE_NIGHT) return 'not in a voting phase';
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i) || eliminated(state, i)) continue;
				if (phase === PHASE_NIGHT && !isWerewolf(state, i)) continue;
				if (voteForPhase(state, i, phase) < 0) return 'not all eligible players have voted';
			}
			return null;
		},
		apply: finishVoteResolution,
	};
}

/** Not a fix-up: only the vote timer proposes it, resolving with the votes cast so far. */
function resolveOnTimeout(options: Pick<MoveSpec, 'name' | 'helpText'>): MoveSpec {
	return {
		...options,
		fixUp: false,
		legal(state, _move, proposer, base) {
			const err = base();
			if (err !== null) return err;
			if (proposer !== ADMIN_PLAYER_INDEX) return 'only the timer may resolve incomplete votes';
			if (!timedVoting(state) || !state.timer('VoteTimer').active()) return 'timed vote resolution is not active';
			return null;
		},
		apply: finishVoteResolution,
	};
}

const ROLES = { Values: { 0: VILLAGER, 1: WEREWOLF }, DefaultValue: 0 };

export const werewolf: GameSpec = {
	name: 'werewolf',
	minNumPlayers: 4,
	maxNumPlayers: 7,
	defaultNumPlayers: 5,

	decks: {},
	enums: {
		phase: { Values: { 0: 'Gathering', 1: 'Day', 2: 'Night' }, DefaultValue: 0 },
		role: ROLES,
		// The enum named "group" makes its values sanitization groups: a
		// property or move may be visible to "same-role" players.
		group: ROLES,
	},

	game: {
		props: { ...phaseProps, RoundNumber: { Type: 'int' }, VoteTimer: { Type: 'timer' } },
	},
	player: {
		props: {
			...seatProps,
			...inactivePlayerProps,
			...eliminationProps,
			...roleProps,
			DayVote: { Type: 'playerIndex' },
			NightVote: { Type: 'playerIndex', Sanitize: 'other:hidden' },
			FellowWolves: { Type: 'playerIndexSlice', Sanitize: 'other:hidden' },
		},
		// -1 means "has not voted".
		initial: () => ({ DayVote: -1, NightVote: -1 }),
		traits: { seat: true, inactive: true, elimination: true, role: true },
	},

	variants: {
		timedvoting: {
			displayName: 'Timed Voting',
			description: 'Resolve each day and night vote after 45 seconds, using the votes cast so far.',
			default: 'off',
			values: { off: { displayName: 'Off' }, on: { displayName: '45 seconds' } },
		},
	},

	moves: combine(
		add(seatPlayer()),
		addOrderedForPhase(
			PHASE_GATHERING,
			optional(activateInactivePlayer({ nameSuffix: 'Gathering' })),
			waitForEnoughPlayers({ nameSuffix: 'Gathering' }),
			optional(inactivateEmptySeat({ nameSuffix: 'Gathering' })),
			beginGame,
		),
		addForPhase(
			PHASE_DAY,
			castVote({ name: 'Cast Vote', helpText: 'Vote for a player to eliminate.' }),
			resolveVotesMove({ name: 'Resolve Votes', helpText: 'Tallies votes, eliminates the chosen player, and transitions to the next phase.' }),
			resolveOnTimeout({ name: DAY_TIMEOUT, helpText: 'Resolves the daytime deadline using votes cast so far.' }),
		),
		addForPhase(
			PHASE_NIGHT,
			castVote({
				name: 'Cast Night Vote',
				helpText: 'Werewolves choose a target to eliminate.',
				// Villagers must not learn that a werewolf acted, or who did.
				nameSanitization: 'self:visible,same-role:visible',
				hiddenAnimationKey: 'Hidden Action',
			}),
			resolveVotesMove({ name: 'Resolve Night Votes', helpText: 'Tallies werewolf votes, eliminates the target, and transitions to day.' }),
			resolveOnTimeout({ name: NIGHT_TIMEOUT, helpText: 'Resolves the nighttime deadline using votes cast so far.' }),
		),
	),

	setUp: {
		begin(state) {
			for (const player of state.players) {
				player.DayVote = -1;
				player.NightVote = -1;
			}
		},
		distribute: () => {
			throw new Error('DistributeComponentToStarterStack was called, but the component was not stored in a stack');
		},
	},

	// Everyone acts at once while a vote is open; nobody may act while gathering.
	currentPlayerIndex: state => {
		const phase = state.currentPhaseKey();
		return phase === PHASE_DAY || phase === PHASE_NIGHT ? ANY_PLAYER_INDEX : OBSERVER_PLAYER_INDEX;
	},

	end: {
		// Villagers win when no werewolf is left; werewolves when they are
		// at least as many as the villagers.
		check(state) {
			if (state.currentPhaseKey() === PHASE_GATHERING) return { finished: false, winners: [] };
			let wolves = 0;
			let villagers = 0;
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i) || eliminated(state, i)) continue;
				if (isWerewolf(state, i)) wolves++;
				else villagers++;
			}
			const villagersWin = wolves === 0;
			const werewolvesWin = !villagersWin && wolves >= villagers;
			if (!villagersWin && !werewolvesWin) return { finished: false, winners: [] };
			const winners: number[] = [];
			for (let i = 0; i < state.numPlayers(); i++) {
				if (state.playerIsInactive(i)) continue;
				if (villagersWin !== isWerewolf(state, i)) winners.push(i);
			}
			return { finished: true, winners };
		},
	},
};
