/**
 * Start-of-game detection shared by the base game renderer (which blocks a
 * start move while the game reports a ReadyToStartError) and the webapp's
 * gathering panel. Pure reads of expanded state; no store or API types.
 */

/**
 * @deprecated Use IsGatheringStart on MoveForm instead.
 * Kept as a fallback for servers that haven't been updated yet.
 */
export const START_MOVE_NAMES = new Set([
  'Confirm Players',
  'Close All Seats',
  'Start Game',
  'Begin Game',
  'Finalize Set Up',
]);

// ---- State accessors ----

export function optionalRecord(value: unknown, path: string): Readonly<Record<string, unknown>> | null {
  if (value === null || value === undefined) return null;
  if (typeof value !== 'object' || Array.isArray(value)) {
    throw new Error(`Gathering state ${path} must be an object when present`);
  }
  return value as Readonly<Record<string, unknown>>;
}

export function gameComputed(state: unknown): Readonly<Record<string, unknown>> | null {
  const root = optionalRecord(state, 'root');
  const game = optionalRecord(root?.['Game'], 'Game');
  return optionalRecord(game?.['Computed'], 'Game.Computed');
}

/** Get ReadyToStartError from expanded global computed properties. */
export function getReadyToStartError(state: unknown): string {
  const value = gameComputed(state)?.['ReadyToStartError'];
  if (value === undefined || value === null) return '';
  if (typeof value !== 'string') {
    throw new Error('Gathering state Game.Computed.ReadyToStartError must be a string');
  }
  return value;
}
