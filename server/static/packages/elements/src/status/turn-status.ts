export const ObserverPlayerIndex = -1 as const;
export const AdminPlayerIndex = -2 as const;
export const AnyPlayerIndex = -3 as const;

export type SpecialPlayerIndex =
  | typeof ObserverPlayerIndex
  | typeof AdminPlayerIndex
  | typeof AnyPlayerIndex;

export interface TurnStatusContext {
  readonly currentPlayerIndex: number;
  readonly viewerPlayerIndex: number;
  readonly finished: boolean;
  readonly animating: boolean;
}

export type TurnStatusKind = 'active' | 'simultaneous' | 'waiting';

export interface TurnStatusPresentation {
  readonly kind: TurnStatusKind;
  readonly message: string;
}

export function isConcretePlayerIndex(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0;
}

export function isKnownPlayerIndex(value: unknown): value is number {
  return isConcretePlayerIndex(value)
    || value === ObserverPlayerIndex
    || value === AdminPlayerIndex
    || value === AnyPlayerIndex;
}

export function turnStatusPresentation(
  context: TurnStatusContext,
  playerLabels: readonly string[] = [],
  activeLabel = 'Your turn',
  simultaneousLabel = 'All players may act',
): TurnStatusPresentation | null {
  validateContext(context);
  validateLabels(playerLabels);
  if (typeof activeLabel !== 'string' || !activeLabel.trim()) {
    throw new Error('boardgame-turn-status: activeLabel must be a non-empty string');
  }
  if (typeof simultaneousLabel !== 'string' || !simultaneousLabel.trim()) {
    throw new Error('boardgame-turn-status: simultaneousLabel must be a non-empty string');
  }
  if (context.finished || context.animating) return null;

  if (context.currentPlayerIndex === AnyPlayerIndex) {
    return isConcretePlayerIndex(context.viewerPlayerIndex)
      ? { kind: 'active', message: activeLabel.trim() }
      : { kind: 'simultaneous', message: simultaneousLabel.trim() };
  }
  if (!isConcretePlayerIndex(context.currentPlayerIndex)) return null;
  if (context.currentPlayerIndex === context.viewerPlayerIndex) {
    return { kind: 'active', message: activeLabel.trim() };
  }

  const label = playerLabels[context.currentPlayerIndex] ?? `Player ${context.currentPlayerIndex + 1}`;
  return { kind: 'waiting', message: `${label}'s turn` };
}

function validateContext(context: TurnStatusContext): void {
  if (!context || typeof context !== 'object') {
    throw new Error('boardgame-turn-status: turn must be a renderer turn-status context');
  }
  const keys = Object.keys(context).sort();
  const expected = ['animating', 'currentPlayerIndex', 'finished', 'viewerPlayerIndex'];
  if (keys.length !== expected.length || keys.some((key, index) => key !== expected[index])) {
    throw new Error(`boardgame-turn-status: turn must contain exactly ${expected.join(', ')}`);
  }
  if (!isKnownPlayerIndex(context.currentPlayerIndex)) {
    throw new Error('boardgame-turn-status: currentPlayerIndex must be a concrete player or a known framework sentinel');
  }
  if (!isKnownPlayerIndex(context.viewerPlayerIndex) || context.viewerPlayerIndex === AnyPlayerIndex) {
    throw new Error('boardgame-turn-status: viewerPlayerIndex must be a concrete player, ObserverPlayerIndex, or AdminPlayerIndex');
  }
  if (typeof context.finished !== 'boolean' || typeof context.animating !== 'boolean') {
    throw new Error('boardgame-turn-status: finished and animating must be booleans');
  }
}

function validateLabels(playerLabels: readonly string[]): void {
  if (!Array.isArray(playerLabels)) throw new Error('boardgame-turn-status: playerLabels must be an array');
  for (const label of playerLabels) {
    if (typeof label !== 'string' || !label.trim()) {
      throw new Error('boardgame-turn-status: playerLabels must contain only non-empty strings');
    }
  }
}
