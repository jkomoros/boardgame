/**
 * State expansion: turns a raw state snapshot from the wire (component indexes
 * into decks) into the expanded state renderers bind to (component objects,
 * boards keyed by space, computed properties copied onto their leaves).
 *
 * Pure, with no store, selector or DOM dependency, so every renderer host
 * expands a snapshot the same way. The webapp's memoized selectors in
 * `selectors.ts` call this; a host without Redux calls it directly.
 */
import type { ExpandedGameState, GameChest } from '../types/wire';
import type { RawGameState, TimerInfo } from '../types/game-state';

/** Expand an API snapshot using exactly the same path as live Redux state. */
export function expandGameStateSnapshot(
    rawState: RawGameState,
    chest: GameChest,
    gameName: string,
    timerInfos: Record<string, TimerInfo> | null = null,
): ExpandedGameState {
    return {
        ...rawState,
        Game: expandLeafState(rawState, rawState.Game, chest, gameName, timerInfos, timerInfos === null),
        Players: rawState.Players.map(player =>
            expandLeafState(rawState, player, chest, gameName, timerInfos, timerInfos === null)),
    };
}

/**
 * Pure function to expand a leaf state object.
 * Walks properties and expands stacks and timers inline.
 *
 * @param skipTimers - If true, timer objects are left unexpanded (performance optimization)
 */
const expandLeafState = (
    wholeState: any,
    leafState: any,
    chest: GameChest,
    gameName: string,
    timerInfos: Record<string, any> | null,
    skipTimers: boolean = false
): any => {
    const result = { ...leafState };

    Object.entries(leafState).forEach(([key, val]) => {
        // Skip null and non-objects
        if (!val || typeof val !== 'object') return;

        // Expand stacks (objects with Deck property)
        if ((val as any).Deck) {
            result[key] = expandStack(val, wholeState, chest, gameName);
        }
        // Expand boards (objects with Spaces array of stacks)
        else if (Array.isArray((val as any).Spaces) && (val as any).Spaces.length > 0 && (val as any).Spaces[0]?.Deck) {
            const spaces = (val as any).Spaces.map((space: any) =>
                expandStack(space, wholeState, chest, gameName)
            );
            const keys = Array.isArray((val as any).Keys) ? (val as any).Keys : undefined;
            result[key] = {
                ...(val as any),
                Spaces: spaces,
                ...(keys?.length === spaces.length ? {
                    SpacesByKey: Object.fromEntries(keys.map((boardKey: string, index: number) =>
                        [boardKey, spaces[index]]
                    )),
                } : {}),
            };
        }
        // Expand timers (objects with IsTimer property) - skip if requested
        else if ((val as any).IsTimer && !skipTimers) {
            result[key] = expandTimer(val, timerInfos);
        }
    });

    // Copy in computed state if it exists
    const pathToLeaf = getPathToLeaf(wholeState, leafState);
    if (pathToLeaf?.length === 2 && pathToLeaf[0] === 'Players') {
        const playerIndex = pathToLeaf[1];
        if (playerIndex !== undefined && wholeState.Computed?.Players?.[playerIndex]) {
            result.Computed = wholeState.Computed.Players[playerIndex];
        }
    }
    if (pathToLeaf?.length === 1 && pathToLeaf[0] === 'Game') {
        if (wholeState.Computed?.Global) {
            result.Computed = wholeState.Computed.Global;
        }
    }

    return result;
};

/**
 * Pure function to expand a stack (deck of components).
 * Returns new object with Components and GameName added.
 */
const expandStack = (
    stack: any,
    wholeState: any,
    chest: GameChest,
    gameName: string
): any => {
    if (!stack.Deck) return stack;

    const components = stack.Indexes.map((index: number, i: number) => {
        if (index === -1) return null;

        // Generic component (index -2)
        if (index === -2) return {};

        // Resolve actual component
        const component = componentForDeckAndIndex(stack.Deck, index, wholeState, chest);
        if (!component) return null;

        // Add ID if available
        const result = { ...component };
        if (stack.IDs?.[i]) {
            result.ID = stack.IDs[i];
        }
        result.Deck = stack.Deck;
        result.GameName = gameName;

        return result;
    });

    return {
        ...stack,
        GameName: gameName,
        Components: components,
    };
};

/**
 * Pure function to expand a timer.
 * Returns new object with TimeLeft and originalTimeLeft added.
 * originalTimeLeft is the value when first received; TimeLeft is updated by tick.
 */
const expandTimer = (
    timer: any,
    timerInfos: Record<string, any> | null
): any => {
    const result = {
        ...timer,
        TimeLeft: 0,
        originalTimeLeft: 0,
    };

    if (timerInfos?.[timer.ID]) {
        const info = timerInfos[timer.ID];
        result.TimeLeft = info.TimeLeft;
        // originalTimeLeft should be preserved from when timer was first installed
        result.originalTimeLeft = info.originalTimeLeft ?? info.TimeLeft;
    }

    return result;
};

/**
 * Helper to get component for a deck and index.
 */
const componentForDeckAndIndex = (
    deckName: string,
    index: number,
    wholeState: any,
    chest: GameChest
): any | null => {
    const deck = (chest as any).Decks?.[deckName];
    if (!deck) return null;

    const result = { ...deck[index] };

    // Add dynamic values if available
    if (wholeState.Components?.[deckName]?.[index]) {
        result.DynamicValues = wholeState.Components[deckName][index];
    }

    return result;
};

/**
 * Helper to determine the path to a leaf state within the whole state.
 * Used to identify Player states for Computed property copying.
 */
const getPathToLeaf = (wholeState: any, leafState: any): (string | number)[] | null => {
    // Check if it's the Game
    if (wholeState.Game === leafState) {
        return ['Game'];
    }

    // Check if it's a Player
    if (wholeState.Players) {
        const index = wholeState.Players.indexOf(leafState);
        if (index !== -1) {
            return ['Players', index];
        }
    }

    return null;
};
